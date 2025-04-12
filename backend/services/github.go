package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v60/github"
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

}

// NewGitHubService creates a new GitHub service instance
func NewGitHubService() GitHubService {
	return &gitHubService{
		client: &http.Client{},
		apiURL: "https://api.github.com",
	}
}

// gitHubService implements the GitHubService interface
type gitHubService struct {
	client *http.Client
	apiURL string
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
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Owner    struct {
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


func (s *gitHubService) CreateRepository(owner, name, url string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	repo := &github.Repository{
		Name:    &name,
		Private: github.Bool(false), // Set to true if you want a private repository
	}

	createdRepo, _, err := client.Repositories.Create(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to create repository: %w", err)
	}

	return fmt.Sprintf("%d", createdRepo.GetID()), nil
	
	
}

func (s *gitHubService) ListRepositories() ([]*Repository, error) {
	ctx := context.Background()
    client := github.NewClient(nil)
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
	// This will be implemented using our own scanning mechanism
    // For now, return an empty slice as placeholder
    return []*Vulnerability{}, nil
}
