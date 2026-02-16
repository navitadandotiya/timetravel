package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/handler/v1"
)

// -------------------------
// Simulated errors
// -------------------------
var (
	ErrNotFound   = errors.New("record not found")
	ErrInvalidID  = errors.New("invalid ID")
)

// -------------------------
// Minimal mock service
// -------------------------
type mockRecordService struct{}

func (m *mockRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrInvalidID
	}
	if id == 42 {
		return entity.Record{ID: 42, Data: map[string]string{"info": "ok"}}, nil
	}
	return entity.Record{}, ErrNotFound
}

func (m *mockRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrInvalidID
	}
	rec := entity.Record{ID: id, Data: make(map[string]string)}
	for k, v := range updates {
		if v != nil {
			rec.Data[k] = *v
		}
	}
	return rec, nil
}

func (m *mockRecordService) CreateRecord(ctx context.Context, rec entity.Record) error {
	if rec.ID <= 0 {
		return ErrInvalidID
	}
	return nil
}

// -------------------------
// Helper functions
// -------------------------
func newTestAPI() *v1.API {
	return v1.NewAPI(&mockRecordService{})
}

func newTestRouter() *mux.Router {
	api := newTestAPI()
	r := mux.NewRouter()
	api.CreateRoutes(r)
	return r
}

func ptr(s string) *string { return &s }

// -------------------------
// Tests
// -------------------------

func TestGetRecord_Success(t *testing.T) {
	router := newTestRouter()
	req := httptest.NewRequest("GET", "/records/42", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp entity.Record
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.ID != 42 || resp.Data["info"] != "ok" {
		t.Fatalf("unexpected record data: %+v", resp)
	}
}


func TestGetRecord_InvalidID(t *testing.T) {
	router := newTestRouter()
	req := httptest.NewRequest("GET", "/records/0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestUpdateRecord_Success(t *testing.T) {
	router := newTestRouter()
	updates := map[string]*string{"info": ptr("updated")}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest("POST", "/records/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp entity.Record
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != 1 || resp.Data["info"] != "updated" {
		t.Fatalf("unexpected record data: %+v", resp)
	}
}

func TestUpdateRecord_InvalidID(t *testing.T) {
	router := newTestRouter()
	updates := map[string]*string{"info": ptr("x")}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest("POST", "/records/0", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}


func TestGetRecord_NotFound(t *testing.T) {
	mockSvc := &mockRecordService{} // mock returns ErrNotFound for any ID
	api := v1.NewAPI(mockSvc)
	router := mux.NewRouter()
	api.CreateRoutes(router)

	// Use a valid positive ID that doesn't exist
	req := httptest.NewRequest("GET", "/records/9999", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 404 Not Found, got %d", rec.Code)
	}
}



