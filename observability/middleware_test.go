package observability

import (
	"net/http"
	"net/http/httptest"
	
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/rainbowmga/timetravel/common"
)

func TestPathForMetrics(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/records/123", "/records/{id}"},
		{"/records/123/versions/456", "/records/{id}/versions/{id}"},
		{"/static/path", "/static/path"},
		{"", ""},
	}

	for _, tt := range tests {
		got := pathForMetrics(tt.input)
		if got != tt.expected {
			t.Errorf("pathForMetrics(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestResponseWriterCapturesStatusAndSize(t *testing.T) {
	rr := httptest.NewRecorder()
	w := &responseWriter{ResponseWriter: rr, status: http.StatusOK}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("hello"))

	if w.status != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.status)
	}

	if w.size != 5 {
		t.Errorf("expected size 5, got %d", w.size)
	}
}

func TestLoggingAndMetrics(t *testing.T) {
	// reset metric counter before test
	before := testutil.ToFloat64(
		HTTPRequestTotal.WithLabelValues("GET", "/records/{id}", "200"),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	wrapped := LoggingAndMetrics(handler)

	req := httptest.NewRequest("GET", "/records/123", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK")
	}

	after := testutil.ToFloat64(
		HTTPRequestTotal.WithLabelValues("GET", "/records/{id}", "200"),
	)

	if after != before+1 {
		t.Errorf("expected metrics counter increment")
	}
}

func TestLoggingAndMetrics_EmptyPath(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := LoggingAndMetrics(handler)

	req := httptest.NewRequest("GET", "/", nil)

	// Simulate empty path AFTER creation
	req.URL.Path = ""

	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK")
	}
}


func TestRequireUserContext_MissingHeader(t *testing.T) {
	handler := RequireUserContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header")
	}
}

func TestRequireUserContext_InvalidHeader(t *testing.T) {
	handler := RequireUserContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-User-ID", "abc")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid header")
	}
}

func TestRequireUserContext_ValidHeader(t *testing.T) {
	handler := RequireUserContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(common.UserIDKey)
		if val == nil {
			t.Errorf("expected user ID in context")
		}
		if val.(int64) != 42 {
			t.Errorf("expected user ID 42")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-User-ID", "42")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK")
	}
}
