package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/spf13/cobra"

	"github.com/shreyasrama/cph/pkg/awsutil"
	"github.com/shreyasrama/cph/pkg/helpers"
)

// approveCmd represents the approve command
var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve CodePipelines based on a provided search term.",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("You must pass in a --name flag when approving pipelines.")
			return err
		}
		message, _ := cmd.Flags().GetString("message")

		return approvePipelines(name, message)
	},
}

func init() {
	rootCmd.AddCommand(approveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	approveCmd.PersistentFlags().String("name", "", "Use a name or part of a name to filter the runnable pipelines.")
	approveCmd.PersistentFlags().String("message", "", "Add a message for the approval action")
	// TODO: add pipeline flag?
}

// Core logic for the approve feature.
// Notable data structures/variables:
// pipelineNames []string - names of the pipeline that the search returned.
// pipelineMap (map[int]string) - maps the number the pipeline corresponds to in the search results to its name.
// executionTable (var) - table that presents the output from the run command.
func approvePipelines(searchTerm string, message string) error {
	cp, err := awsutil.CreateCodePipelineSession()
	if err != nil {
		return err
	}

	pipelineNames, err := awsutil.GetPipelineNames(cp, searchTerm)
	if err != nil {
		return err
	}

	// Iterate over pipeline names and get the most recent pipeline
	// execution status and create a map of names to StageInfo
	stagesToApprove := make(map[string]awsutil.StageInfo)
	for _, name := range pipelineNames {
		stageInfo, err := awsutil.GetLastExecutedStage(cp, name)
		if err != nil {
			return err
		}
		if stageInfo.Status == "InProgress" {
			stagesToApprove[name] = stageInfo
		}
	}

	if len(stagesToApprove) == 0 {
		fmt.Println("No pipelines to approve.")
		return nil
	}

	// Print and confirm pipelines to be approved
	pipelineMap := make(map[int]string)
	fmt.Printf("\n%s\n", "The following pipelines have been found:")
	i := 0
	// TODO: executionTable?
	for pipeline := range stagesToApprove {
		pipelineMap[i+1] = pipeline
		fmt.Printf("    [%v] %s (%s)\n", i+1, pipeline, stagesToApprove[pipeline].StageName)
		i++
	}

	var s string

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf(
		"\n%s",
		`Do you want to approve these pipelines?
Enter 'yes' to approve all, 'no' to cancel, 'reject' to reject all, a number for a specific pipeline, or provide a range or list: `)
	if scanner.Scan() {
		s = scanner.Text()
	}
	if len(message) > 0 {
		scanner = bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			message = scanner.Text()
		}
		fmt.Printf(
			"\n%s",
			`Do you want to add a message? E.g. A Change Order number. This will apply for all pipelines.`)
		if scanner.Scan() {
			s = scanner.Text()
		}
	}

	if i, err := strconv.Atoi(s); err == nil { // User enters a single number
		stageToApprove := make(map[string]awsutil.StageInfo)
		stageToApprove[pipelineMap[i]] = stagesToApprove[pipelineMap[i]]
		err := awsutil.ApprovePipelines(cp, stageToApprove, codepipeline.ApprovalStatusApproved, message)
		if err != nil {
			return err
		}
		fmt.Printf("Approved %s\n", pipelineMap[i])

	} else if strings.EqualFold(s, "yes") {
		fmt.Println("Approving pipelines...")
		err := awsutil.ApprovePipelines(cp, stagesToApprove, codepipeline.ApprovalStatusApproved, message)
		if err != nil {
			return err
		}

		for name := range stagesToApprove {
			fmt.Printf("Approved %s\n", name)
		}

	} else if strings.EqualFold(s, "no") {
		fmt.Println("Cancelled.")
		// exit?

	} else if strings.EqualFold(s, "reject") {
		fmt.Println("Rejecting pipelines...")
		err := awsutil.ApprovePipelines(cp, stagesToApprove, codepipeline.ApprovalStatusRejected, message)
		if err != nil {
			return err
		}

		for name := range stagesToApprove {
			fmt.Printf("Rejected %s\n", name)
		}
	} else { // User enters a range
		rangeMatch, _ := regexp.MatchString("^[0-9]{1,2}-[0-9]{1,2}$", s) // e.g. 1-3, 2-6
		selectionMatch, _ := regexp.MatchString(`(\d+)(,\s*\d+)*`, s)     // e.g. 1,3,5

		if rangeMatch {
			pipelinesToApprove, err := helpers.ProcessInputRange(s, len(pipelineNames))
			if err != nil {
				return err
			}

			// Approve pipelines
			approveStages := make(map[string]awsutil.StageInfo)
			for i := range pipelinesToApprove {
				approveStages[pipelineMap[pipelinesToApprove[i]]] = stagesToApprove[pipelineMap[pipelinesToApprove[i]]]
			}

			approveMultiInputPipelines(cp, approveStages, pipelineMap, message)

		} else if selectionMatch {
			pipelinesToApprove, err := helpers.ProcessInputSelection(s, len(pipelineNames))
			if err != nil {
				return err
			}

			// Approve pipelines
			approveStages := make(map[string]awsutil.StageInfo)
			for i := range pipelinesToApprove {
				approveStages[pipelineMap[pipelinesToApprove[i]]] = stagesToApprove[pipelineMap[pipelinesToApprove[i]]]
			}

			approveMultiInputPipelines(cp, approveStages, pipelineMap, message)

		} else {
			fmt.Println("Input not recognised.")
			// exit?
		}
	}

	return nil
}

// For range and selection inputs.
// Takes map of pipeline names -> their approval stage to approve the appropriate pipelines
func approveMultiInputPipelines(cp *codepipeline.CodePipeline, stagesToApprove map[string]awsutil.StageInfo, pipelineMap map[int]string, message string) error {
	fmt.Println("Approving pipelines...")
	err := awsutil.ApprovePipelines(cp, stagesToApprove, codepipeline.ApprovalStatusApproved, message)
	if err != nil {
		return err
	}
	for name := range stagesToApprove {
		fmt.Printf("Approved %s\n", name)
	}

	return nil
}
