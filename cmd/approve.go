package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shreyasrama/cph/pkg/awsutil"

	"github.com/spf13/cobra"
)

// approveCmd represents the approve command
var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve CodePipelines based on a provided search term.",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		return approvePipelines(name)
	},
}

func init() {
	rootCmd.AddCommand(approveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	approveCmd.PersistentFlags().String("name", "", "Use a name or part of a name to filter the runnable pipelines.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// TODO: add pipeline flag?
}

func approvePipelines(searchTerm string) error {
	cp, err := awsutil.CreateCodePipelineSession()
	if err != nil {
		return err
	}

	pipeline_names, err := awsutil.GetPipelineNames(cp, searchTerm)
	if err != nil {
		return err
	}

	// Iterate over pipeline names and get the most recent pipeline
	// execution status and create a slice of structs
	stagesToApprove := make(map[string]awsutil.StageInfo)
	for _, name := range pipeline_names {
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
	pipeline_map := make(map[int]string)
	fmt.Printf("\n%s\n", "The following pipelines have been found:")
	i := 0
	for pipeline := range stagesToApprove {
		pipeline_map[i+1] = pipeline
		fmt.Printf("    [%v] %s (%s)\n", i+1, pipeline, stagesToApprove[pipeline].StageName)
		i++
	}
	var s string
	fmt.Printf(
		"\n%s",
		`Do you want to approve these pipelines?
Enter 'yes' to run all, 'no' to cancel, or a number for a specific pipeline: `)
	fmt.Scan(&s)
	// Check if number is entered
	if i, err := strconv.Atoi(s); err == nil {
		stageToApprove := make(map[string]awsutil.StageInfo)
		stageToApprove[pipeline_map[i]] = stagesToApprove[pipeline_map[i]]
		err := awsutil.ApprovePipelines(cp, stageToApprove)
		if err != nil {
			return err
		}
		fmt.Printf("Approved %s\n", pipeline_map[i]) //
	} else if strings.EqualFold(s, "yes") {
		fmt.Println("Approving pipelines...")
		err := awsutil.ApprovePipelines(cp, stagesToApprove)
		if err != nil {
			return err
		}

		for name := range stagesToApprove {
			fmt.Printf("Approved %s\n", name)
		}
	} else if strings.EqualFold(s, "no") {
		fmt.Println("Cancelled.")
		// exit?
	} else {
		fmt.Println("Input not recognised.")
		// exit?
	}

	return nil
}
