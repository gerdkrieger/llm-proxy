// Package openai provides OpenAI API client
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

const (
	baseURL = "https://api.openai.com/v1"
)

// Client is an OpenAI API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *logger.Logger
}

// NewClient creates a new OpenAI client
func NewClient(apiKey string, logger *logger.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger: logger,
	}
}

// normalizeRequestParams adjusts request parameters for model compatibility.
// Newer OpenAI models (GPT-5.x, o-series) require max_completion_tokens instead of max_tokens.
func (c *Client) normalizeRequestParams(req *models.OpenAIRequest) {
	if !requiresMaxCompletionTokens(req.Model) {
		return
	}

	// If max_tokens is set but max_completion_tokens is not, convert it
	if req.MaxTokens != nil && req.MaxCompletionTokens == nil {
		c.logger.Debugf("Converting max_tokens=%d to max_completion_tokens for model %s", *req.MaxTokens, req.Model)
		req.MaxCompletionTokens = req.MaxTokens
		req.MaxTokens = nil
	}
}

// requiresMaxCompletionTokens returns true for models that need max_completion_tokens
// instead of the legacy max_tokens parameter.
func requiresMaxCompletionTokens(model string) bool {
	prefixes := []string{"gpt-5", "o1", "o3", "o4"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

// CreateMessage sends a chat completion request to OpenAI
// Since the format is already OpenAI-compatible, we can pass through
func (c *Client) CreateMessage(ctx context.Context, req *models.OpenAIRequest) (*models.OpenAIResponse, error) {
	c.normalizeRequestParams(req)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp models.OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &openAIResp, nil
}

// StreamEvent represents a streaming event from the OpenAI API
type StreamEvent struct {
	Data  string // Raw JSON string of the chunk, or "[DONE]"
	Error error
}

// CreateMessageStream sends a streaming chat completion request to OpenAI.
// Since the response is already in OpenAI SSE format, we parse each chunk
// and forward it. The caller is responsible for writing SSE to the client.
func (c *Client) CreateMessageStream(ctx context.Context, req *models.OpenAIRequest) (<-chan StreamEvent, error) {
	// Force streaming with usage reporting
	req.Stream = true
	req.StreamOptions = &models.OpenAIStreamOptions{IncludeUsage: true}

	// Normalize parameters for model compatibility (e.g. max_tokens -> max_completion_tokens)
	c.normalizeRequestParams(req)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Use a client without timeout for streaming (context handles cancellation)
	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	eventChan := make(chan StreamEvent, 10)
	go c.readOpenAIStreamEvents(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// readOpenAIStreamEvents reads SSE events from the OpenAI streaming response.
// OpenAI sends lines like "data: {json}\n\n" and terminates with "data: [DONE]\n\n".
func (c *Client) readOpenAIStreamEvents(ctx context.Context, body io.ReadCloser, eventChan chan<- StreamEvent) {
	defer close(eventChan)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		// Check context cancellation
		select {
		case <-ctx.Done():
			c.logger.Debug("OpenAI stream context cancelled")
			return
		default:
		}

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Parse "data: ..." lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream termination
		if data == "[DONE]" {
			eventChan <- StreamEvent{Data: "[DONE]"}
			return
		}

		// Forward the raw JSON chunk
		eventChan <- StreamEvent{Data: data}
	}

	if err := scanner.Err(); err != nil {
		c.logger.Errorf(err, "Error reading OpenAI stream")
		eventChan <- StreamEvent{Error: err}
	}
}

// ListModels retrieves available models from OpenAI
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Filter for chat models only (GPT-4, GPT-3.5)
	var models []string
	for _, model := range result.Data {
		if isValidChatModel(model.ID) {
			models = append(models, model.ID)
		}
	}

	return models, nil
}

// isValidChatModel checks if model ID is a chat model
func isValidChatModel(modelID string) bool {
	validPrefixes := []string{
		"gpt-4",
		"gpt-3.5",
		"gpt-5",
		"o1",
		"o3",
		"o4",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(modelID, prefix) {
			return true
		}
	}
	return false
}

// Health checks if the OpenAI API is accessible
func (c *Client) Health(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI API unhealthy (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAPIKey returns masked API key for logging
func (c *Client) GetAPIKey() string {
	if len(c.apiKey) < 10 {
		return "***"
	}
	return c.apiKey[:7] + "..." + c.apiKey[len(c.apiKey)-4:]
}
