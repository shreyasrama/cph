package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/shreyasrama/cph/pkg/awsutil"
	"github.com/shreyasrama/cph/pkg/helpers"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run CodePipelines based on a provided search term.",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		return runPipelines(name)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	runCmd.PersistentFlags().String("name", "", "Use a name or part of a name to filter the runnable pipelines.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// TODO: add pipeline flag?
}

// Core logic for the run feature.
// Notable data structures/variables:
// pipelineNames []string - names of the pipeline that the search returned.
// pipelineMap (map[int]string) - maps the number the pipeline corresponds to in the search results to its name.
// executionTable (var) - table that presents the output from the run command.
func runPipelines(searchTerm string) error {
	cp, err := awsutil.CreateCodePipelineSession()
	if err != nil {
		return err
	}

	pipelineNames, err := awsutil.GetPipelineNames(cp, searchTerm)
	if err != nil {
		return err
	}

	// Print and confirm pipelines to be run
	pipelineMap := make(map[int]string)
	fmt.Printf("\n%s\n", "The following pipelines have been found:")
	for i, pipeline := range pipelineNames {
		pipelineMap[i+1] = pipeline
		fmt.Printf("    [%v] %s\n", i+1, pipeline)
	}

	executionTable := helpers.SetupTable([]string{"Pipeline", "Execution ID"})

	var s string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf(
		"\n%s",
		`Do you want to run these pipelines?
Enter 'yes' to run all, 'no' to cancel, a number for a specific pipeline, or provide a range or list: `)
	if scanner.Scan() {
		s = scanner.Text()
	}

	if i, err := strconv.Atoi(s); err == nil { // User enters a single number
		executionId, err := awsutil.RunPipeline(cp, pipelineMap[i])
		if err != nil {
			return err
		}
		fmt.Printf("Started execution of %s. Execution ID: %s", pipelineMap[i], executionId)

	} else if strings.EqualFold(s, "yes") {
		fmt.Println("Running pipelines...")
		executionIds, err := awsutil.RunPipelines(cp, pipelineNames)
		if err != nil {
			return err
		}
		renderExecutionTable(executionIds, executionTable)

	} else if strings.EqualFold(s, "no") {
		fmt.Println("Cancelled.")
		// exit?

	} else { // User enters a range
		rangeMatch, _ := regexp.MatchString("^[0-9]{1,2}-[0-9]{1,2}$", s) // e.g. 1-3, 2-6
		selectionMatch, _ := regexp.MatchString(`(\d+)(,\s*\d+)*`, s)     // e.g. 1,3,5

		if rangeMatch {
			pipelinesToRun, err := helpers.ProcessInputRange(s, len(pipelineNames))
			if err != nil {
				return err
			}

			runMultiInputPipelines(cp, pipelinesToRun, pipelineMap, executionTable)

		} else if selectionMatch {
			pipelinesToRun, err := helpers.ProcessInputSelection(s, len(pipelineNames))
			if err != nil {
				return err
			}

			// Run pipelines and set up table
			runMultiInputPipelines(cp, pipelinesToRun, pipelineMap, executionTable)

		} else {
			fmt.Println("Input not recognised.")
			// exit?
		}
	}

	return nil
}

// Helper function that will render the table to the terminal.
func renderExecutionTable(executionIds map[string]string, executionTable *tablewriter.Table) {
	for id, name := range executionIds {
		executionTable.Append([]string{
			name,
			id,
		})
	}
	executionTable.Render()
}

// For range and selection inputs.
// Takes processed user input and the pipelineMap to run the appropriate pipelines
// and display the results.
func runMultiInputPipelines(cp *codepipeline.CodePipeline, pipelinesToRun []int, pipelineMap map[int]string, executionTable *tablewriter.Table) error {
	fmt.Println("Running pipelines...")
	executionIds := make(map[string]string)

	for i := range pipelinesToRun {
		executionId, err := awsutil.RunPipeline(cp, pipelineMap[pipelinesToRun[i]])
		if err != nil {
			return err
		}
		executionIds[executionId] = pipelineMap[pipelinesToRun[i]]
	}

	renderExecutionTable(executionIds, executionTable)

	return nil
}
