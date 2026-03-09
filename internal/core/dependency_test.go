package core

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// helper to build a pool with preconfigured tasks.
func buildTestPool(tasks ...*Task) *TaskPool {
	pool := NewTaskPool()
	for _, t := range tasks {
		pool.AddTask(t)
	}
	return pool
}

func TestHasUnmetDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() (*Task, *TaskPool)
		wantMet bool // true = no unmet deps (HasUnmetDependencies returns false)
	}{
		{
			name: "no dependencies",
			setup: func() (*Task, *TaskPool) {
				a := NewTask("Task A")
				return a, buildTestPool(a)
			},
			wantMet: true,
		},
		{
			name: "all deps complete",
			setup: func() (*Task, *TaskPool) {
				b := NewTask("Task B")
				b.Status = StatusComplete
				c := NewTask("Task C")
				c.Status = StatusComplete
				a := NewTask("Task A")
				a.DependsOn = []string{b.ID, c.ID}
				return a, buildTestPool(a, b, c)
			},
			wantMet: true,
		},
		{
			name: "one dep incomplete",
			setup: func() (*Task, *TaskPool) {
				b := NewTask("Task B")
				b.Status = StatusTodo
				c := NewTask("Task C")
				c.Status = StatusComplete
				a := NewTask("Task A")
				a.DependsOn = []string{b.ID, c.ID}
				return a, buildTestPool(a, b, c)
			},
			wantMet: false,
		},
		{
			name: "orphaned dep treated as unmet",
			setup: func() (*Task, *TaskPool) {
				a := NewTask("Task A")
				a.DependsOn = []string{"nonexistent-id"}
				return a, buildTestPool(a)
			},
			wantMet: false,
		},
		{
			name: "dep in progress treated as unmet",
			setup: func() (*Task, *TaskPool) {
				b := NewTask("Task B")
				b.Status = StatusInProgress
				a := NewTask("Task A")
				a.DependsOn = []string{b.ID}
				return a, buildTestPool(a, b)
			},
			wantMet: false,
		},
		{
			name: "empty depends_on slice",
			setup: func() (*Task, *TaskPool) {
				a := NewTask("Task A")
				a.DependsOn = []string{}
				return a, buildTestPool(a)
			},
			wantMet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task, pool := tt.setup()
			got := HasUnmetDependencies(task, pool)
			want := !tt.wantMet
			if got != want {
				t.Errorf("HasUnmetDependencies() = %v, want %v", got, want)
			}
		})
	}
}

func TestGetBlockingDependencies(t *testing.T) {
	t.Parallel()

	t.Run("returns incomplete deps", func(t *testing.T) {
		t.Parallel()
		b := NewTask("Task B")
		b.Status = StatusTodo
		c := NewTask("Task C")
		c.Status = StatusComplete
		d := NewTask("Task D")
		d.Status = StatusInProgress
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID, c.ID, d.ID}
		pool := buildTestPool(a, b, c, d)

		blocking := GetBlockingDependencies(a, pool)
		if len(blocking) != 2 {
			t.Fatalf("expected 2 blocking deps, got %d", len(blocking))
		}

		ids := map[string]bool{}
		for _, bl := range blocking {
			ids[bl.ID] = true
		}
		if !ids[b.ID] {
			t.Error("expected Task B to be blocking")
		}
		if !ids[d.ID] {
			t.Error("expected Task D to be blocking")
		}
	})

	t.Run("orphaned dep returns placeholder", func(t *testing.T) {
		t.Parallel()
		a := NewTask("Task A")
		a.DependsOn = []string{"orphaned-id"}
		pool := buildTestPool(a)

		blocking := GetBlockingDependencies(a, pool)
		if len(blocking) != 1 {
			t.Fatalf("expected 1 blocking dep, got %d", len(blocking))
		}
		if blocking[0].ID != "orphaned-id" {
			t.Errorf("expected orphaned-id, got %s", blocking[0].ID)
		}
		if blocking[0].Text != "[deleted task]" {
			t.Errorf("expected placeholder text, got %q", blocking[0].Text)
		}
	})

	t.Run("all deps complete returns nil", func(t *testing.T) {
		t.Parallel()
		b := NewTask("Task B")
		b.Status = StatusComplete
		a := NewTask("Task A")
		a.DependsOn = []string{b.ID}
		pool := buildTestPool(a, b)

		blocking := GetBlockingDependencies(a, pool)
		if len(blocking) != 0 {
			t.Errorf("expected nil/empty blocking list, got %d", len(blocking))
		}
	})

	t.Run("no dependencies returns nil", func(t *testing.T) {
		t.Parallel()
		a := NewTask("Task A")
		pool := buildTestPool(a)

		blocking := GetBlockingDependencies(a, pool)
		if blocking != nil {
			t.Errorf("expected nil, got %v", blocking)
		}
	})
}

func TestWouldCreateCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() *TaskPool
		taskID    string
		depID     string
		wantCycle bool
	}{
		{
			name: "self-dependency",
			setup: func() *TaskPool {
				a := NewTask("Task A")
				a.ID = "a"
				return buildTestPool(a)
			},
			taskID:    "a",
			depID:     "a",
			wantCycle: true,
		},
		{
			name: "direct cycle A->B, adding B->A",
			setup: func() *TaskPool {
				a := &Task{ID: "a", Text: "A", DependsOn: []string{"b"}}
				b := &Task{ID: "b", Text: "B"}
				return buildTestPool(a, b)
			},
			taskID:    "b",
			depID:     "a",
			wantCycle: true,
		},
		{
			name: "transitive cycle A->B->C, adding C->A",
			setup: func() *TaskPool {
				a := &Task{ID: "a", Text: "A", DependsOn: []string{"b"}}
				b := &Task{ID: "b", Text: "B", DependsOn: []string{"c"}}
				c := &Task{ID: "c", Text: "C"}
				return buildTestPool(a, b, c)
			},
			taskID:    "c",
			depID:     "a",
			wantCycle: true,
		},
		{
			name: "no cycle - independent chains",
			setup: func() *TaskPool {
				a := &Task{ID: "a", Text: "A", DependsOn: []string{"b"}}
				b := &Task{ID: "b", Text: "B"}
				c := &Task{ID: "c", Text: "C", DependsOn: []string{"d"}}
				d := &Task{ID: "d", Text: "D"}
				return buildTestPool(a, b, c, d)
			},
			taskID:    "a",
			depID:     "c",
			wantCycle: false,
		},
		{
			name: "no cycle - simple linear chain",
			setup: func() *TaskPool {
				a := &Task{ID: "a", Text: "A"}
				b := &Task{ID: "b", Text: "B"}
				return buildTestPool(a, b)
			},
			taskID:    "a",
			depID:     "b",
			wantCycle: false,
		},
		{
			name: "no cycle - diamond shape",
			setup: func() *TaskPool {
				// A->B, A->C, B->D, C->D — adding E->A is fine
				a := &Task{ID: "a", Text: "A", DependsOn: []string{"b", "c"}}
				b := &Task{ID: "b", Text: "B", DependsOn: []string{"d"}}
				c := &Task{ID: "c", Text: "C", DependsOn: []string{"d"}}
				d := &Task{ID: "d", Text: "D"}
				e := &Task{ID: "e", Text: "E"}
				return buildTestPool(a, b, c, d, e)
			},
			taskID:    "e",
			depID:     "a",
			wantCycle: false,
		},
		{
			name: "cycle through diamond",
			setup: func() *TaskPool {
				// A->B, A->C, B->D, C->D — adding D->A creates cycle
				a := &Task{ID: "a", Text: "A", DependsOn: []string{"b", "c"}}
				b := &Task{ID: "b", Text: "B", DependsOn: []string{"d"}}
				c := &Task{ID: "c", Text: "C", DependsOn: []string{"d"}}
				d := &Task{ID: "d", Text: "D"}
				return buildTestPool(a, b, c, d)
			},
			taskID:    "d",
			depID:     "a",
			wantCycle: true,
		},
		{
			name: "dep target does not exist",
			setup: func() *TaskPool {
				a := &Task{ID: "a", Text: "A"}
				return buildTestPool(a)
			},
			taskID:    "a",
			depID:     "nonexistent",
			wantCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := tt.setup()
			got := WouldCreateCycle(tt.taskID, tt.depID, pool)
			if got != tt.wantCycle {
				t.Errorf("WouldCreateCycle(%q, %q) = %v, want %v", tt.taskID, tt.depID, got, tt.wantCycle)
			}
		})
	}
}

func TestGetNewlyUnblockedTasks(t *testing.T) {
	t.Parallel()

	t.Run("single dependency completed", func(t *testing.T) {
		t.Parallel()
		b := &Task{ID: "b", Text: "B", Status: StatusComplete}
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b"}}
		pool := buildTestPool(a, b)

		unblocked := GetNewlyUnblockedTasks("b", pool)
		if len(unblocked) != 1 {
			t.Fatalf("expected 1 unblocked task, got %d", len(unblocked))
		}
		if unblocked[0].ID != "a" {
			t.Errorf("expected task a, got %s", unblocked[0].ID)
		}
	})

	t.Run("still has other unmet deps", func(t *testing.T) {
		t.Parallel()
		b := &Task{ID: "b", Text: "B", Status: StatusComplete}
		c := &Task{ID: "c", Text: "C", Status: StatusTodo}
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b", "c"}}
		pool := buildTestPool(a, b, c)

		unblocked := GetNewlyUnblockedTasks("b", pool)
		if len(unblocked) != 0 {
			t.Errorf("expected 0 unblocked tasks (c still unmet), got %d", len(unblocked))
		}
	})

	t.Run("multiple tasks unblocked by same completion", func(t *testing.T) {
		t.Parallel()
		c := &Task{ID: "c", Text: "C", Status: StatusComplete}
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"c"}}
		b := &Task{ID: "b", Text: "B", Status: StatusTodo, DependsOn: []string{"c"}}
		pool := buildTestPool(a, b, c)

		unblocked := GetNewlyUnblockedTasks("c", pool)
		if len(unblocked) != 2 {
			t.Fatalf("expected 2 unblocked tasks, got %d", len(unblocked))
		}
	})

	t.Run("no tasks depend on completed", func(t *testing.T) {
		t.Parallel()
		a := &Task{ID: "a", Text: "A", Status: StatusComplete}
		b := &Task{ID: "b", Text: "B", Status: StatusTodo}
		pool := buildTestPool(a, b)

		unblocked := GetNewlyUnblockedTasks("a", pool)
		if len(unblocked) != 0 {
			t.Errorf("expected 0 unblocked tasks, got %d", len(unblocked))
		}
	})
}

func TestDependsOnYAMLRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("preserves depends_on through round-trip", func(t *testing.T) {
		t.Parallel()
		task := NewTask("Deploy to production")
		task.DependsOn = []string{"task-id-1", "task-id-2"}

		data, err := yaml.Marshal(task)
		if err != nil {
			t.Fatalf("yaml.Marshal: %v", err)
		}

		var restored Task
		if err := yaml.Unmarshal(data, &restored); err != nil {
			t.Fatalf("yaml.Unmarshal: %v", err)
		}

		if len(restored.DependsOn) != 2 {
			t.Fatalf("expected 2 depends_on entries, got %d", len(restored.DependsOn))
		}
		if restored.DependsOn[0] != "task-id-1" {
			t.Errorf("expected task-id-1, got %s", restored.DependsOn[0])
		}
		if restored.DependsOn[1] != "task-id-2" {
			t.Errorf("expected task-id-2, got %s", restored.DependsOn[1])
		}
	})

	t.Run("omits depends_on when nil", func(t *testing.T) {
		t.Parallel()
		task := NewTask("Simple task")

		data, err := yaml.Marshal(task)
		if err != nil {
			t.Fatalf("yaml.Marshal: %v", err)
		}

		yamlStr := string(data)
		if contains(yamlStr, "depends_on") {
			t.Error("expected depends_on to be omitted from YAML when nil")
		}
	})

	t.Run("omits depends_on when empty slice", func(t *testing.T) {
		t.Parallel()
		task := NewTask("Simple task")
		task.DependsOn = []string{}

		data, err := yaml.Marshal(task)
		if err != nil {
			t.Fatalf("yaml.Marshal: %v", err)
		}

		yamlStr := string(data)
		if contains(yamlStr, "depends_on") {
			t.Error("expected depends_on to be omitted from YAML when empty slice")
		}
	})
}

func TestClearCompletedDependency(t *testing.T) {
	t.Parallel()

	t.Run("removes completed dep from all tasks", func(t *testing.T) {
		t.Parallel()
		c := &Task{ID: "c", Text: "C", Status: StatusComplete}
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"c", "d"}}
		b := &Task{ID: "b", Text: "B", Status: StatusTodo, DependsOn: []string{"c"}}
		d := &Task{ID: "d", Text: "D", Status: StatusTodo}
		pool := buildTestPool(a, b, c, d)

		ClearCompletedDependency("c", pool)

		gotA := pool.GetTask("a")
		if len(gotA.DependsOn) != 1 || gotA.DependsOn[0] != "d" {
			t.Errorf("task A DependsOn = %v, want [d]", gotA.DependsOn)
		}
		gotB := pool.GetTask("b")
		if gotB.DependsOn != nil {
			t.Errorf("task B DependsOn = %v, want nil", gotB.DependsOn)
		}
	})

	t.Run("no-op when no tasks depend on completed", func(t *testing.T) {
		t.Parallel()
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b"}}
		b := &Task{ID: "b", Text: "B", Status: StatusTodo}
		pool := buildTestPool(a, b)

		ClearCompletedDependency("nonexistent", pool)

		gotA := pool.GetTask("a")
		if len(gotA.DependsOn) != 1 || gotA.DependsOn[0] != "b" {
			t.Errorf("task A DependsOn = %v, want [b]", gotA.DependsOn)
		}
	})

	t.Run("sets DependsOn to nil when last dep cleared", func(t *testing.T) {
		t.Parallel()
		a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b"}}
		b := &Task{ID: "b", Text: "B", Status: StatusComplete}
		pool := buildTestPool(a, b)

		ClearCompletedDependency("b", pool)

		gotA := pool.GetTask("a")
		if gotA.DependsOn != nil {
			t.Errorf("task A DependsOn = %v, want nil", gotA.DependsOn)
		}
	})
}

func TestCascadingUnblock(t *testing.T) {
	t.Parallel()

	// Chain: A depends on B, B depends on C
	c := &Task{ID: "c", Text: "C", Status: StatusTodo}
	b := &Task{ID: "b", Text: "B", Status: StatusTodo, DependsOn: []string{"c"}}
	a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b"}}
	pool := buildTestPool(a, b, c)

	// Both A and B should have unmet deps
	if !HasUnmetDependencies(a, pool) {
		t.Fatal("A should have unmet deps initially")
	}
	if !HasUnmetDependencies(b, pool) {
		t.Fatal("B should have unmet deps initially")
	}

	// Complete C
	c.Status = StatusComplete
	unblockedByC := GetNewlyUnblockedTasks("c", pool)
	if len(unblockedByC) != 1 || unblockedByC[0].ID != "b" {
		t.Fatalf("completing C should unblock B, got %v", unblockedByC)
	}
	ClearCompletedDependency("c", pool)

	// A is still blocked (B not complete yet)
	if !HasUnmetDependencies(a, pool) {
		t.Error("A should still have unmet deps (B not complete)")
	}

	// Complete B
	b.Status = StatusComplete
	unblockedByB := GetNewlyUnblockedTasks("b", pool)
	if len(unblockedByB) != 1 || unblockedByB[0].ID != "a" {
		t.Fatalf("completing B should unblock A, got %v", unblockedByB)
	}
	ClearCompletedDependency("b", pool)

	// A is now unblocked
	if HasUnmetDependencies(a, pool) {
		t.Error("A should have no unmet deps after B completes")
	}
}

func TestAutoUnblockOnlyFiresForNewlyUnblocked(t *testing.T) {
	t.Parallel()

	// A depends on B and C. B completes but C is still todo.
	b := &Task{ID: "b", Text: "B", Status: StatusComplete}
	c := &Task{ID: "c", Text: "C", Status: StatusTodo}
	a := &Task{ID: "a", Text: "A", Status: StatusTodo, DependsOn: []string{"b", "c"}}
	pool := buildTestPool(a, b, c)

	unblocked := GetNewlyUnblockedTasks("b", pool)
	if len(unblocked) != 0 {
		t.Errorf("A still has unmet dep C, should not be unblocked, got %d", len(unblocked))
	}
}

func TestDependencyLargeFanOut(t *testing.T) {
	t.Parallel()

	// Task with 15 dependencies — verify correctness at scale
	deps := make([]*Task, 15)
	depIDs := make([]string, 15)
	for i := range deps {
		deps[i] = NewTask("Dep task")
		deps[i].Status = StatusComplete
		depIDs[i] = deps[i].ID
	}

	a := NewTask("Task A")
	a.DependsOn = depIDs

	allTasks := append([]*Task{a}, deps...)
	pool := buildTestPool(allTasks...)

	// All complete — should not be blocked
	if HasUnmetDependencies(a, pool) {
		t.Error("all deps are complete, should not have unmet dependencies")
	}

	// Make one incomplete
	deps[7].Status = StatusTodo
	if !HasUnmetDependencies(a, pool) {
		t.Error("dep 7 is todo, should have unmet dependencies")
	}

	blocking := GetBlockingDependencies(a, pool)
	if len(blocking) != 1 {
		t.Fatalf("expected 1 blocking dep, got %d", len(blocking))
	}
	if blocking[0].ID != deps[7].ID {
		t.Errorf("expected dep 7 to be blocking, got %s", blocking[0].ID)
	}
}
