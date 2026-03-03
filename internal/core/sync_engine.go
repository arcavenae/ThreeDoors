package core

import "fmt"

// ChangeSet represents the differences detected between local and remote task lists.
type ChangeSet struct {
	NewTasks      []*Task    // exist in remote, not in lastSync
	DeletedTasks  []string   // exist in lastSync, not in remote (task IDs)
	ModifiedTasks []*Task    // exist in both, remote.UpdatedAt > lastSync snapshot
	Conflicts     []Conflict // modified both locally (dirty) and remotely
}

// Conflict represents a task modified both locally and remotely.
type Conflict struct {
	LocalTask  *Task
	RemoteTask *Task
}

// Resolution represents the outcome of a conflict resolution.
type Resolution struct {
	TaskID          string
	Winner          string // "local" or "remote"
	WinningTask     *Task
	LocalOverridden bool
	Message         string // user-facing notification
}

// SyncResult summarizes the outcome of a sync operation.
type SyncResult struct {
	Added     int
	Updated   int
	Removed   int
	Conflicts int
	Overrides []Resolution // only where LocalOverridden=true
	Errors    []error
	Summary   string // e.g. "Synced: 2 new, 1 updated, 1 removed"
}

// SyncEngine performs three-way sync between local and remote task lists.
type SyncEngine struct{}

// NewSyncEngine creates a new SyncEngine.
func NewSyncEngine() *SyncEngine {
	return &SyncEngine{}
}

// DetectChanges compares local and remote tasks against the last-known sync state
// to identify new, deleted, modified tasks, and conflicts.
func (e *SyncEngine) DetectChanges(lastSync SyncState, local []*Task, remote []*Task) ChangeSet {
	var changes ChangeSet

	// Build lookup maps
	remoteByID := make(map[string]*Task, len(remote))
	for _, t := range remote {
		remoteByID[t.ID] = t
	}

	localByID := make(map[string]*Task, len(local))
	for _, t := range local {
		localByID[t.ID] = t
	}

	// Detect new and modified remote tasks
	for _, rt := range remote {
		snap, existsInSync := lastSync.TaskSnapshots[rt.ID]
		if !existsInSync {
			// Task not in last sync state - check if it exists locally
			if _, existsLocal := localByID[rt.ID]; !existsLocal {
				// Not in sync state AND not local → truly new task
				changes.NewTasks = append(changes.NewTasks, rt)
			}
			// If exists locally but not in sync state, both sides have it already (first sync scenario)
			continue
		}

		// Task exists in sync state - check if modified remotely
		if rt.UpdatedAt.After(snap.UpdatedAt) {
			// Remote was modified since last sync
			if snap.Dirty {
				// Also modified locally → conflict
				lt := localByID[rt.ID]
				if lt != nil {
					changes.Conflicts = append(changes.Conflicts, Conflict{
						LocalTask:  lt,
						RemoteTask: rt,
					})
				}
			} else {
				// Only remote changed → modified
				changes.ModifiedTasks = append(changes.ModifiedTasks, rt)
			}
		}
	}

	// Detect deleted tasks (in sync state but not in remote)
	for id := range lastSync.TaskSnapshots {
		if _, existsRemote := remoteByID[id]; !existsRemote {
			changes.DeletedTasks = append(changes.DeletedTasks, id)
		}
	}

	return changes
}

// ResolveConflicts applies last-write-wins strategy to resolve conflicts.
// If timestamps are equal, remote wins (tiebreak).
// If changes are identical (same text and status), no override is reported.
func (e *SyncEngine) ResolveConflicts(conflicts []Conflict) []Resolution {
	resolutions := make([]Resolution, 0, len(conflicts))

	for _, c := range conflicts {
		r := Resolution{
			TaskID: c.LocalTask.ID,
		}

		// Check for identical changes (same text and status = no real conflict)
		identical := c.LocalTask.Text == c.RemoteTask.Text &&
			c.LocalTask.Status == c.RemoteTask.Status

		if identical {
			// No real conflict - pick remote as canonical, no override
			r.Winner = "remote"
			r.WinningTask = c.RemoteTask
			r.LocalOverridden = false
			r.Message = ""
			resolutions = append(resolutions, r)
			continue
		}

		// Last-write-wins: compare UpdatedAt timestamps
		// If equal, remote wins (tiebreak)
		if c.LocalTask.UpdatedAt.After(c.RemoteTask.UpdatedAt) {
			r.Winner = "local"
			r.WinningTask = c.LocalTask
			r.LocalOverridden = false
			r.Message = fmt.Sprintf("Local change to '%s' kept (newer than iPhone edit)", c.LocalTask.Text)
		} else {
			r.Winner = "remote"
			r.WinningTask = c.RemoteTask
			r.LocalOverridden = true
			r.Message = fmt.Sprintf("Your change to '%s' was overridden by a newer iPhone edit", c.LocalTask.Text)
		}

		resolutions = append(resolutions, r)
	}

	return resolutions
}

// ApplyChanges merges the detected changes and resolutions into the TaskPool.
func (e *SyncEngine) ApplyChanges(pool *TaskPool, changes ChangeSet, resolutions []Resolution) SyncResult {
	var result SyncResult

	// Add new tasks
	for _, t := range changes.NewTasks {
		pool.AddTask(t)
		result.Added++
	}

	// Remove deleted tasks
	for _, id := range changes.DeletedTasks {
		pool.RemoveTask(id)
		result.Removed++
	}

	// Update modified tasks (non-conflicting)
	for _, t := range changes.ModifiedTasks {
		pool.UpdateTask(t)
		result.Updated++
	}

	// Apply conflict resolutions
	for _, r := range resolutions {
		pool.UpdateTask(r.WinningTask)
		result.Conflicts++
		if r.LocalOverridden {
			result.Overrides = append(result.Overrides, r)
		}
	}

	// Generate summary
	result.Summary = fmt.Sprintf("Synced: %d new, %d updated, %d removed", result.Added, result.Updated, result.Removed)
	if result.Conflicts > 0 {
		result.Summary += fmt.Sprintf(", %d conflicts resolved", result.Conflicts)
	}

	return result
}

// Sync performs a full sync cycle: load remote, detect changes, resolve conflicts, apply.
// Returns SyncResult with counts and any override notifications.
// On provider error, returns error and leaves pool unchanged.
func (e *SyncEngine) Sync(provider TaskProvider, syncState SyncState, pool *TaskPool) (SyncResult, error) {
	remote, err := provider.LoadTasks()
	if err != nil {
		return SyncResult{}, fmt.Errorf("sync: load remote: %w", err)
	}

	// Filter invalid remote tasks
	var validRemote []*Task
	var syncErrors []error
	for _, t := range remote {
		if t == nil {
			syncErrors = append(syncErrors, fmt.Errorf("sync: nil task in remote list"))
			continue
		}
		if t.ID == "" {
			syncErrors = append(syncErrors, fmt.Errorf("sync: task with empty ID skipped"))
			continue
		}
		if t.Text == "" {
			syncErrors = append(syncErrors, fmt.Errorf("sync: task %q has empty text, skipped", t.ID))
			continue
		}
		validRemote = append(validRemote, t)
	}

	local := pool.GetAllTasks()
	changes := e.DetectChanges(syncState, local, validRemote)
	resolutions := e.ResolveConflicts(changes.Conflicts)
	result := e.ApplyChanges(pool, changes, resolutions)
	result.Errors = append(result.Errors, syncErrors...)

	return result, nil
}
