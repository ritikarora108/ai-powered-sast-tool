package temporal

import (
	"time"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ScanWorkflowInput represents the input for the scan workflow
// This struct contains all the information needed to start a repository scan
type ScanWorkflowInput struct {
	RepositoryID   string   // Unique identifier for the repository
	Owner          string   // GitHub repository owner (username or organization)
	Name           string   // GitHub repository name
	CloneURL       string   // URL to clone the repository (HTTPS or SSH)
	VulnTypes      []string // Types of vulnerabilities to scan for (e.g., "INJECTION", "XSS")
	FileExtensions []string // File extensions to include in the scan (e.g., ".go", ".js")
	NotifyEmail    bool     // Indicates whether email notification should be sent
	Email          string   // Store the submitter's email address
}

// ScanWorkflowOutput represents the output from the scan workflow
// This struct contains the results of the scan, including any vulnerabilities found
type ScanWorkflowOutput struct {
	RepositoryID    string                    // ID of the repository that was scanned
	ScanID          string                    // Unique identifier for this scan
	Status          string                    // Status of the scan (e.g., "completed", "failed")
	Message         string                    // Human-readable message about the scan result
	StartTime       time.Time                 // When the scan started
	EndTime         time.Time                 // When the scan completed
	Vulnerabilities []*services.Vulnerability // List of detected vulnerabilities
}

// ScanWorkflow orchestrates the repository scanning process
// This is the main workflow that coordinates the entire scanning process
// It follows these steps:
// 1. Clone the repository
// 2. Scan the repository for vulnerabilities
// 3. Return the scan results
func ScanWorkflow(ctx workflow.Context, input ScanWorkflowInput) (*ScanWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting repository scan workflow", "repository", input.Owner+"/"+input.Name)

	// Record workflow start time for tracking scan duration
	startTime := workflow.Now(ctx)

	// Step 1: Clone repository
	// This executes the CloneRepositoryActivity to download the repository code
	var cloneOutput CloneActivityOutput
	cloneCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Minute, // Allow up to 60 minutes for cloning (large repos may take time)
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3, // Retry up to 3 times if cloning fails
		},
	})

	// Execute the clone activity and wait for it to complete
	cloneErr := workflow.ExecuteActivity(cloneCtx, CloneRepositoryActivity, CloneActivityInput{
		RepositoryID: input.RepositoryID,
		CloneURL:     input.CloneURL,
	}).Get(ctx, &cloneOutput)

	// If cloning fails, return an error result
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
	// This executes the ScanRepositoryActivity to analyze the code for security issues
	var scanOutput ScanActivityOutput
	scanCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute, // Allow up to 30 minutes for scanning
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2, // Retry up to 2 times if scanning fails
		},
	})

	// Execute the scan activity and wait for it to complete
	scanErr := workflow.ExecuteActivity(scanCtx, ScanRepositoryActivity, ScanActivityInput{
		RepositoryID:   input.RepositoryID,
		RepoDir:        cloneOutput.RepoDir,
		VulnTypes:      input.VulnTypes,
		FileExtensions: input.FileExtensions,
		NotifyEmail:    input.NotifyEmail,
		Email:          input.Email,
	}).Get(ctx, &scanOutput)

	// If scanning fails, return an error result
	if scanErr != nil {
		return &ScanWorkflowOutput{
			RepositoryID: input.RepositoryID,
			Status:       "failed",
			Message:      "Failed to scan repository: " + scanErr.Error(),
			StartTime:    startTime,
			EndTime:      workflow.Now(ctx),
		}, scanErr
	}

	// Convert the vulnerabilities from the activity output to the workflow output format
	// This step ensures proper type conversion between internal types
	var vulnerabilities []*services.Vulnerability
	for _, v := range scanOutput.VulnerabilitiesFound {
		vuln := &services.Vulnerability{
			ID:          v.ID,
			Type:        services.VulnerabilityType(v.Type),
			FilePath:    v.FilePath,
			LineStart:   v.LineStart,
			LineEnd:     v.LineEnd,
			Severity:    v.Severity,
			Description: v.Description,
			Remediation: v.Remediation,
			Code:        v.Code,
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	// Register query handler to expose results
	// This allows external systems to query the current status of the workflow
	workflow.SetQueryHandler(ctx, "scan_result", func() (*ScanWorkflowOutput, error) {
		return &ScanWorkflowOutput{
			RepositoryID:    input.RepositoryID,
			ScanID:          scanOutput.ScanID,
			Status:          "completed",
			Message:         "Scan completed successfully",
			StartTime:       startTime,
			EndTime:         workflow.Now(ctx),
			Vulnerabilities: vulnerabilities,
		}, nil
	})

	// Successfully completed - return the final scan results
	return &ScanWorkflowOutput{
		RepositoryID:    input.RepositoryID,
		ScanID:          scanOutput.ScanID,
		Status:          "completed",
		Message:         "Scan completed successfully",
		StartTime:       startTime,
		EndTime:         workflow.Now(ctx),
		Vulnerabilities: vulnerabilities,
	}, nil
}
