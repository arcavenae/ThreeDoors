package core

import (
	"testing"
	"time"
)

func TestTaskPool_GetAvailableForDoors_ExcludesOrphaned(t *testing.T) {
	t.Parallel()

	t.Run("orphaned tasks excluded from doors", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		normal := NewTask("Normal task")
		orphaned := NewTask("Orphaned task")
		now := time.Now().UTC()
		orphaned.Orphaned = true
		orphaned.OrphanedAt = &now
		pool.AddTask(normal)
		pool.AddTask(orphaned)

		available := pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available task, got %d", len(available))
		}
		if available[0].ID != normal.ID {
			t.Errorf("expected normal task, got %s", available[0].ID)
		}
	})

	t.Run("orphaned tasks excluded even in fallback with few tasks", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		orphaned := NewTask("Orphaned task")
		now := time.Now().UTC()
		orphaned.Orphaned = true
		orphaned.OrphanedAt = &now
		pool.AddTask(orphaned)

		// Only orphaned tasks — should return empty even in fallback path
		available := pool.GetAvailableForDoors()
		if len(available) != 0 {
			t.Errorf("expected 0 available tasks (all orphaned), got %d", len(available))
		}
	})

	t.Run("mix of orphaned and normal in fallback", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		normal := NewTask("Normal task")
		orphaned := NewTask("Orphaned task")
		now := time.Now().UTC()
		orphaned.Orphaned = true
		orphaned.OrphanedAt = &now
		pool.AddTask(normal)
		pool.AddTask(orphaned)
		// Mark normal as recently shown — triggers fallback
		pool.MarkRecentlyShown(normal.ID)

		available := pool.GetAvailableForDoors()
		// Fallback: < 3 non-recent, includes recently shown but still excludes orphaned
		if len(available) != 1 {
			t.Fatalf("expected 1 available task in fallback, got %d", len(available))
		}
		if available[0].ID != normal.ID {
			t.Errorf("expected normal task in fallback, got %s", available[0].ID)
		}
	})
}

func TestTaskPool_GetOrphanedTasks(t *testing.T) {
	t.Parallel()

	t.Run("returns only orphaned tasks", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		normal := NewTask("Normal")
		orphaned := NewTask("Orphaned")
		now := time.Now().UTC()
		orphaned.Orphaned = true
		orphaned.OrphanedAt = &now
		pool.AddTask(normal)
		pool.AddTask(orphaned)

		result := pool.GetOrphanedTasks()
		if len(result) != 1 {
			t.Fatalf("expected 1 orphaned task, got %d", len(result))
		}
		if result[0].ID != orphaned.ID {
			t.Errorf("expected orphaned task, got %s", result[0].ID)
		}
	})

	t.Run("returns empty when no orphaned tasks", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		pool.AddTask(NewTask("Normal"))

		result := pool.GetOrphanedTasks()
		if len(result) != 0 {
			t.Errorf("expected 0 orphaned tasks, got %d", len(result))
		}
	})
}

func TestTaskPool_KeepOrphanedTask(t *testing.T) {
	t.Parallel()

	t.Run("converts orphaned task to local", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		task := NewTask("Orphaned task")
		now := time.Now().UTC()
		task.Orphaned = true
		task.OrphanedAt = &now
		task.SourceProvider = "todoist"
		task.SourceRefs = []SourceRef{{Provider: "todoist", NativeID: "abc123"}}
		pool.AddTask(task)

		kept := pool.KeepOrphanedTask(task.ID)
		if kept == nil {
			t.Fatal("expected kept task, got nil")
		}
		if kept.Orphaned {
			t.Error("task should no longer be orphaned")
		}
		if kept.OrphanedAt != nil {
			t.Error("OrphanedAt should be nil")
		}
		if kept.SourceProvider != "" {
			t.Errorf("SourceProvider should be empty, got %q", kept.SourceProvider)
		}
		if len(kept.SourceRefs) != 0 {
			t.Errorf("SourceRefs should be empty, got %d", len(kept.SourceRefs))
		}
	})

	t.Run("returns nil for non-orphaned task", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		task := NewTask("Normal task")
		pool.AddTask(task)

		kept := pool.KeepOrphanedTask(task.ID)
		if kept != nil {
			t.Error("expected nil for non-orphaned task")
		}
	})

	t.Run("returns nil for missing task", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()

		kept := pool.KeepOrphanedTask("nonexistent")
		if kept != nil {
			t.Error("expected nil for missing task")
		}
	})

	t.Run("kept task reappears in door selection", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		task := NewTask("Was orphaned")
		now := time.Now().UTC()
		task.Orphaned = true
		task.OrphanedAt = &now
		pool.AddTask(task)

		// Verify orphaned task excluded from doors
		available := pool.GetAvailableForDoors()
		if len(available) != 0 {
			t.Fatalf("expected 0 available before keep, got %d", len(available))
		}

		pool.KeepOrphanedTask(task.ID)

		// After keep, task should appear in door selection
		available = pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available after keep, got %d", len(available))
		}
		if available[0].ID != task.ID {
			t.Errorf("expected kept task, got %s", available[0].ID)
		}
	})
}

func TestTaskPool_DeleteOrphanedTask(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("Orphaned task")
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now
	pool.AddTask(task)

	pool.RemoveTask(task.ID)

	if pool.GetTask(task.ID) != nil {
		t.Error("expected orphaned task to be removed")
	}
	if pool.Count() != 0 {
		t.Errorf("expected 0 tasks, got %d", pool.Count())
	}
}

func TestSyncEngine_ApplyChangesWithOrphans_MarksOrphaned(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("Remote task")
	task.ID = "remote-1"
	pool.AddTask(task)

	engine := NewSyncEngine()
	changes := ChangeSet{
		DeletedTasks: []string{"remote-1"},
	}

	result := engine.ApplyChangesWithOrphans(pool, changes, nil)

	// Task should still exist but be orphaned
	orphaned := pool.GetTask("remote-1")
	if orphaned == nil {
		t.Fatal("task should still exist (orphaned, not deleted)")
	}
	if !orphaned.Orphaned {
		t.Error("task should be marked as orphaned")
	}
	if orphaned.OrphanedAt == nil {
		t.Error("OrphanedAt should be set")
	}

	// Should not appear in doors
	available := pool.GetAvailableForDoors()
	if len(available) != 0 {
		t.Errorf("expected 0 available tasks (orphaned excluded), got %d", len(available))
	}

	if result.Removed != 1 {
		t.Errorf("expected 1 removal counted, got %d", result.Removed)
	}
}
