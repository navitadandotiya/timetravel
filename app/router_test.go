package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rainbowmga/timetravel/app"
	"github.com/rainbowmga/timetravel/gateways"
)

// setupSharedInMemoryDB creates a shared in-memory DB and all required tables.
func setupSharedInMemoryDB(t *testing.T) string {
	t.Helper()
	dbPath := "file:testdb?mode=memory&cache=shared"
	db := gateways.ConnectDB(dbPath, "") // no file
	schema := `
	PRAGMA foreign_keys = OFF;

--------------------------------------------------
-- FEATURE FLAGS
--------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    rollout_percentage INTEGER DEFAULT 100
);

INSERT INTO feature_flags(flag_key, enabled, description, updated_at, rollout_percentage)
VALUES
('enable_v2_api', 1, 'Enable version 2 API', CURRENT_TIMESTAMP, 100),
('enable_audit_logging', 1, 'Audit history logging', CURRENT_TIMESTAMP, 100),
('enable_metrics', 1, 'Observability metrics', CURRENT_TIMESTAMP, 100)
ON CONFLICT(flag_key) DO UPDATE SET
    enabled = excluded.enabled,
    description = excluded.description,
    updated_at = CURRENT_TIMESTAMP,
    rollout_percentage = excluded.rollout_percentage;

CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    country_code TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS policyholder_records (
    record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    policyholder_id INTEGER NOT NULL,
    data TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(policyholder_id) REFERENCES policyholders(policyholder_id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS audit_history (
    audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL,
    event_type TEXT NOT NULL,
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS event_logs (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id) ON DELETE CASCADE
);

--------------------------------------------------
-- OBSERVABILITY METRICS
--------------------------------------------------
CREATE TABLE IF NOT EXISTS observability_metrics (
    metric_id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    value REAL,
    region TEXT,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- SCHEMA MIGRATION TRACKING
--------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY);

INSERT INTO schema_migrations(version)
VALUES('001_create_users.sql')
ON CONFLICT(version) DO NOTHING;


PRAGMA foreign_keys = ON;
`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	// leave db open if needed, or just return dbPath for BuildRouter
	return dbPath
}

// -------------------------
// Helper to assert {"ok": true}
// -------------------------
func assertOkResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	var resp map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if ok, exists := resp["ok"]; !exists || !ok {
		t.Fatalf("expected ok=true in response, got %v", resp)
	}
}

// -------------------------
// Tests
// -------------------------
func TestBuildRouter_V1Health(t *testing.T) {
	dbPath := setupSharedInMemoryDB(t)

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assertOkResponse(t, rec)
}

func TestBuildRouter_V2Health(t *testing.T) {
	dbPath := setupSharedInMemoryDB(t)

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v2/health", nil)
	req.Header.Set("X-User-ID", "123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assertOkResponse(t, rec)
}

func TestBuildRouter_Metrics(t *testing.T) {
	dbPath := setupSharedInMemoryDB(t)

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
}

func TestBuildRouter_V2ProtectedRoute_NoUser(t *testing.T) {
	dbPath := setupSharedInMemoryDB(t)

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v2/records", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusNotFound {
		t.Fatalf("expected 401 Unauthorized or 404, got %d", rec.Code)
	}
}

func TestBuildRouter_V2ProtectedRoute_WithUser(t *testing.T) {
	dbPath := setupSharedInMemoryDB(t)

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v2/records", nil)
	req.Header.Set("X-User-ID", "123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Fatalf("expected 200 OK or 404 Not Found, got %d", rec.Code)
	}
}
