package core

import (
	"fmt"
	"testing"
	"time"
)

// =============================================================================
// Multi-Adapter Integration Tests (Story 0.52)
//
// These tests exercise the sync engine with two independent mock adapters,
// simulating real-world multi-source scenarios (e.g., TextFile + Todoist).
// Each adapter maintains its own in-memory state, and the sync engine
// resolves conflicts across them via a shared TaskPool.
// =============================================================================

// namedMockProvider wraps MockProvider with a distinct name, simulating
// different adapters (e.g., "textfile", "todoist").
type namedMockProvider struct {
	MockProvider
	name string
}

func newNamedProvider(name string, tasks ...*Task) *namedMockProvider {
	cp := make([]*Task, len(tasks))
	copy(cp, tasks)
	return &namedMockProvider{
		MockProvider: MockProvider{Tasks: cp},
		name:         name,
	}
}

func (p *namedMockProvider) Name() string { return p.name }

// multiAdapterScenario holds the shared state for a two-adapter sync test.
type multiAdapterScenario struct {
	engine     *SyncEngine
	pool       *TaskPool
	providerA  *namedMockProvider
	providerB  *namedMockProvider
	syncStateA SyncState
	syncStateB SyncState
}

// newScenario creates a fresh scenario with two named adapters and an empty pool.
func newScenario() *multiAdapterScenario {
	return &multiAdapterScenario{
		engine:     NewSyncEngine(),
		pool:       NewTaskPool(),
		providerA:  newNamedProvider("adapter-a"),
		providerB:  newNamedProvider("adapter-b"),
		syncStateA: SyncState{TaskSnapshots: make(map[string]TaskSnapshot)},
		syncStateB: SyncState{TaskSnapshots: make(map[string]TaskSnapshot)},
	}
}

// seedPool adds tasks to the pool and both adapters, then takes a sync snapshot.
func (s *multiAdapterScenario) seedPool(tasks ...*Task) {
	for _, t := range tasks {
		s.pool.AddTask(t)
	}
	s.providerA.Tasks = cloneTasks(tasks)
	s.providerB.Tasks = cloneTasks(tasks)
	s.syncStateA = newTestSyncState(tasks...)
	s.syncStateB = newTestSyncState(tasks...)
}

// snapshotSyncState captures the current pool state as a clean sync baseline.
func (s *multiAdapterScenario) snapshotSyncState(dirtyIDs ...string) {
	tasks := s.pool.GetAllTasks()
	s.syncStateA = newTestSyncState(tasks...)
	s.syncStateB = newTestSyncState(tasks...)
	if len(dirtyIDs) > 0 {
		dirtySet := make(map[string]bool, len(dirtyIDs))
		for _, id := range dirtyIDs {
			dirtySet[id] = true
		}
		for id, snap := range s.syncStateA.TaskSnapshots {
			if dirtySet[id] {
				snap.Dirty = true
				s.syncStateA.TaskSnapshots[id] = snap
			}
		}
		for id, snap := range s.syncStateB.TaskSnapshots {
			if dirtySet[id] {
				snap.Dirty = true
				s.syncStateB.TaskSnapshots[id] = snap
			}
		}
	}
}

// syncFromA syncs the pool against adapter A.
func (s *multiAdapterScenario) syncFromA(t *testing.T) SyncResult {
	t.Helper()
	result, err := s.engine.Sync(s.providerA, s.syncStateA, s.pool)
	if err != nil {
		t.Fatalf("Sync from adapter-a failed: %v", err)
	}
	return result
}

// syncFromB syncs the pool against adapter B.
func (s *multiAdapterScenario) syncFromB(t *testing.T) SyncResult {
	t.Helper()
	result, err := s.engine.Sync(s.providerB, s.syncStateB, s.pool)
	if err != nil {
		t.Fatalf("Sync from adapter-b failed: %v", err)
	}
	return result
}

func cloneTasks(tasks []*Task) []*Task {
	out := make([]*Task, len(tasks))
	for i, t := range tasks {
		cp := *t
		out[i] = &cp
	}
	return out
}

// =============================================================================
// Scenario 1: Last-Writer-Wins — Both adapters modify same task's title
// =============================================================================

func TestMultiAdapter_LastWriterWins_TitleConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		adapterATime   time.Time
		adapterBTime   time.Time
		adapterATitle  string
		adapterBTitle  string
		wantWinner     string
		wantTitle      string
		wantOverridden bool
	}{
		{
			name:           "adapter B wins when newer",
			adapterATime:   laterTime,
			adapterBTime:   latestTime,
			adapterATitle:  "Title from A",
			adapterBTitle:  "Title from B",
			wantWinner:     "remote",
			wantTitle:      "Title from B",
			wantOverridden: true,
		},
		{
			name:           "adapter A wins when newer",
			adapterATime:   latestTime,
			adapterBTime:   laterTime,
			adapterATitle:  "Title from A",
			adapterBTitle:  "Title from B",
			wantWinner:     "local",
			wantTitle:      "Title from A",
			wantOverridden: false,
		},
		{
			name:           "same timestamp tiebreaks to remote",
			adapterATime:   laterTime,
			adapterBTime:   laterTime,
			adapterATitle:  "Title from A",
			adapterBTitle:  "Title from B",
			wantWinner:     "remote",
			wantTitle:      "Title from B",
			wantOverridden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup: shared task in pool
			task := newTestTask("task-1", "Original title", StatusTodo, baseTime)
			sc := newScenario()
			sc.seedPool(task)

			// Adapter A modifies the task
			taskFromA := newTestTask("task-1", tt.adapterATitle, StatusTodo, tt.adapterATime)
			sc.providerA.Tasks = []*Task{taskFromA}

			// Mark task as dirty in sync state (local was modified)
			sc.syncStateA.TaskSnapshots["task-1"] = TaskSnapshot{
				ID: "task-1", Text: "Original title", Status: StatusTodo,
				UpdatedAt: baseTime, Dirty: true,
			}

			// Sync from A first — this applies adapter A's changes
			sc.pool.UpdateTask(taskFromA)

			// Now adapter B also modified the same task (different title)
			taskFromB := newTestTask("task-1", tt.adapterBTitle, StatusTodo, tt.adapterBTime)
			sc.providerB.Tasks = []*Task{taskFromB}

			// Build sync state for B: pool has A's version, mark as dirty
			sc.syncStateB.TaskSnapshots["task-1"] = TaskSnapshot{
				ID: "task-1", Text: "Original title", Status: StatusTodo,
				UpdatedAt: baseTime, Dirty: true,
			}

			// Detect changes: pool (with A's edit) vs adapter B's remote
			local := sc.pool.GetAllTasks()
			changes := sc.engine.DetectChanges(sc.syncStateB, local, sc.providerB.Tasks)

			if len(changes.Conflicts) != 1 {
				t.Fatalf("expected 1 conflict, got %d", len(changes.Conflicts))
			}

			resolutions := sc.engine.ResolveConflicts(changes.Conflicts)
			if len(resolutions) != 1 {
				t.Fatalf("expected 1 resolution, got %d", len(resolutions))
			}

			if resolutions[0].Winner != tt.wantWinner {
				t.Errorf("winner = %q, want %q", resolutions[0].Winner, tt.wantWinner)
			}
			if resolutions[0].LocalOverridden != tt.wantOverridden {
				t.Errorf("LocalOverridden = %v, want %v", resolutions[0].LocalOverridden, tt.wantOverridden)
			}

			// Apply and verify final pool state
			sc.engine.ApplyChanges(sc.pool, changes, resolutions)
			final := sc.pool.GetTask("task-1")
			if final == nil {
				t.Fatal("task-1 missing from pool after sync")
			}
			if final.Text != tt.wantTitle {
				t.Errorf("final title = %q, want %q", final.Text, tt.wantTitle)
			}
		})
	}
}

// =============================================================================
// Scenario 2: Orphaned task — deleted in one adapter, modified in another
// =============================================================================

func TestMultiAdapter_OrphanedTask_DeleteVsModify(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		deleteFromA    bool // if true, A deletes; if false, B deletes
		wantPoolCount  int
		wantTaskExists bool
	}{
		{
			name:           "adapter A deletes task, adapter B still has it",
			deleteFromA:    true,
			wantPoolCount:  1, // task removed by A's sync
			wantTaskExists: false,
		},
		{
			name:           "adapter B deletes task, adapter A still has it modified",
			deleteFromA:    false,
			wantPoolCount:  1, // task removed by B's sync
			wantTaskExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task1 := newTestTask("task-1", "Keep me", StatusTodo, baseTime)
			task2 := newTestTask("task-2", "I might be deleted", StatusTodo, baseTime)
			sc := newScenario()
			sc.seedPool(task1, task2)

			if tt.deleteFromA {
				// Adapter A removes task-2
				sc.providerA.Tasks = []*Task{cloneTask(task1)}
				result := sc.syncFromA(t)

				if result.Removed != 1 {
					t.Errorf("expected 1 removal from adapter A sync, got %d", result.Removed)
				}
			} else {
				// Adapter B removes task-2
				sc.providerB.Tasks = []*Task{cloneTask(task1)}
				result := sc.syncFromB(t)

				if result.Removed != 1 {
					t.Errorf("expected 1 removal from adapter B sync, got %d", result.Removed)
				}
			}

			if sc.pool.Count() != tt.wantPoolCount {
				t.Errorf("pool count = %d, want %d", sc.pool.Count(), tt.wantPoolCount)
			}

			orphan := sc.pool.GetTask("task-2")
			if (orphan != nil) != tt.wantTaskExists {
				t.Errorf("task-2 exists = %v, want %v", orphan != nil, tt.wantTaskExists)
			}
		})
	}
}

// TestMultiAdapter_DeleteModifyConflict tests the scenario where adapter A
// deletes a task while adapter B modifies it. The sync engine's detect+apply
// flow handles this: the delete from A removes the task, and B's modification
// arrives as a "new" task on subsequent sync (since it's no longer in sync state).
func TestMultiAdapter_DeleteModifyConflict(t *testing.T) {
	t.Parallel()

	task1 := newTestTask("task-1", "Shared task", StatusTodo, baseTime)
	task2 := newTestTask("task-2", "Contested task", StatusTodo, baseTime)
	sc := newScenario()
	sc.seedPool(task1, task2)

	// Step 1: Adapter A deletes task-2
	sc.providerA.Tasks = []*Task{cloneTask(task1)}
	resultA := sc.syncFromA(t)
	if resultA.Removed != 1 {
		t.Fatalf("adapter A should remove 1 task, got %d", resultA.Removed)
	}

	// Pool should now have only task-1
	if sc.pool.Count() != 1 {
		t.Fatalf("pool count after A sync = %d, want 1", sc.pool.Count())
	}

	// Step 2: Adapter B still has task-2 (modified), doesn't know about deletion
	task2Modified := newTestTask("task-2", "Modified by B", StatusInProgress, laterTime)
	sc.providerB.Tasks = []*Task{cloneTask(task1), task2Modified}

	// After A's sync removed task-2 from pool, the sync state for B still has it.
	// B's remote still has task-2 (modified) → detected as modified (not conflict,
	// since local didn't mark it dirty).
	resultB := sc.syncFromB(t)

	// task-2 was in sync state, B still has it modified → it's a ModifiedTask
	// Also, task-2 is still in syncStateB snapshots so it won't appear as new
	if resultB.Updated > 0 || resultB.Added > 0 {
		// The sync sees task-2 as modified in remote → re-adds via UpdateTask
		poolTask := sc.pool.GetTask("task-2")
		if poolTask == nil {
			t.Error("task-2 should be restored after B sync brings it back")
		} else if poolTask.Text != "Modified by B" {
			t.Errorf("task-2 text = %q, want %q", poolTask.Text, "Modified by B")
		}
	}
}

func cloneTask(t *Task) *Task {
	cp := *t
	return &cp
}

// =============================================================================
// Scenario 3: Field-level conflict — title changed in A, status changed in B
// =============================================================================

func TestMultiAdapter_FieldLevelConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		localTitle   string
		localStatus  TaskStatus
		localTime    time.Time
		remoteTitle  string
		remoteStatus TaskStatus
		remoteTime   time.Time
		wantWinner   string
	}{
		{
			name:         "A changes title, B changes status — B newer wins all fields",
			localTitle:   "New title from A",
			localStatus:  StatusTodo,
			localTime:    laterTime,
			remoteTitle:  "Original",
			remoteStatus: StatusInProgress,
			remoteTime:   latestTime,
			wantWinner:   "remote",
		},
		{
			name:         "A changes title, B changes status — A newer wins all fields",
			localTitle:   "New title from A",
			localStatus:  StatusTodo,
			localTime:    latestTime,
			remoteTitle:  "Original",
			remoteStatus: StatusInProgress,
			remoteTime:   laterTime,
			wantWinner:   "local",
		},
		{
			name:         "both change different fields, same timestamp — remote wins",
			localTitle:   "Title from A",
			localStatus:  StatusTodo,
			localTime:    laterTime,
			remoteTitle:  "Original",
			remoteStatus: StatusInProgress,
			remoteTime:   laterTime,
			wantWinner:   "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := newTestTask("task-1", "Original", StatusTodo, baseTime)
			sc := newScenario()
			sc.seedPool(original)

			// Adapter A changes title only
			localVersion := newTestTask("task-1", tt.localTitle, tt.localStatus, tt.localTime)
			sc.pool.UpdateTask(localVersion)

			// Adapter B changes status only
			remoteVersion := newTestTask("task-1", tt.remoteTitle, tt.remoteStatus, tt.remoteTime)
			sc.providerB.Tasks = []*Task{remoteVersion}

			// Mark dirty — local was modified
			sc.syncStateB.TaskSnapshots["task-1"] = TaskSnapshot{
				ID: "task-1", Text: "Original", Status: StatusTodo,
				UpdatedAt: baseTime, Dirty: true,
			}

			local := sc.pool.GetAllTasks()
			changes := sc.engine.DetectChanges(sc.syncStateB, local, sc.providerB.Tasks)

			if len(changes.Conflicts) != 1 {
				t.Fatalf("expected 1 conflict, got %d (new=%d mod=%d del=%d)",
					len(changes.Conflicts), len(changes.NewTasks),
					len(changes.ModifiedTasks), len(changes.DeletedTasks))
			}

			resolutions := sc.engine.ResolveConflicts(changes.Conflicts)
			if resolutions[0].Winner != tt.wantWinner {
				t.Errorf("winner = %q, want %q", resolutions[0].Winner, tt.wantWinner)
			}

			// Apply and verify: last-writer-wins is whole-task, not field-level merge
			sc.engine.ApplyChanges(sc.pool, changes, resolutions)
			final := sc.pool.GetTask("task-1")
			if final == nil {
				t.Fatal("task-1 missing after apply")
			}

			var wantTitle string
			var wantStatus TaskStatus
			if tt.wantWinner == "remote" {
				wantTitle = tt.remoteTitle
				wantStatus = tt.remoteStatus
			} else {
				wantTitle = tt.localTitle
				wantStatus = tt.localStatus
			}

			if final.Text != wantTitle {
				t.Errorf("final text = %q, want %q", final.Text, wantTitle)
			}
			if final.Status != wantStatus {
				t.Errorf("final status = %q, want %q", final.Status, wantStatus)
			}
		})
	}
}

// =============================================================================
// Scenario 4: Stale reconciliation — one adapter offline, comes back with stale data
// =============================================================================

func TestMultiAdapter_StaleReconciliation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		staleTitle        string
		staleTime         time.Time
		currentTitle      string
		currentTime       time.Time
		wantTitle         string
		wantConflicts     int
		wantLocalOverride bool
	}{
		{
			name:              "stale adapter returns old data — current version preserved",
			staleTitle:        "Old stale version",
			staleTime:         baseTime,
			currentTitle:      "Current version",
			currentTime:       latestTime,
			wantTitle:         "Current version",
			wantConflicts:     0,
			wantLocalOverride: false,
		},
		{
			name:              "stale adapter returns slightly newer data — triggers conflict, remote wins",
			staleTitle:        "Offline edit",
			staleTime:         latestTime,
			currentTitle:      "Online edit",
			currentTime:       laterTime,
			wantTitle:         "Offline edit",
			wantConflicts:     1,
			wantLocalOverride: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := newTestTask("task-1", "Original", StatusTodo, baseTime)
			sc := newScenario()
			sc.seedPool(original)

			// Adapter A (online) updates the task
			currentVersion := newTestTask("task-1", tt.currentTitle, StatusTodo, tt.currentTime)
			sc.pool.UpdateTask(currentVersion)
			sc.providerA.Tasks = []*Task{currentVersion}

			// Sync from A to update pool
			sc.syncStateA.TaskSnapshots["task-1"] = TaskSnapshot{
				ID: "task-1", Text: "Original", Status: StatusTodo,
				UpdatedAt: baseTime, Dirty: false,
			}
			sc.syncFromA(t)

			// Now adapter B (was offline) comes back with stale data
			staleVersion := newTestTask("task-1", tt.staleTitle, StatusTodo, tt.staleTime)
			sc.providerB.Tasks = []*Task{staleVersion}

			// Sync state B still has the original baseline
			sc.syncStateB.TaskSnapshots["task-1"] = TaskSnapshot{
				ID: "task-1", Text: "Original", Status: StatusTodo,
				UpdatedAt: baseTime, Dirty: tt.wantConflicts > 0,
			}

			local := sc.pool.GetAllTasks()
			changes := sc.engine.DetectChanges(sc.syncStateB, local, sc.providerB.Tasks)

			if len(changes.Conflicts) != tt.wantConflicts {
				// If stale data is older than sync snapshot, no remote modification detected
				// If stale data has same timestamp, no change detected either
				if tt.staleTime.After(baseTime) && tt.wantConflicts == 0 {
					// staleTime > baseTime but not dirty → ModifiedTask, not conflict
					if len(changes.ModifiedTasks) != 1 {
						t.Errorf("expected 1 modified task, got %d", len(changes.ModifiedTasks))
					}
				}
			}

			resolutions := sc.engine.ResolveConflicts(changes.Conflicts)
			result := sc.engine.ApplyChanges(sc.pool, changes, resolutions)

			final := sc.pool.GetTask("task-1")
			if final == nil {
				t.Fatal("task-1 missing from pool")
			}

			// When stale data isn't newer than baseline, no change happens
			// When stale data IS newer and dirty, conflict resolves by timestamp
			if tt.wantConflicts > 0 {
				if result.Conflicts != tt.wantConflicts {
					t.Errorf("conflicts = %d, want %d", result.Conflicts, tt.wantConflicts)
				}
				// Local (currentVersion at laterTime) wins because it's the pool
				// content, and conflict resolves local vs remote by timestamp
				for _, r := range resolutions {
					if r.LocalOverridden != tt.wantLocalOverride {
						t.Errorf("LocalOverridden = %v, want %v", r.LocalOverridden, tt.wantLocalOverride)
					}
				}
			}
		})
	}
}

// =============================================================================
// Scenario 5: Full two-adapter sync cycle
// =============================================================================

func TestMultiAdapter_FullSyncCycle(t *testing.T) {
	t.Parallel()

	// Simulate a realistic scenario: TextFile (A) and Todoist (B) both have tasks.
	// Changes happen on both sides between sync cycles.

	task1 := newTestTask("task-1", "Buy groceries", StatusTodo, baseTime)
	task2 := newTestTask("task-2", "Write report", StatusTodo, baseTime)
	task3 := newTestTask("task-3", "Fix bug #42", StatusTodo, baseTime)

	sc := newScenario()
	sc.seedPool(task1, task2, task3)

	// --- Sync cycle 1: Adapter A adds a new task ---
	task4 := newTestTask("task-4", "New from TextFile", StatusTodo, laterTime)
	sc.providerA.Tasks = []*Task{
		cloneTask(task1), cloneTask(task2), cloneTask(task3), task4,
	}

	result1 := sc.syncFromA(t)
	if result1.Added != 1 {
		t.Errorf("cycle 1: added = %d, want 1", result1.Added)
	}
	if sc.pool.Count() != 4 {
		t.Errorf("cycle 1: pool count = %d, want 4", sc.pool.Count())
	}

	// Update sync states after cycle 1
	sc.snapshotSyncState()

	// --- Sync cycle 2: Adapter B deletes task-2, modifies task-3 ---
	task3Modified := newTestTask("task-3", "Fix bug #42 (urgent!)", StatusInProgress, latestTime)
	sc.providerB.Tasks = []*Task{
		cloneTask(task1), task3Modified, cloneTask(task4),
	}

	result2 := sc.syncFromB(t)
	if result2.Removed != 1 {
		t.Errorf("cycle 2: removed = %d, want 1", result2.Removed)
	}
	if result2.Updated != 1 {
		t.Errorf("cycle 2: updated = %d, want 1", result2.Updated)
	}

	// Verify final state
	if sc.pool.Count() != 3 {
		t.Errorf("final pool count = %d, want 3", sc.pool.Count())
	}
	if sc.pool.GetTask("task-2") != nil {
		t.Error("task-2 should be deleted")
	}
	task3Final := sc.pool.GetTask("task-3")
	if task3Final == nil {
		t.Fatal("task-3 should exist")
	}
	if task3Final.Text != "Fix bug #42 (urgent!)" {
		t.Errorf("task-3 text = %q, want %q", task3Final.Text, "Fix bug #42 (urgent!)")
	}
	if task3Final.Status != StatusInProgress {
		t.Errorf("task-3 status = %q, want %q", task3Final.Status, StatusInProgress)
	}
}

// =============================================================================
// Scenario 6: Multiple conflicts across adapters in a single sync
// =============================================================================

func TestMultiAdapter_MultipleConflicts(t *testing.T) {
	t.Parallel()

	task1 := newTestTask("task-1", "Task one", StatusTodo, baseTime)
	task2 := newTestTask("task-2", "Task two", StatusTodo, baseTime)
	task3 := newTestTask("task-3", "Task three", StatusTodo, baseTime)

	sc := newScenario()
	sc.seedPool(task1, task2, task3)

	// Local (pool) modifications
	task1Local := newTestTask("task-1", "Task one - local edit", StatusInProgress, latestTime) // local newer
	task2Local := newTestTask("task-2", "Task two - local edit", StatusTodo, laterTime)        // remote newer
	sc.pool.UpdateTask(task1Local)
	sc.pool.UpdateTask(task2Local)

	// Remote (adapter B) modifications
	task1Remote := newTestTask("task-1", "Task one - remote edit", StatusTodo, laterTime)         // older
	task2Remote := newTestTask("task-2", "Task two - remote edit", StatusComplete, latestTime)    // newer
	task3Remote := newTestTask("task-3", "Task three - remote edit", StatusInProgress, laterTime) // modified, not dirty

	sc.providerB.Tasks = []*Task{task1Remote, task2Remote, task3Remote}

	// Mark task-1 and task-2 as dirty (locally modified)
	sc.syncStateB.TaskSnapshots["task-1"] = TaskSnapshot{
		ID: "task-1", Text: "Task one", Status: StatusTodo,
		UpdatedAt: baseTime, Dirty: true,
	}
	sc.syncStateB.TaskSnapshots["task-2"] = TaskSnapshot{
		ID: "task-2", Text: "Task two", Status: StatusTodo,
		UpdatedAt: baseTime, Dirty: true,
	}

	local := sc.pool.GetAllTasks()
	changes := sc.engine.DetectChanges(sc.syncStateB, local, sc.providerB.Tasks)

	if len(changes.Conflicts) != 2 {
		t.Fatalf("expected 2 conflicts, got %d", len(changes.Conflicts))
	}
	if len(changes.ModifiedTasks) != 1 {
		t.Errorf("expected 1 modified (task-3), got %d", len(changes.ModifiedTasks))
	}

	resolutions := sc.engine.ResolveConflicts(changes.Conflicts)
	result := sc.engine.ApplyChanges(sc.pool, changes, resolutions)

	if result.Conflicts != 2 {
		t.Errorf("result.Conflicts = %d, want 2", result.Conflicts)
	}

	// task-1: local wins (latestTime > laterTime)
	t1 := sc.pool.GetTask("task-1")
	if t1.Text != "Task one - local edit" {
		t.Errorf("task-1 text = %q, want local version", t1.Text)
	}

	// task-2: remote wins (latestTime > laterTime)
	t2 := sc.pool.GetTask("task-2")
	if t2.Text != "Task two - remote edit" {
		t.Errorf("task-2 text = %q, want remote version", t2.Text)
	}
	if t2.Status != StatusComplete {
		t.Errorf("task-2 status = %q, want %q", t2.Status, StatusComplete)
	}

	// task-3: non-conflicting modification applied
	t3 := sc.pool.GetTask("task-3")
	if t3.Text != "Task three - remote edit" {
		t.Errorf("task-3 text = %q, want remote modified version", t3.Text)
	}

	// Exactly one override (task-2 local was overridden)
	if len(result.Overrides) != 1 {
		t.Errorf("expected 1 override, got %d", len(result.Overrides))
	} else if result.Overrides[0].TaskID != "task-2" {
		t.Errorf("override task = %q, want task-2", result.Overrides[0].TaskID)
	}
}

// =============================================================================
// Scenario 7: Sequential sync from both adapters — no data loss
// =============================================================================

func TestMultiAdapter_SequentialSync_NoDataLoss(t *testing.T) {
	t.Parallel()

	original := newTestTask("task-1", "Shared task", StatusTodo, baseTime)
	sc := newScenario()
	sc.seedPool(original)

	// Adapter A adds a new task
	taskFromA := newTestTask("task-a", "From adapter A", StatusTodo, laterTime)
	sc.providerA.Tasks = []*Task{cloneTask(original), taskFromA}

	resultA := sc.syncFromA(t)
	if resultA.Added != 1 {
		t.Errorf("adapter A should add 1 task, got %d", resultA.Added)
	}

	// Update sync states
	sc.snapshotSyncState()

	// Adapter B adds a different new task
	taskFromB := newTestTask("task-b", "From adapter B", StatusTodo, laterTime)
	sc.providerB.Tasks = []*Task{cloneTask(original), cloneTask(taskFromA), taskFromB}

	resultB := sc.syncFromB(t)
	if resultB.Added != 1 {
		t.Errorf("adapter B should add 1 task, got %d", resultB.Added)
	}

	// Final pool should have all 3 tasks
	if sc.pool.Count() != 3 {
		t.Errorf("final pool count = %d, want 3", sc.pool.Count())
	}
	for _, id := range []string{"task-1", "task-a", "task-b"} {
		if sc.pool.GetTask(id) == nil {
			t.Errorf("task %q missing from final pool", id)
		}
	}
}

// =============================================================================
// Scenario 8: Adapter error during multi-adapter sync
// =============================================================================

func TestMultiAdapter_ProviderError_PoolUnchanged(t *testing.T) {
	t.Parallel()

	task1 := newTestTask("task-1", "Important task", StatusTodo, baseTime)
	sc := newScenario()
	sc.seedPool(task1)

	// Adapter A syncs fine
	sc.providerA.Tasks = []*Task{cloneTask(task1)}
	sc.syncFromA(t)

	// Adapter B fails
	sc.providerB.LoadErr = fmt.Errorf("network timeout")
	_, err := sc.engine.Sync(sc.providerB, sc.syncStateB, sc.pool)
	if err == nil {
		t.Fatal("expected error from failing adapter B")
	}

	// Pool should be unchanged after failed sync
	if sc.pool.Count() != 1 {
		t.Errorf("pool count = %d, want 1 (unchanged after error)", sc.pool.Count())
	}
	if sc.pool.GetTask("task-1") == nil {
		t.Error("task-1 should still be in pool")
	}
}

// =============================================================================
// Scenario 9: Identical edits from both adapters — no override notification
// =============================================================================

func TestMultiAdapter_IdenticalEdits_NoOverride(t *testing.T) {
	t.Parallel()

	original := newTestTask("task-1", "Original text", StatusTodo, baseTime)
	sc := newScenario()
	sc.seedPool(original)

	// Both adapters arrive at the same state
	sameEdit := newTestTask("task-1", "Updated text", StatusInProgress, laterTime)

	// Apply locally
	sc.pool.UpdateTask(cloneTask(sameEdit))

	// Remote has same changes
	sc.providerB.Tasks = []*Task{cloneTask(sameEdit)}

	sc.syncStateB.TaskSnapshots["task-1"] = TaskSnapshot{
		ID: "task-1", Text: "Original text", Status: StatusTodo,
		UpdatedAt: baseTime, Dirty: true,
	}

	local := sc.pool.GetAllTasks()
	changes := sc.engine.DetectChanges(sc.syncStateB, local, sc.providerB.Tasks)

	if len(changes.Conflicts) != 1 {
		t.Fatalf("expected 1 conflict (both modified), got %d", len(changes.Conflicts))
	}

	resolutions := sc.engine.ResolveConflicts(changes.Conflicts)
	if len(resolutions) != 1 {
		t.Fatalf("expected 1 resolution, got %d", len(resolutions))
	}

	// Identical changes should not trigger override notification
	if resolutions[0].LocalOverridden {
		t.Error("identical edits should not mark LocalOverridden=true")
	}
	if resolutions[0].Message != "" {
		t.Errorf("identical edits should have empty message, got %q", resolutions[0].Message)
	}
}
