package service_test

import (
	"database/sql"
	"os"
	"sync"
	"testing"

	"github.com/rainbowmga/timetravel/service"
	_ "github.com/mattn/go-sqlite3"
)

// createTempDB creates a temporary sqlite file for testing
func createTempDB(t *testing.T) (string, func()) {
	t.Helper()

	file, err := os.CreateTemp("", "featureflag-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db file: %v", err)
	}
	path := file.Name()
	file.Close()

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("failed to open temp db: %v", err)
	}

	_, err = db.Exec(`
	CREATE TABLE feature_flags (
		flag_key TEXT PRIMARY KEY,
		enabled BOOLEAN,
		rollout_percentage INTEGER
	);
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	db.Close()

	cleanup := func() {
		os.Remove(path)
	}

	return path, cleanup
}

func TestNewFeatureFlagService(t *testing.T) {
	path, cleanup := createTempDB(t)
	defer cleanup()

	db, _ := sql.Open("sqlite3", path)
	_, err := db.Exec(`
	INSERT INTO feature_flags (flag_key, enabled, rollout_percentage)
	VALUES ('enable_v2_api', 1, 100)
	`)
	if err != nil {
		t.Fatalf("failed to insert flag: %v", err)
	}
	db.Close()

	s, err := service.NewFeatureFlagService(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.IsEnabled("enable_v2_api", 1) {
		t.Errorf("expected flag to be enabled")
	}
}

func TestIsEnabled_TableDriven(t *testing.T) {
	path, cleanup := createTempDB(t)
	defer cleanup()

	db, _ := sql.Open("sqlite3", path)
	_, err := db.Exec(`
	INSERT INTO feature_flags (flag_key, enabled, rollout_percentage) VALUES
	('enabled_100', 1, 100),
	('disabled_flag', 0, 100),
	('zero_rollout', 1, 0),
	('partial_flag', 1, 50)
	`)
	if err != nil {
		t.Fatalf("failed to insert flags: %v", err)
	}
	db.Close()

	s, err := service.NewFeatureFlagService(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		flag     string
		userID   int64
		expected bool
	}{
		{"flag_not_exist", "missing", 1, false},
		{"disabled_flag", "disabled_flag", 1, false},
		{"100_percent", "enabled_100", 1, true},
		{"zero_percent", "zero_rollout", 1, false},
		{"partial_deterministic", "partial_flag", 42, s.IsEnabled("partial_flag", 42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.IsEnabled(tt.flag, tt.userID)
			if got != tt.expected {
				t.Errorf("IsEnabled(%s) = %v, want %v", tt.flag, got, tt.expected)
			}
		})
	}
}

func TestRefresh_UpdatesCache(t *testing.T) {
	path, cleanup := createTempDB(t)
	defer cleanup()

	db, _ := sql.Open("sqlite3", path)

	_, _ = db.Exec(`
	INSERT INTO feature_flags (flag_key, enabled, rollout_percentage)
	VALUES ('test_flag', 1, 100)
	`)
	db.Close()

	s, err := service.NewFeatureFlagService(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !s.IsEnabled("test_flag", 1) {
		t.Errorf("expected flag enabled before refresh update")
	}

	// update flag to disabled
	db2, _ := sql.Open("sqlite3", path)
	_, _ = db2.Exec(`
	UPDATE feature_flags SET enabled = 0 WHERE flag_key = 'test_flag'
	`)
	db2.Close()

	if err := s.Refresh(); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if s.IsEnabled("test_flag", 1) {
		t.Errorf("expected flag disabled after refresh")
	}
}

func TestIsEnabled_ConcurrentAccess(t *testing.T) {
	path, cleanup := createTempDB(t)
	defer cleanup()

	db, _ := sql.Open("sqlite3", path)
	_, _ = db.Exec(`
	INSERT INTO feature_flags (flag_key, enabled, rollout_percentage)
	VALUES ('enable_v2_api', 1, 100)
	`)
	db.Close()

	s, err := service.NewFeatureFlagService(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			_ = s.IsEnabled("enable_v2_api", userID)
		}(int64(i))
	}
	wg.Wait()
}

func TestNewFeatureFlagService_InvalidDB(t *testing.T) {
	_, err := service.NewFeatureFlagService("/invalid/path/db.sqlite")
	if err == nil {
		t.Errorf("expected error for invalid DB path")
	}
}
