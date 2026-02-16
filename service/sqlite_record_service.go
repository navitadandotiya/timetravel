package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rainbowmga/timetravel/entity"
)

var (
	ErrRecordDoesNotExist = errors.New("record does not exist")
)

// SQLiteRecordService implements v2 persistent storage with versioning
type SQLiteRecordService struct {
	db *sql.DB
}

// define interface for the controller to allow testing with mocks
type SQLiteRecordServiceInterface interface {
	Get(int64) (*entity.PolicyholderRecord, error)
	CreateOrUpdate(int64, map[string]string) (*entity.PolicyholderRecord, error)
	GetVersion(int64, int) (map[string]string, error)
	ListVersions(int64) ([]int, error)
}

// NewSQLiteRecordService initializes the service with DB connection
func NewSQLiteRecordService(dbPath string) (*SQLiteRecordService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enforce foreign keys
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")
	return &SQLiteRecordService{db: db}, nil
}

// CreateOrUpdate inserts or updates a policyholder record, increments version, writes audit + event log
func (s *SQLiteRecordService) CreateOrUpdate(policyholderID int64, data map[string]string) (*entity.PolicyholderRecord, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	dataJSON, _ := json.Marshal(data)
	now := time.Now().UTC()

	// --- Step 0: Ensure policyholder exists ---
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO policyholders (policyholder_id, name)
		VALUES (?, ?)`, policyholderID, data["name"])
	if err != nil {
		return nil, fmt.Errorf("failed to ensure policyholder exists: %w", err)
	}

	// --- Step 1: Check if record exists ---
	var recordID int64
	var currentVersion int
	row := tx.QueryRow(`
		SELECT record_id, version
		FROM policyholder_records
		WHERE policyholder_id = ?`, policyholderID)
	err = row.Scan(&recordID, &currentVersion)

	if err == sql.ErrNoRows {
		// Insert new record
		res, err := tx.Exec(`
			INSERT INTO policyholder_records 
			(policyholder_id, data, version, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`,
			policyholderID, string(dataJSON), 1, now, now,
		)
		if err != nil {
			return nil, err
		}

		recordID, _ = res.LastInsertId()
		currentVersion = 1

		// Insert audit history
		_, err = tx.Exec(`
			INSERT INTO audit_history 
			(record_id, version, data, changed_at, event_type)
			VALUES (?, ?, ?, ?, ?)`,
			recordID, currentVersion, string(dataJSON), now, "create",
		)
		if err != nil {
			return nil, err
		}

		// Insert event log
		_, err = tx.Exec(`
			INSERT INTO event_logs 
			(record_id, action, timestamp, details)
			VALUES (?, ?, ?, ?)`,
			recordID, "create", now, string(dataJSON),
		)
		if err != nil {
			return nil, err
		}

	} else if err == nil {
		// Update existing record
		currentVersion++
		_, err = tx.Exec(`
			UPDATE policyholder_records
			SET data = ?, version = ?, updated_at = ?
			WHERE record_id = ?`,
			string(dataJSON), currentVersion, now, recordID,
		)
		if err != nil {
			return nil, err
		}

		// Insert audit history
		_, err = tx.Exec(`
			INSERT INTO audit_history 
			(record_id, version, data, changed_at, event_type)
			VALUES (?, ?, ?, ?, ?)`,
			recordID, currentVersion, string(dataJSON), now, "update",
		)
		if err != nil {
			return nil, err
		}

		// Insert event log
		_, err = tx.Exec(`
			INSERT INTO event_logs 
			(record_id, action, timestamp, details)
			VALUES (?, ?, ?, ?)`,
			recordID, "update", now, string(dataJSON),
		)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &entity.PolicyholderRecord{
		ID:        recordID,
		Data:      data,
		Version:   currentVersion,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Get retrieves a record by policyholder ID
func (s *SQLiteRecordService) Get(policyholderID int64) (*entity.PolicyholderRecord, error) {
	row := s.db.QueryRow(`
		SELECT record_id, data, version, created_at, updated_at
		FROM policyholder_records
		WHERE policyholder_id = ?`, policyholderID)

	var recordID int64
	var dataJSON string
	var version int
	var createdAt, updatedAt string

	err := row.Scan(&recordID, &dataJSON, &version, &createdAt, &updatedAt)
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
		Version:   version,
		CreatedAt: createdTime,
		UpdatedAt: updatedTime,
	}, nil
}

// GetVersion returns a specific historical version
func (s *SQLiteRecordService) GetVersion(policyholderID int64, version int) (map[string]string, error) {
	row := s.db.QueryRow(`
		SELECT ah.data
		FROM audit_history ah
		JOIN policyholder_records pr ON pr.record_id = ah.record_id
		WHERE pr.policyholder_id = ?
		AND ah.version = ?`,
		policyholderID, version,
	)

	var jsonData string
	if err := row.Scan(&jsonData); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordDoesNotExist
		}
		return nil, err
	}

	var data map[string]string
	_ = json.Unmarshal([]byte(jsonData), &data)

	return data, nil
}

// ListVersions returns all versions for a record
func (s *SQLiteRecordService) ListVersions(policyholderID int64) ([]int, error) {
	rows, err := s.db.Query(`
		SELECT ah.version
		FROM audit_history ah
		JOIN policyholder_records pr ON pr.record_id = ah.record_id
		WHERE pr.policyholder_id = ?
		ORDER BY ah.version`,
		policyholderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		rows.Scan(&v)
		versions = append(versions, v)
	}

	return versions, nil
}
