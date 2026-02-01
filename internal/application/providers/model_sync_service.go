// Package providers provides model synchronization service.
package providers

import (
	"context"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ModelSyncService handles automatic model synchronization to database
type ModelSyncService struct {
	providerModelRepo *repositories.ProviderModelRepository
	logger            *logger.Logger
}

// NewModelSyncService creates a new model sync service
func NewModelSyncService(
	providerModelRepo *repositories.ProviderModelRepository,
	log *logger.Logger,
) *ModelSyncService {
	return &ModelSyncService{
		providerModelRepo: providerModelRepo,
		logger:            log,
	}
}

// ModelDefinition represents a model with its metadata
type ModelDefinition struct {
	ID           string
	Name         string
	ProviderID   string
	Capabilities []string
	Description  string
}

// GetAllKnownModels returns all models we support (based on official API documentation)
func GetAllKnownModels() []ModelDefinition {
	return []ModelDefinition{
		// ===================================================
		// CLAUDE MODELS (Anthropic)
		// Source: https://docs.anthropic.com/en/docs/about-claude/models
		// ===================================================

		// Claude 4.5 Series (LATEST - Jan 2026)
		{
			ID:           "claude-sonnet-4-5",
			Name:         "Claude Sonnet 4.5 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context", "1m_context_beta"},
			Description:  "Smart model for complex agents and coding (auto-updated alias)",
		},
		{
			ID:           "claude-sonnet-4-5-20250929",
			Name:         "Claude Sonnet 4.5 (Sep 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context", "1m_context_beta"},
			Description:  "Smart model for complex agents and coding tasks",
		},
		{
			ID:           "claude-haiku-4-5",
			Name:         "Claude Haiku 4.5 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Fastest model with near-frontier intelligence (auto-updated alias)",
		},
		{
			ID:           "claude-haiku-4-5-20251001",
			Name:         "Claude Haiku 4.5 (Oct 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Fastest model with near-frontier intelligence",
		},
		{
			ID:           "claude-opus-4-5",
			Name:         "Claude Opus 4.5 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Premium model combining maximum intelligence with practical performance (auto-updated alias)",
		},
		{
			ID:           "claude-opus-4-5-20251101",
			Name:         "Claude Opus 4.5 (Nov 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Premium model combining maximum intelligence with practical performance",
		},

		// Claude 4.1 and 4 Series (Legacy but available)
		{
			ID:           "claude-opus-4-1",
			Name:         "Claude Opus 4.1 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 4.1 Opus (auto-updated alias)",
		},
		{
			ID:           "claude-opus-4-1-20250805",
			Name:         "Claude Opus 4.1 (Aug 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 4.1 Opus release",
		},
		{
			ID:           "claude-sonnet-4-0",
			Name:         "Claude Sonnet 4 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context", "1m_context_beta"},
			Description:  "Legacy Claude 4 Sonnet (auto-updated alias)",
		},
		{
			ID:           "claude-sonnet-4-20250514",
			Name:         "Claude Sonnet 4 (May 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context", "1m_context_beta"},
			Description:  "Legacy Claude 4 Sonnet release",
		},
		{
			ID:           "claude-3-7-sonnet-latest",
			Name:         "Claude 3.7 Sonnet (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 3.7 Sonnet (auto-updated alias)",
		},
		{
			ID:           "claude-3-7-sonnet-20250219",
			Name:         "Claude 3.7 Sonnet (Feb 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 3.7 Sonnet release",
		},
		{
			ID:           "claude-opus-4-0",
			Name:         "Claude Opus 4 (Alias)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 4 Opus (auto-updated alias)",
		},
		{
			ID:           "claude-opus-4-20250514",
			Name:         "Claude Opus 4 (May 2025)",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context"},
			Description:  "Legacy Claude 4 Opus release",
		},

		// Claude 3 Series (Legacy)
		{
			ID:           "claude-3-haiku-20240307",
			Name:         "Claude 3 Haiku",
			ProviderID:   "claude",
			Capabilities: []string{"vision", "200k_context"},
			Description:  "Legacy fast and affordable Claude 3 model",
		},

		// ===================================================
		// OPENAI MODELS
		// Source: https://platform.openai.com/docs/models
		// ===================================================

		// GPT-5 Series (LATEST - Jan 2026)
		{
			ID:           "gpt-5.2",
			Name:         "GPT-5.2",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning", "coding"},
			Description:  "Best model for coding and agentic tasks across industries",
		},
		{
			ID:           "gpt-5.2-pro",
			Name:         "GPT-5.2 Pro",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning", "coding"},
			Description:  "GPT-5.2 with more compute for smarter and more precise responses",
		},
		{
			ID:           "gpt-5.2-codex",
			Name:         "GPT-5.2 Codex",
			ProviderID:   "openai",
			Capabilities: []string{"coding", "reasoning", "agentic"},
			Description:  "Most intelligent coding model optimized for long-horizon agentic coding tasks",
		},
		{
			ID:           "gpt-5.1",
			Name:         "GPT-5.1",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning", "coding"},
			Description:  "Intelligent reasoning model with configurable reasoning effort",
		},
		{
			ID:           "gpt-5.1-codex",
			Name:         "GPT-5.1 Codex",
			ProviderID:   "openai",
			Capabilities: []string{"coding", "reasoning", "agentic"},
			Description:  "GPT-5.1 optimized for agentic coding in Codex",
		},
		{
			ID:           "gpt-5.1-codex-max",
			Name:         "GPT-5.1 Codex Max",
			ProviderID:   "openai",
			Capabilities: []string{"coding", "reasoning", "agentic"},
			Description:  "GPT-5.1 Codex optimized for long running tasks",
		},
		{
			ID:           "gpt-5.1-codex-mini",
			Name:         "GPT-5.1 Codex Mini",
			ProviderID:   "openai",
			Capabilities: []string{"coding", "reasoning"},
			Description:  "Smaller, more cost-effective version of GPT-5.1 Codex",
		},
		{
			ID:           "gpt-5",
			Name:         "GPT-5",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning", "coding"},
			Description:  "Previous intelligent reasoning model with configurable reasoning effort",
		},
		{
			ID:           "gpt-5-pro",
			Name:         "GPT-5 Pro",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning"},
			Description:  "GPT-5 with more compute for better responses",
		},
		{
			ID:           "gpt-5-codex",
			Name:         "GPT-5 Codex",
			ProviderID:   "openai",
			Capabilities: []string{"coding", "reasoning", "agentic"},
			Description:  "GPT-5 optimized for agentic coding in Codex",
		},
		{
			ID:           "gpt-5-mini",
			Name:         "GPT-5 Mini",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning"},
			Description:  "Faster, cost-efficient version of GPT-5 for well-defined tasks",
		},
		{
			ID:           "gpt-5-nano",
			Name:         "GPT-5 Nano",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "Fastest, most cost-efficient version of GPT-5",
		},

		// GPT-4.1 Series (Latest Non-Reasoning)
		{
			ID:           "gpt-4.1",
			Name:         "GPT-4.1",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "Smartest non-reasoning model",
		},
		{
			ID:           "gpt-4.1-mini",
			Name:         "GPT-4.1 Mini",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "Smaller, faster version of GPT-4.1",
		},
		{
			ID:           "gpt-4.1-nano",
			Name:         "GPT-4.1 Nano",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "Fastest, most cost-efficient version of GPT-4.1",
		},

		// o-Series Reasoning Models (Succeeded by GPT-5)
		{
			ID:           "o3",
			Name:         "o3",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Reasoning model for complex tasks, succeeded by GPT-5",
		},
		{
			ID:           "o3-pro",
			Name:         "o3 Pro",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "o3 with more compute for better responses",
		},
		{
			ID:           "o3-mini",
			Name:         "o3 Mini",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Smaller alternative to o3",
		},
		{
			ID:           "o4-mini",
			Name:         "o4 Mini",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Fast, cost-efficient reasoning model, succeeded by GPT-5 mini",
		},
		{
			ID:           "o1",
			Name:         "o1",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Previous full o-series reasoning model",
		},
		{
			ID:           "o1-pro",
			Name:         "o1 Pro",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "o1 with more compute for better responses",
		},
		{
			ID:           "o1-mini",
			Name:         "o1 Mini",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Smaller alternative to o1",
		},
		{
			ID:           "o1-preview",
			Name:         "o1 Preview",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning"},
			Description:  "Preview of first o-series reasoning model",
		},

		// Deep Research Models
		{
			ID:           "o3-deep-research",
			Name:         "o3 Deep Research",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning", "research"},
			Description:  "Most powerful deep research model",
		},
		{
			ID:           "o4-mini-deep-research",
			Name:         "o4 Mini Deep Research",
			ProviderID:   "openai",
			Capabilities: []string{"reasoning", "research"},
			Description:  "Faster, more affordable deep research model",
		},

		// GPT-4o Series (Still Available)
		{
			ID:           "gpt-4o",
			Name:         "GPT-4o",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "audio"},
			Description:  "Fast, intelligent, flexible GPT model",
		},
		{
			ID:           "gpt-4o-2024-11-20",
			Name:         "GPT-4o (Nov 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "structured_outputs"},
			Description:  "GPT-4o snapshot from November 2024",
		},
		{
			ID:           "gpt-4o-2024-08-06",
			Name:         "GPT-4o (Aug 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode", "structured_outputs"},
			Description:  "GPT-4o snapshot from August 2024 with structured outputs",
		},
		{
			ID:           "gpt-4o-2024-05-13",
			Name:         "GPT-4o (May 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "First GPT-4o release from May 2024",
		},
		{
			ID:           "gpt-4o-mini",
			Name:         "GPT-4o Mini",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "Affordable and intelligent small model for fast, lightweight tasks",
		},
		{
			ID:           "gpt-4o-mini-2024-07-18",
			Name:         "GPT-4o Mini (Jul 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "GPT-4o Mini snapshot from July 2024",
		},

		// Realtime and Audio Models
		{
			ID:           "gpt-realtime",
			Name:         "GPT Realtime",
			ProviderID:   "openai",
			Capabilities: []string{"realtime", "audio", "text"},
			Description:  "Model capable of realtime text and audio inputs and outputs",
		},
		{
			ID:           "gpt-realtime-mini",
			Name:         "GPT Realtime Mini",
			ProviderID:   "openai",
			Capabilities: []string{"realtime", "audio", "text"},
			Description:  "Cost-efficient version of GPT Realtime",
		},
		{
			ID:           "gpt-audio",
			Name:         "GPT Audio",
			ProviderID:   "openai",
			Capabilities: []string{"audio"},
			Description:  "For audio inputs and outputs with Chat Completions API",
		},
		{
			ID:           "gpt-audio-mini",
			Name:         "GPT Audio Mini",
			ProviderID:   "openai",
			Capabilities: []string{"audio"},
			Description:  "Cost-efficient version of GPT Audio",
		},
		{
			ID:           "gpt-4o-audio-preview",
			Name:         "GPT-4o Audio Preview",
			ProviderID:   "openai",
			Capabilities: []string{"audio", "vision"},
			Description:  "GPT-4o models capable of audio inputs and outputs",
		},
		{
			ID:           "gpt-4o-mini-audio-preview",
			Name:         "GPT-4o Mini Audio Preview",
			ProviderID:   "openai",
			Capabilities: []string{"audio"},
			Description:  "Smaller model capable of audio inputs and outputs",
		},
		{
			ID:           "gpt-4o-realtime-preview",
			Name:         "GPT-4o Realtime Preview",
			ProviderID:   "openai",
			Capabilities: []string{"realtime", "audio", "text"},
			Description:  "Model capable of realtime text and audio inputs and outputs",
		},
		{
			ID:           "gpt-4o-mini-realtime-preview",
			Name:         "GPT-4o Mini Realtime Preview",
			ProviderID:   "openai",
			Capabilities: []string{"realtime", "audio", "text"},
			Description:  "Smaller realtime model for text and audio inputs and outputs",
		},

		// GPT-4 Turbo Series (Legacy but Available)
		{
			ID:           "gpt-4-turbo",
			Name:         "GPT-4 Turbo",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "Older high-intelligence GPT model with vision",
		},
		{
			ID:           "gpt-4-turbo-2024-04-09",
			Name:         "GPT-4 Turbo (Apr 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"vision", "function_calling", "json_mode"},
			Description:  "GPT-4 Turbo snapshot from April 2024",
		},
		{
			ID:           "gpt-4-turbo-preview",
			Name:         "GPT-4 Turbo Preview",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "Older fast GPT model preview",
		},
		{
			ID:           "gpt-4-0125-preview",
			Name:         "GPT-4 Turbo (Jan 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "GPT-4 Turbo from January 2024",
		},
		{
			ID:           "gpt-4-1106-preview",
			Name:         "GPT-4 Turbo (Nov 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "GPT-4 Turbo from November 2023",
		},
		{
			ID:           "gpt-4-vision-preview",
			Name:         "GPT-4 Vision Preview",
			ProviderID:   "openai",
			Capabilities: []string{"vision"},
			Description:  "GPT-4 with vision capabilities (preview)",
		},
		{
			ID:           "gpt-4-1106-vision-preview",
			Name:         "GPT-4 Vision (Nov 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"vision"},
			Description:  "GPT-4 Vision from November 2023",
		},

		// GPT-4 Series (Legacy)
		{
			ID:           "gpt-4",
			Name:         "GPT-4",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "Older high-intelligence GPT model (8K context)",
		},
		{
			ID:           "gpt-4-0613",
			Name:         "GPT-4 (Jun 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-4 snapshot from June 2023",
		},
		{
			ID:           "gpt-4-32k",
			Name:         "GPT-4 32K",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-4 with extended 32K context window",
		},
		{
			ID:           "gpt-4-32k-0613",
			Name:         "GPT-4 32K (Jun 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-4 32K snapshot from June 2023",
		},

		// GPT-3.5 Turbo Series (Legacy)
		{
			ID:           "gpt-3.5-turbo",
			Name:         "GPT-3.5 Turbo",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "Legacy GPT model for cheaper chat and non-chat tasks",
		},
		{
			ID:           "gpt-3.5-turbo-0125",
			Name:         "GPT-3.5 Turbo (Jan 2024)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "GPT-3.5 Turbo from January 2024",
		},
		{
			ID:           "gpt-3.5-turbo-1106",
			Name:         "GPT-3.5 Turbo (Nov 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling", "json_mode"},
			Description:  "GPT-3.5 Turbo from November 2023",
		},
		{
			ID:           "gpt-3.5-turbo-16k",
			Name:         "GPT-3.5 Turbo 16K",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-3.5 with extended 16K context",
		},
		{
			ID:           "gpt-3.5-turbo-0613",
			Name:         "GPT-3.5 Turbo (Jun 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-3.5 snapshot from June 2023",
		},
		{
			ID:           "gpt-3.5-turbo-16k-0613",
			Name:         "GPT-3.5 Turbo 16K (Jun 2023)",
			ProviderID:   "openai",
			Capabilities: []string{"function_calling"},
			Description:  "GPT-3.5 16K snapshot from June 2023",
		},
	}
}

// SyncModelsToDatabase ensures all known models exist in database with default enabled status
func (s *ModelSyncService) SyncModelsToDatabase(ctx context.Context) error {
	s.logger.Info("Starting model synchronization to database...")

	allModels := GetAllKnownModels()
	syncedCount := 0
	errorCount := 0

	for _, model := range allModels {
		// Check if model already exists
		existing, err := s.providerModelRepo.GetByProviderAndModel(ctx, model.ProviderID, model.ID)

		if err != nil || existing == nil {
			// Model doesn't exist, create it as enabled by default
			desc := model.Description
			dbModel := &repositories.ProviderModel{
				ID:          uuid.New(),
				ProviderID:  model.ProviderID,
				ModelID:     model.ID,
				ModelName:   model.Name,
				Enabled:     true, // Default: enabled
				Description: &desc,
				Capabilities: map[string]interface{}{
					"features": model.Capabilities,
				},
				Pricing: map[string]interface{}{},
			}

			if err := s.providerModelRepo.Create(ctx, dbModel); err != nil {
				s.logger.Warnf("Failed to sync model %s: %v", model.ID, err)
				errorCount++
				continue
			}

			s.logger.Debugf("Synced new model: %s (%s)", model.Name, model.ID)
			syncedCount++
		} else {
			s.logger.Debugf("Model already exists: %s (enabled: %v)", model.ID, existing.Enabled)
		}
	}

	s.logger.Infof("Model sync completed: %d new models synced, %d errors, %d total models",
		syncedCount, errorCount, len(allModels))

	return nil
}
