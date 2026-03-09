package core

// HasUnmetDependencies returns true if any of the task's dependencies
// are not in complete status. Missing dependencies (orphaned IDs) are
// treated as unmet (pessimistic).
func HasUnmetDependencies(task *Task, pool *TaskPool) bool {
	for _, depID := range task.DependsOn {
		dep := pool.GetTask(depID)
		if dep == nil || dep.Status != StatusComplete {
			return true
		}
	}
	return false
}

// GetBlockingDependencies returns the subset of a task's dependencies
// that are not yet complete. Returns nil if all dependencies are met.
// Orphaned IDs return a placeholder with text "[deleted task]".
func GetBlockingDependencies(task *Task, pool *TaskPool) []*Task {
	var blocking []*Task
	for _, depID := range task.DependsOn {
		dep := pool.GetTask(depID)
		if dep == nil {
			blocking = append(blocking, &Task{
				ID:   depID,
				Text: "[deleted task]",
			})
		} else if dep.Status != StatusComplete {
			blocking = append(blocking, dep)
		}
	}
	return blocking
}

// WouldCreateCycle returns true if adding a dependency from taskID to
// depID would create a circular dependency chain. Uses depth-first
// search through the dependency graph. Self-dependencies (taskID == depID)
// are detected as cycles.
func WouldCreateCycle(taskID, depID string, pool *TaskPool) bool {
	if taskID == depID {
		return true
	}
	visited := make(map[string]bool)
	return hasDependencyPath(depID, taskID, pool, visited)
}

// hasDependencyPath returns true if there is a dependency path from
// fromID to toID in the pool's dependency graph.
func hasDependencyPath(fromID, toID string, pool *TaskPool, visited map[string]bool) bool {
	if fromID == toID {
		return true
	}
	if visited[fromID] {
		return false
	}
	visited[fromID] = true

	task := pool.GetTask(fromID)
	if task == nil {
		return false
	}
	for _, depID := range task.DependsOn {
		if hasDependencyPath(depID, toID, pool, visited) {
			return true
		}
	}
	return false
}

// GetNewlyUnblockedTasks returns tasks that depended on completedTaskID
// and now have all their dependencies met.
func GetNewlyUnblockedTasks(completedTaskID string, pool *TaskPool) []*Task {
	var unblocked []*Task
	for _, task := range pool.GetAllTasks() {
		for _, depID := range task.DependsOn {
			if depID == completedTaskID {
				if !HasUnmetDependencies(task, pool) {
					unblocked = append(unblocked, task)
				}
				break
			}
		}
	}
	return unblocked
}
