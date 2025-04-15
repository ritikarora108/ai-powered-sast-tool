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
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
)

// Repository represents a GitHub repository
type Repository struct {
	ID          string
	Name        string
	Owner       string
	URL         string
	CloneURL    string
	Description string
	CreatedAt   string
	UpdatedAt   string
	LastScanAt  *string
	Status      string
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
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
		HTMLURL  string `json:"html_url"`
		CloneURL string `json:"clone_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Generate a UUID v5 from the repository ID
	// This creates a consistent UUID based on the GitHub repo ID
	repoIDStr := fmt.Sprintf("github-repo-%d", repoInfo.ID)
	repoUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(repoIDStr))

	return &Repository{
		ID:          repoUUID.String(),
		Name:        repoInfo.Name,
		Owner:       repoInfo.Owner.Login,
		URL:         repoInfo.HTMLURL,
		CloneURL:    repoInfo.CloneURL,
		Description: repoInfo.Description,
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

	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if repositories table exists
	var tableExists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'repositories'
		)
	`).Scan(&tableExists)

	if err != nil || !tableExists {
		// If there's an error or the table doesn't exist, return the error
		return nil, fmt.Errorf("repositories table does not exist: %w", err)
	}

	// Query repositories from the database
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, owner, url, clone_url, created_at, updated_at, last_scan_at, status
		FROM repositories
		ORDER BY updated_at DESC
	`)
	if err != nil {
		// If there's a query error, return the error
		return nil, fmt.Errorf("failed to query repositories: %w", err)
	}
	defer rows.Close()

	var repositories []*Repository
	for rows.Next() {
		repo := &Repository{}
		var lastScanAt sql.NullString
		var status sql.NullString

		err := rows.Scan(
			&repo.ID,
			&repo.Name,
			&repo.Owner,
			&repo.URL,
			&repo.CloneURL,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&lastScanAt,
			&status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan repository row: %w", err)
		}

		if lastScanAt.Valid {
			repo.LastScanAt = &lastScanAt.String
		}
		if status.Valid {
			repo.Status = status.String
		} else {
			repo.Status = "pending"
		}

		repositories = append(repositories, repo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over repository rows: %w", err)
	}

	return repositories, nil
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

	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if repository already exists
	var existingRepoID string
	err = db.QueryRowContext(ctx,
		`SELECT id FROM repositories WHERE owner = $1 AND name = $2`,
		owner, name).Scan(&existingRepoID)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error checking for existing repository: %w", err)
	}

	if err == sql.ErrNoRows {
		// Repository doesn't exist, create it
		_, err = db.ExecContext(ctx,
			`INSERT INTO repositories (id, owner, name, url, clone_url, description, created_at, updated_at, status, created_by) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			repoInfo.ID, owner, name, repoInfo.URL, repoInfo.CloneURL, repoInfo.Description,
			time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), "pending", userID)
		if err != nil {
			return nil, fmt.Errorf("failed to store repository information: %w", err)
		}

		// Also add repository to user_repositories join table
		_, err = db.ExecContext(ctx,
			`INSERT INTO user_repositories (user_id, repository_id) VALUES ($1, $2)`,
			userID, repoInfo.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to associate repository with user: %w", err)
		}
	} else {
		// Repository exists, update it
		_, err = db.ExecContext(ctx,
			`UPDATE repositories SET url = $1, clone_url = $2, updated_at = $3 WHERE id = $4`,
			repoInfo.URL, repoInfo.CloneURL, time.Now().Format(time.RFC3339), existingRepoID)
		if err != nil {
			return nil, fmt.Errorf("failed to update repository information: %w", err)
		}

		// Check if repository is already associated with user
		var exists bool
		err = db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM user_repositories WHERE user_id = $1 AND repository_id = $2)`,
			userID, existingRepoID).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("error checking user-repository association: %w", err)
		}

		if !exists {
			// Add repository to user_repositories join table
			_, err = db.ExecContext(ctx,
				`INSERT INTO user_repositories (user_id, repository_id) VALUES ($1, $2)`,
				userID, existingRepoID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate repository with user: %w", err)
			}
		}

		// Use the existing ID
		repoInfo.ID = existingRepoID
	}

	return repoInfo, nil
}

func (s *gitHubService) GetRepository(id string) (*Repository, error) {
	ctx := context.Background()

	// Check if this is a sample repository ID and return it if so
	if strings.HasPrefix(id, "sample-") {
		return nil, fmt.Errorf("repository with ID %s not found", id)
	}

	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if repositories table exists
	var tableExists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'repositories'
		)
	`).Scan(&tableExists)

	if err != nil || !tableExists {
		return nil, fmt.Errorf("repositories table does not exist")
	}

	// Query repository from the database
	repo := &Repository{}
	var lastScanAt sql.NullString
	var status sql.NullString

	err = db.QueryRowContext(ctx, `
		SELECT id, name, owner, url, clone_url, created_at, updated_at, last_scan_at, status
		FROM repositories
		WHERE id = $1
	`, id).Scan(
		&repo.ID,
		&repo.Name,
		&repo.Owner,
		&repo.URL,
		&repo.CloneURL,
		&repo.CreatedAt,
		&repo.UpdatedAt,
		&lastScanAt,
		&status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("repository with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	if lastScanAt.Valid {
		repo.LastScanAt = &lastScanAt.String
	}
	if status.Valid {
		repo.Status = status.String
	} else {
		repo.Status = "pending"
	}

	return repo, nil
}

func (s *gitHubService) GetRepositoryVulnerabilities(ctx context.Context, repoID string) ([]*Vulnerability, error) {
	// Check if this is a sample repository ID and return an error
	if strings.HasPrefix(repoID, "sample-") {
		return nil, fmt.Errorf("repository with ID %s not found", repoID)
	}

	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Check if necessary tables exist
	var tablesExist bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'scans'
		) AND EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = 'vulnerabilities'
		)
	`).Scan(&tablesExist)

	if err != nil || !tablesExist {
		// If tables don't exist, return empty list
		return []*Vulnerability{}, nil
	}

	// First, find the latest scan for this repository
	var scanID string
	err = db.QueryRowContext(ctx,
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
		var vulnerabilityType string
		var remediation, codeSnippet sql.NullString

		err := rows.Scan(
			&vuln.ID,
			&vulnerabilityType,
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

		vuln.Type = VulnerabilityType(vulnerabilityType)

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

	return "", "", fmt.Errorf("unsupported GitHub URL format")
}

func (s *gitHubService) CreateRepository(owner, name, url string) (string, error) {
	// Get the database connection
	db := s.db.GetDB()
	if db == nil {
		return "", fmt.Errorf("database connection not available")
	}

	// Generate a repository ID (using nano timestamp as a simple solution)
	repoID := fmt.Sprintf("repo-%d", time.Now().UnixNano())

	// Parse the URL to get the clone URL
	parsedURL := url
	if !strings.HasSuffix(parsedURL, ".git") {
		parsedURL = parsedURL + ".git"
	}

	// Insert the repository into the database
	_, err := db.Exec(
		`INSERT INTO repositories (id, owner, name, url, clone_url, created_at, updated_at, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		repoID, owner, name, url, parsedURL,
		time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), "pending")
	if err != nil {
		return "", fmt.Errorf("failed to store repository information: %w", err)
	}

	return repoID, nil
}

// GetDatabaseConnection returns the database connection
func (s *gitHubService) GetDatabaseConnection() *sql.DB {
	if s.db != nil {
		return s.db.GetDB()
	}
	return nil
}
