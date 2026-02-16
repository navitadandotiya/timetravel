package gateways_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"os"
	"github.com/rainbowmga/timetravel/gateways"
	_ "github.com/mattn/go-sqlite3"
)

// setupInMemoryDB creates an in-memory SQLite DB with minimal tables.
func setupInMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db := gateways.ConnectDB(":memory:", "") // empty path skips SQL file

	schema := `
	CREATE TABLE IF NOT EXISTS feature_flags (
		flag_key TEXT PRIMARY KEY,
		enabled BOOLEAN NOT NULL,
		description TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		rollout_percentage INTEGER DEFAULT 100
	);
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY
	);
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create in-memory schema: %v", err)
	}

	return db
}

// ---------------- TESTS -----------------

func TestConnectDB_InMemory(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	// Verify table exists
	var count int
	err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='feature_flags';").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected table 'feature_flags' to exist")
	}
}

func TestRunMigrations(t *testing.T) {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    // schema_migrations table must exist first
    _, err = db.Exec("CREATE TABLE schema_migrations(version TEXT PRIMARY KEY);")
    if err != nil {
        t.Fatal(err)
    }

    // Create migration file(s) with "IF NOT EXISTS"
    migSQL := "CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY);"
    migFile := filepath.Join(t.TempDir(), "001_create_test.sql")
    os.WriteFile(migFile, []byte(migSQL), 0644)

    err = gateways.RunMigrations(db, filepath.Dir(migFile))
    if err != nil {
        t.Fatalf("RunMigrations failed: %v", err)
    }

    // Optionally rerun to test idempotency
    err = gateways.RunMigrations(db, filepath.Dir(migFile))
    if err != nil {
        t.Fatalf("RunMigrations failed on second run: %v", err)
    }
}

