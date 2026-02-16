package gateways

import (
	"database/sql"
)


// MetricsRepo defines what our observability code expects
type MetricsRepo interface {
	InsertMetric(metricType, metricName string, value float64, region string) error
	SumMetric(metricName string) (float64, error) // optional for DB collector
}

// change the global variable to the interface type
var metricsRepo MetricsRepo

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

func (r *MetricsRepository) SumMetric(metricName string) (float64, error) {
    var sum float64
    err := r.DB.QueryRow(`SELECT SUM(value) FROM metrics WHERE name = ?`, metricName).Scan(&sum)
    if err != nil {
        return 0, err
    }
    return sum, nil
}
