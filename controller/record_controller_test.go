package controller_test

import (
	"context"
	"testing"

	"github.com/rainbowmga/timetravel/controller"
	"github.com/rainbowmga/timetravel/entity"
)

func TestInMemoryRecordService(t *testing.T) {
	ctx := context.Background()
	svc := controller.NewInMemoryRecordService()

	// Sample record for tests
	record := entity.Record{
		ID:   1,
		Data: map[string]string{"a": "1", "b": "2"},
	}

	t.Run("CreateRecord success", func(t *testing.T) {
		if err := svc.CreateRecord(ctx, record); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("CreateRecord invalid ID", func(t *testing.T) {
		rec := entity.Record{ID: 0, Data: map[string]string{}}
		if err := svc.CreateRecord(ctx, rec); err != controller.ErrRecordIDInvalid {
			t.Fatalf("expected ErrRecordIDInvalid, got %v", err)
		}
	})

	t.Run("CreateRecord already exists", func(t *testing.T) {
		if err := svc.CreateRecord(ctx, record); err != controller.ErrRecordAlreadyExists {
			t.Fatalf("expected ErrRecordAlreadyExists, got %v", err)
		}
	})

	t.Run("GetRecord success", func(t *testing.T) {
		got, err := svc.GetRecord(ctx, 1)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if got.ID != record.ID || got.Data["a"] != "1" {
			t.Errorf("unexpected record: %+v", got)
		}
	})

	t.Run("GetRecord missing", func(t *testing.T) {
		_, err := svc.GetRecord(ctx, 99)
		if err != controller.ErrRecordDoesNotExist {
			t.Fatalf("expected ErrRecordDoesNotExist, got %v", err)
		}
	})

	t.Run("UpdateRecord add/update keys", func(t *testing.T) {
		updates := map[string]*string{
			"a": strPtr("updated"),
			"c": strPtr("new"),
		}
		updated, err := svc.UpdateRecord(ctx, 1, updates)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if updated.Data["a"] != "updated" || updated.Data["c"] != "new" {
			t.Errorf("unexpected data after update: %+v", updated.Data)
		}
	})

	t.Run("UpdateRecord delete key", func(t *testing.T) {
		updates := map[string]*string{
			"b": nil, // should delete key
		}
		updated, err := svc.UpdateRecord(ctx, 1, updates)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if _, ok := updated.Data["b"]; ok {
			t.Errorf("key 'b' should have been deleted")
		}
	})

	t.Run("UpdateRecord missing record", func(t *testing.T) {
		updates := map[string]*string{"x": strPtr("y")}
		_, err := svc.UpdateRecord(ctx, 99, updates)
		if err != controller.ErrRecordDoesNotExist {
			t.Fatalf("expected ErrRecordDoesNotExist, got %v", err)
		}
	})
}

// helper to return pointer to string
func strPtr(s string) *string {
	return &s
}
