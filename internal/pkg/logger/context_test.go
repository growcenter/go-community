package logger

import (
"context"
"testing"
)

func TestAddProcess(t *testing.T) {
	ctx := context.Background()
	event := NewWideEvent("req-123", "GET", "/", "127.0.0.1", "curl")
	ctx = WithWideEvent(ctx, event)

	AddProcess(ctx, "db_operation", "event.create")
	AddProcess(ctx, "db_operation", "event.update")

	data := event.GetBusinessData()
	
	ops, ok := data["db_operation"]
	if !ok {
		t.Fatalf("db_operation not found in business data")
	}

	opsSlice, ok := ops.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", ops)
	}

	if len(opsSlice) != 2 {
		t.Fatalf("expected length 2, got %d", len(opsSlice))
	}

	if opsSlice[0] != "event.create" || opsSlice[1] != "event.update" {
		t.Fatalf("unexpected slice values: %v", opsSlice)
	}
}
