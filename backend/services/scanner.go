package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/baml"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
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
		bamlClient:    baml.NewCodeScannerClient(),
	}
}

// scannerService implements the ScannerService interface
type scannerService struct {
	githubService GitHubService
	bamlClient    *baml.CodeScannerClient
	openAIService OpenAIService // We'll set this in the constructor if needed
}

func (s *scannerService) ScanRepository(ctx context.Context, repoDir string, options *ScanOptions) (*ScanResult, error) {
	log := logger.FromContext(ctx)
	if log == nil {
		log = logger.Get()
	}

	log.Info("Starting repository scan", zap.String("repo_dir", repoDir))

	// Create a scan record with a unique ID
	scanID := uuid.New().String()

	// Check if options are provided
	if options == nil {
		options = &ScanOptions{
			VulnerabilityTypes: []VulnerabilityType{
				Injection,
				BrokenAccessControl,
				CryptographicFailures,
				InsecureDesign,
				SecurityMisconfiguration,
				VulnerableComponents,
				IdentificationAuthFailures,
				SoftwareIntegrityFailures,
				SecurityLoggingFailures,
				ServerSideRequestForgery,
			},
			MaxFiles:       100,
			FileExtensions: []string{".go", ".js", ".py", ".java", ".php", ".html", ".css", ".ts", ".jsx", ".tsx"},
		}
	}

	// Find all eligible files for scanning
	var filesToScan []string
	log.Debug("Finding files to scan", zap.Strings("extensions", options.FileExtensions))

	// Define directories to skip (common dependency and non-application directories)
	dirsToSkip := map[string]bool{
		".git":              true,
		"node_modules":      true,
		"vendor":            true,
		"venv":              true,
		"env":               true,
		"lib":               true,
		"bin":               true,
		"dist":              true,
		"build":             true,
		"site-packages":     true,
		".github":           true,
		"__pycache__":       true,
		".pytest_cache":     true,
		".cache":            true,
		"package-lock.json": true,
		"yarn.lock":         true,
	}

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories that are likely not application code
			if dirsToSkip[info.Name()] {
				log.Debug("Skipping dependency directory", zap.String("dir", info.Name()))
				return filepath.SkipDir
			}

			// Also skip directories that have paths containing site-packages or other common dependency paths
			if strings.Contains(path, "site-packages") ||
				strings.Contains(path, "node_modules") ||
				strings.Contains(path, "vendor") ||
				strings.Contains(path, ".cache") {
				log.Debug("Skipping dependency path", zap.String("path", path))
				return filepath.SkipDir
			}

			return nil
		}

		// Check if file has one of the target extensions
		ext := filepath.Ext(path)
		for _, targetExt := range options.FileExtensions {
			if ext == targetExt {
				// Skip minified JavaScript files, which are not typically vulnerabilities
				if (ext == ".js" || ext == ".css") && strings.Contains(path, ".min.") {
					return nil
				}

				// Skip test files if they're not important for vulnerability scanning
				if strings.Contains(path, "_test.go") ||
					strings.Contains(path, "test_") ||
					strings.Contains(path, "spec.") {
					return nil
				}

				log.Debug("Adding file to scan list", zap.String("file", path))
				filesToScan = append(filesToScan, path)
				break
			}
		}

		// Limit the number of files to scan
		if options.MaxFiles > 0 && len(filesToScan) >= options.MaxFiles {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Error("Error walking repository directory", zap.Error(err))
		return nil, fmt.Errorf("failed to scan repository files: %w", err)
	}

	log.Info("Found files to scan", zap.Int("file_count", len(filesToScan)))

	// Convert vulnerability types to strings for BAML
	var vulnTypeStrings []string
	for _, vt := range options.VulnerabilityTypes {
		vulnTypeStrings = append(vulnTypeStrings, string(vt))
	}

	// Scan each file
	var allVulnerabilities []*Vulnerability

	for _, filePath := range filesToScan {
		relPath, err := filepath.Rel(repoDir, filePath)
		if err != nil {
			log.Warn("Could not get relative path", zap.String("file", filePath), zap.Error(err))
			relPath = filePath
		}

		log.Debug("Scanning file", zap.String("file", relPath))

		// Read the file content
		codeBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Warn("Failed to read file", zap.String("file", relPath), zap.Error(err))
			continue
		}

		code := string(codeBytes)
		language := getLanguageFromExt(filepath.Ext(filePath))

		// Use BAML client to scan the code
		result, err := s.bamlClient.ScanCode(ctx, code, language, relPath, vulnTypeStrings)
		if err != nil {
			log.Warn("Failed to scan file with BAML", zap.String("file", relPath), zap.Error(err))
			continue
		}

		// Convert BAML vulnerabilities to our format
		for _, v := range result.Vulnerabilities {
			vuln := &Vulnerability{
				ID:          uuid.New().String(),
				Type:        VulnerabilityType(v.VulnerabilityType),
				FilePath:    relPath,
				LineStart:   v.LineStart,
				LineEnd:     v.LineEnd,
				Severity:    v.Severity,
				Description: v.Description,
				Remediation: v.Remediation,
				Code:        v.CodeSnippet,
			}
			allVulnerabilities = append(allVulnerabilities, vuln)
		}
	}

	log.Info("Scan completed",
		zap.String("scan_id", scanID),
		zap.Int("vulnerability_count", len(allVulnerabilities)))

	// Normally, you would save the scan results to a database here

	return &ScanResult{
		RepositoryID:    repoDir,
		Vulnerabilities: allVulnerabilities,
		ScanTime:        time.Now().Unix(),
	}, nil
}

// ScanFile performs a vulnerability scan on a single file
func (s *scannerService) ScanFile(ctx context.Context, filePath string, options *ScanOptions) ([]*Vulnerability, error) {
	log := logger.FromContext(ctx)
	if log == nil {
		log = logger.Get()
	}

	log.Debug("Scanning individual file", zap.String("file", filePath))

	// Read the file content
	codeBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	code := string(codeBytes)
	language := getLanguageFromExt(filepath.Ext(filePath))

	// Convert vulnerability types to strings
	var vulnTypeStrings []string
	for _, vt := range options.VulnerabilityTypes {
		vulnTypeStrings = append(vulnTypeStrings, string(vt))
	}

	// Use BAML client to scan the code
	result, err := s.bamlClient.ScanCode(ctx, code, language, filePath, vulnTypeStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to scan file with BAML: %w", err)
	}

	// Convert BAML vulnerabilities to our format
	var vulnerabilities []*Vulnerability
	for _, v := range result.Vulnerabilities {
		vuln := &Vulnerability{
			ID:          uuid.New().String(),
			Type:        VulnerabilityType(v.VulnerabilityType),
			FilePath:    filePath,
			LineStart:   v.LineStart,
			LineEnd:     v.LineEnd,
			Severity:    v.Severity,
			Description: v.Description,
			Remediation: v.Remediation,
			Code:        v.CodeSnippet,
		}
		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities, nil
}

// Helper function to determine language from file extension
func getLanguageFromExt(ext string) string {
	switch ext {
	case ".go":
		return "Go"
	case ".js", ".jsx":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".php":
		return "PHP"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	default:
		return "Unknown"
	}
}
