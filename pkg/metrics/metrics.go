// Package metrics provides Prometheus metrics instrumentation
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// LLM API metrics
	LLMRequestsTotal   *prometheus.CounterVec
	LLMRequestDuration *prometheus.HistogramVec
	LLMTokensTotal     *prometheus.CounterVec
	LLMCostTotal       *prometheus.CounterVec
	LLMErrorsTotal     *prometheus.CounterVec

	// Cache metrics
	CacheRequestsTotal *prometheus.CounterVec
	CacheHitsTotal     prometheus.Counter
	CacheMissesTotal   prometheus.Counter
	CacheErrorsTotal   prometheus.Counter
	CacheSize          prometheus.Gauge
	CacheDuration      *prometheus.HistogramVec

	// OAuth metrics
	OAuthTokensIssued  *prometheus.CounterVec
	OAuthTokensRevoked prometheus.Counter
	OAuthErrorsTotal   *prometheus.CounterVec

	// Database metrics
	DBConnectionsOpen prometheus.Gauge
	DBConnectionsIdle prometheus.Gauge
	DBQueryDuration   *prometheus.HistogramVec
	DBQueriesTotal    *prometheus.CounterVec
	DBErrorsTotal     prometheus.Counter

	// Provider metrics
	ProviderHealthStatus  *prometheus.GaugeVec
	ProviderRequestsTotal *prometheus.CounterVec
}

// New creates and registers all Prometheus metrics
func New(namespace string) *Metrics {
	if namespace == "" {
		namespace = "llm_proxy"
	}

	m := &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),

		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// LLM API metrics
		LLMRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "llm_requests_total",
				Help:      "Total number of LLM API requests",
			},
			[]string{"model", "provider", "status"},
		),

		LLMRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "llm_request_duration_seconds",
				Help:      "LLM API request duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"model", "provider"},
		),

		LLMTokensTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "llm_tokens_total",
				Help:      "Total number of tokens processed",
			},
			[]string{"model", "provider", "type"}, // type: input, output
		),

		LLMCostTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "llm_cost_total",
				Help:      "Total cost of LLM requests in USD",
			},
			[]string{"model", "provider", "client_id"},
		),

		LLMErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "llm_errors_total",
				Help:      "Total number of LLM API errors",
			},
			[]string{"model", "provider", "error_type"},
		),

		// Cache metrics
		CacheRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_requests_total",
				Help:      "Total number of cache requests",
			},
			[]string{"operation"}, // operation: get, set, delete
		),

		CacheHitsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),

		CacheMissesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),

		CacheErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_errors_total",
				Help:      "Total number of cache errors",
			},
		),

		CacheSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cache_size_bytes",
				Help:      "Current size of cache in bytes",
			},
		),

		CacheDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cache_operation_duration_seconds",
				Help:      "Cache operation duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation"},
		),

		// OAuth metrics
		OAuthTokensIssued: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "oauth_tokens_issued_total",
				Help:      "Total number of OAuth tokens issued",
			},
			[]string{"grant_type", "client_id"},
		),

		OAuthTokensRevoked: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "oauth_tokens_revoked_total",
				Help:      "Total number of OAuth tokens revoked",
			},
		),

		OAuthErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "oauth_errors_total",
				Help:      "Total number of OAuth errors",
			},
			[]string{"error_type"},
		),

		// Database metrics
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Current number of open database connections",
			},
		),

		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Current number of idle database connections",
			},
		),

		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"query_type"},
		),

		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"query_type", "status"},
		),

		DBErrorsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_errors_total",
				Help:      "Total number of database errors",
			},
		),

		// Provider metrics
		ProviderHealthStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "provider_health_status",
				Help:      "Health status of LLM providers (1=healthy, 0=unhealthy)",
			},
			[]string{"provider", "api_key_id"},
		),

		ProviderRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "provider_requests_total",
				Help:      "Total number of requests to each provider",
			},
			[]string{"provider", "api_key_id"},
		),
	}

	return m
}

// RecordHTTPRequest records HTTP request metrics
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordLLMRequest records LLM API request metrics
func (m *Metrics) RecordLLMRequest(model, provider, status string, duration time.Duration, inputTokens, outputTokens int, cost float64, clientID string) {
	m.LLMRequestsTotal.WithLabelValues(model, provider, status).Inc()
	m.LLMRequestDuration.WithLabelValues(model, provider).Observe(duration.Seconds())

	if inputTokens > 0 {
		m.LLMTokensTotal.WithLabelValues(model, provider, "input").Add(float64(inputTokens))
	}
	if outputTokens > 0 {
		m.LLMTokensTotal.WithLabelValues(model, provider, "output").Add(float64(outputTokens))
	}
	if cost > 0 {
		m.LLMCostTotal.WithLabelValues(model, provider, clientID).Add(cost)
	}
}

// RecordLLMError records LLM API error
func (m *Metrics) RecordLLMError(model, provider, errorType string) {
	m.LLMErrorsTotal.WithLabelValues(model, provider, errorType).Inc()
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(duration time.Duration) {
	m.CacheHitsTotal.Inc()
	m.CacheRequestsTotal.WithLabelValues("get").Inc()
	m.CacheDuration.WithLabelValues("get").Observe(duration.Seconds())
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(duration time.Duration) {
	m.CacheMissesTotal.Inc()
	m.CacheRequestsTotal.WithLabelValues("get").Inc()
	m.CacheDuration.WithLabelValues("get").Observe(duration.Seconds())
}

// RecordCacheSet records a cache set operation
func (m *Metrics) RecordCacheSet(duration time.Duration) {
	m.CacheRequestsTotal.WithLabelValues("set").Inc()
	m.CacheDuration.WithLabelValues("set").Observe(duration.Seconds())
}

// RecordCacheError records a cache error
func (m *Metrics) RecordCacheError() {
	m.CacheErrorsTotal.Inc()
}

// RecordOAuthToken records an OAuth token issuance
func (m *Metrics) RecordOAuthToken(grantType, clientID string) {
	m.OAuthTokensIssued.WithLabelValues(grantType, clientID).Inc()
}

// RecordOAuthRevoke records an OAuth token revocation
func (m *Metrics) RecordOAuthRevoke() {
	m.OAuthTokensRevoked.Inc()
}

// RecordOAuthError records an OAuth error
func (m *Metrics) RecordOAuthError(errorType string) {
	m.OAuthErrorsTotal.WithLabelValues(errorType).Inc()
}

// UpdateDBStats updates database connection metrics
func (m *Metrics) UpdateDBStats(open, idle int) {
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsIdle.Set(float64(idle))
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(queryType, status string, duration time.Duration) {
	m.DBQueriesTotal.WithLabelValues(queryType, status).Inc()
	m.DBQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// RecordDBError records a database error
func (m *Metrics) RecordDBError() {
	m.DBErrorsTotal.Inc()
}

// UpdateProviderHealth updates provider health status
func (m *Metrics) UpdateProviderHealth(provider, apiKeyID string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.ProviderHealthStatus.WithLabelValues(provider, apiKeyID).Set(value)
}

// RecordProviderRequest records a request to a provider
func (m *Metrics) RecordProviderRequest(provider, apiKeyID string) {
	m.ProviderRequestsTotal.WithLabelValues(provider, apiKeyID).Inc()
}
