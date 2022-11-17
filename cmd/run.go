package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shreyasrama/cph/pkg/awsutil"
	"github.com/shreyasrama/cph/pkg/tablewriter"

	"github.com/spf13/cobra"
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

func runPipelines(searchTerm string) error {
	cp, err := awsutil.CreateCodePipelineSession()
	if err != nil {
		return err
	}

	pipeline_names, err := awsutil.GetPipelineNames(cp, searchTerm)
	if err != nil {
		return err
	}

	// Print and confirm pipelines to be run
	pipeline_map := make(map[int]string)
	fmt.Printf("\n%s\n", "The following pipelines have been found:")
	for i, pipeline := range pipeline_names {
		pipeline_map[i+1] = pipeline
		fmt.Printf("    [%v] %s\n", i+1, pipeline)
	}
	var s string
	fmt.Printf(
		"\n%s",
		`Do you want to run these pipelines?
Enter 'yes' to run all, 'no' to cancel, or a number for a specific pipeline: `)
	fmt.Scan(&s)
	// Check if number is entered
	if i, err := strconv.Atoi(s); err == nil {
		executionId, err := awsutil.RunPipeline(cp, pipeline_map[i])
		if err != nil {
			return err
		}
		fmt.Printf("Started execution of %s. Execution ID: %s", pipeline_map[i], executionId)
	} else if strings.EqualFold(s, "yes") {
		fmt.Println("Running pipelines...")
		executionIds, err := awsutil.RunPipelines(cp, pipeline_names)
		if err != nil {
			return err
		}

		table := tablewriter.SetupTable([]string{"Pipeline", "Execution ID"})
		for id, name := range executionIds {
			table.Append([]string{
				name,
				id,
			})
		}
		table.Render()

	} else if strings.EqualFold(s, "no") {
		fmt.Println("Cancelled.")
		// exit?
	} else {
		// User provided a range
		match, err := regexp.MatchString("^[0-9]{1,2}-[0-9]{1,2}$", s)
		if err != nil {
			return err
		}
		if match {
			inputhelper.
		}
		fmt.Println("Input not recognised.")
		// exit?
	}

	return nil
}
