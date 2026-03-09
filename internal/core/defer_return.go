package core

import "time"

// CheckDeferredReturns iterates deferred tasks in the pool and returns any
// whose DeferUntil has passed back to todo status. Tasks with nil DeferUntil
// ("Someday" snoozes) are left untouched. Returns the count of tasks returned.
func CheckDeferredReturns(pool *TaskPool) int {
	now := time.Now().UTC()
	returned := 0
	for _, t := range pool.GetTasksByStatus(StatusDeferred) {
		if t.DeferUntil != nil && !t.DeferUntil.After(now) {
			if err := t.UpdateStatus(StatusTodo); err == nil {
				t.DeferUntil = nil
				returned++
			}
		}
	}
	return returned
}
