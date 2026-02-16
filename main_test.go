package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rainbowmga/timetravel/gateways"
	"github.com/rainbowmga/timetravel/app"
	"time"
)

// prepareTestDB creates minimal tables for v2 controllers
func prepareTestDB() (string, func(), error) {
	dbPath := "file:testdb?mode=memory&cache=shared"

	db := gateways.ConnectDB(dbPath,"script/create_v2_tables.sql")
	_ = gateways.RunMigrations(db, "../script/migrations")


	_, _ = db.Exec(`
	CREATE TABLE IF NOT EXISTS feature_flags (
		flag_key TEXT PRIMARY KEY,
		enabled BOOLEAN NOT NULL,
		description TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		rollout_percentage INTEGER DEFAULT 100
	);
	CREATE TABLE IF NOT EXISTS policyholders (
		policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS policyholder_records (
		record_id INTEGER PRIMARY KEY AUTOINCREMENT,
		policyholder_id INTEGER NOT NULL,
		data TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS audit_history (
		audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		data TEXT NOT NULL,
		event_type TEXT NOT NULL,
		changed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS event_logs (
		event_id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER,
		action TEXT NOT NULL,
		details TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`)

	cleanup := func() {
		db.Close()
	}

	return dbPath, cleanup, nil
}


func TestBuildRouter_V1HealthExists(t *testing.T) {
	dbPath, cleanup, err := prepareTestDB()
	if err != nil {
		t.Fatalf("failed to prepare test DB: %v", err)
	}
	defer cleanup()

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.Code)
	}
}

func TestMetricsEndpointExists(t *testing.T) {
	dbPath, cleanup, err := prepareTestDB()
	if err != nil {
		t.Fatalf("failed to prepare test DB: %v", err)
	}
	defer cleanup()

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.Code)
	}
}

func TestV2RequiresUserContext(t *testing.T) {
	dbPath, cleanup, err := prepareTestDB()
	if err != nil {
		t.Fatalf("failed to prepare test DB: %v", err)
	}
	defer cleanup()

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	t.Log("=== v2 Health Check: missing X-User-ID ===")
	// Missing header → expect 401 Unauthorized
	reqMissing := httptest.NewRequest("POST", "/api/v2/health", nil)
	recMissing := httptest.NewRecorder()
	router.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized when X-User-ID missing, got %d", recMissing.Code)
	}

	t.Log("=== v2 Health Check: invalid X-User-ID ===")
	// Invalid header → expect 400 Bad Request
	reqInvalid := httptest.NewRequest("POST", "/api/v2/health", nil)
	reqInvalid.Header.Set("X-User-ID", "abc") // invalid int64
	recInvalid := httptest.NewRecorder()
	router.ServeHTTP(recInvalid, reqInvalid)
	if recInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request with invalid X-User-ID, got %d", recInvalid.Code)
	}

	t.Log("=== v2 Health Check: valid X-User-ID ===")
	// Valid header → expect 200 OK
	reqValid := httptest.NewRequest("POST", "/api/v2/health", nil)
	reqValid.Header.Set("X-User-ID", "123")
	recValid := httptest.NewRecorder()
	router.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("expected 200 OK with valid X-User-ID header, got %d", recValid.Code)
	}
}

// TestRunServer initializes server with test config and shuts it down
func TestRunServer(t *testing.T) {
	// Run server in goroutine to avoid blocking
	go func() {
		err := RunServer("conf/config.yaml", "")
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("server failed to start: %v", err)
		}
	}()

	dbPath, cleanup, err := prepareTestDB()
	if err != nil {
		t.Fatalf("failed to prepare test DB: %v", err)
	}
	defer cleanup()

	router, err := app.BuildRouter(dbPath, false)
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	// Wait a short moment for server to start
	time.Sleep(200 * time.Millisecond)

	// Optionally, make an HTTP request to test /health endpoint
	req := httptest.NewRequest("POST", "/api/v1/health", nil) // use GET instead of POST
	rec := httptest.NewRecorder()
	req.Header.Set("X-User-ID", "123")
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	// TODO: Shutdown server gracefully if needed
}