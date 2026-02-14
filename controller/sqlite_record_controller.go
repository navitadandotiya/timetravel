package controller

import (
	"context"
	"time"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)

type SQLiteRecordController struct {
	service *service.SQLiteRecordService
}

//
// Constructor
//

func NewSQLiteRecordController(dbPath string) (*SQLiteRecordController, error) {
	svc, err := service.NewSQLiteRecordService(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLiteRecordController{service: svc}, nil
}

//
// GET RECORD
//
func (c *SQLiteRecordController) GetRecord(ctx context.Context, id int64) (entity.PolicyholderRecord, error) {
	if id <= 0 {
		return entity.PolicyholderRecord{}, ErrRecordIDInvalid
	}

	rec, err := c.service.Get(id)
	if err != nil {
		if err == service.ErrRecordDoesNotExist {
			return entity.PolicyholderRecord{}, ErrRecordDoesNotExist
		}
		return entity.PolicyholderRecord{}, err
	}

	return *rec, nil
}


//
// UPSERT (CREATE OR UPDATE)
// used by handler to keep logic simple & generic
//

func (c *SQLiteRecordController) UpsertRecord(
	ctx context.Context,
	policyholderID int64,
	data map[string]string,
) (entity.PolicyholderRecord, error) {

	if policyholderID <= 0 {
		return entity.PolicyholderRecord{}, ErrRecordIDInvalid
	}

	rec, err := c.service.CreateOrUpdate(policyholderID, data)
	if err != nil {
		return entity.PolicyholderRecord{}, err
	}

	return *rec, nil
}

//
// UPDATE RECORD (PATCH-style)
// supports deletion via nil values
//

func (c *SQLiteRecordController) UpdateRecord(
	ctx context.Context,
	id int,
	updates map[string]*string,
) (entity.PolicyholderRecord, error) {

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

	// apply updates
	for k, v := range updates {
		if v == nil {
			delete(rec.Data, k)
		} else {
			rec.Data[k] = *v
		}
	}

	rec.UpdatedAt = time.Now().UTC()

	updated, err := c.service.CreateOrUpdate(int64(id), rec.Data)
	if err != nil {
		return entity.PolicyholderRecord{}, err
	}

	return *updated, nil
}
