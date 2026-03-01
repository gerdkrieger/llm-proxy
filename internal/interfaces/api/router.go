// Package api provides HTTP routing and middleware.
package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/cache"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
	customMiddleware "github.com/llm-proxy/llm-proxy/internal/interfaces/middleware"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// Router holds the HTTP router and dependencies
type Router struct {
	chi.Router
	config *config.Config
	db     *database.DB
	redis  *cache.RedisClient
	logger *logger.Logger
}

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(
	cfg *config.Config,
	db *database.DB,
	redis *cache.RedisClient,
	log *logger.Logger,
	oauthHandler *OAuthHandler,
	chatHandler *ChatHandler,
	modelsHandler *ModelsHandler,
	adminHandler *AdminHandler,
	filterHandler *ContentFilterHandler,
	providerMgmtHandler *ProviderManagementHandler,
	apiKeyMiddleware *customMiddleware.APIKeyMiddleware,
	apiKeyAuthMiddleware *customMiddleware.APIKeyAuthMiddleware,
	oauthMiddleware *customMiddleware.OAuthMiddleware,
	adminMiddleware *customMiddleware.AdminMiddleware,
	requestLoggerMiddleware *customMiddleware.RequestLoggerMiddleware,
	metricsMiddleware func(http.Handler) http.Handler,
	metricsHandler http.Handler,
) *Router {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.Server.Timeout))

	// Custom logging middleware
	r.Use(LoggingMiddleware(log))

	// Request logger middleware (logs to database for Live Monitor)
	r.Use(requestLoggerMiddleware.Middleware)

	// Metrics middleware (must be before routes)
	r.Use(metricsMiddleware)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3005",
			"http://localhost:3000",
			"https://llmproxy.aitrail.ch",
			"https://llmproxy.aitrail.ch:3005",
			"https://chat.aitrail.ch",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Admin-API-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Initialize handlers
	healthHandler := NewHealthHandler(db, redis, log)

	// Metrics endpoint (public, no auth)
	r.Get("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metricsHandler.ServeHTTP(w, req)
	})

	// Health check routes (public, no auth)
	r.Get("/health", healthHandler.SimpleHealth)
	r.Get("/health/ready", healthHandler.DetailedHealth)
	r.Get("/health/detailed", healthHandler.DetailedHealthWithStats)

	// OAuth endpoints (public)
	r.Post("/oauth/token", oauthHandler.Token)
	r.Post("/oauth/revoke", oauthHandler.Revoke)

	// OpenAI-compatible API endpoints (Static API Key OR DB API Key OR OAuth protected)
	r.Route("/v1", func(r chi.Router) {
		// Try static API key first (pass through if not static key)
		r.Use(apiKeyMiddleware.Authenticate)
		// Then try DB-based API key (client secrets from oauth_clients table)
		r.Use(apiKeyAuthMiddleware.Authenticate)
		// Finally OAuth (handles OAuth tokens)
		r.Use(oauthMiddleware.Authenticate)

		// Chat completions endpoint (requires 'write' scope)
		r.Group(func(r chi.Router) {
			r.Use(oauthMiddleware.RequireScope("write"))
			r.Post("/chat/completions", chatHandler.CreateCompletion)
		})

		// Models endpoints (requires 'read' scope)
		r.Group(func(r chi.Router) {
			r.Use(oauthMiddleware.RequireScope("read"))
			r.Get("/models", modelsHandler.ListModels)
			r.Get("/models/{model}", modelsHandler.GetModel)
		})
	})

	// Admin API endpoints (Admin API Key protected)
	r.Route("/admin", func(r chi.Router) {
		r.Use(adminMiddleware.Authenticate)

		// Client Management
		r.Get("/clients", adminHandler.ListClients)
		r.Post("/clients", adminHandler.CreateClient)
		r.Get("/clients/{client_id}", adminHandler.GetClient)
		r.Patch("/clients/{client_id}", adminHandler.UpdateClient)
		r.Delete("/clients/{client_id}", adminHandler.DeleteClient)
		r.Post("/clients/{client_id}/reset-secret", adminHandler.ResetClientSecret)

		// Cache Management
		r.Get("/cache/stats", adminHandler.GetCacheStats)
		r.Post("/cache/clear", adminHandler.ClearCache)
		r.Post("/cache/invalidate/{model}", adminHandler.InvalidateCacheByModel)

		// Dashboard
		r.Get("/dashboard", adminHandler.GetDashboardData)

		// Usage Statistics
		r.Get("/stats/usage", adminHandler.GetUsageStats)

		// Request Logs (for Live Monitor)
		r.Get("/requests", adminHandler.GetRequestLogs)
		r.Get("/requests/{id}", adminHandler.GetRequestLogDetails)

		// Provider Management
		r.Post("/providers/sync-models", providerMgmtHandler.SyncProviderModels) // POST /admin/providers/sync-models (must be before /{id})
		r.Get("/providers/status", adminHandler.GetProviderStatus)               // GET /admin/providers/status
		r.Get("/providers", adminHandler.GetProviderDetails)                     // GET /admin/providers
		r.Post("/providers", providerMgmtHandler.CreateProvider)                 // POST /admin/providers

		r.Route("/providers/{id}", func(r chi.Router) {
			r.Put("/", providerMgmtHandler.UpdateProvider)                           // PUT /admin/providers/{id}
			r.Delete("/", providerMgmtHandler.DeleteProvider)                        // DELETE /admin/providers/{id}
			r.Get("/config", providerMgmtHandler.GetProviderConfig)                  // GET /admin/providers/{id}/config
			r.Post("/test", providerMgmtHandler.TestProvider)                        // POST /admin/providers/{id}/test
			r.Put("/toggle", providerMgmtHandler.ToggleProvider)                     // PUT /admin/providers/{id}/toggle
			r.Get("/models", providerMgmtHandler.GetProviderModels)                  // GET /admin/providers/{id}/models
			r.Post("/models/import", providerMgmtHandler.ImportProviderModels)       // POST /admin/providers/{id}/models/import
			r.Post("/models/configure", providerMgmtHandler.ConfigureProviderModels) // POST /admin/providers/{id}/models/configure

			// Provider API Key Management
			r.Get("/keys", providerMgmtHandler.ListProviderAPIKeys)                  // GET /admin/providers/{id}/keys
			r.Post("/keys", providerMgmtHandler.AddProviderAPIKey)                   // POST /admin/providers/{id}/keys
			r.Delete("/keys/{key_id}", providerMgmtHandler.DeleteProviderAPIKey)     // DELETE /admin/providers/{id}/keys/{key_id}
			r.Put("/keys/{key_id}/toggle", providerMgmtHandler.ToggleProviderAPIKey) // PUT /admin/providers/{id}/keys/{key_id}/toggle
		})

		// Content Filter Management
		r.Route("/filters", func(r chi.Router) {
			r.Get("/", filterHandler.ListFilters)                   // GET /admin/filters
			r.Post("/", filterHandler.CreateFilter)                 // POST /admin/filters
			r.Get("/stats", filterHandler.GetFilterStats)           // GET /admin/filters/stats
			r.Get("/matches", adminHandler.GetFilterMatches)        // GET /admin/filters/matches
			r.Post("/test", filterHandler.TestFilter)               // POST /admin/filters/test
			r.Post("/refresh", filterHandler.RefreshFilters)        // POST /admin/filters/refresh
			r.Post("/bulk-import", filterHandler.BulkImportFilters) // POST /admin/filters/bulk-import

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", filterHandler.GetFilter)       // GET /admin/filters/{id}
				r.Put("/", filterHandler.UpdateFilter)    // PUT /admin/filters/{id}
				r.Delete("/", filterHandler.DeleteFilter) // DELETE /admin/filters/{id}
				r.Post("/test", filterHandler.TestFilter) // POST /admin/filters/{id}/test
			})
		})

		// System Settings Management
		r.Get("/settings", adminHandler.GetSettings)   // GET /admin/settings
		r.Put("/settings", adminHandler.UpdateSetting) // PUT /admin/settings
	})

	return &Router{
		Router: r,
		config: cfg,
		db:     db,
		redis:  redis,
		logger: log,
	}
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request
			duration := time.Since(start)

			requestID := middleware.GetReqID(r.Context())

			log.GetZerolog().Info().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration_ms", duration).
				Str("user_agent", r.UserAgent()).
				Msg("HTTP request processed")
		})
	}
}
