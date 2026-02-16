package service_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/rainbowmga/timetravel/service"
	_ "github.com/mattn/go-sqlite3"
)

func createRecordTestDB(t *testing.T) (string, func()) {
	t.Helper()

	file, err := os.CreateTemp("", "record-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	path := file.Name()
	file.Close()

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	schema := `
	CREATE TABLE policyholders (
		policyholder_id INTEGER PRIMARY KEY,
		name TEXT
	);

	CREATE TABLE policyholder_records (
		record_id INTEGER PRIMARY KEY AUTOINCREMENT,
		policyholder_id INTEGER,
		data TEXT,
		version INTEGER,
		created_at TEXT,
		updated_at TEXT
	);

	CREATE TABLE audit_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER,
		version INTEGER,
		data TEXT,
		changed_at TEXT,
		event_type TEXT
	);

	CREATE TABLE event_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER,
		action TEXT,
		timestamp TEXT,
		details TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	db.Close()

	return path, func() { os.Remove(path) }
}

func TestCreateOrUpdate_CreateThenUpdate(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, err := service.NewSQLiteRecordService(path)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// ---- CREATE ----
	data := map[string]string{"name": "John"}
	record, err := svc.CreateOrUpdate(1, data)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if record.Version != 1 {
		t.Errorf("expected version 1, got %d", record.Version)
	}

	// ---- UPDATE ----
	data2 := map[string]string{"name": "John Updated"}
	record2, err := svc.CreateOrUpdate(1, data2)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if record2.Version != 2 {
		t.Errorf("expected version 2, got %d", record2.Version)
	}
}

func TestGet_RecordExists(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "Alice"})

	record, err := svc.Get(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.Version != 1 {
		t.Errorf("expected version 1, got %d", record.Version)
	}

	if record.Data["name"] != "Alice" {
		t.Errorf("unexpected data")
	}
}

func TestGet_RecordDoesNotExist(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	_, err := svc.Get(999)
	if err != service.ErrRecordDoesNotExist {
		t.Errorf("expected ErrRecordDoesNotExist, got %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "V1"})
	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "V2"})

	v1, err := svc.GetVersion(1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v1["name"] != "V1" {
		t.Errorf("expected V1")
	}

	v2, err := svc.GetVersion(1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v2["name"] != "V2" {
		t.Errorf("expected V2")
	}
}

func TestGetVersion_NotFound(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	_, err := svc.GetVersion(1, 1)
	if err != service.ErrRecordDoesNotExist {
		t.Errorf("expected ErrRecordDoesNotExist")
	}
}

func TestListVersions(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "V1"})
	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "V2"})
	_, _ = svc.CreateOrUpdate(1, map[string]string{"name": "V3"})

	versions, err := svc.ListVersions(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}

	if versions[0] != 1 || versions[2] != 3 {
		t.Errorf("unexpected version list: %v", versions)
	}
}

func TestTimestampsAreSet(t *testing.T) {
	path, cleanup := createRecordTestDB(t)
	defer cleanup()

	svc, _ := service.NewSQLiteRecordService(path)

	record, err := svc.CreateOrUpdate(1, map[string]string{"name": "TimeTest"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.CreatedAt.IsZero() || record.UpdatedAt.IsZero() {
		t.Errorf("timestamps should be set")
	}

	// sanity check: within reasonable range
	if time.Since(record.CreatedAt) > time.Minute {
		t.Errorf("created_at too old")
	}
}
