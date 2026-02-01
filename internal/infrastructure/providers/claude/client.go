// Package claude provides Anthropic Claude API client implementation.
package claude

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

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

const (
	claudeAPIBaseURL = "https://api.anthropic.com"
	claudeAPIVersion = "2023-06-01"
)

// Client represents a Claude API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *logger.Logger
	config     config.ClaudeConfig
}

// NewClient creates a new Claude API client
func NewClient(apiKey string, cfg config.ClaudeConfig, log *logger.Logger) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: log,
		config: cfg,
	}
}

// CreateMessage sends a chat completion request to Claude API
func (c *Client) CreateMessage(ctx context.Context, req *models.ClaudeRequest) (*models.ClaudeResponse, error) {
	// Validate request
	if err := c.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Build HTTP request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/messages", claudeAPIBaseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)

	// Send request
	c.logger.Debugf("Sending request to Claude API: %s", req.Model)
	start := time.Now()

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	duration := time.Since(start)
	c.logger.Debugf("Claude API responded in %v with status %d", duration, httpResp.StatusCode)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(httpResp.StatusCode, respBody)
	}

	// Parse successful response
	var claudeResp models.ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &claudeResp, nil
}

// validateRequest validates the Claude request
func (c *Client) validateRequest(req *models.ClaudeRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	if req.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	// Model validation removed - database handles this via /v1/models endpoint
	return nil
}

// handleErrorResponse handles error responses from Claude API
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var errorResp models.ClaudeErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		// If we can't parse the error, return a generic error
		return &APIError{
			StatusCode: statusCode,
			Message:    string(body),
			Type:       "unknown_error",
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    errorResp.Error.Message,
		Type:       errorResp.Error.Type,
	}
}

// APIError represents an error from Claude API
type APIError struct {
	StatusCode int
	Message    string
	Type       string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("Claude API error (status %d, type %s): %s", e.StatusCode, e.Type, e.Message)
}

// IsRateLimitError checks if the error is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	return e.StatusCode == 429 || e.Type == "rate_limit_error"
}

// IsRetryable checks if the error is retryable
func (e *APIError) IsRetryable() bool {
	// Retry on 5xx errors and rate limits
	return e.StatusCode >= 500 || e.IsRateLimitError()
}

// GetAPIKey returns the API key (for testing/debugging)
func (c *Client) GetAPIKey() string {
	// Only return masked version for security
	if len(c.apiKey) < 8 {
		return "***"
	}
	return c.apiKey[:8] + "..." + c.apiKey[len(c.apiKey)-4:]
}

// CreateMessageStream sends a streaming chat completion request to Claude API
func (c *Client) CreateMessageStream(ctx context.Context, req *models.ClaudeRequest) (<-chan StreamEvent, error) {
	// Validate request
	if err := c.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Force streaming
	req.Stream = true

	// Build HTTP request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/messages", claudeAPIBaseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("anthropic-version", claudeAPIVersion)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	c.logger.Debugf("Sending streaming request to Claude API: %s", req.Model)
	start := time.Now()

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Check for error response
	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		respBody, _ := io.ReadAll(httpResp.Body)
		return nil, c.handleErrorResponse(httpResp.StatusCode, respBody)
	}

	c.logger.Debugf("Claude API streaming started in %v", time.Since(start))

	// Create event channel
	eventChan := make(chan StreamEvent, 10)

	// Start goroutine to read streaming events
	go c.readStreamEvents(ctx, httpResp.Body, eventChan)

	return eventChan, nil
}

// readStreamEvents reads SSE events from the response body
func (c *Client) readStreamEvents(ctx context.Context, body io.ReadCloser, eventChan chan<- StreamEvent) {
	defer close(eventChan)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentEvent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check context cancellation
		select {
		case <-ctx.Done():
			c.logger.Debug("Stream context cancelled")
			return
		default:
		}

		// Empty line indicates end of event
		if line == "" {
			if currentEvent.Len() > 0 {
				// Parse and send the event
				c.parseAndSendEvent(currentEvent.String(), eventChan)
				currentEvent.Reset()
			}
			continue
		}

		// Add line to current event
		currentEvent.WriteString(line)
		currentEvent.WriteString("\n")
	}

	// Handle scanner error
	if err := scanner.Err(); err != nil {
		c.logger.Errorf(err, "Error reading stream")
		eventChan <- StreamEvent{
			Type:  StreamEventError,
			Error: err,
		}
	}
}

// parseAndSendEvent parses an SSE event and sends it to the channel
func (c *Client) parseAndSendEvent(eventText string, eventChan chan<- StreamEvent) {
	lines := strings.Split(eventText, "\n")
	var data strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLine := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data.Len() > 0 {
				data.WriteString("\n")
			}
			data.WriteString(dataLine)
		}
		// We ignore the "event:" line as the type is in the JSON data
	}

	dataStr := data.String()
	if dataStr == "" {
		return
	}

	// Parse the data as JSON
	var claudeEvent models.ClaudeStreamEvent
	if err := json.Unmarshal([]byte(dataStr), &claudeEvent); err != nil {
		c.logger.Warnf("Failed to parse stream event: %v", err)
		return
	}

	// Convert to our StreamEvent type
	streamEvent := StreamEvent{
		Type:        StreamEventType(claudeEvent.Type),
		ClaudeEvent: &claudeEvent,
	}

	eventChan <- streamEvent
}

// StreamEventType represents the type of stream event
type StreamEventType string

const (
	StreamEventMessageStart      StreamEventType = "message_start"
	StreamEventContentBlockStart StreamEventType = "content_block_start"
	StreamEventContentBlockDelta StreamEventType = "content_block_delta"
	StreamEventContentBlockStop  StreamEventType = "content_block_stop"
	StreamEventMessageDelta      StreamEventType = "message_delta"
	StreamEventMessageStop       StreamEventType = "message_stop"
	StreamEventPing              StreamEventType = "ping"
	StreamEventError             StreamEventType = "error"
)

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type        StreamEventType
	ClaudeEvent *models.ClaudeStreamEvent
	Error       error
}

// Health checks if the Claude API is accessible
func (c *Client) Health(ctx context.Context) error {
	// Simple ping test - we can't actually ping Claude API without making a real request
	// So we just check if we have an API key configured
	if c.apiKey == "" {
		return fmt.Errorf("no API key configured")
	}
	return nil
}
