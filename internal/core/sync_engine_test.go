package core

import (
	"fmt"
	"testing"
	"time"
)

// =============================================================================
// Change Detection Tests
// =============================================================================

func TestDetectChanges_NoChanges(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	syncState := newTestSyncState(taskA, taskB)
	local := []*Task{taskA, taskB}
	remote := []*Task{taskA, taskB}

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.NewTasks) != 0 {
		t.Errorf("NewTasks = %d, want 0", len(changes.NewTasks))
	}
	if len(changes.DeletedTasks) != 0 {
		t.Errorf("DeletedTasks = %d, want 0", len(changes.DeletedTasks))
	}
	if len(changes.ModifiedTasks) != 0 {
		t.Errorf("ModifiedTasks = %d, want 0", len(changes.ModifiedTasks))
	}
	if len(changes.Conflicts) != 0 {
		t.Errorf("Conflicts = %d, want 0", len(changes.Conflicts))
	}
}

func TestDetectChanges_NewRemoteTask(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskNew := newTestTask("ddd", "New iPhone task", StatusTodo, laterTime)

	syncState := newTestSyncState(taskA)
	local := []*Task{taskA}
	remote := []*Task{taskA, taskNew}

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.NewTasks) != 1 {
		t.Fatalf("NewTasks = %d, want 1", len(changes.NewTasks))
	}
	if changes.NewTasks[0].ID != "ddd" {
		t.Errorf("NewTasks[0].ID = %q, want %q", changes.NewTasks[0].ID, "ddd")
	}
	if changes.NewTasks[0].Text != "New iPhone task" {
		t.Errorf("NewTasks[0].Text = %q, want %q", changes.NewTasks[0].Text, "New iPhone task")
	}
}

func TestDetectChanges_DeletedRemoteTask(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	syncState := newTestSyncState(taskA, taskB)
	local := []*Task{taskA, taskB}
	remote := []*Task{taskA} // taskB deleted on iPhone

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.DeletedTasks) != 1 {
		t.Fatalf("DeletedTasks = %d, want 1", len(changes.DeletedTasks))
	}
	if changes.DeletedTasks[0] != "bbb" {
		t.Errorf("DeletedTasks[0] = %q, want %q", changes.DeletedTasks[0], "bbb")
	}
}

func TestDetectChanges_ModifiedRemoteTask(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskARemote := newTestTask("aaa", "Task A updated", StatusTodo, laterTime)

	syncState := newTestSyncState(taskA)
	local := []*Task{taskA}
	remote := []*Task{taskARemote}

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.ModifiedTasks) != 1 {
		t.Fatalf("ModifiedTasks = %d, want 1", len(changes.ModifiedTasks))
	}
	if changes.ModifiedTasks[0].Text != "Task A updated" {
		t.Errorf("ModifiedTasks[0].Text = %q, want %q", changes.ModifiedTasks[0].Text, "Task A updated")
	}
}

func TestDetectChanges_Conflict(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskALocal := newTestTask("aaa", "Task A local edit", StatusInProgress, latestTime)
	taskARemote := newTestTask("aaa", "Task A remote edit", StatusTodo, laterTime)

	// Mark task A as dirty (locally modified)
	syncState := newDirtySyncState([]string{"aaa"}, taskA)
	local := []*Task{taskALocal}
	remote := []*Task{taskARemote}

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.Conflicts) != 1 {
		t.Fatalf("Conflicts = %d, want 1", len(changes.Conflicts))
	}
	if changes.Conflicts[0].LocalTask.Text != "Task A local edit" {
		t.Errorf("Conflict LocalTask.Text = %q, want %q", changes.Conflicts[0].LocalTask.Text, "Task A local edit")
	}
	if changes.Conflicts[0].RemoteTask.Text != "Task A remote edit" {
		t.Errorf("Conflict RemoteTask.Text = %q, want %q", changes.Conflicts[0].RemoteTask.Text, "Task A remote edit")
	}
}

func TestDetectChanges_FirstSync_EmptySyncState(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)
	taskNew := newTestTask("ddd", "New iPhone task", StatusTodo, laterTime)

	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	local := []*Task{taskA, taskB}
	remote := []*Task{taskA, taskNew}

	engine := NewSyncEngine()
	changes := engine.DetectChanges(emptySyncState, local, remote)

	// taskNew should be detected as new (not in SyncState)
	if len(changes.NewTasks) != 1 {
		t.Errorf("NewTasks = %d, want 1 (taskNew)", len(changes.NewTasks))
	}
	// No deletions expected on first sync (nothing in SyncState to compare)
	if len(changes.DeletedTasks) != 0 {
		t.Errorf("DeletedTasks = %d, want 0 (first sync)", len(changes.DeletedTasks))
	}
}

func TestDetectChanges_MixedChanges(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)
	taskC := newTestTask("ccc", "Task C", StatusTodo, baseTime)
	taskARemote := newTestTask("aaa", "Task A updated", StatusTodo, laterTime)
	taskNew := newTestTask("ddd", "New task", StatusTodo, laterTime)

	syncState := newTestSyncState(taskA, taskB, taskC)
	local := []*Task{taskA, taskB, taskC}
	remote := []*Task{taskARemote, taskC, taskNew} // A modified, B deleted, new D

	engine := NewSyncEngine()
	changes := engine.DetectChanges(syncState, local, remote)

	if len(changes.ModifiedTasks) != 1 {
		t.Errorf("ModifiedTasks = %d, want 1", len(changes.ModifiedTasks))
	}
	if len(changes.DeletedTasks) != 1 {
		t.Errorf("DeletedTasks = %d, want 1", len(changes.DeletedTasks))
	}
	if len(changes.NewTasks) != 1 {
		t.Errorf("NewTasks = %d, want 1", len(changes.NewTasks))
	}
}

// =============================================================================
// Conflict Resolution Tests
// =============================================================================

func TestResolveConflicts_RemoteNewer(t *testing.T) {
	localTask := newTestTask("aaa", "Local edit", StatusTodo, baseTime)
	remoteTask := newTestTask("aaa", "Remote edit", StatusTodo, laterTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("Resolutions = %d, want 1", len(resolutions))
	}
	if resolutions[0].Winner != "remote" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "remote")
	}
	if !resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be true")
	}
	if resolutions[0].WinningTask.Text != "Remote edit" {
		t.Errorf("WinningTask.Text = %q, want %q", resolutions[0].WinningTask.Text, "Remote edit")
	}
}

func TestResolveConflicts_LocalNewer(t *testing.T) {
	localTask := newTestTask("aaa", "Local edit", StatusTodo, latestTime)
	remoteTask := newTestTask("aaa", "Remote edit", StatusTodo, laterTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("Resolutions = %d, want 1", len(resolutions))
	}
	if resolutions[0].Winner != "local" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "local")
	}
	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false when local wins")
	}
	if resolutions[0].WinningTask.Text != "Local edit" {
		t.Errorf("WinningTask.Text = %q, want %q", resolutions[0].WinningTask.Text, "Local edit")
	}
}

func TestResolveConflicts_SameTimestamp_RemoteWins(t *testing.T) {
	localTask := newTestTask("aaa", "Local edit", StatusTodo, laterTime)
	remoteTask := newTestTask("aaa", "Remote edit", StatusTodo, laterTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("Resolutions = %d, want 1", len(resolutions))
	}
	// Tiebreak: remote wins
	if resolutions[0].Winner != "remote" {
		t.Errorf("Winner = %q, want %q (tiebreak to remote)", resolutions[0].Winner, "remote")
	}
}

func TestResolveConflicts_RemoteCompleted(t *testing.T) {
	localTask := newTestTask("aaa", "Local edit", StatusInProgress, laterTime)
	remoteTask := newCompletedTestTask("aaa", "Remote completed", latestTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("Resolutions = %d, want 1", len(resolutions))
	}
	if resolutions[0].Winner != "remote" {
		t.Errorf("Winner = %q, want %q (remote completion wins)", resolutions[0].Winner, "remote")
	}
	if resolutions[0].WinningTask.Status != StatusComplete {
		t.Errorf("WinningTask.Status = %q, want %q", resolutions[0].WinningTask.Status, StatusComplete)
	}
}

func TestResolveConflicts_IdenticalChanges_NoConflict(t *testing.T) {
	// Both local and remote have the same text and status - should be a no-op
	localTask := newTestTask("aaa", "Same text", StatusInProgress, laterTime)
	remoteTask := newTestTask("aaa", "Same text", StatusInProgress, laterTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	// Identical changes should resolve without marking as override
	if len(resolutions) != 1 {
		t.Fatalf("Resolutions = %d, want 1", len(resolutions))
	}
	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false for identical changes")
	}
}

func TestResolveConflicts_MultipleConflicts(t *testing.T) {
	conflicts := []Conflict{
		{
			LocalTask:  newTestTask("aaa", "A local", StatusTodo, baseTime),
			RemoteTask: newTestTask("aaa", "A remote", StatusTodo, laterTime),
		},
		{
			LocalTask:  newTestTask("bbb", "B local", StatusTodo, latestTime),
			RemoteTask: newTestTask("bbb", "B remote", StatusTodo, laterTime),
		},
	}

	engine := NewSyncEngine()
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) != 2 {
		t.Fatalf("Resolutions = %d, want 2", len(resolutions))
	}

	// First: remote newer → remote wins
	if resolutions[0].Winner != "remote" {
		t.Errorf("Resolution[0].Winner = %q, want %q", resolutions[0].Winner, "remote")
	}
	// Second: local newer → local wins
	if resolutions[1].Winner != "local" {
		t.Errorf("Resolution[1].Winner = %q, want %q", resolutions[1].Winner, "local")
	}
}

func TestResolveConflicts_GeneratesMessage(t *testing.T) {
	localTask := newTestTask("aaa", "Fix the bug", StatusTodo, baseTime)
	remoteTask := newTestTask("aaa", "Fix the bug (updated)", StatusTodo, laterTime)

	engine := NewSyncEngine()
	conflicts := []Conflict{{LocalTask: localTask, RemoteTask: remoteTask}}
	resolutions := engine.ResolveConflicts(conflicts)

	if len(resolutions) == 0 {
		t.Fatal("Resolutions should not be empty")
	}
	if resolutions[0].Message == "" {
		t.Error("Resolution should have a user-facing message")
	}
}

// =============================================================================
// Apply Changes Tests
// =============================================================================

func TestApplyChanges_AddNewTasks(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskNew := newTestTask("ddd", "New task", StatusTodo, laterTime)

	pool := poolFromTasks(taskA)
	changes := ChangeSet{NewTasks: []*Task{taskNew}}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, nil)

	if pool.Count() != 2 {
		t.Errorf("Pool count = %d, want 2", pool.Count())
	}
	if pool.GetTask("ddd") == nil {
		t.Error("New task 'ddd' not found in pool")
	}
	if result.Added != 1 {
		t.Errorf("Result.Added = %d, want 1", result.Added)
	}
}

func TestApplyChanges_RemoveDeletedTasks(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	pool := poolFromTasks(taskA, taskB)
	changes := ChangeSet{DeletedTasks: []string{"bbb"}}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, nil)

	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1", pool.Count())
	}
	if pool.GetTask("bbb") != nil {
		t.Error("Deleted task 'bbb' still in pool")
	}
	if result.Removed != 1 {
		t.Errorf("Result.Removed = %d, want 1", result.Removed)
	}
}

func TestApplyChanges_UpdateModifiedTasks(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskARemote := newTestTask("aaa", "Task A updated", StatusTodo, laterTime)

	pool := poolFromTasks(taskA)
	changes := ChangeSet{ModifiedTasks: []*Task{taskARemote}}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, nil)

	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1", pool.Count())
	}
	updated := pool.GetTask("aaa")
	if updated.Text != "Task A updated" {
		t.Errorf("Updated task text = %q, want %q", updated.Text, "Task A updated")
	}
	if result.Updated != 1 {
		t.Errorf("Result.Updated = %d, want 1", result.Updated)
	}
}

func TestApplyChanges_MixedChanges(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)
	taskARemote := newTestTask("aaa", "Task A updated", StatusTodo, laterTime)
	taskNew := newTestTask("ccc", "New task", StatusTodo, laterTime)

	pool := poolFromTasks(taskA, taskB)
	changes := ChangeSet{
		NewTasks:      []*Task{taskNew},
		DeletedTasks:  []string{"bbb"},
		ModifiedTasks: []*Task{taskARemote},
	}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, nil)

	if pool.Count() != 2 {
		t.Errorf("Pool count = %d, want 2 (A updated + C new)", pool.Count())
	}
	if pool.GetTask("aaa").Text != "Task A updated" {
		t.Error("Task A should be updated")
	}
	if pool.GetTask("bbb") != nil {
		t.Error("Task B should be removed")
	}
	if pool.GetTask("ccc") == nil {
		t.Error("Task C should be added")
	}
	if result.Added != 1 || result.Updated != 1 || result.Removed != 1 {
		t.Errorf("Result = {Added:%d, Updated:%d, Removed:%d}, want {1, 1, 1}",
			result.Added, result.Updated, result.Removed)
	}
}

func TestApplyChanges_WithConflictResolutions(t *testing.T) {
	taskA := newTestTask("aaa", "Task A local", StatusTodo, baseTime)
	taskAResolved := newTestTask("aaa", "Task A remote wins", StatusTodo, laterTime)

	pool := poolFromTasks(taskA)
	changes := ChangeSet{} // conflicts handled separately via resolutions
	resolutions := []Resolution{
		{
			TaskID:          "aaa",
			Winner:          "remote",
			WinningTask:     taskAResolved,
			LocalOverridden: true,
			Message:         "Your change to 'Task A local' was overridden",
		},
	}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, resolutions)

	updated := pool.GetTask("aaa")
	if updated.Text != "Task A remote wins" {
		t.Errorf("Task text = %q, want %q", updated.Text, "Task A remote wins")
	}
	if result.Conflicts != 1 {
		t.Errorf("Result.Conflicts = %d, want 1", result.Conflicts)
	}
	if len(result.Overrides) != 1 {
		t.Errorf("Result.Overrides = %d, want 1", len(result.Overrides))
	}
}

func TestApplyChanges_GeneratesSummary(t *testing.T) {
	taskNew := newTestTask("ddd", "New task", StatusTodo, laterTime)
	pool := NewTaskPool()
	changes := ChangeSet{NewTasks: []*Task{taskNew}}

	engine := NewSyncEngine()
	result := engine.ApplyChanges(pool, changes, nil)

	if result.Summary == "" {
		t.Error("Result.Summary should not be empty")
	}
}

// =============================================================================
// Integration Tests: Full Sync Cycle
// =============================================================================

func TestSync_FullCycle(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)
	taskARemote := newTestTask("aaa", "Task A updated", StatusTodo, laterTime)
	taskNew := newTestTask("ccc", "New task from iPhone", StatusTodo, laterTime)

	// Setup: local has A and B, remote has A(modified) and C(new), B deleted
	provider := &MockProvider{
		Tasks: []*Task{taskARemote, taskNew},
	}
	syncState := newTestSyncState(taskA, taskB)
	pool := poolFromTasks(taskA, taskB)

	engine := NewSyncEngine()

	// Run full sync
	result, err := engine.Sync(provider, syncState, pool)
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}

	// Verify pool state
	if pool.Count() != 2 {
		t.Errorf("Pool count = %d, want 2", pool.Count())
	}
	if pool.GetTask("aaa").Text != "Task A updated" {
		t.Error("Task A should be updated")
	}
	if pool.GetTask("bbb") != nil {
		t.Error("Task B should be removed")
	}
	if pool.GetTask("ccc") == nil {
		t.Error("Task C should be added")
	}

	// Verify sync result
	if result.Added != 1 {
		t.Errorf("Result.Added = %d, want 1", result.Added)
	}
	if result.Updated != 1 {
		t.Errorf("Result.Updated = %d, want 1", result.Updated)
	}
	if result.Removed != 1 {
		t.Errorf("Result.Removed = %d, want 1", result.Removed)
	}
}

func TestSync_ProviderLoadError(t *testing.T) {
	provider := &MockProvider{
		LoadErr: fmt.Errorf("Apple Notes unavailable"),
	}
	syncState := newTestSyncState()
	pool := poolFromTasks(newTestTask("aaa", "Task A", StatusTodo, baseTime))

	engine := NewSyncEngine()
	_, err := engine.Sync(provider, syncState, pool)

	if err == nil {
		t.Error("Sync() should return error when provider fails")
	}

	// Pool should be unchanged
	if pool.Count() != 1 {
		t.Errorf("Pool should be unchanged after sync error, count = %d, want 1", pool.Count())
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestSync_FirstSync_EmptySyncState(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	provider := &MockProvider{Tasks: []*Task{taskA, taskB}}
	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	pool := NewTaskPool()

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, emptySyncState, pool)
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}
	if pool.Count() != 2 {
		t.Errorf("Pool count = %d, want 2 (all remote tasks added)", pool.Count())
	}
	if result.Added != 2 {
		t.Errorf("Result.Added = %d, want 2", result.Added)
	}
}

func TestSync_EmptyRemote_AllDeleted(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	provider := &MockProvider{Tasks: []*Task{}} // all deleted
	syncState := newTestSyncState(taskA, taskB)
	pool := poolFromTasks(taskA, taskB)

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, syncState, pool)
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}
	if pool.Count() != 0 {
		t.Errorf("Pool count = %d, want 0 (all deleted)", pool.Count())
	}
	if result.Removed != 2 {
		t.Errorf("Result.Removed = %d, want 2", result.Removed)
	}
}

func TestSync_CorruptRemoteTask_Skipped(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	corruptTask := &Task{ID: "", Text: "", Status: StatusTodo} // empty ID

	provider := &MockProvider{Tasks: []*Task{taskA, corruptTask}}
	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	pool := NewTaskPool()

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, emptySyncState, pool)
	if err != nil {
		t.Fatalf("Sync() should not fail for corrupt tasks, got: %v", err)
	}
	// Only valid task should be added
	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1 (corrupt task skipped)", pool.Count())
	}
	if len(result.Errors) == 0 {
		t.Error("Result.Errors should contain entry for corrupt task")
	}
}

func TestSync_EmptyTaskText_Skipped(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	emptyText := newTestTask("bbb", "", StatusTodo, baseTime)

	provider := &MockProvider{Tasks: []*Task{taskA, emptyText}}
	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	pool := NewTaskPool()

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, emptySyncState, pool)
	if err != nil {
		t.Fatalf("Sync() should not fail for empty text tasks, got: %v", err)
	}
	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1 (empty text task skipped)", pool.Count())
	}
	if len(result.Errors) == 0 {
		t.Error("Result.Errors should contain entry for empty text task")
	}
}

func TestSync_ZeroTasks_BothSides(t *testing.T) {
	provider := &MockProvider{Tasks: []*Task{}}
	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	pool := NewTaskPool()

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, emptySyncState, pool)
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}
	if result.Added != 0 || result.Updated != 0 || result.Removed != 0 || result.Conflicts != 0 {
		t.Errorf("Result should be all zeros, got Added:%d Updated:%d Removed:%d Conflicts:%d",
			result.Added, result.Updated, result.Removed, result.Conflicts)
	}
}

func TestSync_NilRemoteTasks_Skipped(t *testing.T) {
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)

	// Provider returns list with nil entry
	provider := &MockProvider{Tasks: []*Task{taskA, nil}}
	emptySyncState := SyncState{TaskSnapshots: make(map[string]TaskSnapshot)}
	pool := NewTaskPool()

	engine := NewSyncEngine()
	result, err := engine.Sync(provider, emptySyncState, pool)
	if err != nil {
		t.Fatalf("Sync() should not fail for nil tasks, got: %v", err)
	}
	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1 (nil task skipped)", pool.Count())
	}
	if len(result.Errors) == 0 {
		t.Error("Result.Errors should contain entry for nil task")
	}
}

// =============================================================================
// Performance Test
// =============================================================================

func TestSync_Performance100Tasks(t *testing.T) {
	const taskCount = 100

	remoteTasks := make([]*Task, taskCount)
	localTasks := make([]*Task, taskCount)
	for i := range taskCount {
		id := fmt.Sprintf("id-%d", i)
		remoteTasks[i] = newTestTask(id, fmt.Sprintf("Task %d", i), StatusTodo, baseTime)
		localTasks[i] = newTestTask(id, fmt.Sprintf("Task %d", i), StatusTodo, baseTime)
	}

	// Modify 30% remotely
	for i := range 30 {
		remoteTasks[i] = newTestTask(fmt.Sprintf("id-%d", i), fmt.Sprintf("Task %d modified", i), StatusTodo, laterTime)
	}

	// Delete 20% remotely (remove from slice)
	remoteTasks = remoteTasks[:80]

	// Add 10 new remote tasks
	for i := range 10 {
		remoteTasks = append(remoteTasks, newTestTask(
			fmt.Sprintf("new-%d", i), fmt.Sprintf("New task %d", i), StatusTodo, laterTime,
		))
	}

	provider := &MockProvider{Tasks: remoteTasks}
	syncState := newTestSyncState(localTasks...)
	pool := poolFromTasks(localTasks...)

	engine := NewSyncEngine()

	start := time.Now()
	_, err := engine.Sync(provider, syncState, pool)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("Sync took %v, expected < 2s", elapsed)
	}
	t.Logf("Sync of %d tasks completed in %v", taskCount, elapsed)
}
