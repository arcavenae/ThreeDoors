package core

import "testing"

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
	// mockProvider returns empty HealthCheckResult - WAL delegates to inner
	result := wp.HealthCheck()
	if len(result.Items) != 0 {
		t.Errorf("expected 0 items (from mock), got %d", len(result.Items))
	}
}
