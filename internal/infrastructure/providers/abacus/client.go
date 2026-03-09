// Package abacus provides Abacus.ai API client implementation.
package abacus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

const (
	abacusAPIBaseURL = "https://api.abacus.ai/api/v0"
)

// Client represents an Abacus.ai API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *logger.Logger
	config     config.AbacusConfig
}

// AbacusRequest represents an Abacus.ai chat request
type AbacusRequest struct {
	DeploymentID      string                 `json:"deploymentId"`
	Messages          []AbacusMessage        `json:"messages"`
	Temperature       float64                `json:"temperature,omitempty"`
	MaxTokens         int                    `json:"maxTokens,omitempty"`
	TopP              float64                `json:"topP,omitempty"`
	Stream            bool                   `json:"stream,omitempty"`
	ConversationID    string                 `json:"conversationId,omitempty"`
	ExternalSessionID string                 `json:"externalSessionId,omitempty"`
	LLMName           string                 `json:"llmName,omitempty"`
	SystemMessage     string                 `json:"systemMessage,omitempty"`
	SearchDocuments   bool                   `json:"searchDocuments,omitempty"`
	KeywordFilters    map[string]interface{} `json:"keywordFilters,omitempty"`
	DocFilters        map[string]interface{} `json:"docFilters,omitempty"`
}

// AbacusMessage represents a message in Abacus.ai format
type AbacusMessage struct {
	Role    string `json:"role"` // "user" or "assistant"
	Content string `json:"content"`
}

// AbacusResponse represents an Abacus.ai chat response
type AbacusResponse struct {
	Success   bool         `json:"success"`
	Result    AbacusResult `json:"result,omitempty"`
	Error     string       `json:"error,omitempty"`
	ErrorType string       `json:"errorType,omitempty"`
	ErrorCode int          `json:"errorCode,omitempty"`
}

// AbacusResult represents the result part of Abacus.ai response
type AbacusResult struct {
	Messages         []AbacusMessage        `json:"messages"`
	ConversationID   string                 `json:"conversationId,omitempty"`
	DeploymentID     string                 `json:"deploymentId"`
	LLMDisplayName   string                 `json:"llmDisplayName,omitempty"`
	SegmentIDs       []string               `json:"segmentIds,omitempty"`
	DocIDs           []string               `json:"docIds,omitempty"`
	KeywordArguments map[string]interface{} `json:"keywordArguments,omitempty"`
}

// NewClient creates a new Abacus.ai API client
func NewClient(apiKey string, cfg config.AbacusConfig, log *logger.Logger) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: log,
		config: cfg,
	}
}

// CreateChatCompletion sends a chat completion request to Abacus.ai API
func (c *Client) CreateChatCompletion(ctx context.Context, req *AbacusRequest) (*AbacusResponse, error) {
	// Build HTTP request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chatLLM", abacusAPIBaseURL)
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers - Abacus.ai uses "apiKey" in header
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apiKey", c.apiKey)

	// Send request
	c.logger.Debugf("Sending request to Abacus.ai API: deployment=%s", req.DeploymentID)
	start := time.Now()

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	duration := time.Since(start)
	c.logger.Debugf("Abacus.ai API responded in %v with status %d", duration, httpResp.StatusCode)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response (Abacus.ai returns JSON with success field)
	var abacusResp AbacusResponse
	if err := json.Unmarshal(respBody, &abacusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if !abacusResp.Success {
		return nil, c.handleErrorResponse(&abacusResp, httpResp.StatusCode)
	}

	return &abacusResp, nil
}

// GetAvailableModels returns list of available Abacus.ai deployments/models
func (c *Client) GetAvailableModels(ctx context.Context) ([]string, error) {
	// Abacus.ai uses deployments, not static model lists
	// For now, return common LLM names that Abacus.ai supports
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"llama-3-70b",
		"llama-3-8b",
		"mistral-large",
		"mistral-medium",
	}, nil
}

// handleErrorResponse handles Abacus.ai API error responses
func (c *Client) handleErrorResponse(resp *AbacusResponse, statusCode int) error {
	if resp.Error != "" {
		return fmt.Errorf("abacus.ai API error (status=%d, type=%s, code=%d): %s",
			statusCode, resp.ErrorType, resp.ErrorCode, resp.Error)
	}
	return fmt.Errorf("abacus.ai API error: status=%d", statusCode)
}

// ValidateRequest validates an Abacus.ai request
func (c *Client) ValidateRequest(req *AbacusRequest) error {
	if req.DeploymentID == "" {
		return fmt.Errorf("deploymentId is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}
	return nil
}

// GetAPIKey returns the masked API key for logging purposes
func (c *Client) GetAPIKey() string {
	if len(c.apiKey) <= 8 {
		return "***"
	}
	return c.apiKey[:4] + "..." + c.apiKey[len(c.apiKey)-4:]
}

// Health checks if the Abacus.ai API is accessible
func (c *Client) Health(ctx context.Context) error {
	// For now, just check if we have an API key configured
	// In the future, we could make a lightweight API call to verify connectivity
	if c.apiKey == "" {
		return fmt.Errorf("no API key configured")
	}
	return nil
}
