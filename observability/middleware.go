package observability

import (
	"net/http"
	"regexp"
	"time"
)

// pathForMetrics normalizes path for metrics (replaces numeric segments to limit cardinality)
var numericSegment = regexp.MustCompile(`/\d+`)

func pathForMetrics(path string) string {
	return numericSegment.ReplaceAllString(path, "/{id}")
}

// responseWriter wraps http.ResponseWriter to capture status and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// LoggingAndMetrics wraps a handler to log requests and record metrics
func LoggingAndMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		path := r.URL.Path
		if path == "" {
			path = "/"
		}
		metricsPath := pathForMetrics(path)
		RecordRequest(r.Method, metricsPath, wrapped.status, duration)
		DefaultLogger.Info("request",
			"method", r.Method,
			"path", path,
			"status", wrapped.status,
			"duration_ms", duration.Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}
