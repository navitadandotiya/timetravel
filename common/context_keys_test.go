package common

import (
	"context"
	"testing"
)

func TestGetUserID_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected int64
		ok       bool
	}{
		{
			name:     "valid int64",
			ctx:      context.WithValue(context.Background(), UserIDKey, int64(10)),
			expected: 10,
			ok:       true,
		},
		{
			name:     "wrong type",
			ctx:      context.WithValue(context.Background(), UserIDKey, "bad"),
			expected: 0,
			ok:       false,
		},
		{
			name:     "not present",
			ctx:      context.Background(),
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := GetUserID(tt.ctx)

			if ok != tt.ok {
				t.Fatalf("expected ok %v, got %v", tt.ok, ok)
			}

			if id != tt.expected {
				t.Fatalf("expected id %d, got %d", tt.expected, id)
			}
		})
	}
}
