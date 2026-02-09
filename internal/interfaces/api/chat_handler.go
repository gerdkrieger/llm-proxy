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
	"github.com/llm-proxy/llm-proxy/internal/application/attachment"
	"github.com/llm-proxy/llm-proxy/internal/application/caching"
	"github.com/llm-proxy/llm-proxy/internal/application/filtering"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/claude"
	openaiProvider "github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/openai"
	mw "github.com/llm-proxy/llm-proxy/internal/interfaces/middleware"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
	"github.com/llm-proxy/llm-proxy/pkg/metrics"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
	providerManager   *providers.ProviderManager
	requestLogRepo    *repositories.RequestLogRepository
	filterMatchRepo   *repositories.FilterMatchRepository
	clientRepo        *repositories.OAuthClientRepository
	cacheService      *caching.Service
	filterService     *filtering.Service
	attachmentService *attachment.Service
	metrics           *metrics.Metrics
	logger            *logger.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(
	providerManager *providers.ProviderManager,
	requestLogRepo *repositories.RequestLogRepository,
	filterMatchRepo *repositories.FilterMatchRepository,
	clientRepo *repositories.OAuthClientRepository,
	cacheService *caching.Service,
	filterService *filtering.Service,
	attachmentService *attachment.Service,
	metricsCollector *metrics.Metrics,
	log *logger.Logger,
) *ChatHandler {
	return &ChatHandler{
		providerManager:   providerManager,
		requestLogRepo:    requestLogRepo,
		filterMatchRepo:   filterMatchRepo,
		clientRepo:        clientRepo,
		cacheService:      cacheService,
		filterService:     filterService,
		attachmentService: attachmentService,
		metrics:           metricsCollector,
		logger:            log,
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
		h.logRequest(ctx, r, clientID, requestID, "", http.StatusBadRequest, 0, 0, 0, startTime, err)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "Invalid request body: "+err.Error())
		return
	}

	// Validate request
	if openAIReq.Model == "" {
		h.logRequest(ctx, r, clientID, requestID, "", http.StatusBadRequest, 0, 0, 0, startTime, nil)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}

	if len(openAIReq.Messages) == 0 {
		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, http.StatusBadRequest, 0, 0, 0, startTime, nil)
		h.respondError(w, http.StatusBadRequest, "invalid_request_error", "messages are required")
		return
	}

	// Check if client has access to the requested model
	if clientID != "" {
		hasAccess, err := h.checkModelAccess(ctx, clientID, openAIReq.Model)
		if err != nil {
			h.logger.Errorf(err, "Failed to check model access for client %s", clientID)
			h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, http.StatusInternalServerError, 0, 0, 0, startTime, err)
			h.respondError(w, http.StatusInternalServerError, "server_error", "Failed to verify model access")
			return
		}
		if !hasAccess {
			h.logger.Warnf("Client %s attempted to access unauthorized model: %s", clientID, openAIReq.Model)
			h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, http.StatusForbidden, 0, 0, 0, startTime, nil)
			h.respondError(w, http.StatusForbidden, "model_not_allowed", fmt.Sprintf("no access to model '%s'", openAIReq.Model))
			return
		}
	}

	// Analyze attachments for sensitive content
	if h.attachmentService != nil {
		attachmentResult, err := h.attachmentService.AnalyzeAttachments(ctx, openAIReq.Messages)
		if err != nil {
			h.logger.Warnf("Failed to analyze attachments: %v", err)
		} else if attachmentResult != nil {
			if attachmentResult.HasAttachments {
				h.logger.Infof("Request %s contains %d attachments (%d blocked)",
					requestID, attachmentResult.TotalAttachments, attachmentResult.BlockedAttachments)
			}

			// If attachments were blocked or redacted, log them
			if len(attachmentResult.FilterMatches) > 0 {
				if attachmentResult.BlockedAttachments > 0 {
					h.logger.Warnf("Blocked %d attachments due to sensitive content in request %s",
						attachmentResult.BlockedAttachments, requestID)
				} else {
					h.logger.Infof("Redacted PII in %d attachments for request %s",
						attachmentResult.TotalAttachments, requestID)
				}
				go h.logFilterMatches(r, clientID, requestID, openAIReq.Model, attachmentResult.FilterMatches)
			}

			// Use cleaned messages (with blocked attachments removed)
			openAIReq.Messages = attachmentResult.CleanedMessages
		}
	}

	// Apply content filters to user messages
	if h.filterService != nil {
		filteredMessages, matches, err := h.filterService.ApplyFilters(ctx, openAIReq.Messages)
		if err != nil {
			h.logger.Warnf("Failed to apply content filters: %v", err)
			// Continue without filtering on error
		} else {
			openAIReq.Messages = filteredMessages
			if len(matches) > 0 {
				h.logger.Infof("Applied %d content filters to request %s (client: %s)", len(matches), requestID, clientID)
				// Log filter matches to database (async, fire-and-forget)
				go h.logFilterMatches(r, clientID, requestID, openAIReq.Model, matches)
			}
		}
	}

	// Determine provider based on model name
	provider := h.providerManager.DetermineProvider(openAIReq.Model)
	h.logger.Infof("Routing request to provider: %s (model: %s)", provider, openAIReq.Model)

	// Check cache first (only for non-streaming requests)
	var cacheKey string

	if !openAIReq.Stream {
		cacheKey = h.cacheService.GenerateCacheKey(&openAIReq)
		if cachedResp, found := h.cacheService.Get(ctx, cacheKey); found {
			h.logger.Infof("Cache hit for request: %s", requestID)
			if h.metrics != nil {
				h.metrics.RecordCacheHit(time.Since(startTime))
			}

			// Update request ID
			cachedResp.ID = requestID

			// Log cached request
			duration := time.Since(startTime)
			h.logRequest(
				ctx,
				r,
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
		} else if h.metrics != nil {
			h.metrics.RecordCacheMiss(time.Since(startTime))
		}
	}

	// Route to appropriate provider
	if provider == "openai" {
		h.handleOpenAICompletion(w, r, ctx, clientID, requestID, &openAIReq, cacheKey, startTime)
	} else {
		h.handleClaudeCompletion(w, r, ctx, clientID, requestID, &openAIReq, cacheKey, startTime)
	}
}

// logRequest logs the request to database (async, fire-and-forget)
func (h *ChatHandler) logRequest(
	ctx context.Context,
	r *http.Request,
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
	// Extract IP address and user agent from request
	ipAddr := h.extractIPAddress(r)
	userAgent := r.UserAgent()
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// Determine provider based on model
	provider := h.providerManager.DetermineProvider(model)

	// Create log entry
	log := &repositories.RequestLog{
		ID:               uuid.New(),
		RequestID:        requestID,
		Method:           r.Method,
		Path:             r.URL.Path,
		Model:            model,
		Provider:         provider,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		CostUSD:          claude.CalculateCost(model, promptTokens, completionTokens),
		DurationMS:       int(durationMS),
		StatusCode:       statusCode,
		Cached:           false,
		IPAddress:        ipAddr,
		UserAgent:        userAgentPtr,
	}

	// Get client UUID and auth info from client_id string
	if clientID != "" {
		client, err := h.clientRepo.GetByClientID(ctx, clientID)
		if err == nil {
			log.ClientID = &client.ID
			authType := "api_key"
			log.AuthType = &authType
			log.APIKeyName = &client.Name
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

// extractIPAddress extracts the client IP address from the request
// Checks X-Forwarded-For, X-Real-IP headers first, then falls back to RemoteAddr
func (h *ChatHandler) extractIPAddress(r *http.Request) *string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' || xff[i] == ' ' {
				ip := xff[:i]
				return &ip
			}
		}
		return &xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return &xri
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// RemoteAddr is in format "IP:port", extract just the IP
		addr := r.RemoteAddr
		for i := len(addr) - 1; i >= 0; i-- {
			if addr[i] == ':' {
				ip := addr[:i]
				return &ip
			}
		}
		return &addr
	}

	// No IP available
	return nil
}

// logFilterMatches logs filter matches to database (async, fire-and-forget)
func (h *ChatHandler) logFilterMatches(
	r *http.Request,
	clientID string,
	requestID string,
	model string,
	matches []filtering.FilterMatch,
) {
	if h.filterMatchRepo == nil || len(matches) == 0 {
		return
	}

	ctx := context.Background()
	ipAddr := h.extractIPAddress(r)
	userAgent := r.UserAgent()
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// Get client UUID from client_id string
	var clientUUID *uuid.UUID
	if clientID != "" {
		if client, err := h.clientRepo.GetByClientID(ctx, clientID); err == nil {
			clientUUID = &client.ID
		}
	}

	// Determine provider (simple heuristic based on model name)
	provider := "claude"
	if len(model) > 0 {
		if model[0] == 'g' && (len(model) < 3 || model[:3] == "gpt") {
			provider = "openai"
		}
	}

	// Log each match
	for _, match := range matches {
		// For attachment redactions (FilterID=0), use NULL filter_id
		var filterIDPtr *int
		if match.FilterID > 0 {
			filterIDPtr = &match.FilterID
		}

		filterMatch := &repositories.FilterMatch{
			ID:          uuid.New(),
			RequestID:   requestID,
			ClientID:    clientUUID,
			FilterID:    filterIDPtr,
			Model:       model,
			Provider:    provider,
			Pattern:     match.Pattern,
			Replacement: match.Replacement,
			FilterType:  "", // Will be set if available
			MatchCount:  match.MatchCount,
			MatchedText: nil, // Don't store actual matched text for privacy
			IPAddress:   ipAddr,
			UserAgent:   userAgentPtr,
		}

		if err := h.filterMatchRepo.Create(ctx, filterMatch); err != nil {
			h.logger.Error(err, "Failed to log filter match")
		}
	}
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

	// Get Claude client for streaming
	claudeClient := h.providerManager.GetClaudeClient()
	if claudeClient == nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "No Claude provider available")
		return
	}

	// Start streaming BEFORE writing SSE headers.
	// This way, if the upstream returns an error, we can send a proper HTTP error
	// response instead of a broken SSE stream that confuses clients like OpenWebUI.
	eventChan, err := claudeClient.CreateMessageStream(ctx, claudeReq)
	if err != nil {
		h.logger.Errorf(err, "Failed to start stream")
		h.logRequest(ctx, r, clientID, requestID, model,
			http.StatusBadGateway, 0, 0, time.Since(startTime).Milliseconds(), startTime, err)
		h.respondError(w, http.StatusBadGateway, "upstream_error", "Failed to start stream: "+err.Error())
		return
	}

	// Upstream connection succeeded - now commit to SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx
	w.WriteHeader(http.StatusOK)

	flusher := &safeFlusher{w: w}
	h.logger.Debugf("Streaming with flusher support: %v", flusher.canFlush())

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
		r,
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

	// Record LLM metrics for streaming
	if h.metrics != nil {
		cost := claude.CalculateCost(model, usage.PromptTokens, usage.CompletionTokens)
		h.metrics.RecordLLMRequest(model, "claude", "success", duration,
			usage.PromptTokens, usage.CompletionTokens, cost, clientID)
	}

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

// checkModelAccess verifies if a client has access to a specific model
// Returns true if:
// - client.AllowedModels is nil (all models allowed)
// - model is in client.AllowedModels array
// Returns false otherwise
func (h *ChatHandler) checkModelAccess(ctx context.Context, clientID string, model string) (bool, error) {
	// Get client from database
	client, err := h.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return false, fmt.Errorf("failed to get client: %w", err)
	}

	// If AllowedModels is nil, all models are allowed
	if client.AllowedModels == nil {
		return true, nil
	}

	// If AllowedModels is empty array, no models are allowed
	if len(client.AllowedModels) == 0 {
		return false, nil
	}

	// Check if requested model is in the allowed list
	for _, allowedModel := range client.AllowedModels {
		if allowedModel == model {
			return true, nil
		}
	}

	// Model not found in allowed list
	return false, nil
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

// handleClaudeCompletion handles completion requests routed to Claude
func (h *ChatHandler) handleClaudeCompletion(
	w http.ResponseWriter,
	r *http.Request,
	ctx context.Context,
	clientID string,
	requestID string,
	openAIReq *models.OpenAIRequest,
	cacheKey string,
	startTime time.Time,
) {
	// Convert OpenAI request to Claude format
	claudeReq, err := claude.MapOpenAIToClaude(openAIReq)
	if err != nil {
		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, http.StatusBadRequest, 0, 0, 0, startTime, err)
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

		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, statusCode, 0, 0, 0, startTime, err)
		if h.metrics != nil {
			h.metrics.RecordLLMError(openAIReq.Model, "claude", errorType)
		}
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
		r,
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

	// Record LLM metrics
	if h.metrics != nil {
		h.metrics.RecordLLMRequest(openAIReq.Model, "claude", "success", duration,
			claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens, cost, clientID)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResp)
}

// handleOpenAICompletion handles completion requests routed to OpenAI
func (h *ChatHandler) handleOpenAICompletion(
	w http.ResponseWriter,
	r *http.Request,
	ctx context.Context,
	clientID string,
	requestID string,
	openAIReq *models.OpenAIRequest,
	cacheKey string,
	startTime time.Time,
) {
	// Get OpenAI client
	openaiClient := h.providerManager.GetOpenAIClient()
	if openaiClient == nil {
		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, http.StatusServiceUnavailable, 0, 0, 0, startTime, nil)
		h.respondError(w, http.StatusServiceUnavailable, "provider_unavailable", "OpenAI provider not available")
		return
	}

	// Handle streaming requests
	if openAIReq.Stream {
		h.handleOpenAIStreamingCompletion(w, r, clientID, requestID, openAIReq, openaiClient, startTime)
		return
	}

	h.logger.Debugf("Sending request to OpenAI API: model=%s", openAIReq.Model)

	// Send request to OpenAI (native format, no conversion needed)
	openAIResp, err := openaiClient.CreateMessage(ctx, openAIReq)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "api_error"

		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model, statusCode, 0, 0, 0, startTime, err)
		if h.metrics != nil {
			h.metrics.RecordLLMError(openAIReq.Model, "openai", errorType)
		}
		h.respondError(w, statusCode, errorType, "Provider error: "+err.Error())
		return
	}

	// Update request ID
	openAIResp.ID = requestID

	// Cache the response (fire and forget)
	if cacheKey != "" {
		go func() {
			if err := h.cacheService.Set(context.Background(), cacheKey, openAIResp); err != nil {
				h.logger.Warnf("Failed to cache response: %v", err)
			}
		}()
	}

	// Log successful request
	duration := time.Since(startTime)
	h.logRequest(
		ctx,
		r,
		clientID,
		requestID,
		openAIReq.Model,
		http.StatusOK,
		openAIResp.Usage.PromptTokens,
		openAIResp.Usage.CompletionTokens,
		duration.Milliseconds(),
		startTime,
		nil,
	)

	h.logger.Infof("Chat completion successful: tokens=%d, duration=%v",
		openAIResp.Usage.TotalTokens, duration)

	// Record LLM metrics
	if h.metrics != nil {
		h.metrics.RecordLLMRequest(openAIReq.Model, "openai", "success", duration,
			openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens, 0, clientID)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResp)
}

// handleOpenAIStreamingCompletion handles streaming chat completion requests for OpenAI models.
// Since OpenAI already returns SSE in OpenAI format, we pipe the chunks through directly.
func (h *ChatHandler) handleOpenAIStreamingCompletion(
	w http.ResponseWriter,
	r *http.Request,
	clientID string,
	requestID string,
	openAIReq *models.OpenAIRequest,
	openaiClient *openaiProvider.Client,
	startTime time.Time,
) {
	ctx := r.Context()

	h.logger.Debugf("Starting streaming request to OpenAI API: model=%s", openAIReq.Model)

	// Start streaming from OpenAI BEFORE writing SSE headers.
	// This way, if the upstream returns an error (e.g. 400 bad request),
	// we can still send a proper HTTP error response instead of a broken SSE stream
	// that confuses clients like OpenWebUI.
	eventChan, err := openaiClient.CreateMessageStream(ctx, openAIReq)
	if err != nil {
		h.logger.Errorf(err, "Failed to start OpenAI stream")
		h.logRequest(ctx, r, clientID, requestID, openAIReq.Model,
			http.StatusBadGateway, 0, 0, time.Since(startTime).Milliseconds(), startTime, err)
		h.respondError(w, http.StatusBadGateway, "upstream_error", "Failed to start stream: "+err.Error())
		return
	}

	// Upstream connection succeeded - now commit to SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	w.WriteHeader(http.StatusOK)
	flusher := &safeFlusher{w: w}

	// Track usage from the stream
	var promptTokens, completionTokens int

	// Process streaming events - pipe through directly since format is already OpenAI
	for event := range eventChan {
		if event.Error != nil {
			h.logger.Errorf(event.Error, "OpenAI stream error")
			h.writeSSEError(w, flusher, "Stream error: "+event.Error.Error())
			return
		}

		if event.Data == "[DONE]" {
			h.writeSSEData(w, flusher, "[DONE]")
			break
		}

		// Try to extract usage info from the chunk (OpenAI includes it in the last chunk)
		var chunk models.OpenAIStreamResponse
		if err := json.Unmarshal([]byte(event.Data), &chunk); err == nil {
			if chunk.Usage != nil {
				promptTokens = chunk.Usage.PromptTokens
				completionTokens = chunk.Usage.CompletionTokens
			}
		}

		// Forward the chunk as-is (already in OpenAI format)
		h.writeSSEData(w, flusher, event.Data)
	}

	// Log request
	duration := time.Since(startTime)
	h.logRequest(ctx, r, clientID, requestID, openAIReq.Model,
		http.StatusOK, promptTokens, completionTokens, duration.Milliseconds(), startTime, nil)

	// Record LLM metrics for streaming
	if h.metrics != nil {
		h.metrics.RecordLLMRequest(openAIReq.Model, "openai", "success", duration,
			promptTokens, completionTokens, 0, clientID)
	}

	h.logger.Infof("OpenAI streaming completion successful: prompt=%d, completion=%d, duration=%v",
		promptTokens, completionTokens, duration)
}

// safeFlusher wraps a ResponseWriter and provides safe flushing
type safeFlusher struct {
	w http.ResponseWriter
}

// Flush flushes the response writer if it supports flushing
func (sf *safeFlusher) Flush() {
	if f, ok := sf.w.(http.Flusher); ok {
		f.Flush()
	}
}

// canFlush checks if the underlying writer supports flushing
func (sf *safeFlusher) canFlush() bool {
	_, ok := sf.w.(http.Flusher)
	return ok
}
