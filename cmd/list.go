package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List AWS CodePipelines you have access to",
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
	listCmd.PersistentFlags().String("name", "", "Use a name or part of a name to filter the results.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type pipelineExecSummary struct {
	PipelineName        string
	PipelineExecSummary codepipeline.PipelineExecutionSummary
}

// List all pipelines TODO: return error
func listPipelines(searchTerm string) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Must be set to enable
		Profile:           os.Getenv("AWS_PROFILE"),
	})
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	cp := codepipeline.New(sess)

	// List all pipelines
	params := &codepipeline.ListPipelinesInput{
		MaxResults: aws.Int64(300),
	}
	result, err := cp.ListPipelines(params)
	if err != nil {
		fmt.Println("Error listing pipelines: ", err)
		os.Exit(1)
	}

	// Iterate over pipelines and create a slice of names
	var pipeline_names []string
	for _, p := range result.Pipelines {
		if searchTerm != "" {
			if strings.Contains(*p.Name, searchTerm) {
				pipeline_names = append(pipeline_names, *p.Name)
			}
		} else {
			pipeline_names = append(pipeline_names, *p.Name)
		}
	}

	// Iterate over pipeline names and get the most recent pipeline
	// execution status and create a slice of structs
	var pipeline_status []pipelineExecSummary
	for _, name := range pipeline_names {
		params := &codepipeline.ListPipelineExecutionsInput{
			MaxResults:   aws.Int64(1),
			PipelineName: aws.String(name),
		}
		result, err := cp.ListPipelineExecutions(params)
		if err != nil {
			fmt.Println("Error listing pipeline executions: ", err)
			os.Exit(1)
		}
		pipeline_status = append(pipeline_status, pipelineExecSummary{PipelineName: name, PipelineExecSummary: *result.PipelineExecutionSummaries[0]})
	}

	// Print output in readable format
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()
	for _, pipeline := range pipeline_status {
		loc, err := time.LoadLocation("Local")
		if err != nil {
			fmt.Println("Error loading timezone data: ", err)
			os.Exit(1)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			pipeline.PipelineName,
			getStatusColor(pipeline.PipelineExecSummary),
			pipeline.PipelineExecSummary.LastUpdateTime.In(loc).Format("Jan 02 2006 15:04:05"),
		)
	}
}

func getStatusColor(pes codepipeline.PipelineExecutionSummary) string {
	switch *pes.Status {
	case "InProgress":
		blue := color.New(color.FgBlue).SprintFunc()
		return blue(*pes.Status)
	case "Failed", "Stopped":
		red := color.New(color.FgRed).SprintFunc()
		return red(*pes.Status)
	case "Stopping":
		yellow := color.New(color.FgYellow).SprintFunc()
		return yellow(*pes.Status)
	case "Succeeded":
		green := color.New(color.FgGreen).SprintFunc()
		return green(*pes.Status)
	case "Superseded":
		black := color.New(color.FgBlack).SprintFunc()
		return black(*pes.Status)
	default:
		return *pes.Status
	}
}
