package services

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AnalysisRequest represents a request to analyze code for vulnerabilities
type AnalysisRequest struct {
	Code      string
	Language  string
	FilePath  string
	VulnTypes []VulnerabilityType
}

// AnalysisResponse represents the response from OpenAI after code analysis
type AnalysisResponse struct {
	Vulnerabilities []*Vulnerability
	RawResponse     string // Raw response from OpenAI for debugging
}

// OpenAIService defines the interface for OpenAI operations
type OpenAIService interface {
	// AnalyzeCode analyzes code for vulnerabilities using OpenAI
	AnalyzeCode(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error)

	// AnalyzeMultipleFiles analyzes multiple files in batch
	AnalyzeMultipleFiles(ctx context.Context, requests []*AnalysisRequest) ([]*AnalysisResponse, error)
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService() OpenAIService {
	return &openAIService{}
}

// openAIService implements the OpenAIService interface
type openAIService struct {
	// Configuration fields
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
}

func (s *openAIService) AnalyzeCode(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error) {
	// Implement OpenAI API call for code analysis
	client := &http.Client{}
	reqBody, err := json.Marshal(map[string]interface{}{
		"model":       s.model,
		"prompt":      req.Code,
		"max_tokens":  s.maxTokens,
		"temperature": s.temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/"+s.model+"/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Choices []struct {
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// This will be replaced with BAML orchestration
	return &AnalysisResponse{
		Vulnerabilities: []*Vulnerability{}, // Placeholder for actual vulnerabilities
		RawResponse:     apiResponse.Choices[0].Text,
	}, nil
}

func (s *openAIService) AnalyzeMultipleFiles(ctx context.Context, requests []*AnalysisRequest) ([]*AnalysisResponse, error) {
	var responses []*AnalysisResponse
	for _, req := range requests {
		response, err := s.AnalyzeCode(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze code for request: %w", err)
		}
		responses = append(responses, response)
	}
	return responses, nil
}
