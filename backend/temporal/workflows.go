package temporal

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// ScanWorkflowInput represents the input for the scan workflow
type ScanWorkflowInput struct {
	RepositoryID   string
	Owner          string
	Name           string
	CloneURL       string
	VulnTypes      []string
	FileExtensions []string
}

// ScanWorkflowOutput represents the output from the scan workflow
type ScanWorkflowOutput struct {
	RepositoryID string
	ScanID       string
	Status       string
	Message      string
	StartTime    time.Time
	EndTime      time.Time
}

// ScanWorkflow orchestrates the repository scanning process
func ScanWorkflow(ctx workflow.Context, input ScanWorkflowInput) (*ScanWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting repository scan workflow", "repository", input.Owner+"/"+input.Name)

	// Record workflow start time
	startTime := workflow.Now(ctx)

	// Step 1: Clone repository
	var cloneOutput CloneActivityOutput
	cloneCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	cloneErr := workflow.ExecuteActivity(cloneCtx, CloneRepositoryActivity, CloneActivityInput{
		RepositoryID: input.RepositoryID,
		CloneURL:     input.CloneURL,
	}).Get(ctx, &cloneOutput)

	if cloneErr != nil {
		return &ScanWorkflowOutput{
			RepositoryID: input.RepositoryID,
			Status:       "failed",
			Message:      "Failed to clone repository: " + cloneErr.Error(),
			StartTime:    startTime,
			EndTime:      workflow.Now(ctx),
		}, cloneErr
	}

	// Step 2: Scan repository for vulnerabilities
	var scanOutput ScanActivityOutput
	scanCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			MaximumAttempts: 2,
		},
	})

	scanErr := workflow.ExecuteActivity(scanCtx, ScanRepositoryActivity, ScanActivityInput{
		RepositoryID:   input.RepositoryID,
		RepoDir:        cloneOutput.RepoDir,
		VulnTypes:      input.VulnTypes,
		FileExtensions: input.FileExtensions,
	}).Get(ctx, &scanOutput)

	if scanErr != nil {
		return &ScanWorkflowOutput{
			RepositoryID: input.RepositoryID,
			Status:       "failed",
			Message:      "Failed to scan repository: " + scanErr.Error(),
			StartTime:    startTime,
			EndTime:      workflow.Now(ctx),
		}, scanErr
	}

	// Successfully completed
	return &ScanWorkflowOutput{
		RepositoryID: input.RepositoryID,
		ScanID:       scanOutput.ScanID,
		Status:       "completed",
		Message:      "Scan completed successfully",
		StartTime:    startTime,
		EndTime:      workflow.Now(ctx),
	}, nil
}
