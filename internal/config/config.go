// Package config provides configuration management using Viper.
// It supports YAML files, environment variables, and defaults.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server        ServerConfig         `mapstructure:"server"`
	Database      DatabaseConfig       `mapstructure:"database"`
	Redis         RedisConfig          `mapstructure:"redis"`
	OAuth         OAuthConfig          `mapstructure:"oauth"`
	Admin         AdminConfig          `mapstructure:"admin"`
	ClientAPIKeys []ClientAPIKeyConfig `mapstructure:"client_api_keys"`
	Providers     ProvidersConfig      `mapstructure:"providers"`
	Cache         CacheConfig          `mapstructure:"cache"`
	RateLimiting  RateLimitingConfig   `mapstructure:"rate_limiting"`
	Logging       LoggingConfig        `mapstructure:"logging"`
	Metrics       MetricsConfig        `mapstructure:"metrics"`
	EncryptionKey string               `mapstructure:"encryption_key"` // 32-byte hex key for encrypting secrets in DB
	SMTP          SMTPConfig           `mapstructure:"smtp"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Timeout      time.Duration `mapstructure:"timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	CORSOrigins  []string      `mapstructure:"cors_origins"`
}

// Environment returns the current environment name from the ENVIRONMENT env var
func Environment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		return "development"
	}
	return env
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	return Environment() == "production"
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Database       string `mapstructure:"database"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	MaxConnections int    `mapstructure:"max_connections"`
	SSLMode        string `mapstructure:"ssl_mode"`
}

// DSN returns the database connection string
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	DB         int    `mapstructure:"db"`
	Password   string `mapstructure:"password"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

// Address returns the Redis address
func (r RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// OAuthConfig holds OAuth 2.0 configuration
type OAuthConfig struct {
	JWTSecret         string                            `mapstructure:"jwt_secret"`
	AccessTokenTTL    time.Duration                     `mapstructure:"access_token_ttl"`
	RefreshTokenTTL   time.Duration                     `mapstructure:"refresh_token_ttl"`
	Issuer            string                            `mapstructure:"issuer"`
	ExternalProviders map[string]ExternalProviderConfig `mapstructure:"external_providers"`
}

// ExternalProviderConfig holds configuration for external OAuth providers
type ExternalProviderConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri"`
}

// AdminConfig holds admin API configuration
type AdminConfig struct {
	APIKeys []string `mapstructure:"api_keys"`
}

// ClientAPIKeyConfig holds a single client API key configuration for OpenAI-compatible clients
type ClientAPIKeyConfig struct {
	Key     string   `mapstructure:"key"`
	Name    string   `mapstructure:"name"`
	Scopes  []string `mapstructure:"scopes"`
	Enabled bool     `mapstructure:"enabled"`
}

// ProvidersConfig holds LLM provider configuration
type ProvidersConfig struct {
	Claude ClaudeConfig `mapstructure:"claude"`
	OpenAI OpenAIConfig `mapstructure:"openai"`
	Abacus AbacusConfig `mapstructure:"abacus"`
}

// ClaudeConfig holds Anthropic Claude configuration
type ClaudeConfig struct {
	Enabled bool           `mapstructure:"enabled"`
	APIKeys []ClaudeAPIKey `mapstructure:"api_keys"`
	Models  []string       `mapstructure:"models"`
	Timeout time.Duration  `mapstructure:"timeout"`
	Retry   RetryConfig    `mapstructure:"retry"`
}

// ClaudeAPIKey holds a single Claude API key configuration
type ClaudeAPIKey struct {
	Key    string `mapstructure:"key"`
	Weight int    `mapstructure:"weight"`
	MaxRPM int    `mapstructure:"max_rpm"`
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	Enabled bool           `mapstructure:"enabled"`
	APIKeys []OpenAIAPIKey `mapstructure:"api_keys"`
	Models  []string       `mapstructure:"models"`
	Timeout time.Duration  `mapstructure:"timeout"`
}

// OpenAIAPIKey holds a single OpenAI API key configuration
type OpenAIAPIKey struct {
	Key    string `mapstructure:"key"`
	Weight int    `mapstructure:"weight"`
	MaxRPM int    `mapstructure:"max_rpm"`
}

// AbacusConfig holds Abacus.ai configuration
type AbacusConfig struct {
	Enabled      bool           `mapstructure:"enabled"`
	APIKeys      []AbacusAPIKey `mapstructure:"api_keys"`
	DeploymentID string         `mapstructure:"deployment_id"` // Default deployment ID
	Models       []string       `mapstructure:"models"`        // Supported LLM names
	Timeout      time.Duration  `mapstructure:"timeout"`
}

// AbacusAPIKey holds a single Abacus.ai API key configuration
type AbacusAPIKey struct {
	Key    string `mapstructure:"key"`
	Weight int    `mapstructure:"weight"`
	MaxRPM int    `mapstructure:"max_rpm"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts       int           `mapstructure:"max_attempts"`
	InitialBackoff    time.Duration `mapstructure:"initial_backoff"`
	MaxBackoff        time.Duration `mapstructure:"max_backoff"`
	BackoffMultiplier float64       `mapstructure:"backoff_multiplier"`
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	TTL     int    `mapstructure:"ttl"`
	MaxSize int    `mapstructure:"max_size"`
	Prefix  string `mapstructure:"prefix"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	Enabled    bool `mapstructure:"enabled"`
	DefaultRPM int  `mapstructure:"default_rpm"`
	DefaultRPD int  `mapstructure:"default_rpd"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// SMTPConfig holds SMTP configuration for sending emails (contact form)
type SMTPConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	From             string `mapstructure:"from"`
	To               string `mapstructure:"to"`               // Recipient for contact form submissions
	TurnstileSecret  string `mapstructure:"turnstile_secret"`  // Cloudflare Turnstile secret key
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file settings
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Read config file (optional - will use defaults if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; use defaults and env vars
	}

	// Environment variables
	v.SetEnvPrefix("") // No prefix for backward compatibility
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit env bindings for keys without defaults (Viper's AutomaticEnv
	// only works for keys it already knows about via defaults or config file)
	_ = v.BindEnv("oauth.jwt_secret", "OAUTH_JWT_SECRET")
	_ = v.BindEnv("oauth.access_token_ttl", "OAUTH_ACCESS_TOKEN_TTL")
	_ = v.BindEnv("oauth.refresh_token_ttl", "OAUTH_REFRESH_TOKEN_TTL")
	_ = v.BindEnv("oauth.issuer", "OAUTH_ISSUER")
	_ = v.BindEnv("admin.api_keys", "ADMIN_API_KEYS")
	_ = v.BindEnv("database.password", "DB_PASSWORD", "DATABASE_PASSWORD")
	_ = v.BindEnv("database.host", "DB_HOST", "DATABASE_HOST")
	_ = v.BindEnv("database.port", "DB_PORT", "DATABASE_PORT")
	_ = v.BindEnv("database.database", "DB_NAME", "DATABASE_DATABASE")
	_ = v.BindEnv("database.user", "DB_USER", "DATABASE_USER")
	_ = v.BindEnv("database.ssl_mode", "DB_SSL_MODE", "DATABASE_SSL_MODE")
	_ = v.BindEnv("encryption_key", "ENCRYPTION_KEY")
	_ = v.BindEnv("smtp.enabled", "SMTP_ENABLED")
	_ = v.BindEnv("smtp.host", "SMTP_HOST")
	_ = v.BindEnv("smtp.port", "SMTP_PORT")
	_ = v.BindEnv("smtp.username", "SMTP_USERNAME")
	_ = v.BindEnv("smtp.password", "SMTP_PASSWORD")
	_ = v.BindEnv("smtp.from", "SMTP_FROM")
	_ = v.BindEnv("smtp.to", "SMTP_TO")
	_ = v.BindEnv("smtp.turnstile_secret", "TURNSTILE_SECRET")

	// Handle comma-separated env vars for slice fields
	if corsOrigins := os.Getenv("SERVER_CORS_ORIGINS"); corsOrigins != "" {
		v.Set("server.cors_origins", strings.Split(corsOrigins, ","))
	}
	if adminKeys := os.Getenv("ADMIN_API_KEYS"); adminKeys != "" {
		v.Set("admin.api_keys", strings.Split(adminKeys, ","))
	}
	// Single ADMIN_API_KEY env var (backwards compatibility)
	if adminKey := os.Getenv("ADMIN_API_KEY"); adminKey != "" && os.Getenv("ADMIN_API_KEYS") == "" {
		v.Set("admin.api_keys", []string{adminKey})
	}
	// Provider API keys from env (simple single-key format)
	if claudeKey := os.Getenv("CLAUDE_API_KEY"); claudeKey != "" {
		v.Set("providers.claude.api_keys", []map[string]interface{}{
			{"key": claudeKey, "weight": 1, "max_rpm": 60},
		})
	}
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		v.Set("providers.openai.enabled", true)
		v.Set("providers.openai.api_keys", []map[string]interface{}{
			{"key": openaiKey, "weight": 1, "max_rpm": 60},
		})
	}

	// Unmarshal into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Environment-aware warnings
	if IsProduction() {
		if config.Database.SSLMode == "disable" {
			log.Println("[WARNING] Production environment with database ssl_mode=disable — consider using 'require' or 'verify-full'")
		}
		if config.Logging.Level == "debug" {
			log.Println("[WARNING] Production environment with log level 'debug' — consider using 'info' or 'warn'")
		}
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.timeout", "300s")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.cors_origins", []string{
		"http://localhost:5173",
		"http://localhost:3005",
		"http://localhost:3000",
	})

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5433)
	v.SetDefault("database.database", "llm_proxy")
	v.SetDefault("database.user", "proxy_user")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.ssl_mode", "disable")

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6380)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.pool_size", 10)

	// OAuth
	v.SetDefault("oauth.access_token_ttl", "3600s")
	v.SetDefault("oauth.refresh_token_ttl", "2592000s")
	v.SetDefault("oauth.issuer", "llm-proxy")

	// Providers - Claude
	v.SetDefault("providers.claude.enabled", true)
	v.SetDefault("providers.claude.timeout", "120s")
	v.SetDefault("providers.claude.retry.max_attempts", 3)
	v.SetDefault("providers.claude.retry.initial_backoff", "1s")
	v.SetDefault("providers.claude.retry.max_backoff", "10s")
	v.SetDefault("providers.claude.retry.backoff_multiplier", 2.0)
	v.SetDefault("providers.claude.models", []string{
		"claude-3-opus-20240229",
		"claude-3-haiku-20240307",
	})

	// Providers - OpenAI
	v.SetDefault("providers.openai.enabled", false)
	v.SetDefault("providers.openai.timeout", "120s")
	v.SetDefault("providers.openai.models", []string{
		"gpt-4-turbo-preview",
		"gpt-4",
		"gpt-3.5-turbo",
	})

	// Cache
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.ttl", 3600)
	v.SetDefault("cache.max_size", 1000)
	v.SetDefault("cache.prefix", "llm-proxy:")

	// Rate Limiting
	v.SetDefault("rate_limiting.enabled", true)
	v.SetDefault("rate_limiting.default_rpm", 1000)
	v.SetDefault("rate_limiting.default_rpd", 50000)

	// Logging
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9090)
	v.SetDefault("metrics.path", "/metrics")

	// SMTP (for contact form)
	v.SetDefault("smtp.enabled", false)
	v.SetDefault("smtp.host", "smtp.protonmail.ch")
	v.SetDefault("smtp.port", 587)
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Server
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Database
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Redis
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	// OAuth
	if cfg.OAuth.JWTSecret == "" {
		return fmt.Errorf("oauth jwt_secret is required")
	}
	if len(cfg.OAuth.JWTSecret) < 32 {
		return fmt.Errorf("oauth jwt_secret must be at least 32 characters (64+ recommended)")
	}

	// Admin
	if len(cfg.Admin.APIKeys) == 0 {
		return fmt.Errorf("at least one admin API key is required")
	}
	for _, key := range cfg.Admin.APIKeys {
		if len(key) < 32 {
			return fmt.Errorf("admin API keys must be at least 32 characters (got %d)", len(key))
		}
	}

	// Providers
	if cfg.Providers.Claude.Enabled {
		if len(cfg.Providers.Claude.APIKeys) == 0 {
			return fmt.Errorf("at least one Claude API key is required when Claude is enabled")
		}
	}

	if cfg.Providers.OpenAI.Enabled {
		if len(cfg.Providers.OpenAI.APIKeys) == 0 {
			return fmt.Errorf("at least one OpenAI API key is required when OpenAI is enabled")
		}
	}

	if cfg.Providers.Abacus.Enabled {
		if len(cfg.Providers.Abacus.APIKeys) == 0 {
			return fmt.Errorf("at least one Abacus API key is required when Abacus is enabled")
		}
	}

	// At least one provider must be enabled
	if !cfg.Providers.Claude.Enabled && !cfg.Providers.OpenAI.Enabled && !cfg.Providers.Abacus.Enabled {
		return fmt.Errorf("at least one provider must be enabled")
	}

	// Logging
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", cfg.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", cfg.Logging.Format)
	}

	return nil
}

// MustLoad loads configuration or panics on error
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}
