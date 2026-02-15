package entity

import "time"

// ------------------------------
// POLICYHOLDER
// ------------------------------
type Policyholder struct {
	ID        int64     `db:"policyholder_id" json:"policyholder_id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Country   string    `db:"country_code" json:"country_code"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ------------------------------
// POLICYHOLDER RECORD (LATEST STATE)
// ------------------------------
type PolicyholderRecord struct {
	ID        int64             `db:"record_id" json:"id"`
	Data      map[string]string `db:"-" json:"data"` // stored as JSON in DB
	Version   int               `db:"version" json:"version"`
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`
}

// ------------------------------
// AUDIT HISTORY (IMMUTABLE SNAPSHOTS)
// ------------------------------
type AuditHistory struct {
	ID        int64             `db:"audit_id" json:"audit_id"`
	RecordID  int64             `db:"record_id" json:"record_id"`
	Version   int               `db:"version" json:"version"` // version snapshot
	Data      map[string]string `db:"-" json:"data"`
	ChangedAt time.Time         `db:"changed_at" json:"changed_at"`
	EventType string            `db:"event_type" json:"event_type"` // create/update/delete
}

// ------------------------------
// EVENT LOG (TRACEABILITY)
// ------------------------------
type EventLog struct {
	ID        int64     `db:"event_id" json:"event_id"`
	RecordID  int64     `db:"record_id" json:"record_id"`
	Action    string    `db:"action" json:"action"`   // create/update/delete/read
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	Details   string    `db:"details" json:"details"` // JSON string or human-readable message
}

// ------------------------------
// OBSERVABILITY METRICS
// ------------------------------
type ObservabilityMetric struct {
	ID         int64     `db:"metric_id" json:"metric_id"`
	MetricType string    `db:"metric_type" json:"metric_type"` // "platform" or "business"
	MetricName string    `db:"metric_name" json:"metric_name"`
	Value      float64   `db:"value" json:"value"`
	Region     string    `db:"region" json:"region"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
}

// ------------------------------
// FEATURE FLAGS (GLOBAL CONTROL)
// ------------------------------
type FeatureFlag struct {
	Key         string    `db:"flag_key" json:"flag_key"`
	Enabled     bool      `db:"enabled" json:"enabled"`
	Description string    `db:"description" json:"description"`
	RolloutPercentage int `db:"description" json:"rollout_percentage"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ------------------------------
// API VERSION CONFIG (ROLL-OUT CONTROL)
// ------------------------------
type APIVersionConfig struct {
	Version           string    `db:"version" json:"version"`
	IsActive          bool      `db:"is_active" json:"is_active"`
	RolloutPercentage int       `db:"rollout_percentage" json:"rollout_percentage"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}