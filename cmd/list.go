package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/shreyasrama/cph/cmd/awsutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
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

// List all pipelines TODO: return error, refactor
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
	// TODO: maybe make this a util function
	var pipeline_status []pipelineExecSummary
	for _, name := range pipeline_names {
		params := &codepipeline.ListPipelineExecutionsInput{
			MaxResults:   aws.Int64(1),
			PipelineName: aws.String(name),
		}
		result, err := cp.ListPipelineExecutions(params)
		if err != nil {
			fmt.Println("Error listing pipeline executions: ", err)
			return err
		}
		pipeline_status = append(pipeline_status, pipelineExecSummary{PipelineName: name, PipelineExecSummary: *result.PipelineExecutionSummaries[0]})
	}

	// Print output in readable format
	// TODO: replace with https://github.com/olekukonko/tablewriter to use colours better and fix formatting
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	fmt.Fprintln(w, "Name\tLatest State\tLast Update")
	for _, pipeline := range pipeline_status {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Println("Error loading timezone data: ", err)
			return err
		}
		date := pipeline.PipelineExecSummary.LastUpdateTime.In(loc).Format("Jan 02 2006 15:04:05")
		stage, err := awsutil.GetLastExecutedStage(cp, pipeline.PipelineName)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t",
			pipeline.PipelineName,
			getStatusColor(pipeline.PipelineExecSummary, stage),
			date,
		)
		fmt.Fprintln(w)
	}
	w.Flush()

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
