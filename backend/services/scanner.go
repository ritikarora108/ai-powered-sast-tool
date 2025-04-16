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
// These are based on the OWASP Top 10 2021 list of security risks
// https://owasp.org/Top10/
type VulnerabilityType string

// Constants for the OWASP Top 10 vulnerability types (2021 version)
// Each type represents a category of security risks that the scanner will detect
const (
	BrokenAccessControl        VulnerabilityType = "Broken Access Control"                      // A01:2021 - Improper access restrictions
	CryptographicFailures      VulnerabilityType = "Cryptographic Failures"                     // A02:2021 - Weak crypto, sensitive data exposure
	Injection                  VulnerabilityType = "Injection"                                  // A03:2021 - SQL, NoSQL, OS, LDAP injection flaws
	InsecureDesign             VulnerabilityType = "Insecure Design"                            // A04:2021 - Design flaws that lead to security issues
	SecurityMisconfiguration   VulnerabilityType = "Security Misconfiguration"                  // A05:2021 - Missing or insecure configurations
	VulnerableComponents       VulnerabilityType = "Vulnerable Components"                      // A06:2021 - Using components with known vulnerabilities
	IdentificationAuthFailures VulnerabilityType = "Identification and Authentication Failures" // A07:2021 - Auth problems
	SoftwareIntegrityFailures  VulnerabilityType = "Software and Data Integrity Failures"       // A08:2021 - Integrity issues
	SecurityLoggingFailures    VulnerabilityType = "Security Logging and Monitoring Failures"   // A09:2021 - Insufficient logging
	ServerSideRequestForgery   VulnerabilityType = "Server-Side Request Forgery"                // A10:2021 - SSRF attacks
)

// Vulnerability represents a detected security vulnerability
// This struct stores all the information about a specific vulnerability found in the code
type Vulnerability struct {
	ID          string            // Unique identifier for the vulnerability
	Type        VulnerabilityType // OWASP category of the vulnerability
	FilePath    string            // Path to the file containing the vulnerability
	LineStart   int               // Starting line number of the vulnerable code
	LineEnd     int               // Ending line number of the vulnerable code
	Severity    string            // "Low", "Medium", "High", "Critical"
	Description string            // Human-readable description of the vulnerability
	Remediation string            // Recommended fix for the vulnerability
	Code        string            // The vulnerable code snippet
}

// ScanResult represents the results of a vulnerability scan
// This contains all vulnerabilities found in a repository and metadata about the scan
type ScanResult struct {
	RepositoryID    string           // ID of the repository that was scanned
	Vulnerabilities []*Vulnerability // List of all vulnerabilities found
	ScanTime        int64            // Unix timestamp when the scan was performed
}

// ScanOptions contains options for the vulnerability scanner
// These settings control how the scan is performed
type ScanOptions struct {
	VulnerabilityTypes []VulnerabilityType // Types of vulnerabilities to scan for
	MaxFiles           int                 // Maximum number of files to scan
	FileExtensions     []string            // File extensions to include in the scan
}

// ScannerService defines the interface for vulnerability scanning
// This interface allows for different scanner implementations
type ScannerService interface {
	// ScanRepository performs a vulnerability scan on a repository
	// It walks through the repository directory, analyzes files, and detects vulnerabilities
	ScanRepository(ctx context.Context, repoDir string, options *ScanOptions) (*ScanResult, error)

	// ScanFile performs a vulnerability scan on a single file
	// Useful for targeted scanning of specific files
	ScanFile(ctx context.Context, filePath string, options *ScanOptions) ([]*Vulnerability, error)
}

// NewScannerService creates a new scanner service instance
// This factory function initializes a scanner with the necessary dependencies
func NewScannerService(githubService GitHubService) ScannerService {
	return &scannerService{
		githubService: githubService,
		bamlClient:    baml.NewCodeScannerClient(), // Initialize the BAML AI client for code scanning
	}
}

// scannerService implements the ScannerService interface
// This is the concrete implementation of the vulnerability scanning service
type scannerService struct {
	githubService GitHubService           // Service to interact with GitHub
	bamlClient    *baml.CodeScannerClient // Client to interact with the AI code scanner
}

// ScanRepository analyzes all eligible files in a repository for security vulnerabilities
// This method is the main entry point for scanning an entire codebase
func (s *scannerService) ScanRepository(ctx context.Context, repoDir string, options *ScanOptions) (*ScanResult, error) {
	log := logger.FromContext(ctx)
	if log == nil {
		log = logger.Get()
	}

	log.Info("Starting repository scan", zap.String("repo_dir", repoDir))

	// Create a scan record with a unique ID
	scanID := uuid.New().String()

	// Use default options if none provided
	// This ensures we have sensible defaults for vulnerability types and file extensions
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
			MaxFiles:       100, // Limit to 100 files to prevent excessive scanning time
			FileExtensions: []string{".go", ".js", ".py", ".java", ".php", ".html", ".css", ".ts", ".jsx", ".tsx"},
		}
	}

	// Find all eligible files for scanning
	// We'll collect paths to all files that match our criteria
	var filesToScan []string
	log.Debug("Finding files to scan", zap.Strings("extensions", options.FileExtensions))

	// Define directories to skip (common dependency and non-application directories)
	// This improves performance by avoiding scanning of third-party code
	dirsToSkip := map[string]bool{
		".git":              true, // Git metadata
		"node_modules":      true, // NPM dependencies
		"vendor":            true, // Go vendor directory
		"venv":              true, // Python virtual environment
		"env":               true, // Python environment
		"lib":               true, // Library code
		"bin":               true, // Binary files
		"dist":              true, // Distribution builds
		"build":             true, // Build artifacts
		"site-packages":     true, // Python packages
		".github":           true, // GitHub configuration
		"__pycache__":       true, // Python cache
		".pytest_cache":     true, // Python test cache
		".cache":            true, // Generic cache
		"package-lock.json": true, // NPM lock file
		"yarn.lock":         true, // Yarn lock file
	}

	// Walk the repository directory tree to find eligible files
	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warn("Error accessing path", zap.String("path", path), zap.Error(err))
			return nil // Continue despite errors
		}

		if info.IsDir() {
			// Skip directories that are likely not application code
			// This prevents scanning dependency directories
			if dirsToSkip[info.Name()] {
				log.Debug("Skipping dependency directory", zap.String("dir", info.Name()))
				return filepath.SkipDir
			}

			// Also skip directories that have paths containing common dependency patterns
			// This catches nested dependencies
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
		// Only scan files with extensions we're interested in
		ext := filepath.Ext(path)
		for _, targetExt := range options.FileExtensions {
			if ext == targetExt {
				// Skip minified JavaScript/CSS files, which are typically not sources of vulnerabilities
				// and can be difficult for the AI to analyze effectively
				if (ext == ".js" || ext == ".css") && strings.Contains(path, ".min.") {
					return nil
				}

				// Skip test files as they often contain sample code that triggers false positives
				// and typically don't run in production
				if strings.Contains(path, "_test.go") ||
					strings.Contains(path, "test_") ||
					strings.Contains(path, "spec.") {
					return nil
				}

				// Add the file to our scan list
				relPath, _ := filepath.Rel(repoDir, path)
				log.Debug("Adding file to scan list", zap.String("file", relPath))
				filesToScan = append(filesToScan, path)
				break
			}
		}

		// Limit the number of files to scan to prevent excessive scanning time
		if options.MaxFiles > 0 && len(filesToScan) >= options.MaxFiles {
			return filepath.SkipDir
		}

		return nil
	})

	// Handle errors or empty file lists
	if err != nil {
		log.Error("Error walking repository directory", zap.Error(err))
		// Continue with any files found instead of failing completely
		if len(filesToScan) == 0 {
			log.Warn("No files found to scan, checking if repository exists")
			// Check if repo directory exists and has content
			if _, statErr := os.Stat(repoDir); statErr != nil {
				return nil, fmt.Errorf("repository directory not found or inaccessible: %w", statErr)
			}

			// Directory exists but no matching files found
			// Try with broader extensions as fallback to find something to scan
			fallbackExts := []string{".txt", ".md", ".json", ".yml", ".yaml", ".xml"}
			log.Info("Trying fallback file types", zap.Strings("extensions", fallbackExts))

			filepath.Walk(repoDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil || info.IsDir() {
					return nil
				}
				ext := filepath.Ext(path)
				for _, fbExt := range fallbackExts {
					if ext == fbExt {
						filesToScan = append(filesToScan, path)
						break
					}
				}
				return nil
			})
		}
	}

	log.Info("Found files to scan", zap.Int("file_count", len(filesToScan)))

	// Convert vulnerability types to strings for the BAML client
	// BAML requires string input rather than our custom VulnerabilityType
	var vulnTypeStrings []string
	for _, vt := range options.VulnerabilityTypes {
		vulnTypeStrings = append(vulnTypeStrings, string(vt))
	}

	// Scan each file and collect all vulnerabilities
	var allVulnerabilities []*Vulnerability

	for _, filePath := range filesToScan {
		// Calculate the relative path from the repo root for better reporting
		relPath, err := filepath.Rel(repoDir, filePath)
		if err != nil {
			log.Warn("Could not get relative path", zap.String("file", filePath), zap.Error(err))
			relPath = filePath
		}

		log.Debug("Scanning file", zap.String("file", relPath))

		// Read the file content for analysis
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
