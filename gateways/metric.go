package gateways

import (
	"database/sql"
)

type MetricsRepository struct {
	DB *sql.DB
}

func NewMetricsRepository(dbPath string) (*MetricsRepository,error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enforce foreign keys
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	return &MetricsRepository{DB: db},nil
}

func (r *MetricsRepository) InsertMetric(metricType, metricName string, value float64, region string) error {
	_, err := r.DB.Exec(`
		INSERT INTO observability_metrics (
			metric_type,
			metric_name,
			value,
			region,
			recorded_at
		) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, metricType, metricName, value, region)

	return err
}
