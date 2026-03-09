package core

import (
	"testing"
	"time"
)

func TestGetIncompleteTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	pool := NewTaskPool()

	// Created two days ago, still todo
	t1 := &Task{ID: "1", Text: "Old todo", Status: StatusTodo, CreatedAt: twoDaysAgo}
	// Created two days ago, in progress
	t2 := &Task{ID: "2", Text: "Old in-progress", Status: StatusInProgress, CreatedAt: twoDaysAgo}
	// Created two days ago, blocked
	t3 := &Task{ID: "3", Text: "Old blocked", Status: StatusBlocked, CreatedAt: twoDaysAgo}
	// Created two days ago, completed
	t4 := &Task{ID: "4", Text: "Old complete", Status: StatusComplete, CreatedAt: twoDaysAgo}
	// Created today, todo
	t5 := &Task{ID: "5", Text: "New todo", Status: StatusTodo, CreatedAt: now}

	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)
	pool.AddTask(t4)
	pool.AddTask(t5)

	tests := []struct {
		name      string
		since     time.Time
		wantCount int
		wantIDs   map[string]bool
	}{
		{
			"since yesterday",
			yesterday,
			3, // t1, t2, t3 (created before yesterday), not t4 (complete), not t5 (created after)
			map[string]bool{"1": true, "2": true, "3": true},
		},
		{
			"since now",
			now,
			4, // t1, t2, t3, t5 (all created <= now)
			map[string]bool{"1": true, "2": true, "3": true, "5": true},
		},
		{
			"since future",
			now.Add(time.Hour),
			4, // all incomplete tasks created before future
			map[string]bool{"1": true, "2": true, "3": true, "5": true},
		},
		{
			"since three days ago",
			now.Add(-72 * time.Hour),
			0, // no tasks created that long ago
			map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pool.GetIncompleteTasks(tt.since)
			if len(got) != tt.wantCount {
				ids := make([]string, len(got))
				for i, task := range got {
					ids[i] = task.ID
				}
				t.Fatalf("GetIncompleteTasks() returned %d tasks %v, want %d", len(got), ids, tt.wantCount)
			}
			for _, task := range got {
				if !tt.wantIDs[task.ID] {
					t.Errorf("unexpected task ID %q in results", task.ID)
				}
			}
		})
	}
}

func TestGetIncompleteTasks_emptyPool(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	got := pool.GetIncompleteTasks(time.Now().UTC())
	if len(got) != 0 {
		t.Errorf("GetIncompleteTasks() on empty pool returned %d tasks, want 0", len(got))
	}
}

func TestGetIncompleteTasks_excludesDeferredAndArchived(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	pool := NewTaskPool()

	pool.AddTask(&Task{ID: "1", Text: "Deferred", Status: StatusDeferred, CreatedAt: now})
	pool.AddTask(&Task{ID: "2", Text: "Archived", Status: StatusArchived, CreatedAt: now})
	pool.AddTask(&Task{ID: "3", Text: "Todo", Status: StatusTodo, CreatedAt: now})

	got := pool.GetIncompleteTasks(now)
	if len(got) != 1 {
		t.Fatalf("GetIncompleteTasks() returned %d tasks, want 1", len(got))
	}
	if got[0].ID != "3" {
		t.Errorf("expected task ID 3, got %q", got[0].ID)
	}
}
