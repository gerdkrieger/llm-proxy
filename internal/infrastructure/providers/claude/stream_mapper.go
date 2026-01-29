// Package claude provides streaming event mapping between Claude and OpenAI formats.
package claude

import (
	"fmt"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
)

// StreamMapper converts Claude streaming events to OpenAI streaming format
type StreamMapper struct {
	requestID     string
	model         string
	contentBuffer string
	inputTokens   int
	outputTokens  int
}

// NewStreamMapper creates a new stream mapper
func NewStreamMapper(requestID, model string) *StreamMapper {
	return &StreamMapper{
		requestID: requestID,
		model:     model,
	}
}

// MapClaudeStreamToOpenAI converts a Claude stream event to OpenAI streaming format
func (m *StreamMapper) MapClaudeStreamToOpenAI(event *models.ClaudeStreamEvent) (*models.OpenAIStreamResponse, error) {
	timestamp := time.Now().Unix()

	switch event.Type {
	case "message_start":
		// First chunk - include role in delta
		return &models.OpenAIStreamResponse{
			ID:      m.requestID,
			Object:  "chat.completion.chunk",
			Created: timestamp,
			Model:   m.model,
			Choices: []models.OpenAIStreamChoice{
				{
					Index: 0,
					Delta: models.OpenAIMessageDelta{
						Role: "assistant",
					},
					FinishReason: nil,
				},
			},
		}, nil

	case "content_block_start":
		// Content block started - no content yet
		return nil, nil

	case "content_block_delta":
		// Extract text delta
		if event.Delta != nil && event.Delta.Text != "" {
			m.contentBuffer += event.Delta.Text
			m.outputTokens++ // Rough estimation

			return &models.OpenAIStreamResponse{
				ID:      m.requestID,
				Object:  "chat.completion.chunk",
				Created: timestamp,
				Model:   m.model,
				Choices: []models.OpenAIStreamChoice{
					{
						Index: 0,
						Delta: models.OpenAIMessageDelta{
							Content: event.Delta.Text,
						},
						FinishReason: nil,
					},
				},
			}, nil
		}
		return nil, nil

	case "content_block_stop":
		// Content block stopped - no action needed
		return nil, nil

	case "message_delta":
		// Message delta - might include usage or stop reason
		if event.Delta != nil && event.Delta.StopReason != "" {
			// Final chunk with finish reason
			finishReason := m.mapStopReason(event.Delta.StopReason)
			return &models.OpenAIStreamResponse{
				ID:      m.requestID,
				Object:  "chat.completion.chunk",
				Created: timestamp,
				Model:   m.model,
				Choices: []models.OpenAIStreamChoice{
					{
						Index:        0,
						Delta:        models.OpenAIMessageDelta{},
						FinishReason: &finishReason,
					},
				},
			}, nil
		}
		return nil, nil

	case "message_stop":
		// Message stopped - send final chunk with usage if available
		if event.Usage != nil {
			m.inputTokens = event.Usage.InputTokens
			m.outputTokens = event.Usage.OutputTokens
		}

		// Send final [DONE] marker
		return nil, fmt.Errorf("[DONE]")

	case "ping":
		// Ping event - ignore
		return nil, nil

	case "error":
		// Error event
		return nil, fmt.Errorf("streaming error from Claude API")

	default:
		// Unknown event type - ignore
		return nil, nil
	}
}

// mapStopReason maps Claude stop reason to OpenAI finish reason
func (m *StreamMapper) mapStopReason(claudeReason string) string {
	switch claudeReason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	default:
		return "stop"
	}
}

// GetUsage returns the accumulated token usage
func (m *StreamMapper) GetUsage() *models.OpenAIUsage {
	return &models.OpenAIUsage{
		PromptTokens:     m.inputTokens,
		CompletionTokens: m.outputTokens,
		TotalTokens:      m.inputTokens + m.outputTokens,
	}
}

// GetContent returns the accumulated content
func (m *StreamMapper) GetContent() string {
	return m.contentBuffer
}
