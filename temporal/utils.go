package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// WorkflowStatus represents the status of a workflow.
type WorkflowStatus struct {
	WorkflowID string
	RunID      string
	Status     enums.WorkflowExecutionStatus
	StartTime  int64
	CloseTime  int64
}

// ListAllWorkflows returns all workflows matching the given query.
// The query uses Temporal's visibility query syntax.
// Example: "WorkflowType='MyWorkflow' AND ExecutionStatus='Running'"
func ListAllWorkflows(ctx context.Context, c client.Client, namespace, query string) ([]WorkflowStatus, error) {
	var workflows []WorkflowStatus
	var nextPageToken []byte

	for {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     namespace,
			PageSize:      100,
			Query:         query,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list workflows: %w", err)
		}

		for _, exec := range resp.Executions {
			status := WorkflowStatus{
				WorkflowID: exec.Execution.WorkflowId,
				RunID:      exec.Execution.RunId,
				Status:     exec.Status,
			}
			if exec.StartTime != nil {
				status.StartTime = exec.StartTime.AsTime().Unix()
			}
			if exec.CloseTime != nil {
				status.CloseTime = exec.CloseTime.AsTime().Unix()
			}
			workflows = append(workflows, status)
		}

		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 {
			break
		}
	}

	return workflows, nil
}

// GetWorkflowStatus returns the status of a specific workflow.
func GetWorkflowStatus(ctx context.Context, c client.Client, workflowID, runID string) (*WorkflowStatus, error) {
	desc, err := c.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to describe workflow: %w", err)
	}

	status := &WorkflowStatus{
		WorkflowID: desc.WorkflowExecutionInfo.Execution.WorkflowId,
		RunID:      desc.WorkflowExecutionInfo.Execution.RunId,
		Status:     desc.WorkflowExecutionInfo.Status,
	}

	if desc.WorkflowExecutionInfo.StartTime != nil {
		status.StartTime = desc.WorkflowExecutionInfo.StartTime.AsTime().Unix()
	}
	if desc.WorkflowExecutionInfo.CloseTime != nil {
		status.CloseTime = desc.WorkflowExecutionInfo.CloseTime.AsTime().Unix()
	}

	return status, nil
}

// IsWorkflowRunning returns true if the workflow is still running.
func IsWorkflowRunning(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_RUNNING
}

// IsWorkflowCompleted returns true if the workflow completed successfully.
func IsWorkflowCompleted(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
}

// IsWorkflowFailed returns true if the workflow failed.
func IsWorkflowFailed(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_FAILED
}

// IsWorkflowCanceled returns true if the workflow was canceled.
func IsWorkflowCanceled(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_CANCELED
}

// IsWorkflowTerminated returns true if the workflow was terminated.
func IsWorkflowTerminated(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_TERMINATED
}

// IsWorkflowTimedOut returns true if the workflow timed out.
func IsWorkflowTimedOut(status enums.WorkflowExecutionStatus) bool {
	return status == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT
}
