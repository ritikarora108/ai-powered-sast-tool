package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/temporal"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

// RepositoryHandler handles repository-related API requests
type RepositoryHandler struct {
	GitHubService  services.GitHubService
	ScannerService services.ScannerService
	OpenAIService  services.OpenAIService
	TemporalClient client.Client
}

// NewRepositoryHandler creates a new repository handler
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

// RegisterRoutes registers the repository routes
func (h *RepositoryHandler) RegisterRoutes(r chi.Router) {
	logger.Info("Registering repository routes")
	r.Post("/repositories", h.CreateRepository)
	r.Get("/repositories", h.ListRepositories)
	r.Get("/repositories/{id}", h.GetRepository)
	r.Post("/repositories/{id}/scan", h.ScanRepository)
	r.Get("/repositories/{id}/vulnerabilities", h.GetVulnerabilities)

	// Register the scan public repository endpoint
	r.Post("/scan", h.ScanPublicRepository)
	r.Get("/scan/{id}/status", h.GetScanStatus)
	r.Get("/scan/{id}/results", h.GetScanResults)

	// Add debug endpoint
	r.Get("/scan/{id}/debug", h.DebugWorkflow)
}

// ScanPublicRepository handles scanning a public GitHub repository by URL
// This endpoint doesn't require authentication or GitHub integration
func (h *RepositoryHandler) ScanPublicRepository(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Handling public repository scan request")

	// Parse request body
	var req struct {
		RepoURL string `json:"repo_url"`
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
	if dbConn != nil {
		// Check if repository already exists
		var existingRepoID string
		err := dbConn.QueryRowContext(r.Context(),
			`SELECT id FROM repositories WHERE owner = $1 AND name = $2`,
			owner, name).Scan(&existingRepoID)

		if err != nil && err != sql.ErrNoRows {
			log.Error("Error checking for existing repository",
				zap.String("owner", owner),
				zap.String("name", name),
				zap.Error(err))
		}

		if err == sql.ErrNoRows {
			// Repository doesn't exist, create it
			_, err = dbConn.ExecContext(r.Context(),
				`INSERT INTO repositories (id, owner, name, url, clone_url) VALUES ($1, $2, $3, $4, $5)`,
				repoInfo.ID, owner, name, repoInfo.URL, repoInfo.CloneURL)
			if err != nil {
				log.Error("Failed to store repository information",
					zap.String("repo_id", repoInfo.ID),
					zap.Error(err))
			} else {
				log.Info("Repository stored in database",
					zap.String("repo_id", repoInfo.ID))
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
			} else {
				log.Info("Repository information updated",
					zap.String("repo_id", repoInfo.ID))
			}
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
		"scan_id":    repoInfo.ID,
		"status":     "scan_initiated",
		"run_id":     we.GetRunID(),
		"repository": req.RepoURL,
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

	status := "unknown"
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
	json.NewEncoder(w).Encode(map[string]string{
		"scan_id": scanID,
		"status":  status,
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

	// First check if the workflow is complete
	workflowID := "scan-workflow-" + scanID

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

	// If the workflow completed, we need to get the results
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_COMPLETED { // COMPLETED
		// Query the workflow for its result
		var result temporal.ScanWorkflowOutput
		response, queryErr := h.TemporalClient.QueryWorkflow(r.Context(), workflowID, "", "scan_result")

		// If successful query, decode the response
		if queryErr == nil && response != nil {
			// Decode the query response
			err = response.Get(&result)
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
		})
		return
	}

	// If workflow failed or was canceled, report the error
	if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_FAILED ||
		workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_CANCELED ||
		workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT { // FAILED, CANCELED, TIMED_OUT
		statusStr := "failed"
		if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_CANCELED {
			statusStr = "canceled"
		} else if workflowStatus == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT {
			statusStr = "timed_out"
		}

		log.Warn("Scan failed or was canceled",
			zap.String("scan_id", scanID),
			zap.String("status", statusStr))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"scan_id":                     scanID,
			"status":                      statusStr,
			"message":                     "Scan failed or was canceled",
			"vulnerabilities_count":       0,
			"vulnerabilities_by_category": map[string][]any{},
		})
		return
	}

	// Fallback to empty results if we couldn't determine the status or get results
	log.Warn("Could not determine scan status or get results",
		zap.String("scan_id", scanID),
		zap.Stringer("workflow_status", workflowStatus))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"scan_id":                     scanID,
		"status":                      "unknown",
		"vulnerabilities_count":       0,
		"vulnerabilities_by_category": map[string][]any{},
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
	_, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	repositories, err := h.GitHubService.ListRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repositories)
}

// GetRepository handles getting a single repository
func (h *RepositoryHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	repo, err := h.GitHubService.GetRepository(id)
	if err != nil {
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

	// Get repository info first to use in workflow
	repo, err := h.GitHubService.GetRepository(id)
	if err != nil {
		log.Error("Failed to get repository info", zap.String("repo_id", id), zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get repository info: %v", err), http.StatusInternalServerError)
		return
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

	// Update repository status to in_progress
	dbConn := h.GitHubService.GetDatabaseConnection()
	if dbConn != nil {
		_, err := dbConn.ExecContext(r.Context(),
			`UPDATE repositories SET status = $1, updated_at = NOW() WHERE id = $2`,
			"in_progress", id)
		if err != nil {
			log.Error("Failed to update repository status", zap.String("repo_id", id), zap.Error(err))
		}
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

	// Get vulnerabilities from GitHub service
	vulnerabilities, err := h.GitHubService.GetRepositoryVulnerabilities(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get vulnerabilities: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vulnerabilities)
}

// parseGitHubRepoURL parses a GitHub URL to extract owner and repo name
func parseGitHubRepoURL(url string) (owner, name string, err error) {
	// Log the parsing attempt
	log := logger.Get()
	log.Debug("Parsing GitHub URL", zap.String("url", url))

	// Parse logic below...
	// GitHub URL formats:
	// - https://github.com/owner/repo
	// - https://github.com/owner/repo.git
	// - git@github.com:owner/repo.git

	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
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
	return "", "", fmt.Errorf("unsupported GitHub URL format")
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
