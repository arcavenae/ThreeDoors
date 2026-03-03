package core

// TaskPool manages an in-memory collection of tasks.
type TaskPool struct {
	tasks            map[string]*Task
	recentlyShown    []string
	recentlyShownIdx int
	maxRecentlyShown int
}

// NewTaskPool creates a new empty TaskPool.
func NewTaskPool() *TaskPool {
	return &TaskPool{
		tasks:            make(map[string]*Task),
		recentlyShown:    make([]string, 10),
		recentlyShownIdx: 0,
		maxRecentlyShown: 10,
	}
}

// AddTask adds a task to the pool.
func (tp *TaskPool) AddTask(task *Task) {
	tp.tasks[task.ID] = task
}

// GetTask retrieves a task by ID.
func (tp *TaskPool) GetTask(id string) *Task {
	return tp.tasks[id]
}

// UpdateTask updates an existing task in the pool.
func (tp *TaskPool) UpdateTask(task *Task) {
	tp.tasks[task.ID] = task
}

// RemoveTask removes a task from the pool by ID.
func (tp *TaskPool) RemoveTask(id string) {
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
// Eligible: status is todo, blocked, or in-progress, and not recently shown.
func (tp *TaskPool) GetAvailableForDoors() []*Task {
	var result []*Task
	for _, t := range tp.tasks {
		if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {
			if !tp.IsRecentlyShown(t.ID) {
				result = append(result, t)
			}
		}
	}
	// If not enough non-recent tasks, include recently shown ones
	if len(result) < 3 {
		result = nil
		for _, t := range tp.tasks {
			if t.Status == StatusTodo || t.Status == StatusBlocked || t.Status == StatusInProgress {
				result = append(result, t)
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
