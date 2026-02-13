package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rainbowmga/timetravel/entity"
)

var (
	ErrRecordDoesNotExist = errors.New("record does not exist")
)

// SQLiteRecordService implements v2 persistent storage
type SQLiteRecordService struct {
	db *sql.DB
}

// NewSQLiteRecordService initializes the service with DB connection
func NewSQLiteRecordService(dbPath string) (*SQLiteRecordService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteRecordService{db: db}, nil
}

// CreateOrUpdate inserts or updates a policyholder record and writes audit + event log
func (s *SQLiteRecordService) CreateOrUpdate(policyholderID int64, data map[string]string) (*entity.PolicyholderRecord, error) {
	// Marshal data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Check if record exists
	var recordID int64
	row := s.db.QueryRow("SELECT record_id FROM policyholder_records WHERE policyholder_id = ?", policyholderID)
	err = row.Scan(&recordID)

	now := time.Now().UTC()

	if err == sql.ErrNoRows {
		// Insert new record
		res, err := s.db.Exec(`
			INSERT INTO policyholder_records (policyholder_id, data, created_at, updated_at)
			VALUES (?, ?, ?, ?)`,
			policyholderID, string(dataJSON), now, now)
		if err != nil {
			return nil, err
		}
		recordID, _ = res.LastInsertId()

		// Write audit history
		_, _ = s.db.Exec(`
			INSERT INTO audit_history (record_id, data, changed_at, event_type)
			VALUES (?, ?, ?, ?)`,
			recordID, string(dataJSON), now, "create")

		// Write event log
		_, _ = s.db.Exec(`
			INSERT INTO event_log (record_id, action, timestamp, details)
			VALUES (?, ?, ?, ?)`,
			recordID, "create", now, string(dataJSON))

	} else if err == nil {
		// Update existing record
		_, err = s.db.Exec(`
			UPDATE policyholder_records
			SET data = ?, updated_at = ?
			WHERE record_id = ?`,
			string(dataJSON), now, recordID)
		if err != nil {
			return nil, err
		}

		// Write audit history
		_, _ = s.db.Exec(`
			INSERT INTO audit_history (record_id, data, changed_at, event_type)
			VALUES (?, ?, ?, ?)`,
			recordID, string(dataJSON), now, "update")

		// Write event log
		_, _ = s.db.Exec(`
			INSERT INTO event_log (record_id, action, timestamp, details)
			VALUES (?, ?, ?, ?)`,
			recordID, "update", now, string(dataJSON))
	} else {
		return nil, err
	}

	return &entity.PolicyholderRecord{
		ID:        recordID,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get retrieves a record by policyholder ID
func (s *SQLiteRecordService) Get(policyholderID int64) (*entity.PolicyholderRecord, error) {
	row := s.db.QueryRow("SELECT record_id, data, created_at, updated_at FROM policyholder_records WHERE policyholder_id = ?", policyholderID)
	var recordID int64
	var dataJSON string
	var createdAt, updatedAt string

	err := row.Scan(&recordID, &dataJSON, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrRecordDoesNotExist
	} else if err != nil {
		return nil, err
	}

	var data map[string]string
	_ = json.Unmarshal([]byte(dataJSON), &data)

	createdTime, _ := time.Parse(time.RFC3339, createdAt)
	updatedTime, _ := time.Parse(time.RFC3339, updatedAt)

	return &entity.PolicyholderRecord{
		ID:        recordID,
		Data:      data,
		CreatedAt: createdTime,
		UpdatedAt: updatedTime,
	}, nil
}
