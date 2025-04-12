package handlers

import (
	"encoding/json"
	"net/http"
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/temporal"
)

// RepositoryHandler handles repository-related API requests
type RepositoryHandler struct {
	GitHubService   services.GitHubService
	ScannerService  services.ScannerService
	OpenAIService   services.OpenAIService
	TemporalClient  client.Client
}

// NewRepositoryHandler creates a new repository handler
func NewRepositoryHandler(
	githubService services.GitHubService,
	scannerService services.ScannerService,
	openAIService services.OpenAIService,
	temporalClient client.Client,
) *RepositoryHandler {
	return &RepositoryHandler{
		GitHubService:   githubService,
		ScannerService:  scannerService,
		OpenAIService:   openAIService,
		TemporalClient:  temporalClient,
	}
}

// RegisterRoutes registers the repository routes
func (h *RepositoryHandler) RegisterRoutes(r chi.Router) {
	r.Post("/repositories", h.CreateRepository)
	r.Get("/repositories", h.ListRepositories)
	r.Get("/repositories/{id}", h.GetRepository)
	r.Post("/repositories/{id}/scan", h.ScanRepository)
	r.Get("/repositories/{id}/vulnerabilities", h.GetVulnerabilities)
}

// CreateRepositoryRequest represents a request to create a new repository
type CreateRepositoryRequest struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	URL   string `json:"url"`
}

// CreateRepository handles creating a new repository
func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	var req CreateRepositoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repoID, err := h.GitHubService.CreateRepository(req.Owner, req.Name, req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id": repoID,
	})
	
}

// ListRepositories handles listing repositories
func (h *RepositoryHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	repositories, err := h.GitHubService.ListRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repositories)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]map[string]string{})
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":    id,
		"owner": "placeholder-owner",
		"name":  "placeholder-name",
		"url":   "placeholder-url",
	})
}

// ScanRepository handles scanning a repository for vulnerabilities
func (h *RepositoryHandler) ScanRepository(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Initiate Temporal workflow for repository scanning
	workflowOptions := client.StartWorkflowOptions{
		ID:        "scan-workflow-" + id,
		TaskQueue: "SCAN_TASK_QUEUE",
	}

	workflowInput := temporal.ScanWorkflowInput{
		RepositoryID: id,
		Owner:        "placeholder-owner",
		Name:         "placeholder-name",
		CloneURL:     "placeholder-url",
		VulnTypes:    []string{"Injection", "Broken Access Control"},
		FileExtensions: []string{".go", ".js"},
	}

	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, temporal.ScanWorkflow, workflowInput)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start scan workflow: %v", err), http.StatusInternalServerError)
		return
	}

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

	// Initiate Temporal workflow to get vulnerabilities
	workflowOptions := client.StartWorkflowOptions{
		ID:        "get-vulnerabilities-workflow-" + id,
		TaskQueue: "SCAN_TASK_QUEUE",
	}

	workflowInput := temporal.ScanWorkflowInput{
		RepositoryID: id,
		Owner:        "placeholder-owner",
		Name:         "placeholder-name",
		CloneURL:     "placeholder-url",
		VulnTypes:    []string{"Injection", "Broken Access Control"},
		FileExtensions: []string{".go", ".js"},
	}

	we, err := h.TemporalClient.ExecuteWorkflow(context.Background(), workflowOptions, temporal.ScanWorkflow, workflowInput)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start get vulnerabilities workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": "vulnerabilities_retrieval_initiated",
		"run_id": we.GetRunID(),
	})
}
