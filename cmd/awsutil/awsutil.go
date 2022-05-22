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

type pipelineExecSummary struct {
	PipelineName        string
	PipelineExecSummary codepipeline.PipelineExecutionSummary
}

type StageInfo struct {
	ActionName string
	StageName  string
	Status     string
	Token      *string
}

// Create an AWS Session with a Code Pipeline client
func CreateCodePipelineSession() (*codepipeline.CodePipeline, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Must be set to enable
		Profile:           os.Getenv("AWS_PROFILE"),
	})
	if err != nil {
		fmt.Println("Error creating CodePipeline session:", err)
		return nil, err
	}

	return codepipeline.New(sess), nil
}

// Given a serch term, return a slice of pipeline names
func GetPipelineNames(client *codepipeline.CodePipeline, searchTerm string) ([]string, error) {
	// List all pipelines
	params := &codepipeline.ListPipelinesInput{
		MaxResults: aws.Int64(300),
	}
	result, err := client.ListPipelines(params)
	if err != nil {
		fmt.Println("Error listing pipelines: ", err)
		return nil, err
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

	return pipeline_names, nil
}

// Given a pipeline name, run that pipeline
func RunPipeline(client *codepipeline.CodePipeline, pipelineName string) (string, error) {
	// Start pipeline execution
	params := &codepipeline.StartPipelineExecutionInput{
		Name: aws.String(pipelineName),
	}
	result, err := client.StartPipelineExecution(params)
	if err != nil {
		fmt.Println("Error starting pipeline execution: ", err)
		return "", err
	}

	return *result.PipelineExecutionId, nil
}

// Given pipeline names, run those pipelines
func RunPipelines(client *codepipeline.CodePipeline, pipelineNames []string) (map[string]string, error) {
	// Start pipeline execution
	m := make(map[string]string)
	for _, p := range pipelineNames {
		executionId, err := RunPipeline(client, p)
		if err != nil {
			return nil, err
		}
		m[executionId] = p
	}

	return m, nil
}

// Given a pipeline name, return the stage that was last executed
func GetLastExecutedStage(client *codepipeline.CodePipeline, pipelineName string) (StageInfo, error) {
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
	var stageInfo StageInfo
	lastStatusChange := time.Date(1970, time.Month(1), 1, 1, 1, 1, 1, time.UTC)
	for _, p := range result.StageStates {
		if p.ActionStates[0].LatestExecution != nil {
			switch *p.ActionStates[0].LatestExecution.Status {
			case "InProgress", "Failed":
				return StageInfo{
					*p.ActionStates[0].ActionName,
					*p.StageName,
					*p.ActionStates[0].LatestExecution.Status,
					p.ActionStates[0].LatestExecution.Token,
				}, nil
			default:
				currentStatusChange := *p.ActionStates[0].LatestExecution.LastStatusChange
				if currentStatusChange.After(lastStatusChange) {
					lastStatusChange = currentStatusChange
					stageInfo = StageInfo{
						*p.ActionStates[0].ActionName,
						*p.StageName,
						*p.ActionStates[0].LatestExecution.Status,
						p.ActionStates[0].LatestExecution.Token,
					}
				}
			}
		}
	}

	return stageInfo, nil
}

func GetLatestPipelineExecution(client *codepipeline.CodePipeline, pipelineName string) (codepipeline.PipelineExecutionSummary, error) {
	// Get one (the latest) pipeline execution
	params := &codepipeline.ListPipelineExecutionsInput{
		MaxResults:   aws.Int64(1),
		PipelineName: aws.String(pipelineName),
	}
	result, err := client.ListPipelineExecutions(params)
	if err != nil {
		fmt.Println("Error listing pipeline executions: ", err)
		return codepipeline.PipelineExecutionSummary{}, err
	}

	return *result.PipelineExecutionSummaries[0], nil
}

func ApprovePipelines(client *codepipeline.CodePipeline, stagesToApprove map[string]StageInfo) error {
	for name, info := range stagesToApprove {
		_, err := client.PutApprovalResult(&codepipeline.PutApprovalResultInput{
			ActionName: &info.ActionName,
			PipelineName: &name,
			Result: &codepipeline.ApprovalResult{
				Status: aws.String(codepipeline.ApprovalStatusApproved),
				Summary: aws.String("Approved"),
			},
			StageName: &info.StageName,
			Token: info.Token,
		})

		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
