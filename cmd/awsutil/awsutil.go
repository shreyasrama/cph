package awsutil

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

// Create an AWS Session with a Code Pipeline client
func CreateCodePipelineSession() *codepipeline.CodePipeline {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Must be set to enable
		Profile:           os.Getenv("AWS_PROFILE"),
	})
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	return codepipeline.New(sess)
}

// Given a serch term, return a slice of pipeline names
func GetPipelineNames(client *codepipeline.CodePipeline, searchTerm string) []string {
	// List all pipelines
	params := &codepipeline.ListPipelinesInput{
		MaxResults: aws.Int64(300),
	}
	result, err := client.ListPipelines(params)
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

	return pipeline_names
}

func RunPipeline(client *codepipeline.CodePipeline, pipelineName string) string {
	return ""
}

// Given a pipeline name, return the stage that was last executed
func GetLastExecutedStage(client *codepipeline.CodePipeline, pipelineName string) string {
	// Get the pipeline state
	params := &codepipeline.GetPipelineStateInput{
		Name: aws.String(pipelineName),
	}
	result, err := client.GetPipelineState(params)
	if err != nil {
		fmt.Println("Error retrieving pipeline state: ", err)
	}

	// Iterate over pipeline stage states.
	// InProgress or Failed means that the pipeline is currently at that given stage.
	var stage string
	lastStatusChange := time.Date(1970, time.Month(1), 1, 1, 1, 1, 1, time.UTC)
	for _, p := range result.StageStates {
		if *p.ActionStates[0].LatestExecution.Status == "InProgress" {
			return *p.StageName
		} else if *p.ActionStates[0].LatestExecution.Status == "Failed" {
			return *p.StageName
		} else {
			currentStatusChange := *p.ActionStates[0].LatestExecution.LastStatusChange
			if currentStatusChange.After(lastStatusChange) {
				lastStatusChange = currentStatusChange
				stage = *p.StageName
			}
		}
	}

	return stage
}
