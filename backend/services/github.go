package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v60/github"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
)

// Repository represents a GitHub repository
type Repository struct {
	ID       string
	Name     string
	Owner    string
	URL      string
	CloneURL string
}

// GitHubService defines the interface for GitHub operations
type GitHubService interface {
	// FetchRepositoryInfo retrieves repository metadata
	FetchRepositoryInfo(ctx context.Context, owner, repo string) (*Repository, error)

	// CloneRepository clones a GitHub repository to the local filesystem
	CloneRepository(ctx context.Context, repo *Repository, targetDir string) error

	// ListFiles lists files in a repository with optional filtering
	ListFiles(ctx context.Context, repoDir string, extensions []string) ([]string, error)

	CreateRepository(owner, name, url string) (string, error)
	ListRepositories() ([]*Repository, error)
	GetRepository(id string) (*Repository, error)

	// GetRepositoryVulnerabilities retrieves vulnerabilities for a repository
	GetRepositoryVulnerabilities(ctx context.Context, repoID string) ([]*Vulnerability, error)

	// AddUserRepository adds a repository for a user
	AddUserRepository(ctx context.Context, userID string, repoURL string) (*Repository, error)

	// GetDatabaseConnection returns the database connection
	GetDatabaseConnection() *sql.DB
}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService(dbQueries *db.Queries) GitHubService {
	return &gitHubService{
		client: &http.Client{},
		apiURL: "https://api.github.com",
		db:     dbQueries,
	}
}

// gitHubService implements the GitHubService interface
type gitHubService struct {
	client *http.Client
	apiURL string
	db     *db.Queries // Add database client
}

func (s *gitHubService) FetchRepositoryInfo(ctx context.Context, owner, repo string) (*Repository, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var repoInfo struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		HTMLURL  string `json:"html_url"`
		CloneURL string `json:"clone_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Repository{
		ID:       fmt.Sprintf("%d", repoInfo.ID),
		Name:     repoInfo.Name,
		Owner:    repoInfo.Owner.Login,
		URL:      repoInfo.HTMLURL,
		CloneURL: repoInfo.CloneURL,
	}, nil
}

func (s *gitHubService) CloneRepository(ctx context.Context, repo *Repository, targetDir string) error {
	// Implement Git clone using go-git
	r, err := git.PlainCloneContext(ctx, targetDir, false, &git.CloneOptions{
		URL:      repo.CloneURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Verify the repository was cloned successfully
	_, err = r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	return nil
}

func (s *gitHubService) ListFiles(ctx context.Context, repoDir string, extensions []string) ([]string, error) {
	files, err := os.ReadDir(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var result []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := filepath.Ext(file.Name())
		if len(extensions) > 0 && !slices.Contains(extensions, ext) {
			continue
		}
		result = append(result, filepath.Join(repoDir, file.Name()))
	}
	return result, nil
}

func (s *gitHubService) ListRepositories() ([]*Repository, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	// Temporary implementation until DB is fully set up
	repos, _, err := client.Repositories.ListByUser(ctx, "ritikarora108", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	var result []*Repository
	for _, repo := range repos {
		result = append(result, &Repository{
			ID:   fmt.Sprintf("%d", repo.GetID()),
			Name: repo.GetName(),
			URL:  repo.GetHTMLURL(),
		})
	}
	return result, nil
}

func (s *gitHubService) AddUserRepository(ctx context.Context, userID string, repoURL string) (*Repository, error) {
	// Parse the GitHub URL to extract owner and repo name
	owner, name, err := parseGitHubURL(repoURL)
	if err != nil {
		return nil, err
	}

	// Fetch repository details from GitHub API
	repoInfo, err := s.FetchRepositoryInfo(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	// Temporary implementation until DB is fully set up
	return repoInfo, nil
}

func (s *gitHubService) GetRepository(id string) (*Repository, error) {
	ctx := context.Background()
	client := github.NewClient(nil)
	repo, _, err := client.Repositories.Get(ctx, "ritikarora108", "test")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return &Repository{
		ID:   fmt.Sprintf("%d", repo.GetID()),
		Name: repo.GetName(),
		URL:  repo.GetHTMLURL(),
	}, nil
}

func (s *gitHubService) GetRepositoryVulnerabilities(ctx context.Context, repoID string) ([]*Vulnerability, error) {
	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// First, find the latest scan for this repository
	var scanID string
	err := db.QueryRowContext(ctx,
		`SELECT id FROM scans WHERE repository_id = $1 ORDER BY created_at DESC LIMIT 1`,
		repoID).Scan(&scanID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No scans found for this repository
			return []*Vulnerability{}, nil
		}
		return nil, fmt.Errorf("failed to find latest scan: %w", err)
	}

	// Query the vulnerabilities for this scan
	rows, err := db.QueryContext(ctx,
		`SELECT id, vulnerability_type, file_path, line_start, line_end, severity, description, 
		remediation, code_snippet FROM vulnerabilities WHERE scan_id = $1`,
		scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to query vulnerabilities: %w", err)
	}
	defer rows.Close()

	var vulnerabilities []*Vulnerability
	for rows.Next() {
		vuln := &Vulnerability{}
		var remediation, codeSnippet sql.NullString

		err := rows.Scan(
			&vuln.ID,
			&vuln.Type,
			&vuln.FilePath,
			&vuln.LineStart,
			&vuln.LineEnd,
			&vuln.Severity,
			&vuln.Description,
			&remediation,
			&codeSnippet,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vulnerability row: %w", err)
		}

		if remediation.Valid {
			vuln.Remediation = remediation.String
		}
		if codeSnippet.Valid {
			vuln.Code = codeSnippet.String
		}

		vulnerabilities = append(vulnerabilities, vuln)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over vulnerability rows: %w", err)
	}

	return vulnerabilities, nil
}

// Helper function to parse GitHub URLs
func parseGitHubURL(url string) (owner, name string, err error) {
	// Simple implementation for now
	return "ritikarora108", "test", nil
}

func (s *gitHubService) CreateRepository(owner, name, url string) (string, error) {
	// Temporary implementation - in a real app, we would create the repo on GitHub
	// and store it in our database
	return "123", nil
}

// GetDatabaseConnection returns the database connection
func (s *gitHubService) GetDatabaseConnection() *sql.DB {
	if s.db != nil {
		return s.db.GetDB()
	}
	return nil
}
