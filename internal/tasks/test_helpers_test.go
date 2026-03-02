package tasks

import (
	"time"
)

// Shared test time constants for deterministic testing.
var (
	baseTime   = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	laterTime  = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	latestTime = time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
)

// newTestTask creates a Task with explicit fields for deterministic testing.
// No UUID generation - caller controls the ID.
func newTestTask(id, text string, status TaskStatus, updatedAt time.Time) *Task {
	return &Task{
		ID:        id,
		Text:      text,
		Status:    status,
		Notes:     []TaskNote{},
		CreatedAt: updatedAt,
		UpdatedAt: updatedAt,
	}
}

// newCompletedTestTask creates a completed Task with CompletedAt set.
func newCompletedTestTask(id, text string, completedAt time.Time) *Task {
	return &Task{
		ID:          id,
		Text:        text,
		Status:      StatusComplete,
		Notes:       []TaskNote{},
		CreatedAt:   completedAt.Add(-1 * time.Hour),
		UpdatedAt:   completedAt,
		CompletedAt: &completedAt,
	}
}

// newTestSyncState creates a SyncState from a list of tasks.
// All task snapshots have Dirty=false.
func newTestSyncState(tasks ...*Task) SyncState {
	state := SyncState{
		LastSyncTime:  time.Now().UTC(),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}
	for _, t := range tasks {
		state.TaskSnapshots[t.ID] = TaskSnapshot{
			ID:        t.ID,
			Text:      t.Text,
			Status:    t.Status,
			UpdatedAt: t.UpdatedAt,
			Dirty:     false,
		}
	}
	return state
}

// newDirtySyncState creates a SyncState where specified task IDs are marked Dirty.
func newDirtySyncState(dirtyIDs []string, tasks ...*Task) SyncState {
	state := newTestSyncState(tasks...)
	dirtySet := make(map[string]bool)
	for _, id := range dirtyIDs {
		dirtySet[id] = true
	}
	for id, snap := range state.TaskSnapshots {
		if dirtySet[id] {
			snap.Dirty = true
			state.TaskSnapshots[id] = snap
		}
	}
	return state
}

// poolFromTasks creates a TaskPool populated with the given tasks.
func poolFromTasks(tasks ...*Task) *TaskPool {
	pool := NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return pool
}
