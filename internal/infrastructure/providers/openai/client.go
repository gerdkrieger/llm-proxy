// Package openai provides OpenAI API client
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// CreateMessage sends a chat completion request to OpenAI
// Since the format is already OpenAI-compatible, we can pass through
func (c *Client) CreateMessage(ctx context.Context, req *models.OpenAIRequest) (*models.OpenAIResponse, error) {
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
	}

	for _, prefix := range validPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
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
