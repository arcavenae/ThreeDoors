package core

import "time"

// CheckDeferredReturns iterates deferred tasks in the pool and returns any
// whose DeferUntil has passed back to todo status. Tasks with nil DeferUntil
// ("Someday" snoozes) are left untouched. Returns the count of tasks returned.
func CheckDeferredReturns(pool *TaskPool) int {
	return CheckDeferredReturnsWithTracker(pool, nil)
}

// CheckDeferredReturnsWithTracker iterates deferred tasks and returns expired
// ones to todo status. If a SessionTracker is provided, each returned task is
// logged as a snooze_return event.
func CheckDeferredReturnsWithTracker(pool *TaskPool, tracker *SessionTracker) int {
	now := time.Now().UTC()
	returned := 0
	for _, t := range pool.GetTasksByStatus(StatusDeferred) {
		if t.DeferUntil != nil && !t.DeferUntil.After(now) {
			if err := t.UpdateStatus(StatusTodo); err == nil {
				taskID := t.ID
				t.DeferUntil = nil
				returned++
				if tracker != nil {
					tracker.RecordSnoozeReturn(taskID)
				}
			}
		}
	}
	return returned
}
