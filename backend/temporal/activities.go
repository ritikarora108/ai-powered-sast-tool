package temporal

import (
	"context"
	"database/sql"
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

	// Check if database is available
	dbQueries := db.NewQueries()
	gitHubService := services.NewGitHubService(dbQueries)

	// Continue with cloning even if database is unavailable
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

	// If database connection is available, try to create a scan record
	var databaseAvailable bool = false
	if sqlDB != nil {
		databaseAvailable = true

		// First, let's get the repository's created_by field if possible
		var createdBy sql.NullString
		err := sqlDB.QueryRowContext(ctx,
			`SELECT created_by FROM repositories WHERE id = $1`,
			input.RepositoryID).Scan(&createdBy)

		if err != nil && err != sql.ErrNoRows {
			log.Warn("Failed to get repository creator information",
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			// Continue anyway, the created_by will remain NULL
		}

		// Create a scan record
		_, err = sqlDB.ExecContext(ctx,
			`INSERT INTO scans (id, repository_id, status, started_at, created_by, error_message) 
			VALUES ($1, $2, $3, NOW(), $4, $5)`,
			scanID, input.RepositoryID, "in_progress", createdBy, "")
		if err != nil {
			log.Error("Failed to create scan record in database",
				zap.String("scan_id", scanID),
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			// Continue with the scan but note that we won't be able to store results
			databaseAvailable = false
		} else {
			log.Info("Created scan record in database",
				zap.String("scan_id", scanID),
				zap.String("repo_id", input.RepositoryID))
		}
	} else {
		log.Warn("Database connection is not available, proceeding without saving to database",
			zap.String("repo_id", input.RepositoryID))
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

		// Update scan status to failed if database is available
		if databaseAvailable && sqlDB != nil {
			errMsg := err.Error()
			if errMsg == "" {
				errMsg = "Unknown scan error occurred"
			}

			_, updateErr := sqlDB.ExecContext(ctx,
				`UPDATE scans SET status = $1, error_message = $2, completed_at = NOW() WHERE id = $3`,
				"failed", errMsg, scanID)
			if updateErr != nil {
				log.Error("Failed to update scan status",
					zap.String("scan_id", scanID),
					zap.Error(updateErr))
			}
		}

		return nil, fmt.Errorf("failed to scan repository: %w", err)
	}

	// Store the vulnerabilities in the database if available
	var vulnList []services.Vulnerability
	var dbErrors []error

	if databaseAvailable && sqlDB != nil && scanResult != nil && len(scanResult.Vulnerabilities) > 0 {
		log.Info("Storing vulnerability findings in database",
			zap.String("scan_id", scanID),
			zap.Int("vuln_count", len(scanResult.Vulnerabilities)))

		// Prepare a transaction for bulk inserts
		tx, err := sqlDB.BeginTx(ctx, nil)
		if err != nil {
			log.Error("Failed to begin transaction for storing vulnerabilities",
				zap.String("scan_id", scanID),
				zap.Error(err))
		} else {
			// Insert vulnerabilities within the transaction
			for _, vuln := range scanResult.Vulnerabilities {
				vulnID := uuid.New().String()
				_, err := tx.ExecContext(ctx,
					`INSERT INTO vulnerabilities (
						id, scan_id, vulnerability_type, file_path, 
						line_start, line_end, severity, description, 
						remediation, code_snippet, created_at, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
					vulnID, scanID, string(vuln.Type), vuln.FilePath,
					vuln.LineStart, vuln.LineEnd, vuln.Severity, vuln.Description,
					vuln.Remediation, vuln.Code)

				if err != nil {
					dbErrors = append(dbErrors, err)
					log.Error("Failed to insert vulnerability",
						zap.String("scan_id", scanID),
						zap.String("vuln_type", string(vuln.Type)),
						zap.Error(err))
					continue
				}

				// Add to the list of vulnerabilities to return
				vulnWithID := services.Vulnerability{
					ID:          vulnID,
					Type:        vuln.Type,
					FilePath:    vuln.FilePath,
					LineStart:   vuln.LineStart,
					LineEnd:     vuln.LineEnd,
					Severity:    vuln.Severity,
					Description: vuln.Description,
					Remediation: vuln.Remediation,
					Code:        vuln.Code,
				}
				vulnList = append(vulnList, vulnWithID)
			}

			// Commit the transaction
			if err := tx.Commit(); err != nil {
				log.Error("Failed to commit transaction for storing vulnerabilities",
					zap.String("scan_id", scanID),
					zap.Error(err))
			} else {
				log.Info("Successfully stored vulnerabilities in database",
					zap.String("scan_id", scanID),
					zap.Int("vuln_count", len(vulnList)))
			}
		}
	} else if scanResult != nil {
		// Database unavailable, but we still have scan results, so include them in the output
		log.Info("Database unavailable for storing vulnerabilities, returning only in memory",
			zap.String("scan_id", scanID),
			zap.Int("vuln_count", len(scanResult.Vulnerabilities)))

		// Still include the vulnerabilities in the output
		for _, vuln := range scanResult.Vulnerabilities {
			vulnWithID := services.Vulnerability{
				ID:          uuid.New().String(), // Generate IDs even if not in DB
				Type:        vuln.Type,
				FilePath:    vuln.FilePath,
				LineStart:   vuln.LineStart,
				LineEnd:     vuln.LineEnd,
				Severity:    vuln.Severity,
				Description: vuln.Description,
				Remediation: vuln.Remediation,
				Code:        vuln.Code,
			}
			vulnList = append(vulnList, vulnWithID)
		}
	}

	// Update scan status to completed
	if databaseAvailable && sqlDB != nil {
		_, err = sqlDB.ExecContext(ctx,
			`UPDATE scans SET status = $1, completed_at = NOW(), results_available = true WHERE id = $2`,
			"completed", scanID)
		if err != nil {
			log.Error("Failed to update scan status",
				zap.String("scan_id", scanID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update scan status: %w", err)
		}

		log.Info("Updated scan status to completed and set results_available flag",
			zap.String("scan_id", scanID))
	}

	// Update repository with last scan time and status
	if databaseAvailable && sqlDB != nil {
		_, err = sqlDB.ExecContext(ctx,
			`UPDATE repositories SET last_scan_at = NOW(), status = $1 WHERE id = $2`,
			"completed", input.RepositoryID)
		if err != nil {
			log.Error("Failed to update repository scan info",
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			// Continue anyway since the scan itself is completed
		} else {
			log.Info("Updated repository with scan info",
				zap.String("repo_id", input.RepositoryID),
				zap.String("status", "completed"))
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
