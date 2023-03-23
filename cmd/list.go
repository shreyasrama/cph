package cmd

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/shreyasrama/cph/pkg/awsutil"
	"github.com/shreyasrama/cph/pkg/helpers"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS CodePipelines you have access to.",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		if name != "" {
			listPipelines(name)
		} else {
			listPipelines("")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	listCmd.PersistentFlags().String("name", "", "Use a name or part of a name to filter the listed pipelines.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type pipelineExecSummary struct {
	PipelineName        string
	PipelineExecSummary codepipeline.PipelineExecutionSummary
}

// List all pipelines
// Makes the following calls to CodePipeline:
// 1. ListPipelines
// 2. ListPipelineExecutions
// 3. GetPipelineState
func listPipelines(searchTerm string) error {
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
	var pipeline_status []pipelineExecSummary
	for _, name := range pipeline_names {
		latestExecution, err := awsutil.GetLatestPipelineExecution(cp, name)
		if err != nil {
			return err
		}
		pipeline_status = append(pipeline_status, pipelineExecSummary{PipelineName: name, PipelineExecSummary: latestExecution})
	}

	// Print output in readable format
	table := helpers.SetupTable([]string{"Name", "Latest State", "Last Update", "Revision"})

	for _, pipeline := range pipeline_status {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Println("Error loading timezone data: ", err)
			return err
		}
		date := pipeline.PipelineExecSummary.LastUpdateTime.In(loc).Format("Jan 02 2006 15:04:05")
		stageInfo, err := awsutil.GetLastExecutedStage(cp, pipeline.PipelineName)
		if err != nil {
			return err
		}

		table.Append([]string{
			pipeline.PipelineName,
			getStatusColor(pipeline.PipelineExecSummary, stageInfo.StageName),
			date,
			*pipeline.PipelineExecSummary.SourceRevisions[0].RevisionSummary,
		})
	}

	table.Render()

	return nil
}

func getStatusColor(pes codepipeline.PipelineExecutionSummary, stage string) string {
	switch *pes.Status {
	case "InProgress":
		blue := color.New(color.FgBlue).SprintFunc()
		return blue(*pes.Status, " - ", stage)
	case "Failed", "Stopped", "Cancelled":
		red := color.New(color.FgRed).SprintFunc()
		return red(*pes.Status, " - ", stage)
	case "Stopping":
		yellow := color.New(color.FgYellow).SprintFunc()
		return yellow(*pes.Status, " - ", stage)
	case "Succeeded":
		green := color.New(color.FgGreen).SprintFunc()
		return green(*pes.Status, " - ", stage)
	case "Superseded":
		black := color.New(color.FgBlack).SprintFunc()
		return black(*pes.Status, " - ", stage)
	default:
		return *pes.Status
	}
}
