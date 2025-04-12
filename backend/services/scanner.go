package services

import (
	"context"
	"fmt"
	"time"

)

// VulnerabilityType represents an OWASP Top 10 vulnerability type
type VulnerabilityType string

const (
	BrokenAccessControl        VulnerabilityType = "Broken Access Control"
	CryptographicFailures      VulnerabilityType = "Cryptographic Failures"
	Injection                  VulnerabilityType = "Injection"
	InsecureDesign             VulnerabilityType = "Insecure Design"
	SecurityMisconfiguration   VulnerabilityType = "Security Misconfiguration"
	VulnerableComponents       VulnerabilityType = "Vulnerable Components"
	IdentificationAuthFailures VulnerabilityType = "Identification and Authentication Failures"
	SoftwareIntegrityFailures  VulnerabilityType = "Software and Data Integrity Failures"
	SecurityLoggingFailures    VulnerabilityType = "Security Logging and Monitoring Failures"
	ServerSideRequestForgery   VulnerabilityType = "Server-Side Request Forgery"
)

// Vulnerability represents a detected security vulnerability
type Vulnerability struct {
	ID          string
	Type        VulnerabilityType
	FilePath    string
	LineStart   int
	LineEnd     int
	Severity    string // "Low", "Medium", "High", "Critical"
	Description string
	Remediation string
	Code        string // The vulnerable code snippet
}

// ScanResult represents the results of a vulnerability scan
type ScanResult struct {
	RepositoryID    string
	Vulnerabilities []*Vulnerability
	ScanTime        int64 // Unix timestamp
}

// ScanOptions contains options for the vulnerability scanner
type ScanOptions struct {
	VulnerabilityTypes []VulnerabilityType
	MaxFiles           int
	FileExtensions     []string
}

// ScannerService defines the interface for vulnerability scanning
type ScannerService interface {
	// ScanRepository performs a vulnerability scan on a repository
	ScanRepository(ctx context.Context, repoDir string, options *ScanOptions) (*ScanResult, error)

	// ScanFile performs a vulnerability scan on a single file
	ScanFile(ctx context.Context, filePath string, options *ScanOptions) ([]*Vulnerability, error)
}

// NewScannerService creates a new scanner service instance
func NewScannerService(githubService GitHubService) ScannerService {
	return &scannerService{
        githubService: githubService,
    }
}

// scannerService implements the ScannerService interface
type scannerService struct {
	githubService GitHubService
}

func (s *scannerService) ScanRepository(ctx context.Context, repoDir string, options *ScanOptions) (*ScanResult, error) {

	vulnerabilities := []*Vulnerability{}
	files, err := s.githubService.ListFiles(ctx, repoDir, options.FileExtensions)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, file := range files {
		fileVulnerabilities, err := s.ScanFile(ctx, file, options)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file %s: %w", file, err)
		}
		vulnerabilities = append(vulnerabilities, fileVulnerabilities...)
	}

	return &ScanResult{
		RepositoryID:    "placeholder-repo-id",
		Vulnerabilities: vulnerabilities,
		ScanTime:        time.Now().Unix(),
	}, nil
	
}

func (s *scannerService) ScanFile(ctx context.Context, filePath string, options *ScanOptions) ([]*Vulnerability, error) {
	openAIService := NewOpenAIService()
	req := &AnalysisRequest{
		Code:      filePath,
		Language:  "go", // Assuming the language is Go, this can be dynamic based on the file extension
		FilePath:  filePath,
		VulnTypes: options.VulnerabilityTypes,
	}

	resp, err := openAIService.AnalyzeCode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze code: %w", err)
	}

	return resp.Vulnerabilities, nil
}
