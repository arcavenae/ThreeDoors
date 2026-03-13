package core

import (
	"errors"
	"testing"
	"time"
)

func TestWALProvider_Name(t *testing.T) {
	t.Parallel()
	wp, _, _ := newTestWALProvider(t)
	name := wp.Name()
	if name != "mock (WAL)" {
		t.Errorf("Name() = %q, want %q", name, "mock (WAL)")
	}
}

func TestWALProvider_Watch(t *testing.T) {
	t.Parallel()
	wp, _, _ := newTestWALProvider(t)
	if wp.Watch() != nil {
		t.Error("expected Watch to return nil")
	}
}

func TestWALProvider_HealthCheck(t *testing.T) {
	t.Parallel()
	wp, _, _ := newTestWALProvider(t)
	// mockProvider returns empty HealthCheckResult; WAL appends wal_queue item
	result := wp.HealthCheck()
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item (wal_queue), got %d", len(result.Items))
	}
	if len(result.Items) > 0 && result.Items[0].Name != "wal_queue" {
		t.Errorf("expected wal_queue item, got %q", result.Items[0].Name)
	}
}

func TestWALProvider_OldestPending_Empty(t *testing.T) {
	t.Parallel()
	wp, _, _ := newTestWALProvider(t)
	got := wp.OldestPending()
	if !got.IsZero() {
		t.Errorf("OldestPending() = %v, want zero time for empty queue", got)
	}
}

func TestWALProvider_OldestPending_WithEntries(t *testing.T) {
	t.Parallel()
	wp, mock, _ := newTestWALProvider(t)
	mock.saveErr = errors.New("offline")

	task1 := &Task{ID: "t1", Text: "first"}
	task2 := &Task{ID: "t2", Text: "second"}

	// Queue two tasks — first should be oldest
	beforeEnqueue := time.Now().UTC()
	_ = wp.SaveTask(task1)
	_ = wp.SaveTask(task2)

	oldest := wp.OldestPending()
	if oldest.IsZero() {
		t.Fatal("OldestPending() returned zero time, expected non-zero")
	}
	if oldest.Before(beforeEnqueue.Add(-time.Second)) {
		t.Errorf("OldestPending() = %v, expected >= %v", oldest, beforeEnqueue)
	}
	if wp.PendingCount() != 2 {
		t.Errorf("PendingCount() = %d, want 2", wp.PendingCount())
	}
}
