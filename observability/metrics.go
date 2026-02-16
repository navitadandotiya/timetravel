package observability

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// -----------------------------
// MetricsRepo Interface
// -----------------------------
type MetricsRepo interface {
	InsertMetric(metricType, metricName string, value float64, region string) error
	// optional for DB collectors: SumMetric(metricName string) (float64, error)
}

// -----------------------------
// Globals
// -----------------------------
var (
	metricsRepo  MetricsRepo
	registerOnce sync.Once

	HTTPRequestTotal           *prometheus.CounterVec
	HTTPRequestDurationSeconds *prometheus.HistogramVec
	FlagEvaluations            *prometheus.CounterVec
)

// -----------------------------
// InitMetricsRepository
// -----------------------------
func InitMetricsRepository(repo MetricsRepo) {
	metricsRepo = repo

	registerOnce.Do(func() {
		HTTPRequestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)

		HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		FlagEvaluations = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "feature_flag_evaluations_total",
				Help: "Total feature flag evaluations",
			},
			[]string{"flag", "enabled"},
		)

		prometheus.MustRegister(HTTPRequestTotal, HTTPRequestDurationSeconds, FlagEvaluations)
	})
}

// -----------------------------
// MetricsHandler
// -----------------------------
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// -----------------------------
// RecordRequest
// -----------------------------
func RecordRequest(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)

	if HTTPRequestTotal != nil {
		HTTPRequestTotal.WithLabelValues(method, path, statusStr).Inc()
		HTTPRequestDurationSeconds.WithLabelValues(method, path).Observe(duration.Seconds())
	}

	if metricsRepo != nil {
		_ = metricsRepo.InsertMetric("platform", "http_requests_total", 1, "us-east-1")
		_ = metricsRepo.InsertMetric("platform", "http_request_duration_seconds", duration.Seconds(), "us-east-1")
	}
}

// -----------------------------
// FlagEvaluated
// -----------------------------
func FlagEvaluated(flag string, enabled bool) {
	state := "false"
	if enabled {
		state = "true"
	}

	if FlagEvaluations != nil {
		FlagEvaluations.WithLabelValues(flag, state).Inc()
	}
}
