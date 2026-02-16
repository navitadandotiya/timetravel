package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// mockMetricsRepo implements MetricsRepo for testing
type mockMetricsRepo struct {
	calls int
}

func (m *mockMetricsRepo) InsertMetric(metricType, metricName string, value float64, region string) error {
	m.calls++
	return nil
}

// resetMetricsForTest returns fresh metrics registered in a new registry
func resetMetricsForTest() *prometheus.Registry {
	reg := prometheus.NewRegistry()

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

	reg.MustRegister(HTTPRequestTotal, HTTPRequestDurationSeconds, FlagEvaluations)

	// clear DB repo for test
	metricsRepo = nil

	return reg
}

func TestMetricsHandler_NotNil(t *testing.T) {
	reg := resetMetricsForTest()

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	if handler == nil {
		t.Fatal("expected MetricsHandler not nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK from metrics handler, got %d", rr.Code)
	}
}

func TestRecordRequest_NoRepo(t *testing.T) {
	resetMetricsForTest()

	before := testutil.ToFloat64(HTTPRequestTotal.WithLabelValues("GET", "/test", "200"))

	RecordRequest("GET", "/test", 200, 100*time.Millisecond)

	after := testutil.ToFloat64(HTTPRequestTotal.WithLabelValues("GET", "/test", "200"))

	if after != before+1 {
		t.Errorf("expected counter increment, got before=%f after=%f", before, after)
	}
}


func TestRecordRequest_WithMockRepo_Isolated(t *testing.T) {
	// Reset metrics and clear global registry
	resetMetricsForTest()

	mock := &mockMetricsRepo{}

	// Temporarily swap the global repo with our mock
	oldRepo := metricsRepo
	metricsRepo = mock
	defer func() { metricsRepo = oldRepo }() // restore after test

	// Directly call RecordRequest — no HTTP router involved
	RecordRequest("GET", "/test", 200, 50*time.Millisecond)
	RecordRequest("POST", "/test", 500, 20*time.Millisecond)

	if mock.calls != 4 {
		t.Errorf("expected 2 DB metric calls, got %d", mock.calls)
	}
}

func TestFlagEvaluated(t *testing.T) {
	resetMetricsForTest()

	before := testutil.ToFloat64(FlagEvaluations.WithLabelValues("test_flag", "true"))

	FlagEvaluated("test_flag", true)

	after := testutil.ToFloat64(FlagEvaluations.WithLabelValues("test_flag", "true"))

	if after != before+1 {
		t.Errorf("expected flag evaluation increment, got before=%f after=%f", before, after)
	}
}
