// Package claude provides mapping between OpenAI and Claude API formats.
package claude

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
)

// MapOpenAIToClaude converts an OpenAI request to Claude format
func MapOpenAIToClaude(openAIReq *models.OpenAIRequest) (*models.ClaudeRequest, error) {
	claudeReq := &models.ClaudeRequest{
		Model:       openAIReq.Model,
		MaxTokens:   4096, // Default
		Stream:      openAIReq.Stream,
		Temperature: openAIReq.Temperature,
		TopP:        openAIReq.TopP,
	}

	// Set max_tokens from OpenAI request
	if openAIReq.MaxTokens != nil && *openAIReq.MaxTokens > 0 {
		claudeReq.MaxTokens = *openAIReq.MaxTokens
	}

	// Convert stop sequences
	if openAIReq.Stop != nil {
		claudeReq.StopSequences = convertStopSequences(openAIReq.Stop)
	}

	// Convert messages
	claudeMessages, systemPrompt, err := convertMessages(openAIReq.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	claudeReq.Messages = claudeMessages
	if systemPrompt != "" {
		claudeReq.System = systemPrompt
	}

	return claudeReq, nil
}

// convertMessages converts OpenAI messages to Claude format
// Returns messages and system prompt (Claude uses separate system field)
func convertMessages(openAIMessages []models.OpenAIMessage) ([]models.ClaudeMessage, string, error) {
	var claudeMessages []models.ClaudeMessage
	var systemPrompt strings.Builder

	for i, msg := range openAIMessages {
		// Extract system messages separately
		if msg.Role == "system" {
			content, err := extractTextContent(msg.Content)
			if err != nil {
				return nil, "", fmt.Errorf("invalid system message content: %w", err)
			}
			if systemPrompt.Len() > 0 {
				systemPrompt.WriteString("\n\n")
			}
			systemPrompt.WriteString(content)
			continue
		}

		// Convert role (Claude only supports "user" and "assistant")
		role := msg.Role
		if role != "user" && role != "assistant" {
			// Tool/function messages become user messages in Claude
			role = "user"
		}

		// Convert content
		contentParts, err := convertContent(msg.Content)
		if err != nil {
			return nil, "", fmt.Errorf("failed to convert message %d content: %w", i, err)
		}

		claudeMessages = append(claudeMessages, models.ClaudeMessage{
			Role:    role,
			Content: contentParts,
		})
	}

	// Claude requires alternating user/assistant messages
	// Ensure we start with a user message
	if len(claudeMessages) > 0 && claudeMessages[0].Role != "user" {
		return nil, "", fmt.Errorf("first message must be from user")
	}

	return claudeMessages, systemPrompt.String(), nil
}

// convertContent converts OpenAI message content to Claude format
func convertContent(content interface{}) ([]models.ClaudeContentPart, error) {
	switch c := content.(type) {
	case string:
		// Simple text content
		return []models.ClaudeContentPart{
			{
				Type: "text",
				Text: c,
			},
		}, nil

	case []interface{}:
		// Multi-modal content (text + images)
		var parts []models.ClaudeContentPart
		for i, part := range c {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid content part %d: expected object", i)
			}

			partType, _ := partMap["type"].(string)
			switch partType {
			case "text":
				text, _ := partMap["text"].(string)
				parts = append(parts, models.ClaudeContentPart{
					Type: "text",
					Text: text,
				})

			case "image_url":
				imageURL, ok := partMap["image_url"].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid image_url in part %d", i)
				}

				url, _ := imageURL["url"].(string)
				if !strings.HasPrefix(url, "data:") {
					return nil, fmt.Errorf("only base64-encoded images are supported")
				}

				// Parse data URL: data:image/png;base64,<data>
				urlParts := strings.SplitN(url, ",", 2)
				if len(urlParts) != 2 {
					return nil, fmt.Errorf("invalid data URL format")
				}

				// Extract media type
				mediaType := "image/png" // default
				if strings.Contains(urlParts[0], ";") {
					mediaTypeParts := strings.Split(urlParts[0], ";")
					if len(mediaTypeParts) > 0 {
						mediaType = strings.TrimPrefix(mediaTypeParts[0], "data:")
					}
				}

				parts = append(parts, models.ClaudeContentPart{
					Type: "image",
					Source: &models.ClaudeImageSource{
						Type:      "base64",
						MediaType: mediaType,
						Data:      urlParts[1],
					},
				})

			default:
				return nil, fmt.Errorf("unsupported content part type: %s", partType)
			}
		}
		return parts, nil

	case []models.OpenAIContentPart:
		// Typed content parts
		var parts []models.ClaudeContentPart
		for _, part := range c {
			switch part.Type {
			case "text":
				parts = append(parts, models.ClaudeContentPart{
					Type: "text",
					Text: part.Text,
				})
			case "image_url":
				if part.ImageURL == nil {
					continue
				}
				url := part.ImageURL.URL
				if !strings.HasPrefix(url, "data:") {
					return nil, fmt.Errorf("only base64-encoded images are supported")
				}

				// Parse data URL
				urlParts := strings.SplitN(url, ",", 2)
				if len(urlParts) != 2 {
					return nil, fmt.Errorf("invalid data URL format")
				}

				mediaType := "image/png"
				if strings.Contains(urlParts[0], ";") {
					mediaTypeParts := strings.Split(urlParts[0], ";")
					if len(mediaTypeParts) > 0 {
						mediaType = strings.TrimPrefix(mediaTypeParts[0], "data:")
					}
				}

				parts = append(parts, models.ClaudeContentPart{
					Type: "image",
					Source: &models.ClaudeImageSource{
						Type:      "base64",
						MediaType: mediaType,
						Data:      urlParts[1],
					},
				})
			}
		}
		return parts, nil

	default:
		return nil, fmt.Errorf("unsupported content type: %T", content)
	}
}

// extractTextContent extracts text from content (helper function)
func extractTextContent(content interface{}) (string, error) {
	switch c := content.(type) {
	case string:
		return c, nil
	case []interface{}:
		var text strings.Builder
		for _, part := range c {
			if partMap, ok := part.(map[string]interface{}); ok {
				if partType, _ := partMap["type"].(string); partType == "text" {
					if textContent, _ := partMap["text"].(string); textContent != "" {
						if text.Len() > 0 {
							text.WriteString(" ")
						}
						text.WriteString(textContent)
					}
				}
			}
		}
		return text.String(), nil
	default:
		return "", fmt.Errorf("cannot extract text from content type: %T", content)
	}
}

// convertStopSequences converts stop parameter to string array
func convertStopSequences(stop interface{}) []string {
	switch s := stop.(type) {
	case string:
		return []string{s}
	case []string:
		return s
	case []interface{}:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// MapClaudeToOpenAI converts a Claude response to OpenAI format
func MapClaudeToOpenAI(claudeResp *models.ClaudeResponse, model string) *models.OpenAIResponse {
	// Extract text content from Claude response
	var content strings.Builder
	for _, part := range claudeResp.Content {
		if part.Type == "text" {
			content.WriteString(part.Text)
		}
	}

	// Map finish reason
	finishReason := mapFinishReason(claudeResp.StopReason)

	return &models.OpenAIResponse{
		ID:      claudeResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []models.OpenAIChoice{
			{
				Index: 0,
				Message: models.OpenAIMessage{
					Role:    "assistant",
					Content: content.String(),
				},
				FinishReason: finishReason,
			},
		},
		Usage: models.OpenAIUsage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}
}

// mapFinishReason maps Claude stop reason to OpenAI finish reason
func mapFinishReason(claudeStopReason string) string {
	switch claudeStopReason {
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

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return "chatcmpl-" + uuid.New().String()
}

// CalculateCost calculates the cost in USD based on token usage
func CalculateCost(model string, inputTokens, outputTokens int) float64 {
	// Claude pricing (as of 2024-01)
	var inputCost, outputCost float64

	switch model {
	case "claude-3-opus-20240229":
		inputCost = 15.00 / 1_000_000  // $15 per 1M input tokens
		outputCost = 75.00 / 1_000_000 // $75 per 1M output tokens
	case "claude-3-5-sonnet-20240620":
		inputCost = 3.00 / 1_000_000   // $3 per 1M input tokens
		outputCost = 15.00 / 1_000_000 // $15 per 1M output tokens
	case "claude-3-sonnet-20240229":
		inputCost = 3.00 / 1_000_000   // $3 per 1M input tokens
		outputCost = 15.00 / 1_000_000 // $15 per 1M output tokens
	case "claude-3-haiku-20240307":
		inputCost = 0.25 / 1_000_000  // $0.25 per 1M input tokens
		outputCost = 1.25 / 1_000_000 // $1.25 per 1M output tokens
	default:
		// Default to Sonnet 3.5 pricing if model is unknown
		inputCost = 3.00 / 1_000_000
		outputCost = 15.00 / 1_000_000
	}

	return (float64(inputTokens) * inputCost) + (float64(outputTokens) * outputCost)
}
