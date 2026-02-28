// Package middleware provides HTTP middleware for the LLM-Proxy.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// RequestLoggerMiddleware logs all API requests to the database
type RequestLoggerMiddleware struct {
	repo         *repositories.RequestLogRepository
	settingsRepo *repositories.SystemSettingsRepository
	logger       *logger.Logger

	// Cached setting for body capture (refreshed periodically)
	captureBodies    bool
	captureBodyMutex sync.RWMutex
	lastSettingCheck time.Time
}

// NewRequestLoggerMiddleware creates a new request logger middleware
func NewRequestLoggerMiddleware(repo *repositories.RequestLogRepository, settingsRepo *repositories.SystemSettingsRepository, log *logger.Logger) *RequestLoggerMiddleware {
	m := &RequestLoggerMiddleware{
		repo:          repo,
		settingsRepo:  settingsRepo,
		logger:        log,
		captureBodies: true, // Default to true
	}

	// Initial load of setting
	go m.refreshCaptureBodySetting()

	// Refresh setting every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			m.refreshCaptureBodySetting()
		}
	}()

	return m
}

// refreshCaptureBodySetting refreshes the cached body capture setting from database
func (m *RequestLoggerMiddleware) refreshCaptureBodySetting() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	capture := m.settingsRepo.GetBool(ctx, "capture_request_response_bodies")

	m.captureBodyMutex.Lock()
	m.captureBodies = capture
	m.lastSettingCheck = time.Now()
	m.captureBodyMutex.Unlock()
}

// shouldCaptureBodies returns whether bodies should be captured
func (m *RequestLoggerMiddleware) shouldCaptureBodies() bool {
	m.captureBodyMutex.RLock()
	defer m.captureBodyMutex.RUnlock()
	return m.captureBodies
}

// loggingResponseWriter is a wrapper that captures status code, response size, and body.
// It also implements http.Flusher so SSE streaming works correctly through this middleware.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	written      bool
	body         *bytes.Buffer // Capture response body
	headers      http.Header   // Capture response headers
	captureBody  bool          // Whether to capture body
}

// WriteHeader captures the status code and headers
func (rw *loggingResponseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		// Capture response headers before writing
		rw.headers = make(http.Header)
		for k, v := range rw.ResponseWriter.Header() {
			rw.headers[k] = v
		}
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write captures bytes written and response body
func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
		// Capture headers on first write
		rw.headers = make(http.Header)
		for k, v := range rw.ResponseWriter.Header() {
			rw.headers[k] = v
		}
	}

	// Capture response body (limit to 100KB to avoid memory issues) - only if enabled
	if rw.captureBody && rw.body.Len() < 100*1024 {
		rw.body.Write(b)
	}

	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Flush implements http.Flusher so SSE streaming works through this middleware
func (rw *loggingResponseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the underlying ResponseWriter (for http.ResponseController support)
func (rw *loggingResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// authInfoKey is the context key for the mutable AuthInfo struct
type authInfoKeyType struct{}

var authInfoKey = authInfoKeyType{}

// AuthInfo is a mutable struct stored in the context by the request logger middleware.
// Auth middlewares can update it so the logger can read auth info after the handler chain completes.
// This solves the problem where auth middlewares use r.WithContext() creating a new request
// that the outer logging middleware never sees.
type AuthInfo struct {
	AuthType   *string
	APIKeyName *string
	ClientID   *uuid.UUID
}

// SetAuthInfo stores auth info in the AuthInfo struct found in the request context.
// Auth middlewares should call this to make auth data available to the request logger.
func SetAuthInfo(ctx context.Context, authType string, apiKeyName string, clientID *uuid.UUID) {
	if info, ok := ctx.Value(authInfoKey).(*AuthInfo); ok {
		info.AuthType = &authType
		info.APIKeyName = &apiKeyName
		info.ClientID = clientID
	}
}

// Middleware logs the request to the database
func (m *RequestLoggerMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Check if we should capture bodies
		captureEnabled := m.shouldCaptureBodies()

		// Capture request body (limit to 100KB) - only if enabled
		// ContentLength can be -1 for chunked requests, so also handle that case
		var requestBody *string
		if captureEnabled && r.Body != nil && (r.ContentLength > 0 && r.ContentLength < 100*1024 || r.ContentLength == -1) {
			bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 100*1024))
			if err == nil && len(bodyBytes) > 0 {
				bodyStr := string(bodyBytes)
				requestBody = &bodyStr
				// Restore the body for the next handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Extract model from request body (for /v1/chat/completions requests)
		var requestModel string
		if requestBody != nil && strings.HasPrefix(r.URL.Path, "/v1/") {
			var bodyObj struct {
				Model string `json:"model"`
			}
			if err := json.Unmarshal([]byte(*requestBody), &bodyObj); err == nil && bodyObj.Model != "" {
				requestModel = bodyObj.Model
			}
		}

		// Capture request headers (sanitize auth headers)
		requestHeaders := sanitizeHeaders(r.Header)

		// Create response writer wrapper
		wrapped := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default if WriteHeader not called
			body:           &bytes.Buffer{},
			captureBody:    captureEnabled,
		}

		// Get or create request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Extract IP address
		ipAddress := extractIPAddress(r)

		// Store a mutable AuthInfo struct in the context. Auth middlewares downstream
		// can update this struct via SetAuthInfo(), and we read it back after the
		// handler chain completes. This avoids the r.WithContext() problem where auth
		// middlewares create a new request the outer middleware never sees.
		authInfo := &AuthInfo{}
		ctx := context.WithValue(r.Context(), authInfoKey, authInfo)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Read auth info that was set by auth middlewares via SetAuthInfo()
		var authType, apiKeyName *string
		var clientID *uuid.UUID
		if authInfo.AuthType != nil {
			authType = authInfo.AuthType
			apiKeyName = authInfo.APIKeyName
			clientID = authInfo.ClientID
		} else {
			// Fallback: try extracting from the original context
			authType, apiKeyName, clientID = extractAuthInfo(r.Context())
		}

		// Calculate duration
		duration := time.Since(start)

		// Extract error message if any
		var errorMsg *string
		if wrapped.statusCode >= 400 {
			// Try to get error from context
			if errCtx := r.Context().Value("error"); errCtx != nil {
				if err, ok := errCtx.(string); ok {
					errorMsg = &err
				}
			}
		}

		// Use model from request body if available, otherwise try context
		model := requestModel
		var provider string
		if model == "" {
			if modelCtx := r.Context().Value("model"); modelCtx != nil {
				if m, ok := modelCtx.(string); ok {
					model = m
				}
			}
		}
		if providerCtx := r.Context().Value("provider"); providerCtx != nil {
			if p, ok := providerCtx.(string); ok {
				provider = p
			}
		}
		// Determine provider from model name if not set from context
		if provider == "" && model != "" {
			if strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o4") {
				provider = "openai"
			} else if strings.HasPrefix(model, "claude") {
				provider = "claude"
			}
		}

		// Extract filtering info from context
		wasFiltered := false
		var filterReason *string
		if filteredCtx := r.Context().Value("filtered"); filteredCtx != nil {
			if filtered, ok := filteredCtx.(bool); ok {
				wasFiltered = filtered
			}
		}
		if reasonCtx := r.Context().Value("filter_reason"); reasonCtx != nil {
			if reason, ok := reasonCtx.(string); ok {
				filterReason = &reason
			}
		}

		// Capture response body and headers - only if enabled
		var responseBody *string
		if captureEnabled && wrapped.body.Len() > 0 {
			// Only capture if it's likely text (not binary/compressed)
			if isTextResponse(wrapped.headers) {
				bodyStr := wrapped.body.String()
				responseBody = &bodyStr
			}
		}

		// Sanitize response headers
		responseHeaders := sanitizeHeaders(wrapped.headers)

		// Calculate response size
		responseSize := int64(wrapped.body.Len())

		// Create log entry
		userAgent := r.UserAgent()
		logEntry := &repositories.RequestLog{
			ID:                uuid.New(),
			ClientID:          clientID,
			RequestID:         requestID,
			Method:            r.Method,
			Path:              r.URL.Path,
			Model:             model,
			Provider:          provider,
			DurationMS:        int(duration.Milliseconds()),
			StatusCode:        wrapped.statusCode,
			IPAddress:         &ipAddress,
			UserAgent:         &userAgent,
			ErrorMessage:      errorMsg,
			AuthType:          authType,
			APIKeyName:        apiKeyName,
			WasFiltered:       wasFiltered,
			FilterReason:      filterReason,
			RequestHeaders:    requestHeaders,
			RequestBody:       requestBody,
			ResponseHeaders:   responseHeaders,
			ResponseBody:      responseBody,
			ResponseSizeBytes: &responseSize,
		}

		// Log to database (fire and forget - don't block response)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := m.repo.Create(ctx, logEntry); err != nil {
				m.logger.Warnf("Failed to log request to database: %v", err)
			} else {
				m.logger.Debugf("Logged request: %s %s -> %d (%dms)", logEntry.Method, logEntry.Path, logEntry.StatusCode, logEntry.DurationMS)
			}
		}()
	})
}

// extractIPAddress extracts the real IP address from the request
func extractIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:Port", so we need to strip the port
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Handle IPv6 localhost
	if ip == "[::1]" || ip == "::1" {
		return "127.0.0.1"
	}

	// Remove IPv6 brackets if present
	ip = strings.Trim(ip, "[]")

	return ip
}

// extractAuthInfo extracts authentication type, key name, and client ID from context
// These values should be set by the authentication middlewares
func extractAuthInfo(ctx context.Context) (*string, *string, *uuid.UUID) {
	var authType, apiKeyName *string
	var clientID *uuid.UUID

	// Check for authenticated client (from new API key middleware)
	if client, ok := GetClientFromContext(ctx); ok {
		apiKeyName = &client.Name
		authTypeVal := "api_key"
		authType = &authTypeVal
		clientID = &client.ID
	}

	// Check for legacy API key authentication
	if keyName := ctx.Value("api_key_name"); keyName != nil {
		if name, ok := keyName.(string); ok {
			if apiKeyName == nil {
				apiKeyName = &name
				authTypeVal := "api_key"
				authType = &authTypeVal
			}
		}
	}

	// Check for OAuth authentication
	if oauthClientID := ctx.Value("oauth_client_id"); oauthClientID != nil {
		if _, ok := oauthClientID.(string); ok {
			authTypeVal := "oauth"
			authType = &authTypeVal
		}
	}

	// Check for admin authentication
	if adminKey := ctx.Value("admin_authenticated"); adminKey != nil {
		if _, ok := adminKey.(bool); ok {
			authTypeVal := "admin"
			authType = &authTypeVal
		}
	}

	// If no authentication found
	if authType == nil {
		noneVal := "none"
		authType = &noneVal
	}

	return authType, apiKeyName, clientID
}

// isTextResponse checks if the response is likely text (not binary/compressed)
func isTextResponse(headers http.Header) bool {
	// Don't capture if response is compressed
	if encoding := headers.Get("Content-Encoding"); encoding != "" {
		// gzip, deflate, br, etc. are binary
		return false
	}

	// Check Content-Type
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		// No content type, assume text for backwards compatibility
		return true
	}

	// List of text-based content types
	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-www-form-urlencoded",
		"application/ld+json",
		"application/graphql",
	}

	for _, textType := range textTypes {
		if strings.Contains(contentType, textType) {
			return true
		}
	}

	// Default to false for unknown types (images, videos, etc.)
	return false
}

// sanitizeHeaders removes or masks sensitive headers
func sanitizeHeaders(headers http.Header) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// List of sensitive headers to mask
	sensitiveHeaders := map[string]bool{
		"Authorization":   true,
		"X-Api-Key":       true,
		"X-Admin-Api-Key": true,
		"Cookie":          true,
		"Set-Cookie":      true,
		"X-Auth-Token":    true,
		"Api-Key":         true,
	}

	for key, values := range headers {
		// Check if this is a sensitive header
		if sensitiveHeaders[key] {
			// Mask it
			sanitized[key] = "***REDACTED***"
		} else if len(values) == 1 {
			// Single value
			sanitized[key] = values[0]
		} else {
			// Multiple values
			sanitized[key] = values
		}
	}

	return sanitized
}
