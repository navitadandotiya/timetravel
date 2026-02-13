package controller

import (
	"context"
	"time"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)


// SQLiteRecordController wraps the SQLite service to match RecordService-like interface
type SQLiteRecordController struct {
	service *service.SQLiteRecordService
}

// Constructor
func NewSQLiteRecordService(dbPath string) (*SQLiteRecordController, error) {
	svc, err := service.NewSQLiteRecordService(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteRecordController{service: svc}, nil
}

// GetRecord retrieves a record by policyholder ID
func (c *SQLiteRecordController) GetRecord(ctx context.Context, id int) (entity.PolicyholderRecord, error) {
	if id <= 0 {
		return entity.PolicyholderRecord{}, ErrRecordIDInvalid
	}

	rec, err := c.service.Get(int64(id))
	if err != nil {
		if err == service.ErrRecordDoesNotExist {
			return entity.PolicyholderRecord{}, ErrRecordDoesNotExist
		}
		return entity.PolicyholderRecord{}, err
	}

	return *rec, nil
}

// CreateRecord inserts a new record
func (c *SQLiteRecordController) CreateRecord(ctx context.Context, record entity.PolicyholderRecord) error {
	if record.ID <= 0 {
		return ErrRecordIDInvalid
	}

	_, err := c.service.CreateOrUpdate(record.ID, record.Data)
	return err
}

// UpdateRecord updates an existing record's fields; nil values delete keys
func (c *SQLiteRecordController) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.PolicyholderRecord, error) {
	if id <= 0 {
		return entity.PolicyholderRecord{}, ErrRecordIDInvalid
	}

	// Get existing record
	rec, err := c.service.Get(int64(id))
	if err != nil {
		if err == service.ErrRecordDoesNotExist {
			return entity.PolicyholderRecord{}, ErrRecordDoesNotExist
		}
		return entity.PolicyholderRecord{}, err
	}

	// Apply updates
	for k, v := range updates {
		if v == nil {
			delete(rec.Data, k)
		} else {
			rec.Data[k] = *v
		}
	}
	rec.UpdatedAt = time.Now().UTC()

	// Persist back
	_, err = c.service.CreateOrUpdate(int64(id), rec.Data)
	if err != nil {
		return entity.PolicyholderRecord{}, err
	}

	return *rec, nil
}
