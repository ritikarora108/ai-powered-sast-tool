package temporal

import (
	"context"
	"fmt"
	"time"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
)

// CloneActivityInput represents the input for the clone repository activity
type CloneActivityInput struct {
	RepositoryID string
	CloneURL     string
}

// CloneActivityOutput represents the output from the clone repository activity
type CloneActivityOutput struct {
	RepositoryID string
	RepoDir      string
}

// ScanActivityInput represents the input for the scan repository activity
type ScanActivityInput struct {
	RepositoryID   string
	RepoDir        string
	VulnTypes      []string
	FileExtensions []string
}

// ScanActivityOutput represents the output from the scan repository activity
type ScanActivityOutput struct {
	RepositoryID  string
	ScanID        string
	VulnCount     int
	ScanTimestamp time.Time
}

// CloneRepositoryActivity clones a GitHub repository to the local filesystem
func CloneRepositoryActivity(ctx context.Context, input CloneActivityInput) (*CloneActivityOutput, error) {
	gitHubService := services.NewGitHubService()
	repo := &services.Repository{
		ID:       input.RepositoryID,
		CloneURL: input.CloneURL,
	}

	// Create a temporary directory for the repository
	repoDir := fmt.Sprintf("/tmp/repos/%s", input.RepositoryID)

	err := gitHubService.CloneRepository(ctx, repo, repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return &CloneActivityOutput{
		RepositoryID: input.RepositoryID,
		RepoDir:      repoDir,
	}, nil
}

// ScanRepositoryActivity scans a repository for vulnerabilities
func ScanRepositoryActivity(ctx context.Context, input ScanActivityInput) (*ScanActivityOutput, error) {
	scannerService := services.NewScannerService(services.NewGitHubService())

	var vulnerabilityTypes []services.VulnerabilityType
	for _, vulnType := range input.VulnTypes {
		vulnerabilityTypes = append(vulnerabilityTypes, services.VulnerabilityType(vulnType))
	}

	scanOptions := &services.ScanOptions{
		VulnerabilityTypes: vulnerabilityTypes,
		FileExtensions:     input.FileExtensions,
	}
	scanResult, err := scannerService.ScanRepository(ctx, input.RepoDir, scanOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	return &ScanActivityOutput{
		RepositoryID:  input.RepositoryID,
		ScanID:        "placeholder-scan-id",
		VulnCount:     len(scanResult.Vulnerabilities),
		ScanTimestamp: time.Now(),
	}, nil
}
