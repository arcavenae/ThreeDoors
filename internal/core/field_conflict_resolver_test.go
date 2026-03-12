package core

import (
	"testing"
	"time"
)

// =============================================================================
// DefaultConflictResolver Tests
// =============================================================================

func TestDefaultConflictResolver_MetadataFields_RemoteWins(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	metadataFields := []string{"text", "status", "context", "notes", "blocker", "depends_on", "completed_at", "defer_until"}
	for _, field := range metadataFields {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			winner := resolver.ResolveField(field)
			if winner != "remote" {
				t.Errorf("ResolveField(%q) = %q, want %q", field, winner, "remote")
			}
			cat := resolver.Category(field)
			if cat != MetadataField {
				t.Errorf("Category(%q) = %d, want MetadataField(%d)", field, cat, MetadataField)
			}
		})
	}
}

func TestDefaultConflictResolver_ThreeDoorsFields_LocalWins(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	threeDoorsFields := []string{"effort", "type", "location", "dev_dispatch"}
	for _, field := range threeDoorsFields {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			winner := resolver.ResolveField(field)
			if winner != "local" {
				t.Errorf("ResolveField(%q) = %q, want %q", field, winner, "local")
			}
			cat := resolver.Category(field)
			if cat != ThreeDoorsField {
				t.Errorf("Category(%q) = %d, want ThreeDoorsField(%d)", field, cat, ThreeDoorsField)
			}
		})
	}
}

func TestDefaultConflictResolver_UnknownField_DefaultsToRemote(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	winner := resolver.ResolveField("unknown_field")
	if winner != "remote" {
		t.Errorf("ResolveField(unknown) = %q, want %q", winner, "remote")
	}
	cat := resolver.Category("unknown_field")
	if cat != MetadataField {
		t.Errorf("Category(unknown) = %d, want MetadataField(%d)", cat, MetadataField)
	}
}

// =============================================================================
// ResolveTaskConflict Tests
// =============================================================================

func TestResolveTaskConflict_MetadataConflict_RemoteWins(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Local text", StatusInProgress, laterTime)
	remote := newTestTask("aaa", "Remote text", StatusTodo, baseTime)

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// Text: remote wins (metadata)
	if merged.Text != "Remote text" {
		t.Errorf("merged.Text = %q, want %q", merged.Text, "Remote text")
	}
	// Status: remote wins (metadata)
	if merged.Status != StatusTodo {
		t.Errorf("merged.Status = %q, want %q", merged.Status, StatusTodo)
	}

	if len(fieldRes) != 2 {
		t.Fatalf("len(fieldRes) = %d, want 2 (text + status)", len(fieldRes))
	}

	for _, fr := range fieldRes {
		if fr.Winner != "remote" {
			t.Errorf("field %q winner = %q, want %q", fr.Field, fr.Winner, "remote")
		}
	}
}

func TestResolveTaskConflict_ThreeDoorsFieldConflict_LocalWins(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Same text", StatusTodo, laterTime)
	local.Effort = EffortQuickWin
	local.Type = TypeCreative
	local.Location = LocationHome

	remote := newTestTask("aaa", "Same text", StatusTodo, baseTime)
	remote.Effort = EffortDeepWork
	remote.Type = TypeTechnical
	remote.Location = LocationWork

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// ThreeDoors fields: local wins
	if merged.Effort != EffortQuickWin {
		t.Errorf("merged.Effort = %q, want %q", merged.Effort, EffortQuickWin)
	}
	if merged.Type != TypeCreative {
		t.Errorf("merged.Type = %q, want %q", merged.Type, TypeCreative)
	}
	if merged.Location != LocationHome {
		t.Errorf("merged.Location = %q, want %q", merged.Location, LocationHome)
	}

	if len(fieldRes) != 3 {
		t.Fatalf("len(fieldRes) = %d, want 3 (effort + type + location)", len(fieldRes))
	}

	for _, fr := range fieldRes {
		if fr.Winner != "local" {
			t.Errorf("field %q winner = %q, want %q", fr.Field, fr.Winner, "local")
		}
	}
}

func TestResolveTaskConflict_MixedConflict(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Local text", StatusTodo, laterTime)
	local.Effort = EffortQuickWin

	remote := newTestTask("aaa", "Remote text", StatusTodo, baseTime)
	remote.Effort = EffortDeepWork

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// Text: remote wins (metadata), Effort: local wins (ThreeDoors)
	if merged.Text != "Remote text" {
		t.Errorf("merged.Text = %q, want %q (remote wins metadata)", merged.Text, "Remote text")
	}
	if merged.Effort != EffortQuickWin {
		t.Errorf("merged.Effort = %q, want %q (local wins ThreeDoors)", merged.Effort, EffortQuickWin)
	}

	if len(fieldRes) != 2 {
		t.Fatalf("len(fieldRes) = %d, want 2", len(fieldRes))
	}

	for _, fr := range fieldRes {
		switch fr.Field {
		case "text":
			if fr.Winner != "remote" {
				t.Errorf("text winner = %q, want remote", fr.Winner)
			}
		case "effort":
			if fr.Winner != "local" {
				t.Errorf("effort winner = %q, want local", fr.Winner)
			}
		default:
			t.Errorf("unexpected field resolution: %q", fr.Field)
		}
	}
}

func TestResolveTaskConflict_IdenticalTasks_NoFieldResolutions(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Same text", StatusTodo, laterTime)
	remote := newTestTask("aaa", "Same text", StatusTodo, baseTime)

	_, fieldRes := ResolveTaskConflict(resolver, local, remote)

	if len(fieldRes) != 0 {
		t.Errorf("len(fieldRes) = %d, want 0 (identical tasks)", len(fieldRes))
	}
}

func TestResolveTaskConflict_PreservesNewerUpdatedAt(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Local text", StatusTodo, latestTime)
	remote := newTestTask("aaa", "Remote text", StatusTodo, laterTime)

	merged, _ := ResolveTaskConflict(resolver, local, remote)

	if !merged.UpdatedAt.Equal(latestTime) {
		t.Errorf("merged.UpdatedAt = %v, want %v (should preserve newer)", merged.UpdatedAt, latestTime)
	}
}

func TestResolveTaskConflict_ContextConflict(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Task", StatusTodo, laterTime)
	local.Context = "local-context"
	remote := newTestTask("aaa", "Task", StatusTodo, baseTime)
	remote.Context = "remote-context"

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// Context is metadata → remote wins
	if merged.Context != "remote-context" {
		t.Errorf("merged.Context = %q, want %q", merged.Context, "remote-context")
	}
	if len(fieldRes) != 1 {
		t.Fatalf("len(fieldRes) = %d, want 1", len(fieldRes))
	}
	if fieldRes[0].Field != "context" || fieldRes[0].Winner != "remote" {
		t.Errorf("field resolution = %+v, want context/remote", fieldRes[0])
	}
}

func TestResolveTaskConflict_BlockerConflict(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Task", StatusTodo, laterTime)
	local.Blocker = "local blocker"
	remote := newTestTask("aaa", "Task", StatusTodo, baseTime)
	remote.Blocker = "remote blocker"

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// Blocker is metadata → remote wins
	if merged.Blocker != "remote blocker" {
		t.Errorf("merged.Blocker = %q, want %q", merged.Blocker, "remote blocker")
	}
	if len(fieldRes) != 1 {
		t.Fatalf("len(fieldRes) = %d, want 1", len(fieldRes))
	}
}

func TestResolveTaskConflict_DeferUntilConflict(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	localDefer := laterTime
	remoteDefer := latestTime

	local := newTestTask("aaa", "Task", StatusTodo, laterTime)
	local.DeferUntil = &localDefer
	remote := newTestTask("aaa", "Task", StatusTodo, baseTime)
	remote.DeferUntil = &remoteDefer

	merged, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// DeferUntil is metadata → remote wins
	if merged.DeferUntil == nil || !merged.DeferUntil.Equal(remoteDefer) {
		t.Errorf("merged.DeferUntil = %v, want %v", merged.DeferUntil, remoteDefer)
	}
	if len(fieldRes) != 1 {
		t.Fatalf("len(fieldRes) = %d, want 1", len(fieldRes))
	}
}

func TestResolveTaskConflict_FieldResolutionValues(t *testing.T) {
	t.Parallel()
	resolver := NewDefaultConflictResolver()

	local := newTestTask("aaa", "Local text", StatusInProgress, laterTime)
	remote := newTestTask("aaa", "Remote text", StatusTodo, baseTime)

	_, fieldRes := ResolveTaskConflict(resolver, local, remote)

	// Verify field resolution records contain correct values
	for _, fr := range fieldRes {
		switch fr.Field {
		case "text":
			if fr.LocalValue != "Local text" {
				t.Errorf("text.LocalValue = %q, want %q", fr.LocalValue, "Local text")
			}
			if fr.RemoteValue != "Remote text" {
				t.Errorf("text.RemoteValue = %q, want %q", fr.RemoteValue, "Remote text")
			}
		case "status":
			if fr.LocalValue != string(StatusInProgress) {
				t.Errorf("status.LocalValue = %q, want %q", fr.LocalValue, StatusInProgress)
			}
			if fr.RemoteValue != string(StatusTodo) {
				t.Errorf("status.RemoteValue = %q, want %q", fr.RemoteValue, StatusTodo)
			}
		}
	}
}

// =============================================================================
// SyncEngine.ResolveFieldConflicts Tests
// =============================================================================

func TestResolveFieldConflicts_MixedFields(t *testing.T) {
	t.Parallel()
	engine := NewSyncEngineWithResolver(NewDefaultConflictResolver())

	local := newTestTask("aaa", "Local text", StatusTodo, laterTime)
	local.Effort = EffortQuickWin
	remote := newTestTask("aaa", "Remote text", StatusTodo, baseTime)
	remote.Effort = EffortDeepWork

	conflicts := []Conflict{{LocalTask: local, RemoteTask: remote}}
	resolutions := engine.ResolveFieldConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("len(resolutions) = %d, want 1", len(resolutions))
	}

	r := resolutions[0]
	if r.Winner != "field-level" {
		t.Errorf("Winner = %q, want %q", r.Winner, "field-level")
	}
	if len(r.FieldResolutions) != 2 {
		t.Fatalf("len(FieldResolutions) = %d, want 2", len(r.FieldResolutions))
	}

	// text → remote wins, effort → local wins
	for _, fr := range r.FieldResolutions {
		switch fr.Field {
		case "text":
			if fr.Winner != "remote" {
				t.Errorf("text.Winner = %q, want remote", fr.Winner)
			}
		case "effort":
			if fr.Winner != "local" {
				t.Errorf("effort.Winner = %q, want local", fr.Winner)
			}
		}
	}

	// Verify merged task
	if r.WinningTask.Text != "Remote text" {
		t.Errorf("WinningTask.Text = %q, want %q", r.WinningTask.Text, "Remote text")
	}
	if r.WinningTask.Effort != EffortQuickWin {
		t.Errorf("WinningTask.Effort = %q, want %q", r.WinningTask.Effort, EffortQuickWin)
	}
}

func TestResolveFieldConflicts_IdenticalTasks(t *testing.T) {
	t.Parallel()
	engine := NewSyncEngineWithResolver(NewDefaultConflictResolver())

	local := newTestTask("aaa", "Same", StatusTodo, laterTime)
	remote := newTestTask("aaa", "Same", StatusTodo, baseTime)

	conflicts := []Conflict{{LocalTask: local, RemoteTask: remote}}
	resolutions := engine.ResolveFieldConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("len(resolutions) = %d, want 1", len(resolutions))
	}
	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false for identical tasks")
	}
	if len(resolutions[0].FieldResolutions) != 0 {
		t.Errorf("FieldResolutions should be empty for identical tasks, got %d", len(resolutions[0].FieldResolutions))
	}
}

func TestResolveFieldConflicts_DefaultResolver(t *testing.T) {
	t.Parallel()
	// Engine without explicit resolver should use default
	engine := NewSyncEngineWithResolver(nil)

	local := newTestTask("aaa", "Local", StatusTodo, laterTime)
	remote := newTestTask("aaa", "Remote", StatusTodo, baseTime)

	conflicts := []Conflict{{LocalTask: local, RemoteTask: remote}}
	resolutions := engine.ResolveFieldConflicts(conflicts)

	if len(resolutions) != 1 {
		t.Fatalf("len(resolutions) = %d, want 1", len(resolutions))
	}
	// Should still resolve correctly with default resolver
	if resolutions[0].WinningTask.Text != "Remote" {
		t.Errorf("WinningTask.Text = %q, want %q", resolutions[0].WinningTask.Text, "Remote")
	}
}

func TestResolveFieldConflicts_LocalOverridden_WhenMetadataChanged(t *testing.T) {
	t.Parallel()
	engine := NewSyncEngineWithResolver(NewDefaultConflictResolver())

	local := newTestTask("aaa", "Local text", StatusTodo, laterTime)
	remote := newTestTask("aaa", "Remote text", StatusTodo, baseTime)

	conflicts := []Conflict{{LocalTask: local, RemoteTask: remote}}
	resolutions := engine.ResolveFieldConflicts(conflicts)

	if !resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be true when remote wins on a metadata field")
	}
}

func TestResolveFieldConflicts_NotOverridden_WhenOnlyThreeDoorsFieldsChanged(t *testing.T) {
	t.Parallel()
	engine := NewSyncEngineWithResolver(NewDefaultConflictResolver())

	local := newTestTask("aaa", "Same text", StatusTodo, laterTime)
	local.Effort = EffortQuickWin
	remote := newTestTask("aaa", "Same text", StatusTodo, baseTime)
	remote.Effort = EffortDeepWork

	conflicts := []Conflict{{LocalTask: local, RemoteTask: remote}}
	resolutions := engine.ResolveFieldConflicts(conflicts)

	if resolutions[0].LocalOverridden {
		t.Error("LocalOverridden should be false when only ThreeDoors fields differ (local wins)")
	}
}

// =============================================================================
// Orphaned Task Tests
// =============================================================================

func TestApplyChangesWithOrphans_MarksOrphaned(t *testing.T) {
	t.Parallel()
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)

	pool := poolFromTasks(taskA, taskB)
	changes := ChangeSet{DeletedTasks: []string{"bbb"}}

	engine := NewSyncEngine()
	result := engine.ApplyChangesWithOrphans(pool, changes, nil)

	// Task should still exist but be marked orphaned
	orphaned := pool.GetTask("bbb")
	if orphaned == nil {
		t.Fatal("Task 'bbb' should still exist in pool (orphaned, not deleted)")
	}
	if !orphaned.Orphaned {
		t.Error("Task 'bbb' should have Orphaned=true")
	}
	if orphaned.OrphanedAt == nil {
		t.Error("Task 'bbb' should have OrphanedAt set")
	}
	if result.Removed != 1 {
		t.Errorf("Result.Removed = %d, want 1", result.Removed)
	}
	// Pool should still have both tasks
	if pool.Count() != 2 {
		t.Errorf("Pool count = %d, want 2 (orphaned tasks remain)", pool.Count())
	}
}

func TestApplyChangesWithOrphans_NonExistentTask(t *testing.T) {
	t.Parallel()
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)

	pool := poolFromTasks(taskA)
	changes := ChangeSet{DeletedTasks: []string{"nonexistent"}}

	engine := NewSyncEngine()
	result := engine.ApplyChangesWithOrphans(pool, changes, nil)

	// Should still count as removed, but no crash
	if result.Removed != 1 {
		t.Errorf("Result.Removed = %d, want 1", result.Removed)
	}
	if pool.Count() != 1 {
		t.Errorf("Pool count = %d, want 1", pool.Count())
	}
}

func TestApplyChangesWithOrphans_SummaryText(t *testing.T) {
	t.Parallel()
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskNew := newTestTask("bbb", "Task B", StatusTodo, laterTime)

	pool := poolFromTasks(taskA)
	changes := ChangeSet{
		NewTasks:     []*Task{taskNew},
		DeletedTasks: []string{"aaa"},
	}

	engine := NewSyncEngine()
	result := engine.ApplyChangesWithOrphans(pool, changes, nil)

	if result.Summary == "" {
		t.Error("Summary should not be empty")
	}
	// Summary should say "orphaned" not "removed"
	if result.Summary != "Synced: 1 new, 0 updated, 1 orphaned" {
		t.Errorf("Summary = %q, unexpected wording", result.Summary)
	}
}

func TestApplyChangesWithOrphans_MixedWithConflicts(t *testing.T) {
	t.Parallel()
	taskA := newTestTask("aaa", "Task A", StatusTodo, baseTime)
	taskB := newTestTask("bbb", "Task B", StatusTodo, baseTime)
	taskC := newTestTask("ccc", "Task C", StatusTodo, baseTime)
	taskNew := newTestTask("ddd", "New", StatusTodo, laterTime)
	taskAResolved := newTestTask("aaa", "Task A resolved", StatusTodo, laterTime)

	pool := poolFromTasks(taskA, taskB, taskC)
	changes := ChangeSet{
		NewTasks:     []*Task{taskNew},
		DeletedTasks: []string{"ccc"},
	}
	resolutions := []Resolution{
		{
			TaskID:          "aaa",
			Winner:          "remote",
			WinningTask:     taskAResolved,
			LocalOverridden: true,
			Message:         "overridden",
		},
	}

	engine := NewSyncEngine()
	result := engine.ApplyChangesWithOrphans(pool, changes, resolutions)

	if result.Added != 1 {
		t.Errorf("Added = %d, want 1", result.Added)
	}
	if result.Removed != 1 {
		t.Errorf("Removed = %d, want 1", result.Removed)
	}
	if result.Conflicts != 1 {
		t.Errorf("Conflicts = %d, want 1", result.Conflicts)
	}

	orphaned := pool.GetTask("ccc")
	if orphaned == nil || !orphaned.Orphaned {
		t.Error("Task 'ccc' should be orphaned")
	}

	resolved := pool.GetTask("aaa")
	if resolved.Text != "Task A resolved" {
		t.Errorf("Task A text = %q, want %q", resolved.Text, "Task A resolved")
	}

	if pool.Count() != 4 {
		t.Errorf("Pool count = %d, want 4", pool.Count())
	}
}

// =============================================================================
// Orphaned Task Struct Tests
// =============================================================================

func TestTask_OrphanedFields(t *testing.T) {
	t.Parallel()
	task := newTestTask("aaa", "Task A", StatusTodo, baseTime)

	// Initially not orphaned
	if task.Orphaned {
		t.Error("New task should not be orphaned")
	}
	if task.OrphanedAt != nil {
		t.Error("New task should have nil OrphanedAt")
	}

	// Mark as orphaned
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now

	if !task.Orphaned {
		t.Error("Task should be orphaned after marking")
	}
	if task.OrphanedAt == nil {
		t.Error("OrphanedAt should be set")
	}
}
