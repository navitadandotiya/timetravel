package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rainbowmga/timetravel/gateways"
)

var (
	// HTTPRequestTotal counts requests by method, path, and status
	HTTPRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDurationSeconds histogram for latency
	HTTPRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// FeatureFlagEvaluations
	FlagEvaluations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feature_flag_evaluations_total",
			Help: "Total feature flag evaluations",
		},
		[]string{"flag", "enabled"},
	)
)

// Handler returns the Prometheus /metrics handler
var metricsRepo *gateways.MetricsRepository

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func InitMetricsRepository(repo *gateways.MetricsRepository) {
	metricsRepo = repo
}



// RecordRequest increments request count and observes duration
func RecordRequest(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)
	HTTPRequestTotal.WithLabelValues(method, path, statusStr).Inc()
	HTTPRequestDurationSeconds.WithLabelValues(method, path).Observe(duration.Seconds())
	// Persist to DB
	if metricsRepo != nil {
		_ = metricsRepo.InsertMetric(
			"platform",
			"http_requests_total",
			1,
			"us-east-1", // you can make region dynamic
		)

		_ = metricsRepo.InsertMetric(
			"platform",
			"http_request_duration_seconds",
			duration.Seconds(),
			"us-east-1",
		)
	}
}

func FlagEvaluated(flag string, enabled bool) {
	state := "false"
	if enabled {
		state = "true"
	}
	FlagEvaluations.WithLabelValues(flag, state).Inc()
}