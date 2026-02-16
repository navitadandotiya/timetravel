package gateways_test

import (
	"database/sql"
	"testing"

	"github.com/rainbowmga/timetravel/gateways"
	_ "github.com/mattn/go-sqlite3"
)

// helper to create observability_metrics table
func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// create table
	_, err = db.Exec(`
	CREATE TABLE observability_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		metric_type TEXT,
		metric_name TEXT,
		value REAL,
		region TEXT,
		recorded_at DATETIME
	);
	`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestNewMetricsRepository(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := &gateways.MetricsRepository{DB: db}
	if repo.DB == nil {
		t.Fatalf("expected DB to be non-nil")
	}
}

func TestInsertMetric(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	repo := &gateways.MetricsRepository{DB: db}

	tests := []struct {
		name       string
		metricType string
		metricName string
		value      float64
		region     string
		wantErr    bool
	}{
		{"normal insert", "gauge", "cpu_usage", 42.5, "us-east-1", false},
		{"empty region", "counter", "requests", 10, "", false},
		{"empty metric type", "", "mem_usage", 99.9, "us-west-2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.InsertMetric(tt.metricType, tt.metricName, tt.value, tt.region)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InsertMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify row was inserted
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM observability_metrics WHERE metric_name = ?", tt.metricName).Scan(&count)
			if err != nil {
				t.Fatalf("failed to query table: %v", err)
			}
			if count != 1 {
				t.Fatalf("expected 1 row inserted for %s, got %d", tt.metricName, count)
			}
		})
	}
}

func TestInsertMetric_Failure(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// drop table to simulate failure
	db.Exec("DROP TABLE observability_metrics")

	repo := &gateways.MetricsRepository{DB: db}
	err := repo.InsertMetric("gauge", "cpu", 1.2, "us-east-1")
	if err == nil {
		t.Fatal("expected error inserting into missing table, got nil")
	}
}
