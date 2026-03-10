package core

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDeferUntilFieldSerializationRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deferUntil *time.Time
		wantKey    bool // whether defer_until key should appear in YAML
	}{
		{
			name:       "nil DeferUntil omits key from YAML",
			deferUntil: nil,
			wantKey:    false,
		},
		{
			name:       "set DeferUntil persists through round-trip",
			deferUntil: timePtr(time.Date(2026, 3, 8, 14, 0, 0, 0, time.UTC)),
			wantKey:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := NewTask("test task")
			original.DeferUntil = tt.deferUntil

			data, err := yaml.Marshal(original)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			// Check YAML key presence
			yamlStr := string(data)
			hasKey := hasDeferSubstring(yamlStr, "defer_until")
			if hasKey != tt.wantKey {
				t.Errorf("YAML contains defer_until=%v, want %v\nYAML:\n%s", hasKey, tt.wantKey, yamlStr)
			}

			var loaded Task
			if err := yaml.Unmarshal(data, &loaded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if tt.deferUntil == nil {
				if loaded.DeferUntil != nil {
					t.Errorf("expected nil DeferUntil, got %v", loaded.DeferUntil)
				}
			} else {
				if loaded.DeferUntil == nil {
					t.Fatal("expected non-nil DeferUntil, got nil")
					return
				}
				if !loaded.DeferUntil.Equal(*tt.deferUntil) {
					t.Errorf("DeferUntil = %v, want %v", loaded.DeferUntil, tt.deferUntil)
				}
			}
		})
	}
}

func TestNewStatusTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		from   TaskStatus
		to     TaskStatus
		wantOK bool
	}{
		{"in-progress to deferred", StatusInProgress, StatusDeferred, true},
		{"blocked to deferred", StatusBlocked, StatusDeferred, true},
		{"todo to deferred (existing)", StatusTodo, StatusDeferred, true},
		{"deferred to todo (existing)", StatusDeferred, StatusTodo, true},
		{"complete to deferred (invalid)", StatusComplete, StatusDeferred, false},
		{"archived to deferred (invalid)", StatusArchived, StatusDeferred, false},
		{"in-review to deferred (invalid)", StatusInReview, StatusDeferred, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsValidTransition(tt.from, tt.to)
			if got != tt.wantOK {
				t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.wantOK)
			}
		})
	}
}

func TestStatusTransitionInProgressToDeferred(t *testing.T) {
	t.Parallel()

	task := NewTask("test")
	if err := task.UpdateStatus(StatusInProgress); err != nil {
		t.Fatalf("transition to in-progress: %v", err)
	}
	if err := task.UpdateStatus(StatusDeferred); err != nil {
		t.Errorf("transition in-progress→deferred should succeed: %v", err)
	}
	if task.Status != StatusDeferred {
		t.Errorf("status = %q, want %q", task.Status, StatusDeferred)
	}
}

func TestStatusTransitionBlockedToDeferred(t *testing.T) {
	t.Parallel()

	task := NewTask("test")
	if err := task.UpdateStatus(StatusBlocked); err != nil {
		t.Fatalf("transition to blocked: %v", err)
	}
	if err := task.UpdateStatus(StatusDeferred); err != nil {
		t.Errorf("transition blocked→deferred should succeed: %v", err)
	}
	if task.Status != StatusDeferred {
		t.Errorf("status = %q, want %q", task.Status, StatusDeferred)
	}
}

func TestCheckDeferredReturns_ExpiredTaskReturnsToTodo(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("deferred task")
	task.Status = StatusDeferred
	past := time.Now().UTC().Add(-1 * time.Hour)
	task.DeferUntil = &past
	pool.AddTask(task)

	returned := CheckDeferredReturns(pool)

	if returned != 1 {
		t.Errorf("returned = %d, want 1", returned)
	}
	if task.Status != StatusTodo {
		t.Errorf("status = %q, want %q", task.Status, StatusTodo)
	}
	if task.DeferUntil != nil {
		t.Errorf("DeferUntil should be nil after return, got %v", task.DeferUntil)
	}
}

func TestCheckDeferredReturns_FutureTaskStaysDeferred(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("future deferred")
	task.Status = StatusDeferred
	future := time.Now().UTC().Add(24 * time.Hour)
	task.DeferUntil = &future
	pool.AddTask(task)

	returned := CheckDeferredReturns(pool)

	if returned != 0 {
		t.Errorf("returned = %d, want 0", returned)
	}
	if task.Status != StatusDeferred {
		t.Errorf("status = %q, want %q", task.Status, StatusDeferred)
	}
	if task.DeferUntil == nil {
		t.Error("DeferUntil should remain set for future task")
	}
}

func TestCheckDeferredReturns_SomedayTaskDoesNotReturn(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("someday task")
	task.Status = StatusDeferred
	task.DeferUntil = nil // Someday snooze
	pool.AddTask(task)

	returned := CheckDeferredReturns(pool)

	if returned != 0 {
		t.Errorf("returned = %d, want 0", returned)
	}
	if task.Status != StatusDeferred {
		t.Errorf("status = %q, want %q", task.Status, StatusDeferred)
	}
}

func TestCheckDeferredReturns_BoundaryExactlyNow(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("boundary task")
	task.Status = StatusDeferred
	now := time.Now().UTC()
	task.DeferUntil = &now
	pool.AddTask(task)

	returned := CheckDeferredReturns(pool)

	// DeferUntil equal to now should trigger return (not strictly after)
	if returned != 1 {
		t.Errorf("returned = %d, want 1 (boundary: DeferUntil == now)", returned)
	}
	if task.Status != StatusTodo {
		t.Errorf("status = %q, want %q", task.Status, StatusTodo)
	}
}

func TestCheckDeferredReturns_MultipleTasksMixed(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	past := time.Now().UTC().Add(-2 * time.Hour)
	future := time.Now().UTC().Add(2 * time.Hour)

	expired1 := NewTask("expired 1")
	expired1.Status = StatusDeferred
	expired1.DeferUntil = &past
	pool.AddTask(expired1)

	expired2 := NewTask("expired 2")
	expired2.Status = StatusDeferred
	pastCopy := past
	expired2.DeferUntil = &pastCopy
	pool.AddTask(expired2)

	futureTask := NewTask("future")
	futureTask.Status = StatusDeferred
	futureTask.DeferUntil = &future
	pool.AddTask(futureTask)

	somedayTask := NewTask("someday")
	somedayTask.Status = StatusDeferred
	somedayTask.DeferUntil = nil
	pool.AddTask(somedayTask)

	returned := CheckDeferredReturns(pool)

	if returned != 2 {
		t.Errorf("returned = %d, want 2", returned)
	}
	if expired1.Status != StatusTodo {
		t.Errorf("expired1 status = %q, want todo", expired1.Status)
	}
	if expired2.Status != StatusTodo {
		t.Errorf("expired2 status = %q, want todo", expired2.Status)
	}
	if futureTask.Status != StatusDeferred {
		t.Errorf("futureTask status = %q, want deferred", futureTask.Status)
	}
	if somedayTask.Status != StatusDeferred {
		t.Errorf("somedayTask status = %q, want deferred", somedayTask.Status)
	}
}

func TestCheckDeferredReturns_EmptyPool(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	returned := CheckDeferredReturns(pool)
	if returned != 0 {
		t.Errorf("returned = %d, want 0 for empty pool", returned)
	}
}

func TestGetAvailableForDoors_ExcludesDeferredTasks(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()

	todo := NewTask("todo task")
	pool.AddTask(todo)

	deferredWithDate := NewTask("deferred with date")
	deferredWithDate.Status = StatusDeferred
	future := time.Now().UTC().Add(24 * time.Hour)
	deferredWithDate.DeferUntil = &future
	pool.AddTask(deferredWithDate)

	deferredSomeday := NewTask("deferred someday")
	deferredSomeday.Status = StatusDeferred
	pool.AddTask(deferredSomeday)

	available := pool.GetAvailableForDoors()

	for _, t2 := range available {
		if t2.Status == StatusDeferred {
			t.Errorf("deferred task %q should not be in available doors", t2.Text)
		}
	}
	if len(available) != 1 {
		t.Errorf("expected 1 available task, got %d", len(available))
	}
}

func TestReturnedTaskEligibleForDoors(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("will return")
	task.Status = StatusDeferred
	past := time.Now().UTC().Add(-1 * time.Hour)
	task.DeferUntil = &past
	pool.AddTask(task)

	// Before return, task should NOT be in doors
	available := pool.GetAvailableForDoors()
	for _, a := range available {
		if a.ID == task.ID {
			t.Fatal("deferred task should not be in doors before return")
		}
	}

	// Run auto-return
	CheckDeferredReturns(pool)

	// After return, task SHOULD be in doors
	available = pool.GetAvailableForDoors()
	found := false
	for _, a := range available {
		if a.ID == task.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("returned task should be eligible for doors after auto-return")
	}
}

// timePtr returns a pointer to the given time.Time value.
func timePtr(t time.Time) *time.Time {
	return &t
}

// hasDeferSubstring checks if substr is present in s.
func hasDeferSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
