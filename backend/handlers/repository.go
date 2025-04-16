package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/temporal"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// VulnerabilityType is imported from services package
type VulnerabilityType = services.VulnerabilityType

// Import vulnerability type constants
var (
	Injection                  = services.Injection
	BrokenAccessControl        = services.BrokenAccessControl
	CryptographicFailures      = services.CryptographicFailures
	InsecureDesign             = services.InsecureDesign
	SecurityMisconfiguration   = services.SecurityMisconfiguration
	VulnerableComponents       = services.VulnerableComponents
	IdentificationAuthFailures = services.IdentificationAuthFailures
	SoftwareIntegrityFailures  = services.SoftwareIntegrityFailures
	SecurityLoggingFailures    = services.SecurityLoggingFailures
	ServerSideRequestForgery   = services.ServerSideRequestForgery
)

// RepositoryHandler handles repository-related API requests
// This is the main handler for all GitHub repository operations including
// repository creation, retrieval, scanning, and reporting scan results
type RepositoryHandler struct {
	GitHubService  services.GitHubService  // Service for GitHub API integration
	ScannerService services.ScannerService // Service for vulnerability scanning
	OpenAIService  services.OpenAIService  // Service for AI-powered analysis
	TemporalClient client.Client           // Client for Temporal workflow engine
}

// NewRepositoryHandler creates a new repository handler with all required dependencies
// This is a factory function that initializes the handler with the services it needs
// to interact with GitHub, scan code, perform AI analysis, and start workflows
func NewRepositoryHandler(
	githubService services.GitHubService,
	scannerService services.ScannerService,
	openAIService services.OpenAIService,
	temporalClient client.Client,
) *RepositoryHandler {
	logger.Info("Initializing repository handler")
	return &RepositoryHandler{
		GitHubService:  githubService,
		ScannerService: scannerService,
		OpenAIService:  openAIService,
		TemporalClient: temporalClient,
	}
}

// ScanPublicRepository handles scanning a public GitHub repository by URL
// This endpoint doesn't require authentication or GitHub integration
func (h *RepositoryHandler) ScanPublicRepository(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Handling public repository scan request")

	// Parse request body
	var req struct {
		RepoURL string `json:"repo_url"`
		Email   string `json:"email"` // Optional email for notification
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RepoURL == "" {
		log.Warn("Empty repository URL received")
		http.Error(w, "Repository URL is required", http.StatusBadRequest)
		return
	}

	log.Debug("Processing repository URL", zap.String("url", req.RepoURL))

	// Parse the GitHub URL to extract owner and repo name
	owner, name, err := parseGitHubRepoURL(req.RepoURL)
	if err != nil {
		log.Error("Invalid GitHub URL", zap.String("url", req.RepoURL), zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid GitHub URL: %v", err), http.StatusBadRequest)
		return
	}

	log.Debug("Extracted repository details",
		zap.String("owner", owner),
		zap.String("name", name))

	// Fetch repository details from GitHub API
	log.Debug("Fetching repository info from GitHub API")
	repoInfo, err := h.GitHubService.FetchRepositoryInfo(r.Context(), owner, name)
	if err != nil {
		log.Error("Failed to fetch repository info",
			zap.String("owner", owner),
			zap.String("name", name),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to fetch repository info: %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Repository info fetched successfully",
		zap.String("id", repoInfo.ID),
		zap.String("url", repoInfo.URL))

	// Store repository information in the database
	// Get database connection
	dbConn := h.GitHubService.GetDatabaseConnection()
	if dbConn == nil {
		log.Error("Database connection is nil, cannot store repository information")
		http.Error(w, "Internal server error: database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Get or create user based on email if provided
	var userID string
	if req.Email != "" {
		log.Info("Email provided for notifications", zap.String("email", req.Email))

		// Check if user with this email already exists
		err = dbConn.QueryRowContext(r.Context(),
			`SELECT id FROM users WHERE email = $1`,
			req.Email).Scan(&userID)

		if err == sql.ErrNoRows {
			// Create a new user with this email
			userID = uuid.New().String()
			_, err = dbConn.ExecContext(r.Context(),
				`INSERT INTO users (id, email, name, created_at, updated_at, role, receive_notifications)
				VALUES ($1, $2, $3, NOW(), NOW(), 'user', true)`,
				userID, req.Email, fmt.Sprintf("User %s", req.Email[:strings.Index(req.Email, "@")]))

			if err != nil {
				log.Error("Failed to create user for notification",
					zap.String("email", req.Email),
					zap.Error(err))
				// Continue without user ID
				userID = ""
			} else {
				log.Info("Created new user for notification",
					zap.String("user_id", userID),
					zap.String("email", req.Email))
			}
		} else if err != nil {
			log.Error("Error checking for existing user",
				zap.String("email", req.Email),
				zap.Error(err))
			// Continue without user ID
			userID = ""
		} else {
			log.Info("Found existing user for notification",
				zap.String("user_id", userID),
				zap.String("email", req.Email))
		}
	} else {
		// Get the user ID from the context (if authenticated)
		contextUserID, ok := r.Context().Value("userID").(string)
		if ok {
			userID = contextUserID
			log.Info("Using authenticated user", zap.String("user_id", userID))
		}
	}

	// Check if repository already exists
	var existingRepoID string
	err = dbConn.QueryRowContext(r.Context(),
		`SELECT id FROM repositories WHERE owner = $1 AND name = $2`,
		owner, name).Scan(&existingRepoID)

	if err != nil && err != sql.ErrNoRows {
		log.Error("Error checking for existing repository",
			zap.String("owner", owner),
			zap.String("name", name),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// Repository doesn't exist, create it
		description := repoInfo.Description
		if description == "" {
			description = "Repository scanned via AI-powered SAST tool"
		}

		// Create the repository with creator information
		_, err = dbConn.ExecContext(r.Context(),
			`INSERT INTO repositories (id, owner, name, url, clone_url, description, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			repoInfo.ID, owner, name, repoInfo.URL, repoInfo.CloneURL, description, sql.NullString{String: userID, Valid: userID != ""})
		if err != nil {
			log.Error("Failed to store repository information",
				zap.String("repo_id", repoInfo.ID),
				zap.Error(err))
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
		log.Info("Repository stored in database",
			zap.String("repo_id", repoInfo.ID))

		// If we have a user ID, add an entry to user_repositories table for association
		if userID != "" {
			// Check if user_repositories table exists
			var joinTableExists bool
			err = dbConn.QueryRowContext(r.Context(), `
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = 'user_repositories'
				)
			`).Scan(&joinTableExists)

			if err != nil {
				log.Error("Error checking user_repositories table existence", zap.Error(err))
				// Continue anyway, this is not critical
			} else if joinTableExists {
				// Add repository to user_repositories table
				_, err = dbConn.ExecContext(r.Context(),
					`INSERT INTO user_repositories (user_id, repository_id) VALUES ($1, $2)
					ON CONFLICT (user_id, repository_id) DO NOTHING`,
					userID, repoInfo.ID)

				if err != nil {
					log.Error("Failed to associate repository with user in user_repositories",
						zap.String("repo_id", repoInfo.ID),
						zap.String("user_id", userID),
						zap.Error(err))
					// Continue anyway, this is not critical
				} else {
					log.Info("Repository associated with user in user_repositories",
						zap.String("repo_id", repoInfo.ID),
						zap.String("user_id", userID))
				}
			}
		}
	} else {
		// Repository exists, update it
		_, err = dbConn.ExecContext(r.Context(),
			`UPDATE repositories SET url = $1, clone_url = $2, updated_at = NOW() WHERE id = $3`,
			repoInfo.URL, repoInfo.CloneURL, repoInfo.ID)
		if err != nil {
			log.Error("Failed to update repository information",
				zap.String("repo_id", repoInfo.ID),
				zap.Error(err))
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
		log.Info("Repository information updated",
			zap.String("repo_id", repoInfo.ID))

		// If we have a user ID, ensure association in user_repositories table
		if userID != "" {
			// Check if user_repositories table exists
			var joinTableExists bool
			err = dbConn.QueryRowContext(r.Context(), `
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = 'user_repositories'
				)
			`).Scan(&joinTableExists)

			if err != nil {
				log.Error("Error checking user_repositories table existence", zap.Error(err))
				// Continue anyway, this is not critical
			} else if joinTableExists {
				// Add repository to user_repositories table if not already there
				_, err = dbConn.ExecContext(r.Context(),
					`INSERT INTO user_repositories (user_id, repository_id) VALUES ($1, $2)
					ON CONFLICT (user_id, repository_id) DO NOTHING`,
					userID, repoInfo.ID)

				if err != nil {
					log.Error("Failed to associate repository with user in user_repositories",
						zap.String("repo_id", repoInfo.ID),
						zap.String("user_id", userID),
						zap.Error(err))
					// Continue anyway, this is not critical
				} else {
					log.Info("Repository associated with user in user_repositories",
						zap.String("repo_id", repoInfo.ID),
						zap.String("user_id", userID))
				}
			}
		}
	}

	// If we have a user ID, make sure to associate them with this repository as creator
	if userID != "" && existingRepoID == "" {
		// Create a scan record with the user as creator
		scanID := uuid.New().String()
		_, err = dbConn.ExecContext(r.Context(),
			`INSERT INTO scans (id, repository_id, status, started_at, created_by)
			VALUES ($1, $2, $3, NOW(), $4)`,
			scanID, repoInfo.ID, "pending", userID)

		if err != nil {
			log.Error("Failed to create scan record with user association",
				zap.String("repo_id", repoInfo.ID),
				zap.String("user_id", userID),
				zap.Error(err))
			// Continue anyway
		} else {
			log.Info("Created scan record with user association",
				zap.String("scan_id", scanID),
				zap.String("repo_id", repoInfo.ID),
				zap.String("user_id", userID))
		}
	}

	// Initiate Temporal workflow for repository scanning
	workflowOptions := client.StartWorkflowOptions{
		ID:        "scan-workflow-" + repoInfo.ID,
		TaskQueue: "SCAN_TASK_QUEUE",
	}

	workflowInput := temporal.ScanWorkflowInput{
		RepositoryID: repoInfo.ID,
		Owner:        repoInfo.Owner,
		Name:         repoInfo.Name,
		CloneURL:     repoInfo.CloneURL,
		VulnTypes: []string{"Injection", "Broken Access Control", "Cryptographic Failures",
			"Insecure Design", "Security Misconfiguration", "Vulnerable Components",
			"Identification and Authentication Failures", "Software and Data Integrity Failures",
			"Security Logging and Monitoring Failures", "Server-Side Request Forgery"},
		FileExtensions: []string{".go", ".js", ".py", ".java", ".php", ".html", ".css", ".ts", ".jsx", ".tsx"},
		NotifyEmail:    req.Email != "", // Flag to indicate whether to send email
		Email:          req.Email,       // Pass the email to the workflow
	}

	log.Debug("Starting Temporal workflow",
		zap.String("workflow_id", workflowOptions.ID),
		zap.String("repository_id", repoInfo.ID))

	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, temporal.ScanWorkflow, workflowInput)
	if err != nil {
		log.Error("Failed to start scan workflow",
			zap.String("repository_id", repoInfo.ID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start scan workflow: %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Scan workflow initiated successfully",
		zap.String("run_id", we.GetRunID()),
		zap.String("scan_id", repoInfo.ID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"scan_id":       repoInfo.ID,
		"status":        "scan_initiated",
		"run_id":        we.GetRunID(),
		"repository":    req.RepoURL,
		"repository_id": repoInfo.ID,
	})
}

// GetScanStatus handles getting the status of a scan
func (h *RepositoryHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	scanID := chi.URLParam(r, "id")
	if scanID == "" {
		log.Warn("Missing scan ID in request")
		http.Error(w, "Scan ID is required", http.StatusBadRequest)
		return
	}

	log.Debug("Getting scan status", zap.String("scan_id", scanID))

	// Initialize default values
	var resultsAvailable bool = false
	var status string = "unknown"

	// First check if results are available in the database
	dbQueries := db.NewQueries()
	dbConn := dbQueries.GetDB()

	// Check if we have a valid database connection
	if dbConn != nil {
		// Query the database for results availability
		err := dbConn.QueryRowContext(r.Context(),
			"SELECT results_available FROM scans WHERE id = $1", scanID).Scan(&resultsAvailable)

		if err != nil && err != sql.ErrNoRows {
			log.Error("Failed to query scan status from database",
				zap.String("scan_id", scanID),
				zap.Error(err))
		}
	} else {
		log.Warn("No database connection available", zap.String("scan_id", scanID))
	}

	// Query the Temporal workflow execution
	workflowID := "scan-workflow-" + scanID

	// Check if workflow is running
	log.Debug("Querying workflow execution", zap.String("workflow_id", workflowID))
	resp, err := h.TemporalClient.DescribeWorkflowExecution(r.Context(), workflowID, "")
	if err != nil {
		log.Error("Failed to get workflow status",
			zap.String("scan_id", scanID),
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get workflow status: %v", err), http.StatusInternalServerError)
		return
	}

	workflowStatus := resp.WorkflowExecutionInfo.Status
	log.Debug("Received workflow status",
		zap.Stringer("status", workflowStatus),
		zap.String("scan_id", scanID))

	switch workflowStatus {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		status = "in_progress"
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		status = "completed"
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		status = "failed"
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		status = "canceled"
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		status = "timed_out"
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		status = "in_progress"
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		status = "terminated"
	}

	log.Info("Scan status retrieved successfully",
		zap.String("scan_id", scanID),
		zap.String("status", status))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scan_id":           scanID,
		"status":            status,
		"results_available": resultsAvailable,
	})
}

// GetScanResults handles getting the results of a scan
func (h *RepositoryHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	scanID := chi.URLParam(r, "id")
	if scanID == "" {
		log.Warn("Missing scan ID in request")
		http.Error(w, "Scan ID is required", http.StatusBadRequest)
		return
	}

	log.Debug("Getting scan results", zap.String("scan_id", scanID))

	// Define workflowID here so it's available throughout the function
	workflowID := "scan-workflow-" + scanID

	// Initialize default values
	var resultsAvailable bool = false
	var scanStatus string = "unknown"

	// First, try to check results availability in the database
	dbQueries := db.NewQueries()
	dbConn := dbQueries.GetDB()

	// Check if we have a valid database connection
	if dbConn != nil {
		// Query the database for results availability
		err := dbConn.QueryRowContext(r.Context(),
			"SELECT results_available, status FROM scans WHERE id = $1", scanID).Scan(&resultsAvailable, &scanStatus)

		if err != nil {
			if err != sql.ErrNoRows {
				log.Error("Failed to query scan status from database",
					zap.String("scan_id", scanID),
					zap.Error(err))
			} else {
				log.Debug("No scan found in database",
					zap.String("scan_id", scanID))
			}
			// Keep using default values
		}
	} else {
		log.Warn("No database connection available", zap.String("scan_id", scanID))
	}

	// If results are not available in DB, check the workflow status
	if !resultsAvailable {
		// Check workflow execution status
		resp, err := h.TemporalClient.DescribeWorkflowExecution(r.Context(), workflowID, "")
		if err != nil {
			log.Error("Failed to get workflow status",
				zap.String("scan_id", scanID),
				zap.String("workflow_id", workflowID),
				zap.Error(err))
			http.Error(w, fmt.Sprintf("Failed to get scan results: %v", err), http.StatusInternalServerError)
			return
		}

		// If workflow is still running, report that scan is in progress
		workflowStatus := resp.WorkflowExecutionInfo.Status
		if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_RUNNING { // RUNNING
			log.Info("Scan is still in progress", zap.String("scan_id", scanID))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"scan_id":                     scanID,
				"status":                      "in_progress",
				"message":                     "Scan is still in progress, results not available yet",
				"vulnerabilities_count":       0,
				"vulnerabilities_by_category": map[string][]any{},
			})
			return
		}

		// Update scanStatus based on workflow status if it's still "unknown"
		if scanStatus == "unknown" {
			switch workflowStatus {
			case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
				scanStatus = "completed"
			case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
				scanStatus = "failed"
			case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
				scanStatus = "canceled"
			case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
				scanStatus = "timed_out"
			}
		}
	}

	// If we reach here, either results are available or workflow has completed
	// So we can try to get vulnerabilities from database
	if scanStatus == "completed" {
		// Query the workflow for its result
		var result temporal.ScanWorkflowOutput
		response, queryErr := h.TemporalClient.QueryWorkflow(r.Context(), workflowID, "", "scan_result")

		// If successful query, decode the response
		if queryErr == nil && response != nil {
			// Decode the query response
			err := response.Get(&result)
			if err != nil {
				log.Error("Failed to decode query result",
					zap.String("scan_id", scanID),
					zap.Error(err))
			}
		} else {
			log.Warn("Failed to query workflow",
				zap.String("scan_id", scanID),
				zap.Error(queryErr))
		}

		// Query the scan results from the GitHubService
		vulnerabilities, err := h.GitHubService.GetRepositoryVulnerabilities(r.Context(), scanID)
		if err != nil {
			log.Error("Failed to get scan results from database",
				zap.String("scan_id", scanID),
				zap.Error(err))

			// Even if we can't get from database, we might have the result from Temporal
			if len(result.Vulnerabilities) > 0 {
				vulnerabilities = result.Vulnerabilities
			}
		}

		// Update results_available flag if the workflow is complete and we have vulnerabilities
		if !resultsAvailable && (len(vulnerabilities) > 0 || len(result.Vulnerabilities) > 0) && dbConn != nil {
			_, err := dbConn.ExecContext(r.Context(),
				"UPDATE scans SET results_available = true WHERE id = $1", scanID)
			if err != nil {
				log.Error("Failed to update results_available flag",
					zap.String("scan_id", scanID),
					zap.Error(err))
			} else {
				log.Info("Updated results_available flag to true",
					zap.String("scan_id", scanID))
				resultsAvailable = true
			}
		}

		// Group vulnerabilities by OWASP category
		categorizedVulns := make(map[string][]*services.Vulnerability)
		for _, vuln := range vulnerabilities {
			category := string(vuln.Type)
			if category == "" {
				category = "Unknown"
			}
			categorizedVulns[category] = append(categorizedVulns[category], vuln)
		}

		log.Info("Retrieved scan results successfully",
			zap.String("scan_id", scanID),
			zap.Int("vulnerability_count", len(vulnerabilities)))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"scan_id":                     scanID,
			"status":                      "completed",
			"vulnerabilities_count":       len(vulnerabilities),
			"vulnerabilities_by_category": categorizedVulns,
			"results_available":           true,
		})
		return
	}

	// If workflow failed or was canceled, report the error
	if scanStatus == "failed" || scanStatus == "canceled" || scanStatus == "timed_out" {
		log.Warn("Scan failed or was canceled",
			zap.String("scan_id", scanID),
			zap.String("status", scanStatus))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"scan_id":                     scanID,
			"status":                      scanStatus,
			"message":                     "Scan failed or was canceled",
			"vulnerabilities_count":       0,
			"vulnerabilities_by_category": map[string][]any{},
			"results_available":           false,
		})
		return
	}

	// Fallback to empty results if we couldn't determine the status or get results
	log.Warn("Could not determine scan status or get results",
		zap.String("scan_id", scanID),
		zap.String("status", scanStatus))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"scan_id":                     scanID,
		"status":                      "unknown",
		"vulnerabilities_count":       0,
		"vulnerabilities_by_category": map[string][]any{},
		"results_available":           false,
	})
}

// CreateRepositoryRequest represents a request to create a new repository
type CreateRepositoryRequest struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// CreateRepository handles creating a new repository
func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RepoURL string `json:"repo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Add repository for the user
	repo, err := h.GitHubService.AddUserRepository(r.Context(), userID, req.RepoURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repo)
}

// ListRepositories handles listing repositories
func (h *RepositoryHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log := logger.FromContext(r.Context())
	log.Debug("Listing repositories for user", zap.String("user_id", userID))

	repositories, err := h.GitHubService.ListRepositories(userID)
	if err != nil {
		log.Error("Error listing repositories", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure we always return a valid JSON array even if empty
	if repositories == nil {
		repositories = []*services.Repository{}
	}

	log.Debug("Returning repository list", zap.Int("count", len(repositories)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repositories)
}

// GetRepository handles getting a single repository
func (h *RepositoryHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log := logger.FromContext(r.Context())

	// Check if repository belongs to this user
	dbConn := h.GitHubService.GetDatabaseConnection()
	if dbConn == nil {
		log.Error("Database connection is unavailable")
		http.Error(w, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// First check if the user_repositories table exists
	var joinTableExists bool
	err := dbConn.QueryRowContext(r.Context(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'user_repositories'
		)
	`).Scan(&joinTableExists)

	if err != nil {
		log.Error("Error checking user_repositories table existence", zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If join table exists, check if the repository belongs to the user
	if joinTableExists {
		var exists bool
		err = dbConn.QueryRowContext(r.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM user_repositories
				WHERE user_id = $1 AND repository_id = $2
			)`,
			userID, id).Scan(&exists)

		if err != nil {
			log.Error("Error checking repository access", zap.Error(err))
			http.Error(w, "Error checking repository access", http.StatusInternalServerError)
			return
		}

		if !exists {
			log.Warn("User attempted to access unauthorized repository",
				zap.String("user_id", userID),
				zap.String("repo_id", id))
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}
	} else {
		// If join table doesn't exist, check if the created_by column exists and matches
		var createdByExists bool
		err = dbConn.QueryRowContext(r.Context(), `
			SELECT EXISTS (
				SELECT column_name
				FROM information_schema.columns
				WHERE table_name = 'repositories'
				AND column_name = 'created_by'
			)
		`).Scan(&createdByExists)

		if err != nil {
			log.Error("Error checking created_by column", zap.Error(err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if createdByExists {
			var exists bool
			err = dbConn.QueryRowContext(r.Context(),
				`SELECT EXISTS(
					SELECT 1 FROM repositories
					WHERE id = $1 AND created_by = $2
				)`,
				id, userID).Scan(&exists)

			if err != nil {
				log.Error("Error checking repository owner", zap.Error(err))
				http.Error(w, "Error checking repository access", http.StatusInternalServerError)
				return
			}

			if !exists {
				log.Warn("User attempted to access unauthorized repository",
					zap.String("user_id", userID),
					zap.String("repo_id", id))
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
		}
		// If neither table exists, skip the authorization check (temporary fallback)
	}

	// Get the repository details
	repo, err := h.GitHubService.GetRepository(id)
	if err != nil {
		log.Error("Error fetching repository", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repo)
}

// ScanRepository handles scanning a repository for vulnerabilities
func (h *RepositoryHandler) ScanRepository(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	id := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if repository belongs to this user
	dbConn := h.GitHubService.GetDatabaseConnection()
	if dbConn == nil {
		log.Error("Database connection is unavailable")
		http.Error(w, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// First check if the user_repositories table exists
	var joinTableExists bool
	err := dbConn.QueryRowContext(r.Context(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'user_repositories'
		)
	`).Scan(&joinTableExists)

	if err != nil {
		log.Error("Error checking user_repositories table existence", zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If join table exists, check if the repository belongs to the user
	if joinTableExists {
		var exists bool
		err = dbConn.QueryRowContext(r.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM user_repositories
				WHERE user_id = $1 AND repository_id = $2
			)`,
			userID, id).Scan(&exists)

		if err != nil {
			log.Error("Error checking repository access", zap.Error(err))
			http.Error(w, "Error checking repository access", http.StatusInternalServerError)
			return
		}

		if !exists {
			log.Warn("User attempted to scan unauthorized repository",
				zap.String("user_id", userID),
				zap.String("repo_id", id))
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}
	} else {
		// If join table doesn't exist, check if the created_by column exists and matches
		var createdByExists bool
		err = dbConn.QueryRowContext(r.Context(), `
			SELECT EXISTS (
				SELECT column_name
				FROM information_schema.columns
				WHERE table_name = 'repositories'
				AND column_name = 'created_by'
			)
		`).Scan(&createdByExists)

		if err != nil {
			log.Error("Error checking created_by column", zap.Error(err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if createdByExists {
			var exists bool
			err = dbConn.QueryRowContext(r.Context(),
				`SELECT EXISTS(
					SELECT 1 FROM repositories
					WHERE id = $1 AND created_by = $2
				)`,
				id, userID).Scan(&exists)

			if err != nil {
				log.Error("Error checking repository owner", zap.Error(err))
				http.Error(w, "Error checking repository access", http.StatusInternalServerError)
				return
			}

			if !exists {
				log.Warn("User attempted to scan unauthorized repository",
					zap.String("user_id", userID),
					zap.String("repo_id", id))
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
		}
		// If neither table exists, skip the authorization check (temporary fallback)
	}

	// Get repository info first to use in workflow
	repo, err := h.GitHubService.GetRepository(id)
	if err != nil {
		log.Error("Failed to get repository info", zap.String("repo_id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get repository info: %v", err), http.StatusInternalServerError)
		return
	}

	// Update repository status to in_progress
	dbConn = h.GitHubService.GetDatabaseConnection()
	if dbConn == nil {
		log.Error("Database connection is unavailable, cannot create scan record", zap.String("repo_id", id))
		http.Error(w, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// Create a scan record first
	scanID := id // Using the repository ID as the scan ID for simplicity
	_, err = dbConn.ExecContext(r.Context(),
		`INSERT INTO scans (id, repository_id, status, started_at)
		VALUES ($1, $2, $3, NOW())`,
		scanID, id, "in_progress")
	if err != nil {
		log.Error("Failed to create scan record",
			zap.String("repo_id", id),
			zap.Error(err))
		http.Error(w, "Failed to create scan record", http.StatusInternalServerError)
		return
	}

	log.Info("Created scan record in database", zap.String("scan_id", scanID))

	// Update repository status to in_progress
	_, err = dbConn.ExecContext(r.Context(),
		`UPDATE repositories SET updated_at = NOW() WHERE id = $1`,
		id)
	if err != nil {
		log.Error("Failed to update repository",
			zap.String("repo_id", id),
			zap.Error(err))
		// Continue anyway since the scan is already created
	}

	// Initiate Temporal workflow for repository scanning
	workflowOptions := client.StartWorkflowOptions{
		ID:        "scan-workflow-" + id,
		TaskQueue: "SCAN_TASK_QUEUE",
	}

	workflowInput := temporal.ScanWorkflowInput{
		RepositoryID:   id,
		Owner:          repo.Owner,
		Name:           repo.Name,
		CloneURL:       repo.CloneURL,
		VulnTypes:      []string{"Injection", "Broken Access Control", "Cryptographic Failures", "Insecure Design", "Security Misconfiguration"},
		FileExtensions: []string{".go", ".js", ".py", ".java", ".php", ".html", ".css", ".ts", ".jsx", ".tsx"},
	}

	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, temporal.ScanWorkflow, workflowInput)
	if err != nil {
		log.Error("Failed to start scan workflow", zap.String("repo_id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start scan workflow: %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Scan workflow initiated successfully", zap.String("repo_id", id), zap.String("run_id", we.GetRunID()))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": "scan_initiated",
		"run_id": we.GetRunID(),
	})
}

// GetVulnerabilities handles getting vulnerabilities for a repository
func (h *RepositoryHandler) GetVulnerabilities(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log := logger.FromContext(r.Context())

	// Check if repository belongs to this user
	dbConn := h.GitHubService.GetDatabaseConnection()
	if dbConn == nil {
		log.Error("Database connection is unavailable")
		http.Error(w, "Database connection unavailable", http.StatusInternalServerError)
		return
	}

	// First check if the user_repositories table exists
	var joinTableExists bool
	err := dbConn.QueryRowContext(r.Context(), `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'user_repositories'
		)
	`).Scan(&joinTableExists)

	if err != nil {
		log.Error("Error checking user_repositories table existence", zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// If join table exists, check if the repository belongs to the user
	if joinTableExists {
		var exists bool
		err = dbConn.QueryRowContext(r.Context(),
			`SELECT EXISTS(
				SELECT 1 FROM user_repositories
				WHERE user_id = $1 AND repository_id = $2
			)`,
			userID, id).Scan(&exists)

		if err != nil {
			log.Error("Error checking repository access", zap.Error(err))
			http.Error(w, "Error checking repository access", http.StatusInternalServerError)
			return
		}

		if !exists {
			log.Warn("User attempted to access unauthorized vulnerabilities",
				zap.String("user_id", userID),
				zap.String("repo_id", id))
			http.Error(w, "Repository not found", http.StatusNotFound)
			return
		}
	} else {
		// If join table doesn't exist, check if the created_by column exists and matches
		var createdByExists bool
		err = dbConn.QueryRowContext(r.Context(), `
			SELECT EXISTS (
				SELECT column_name
				FROM information_schema.columns
				WHERE table_name = 'repositories'
				AND column_name = 'created_by'
			)
		`).Scan(&createdByExists)

		if err != nil {
			log.Error("Error checking created_by column", zap.Error(err))
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if createdByExists {
			var exists bool
			err = dbConn.QueryRowContext(r.Context(),
				`SELECT EXISTS(
					SELECT 1 FROM repositories
					WHERE id = $1 AND created_by = $2
				)`,
				id, userID).Scan(&exists)

			if err != nil {
				log.Error("Error checking repository owner", zap.Error(err))
				http.Error(w, "Error checking repository access", http.StatusInternalServerError)
				return
			}

			if !exists {
				log.Warn("User attempted to access unauthorized vulnerabilities",
					zap.String("user_id", userID),
					zap.String("repo_id", id))
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
		}
		// If neither table exists, skip the authorization check (temporary fallback)
	}

	// Get vulnerabilities from GitHub service
	vulnerabilities, err := h.GitHubService.GetRepositoryVulnerabilities(r.Context(), id)
	if err != nil {
		log.Error("Error fetching vulnerabilities", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get vulnerabilities: %v", err), http.StatusInternalServerError)
		return
	}

	// Organize vulnerabilities by OWASP category
	categorizedVulns := make(map[string][]interface{})

	// Process each vulnerability
	for _, vuln := range vulnerabilities {
		// Determine the appropriate OWASP Top 10 category based on vulnerability type
		owaspCategory := mapVulnerabilityTypeToOWASP(vuln.Type)

		if categorizedVulns[owaspCategory] == nil {
			categorizedVulns[owaspCategory] = []interface{}{}
		}

		categorizedVulns[owaspCategory] = append(categorizedVulns[owaspCategory], map[string]interface{}{
			"id":             vuln.ID,
			"description":    vuln.Description,
			"severity":       vuln.Severity,
			"file_path":      vuln.FilePath,
			"line_number":    vuln.LineStart,
			"code_snippet":   vuln.Code,
			"recommendation": vuln.Remediation,
		})
	}

	// Find latest scan ID for this repository (if not already known)
	var scanID string
	err = dbConn.QueryRowContext(r.Context(),
		`SELECT id FROM scans WHERE repository_id = $1 ORDER BY created_at DESC LIMIT 1`,
		id).Scan(&scanID)
	if err != nil {
		if err == sql.ErrNoRows {
			scanID = "unknown"
		} else {
			log.Error("Error finding latest scan", zap.Error(err))
		}
	}

	// Return a properly formatted response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scan_id":                     scanID,
		"repository_id":               id,
		"status":                      "completed",
		"scan_started_at":             time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		"scan_completed_at":           time.Now().Format(time.RFC3339),
		"vulnerabilities_count":       len(vulnerabilities),
		"vulnerabilities_by_category": categorizedVulns,
		"results_available":           true,
	})
}

// Helper function to map vulnerability types to OWASP categories
func mapVulnerabilityTypeToOWASP(vulnType VulnerabilityType) string {
	switch vulnType {
	case Injection:
		return "A03:2021"
	case BrokenAccessControl:
		return "A01:2021"
	case CryptographicFailures:
		return "A02:2021"
	case InsecureDesign:
		return "A04:2021"
	case SecurityMisconfiguration:
		return "A05:2021"
	case VulnerableComponents:
		return "A06:2021"
	case IdentificationAuthFailures:
		return "A07:2021"
	case SoftwareIntegrityFailures:
		return "A08:2021"
	case SecurityLoggingFailures:
		return "A09:2021"
	case ServerSideRequestForgery:
		return "A10:2021"
	default:
		return "Other"
	}
}

// parseGitHubRepoURL parses a GitHub URL to extract owner and repo name
func parseGitHubRepoURL(url string) (owner, name string, err error) {
	// Log the parsing attempt
	log := logger.Get()
	log.Debug("Parsing GitHub URL", zap.String("url", url))

	// Normalize URL by trimming whitespace and trailing slashes
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, "/")

	// GitHub URL formats:
	// - https://github.com/owner/repo
	// - https://github.com/owner/repo.git
	// - https://github.com/owner/repo/
	// - git@github.com:owner/repo.git
	// - github.com/owner/repo

	// Handle github.com/owner/repo format (without https://)
	if strings.HasPrefix(url, "github.com/") {
		url = "https://" + url
	}

	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}
		owner = parts[0]
		name = strings.TrimSuffix(parts[1], ".git")
		return owner, name, nil
	} else if strings.HasPrefix(url, "http://github.com/") {
		parts := strings.Split(strings.TrimPrefix(url, "http://github.com/"), "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}
		owner = parts[0]
		name = strings.TrimSuffix(parts[1], ".git")
		return owner, name, nil
	} else if strings.HasPrefix(url, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(url, "git@github.com:"), "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid GitHub URL format")
		}
		owner = parts[0]
		name = strings.TrimSuffix(parts[1], ".git")
		return owner, name, nil
	}

	log.Error("Unsupported GitHub URL format", zap.String("url", url))
	return "", "", fmt.Errorf("unsupported GitHub URL format: %s (expected 'https://github.com/owner/repo')", url)
}

// DebugWorkflow provides detailed information about a Temporal workflow
func (h *RepositoryHandler) DebugWorkflow(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	scanID := chi.URLParam(r, "id")
	if scanID == "" {
		http.Error(w, "Scan ID is required", http.StatusBadRequest)
		return
	}

	workflowID := "scan-workflow-" + scanID
	log.Info("Debugging workflow", zap.String("workflow_id", workflowID))

	// Get workflow description
	resp, err := h.TemporalClient.DescribeWorkflowExecution(r.Context(), workflowID, "")
	if err != nil {
		log.Error("Failed to get workflow description", zap.Error(err))
		http.Error(w, "Failed to get workflow information: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return detailed workflow information
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"scan_id": scanID,
		"workflow_info": map[string]any{
			"workflow_id":             workflowID,
			"run_id":                  resp.WorkflowExecutionInfo.Execution.RunId,
			"type":                    resp.WorkflowExecutionInfo.Type.Name,
			"start_time":              resp.WorkflowExecutionInfo.StartTime,
			"status":                  resp.WorkflowExecutionInfo.Status.String(),
			"workflow_execution_info": resp.WorkflowExecutionInfo,
			"pending_activities":      resp.PendingActivities,
			"pending_children":        resp.PendingChildren,
		},
	})
}
