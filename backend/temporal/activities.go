package temporal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.uber.org/zap"
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
	RepositoryID         string
	ScanID               string
	VulnCount            int
	VulnerabilitiesFound []services.Vulnerability
	ScanTimestamp        time.Time
}

// CloneRepositoryActivity clones a GitHub repository to the local filesystem
func CloneRepositoryActivity(ctx context.Context, input CloneActivityInput) (*CloneActivityOutput, error) {
	log := logger.Get()
	log.Info("Starting clone repository activity", zap.String("repo_id", input.RepositoryID))

	gitHubService := services.NewGitHubService(db.NewQueries())
	repo := &services.Repository{
		ID:       input.RepositoryID,
		CloneURL: input.CloneURL,
	}

	// Create a temporary directory for the repository
	tmpDir := os.TempDir()
	repoDir := fmt.Sprintf("%s/repos/%s", tmpDir, input.RepositoryID)

	// Check if the repository directory already exists
	if _, err := os.Stat(repoDir); err == nil {
		log.Info("Repository directory already exists, removing it before cloning",
			zap.String("repo_dir", repoDir))
		// Remove the existing directory
		if err := os.RemoveAll(repoDir); err != nil {
			log.Error("Failed to remove existing repository directory",
				zap.String("repo_dir", repoDir),
				zap.Error(err))
			return nil, fmt.Errorf("failed to remove existing repository directory: %w", err)
		}
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		log.Error("Failed to create repository directory",
			zap.String("repo_dir", repoDir),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}

	log.Info("Cloning repository",
		zap.String("repo_id", input.RepositoryID),
		zap.String("clone_url", input.CloneURL),
		zap.String("repo_dir", repoDir))

	err := gitHubService.CloneRepository(ctx, repo, repoDir)
	if err != nil {
		log.Error("Failed to clone repository",
			zap.String("repo_id", input.RepositoryID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Info("Repository cloned successfully", zap.String("repo_dir", repoDir))

	return &CloneActivityOutput{
		RepositoryID: input.RepositoryID,
		RepoDir:      repoDir,
	}, nil
}

// ScanRepositoryActivity scans a repository for vulnerabilities
func ScanRepositoryActivity(ctx context.Context, input ScanActivityInput) (*ScanActivityOutput, error) {
	log := logger.Get()
	log.Info("Starting repository scan activity",
		zap.String("repo_id", input.RepositoryID),
		zap.String("repo_dir", input.RepoDir))

	// Create instances of required services
	dbQueries := db.NewQueries()
	githubService := services.NewGitHubService(dbQueries)
	scannerService := services.NewScannerService(githubService)

	// Generate a scan ID
	scanID := uuid.New().String()

	// Try to create a scan record in the database
	sqlDB := dbQueries.GetDB()
	if sqlDB != nil {
		// Create a scan record
		_, err := sqlDB.ExecContext(ctx,
			`INSERT INTO scans (id, repository_id, status) VALUES ($1, $2, $3)`,
			scanID, input.RepositoryID, "in_progress")
		if err != nil {
			log.Warn("Failed to create scan record in database",
				zap.String("scan_id", scanID),
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
		} else {
			log.Info("Created scan record in database",
				zap.String("scan_id", scanID),
				zap.String("repo_id", input.RepositoryID))
		}
	}

	// Convert string vulnerability types to the enum type
	var vulnerabilityTypes []services.VulnerabilityType
	for _, vulnType := range input.VulnTypes {
		vulnerabilityTypes = append(vulnerabilityTypes, services.VulnerabilityType(vulnType))
	}

	// Configure scan options
	scanOptions := &services.ScanOptions{
		VulnerabilityTypes: vulnerabilityTypes,
		FileExtensions:     input.FileExtensions,
		MaxFiles:           100, // Limit the number of files to scan
	}

	log.Info("Starting code scan",
		zap.String("scan_id", scanID),
		zap.Strings("vuln_types", input.VulnTypes),
		zap.Strings("file_extensions", input.FileExtensions))

	// Perform the scan
	scanResult, err := scannerService.ScanRepository(ctx, input.RepoDir, scanOptions)
	if err != nil {
		log.Error("Failed to scan repository",
			zap.String("repo_id", input.RepositoryID),
			zap.Error(err))

		// Update scan status to failed
		if sqlDB != nil {
			_, updateErr := sqlDB.ExecContext(ctx,
				`UPDATE scans SET status = $1, error_message = $2, completed_at = NOW() WHERE id = $3`,
				"failed", err.Error(), scanID)
			if updateErr != nil {
				log.Error("Failed to update scan status",
					zap.String("scan_id", scanID),
					zap.Error(updateErr))
			}
		}

		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	// Store the vulnerabilities in the database
	var vulnList []services.Vulnerability
	for _, v := range scanResult.Vulnerabilities {
		vulnList = append(vulnList, services.Vulnerability{
			ID:          v.ID,
			Type:        v.Type,
			FilePath:    v.FilePath,
			LineStart:   v.LineStart,
			LineEnd:     v.LineEnd,
			Severity:    v.Severity,
			Description: v.Description,
			Remediation: v.Remediation,
			Code:        v.Code,
		})

		// Store each vulnerability in the database
		if sqlDB != nil {
			_, err = sqlDB.ExecContext(ctx,
				`INSERT INTO vulnerabilities 
				(id, scan_id, vulnerability_type, file_path, line_start, line_end, severity, description, remediation, code_snippet) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				v.ID, scanID, string(v.Type), v.FilePath, v.LineStart, v.LineEnd,
				v.Severity, v.Description, v.Remediation, v.Code)
			if err != nil {
				log.Warn("Failed to store vulnerability",
					zap.String("scan_id", scanID),
					zap.String("vuln_id", v.ID),
					zap.Error(err))
			}
		}
	}

	// Update scan status to completed
	if sqlDB != nil {
		_, err = sqlDB.ExecContext(ctx,
			`UPDATE scans SET status = $1, completed_at = NOW() WHERE id = $2`,
			"completed", scanID)
		if err != nil {
			log.Error("Failed to update scan status",
				zap.String("scan_id", scanID),
				zap.Error(err))
		}
	}

	log.Info("Repository scan completed and data stored",
		zap.String("scan_id", scanID),
		zap.Int("vulnerability_count", len(scanResult.Vulnerabilities)))

	return &ScanActivityOutput{
		RepositoryID:         input.RepositoryID,
		ScanID:               scanID,
		VulnCount:            len(scanResult.Vulnerabilities),
		VulnerabilitiesFound: vulnList,
		ScanTimestamp:        time.Now(),
	}, nil
}
