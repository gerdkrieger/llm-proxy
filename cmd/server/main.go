// Package main is the entry point for the LLM-Proxy server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/application/attachment"
	"github.com/llm-proxy/llm-proxy/internal/application/caching"
	"github.com/llm-proxy/llm-proxy/internal/application/filtering"
	"github.com/llm-proxy/llm-proxy/internal/application/oauth"
	modelsync "github.com/llm-proxy/llm-proxy/internal/application/providers"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/cache"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/internal/interfaces/api"
	"github.com/llm-proxy/llm-proxy/internal/interfaces/middleware"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
	"github.com/llm-proxy/llm-proxy/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Version information (set during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	log := logger.GetLogger()

	// Log startup information
	log.Infof("Starting LLM-Proxy Server")
	log.Infof("Version: %s", Version)
	log.Infof("Build Time: %s", BuildTime)
	log.Infof("Git Commit: %s", GitCommit)
	log.Infof("Go Version: %s", GoVersion)

	// Connect to database
	log.Info("Connecting to PostgreSQL...")
	db, err := database.New(cfg.Database, log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Info("Database connection established")

	// Connect to Redis
	log.Info("Connecting to Redis...")
	redis, err := cache.New(cfg.Redis, log)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Info("Redis connection established")

	// Initialize repositories
	log.Info("Initializing repositories...")
	clientRepo := repositories.NewOAuthClientRepository(db)
	tokenRepo := repositories.NewOAuthTokenRepository(db)
	requestLogRepo := repositories.NewRequestLogRepository(db)
	contentFilterRepo := repositories.NewContentFilterRepository(db)
	filterMatchRepo := repositories.NewFilterMatchRepository(db)
	providerSettingsRepo := repositories.NewProviderSettingsRepository(db)
	providerModelRepo := repositories.NewProviderModelRepository(db)

	// Initialize caching service
	log.Info("Initializing caching service...")
	cacheService := caching.NewService(redis, cfg.Cache, log)

	// Initialize content filtering service
	log.Info("Initializing content filtering service...")
	filterService := filtering.NewService(contentFilterRepo, log)

	// Initialize attachment service
	log.Info("Initializing attachment analysis service...")
	attachmentService := attachment.NewService(filterService, log)

	// Initialize metrics
	log.Info("Initializing Prometheus metrics...")
	metricsCollector := metrics.New("llm_proxy")

	// Initialize OAuth service
	log.Info("Initializing OAuth service...")
	oauthService, err := oauth.NewService(clientRepo, tokenRepo, cfg.OAuth, log)
	if err != nil {
		log.Fatalf("Failed to initialize OAuth service: %v", err)
	}

	// Initialize provider manager (Claude API)
	log.Info("Initializing provider manager...")
	providerManager := providers.NewProviderManager(cfg.Providers, log)

	// Check provider health
	ctx := context.Background()
	if err := providerManager.Health(ctx); err != nil {
		log.Warnf("Provider health check warning: %v", err)
	} else {
		log.Infof("Provider health check: %d provider(s) available", providerManager.GetProviderCount())
	}

	// Auto-sync models to database (so all models are available by default)
	log.Info("Synchronizing models to database...")
	modelSyncService := modelsync.NewModelSyncService(providerModelRepo, log)
	if err := modelSyncService.SyncModelsToDatabase(ctx); err != nil {
		log.Warnf("Model sync warning: %v", err)
	}

	// Initialize handlers
	log.Info("Initializing API handlers...")
	oauthHandler := api.NewOAuthHandler(oauthService, log)
	chatHandler := api.NewChatHandler(providerManager, requestLogRepo, filterMatchRepo, clientRepo, cacheService, filterService, attachmentService, log)
	modelsHandler := api.NewModelsHandler(providerManager, providerModelRepo, log)
	adminHandler := api.NewAdminHandler(clientRepo, tokenRepo, requestLogRepo, filterMatchRepo, providerModelRepo, cacheService, providerManager, log)
	filterHandler := api.NewContentFilterHandler(contentFilterRepo, filterService, log)
	providerMgmtHandler := api.NewProviderManagementHandler(providerSettingsRepo, providerModelRepo, providerManager, cfg, log)

	// Initialize middleware
	log.Info("Initializing middleware...")
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(cfg, log)
	oauthMiddleware := middleware.NewOAuthMiddleware(oauthService, log)
	adminMiddleware := middleware.NewAdminMiddleware(cfg, log)
	requestLoggerMiddleware := middleware.NewRequestLoggerMiddleware(requestLogRepo, log)
	metricsMiddleware := middleware.MetricsMiddleware(metricsCollector)

	// Create router with all handlers
	log.Info("Initializing router...")
	router := api.NewRouter(cfg, db, redis, log, oauthHandler, chatHandler, modelsHandler, adminHandler, filterHandler, providerMgmtHandler, apiKeyMiddleware, oauthMiddleware, adminMiddleware, requestLoggerMiddleware, metricsMiddleware, promhttp.Handler())

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start metrics updater goroutine (update DB stats every 30 seconds)
	stopMetrics := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Update database connection metrics
				stats := db.Stats()
				metricsCollector.UpdateDBStats(
					int(stats.TotalConns()),
					int(stats.IdleConns()),
				)
			case <-stopMetrics:
				return
			}
		}
	}()

	// Start server in goroutine
	go func() {
		log.Infof("Server listening on %s", server.Addr)
		log.Infof("Health endpoint: http://%s/health", server.Addr)
		log.Infof("Metrics endpoint: http://%s/metrics", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Infof("Received signal: %v", sig)
	log.Info("Initiating graceful shutdown...")

	// Stop metrics updater
	close(stopMetrics)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf(err, "Server forced to shutdown")
	}

	// Log final statistics
	log.Info("Logging final statistics...")
	db.LogStats()
	redis.LogStats()

	log.Info("Server shut down successfully")
}
