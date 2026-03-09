// Package abacus provides mapping between OpenAI and Abacus.ai formats.
package abacus

import (
	"fmt"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
)

// MapOpenAIToAbacus converts an OpenAI chat completion request to Abacus.ai format
func MapOpenAIToAbacus(openAIReq *models.OpenAIRequest, deploymentID string) (*AbacusRequest, error) {
	if openAIReq == nil {
		return nil, fmt.Errorf("openAI request is nil")
	}

	if deploymentID == "" {
		return nil, fmt.Errorf("deployment ID is required")
	}

	// Convert messages
	abacusMessages := make([]AbacusMessage, 0, len(openAIReq.Messages))
	var systemMessage string

	for _, msg := range openAIReq.Messages {
		// Convert content to string (it can be string or []ContentPart)
		content := ""
		switch v := msg.Content.(type) {
		case string:
			content = v
		case []interface{}:
			// For multi-modal, extract text parts only
			for _, part := range v {
				if partMap, ok := part.(map[string]interface{}); ok {
					if text, ok := partMap["text"].(string); ok {
						content += text
					}
				}
			}
		}

		switch msg.Role {
		case "system":
			// Abacus.ai uses separate systemMessage field
			systemMessage = content
		case "user", "assistant":
			abacusMessages = append(abacusMessages, AbacusMessage{
				Role:    msg.Role,
				Content: content,
			})
		default:
			// Skip unknown roles
			continue
		}
	}

	// Build Abacus.ai request
	abacusReq := &AbacusRequest{
		DeploymentID: deploymentID,
		Messages:     abacusMessages,
		Stream:       openAIReq.Stream,
	}

	// Set system message if present
	if systemMessage != "" {
		abacusReq.SystemMessage = systemMessage
	}

	// Map temperature (both use 0.0-1.0 range, but Abacus.ai may have different defaults)
	if openAIReq.Temperature != nil {
		abacusReq.Temperature = *openAIReq.Temperature
	}

	// Map max_tokens
	if openAIReq.MaxTokens != nil {
		abacusReq.MaxTokens = *openAIReq.MaxTokens
	}

	// Map top_p
	if openAIReq.TopP != nil {
		abacusReq.TopP = *openAIReq.TopP
	}

	// Set LLM name from model field (Abacus.ai uses LLM names like "gpt-4", "claude-3-opus")
	if openAIReq.Model != "" {
		abacusReq.LLMName = openAIReq.Model
	}

	return abacusReq, nil
}

// MapAbacusToOpenAI converts an Abacus.ai response to OpenAI format
func MapAbacusToOpenAI(abacusResp *AbacusResponse, model string, stream bool) (*models.OpenAIResponse, error) {
	if abacusResp == nil {
		return nil, fmt.Errorf("abacus response is nil")
	}

	if !abacusResp.Success {
		return nil, fmt.Errorf("abacus.ai error: %s", abacusResp.Error)
	}

	// Extract assistant message from result
	var assistantMessage string
	if len(abacusResp.Result.Messages) > 0 {
		// Get last message (typically the assistant's response)
		lastMsg := abacusResp.Result.Messages[len(abacusResp.Result.Messages)-1]
		if lastMsg.Role == "assistant" {
			assistantMessage = lastMsg.Content
		} else {
			// If last message is not assistant, search for assistant message
			for i := len(abacusResp.Result.Messages) - 1; i >= 0; i-- {
				if abacusResp.Result.Messages[i].Role == "assistant" {
					assistantMessage = abacusResp.Result.Messages[i].Content
					break
				}
			}
		}
	}

	// Build OpenAI-compatible response
	openAIResp := &models.OpenAIResponse{
		ID:      abacusResp.Result.ConversationID, // Use conversation ID as response ID
		Object:  "chat.completion",
		Created: 0, // Abacus.ai doesn't provide timestamp, set to 0
		Model:   model,
		Choices: []models.OpenAIChoice{
			{
				Index: 0,
				Message: models.OpenAIMessage{
					Role:    "assistant",
					Content: assistantMessage,
				},
				FinishReason: "stop",
			},
		},
		Usage: models.OpenAIUsage{
			PromptTokens:     0, // Abacus.ai doesn't provide token counts in standard response
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	return openAIResp, nil
}

// ExtractDeploymentID extracts deployment ID from model name or config
// Model name format: "abacus:deployment-id" or just use config default
func ExtractDeploymentID(model string, defaultDeploymentID string) string {
	// Check if model contains deployment ID in format "abacus:deployment-id"
	if len(model) > 7 && model[:7] == "abacus:" {
		return model[7:]
	}

	// Return default deployment ID from config
	return defaultDeploymentID
}
