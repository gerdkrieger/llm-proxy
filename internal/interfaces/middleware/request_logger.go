// Package middleware provides HTTP middleware for the LLM-Proxy.
package middleware

import (
	"bytes"
	"context"
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

// loggingResponseWriter is a wrapper that captures status code, response size, and body
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

// Middleware logs the request to the database
func (m *RequestLoggerMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Check if we should capture bodies
		captureEnabled := m.shouldCaptureBodies()

		// Capture request body (limit to 100KB) - only if enabled
		var requestBody *string
		if captureEnabled && r.Body != nil && r.ContentLength > 0 && r.ContentLength < 100*1024 {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				bodyStr := string(bodyBytes)
				requestBody = &bodyStr
				// Restore the body for the next handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
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

		// Extract authentication info from context (set by auth middlewares)
		authType, apiKeyName, clientID := extractAuthInfo(r.Context())

		// Call next handler
		next.ServeHTTP(wrapped, r)

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

		// Extract model and provider info from context (if available)
		var model, provider string
		if modelCtx := r.Context().Value("model"); modelCtx != nil {
			if m, ok := modelCtx.(string); ok {
				model = m
			}
		}
		if providerCtx := r.Context().Value("provider"); providerCtx != nil {
			if p, ok := providerCtx.(string); ok {
				provider = p
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
			bodyStr := wrapped.body.String()
			responseBody = &bodyStr
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
