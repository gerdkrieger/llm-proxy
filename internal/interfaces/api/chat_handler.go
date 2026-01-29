// Package api provides chat completion HTTP handlers.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/application/caching"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/claude"
	mw "github.com/llm-proxy/llm-proxy/internal/interfaces/middleware"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
	providerManager *providers.ProviderManager
	requestLogRepo  *repositories.RequestLogRepository
	clientRepo      *repositories.OAuthClientRepository
	cacheService    *caching.Service
	logger          *logger.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	providerManager *providers.ProviderManager,
	requestLogRepo *repositories.RequestLogRepository,
	clientRepo *repositories.OAuthClientRepository,
	cacheService *caching.Service,
	log *logger.Logger,
) *ChatHandler {
	return &ChatHandler{
		providerManager: providerManager,
		requestLogRepo:  requestLogRepo,
		clientRepo:      clientRepo,
		cacheService:    cacheService,
		logger:          log,
	}
}

// CreateCompletion handles chat completion requests
// POST /v1/chat/completions
func (h *ChatHandler) CreateCompletion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Get client ID from context (set by OAuth middleware)
	clientID := mw.GetClientID(ctx)
	requestID := middleware.GetReqID(ctx)

	h.logger.Infof("Chat completion request from client: %s", clientID)

	// Parse OpenAI request
	var openAIReq models.OpenAIRequest
	if err := json.NewDecoder(r.Body).Decode(&openAIReq); err != nil {
		h.logRequest(ctx, clientID, requestID, "", http.StatusBadRequest, 0, 0, 0, startTime, err)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if openAIReq.Model == "" {
		h.logRequest(ctx, clientID, requestID, "", http.StatusBadRequest, 0, 0, 0, startTime, nil)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}

	if len(openAIReq.Messages) == 0 {
		h.logRequest(ctx, clientID, requestID, openAIReq.Model, http.StatusBadRequest, 0, 0, 0, startTime, nil)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "messages are required")
		return
	}

	// Check cache first (only for non-streaming requests)
	var cacheKey string

	if !openAIReq.Stream {
		cacheKey = h.cacheService.GenerateCacheKey(&openAIReq)
		if cachedResp, found := h.cacheService.Get(ctx, cacheKey); found {
			h.logger.Infof("Cache hit for request: %s", requestID)

			// Update request ID
			cachedResp.ID = requestID

			// Log cached request
			duration := time.Since(startTime)
			h.logRequest(
				ctx,
				clientID,
				requestID,
				openAIReq.Model,
				http.StatusOK,
				cachedResp.Usage.PromptTokens,
				cachedResp.Usage.CompletionTokens,
				duration.Milliseconds(),
				startTime,
				nil,
			)

			// Send cached response
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(cachedResp)
			return
		}
	}

	// Convert OpenAI request to Claude format
	claudeReq, err := claude.MapOpenAIToClaude(&openAIReq)
	if err != nil {
		h.logRequest(ctx, clientID, requestID, openAIReq.Model, http.StatusBadRequest, 0, 0, 0, startTime, err)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "Failed to convert request: "+err.Error())
		return
	}

	// Handle streaming requests differently (no caching for streams)
	if openAIReq.Stream {
		h.handleStreamingCompletion(w, r, clientID, requestID, openAIReq.Model, claudeReq, startTime)
		return
	}

	h.logger.Debugf("Sending request to Claude API: model=%s, max_tokens=%d", claudeReq.Model, claudeReq.MaxTokens)

	// Send request to Claude via provider manager (with retry logic)
	claudeResp, err := h.providerManager.CreateMessage(ctx, claudeReq)
	if err != nil {
		// Check if it's a Claude API error
		statusCode := http.StatusInternalServerError
		errorType := "api_error"

		if apiErr, ok := err.(*claude.APIError); ok {
			statusCode = h.mapClaudeStatusCode(apiErr.StatusCode)
			errorType = apiErr.Type
		}

		h.logRequest(ctx, clientID, requestID, openAIReq.Model, statusCode, 0, 0, 0, startTime, err)
		h.respondError(w, statusCode, errorType, "Provider error: "+err.Error())
		return
	}

	// Convert Claude response to OpenAI format
	openAIResp := claude.MapClaudeToOpenAI(claudeResp, openAIReq.Model)
	openAIResp.ID = requestID

	// Cache the response (fire and forget)
	if cacheKey != "" {
		go func() {
			if err := h.cacheService.Set(context.Background(), cacheKey, openAIResp); err != nil {
				h.logger.Warnf("Failed to cache response: %v", err)
			}
		}()
	}

	// Calculate cost
	cost := claude.CalculateCost(openAIReq.Model, claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)

	// Log successful request
	duration := time.Since(startTime)
	h.logRequest(
		ctx,
		clientID,
		requestID,
		openAIReq.Model,
		http.StatusOK,
		claudeResp.Usage.InputTokens,
		claudeResp.Usage.OutputTokens,
		duration.Milliseconds(),
		startTime,
		nil,
	)

	h.logger.Infof("Chat completion successful: tokens=%d, cost=$%.6f, duration=%v",
		claudeResp.Usage.InputTokens+claudeResp.Usage.OutputTokens, cost, duration)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResp)
}

// logRequest logs the request to database (async, fire-and-forget)
func (h *ChatHandler) logRequest(
	ctx context.Context,
	clientID string,
	requestID string,
	model string,
	statusCode int,
	promptTokens int,
	completionTokens int,
	durationMS int64,
	startTime time.Time,
	err error,
) {
	// Create log entry
	log := &repositories.RequestLog{
		ID:               uuid.New(),
		RequestID:        requestID,
		Method:           "POST",
		Path:             "/v1/chat/completions",
		Model:            model,
		Provider:         "claude",
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		CostUSD:          claude.CalculateCost(model, promptTokens, completionTokens),
		DurationMS:       int(durationMS),
		StatusCode:       statusCode,
		Cached:           false,
		IPAddress:        "", // TODO: Extract from request
		UserAgent:        "", // TODO: Extract from request
	}

	// Get client UUID from client_id string
	if clientID != "" {
		client, err := h.clientRepo.GetByClientID(ctx, clientID)
		if err == nil {
			log.ClientID = &client.ID
		}
	}

	// Add error message if present
	if err != nil {
		errMsg := err.Error()
		log.ErrorMessage = &errMsg
	}

	// Save to database (fire-and-forget, don't block response)
	go func() {
		if err := h.requestLogRepo.Create(context.Background(), log); err != nil {
			h.logger.Error(err, "Failed to log request")
		}
	}()
}

// mapClaudeStatusCode maps Claude API status codes to HTTP status codes
func (h *ChatHandler) mapClaudeStatusCode(claudeStatus int) int {
	// Claude uses standard HTTP status codes, so just pass through
	return claudeStatus
}

// handleStreamingCompletion handles streaming chat completion requests
func (h *ChatHandler) handleStreamingCompletion(
	w http.ResponseWriter,
	r *http.Request,
	clientID string,
	requestID string,
	model string,
	claudeReq *models.ClaudeRequest,
	startTime time.Time,
) {
	ctx := r.Context()

	h.logger.Debugf("Starting streaming request to Claude API: model=%s", claudeReq.Model)

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	// Get Claude client for streaming
	claudeClient := h.providerManager.GetClaudeClient()
	if claudeClient == nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "No Claude provider available")
		return
	}

	// Start streaming
	eventChan, err := claudeClient.CreateMessageStream(ctx, claudeReq)
	if err != nil {
		h.logger.Errorf(err, "Failed to start stream")
		h.writeSSEError(w, flusher, "Failed to start stream: "+err.Error())
		return
	}

	// Create stream mapper
	mapper := claude.NewStreamMapper(requestID, model)

	// Process streaming events
	for event := range eventChan {
		// Check for errors
		if event.Error != nil {
			h.logger.Errorf(event.Error, "Stream error")
			h.writeSSEError(w, flusher, "Stream error: "+event.Error.Error())
			return
		}

		// Map Claude event to OpenAI format
		openAIChunk, err := mapper.MapClaudeStreamToOpenAI(event.ClaudeEvent)
		if err != nil {
			// Check if it's the [DONE] marker
			if err.Error() == "[DONE]" {
				h.writeSSEData(w, flusher, "[DONE]")
				break
			}
			h.logger.Warnf("Failed to map stream event: %v", err)
			continue
		}

		// Skip nil chunks (some events don't produce output)
		if openAIChunk == nil {
			continue
		}

		// Send chunk to client
		chunkJSON, err := json.Marshal(openAIChunk)
		if err != nil {
			h.logger.Errorf(err, "Failed to marshal chunk")
			continue
		}

		h.writeSSEData(w, flusher, string(chunkJSON))
	}

	// Log request (with usage from mapper)
	usage := mapper.GetUsage()
	duration := time.Since(startTime)
	h.logRequest(
		ctx,
		clientID,
		requestID,
		model,
		http.StatusOK,
		usage.PromptTokens,
		usage.CompletionTokens,
		duration.Milliseconds(),
		startTime,
		nil,
	)

	h.logger.Infof("Streaming completion successful: tokens=%d, duration=%v",
		usage.TotalTokens, duration)
}

// writeSSEData writes SSE data to the response
func (h *ChatHandler) writeSSEData(w http.ResponseWriter, flusher http.Flusher, data string) {
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// writeSSEError writes an SSE error event
func (h *ChatHandler) writeSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	errorResponse := models.OpenAIErrorResponse{
		Error: models.OpenAIError{
			Message: message,
			Type:    "stream_error",
		},
	}
	errorJSON, _ := json.Marshal(errorResponse)
	fmt.Fprintf(w, "data: %s\n\n", string(errorJSON))
	flusher.Flush()
}

// respondError sends an OpenAI-compatible error response
func (h *ChatHandler) respondError(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.OpenAIErrorResponse{
		Error: models.OpenAIError{
			Message: message,
			Type:    errorType,
		},
	})
}
