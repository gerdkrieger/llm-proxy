package claude

import (
	"testing"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test MapRequest - Simple User Message
func TestMapRequest_SimpleUserMessage(t *testing.T) {
	req := &models.OpenAIRequest{
		Model: "claude-3-opus-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello, world!"},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	claudeReq := MapRequest(req)

	assert.NotNil(t, claudeReq)
	assert.Equal(t, "claude-3-opus-20240229", claudeReq.Model)
	assert.Equal(t, 1000, claudeReq.MaxTokens)
	assert.Equal(t, 0.7, claudeReq.Temperature)
	assert.Len(t, claudeReq.Messages, 1)
	assert.Equal(t, "user", claudeReq.Messages[0].Role)
	assert.Equal(t, "Hello, world!", claudeReq.Messages[0].Content)
}

// Test MapRequest - With System Message
func TestMapRequest_WithSystemMessage(t *testing.T) {
	req := &models.OpenAIRequest{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is 2+2?"},
		},
		MaxTokens: 500,
	}

	claudeReq := MapRequest(req)

	assert.NotNil(t, claudeReq)
	assert.Equal(t, "You are a helpful assistant.", claudeReq.System)
	assert.Len(t, claudeReq.Messages, 1)
	assert.Equal(t, "user", claudeReq.Messages[0].Role)
	assert.Equal(t, "What is 2+2?", claudeReq.Messages[0].Content)
}

// Test MapRequest - Conversation History
func TestMapRequest_ConversationHistory(t *testing.T) {
	req := &models.OpenAIRequest{
		Model: "claude-3-opus-20240229",
		Messages: []models.Message{
			{Role: "system", Content: "You are a math tutor."},
			{Role: "user", Content: "What is 5+5?"},
			{Role: "assistant", Content: "5+5 equals 10."},
			{Role: "user", Content: "What is 10+10?"},
		},
		MaxTokens: 100,
	}

	claudeReq := MapRequest(req)

	assert.NotNil(t, claudeReq)
	assert.Equal(t, "You are a math tutor.", claudeReq.System)
	assert.Len(t, claudeReq.Messages, 3) // user, assistant, user
	assert.Equal(t, "user", claudeReq.Messages[0].Role)
	assert.Equal(t, "assistant", claudeReq.Messages[1].Role)
	assert.Equal(t, "user", claudeReq.Messages[2].Role)
}

// Test MapRequest - Default Values
func TestMapRequest_DefaultValues(t *testing.T) {
	req := &models.OpenAIRequest{
		Model: "claude-3-opus-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Test"},
		},
		// No Temperature, MaxTokens, TopP specified
	}

	claudeReq := MapRequest(req)

	assert.NotNil(t, claudeReq)
	// Check that values are passed through (may be 0 if not set)
	assert.Equal(t, req.Temperature, claudeReq.Temperature)
	assert.Equal(t, req.MaxTokens, claudeReq.MaxTokens)
}

// Test MapResponse - Simple Response
func TestMapResponse_Simple(t *testing.T) {
	claudeResp := &ClaudeResponse{
		ID:    "msg_123456",
		Model: "claude-3-opus-20240229",
		Role:  "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "Hello! How can I help you?"},
		},
		Usage: ClaudeUsage{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}

	openaiResp := MapResponse(claudeResp)

	assert.NotNil(t, openaiResp)
	assert.Equal(t, "msg_123456", openaiResp.ID)
	assert.Equal(t, "chat.completion", openaiResp.Object)
	assert.Equal(t, "claude-3-opus-20240229", openaiResp.Model)
	assert.Len(t, openaiResp.Choices, 1)
	assert.Equal(t, "assistant", openaiResp.Choices[0].Message.Role)
	assert.Equal(t, "Hello! How can I help you?", openaiResp.Choices[0].Message.Content)
	assert.Equal(t, "stop", openaiResp.Choices[0].FinishReason)
	assert.Equal(t, 10, openaiResp.Usage.PromptTokens)
	assert.Equal(t, 20, openaiResp.Usage.CompletionTokens)
	assert.Equal(t, 30, openaiResp.Usage.TotalTokens)
}

// Test MapResponse - Multiple Content Blocks
func TestMapResponse_MultipleContentBlocks(t *testing.T) {
	claudeResp := &ClaudeResponse{
		ID:    "msg_789",
		Model: "claude-3-sonnet-20240229",
		Role:  "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "First part. "},
			{Type: "text", Text: "Second part."},
		},
		Usage: ClaudeUsage{
			InputTokens:  15,
			OutputTokens: 25,
		},
	}

	openaiResp := MapResponse(claudeResp)

	assert.NotNil(t, openaiResp)
	assert.Len(t, openaiResp.Choices, 1)
	// Content blocks should be concatenated
	assert.Equal(t, "First part. Second part.", openaiResp.Choices[0].Message.Content)
}

// Test MapResponse - Empty Content
func TestMapResponse_EmptyContent(t *testing.T) {
	claudeResp := &ClaudeResponse{
		ID:      "msg_empty",
		Model:   "claude-3-opus-20240229",
		Role:    "assistant",
		Content: []ContentBlock{},
		Usage: ClaudeUsage{
			InputTokens:  5,
			OutputTokens: 0,
		},
	}

	openaiResp := MapResponse(claudeResp)

	assert.NotNil(t, openaiResp)
	assert.Len(t, openaiResp.Choices, 1)
	assert.Empty(t, openaiResp.Choices[0].Message.Content)
}

// Test CalculateCost - Claude 3 Opus
func TestCalculateCost_Claude3Opus(t *testing.T) {
	usage := ClaudeUsage{
		InputTokens:  1000,
		OutputTokens: 2000,
	}

	cost := CalculateCost("claude-3-opus-20240229", usage)

	// Opus: $15/$75 per million tokens (input/output)
	expectedCost := (1000.0/1000000.0)*15.0 + (2000.0/1000000.0)*75.0
	assert.InDelta(t, expectedCost, cost, 0.00001)
	assert.Greater(t, cost, 0.0)
}

// Test CalculateCost - Claude 3 Sonnet
func TestCalculateCost_Claude3Sonnet(t *testing.T) {
	usage := ClaudeUsage{
		InputTokens:  1000,
		OutputTokens: 2000,
	}

	cost := CalculateCost("claude-3-sonnet-20240229", usage)

	// Sonnet: $3/$15 per million tokens
	expectedCost := (1000.0/1000000.0)*3.0 + (2000.0/1000000.0)*15.0
	assert.InDelta(t, expectedCost, cost, 0.00001)
	assert.Greater(t, cost, 0.0)
}

// Test CalculateCost - Claude 3 Haiku
func TestCalculateCost_Claude3Haiku(t *testing.T) {
	usage := ClaudeUsage{
		InputTokens:  1000,
		OutputTokens: 2000,
	}

	cost := CalculateCost("claude-3-haiku-20240307", usage)

	// Haiku: $0.25/$1.25 per million tokens
	expectedCost := (1000.0/1000000.0)*0.25 + (2000.0/1000000.0)*1.25
	assert.InDelta(t, expectedCost, cost, 0.00001)
	assert.Greater(t, cost, 0.0)
}

// Test CalculateCost - Unknown Model (Default)
func TestCalculateCost_UnknownModel(t *testing.T) {
	usage := ClaudeUsage{
		InputTokens:  1000,
		OutputTokens: 2000,
	}

	cost := CalculateCost("unknown-model", usage)

	// Should use Sonnet pricing as default
	expectedCost := (1000.0/1000000.0)*3.0 + (2000.0/1000000.0)*15.0
	assert.InDelta(t, expectedCost, cost, 0.00001)
}

// Test CalculateCost - Zero Usage
func TestCalculateCost_ZeroUsage(t *testing.T) {
	usage := ClaudeUsage{
		InputTokens:  0,
		OutputTokens: 0,
	}

	cost := CalculateCost("claude-3-opus-20240229", usage)
	assert.Equal(t, 0.0, cost)
}

// Test MapRequest - Nil Input
func TestMapRequest_NilInput(t *testing.T) {
	require.NotPanics(t, func() {
		claudeReq := MapRequest(nil)
		assert.Nil(t, claudeReq)
	})
}

// Test MapResponse - Nil Input
func TestMapResponse_NilInput(t *testing.T) {
	require.NotPanics(t, func() {
		openaiResp := MapResponse(nil)
		assert.Nil(t, openaiResp)
	})
}

// Benchmark MapRequest
func BenchmarkMapRequest(b *testing.B) {
	req := &models.OpenAIRequest{
		Model: "claude-3-opus-20240229",
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MapRequest(req)
	}
}

// Benchmark MapResponse
func BenchmarkMapResponse(b *testing.B) {
	claudeResp := &ClaudeResponse{
		ID:    "msg_benchmark",
		Model: "claude-3-opus-20240229",
		Role:  "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "This is a test response for benchmarking purposes."},
		},
		Usage: ClaudeUsage{
			InputTokens:  100,
			OutputTokens: 200,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MapResponse(claudeResp)
	}
}

// Benchmark CalculateCost
func BenchmarkCalculateCost(b *testing.B) {
	usage := ClaudeUsage{
		InputTokens:  10000,
		OutputTokens: 20000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCost("claude-3-opus-20240229", usage)
	}
}
