package v2_test

import (
	"bytes"
	"context"
	
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/handler/v2"
)

//
// ---------------- MOCKS ----------------
//

type mockController struct{}

func (m *mockController) UpsertRecord(ctx context.Context, id int64, data map[string]string) (entity.PolicyholderRecord, error) {
	if id == 500 {
		return entity.PolicyholderRecord{}, errors.New("db error")
	}
	return entity.PolicyholderRecord{
		ID:        1,
		Version:   1,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockController) GetRecord(ctx context.Context, id int64) (entity.PolicyholderRecord, error) {
	if id == 404 {
		return entity.PolicyholderRecord{}, controller.ErrRecordDoesNotExist
	}
	return entity.PolicyholderRecord{
		ID:        1,
		Version:   2,
		Data:      map[string]string{"name": "john"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockController) GetVersion(ctx context.Context, id int, version int) (map[string]string, error) {
	if version == 404 {
		return nil, errors.New("version not found")
	}
	return map[string]string{"name": "v1"}, nil
}

func (m *mockController) ListVersions(ctx context.Context, id int) ([]int, error) {
	if id == 500 {
		return nil, errors.New("error")
	}
	return []int{1, 2, 3}, nil
}

type mockFlags struct {
	enabled bool
}

func (m *mockFlags) IsEnabled(ctx context.Context, key string) bool {
	return m.enabled
}

func (m *mockFlags) Refresh() error {
	return nil
}

//
// ---------------- HELPERS ----------------
//

func newTestRouter(flagEnabled bool) *mux.Router {
	api := &v2.API{
		Controller: &mockController{},
		Flags:      &mockFlags{enabled: flagEnabled},
	}

	r := mux.NewRouter()
	api.CreateRoutes(r)
	return r
}

//
// ---------------- TESTS ----------------
//

func TestHealthCheck(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("POST", "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestHealthCheck_Disabled(t *testing.T) {
	router := newTestRouter(false)

	req := httptest.NewRequest("POST", "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", rec.Code)
	}
}

func TestUpsertRecord_Success(t *testing.T) {
	router := newTestRouter(true)

	body := bytes.NewBuffer([]byte(`{"name":"john"}`))
	req := httptest.NewRequest("POST", "/records/1", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestUpsertRecord_InvalidID(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("POST", "/records/abc", bytes.NewBuffer([]byte(`{}`)))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rec.Code)
	}
}

func TestUpsertRecord_InvalidJSON(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("POST", "/records/1", bytes.NewBuffer([]byte(`invalid`)))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rec.Code)
	}
}

func TestGetRecord_Success(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("GET", "/records/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestGetRecord_NotFound(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("GET", "/records/404", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rec.Code)
	}
}

func TestRefreshFlags(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("POST", "/admin/refresh-flags", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestListVersions(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("GET", "/records/1/versions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestGetVersion_Success(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("GET", "/records/1/versions/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
}

func TestGetVersion_NotFound(t *testing.T) {
	router := newTestRouter(true)

	req := httptest.NewRequest("GET", "/records/1/versions/404", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rec.Code)
	}
}
