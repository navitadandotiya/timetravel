package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/rainbowmga/timetravel/common"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type featureFlagServiceWrapper struct {
	enabledFlags map[string]bool
	refreshErr   error
}

func (m *featureFlagServiceWrapper) IsEnabled(flag string, userID int64) bool {
	return m.enabledFlags[flag]
}

func (m *featureFlagServiceWrapper) Refresh() error {
	return m.refreshErr
}

func TestFeatureFlagController_IsEnabled(t *testing.T) {
	mockSvc := &featureFlagServiceWrapper{
		enabledFlags: map[string]bool{"feature_a": true, "feature_b": false},
	}
	ctrl := NewFeatureFlagControllerWithService(mockSvc)

	mockObs := &mockLogger{}
	restoreLogger := setLoggerForTest(mockObs)
	defer restoreLogger()

	tests := []struct {
		name    string
		flag    string
		ctxOK   bool
		want    bool
		wantLog bool
	}{
		{"enabled flag with context", "feature_a", true, true, false},
		{"disabled flag with context", "feature_b", true, false, false},
		{"missing context", "feature_a", false, false, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if tt.ctxOK {
				ctx = context.WithValue(context.Background(), common.UserIDKey, int64(123))
			} else {
				ctx = context.Background()
			}

			got := ctrl.IsEnabled(ctx, tt.flag)
			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}

			if tt.wantLog && len(mockObs.Errors) == 0 {
				t.Errorf("expected error log for missing context, got none")
			}

			// Reset logs
			mockObs.Errors = nil
		})
	}

	// Refresh tests
	t.Run("Refresh success", func(t *testing.T) {
		mockSvc.refreshErr = nil
		if err := ctrl.Refresh(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("Refresh failure", func(t *testing.T) {
		mockSvc.refreshErr = errors.New("db failed")
		if err := ctrl.Refresh(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}


func TestFeatureFlagController_Integration(t *testing.T) {
	// temporary SQLite file
	dbFile := "test_flags.db"
	os.Remove(dbFile)
	defer os.Remove(dbFile)

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// create table
	_, err = db.Exec(`
	CREATE TABLE feature_flags (
		flag_key TEXT PRIMARY KEY,
		enabled BOOLEAN NOT NULL,
		description TEXT,
		rollout_percentage INTEGER DEFAULT 100,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		t.Fatal(err)
	}

	// insert flags
	_, err = db.Exec(`
	INSERT INTO feature_flags (flag_key, enabled, rollout_percentage) VALUES
	('feature_a', 1, 100),
	('feature_b', 0, 100)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Use a mock logger to capture logs
	mockObs := &mockLogger{}
	restore := setLoggerForTest(mockObs)
	defer restore()

	// initialize controller (internally creates FeatureFlagService)
	ctrl, err := NewFeatureFlagController(dbFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), common.UserIDKey, int64(42))

	// enabled flag
	if !ctrl.IsEnabled(ctx, "feature_a") {
		t.Errorf("feature_a should be enabled")
	}

	// disabled flag
	if ctrl.IsEnabled(ctx, "feature_b") {
		t.Errorf("feature_b should be disabled")
	}

	// missing context triggers error log
	ctrl.IsEnabled(context.Background(), "feature_a")
	if len(mockObs.Errors) == 0 {
		t.Errorf("expected error log for missing context")
	}

	// Refresh reloads from DB
	if err := ctrl.Refresh(); err != nil {
		t.Errorf("Refresh failed: %v", err)
	}
}

