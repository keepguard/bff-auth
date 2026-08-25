package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics representa o sistema de métricas
type Metrics struct {
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	rateLimitBlockedTotal *prometheus.CounterVec
	upstreamRequests      *prometheus.CounterVec
	upstreamDuration      *prometheus.HistogramVec
	upstreamErrors        *prometheus.CounterVec
	cacheHits             *prometheus.CounterVec
	cacheMisses           *prometheus.CounterVec
	circuitBreakerState   *prometheus.GaugeVec
	rabbitMQPublish       *prometheus.CounterVec
}

// New cria um novo sistema de métricas
func New() *Metrics {
	return &Metrics{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_server_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "route", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_server_requests_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
		rateLimitBlockedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_blocked_total",
				Help: "Total number of requests blocked by rate limiting",
			},
			[]string{"action", "identifier_type"},
		),
		upstreamRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "upstream_requests_total",
				Help: "Total number of upstream service requests",
			},
			[]string{"service", "method", "endpoint", "status"},
		),
		upstreamDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "upstream_request_seconds",
				Help:    "Upstream service request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint"},
		),
		upstreamErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "upstream_errors_total",
				Help: "Total number of upstream service errors",
			},
			[]string{"service", "method", "endpoint", "error_type"},
		),
		cacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache", "key_pattern"},
		),
		cacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache", "key_pattern"},
		),
		circuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"service"},
		),
		rabbitMQPublish: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rabbitmq_publish_total",
				Help: "Total number of RabbitMQ message publications",
			},
			[]string{"exchange", "routing_key", "status"},
		),
	}
}

// RecordHTTPRequest registra uma requisição HTTP
func (m *Metrics) RecordHTTPRequest(method, route string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)

	m.httpRequestsTotal.WithLabelValues(method, route, status).Inc()
	m.httpRequestDuration.WithLabelValues(method, route, status).Observe(duration.Seconds())
}

// RecordRateLimitBlocked registra um bloqueio por rate limit
func (m *Metrics) RecordRateLimitBlocked(action, identifierType string) {
	m.rateLimitBlockedTotal.WithLabelValues(action, identifierType).Inc()
}

// RecordUpstreamRequest registra uma requisição para um serviço upstream
func (m *Metrics) RecordUpstreamRequest(service, method, endpoint string, statusCode int, duration time.Duration) {
	status := http.StatusText(statusCode)
	if status == "" {
		status = "unknown"
	}

	m.upstreamRequests.WithLabelValues(service, method, endpoint, status).Inc()
	m.upstreamDuration.WithLabelValues(service, method, endpoint).Observe(duration.Seconds())
}

// RecordUpstreamError registra um erro de serviço upstream
func (m *Metrics) RecordUpstreamError(service, method, endpoint, errorType string) {
	m.upstreamErrors.WithLabelValues(service, method, endpoint, errorType).Inc()
}

// RecordCacheHit registra um hit no cache
func (m *Metrics) RecordCacheHit(cache, keyPattern string) {
	m.cacheHits.WithLabelValues(cache, keyPattern).Inc()
}

// RecordCacheMiss registra um miss no cache
func (m *Metrics) RecordCacheMiss(cache, keyPattern string) {
	m.cacheMisses.WithLabelValues(cache, keyPattern).Inc()
}

// SetCircuitBreakerState define o estado do circuit breaker
func (m *Metrics) SetCircuitBreakerState(service string, state int) {
	m.circuitBreakerState.WithLabelValues(service).Set(float64(state))
}

// RecordRabbitMQPublish registra uma publicação no RabbitMQ
func (m *Metrics) RecordRabbitMQPublish(exchange, routingKey, status string, duration time.Duration) {
	m.rabbitMQPublish.WithLabelValues(exchange, routingKey, status).Inc()
}

// CircuitBreakerState representa o estado do circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// Handler retorna o handler HTTP para métricas Prometheus
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}
