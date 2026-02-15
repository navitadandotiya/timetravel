package observability

import (
	"net/http"
	"regexp"
	"time"
	"context"
	"strconv"

	"github.com/rainbowmga/timetravel/common"
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

func RequireUserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userHeader := r.Header.Get("X-User-ID")
		if userHeader == "" {
			http.Error(w, "missing X-User-ID header", http.StatusUnauthorized)
			DefaultLogger.Error("missing X-User-ID header")
			return
		}

		userID, err := strconv.ParseInt(userHeader, 10, 64)
		if err != nil || userID <= 0 {
			DefaultLogger.Error("invalid X-User-ID header")
			http.Error(w, "invalid X-User-ID header", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), common.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}