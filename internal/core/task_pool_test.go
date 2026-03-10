package core

import "testing"

func TestTaskPool_AddAndGet(t *testing.T) {
	pool := NewTaskPool()
	task := NewTask("Test task")
	pool.AddTask(task)

	got := pool.GetTask(task.ID)
	if got == nil {
		t.Fatal("Expected to get task back")
		return
	}
	if got.Text != task.Text {
		t.Errorf("Expected %q, got %q", task.Text, got.Text)
	}
}

func TestTaskPool_RemoveTask(t *testing.T) {
	pool := NewTaskPool()
	task := NewTask("Test")
	pool.AddTask(task)
	pool.RemoveTask(task.ID)

	if pool.GetTask(task.ID) != nil {
		t.Error("Expected task to be removed")
	}
	if pool.Count() != 0 {
		t.Errorf("Expected 0 tasks, got %d", pool.Count())
	}
}

func TestTaskPool_GetTasksByStatus(t *testing.T) {
	pool := NewTaskPool()
	t1 := NewTask("Todo task")
	t2 := NewTask("Blocked task")
	_ = t2.UpdateStatus(StatusBlocked)
	pool.AddTask(t1)
	pool.AddTask(t2)

	todos := pool.GetTasksByStatus(StatusTodo)
	if len(todos) != 1 {
		t.Errorf("Expected 1 todo task, got %d", len(todos))
	}

	blocked := pool.GetTasksByStatus(StatusBlocked)
	if len(blocked) != 1 {
		t.Errorf("Expected 1 blocked task, got %d", len(blocked))
	}
}

func TestTaskPool_GetAvailableForDoors(t *testing.T) {
	pool := NewTaskPool()
	for i := 0; i < 5; i++ {
		pool.AddTask(NewTask("Task"))
	}

	available := pool.GetAvailableForDoors()
	if len(available) != 5 {
		t.Errorf("Expected 5 available tasks, got %d", len(available))
	}

	// Complete one task
	allTasks := pool.GetAllTasks()
	_ = allTasks[0].UpdateStatus(StatusComplete)
	pool.UpdateTask(allTasks[0])

	available = pool.GetAvailableForDoors()
	if len(available) != 4 {
		t.Errorf("Expected 4 available tasks after completing one, got %d", len(available))
	}
}

func TestTaskPool_RecentlyShown(t *testing.T) {
	pool := NewTaskPool()
	task := NewTask("Test")
	pool.AddTask(task)

	if pool.IsRecentlyShown(task.ID) {
		t.Error("Task should not be recently shown initially")
	}

	pool.MarkRecentlyShown(task.ID)
	if !pool.IsRecentlyShown(task.ID) {
		t.Error("Task should be recently shown after marking")
	}
}

func TestTaskPool_GetAvailableForDoors_FewTasks(t *testing.T) {
	pool := NewTaskPool()
	t1 := NewTask("Only task")
	pool.AddTask(t1)
	pool.MarkRecentlyShown(t1.ID)

	// With < 3 tasks, should include recently shown
	available := pool.GetAvailableForDoors()
	if len(available) != 1 {
		t.Errorf("Expected 1 available task (including recently shown), got %d", len(available))
	}
}

func TestTaskPool_GetAvailableForDoors_DependencyBlocked(t *testing.T) {
	t.Parallel()

	t.Run("blocked tasks excluded from doors", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		b := NewTask("Task B") // no deps, should appear
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID} // A depends on B (todo), should be excluded
		pool.AddTask(a)
		pool.AddTask(b)

		available := pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available task, got %d", len(available))
		}
		if available[0].ID != b.ID {
			t.Errorf("expected task B, got %s", available[0].ID)
		}
	})

	t.Run("task with all deps complete appears in doors", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		b := NewTask("Task B")
		b.Status = StatusComplete
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID}
		pool.AddTask(a)
		pool.AddTask(b)

		available := pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available task (A with met deps), got %d", len(available))
		}
		if available[0].ID != a.ID {
			t.Errorf("expected task A, got %s", available[0].ID)
		}
	})

	t.Run("tasks without deps unaffected", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		for i := 0; i < 4; i++ {
			pool.AddTask(NewTask("Task"))
		}

		available := pool.GetAvailableForDoors()
		if len(available) != 4 {
			t.Errorf("expected 4 available tasks (no deps), got %d", len(available))
		}
	})

	t.Run("fallback with few tasks still excludes dep-blocked", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		b := NewTask("Task B") // only non-blocked task
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID} // blocked
		pool.AddTask(a)
		pool.AddTask(b)

		// Only 1 available (< 3), fallback triggered — but still excludes dep-blocked
		available := pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available task in fallback, got %d", len(available))
		}
		if available[0].ID != b.ID {
			t.Errorf("expected task B in fallback, got %s", available[0].ID)
		}
	})

	t.Run("complete dep then clear makes dependent available", func(t *testing.T) {
		t.Parallel()
		pool := NewTaskPool()
		b := NewTask("Task B")
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID}
		pool.AddTask(a)
		pool.AddTask(b)

		// Initially A is blocked
		available := pool.GetAvailableForDoors()
		for _, t2 := range available {
			if t2.ID == a.ID {
				t.Fatal("task A should not be available before dep completes")
			}
		}

		// Complete B and clear dep reference (simulating completion flow)
		b.Status = StatusComplete
		ClearCompletedDependency(b.ID, pool)
		pool.RemoveTask(b.ID)

		// Now A should be available
		available = pool.GetAvailableForDoors()
		if len(available) != 1 {
			t.Fatalf("expected 1 available task after dep clears, got %d", len(available))
		}
		if available[0].ID != a.ID {
			t.Errorf("expected task A, got %s", available[0].ID)
		}
	})
}

func TestTaskPool_FindBySourceRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		refs     []SourceRef
		provider string
		nativeID string
		wantNil  bool
	}{
		{
			name:     "finds task by exact match",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "jira",
			nativeID: "PROJ-42",
			wantNil:  false,
		},
		{
			name:     "returns nil for missing ref",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "jira",
			nativeID: "PROJ-99",
			wantNil:  true,
		},
		{
			name:     "returns nil for different provider",
			refs:     []SourceRef{{Provider: "jira", NativeID: "PROJ-42"}},
			provider: "obsidian",
			nativeID: "PROJ-42",
			wantNil:  true,
		},
		{
			name: "finds via second ref",
			refs: []SourceRef{
				{Provider: "textfile", NativeID: "abc"},
				{Provider: "jira", NativeID: "PROJ-42"},
			},
			provider: "jira",
			nativeID: "PROJ-42",
			wantNil:  false,
		},
		{
			name:     "empty pool returns nil",
			refs:     nil,
			provider: "jira",
			nativeID: "PROJ-42",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := NewTaskPool()
			if len(tt.refs) > 0 {
				task := NewTask("test task")
				task.SourceRefs = tt.refs
				pool.AddTask(task)
			}

			got := pool.FindBySourceRef(tt.provider, tt.nativeID)
			if tt.wantNil && got != nil {
				t.Errorf("FindBySourceRef(%q, %q) = %v, want nil", tt.provider, tt.nativeID, got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("FindBySourceRef(%q, %q) = nil, want task", tt.provider, tt.nativeID)
			}
		})
	}
}

func TestTaskPool_FindBySourceRef_IndexUpdatedOnUpdate(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("task")
	task.AddSourceRef("jira", "PROJ-1")
	pool.AddTask(task)

	if pool.FindBySourceRef("jira", "PROJ-1") == nil {
		t.Fatal("expected to find task after add")
	}

	// Update task with new ref, removing old
	task.SourceRefs = []SourceRef{{Provider: "obsidian", NativeID: "note-1"}}
	pool.UpdateTask(task)

	if pool.FindBySourceRef("jira", "PROJ-1") != nil {
		t.Error("old ref should be removed from index after update")
	}
	if pool.FindBySourceRef("obsidian", "note-1") == nil {
		t.Error("new ref should be in index after update")
	}
}

func TestTaskPool_FindBySourceRef_IndexCleanedOnRemove(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("task")
	task.AddSourceRef("jira", "PROJ-1")
	pool.AddTask(task)

	pool.RemoveTask(task.ID)

	if pool.FindBySourceRef("jira", "PROJ-1") != nil {
		t.Error("ref should be removed from index after task removal")
	}
}

func TestTaskPool_FindBySourceRef_NoRefsTask(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("no refs task")
	pool.AddTask(task)

	// Should not panic, just return nil for any lookup
	if pool.FindBySourceRef("any", "any") != nil {
		t.Error("expected nil for task with no source refs")
	}
}
