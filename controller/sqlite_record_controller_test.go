package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/observability"
)

// --- Mock Logger ---

type mockLogger struct {
	Infos  []string
	Errors []string
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  { m.Infos = append(m.Infos, msg) }
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) { m.Errors = append(m.Errors, msg) }

// --- Mock SQLite Service ---

type mockSQLiteService struct {
	records map[int64]*entity.PolicyholderRecord
	getErr  error
	updErr  error
}

func (m *mockSQLiteService) Get(id int64) (*entity.PolicyholderRecord, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	rec, ok := m.records[id]
	if !ok {
		return nil, controller.ErrRecordDoesNotExist
	}
	return rec, nil
}

func (m *mockSQLiteService) CreateOrUpdate(id int64, data map[string]string) (*entity.PolicyholderRecord, error) {
	if m.records == nil {
		m.records = make(map[int64]*entity.PolicyholderRecord)
	}
	if m.updErr != nil {
		return nil, m.updErr
	}
	rec := &entity.PolicyholderRecord{
		ID:        id,
		Data:      make(map[string]string),
		UpdatedAt: time.Now().UTC(),
	}
	for k, v := range data {
		rec.Data[k] = v
	}
	m.records[id] = rec
	return rec, nil
}

func (m *mockSQLiteService) GetVersion(id int64, v int) (map[string]string, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return map[string]string{"version": "data"}, nil
}

func (m *mockSQLiteService) ListVersions(id int64) ([]int, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return []int{1, 2}, nil
}

// --- Test Helpers ---

func newControllerWithMocks() (*controller.SQLiteRecordController, *mockSQLiteService, *mockLogger) {
	mockSvc := &mockSQLiteService{
		records: make(map[int64]*entity.PolicyholderRecord),
	}
	mockLog := &mockLogger{}
	origLogger := observability.DefaultLogger
	observability.DefaultLogger = mockLog
	cleanup := func() { observability.DefaultLogger = origLogger }
	t := &testing.T{}
	t.Cleanup(cleanup)

	ctrl := controller.NewSQLiteRecordControllerForTest(mockSvc)
	return ctrl, mockSvc, mockLog
}

// --- Tests ---

func TestSQLiteRecordController_GetRecord(t *testing.T) {
	ctrl, mockSvc, _ := newControllerWithMocks()

	mockSvc.records[1] = &entity.PolicyholderRecord{ID: 1, Data: map[string]string{"foo": "bar"}}

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"valid id", 1, false},
		{"nonexistent id", 2, true},
		{"invalid id", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctrl.GetRecord(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got.ID != tt.id {
				t.Errorf("GetRecord() ID = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestSQLiteRecordController_UpsertRecord(t *testing.T) {
	ctrl, _, mockLog := newControllerWithMocks()

	tests := []struct {
		name    string
		id      int64
		data    map[string]string
		wantErr bool
	}{
		{"success create", 1, map[string]string{"key": "val"}, false},
		{"invalid id", 0, map[string]string{"key": "val"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctrl.UpsertRecord(context.Background(), tt.id, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpsertRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got.Data["key"] != "val" {
				t.Errorf("UpsertRecord() data = %v, want %v", got.Data["key"], "val")
			}
		})
	}

	if len(mockLog.Errors) != 0 {
		t.Errorf("unexpected log errors: %v", mockLog.Errors)
	}
}

func TestSQLiteRecordController_UpdateRecord(t *testing.T) {
	ctrl, mockSvc, _ := newControllerWithMocks()

	mockSvc.records[1] = &entity.PolicyholderRecord{ID: 1, Data: map[string]string{"a": "b"}}

	tests := []struct {
		name    string
		id      int
		updates map[string]*string
		want    map[string]string
		wantErr bool
	}{
		{"update value", 1, map[string]*string{"a": strPtr("c")}, map[string]string{"a": "c"}, false},
		{"delete value", 1, map[string]*string{"a": nil}, map[string]string{}, false},
		{"nonexistent id", 2, map[string]*string{"x": strPtr("y")}, nil, true},
		{"invalid id", 0, map[string]*string{"x": strPtr("y")}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ctrl.UpdateRecord(context.Background(), tt.id, tt.updates)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				for k, v := range tt.want {
					if got.Data[k] != v {
						t.Errorf("UpdateRecord() key %v = %v, want %v", k, got.Data[k], v)
					}
				}
			}
		})
	}
}

func TestSQLiteRecordController_Versions(t *testing.T) {
	ctrl, _, _ := newControllerWithMocks()

	t.Run("GetVersion valid", func(t *testing.T) {
		got, err := ctrl.GetVersion(context.Background(), 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got["version"] != "data" {
			t.Errorf("GetVersion() = %v", got)
		}
	})

	t.Run("ListVersions valid", func(t *testing.T) {
		got, err := ctrl.ListVersions(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Errorf("ListVersions() len = %v, want 2", len(got))
		}
	})
}
