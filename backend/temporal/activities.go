package temporal

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.uber.org/zap"
)

// CloneActivityInput represents the input for the clone repository activity
// It contains the required information to clone a Git repository
type CloneActivityInput struct {
	RepositoryID string // Unique identifier for the repository
	CloneURL     string // Git URL to clone the repository (HTTPS or SSH)
}

// CloneActivityOutput represents the output from the clone repository activity
// It provides information about the cloned repository, including the local directory
type CloneActivityOutput struct {
	RepositoryID string // Repository identifier (for correlation)
	RepoDir      string // Local file system path where the repository was cloned
}

// ScanActivityInput represents the input for the scan repository activity
// It contains all parameters required to perform a security scan on the cloned repo
type ScanActivityInput struct {
	RepositoryID   string   // Unique identifier for the repository
	RepoDir        string   // Directory path where the repository was cloned
	VulnTypes      []string // Types of vulnerabilities to scan for
	FileExtensions []string // File extensions to include in the scan
	NotifyEmail    bool     // Whether to send an email notification when scan completes
	Email          string   // Email address to notify when scan completes
}

// ScanActivityOutput represents the output from the scan repository activity
// It contains the results of the security scan, including detected vulnerabilities
type ScanActivityOutput struct {
	RepositoryID         string                   // Repository identifier (for correlation)
	ScanID               string                   // Unique identifier for this scan
	VulnCount            int                      // Total count of vulnerabilities found
	VulnerabilitiesFound []services.Vulnerability // List of detected vulnerabilities
	ScanTimestamp        time.Time                // When the scan was performed
}

// CloneRepositoryActivity clones a GitHub repository to the local filesystem
// This activity is responsible for downloading the source code from Git repositories
// It handles both public and private repositories, using authentication when needed
func CloneRepositoryActivity(ctx context.Context, input CloneActivityInput) (*CloneActivityOutput, error) {
	log := logger.Get()
	log.Info("Starting clone repository activity", zap.String("repo_id", input.RepositoryID))

	// Check if database is available and initialize services
	dbQueries := db.NewQueries()
	gitHubService := services.NewGitHubService(dbQueries)

	// Create a repository object for the clone operation
	repo := &services.Repository{
		ID:       input.RepositoryID,
		CloneURL: input.CloneURL,
	}

	// Create a temporary directory for the repository
	// The repository will be cloned into a unique subdirectory
	tmpDir := os.TempDir()
	repoDir := fmt.Sprintf("%s/repos/%s", tmpDir, input.RepositoryID)

	// Check if the repository directory already exists
	// If it does, remove it to ensure a clean clone
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
	// This ensures the parent directories exist for the clone operation
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

	// First try without authentication (for public repos)
	// This will succeed for public repositories without requiring credentials
	err := gitHubService.CloneRepository(ctx, repo, repoDir)
	if err != nil {
		// If we get an authentication error, check if GITHUB_TOKEN is set
		// This handles private repositories that require authentication
		if strings.Contains(err.Error(), "authentication required") || strings.Contains(err.Error(), "Invalid username or password") {
			log.Info("Authentication required, checking for GitHub token")
			githubToken := os.Getenv("GITHUB_TOKEN")

			if githubToken == "" {
				log.Warn("Repository requires authentication but GITHUB_TOKEN is not set")
				return nil, fmt.Errorf("repository requires authentication but GITHUB_TOKEN environment variable is not set")
			}

			// For GitHub URLs, construct an authenticated URL with the token
			// This modifies the URL to include the access token for authentication
			if strings.HasPrefix(input.CloneURL, "https://github.com") {
				authenticatedURL := strings.Replace(input.CloneURL, "https://github.com", fmt.Sprintf("https://%s@github.com", githubToken), 1)
				log.Info("Retrying with authenticated URL")

				// Create a new repo object with the authenticated URL
				authRepo := &services.Repository{
					ID:       input.RepositoryID,
					CloneURL: authenticatedURL,
				}

				// Try cloning again with authentication
				err = gitHubService.CloneRepository(ctx, authRepo, repoDir)
				if err != nil {
					log.Error("Failed to clone repository with authentication",
						zap.String("repo_id", input.RepositoryID),
						zap.Error(err))
					return nil, fmt.Errorf("failed to clone repository with authentication: %w", err)
				}

				log.Info("Repository cloned successfully with authenticated URL")
			} else {
				return nil, fmt.Errorf("repository requires authentication but URL format is not supported for authentication")
			}
		} else {
			log.Error("Failed to clone repository",
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		log.Info("Repository cloned successfully without authentication")
	}

	log.Info("Repository cloned successfully", zap.String("repo_dir", repoDir))

	// Return the output with the repository directory where the code was cloned
	return &CloneActivityOutput{
		RepositoryID: input.RepositoryID,
		RepoDir:      repoDir,
	}, nil
}

// ScanRepositoryActivity scans a repository for vulnerabilities
// This activity analyzes the source code to detect security issues and vulnerabilities
// It processes the code using AI models to identify OWASP Top 10 security risks
func ScanRepositoryActivity(ctx context.Context, input ScanActivityInput) (*ScanActivityOutput, error) {
	log := logger.Get()
	log.Info("Starting repository scan activity",
		zap.String("repo_id", input.RepositoryID),
		zap.String("repo_dir", input.RepoDir))

	// Create instances of required services
	// These services handle the various aspects of the scanning process
	dbQueries := db.NewQueries()
	githubService := services.NewGitHubService(dbQueries)
	scannerService := services.NewScannerService(githubService)

	// Generate a unique scan ID to track this specific scan operation
	scanID := uuid.New().String()

	// Get the database connection to record scan information
	sqlDB := dbQueries.GetDB()

	// Check if database connection is available to persist scan results
	var databaseAvailable bool = false
	var submitterEmail string    // Track the email of the user who submitted the scan
	var createdBy sql.NullString // Define createdBy at a broader scope
	if sqlDB != nil {
		databaseAvailable = true

		// First, retrieve the repository's created_by field (the user who added the repository)
		// This helps us track who initiated the repository scan
		err := sqlDB.QueryRowContext(ctx,
			`SELECT created_by FROM repositories WHERE id = $1`,
			input.RepositoryID).Scan(&createdBy)

		if err != nil && err != sql.ErrNoRows {
			log.Warn("Failed to get repository creator information",
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			// Continue anyway, the created_by will remain NULL
		} else if createdBy.Valid && createdBy.String != "" {
			// If we have a creator ID, try to get their email for later notification
			// This email will be used to send scan results if notifications are enabled
			err = sqlDB.QueryRowContext(ctx,
				`SELECT email FROM users WHERE id = $1 AND email IS NOT NULL AND email != ''`,
				createdBy.String).Scan(&submitterEmail)

			if err != nil {
				log.Warn("Failed to get submitter email",
					zap.String("user_id", createdBy.String),
					zap.Error(err))
				// Continue anyway, the email notification will be skipped
			} else {
				log.Info("Found submitter email for notifications",
					zap.String("email", submitterEmail))
			}
		}

		// Create a scan record in the database to track the scan progress
		// This record will be updated when the scan completes or fails
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

		// Send email notification to the scan submitter
		var repoName string
		err = sqlDB.QueryRowContext(ctx,
			`SELECT name FROM repositories WHERE id = $1`,
			input.RepositoryID).Scan(&repoName)

		if err != nil {
			log.Error("Failed to fetch repository name for email notification",
				zap.String("repo_id", input.RepositoryID),
				zap.Error(err))
			repoName = "Unknown Repository"
		}

		// Initialize email service for sending notifications
		emailService := services.NewEmailService(dbQueries)

		vulnCount := 0
		if scanResult != nil {
			vulnCount = len(scanResult.Vulnerabilities)
		}

		// First try to use the email from the database
		emailToNotify := submitterEmail

		// If no email found in database, try using the one provided in the input
		if emailToNotify == "" && input.Email != "" {
			emailToNotify = input.Email
			log.Info("Using email from scan input for notification",
				zap.String("email", emailToNotify))
		}

		// Send email if we have an email address and notification is requested
		shouldSendEmail := input.NotifyEmail || emailToNotify != ""

		if shouldSendEmail && emailToNotify != "" {
			err = emailService.SendScanCompletionEmail(
				emailToNotify,
				repoName,
				input.RepositoryID,
				vulnCount)

			if err != nil {
				log.Error("Failed to send scan completion email",
					zap.String("email", emailToNotify),
					zap.String("repo_name", repoName),
					zap.Error(err))
			} else {
				log.Info("Scan completion email sent successfully",
					zap.String("email", emailToNotify),
					zap.String("repo_name", repoName))
			}
		} else {
			log.Warn("Skipping email notification",
				zap.String("repo_id", input.RepositoryID),
				zap.String("repo_name", repoName),
				zap.Bool("notify_email", input.NotifyEmail),
				zap.Bool("has_email", emailToNotify != ""),
				zap.String("input_email", input.Email))
		}

		// Always save notification in the database for UI notifications
		if createdBy.Valid && createdBy.String != "" {
			_, err = sqlDB.ExecContext(ctx,
				`INSERT INTO notifications (id, user_id, type, title, message, read, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
				uuid.New().String(),
				createdBy.String,
				"scan_completed",
				"Scan Completed: "+repoName,
				fmt.Sprintf("Scan for repository %s completed with %d vulnerabilities found.", repoName, vulnCount),
				false)

			if err != nil {
				log.Error("Failed to create notification record",
					zap.String("user_id", createdBy.String),
					zap.String("repo_name", repoName),
					zap.Error(err))
			} else {
				log.Info("Created UI notification record",
					zap.String("user_id", createdBy.String),
					zap.String("repo_name", repoName))
			}
		}
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
