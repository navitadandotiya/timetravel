package entity

import "time"

// Example struct
type Policyholder struct {
    ID        int64     `db:"policyholder_id" json:"policyholder_id"`
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    Country   string    `db:"country_code" json:"country_code"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

// Policyholder record
type PolicyholderRecord struct {
	ID        int64             `db:"record_id" json:"id"`
	Data      map[string]string `db:"-" json:"data"` // stored as JSON in DB
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`
}

// AuditHistory keeps track of changes
type AuditHistory struct {
	ID        int64             `db:"audit_id" json:"audit_id"`
	RecordID  int64             `db:"record_id" json:"record_id"`
	Data      map[string]string `db:"-" json:"data"`
	ChangedAt time.Time         `db:"changed_at" json:"changed_at"`
	EventType string            `db:"event_type" json:"event_type"` // create/update/delete
}