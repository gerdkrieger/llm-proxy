// Package models defines OpenAI-compatible API models.
package models

// OpenAI API Models (compatible with OpenAI API v1)

// OpenAIRequest represents an OpenAI chat completion request
type OpenAIRequest struct {
	Model               string                `json:"model"`
	Messages            []OpenAIMessage       `json:"messages"`
	Temperature         *float64              `json:"temperature,omitempty"`
	TopP                *float64              `json:"top_p,omitempty"`
	N                   *int                  `json:"n,omitempty"`
	Stream              bool                  `json:"stream,omitempty"`
	StreamOptions       *OpenAIStreamOptions  `json:"stream_options,omitempty"`
	Stop                interface{}           `json:"stop,omitempty"` // string or []string
	MaxTokens           *int                  `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int                  `json:"max_completion_tokens,omitempty"`
	PresencePenalty     *float64              `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64              `json:"frequency_penalty,omitempty"`
	LogitBias           map[string]float64    `json:"logit_bias,omitempty"`
	User                string                `json:"user,omitempty"`
	ResponseFormat      *OpenAIResponseFormat `json:"response_format,omitempty"`
	Tools               []OpenAITool          `json:"tools,omitempty"`
	ToolChoice          interface{}           `json:"tool_choice,omitempty"`
}

// OpenAIStreamOptions configures streaming behavior
type OpenAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role       string      `json:"role"`    // "system", "user", "assistant", "tool"
	Content    interface{} `json:"content"` // string or []ContentPart
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
}

// OpenAIContentPart represents a content part (for multi-modal messages)
type OpenAIContentPart struct {
	Type     string          `json:"type"` // "text" or "image_url"
	Text     string          `json:"text,omitempty"`
	ImageURL *OpenAIImageURL `json:"image_url,omitempty"`
}

// OpenAIImageURL represents an image URL
type OpenAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"` // "auto", "low", "high"
}

// OpenAIResponseFormat specifies response format
type OpenAIResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// OpenAITool represents a function tool
type OpenAITool struct {
	Type     string             `json:"type"` // "function"
	Function OpenAIToolFunction `json:"function"`
}

// OpenAIToolFunction represents a function definition
type OpenAIToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"` // JSON Schema
}

// ToolCall represents a tool call
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction represents a function call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// OpenAIResponse represents an OpenAI chat completion response
type OpenAIResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"` // "chat.completion"
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Choices           []OpenAIChoice `json:"choices"`
	Usage             OpenAIUsage    `json:"usage"`
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
}

// OpenAIChoice represents a completion choice
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"` // "stop", "length", "tool_calls", "content_filter"
	LogProbs     interface{}   `json:"logprobs,omitempty"`
}

// OpenAIUsage represents token usage
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIStreamResponse represents a streaming chunk
type OpenAIStreamResponse struct {
	ID                string               `json:"id"`
	Object            string               `json:"object"` // "chat.completion.chunk"
	Created           int64                `json:"created"`
	Model             string               `json:"model"`
	Choices           []OpenAIStreamChoice `json:"choices"`
	Usage             *OpenAIUsage         `json:"usage,omitempty"` // Present in final chunk when stream_options.include_usage is true
	SystemFingerprint string               `json:"system_fingerprint,omitempty"`
}

// OpenAIStreamChoice represents a streaming choice
type OpenAIStreamChoice struct {
	Index        int                `json:"index"`
	Delta        OpenAIMessageDelta `json:"delta"`
	FinishReason *string            `json:"finish_reason"`
	LogProbs     interface{}        `json:"logprobs,omitempty"`
}

// OpenAIMessageDelta represents a message delta in streaming
type OpenAIMessageDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// OpenAIErrorResponse represents an error response
type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

// OpenAIError represents error details
type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param,omitempty"`
	Code    interface{} `json:"code,omitempty"`
}

// OpenAIModel represents a model info
type OpenAIModel struct {
	ID         string        `json:"id"`
	Object     string        `json:"object"` // "model"
	Created    int64         `json:"created"`
	OwnedBy    string        `json:"owned_by"`
	Permission []interface{} `json:"permission,omitempty"`
	Root       string        `json:"root,omitempty"`
	Parent     string        `json:"parent,omitempty"`
}

// OpenAIModelList represents a list of models
type OpenAIModelList struct {
	Object string        `json:"object"` // "list"
	Data   []OpenAIModel `json:"data"`
}
