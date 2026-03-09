package core

import "time"

// sourceRefKey is the composite key for the SourceRef index.
type sourceRefKey struct {
	Provider string
	NativeID string
}

// TaskPool manages an in-memory collection of tasks.
type TaskPool struct {
	tasks            map[string]*Task
	sourceRefIndex   map[sourceRefKey]string // sourceRefKey → task ID
	recentlyShown    []string
	recentlyShownIdx int
	maxRecentlyShown int
}

// NewTaskPool creates a new empty TaskPool.
func NewTaskPool() *TaskPool {
	return &TaskPool{
		tasks:            make(map[string]*Task),
		sourceRefIndex:   make(map[sourceRefKey]string),
		recentlyShown:    make([]string, 10),
		recentlyShownIdx: 0,
		maxRecentlyShown: 10,
	}
}

// AddTask adds a task to the pool and indexes its SourceRefs.
func (tp *TaskPool) AddTask(task *Task) {
	tp.tasks[task.ID] = task
	tp.indexSourceRefs(task)
}

// GetTask retrieves a task by ID.
func (tp *TaskPool) GetTask(id string) *Task {
	return tp.tasks[id]
}

// UpdateTask updates an existing task in the pool and re-indexes its SourceRefs.
func (tp *TaskPool) UpdateTask(task *Task) {
	// Remove old index entries by iterating stored keys (handles same-pointer mutation).
	tp.removeSourceRefIndexByID(task.ID)
	tp.tasks[task.ID] = task
	tp.indexSourceRefs(task)
}

// RemoveTask removes a task from the pool by ID and cleans up its SourceRef index.
func (tp *TaskPool) RemoveTask(id string) {
	if task, ok := tp.tasks[id]; ok {
		tp.removeSourceRefIndex(task)
	}
	delete(tp.tasks, id)
}

// GetAllTasks returns all tasks in the pool.
func (tp *TaskPool) GetAllTasks() []*Task {
	result := make([]*Task, 0, len(tp.tasks))
	for _, t := range tp.tasks {
		result = append(result, t)
	}
	return result
}

// GetTasksByStatus returns tasks filtered by status.
func (tp *TaskPool) GetTasksByStatus(status TaskStatus) []*Task {
	var result []*Task
	for _, t := range tp.tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

// GetAvailableForDoors returns tasks eligible for door selection.
// Eligible: status is todo, blocked, or in-progress, not recently shown,
// and not blocked by unmet dependencies.
func (tp *TaskPool) GetAvailableForDoors() []*Task {
	var result []*Task
	for _, t := range tp.tasks {
		if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {
			if !tp.IsRecentlyShown(t.ID) && !HasUnmetDependencies(t, tp) {
				result = append(result, t)
			}
		}
	}
	// If not enough non-recent tasks, include recently shown ones (still excluding dependency-blocked)
	if len(result) < 3 {
		result = nil
		for _, t := range tp.tasks {
			if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {
				if !HasUnmetDependencies(t, tp) {
					result = append(result, t)
				}
			}
		}
	}
	return result
}

// MarkRecentlyShown adds a task ID to the recently shown ring buffer.
func (tp *TaskPool) MarkRecentlyShown(taskID string) {
	tp.recentlyShown[tp.recentlyShownIdx%tp.maxRecentlyShown] = taskID
	tp.recentlyShownIdx++
}

// IsRecentlyShown checks if a task ID is in the recently shown buffer.
func (tp *TaskPool) IsRecentlyShown(taskID string) bool {
	for _, id := range tp.recentlyShown {
		if id == taskID {
			return true
		}
	}
	return false
}

// Count returns the total number of tasks.
func (tp *TaskPool) Count() int {
	return len(tp.tasks)
}

// FindBySourceRef returns the task matching the given provider and native ID,
// or nil if no match is found. Uses an internal index for O(1) lookup.
func (tp *TaskPool) FindBySourceRef(provider, nativeID string) *Task {
	key := sourceRefKey{Provider: provider, NativeID: nativeID}
	if id, ok := tp.sourceRefIndex[key]; ok {
		return tp.tasks[id]
	}
	return nil
}

// indexSourceRefs adds all SourceRefs of a task to the index.
func (tp *TaskPool) indexSourceRefs(task *Task) {
	for _, ref := range task.SourceRefs {
		key := sourceRefKey(ref)
		tp.sourceRefIndex[key] = task.ID
	}
}

// removeSourceRefIndex removes all SourceRefs of a task from the index.
func (tp *TaskPool) removeSourceRefIndex(task *Task) {
	for _, ref := range task.SourceRefs {
		key := sourceRefKey(ref)
		delete(tp.sourceRefIndex, key)
	}
}

// FindByPrefix returns all tasks whose ID starts with the given prefix.
func (tp *TaskPool) FindByPrefix(prefix string) []*Task {
	var matches []*Task
	for id, t := range tp.tasks {
		if len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			matches = append(matches, t)
		}
	}
	return matches
}

// GetIncompleteTasks returns tasks that were in todo, in-progress, or blocked
// status and were created on or before the given timestamp.
func (tp *TaskPool) GetIncompleteTasks(since time.Time) []*Task {
	var result []*Task
	for _, t := range tp.tasks {
		if t.CreatedAt.After(since) {
			continue
		}
		if t.Status == StatusTodo || t.Status == StatusInProgress || t.Status == StatusBlocked {
			result = append(result, t)
		}
	}
	return result
}

// removeSourceRefIndexByID removes all index entries pointing to the given task ID.
// This is safe even when the task's SourceRefs have already been mutated in place.
func (tp *TaskPool) removeSourceRefIndexByID(taskID string) {
	for key, id := range tp.sourceRefIndex {
		if id == taskID {
			delete(tp.sourceRefIndex, key)
		}
	}
}
