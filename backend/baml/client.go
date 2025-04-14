package baml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
)

// Note: This is a placeholder for actual BAML SDK integration.
// The actual implementation would use the BAML Go SDK to call the prompts.

// Vulnerability represents a security vulnerability detected by the AI scan
type Vulnerability struct {
	VulnerabilityType string `json:"vulnerability_type"`
	LineStart         int    `json:"line_start"`
	LineEnd           int    `json:"line_end"`
	Severity          string `json:"severity"`
	Description       string `json:"description"`
	Remediation       string `json:"remediation"`
	CodeSnippet       string `json:"code_snippet"`
}

// CodeScanResult represents the result of a code scan
type CodeScanResult struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
}

// OpenAIRequestPayload represents a request to the OpenAI API
type OpenAIRequestPayload struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message is part of the OpenAI chat API request
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponsePayload represents a response from the OpenAI API
type OpenAIResponsePayload struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// CodeScannerClient is a client for the BAML code scanner prompt
type CodeScannerClient struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
}

// NewCodeScannerClient creates a new code scanner client
func NewCodeScannerClient() *CodeScannerClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Warn("OPENAI_API_KEY environment variable not set, BAML scans will fail")
	}

	return &CodeScannerClient{
		apiKey:      apiKey,
		model:       "gpt-4-turbo", // Use the model specified in the BAML file
		maxTokens:   4000,
		temperature: 0.0,
	}
}

// ScanCode scans code for vulnerabilities using the BAML code scanner prompt
func (c *CodeScannerClient) ScanCode(ctx context.Context, code, language, filepath string, vulnerabilityTypes []string) (*CodeScanResult, error) {
	log := logger.FromContext(ctx)
	if log == nil {
		log = logger.Get()
	}

	log.Debug("BAML scanning code",
		zap.String("filepath", filepath),
		zap.String("language", language),
		zap.Strings("vulnerability_types", vulnerabilityTypes))

	if c.apiKey == "" {
		log.Error("OpenAI API key not set, cannot scan code")
		return &CodeScanResult{Vulnerabilities: []Vulnerability{}}, fmt.Errorf("OpenAI API key not set")
	}

	// Build prompt from code_scanner.baml
	promptTemplate := `You are a security expert performing an automated code scan for OWASP Top 10 vulnerabilities.
Your task is to identify potential security vulnerabilities in the provided code.

Please analyze the following code for these specific vulnerabilities: %s

Code language: %s
File path: %s

CODE:
%s

Your task:
1. Thoroughly analyze the provided code for security vulnerabilities.
2. Focus on OWASP Top 10 vulnerabilities, especially the ones specified.
3. For each vulnerability you find, provide:
   - Vulnerability type (from the OWASP Top 10)
   - Location (line numbers where the vulnerability exists)
   - Severity (Critical, High, Medium, Low)
   - Description of the vulnerability
   - A suggested remediation

Provide output in JSON format as follows:
{
  "vulnerabilities": [
    {
      "vulnerability_type": "Injection",
      "line_start": 10,
      "line_end": 15,
      "severity": "High",
      "description": "SQL injection vulnerability due to unparameterized query",
      "remediation": "Use prepared statements or an ORM",
      "code_snippet": "select * from users where name = '" + username + "'"
    }
  ]
}

If no vulnerabilities are found, return: {"vulnerabilities": []}
`

	// Format the prompt with the actual values
	vulnTypesStr := strings.Join(vulnerabilityTypes, ", ")
	formattedPrompt := fmt.Sprintf(promptTemplate, vulnTypesStr, language, filepath, code)

	// Build the OpenAI API request
	payload := OpenAIRequestPayload{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a security expert assistant that analyzes code for vulnerabilities.",
			},
			{
				Role:    "user",
				Content: formattedPrompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	client := &http.Client{
		Timeout: 10 * time.Minute, // Add a 2-minute timeout for scanning large files
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var openAIResp OpenAIResponsePayload
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI API returned no choices")
	}

	// Extract the content from the response
	content := openAIResp.Choices[0].Message.Content

	// Try to extract JSON from the content (the model might return markdown or other text)
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart >= 0 && jsonEnd >= 0 && jsonEnd > jsonStart {
		content = content[jsonStart : jsonEnd+1]
	}

	// Parse the JSON result
	var result CodeScanResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Error("Failed to parse OpenAI response as JSON",
			zap.String("content", content),
			zap.Error(err))
		return &CodeScanResult{Vulnerabilities: []Vulnerability{}}, nil
	}

	log.Debug("BAML scan completed",
		zap.String("filepath", filepath),
		zap.Int("vulnerabilities_found", len(result.Vulnerabilities)))

	return &result, nil
}
