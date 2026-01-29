// Package models defines data models for LLM providers.
package models

// Claude API Models (based on Anthropic API v1)

// ClaudeRequest represents a request to Claude API
type ClaudeRequest struct {
	Model         string          `json:"model"`
	Messages      []ClaudeMessage `json:"messages"`
	MaxTokens     int             `json:"max_tokens"`
	Temperature   *float64        `json:"temperature,omitempty"`
	TopP          *float64        `json:"top_p,omitempty"`
	TopK          *int            `json:"top_k,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
	Stream        bool            `json:"stream,omitempty"`
	System        string          `json:"system,omitempty"`
	Metadata      interface{}     `json:"metadata,omitempty"`
}

// ClaudeMessage represents a message in Claude API format
type ClaudeMessage struct {
	Role    string              `json:"role"` // "user" or "assistant"
	Content []ClaudeContentPart `json:"content"`
}

// ClaudeContentPart represents a content part (text or image)
type ClaudeContentPart struct {
	Type   string             `json:"type"` // "text" or "image"
	Text   string             `json:"text,omitempty"`
	Source *ClaudeImageSource `json:"source,omitempty"`
}

// ClaudeImageSource represents an image source
type ClaudeImageSource struct {
	Type      string `json:"type"` // "base64"
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// ClaudeResponse represents a response from Claude API
type ClaudeResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"` // "message"
	Role         string              `json:"role"` // "assistant"
	Content      []ClaudeContentPart `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason,omitempty"`
	StopSequence string              `json:"stop_sequence,omitempty"`
	Usage        ClaudeUsage         `json:"usage"`
}

// ClaudeUsage represents token usage information
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeErrorResponse represents an error response from Claude API
type ClaudeErrorResponse struct {
	Type  string      `json:"type"` // "error"
	Error ClaudeError `json:"error"`
}

// ClaudeError represents an error details
type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ClaudeStreamEvent represents a streaming event from Claude API
type ClaudeStreamEvent struct {
	Type         string              `json:"type"`
	Index        int                 `json:"index,omitempty"`
	Delta        *ClaudeContentDelta `json:"delta,omitempty"`
	Message      *ClaudeResponse     `json:"message,omitempty"`
	ContentBlock *ClaudeContentPart  `json:"content_block,omitempty"`
	Usage        *ClaudeUsage        `json:"usage,omitempty"`
}

// ClaudeContentDelta represents a content delta in streaming
type ClaudeContentDelta struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}
