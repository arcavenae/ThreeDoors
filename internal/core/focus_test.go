package core

import (
	"testing"
	"time"
)

func TestHasFocusTag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"no focus", "Buy groceries", false},
		{"has focus", "Buy groceries +focus", true},
		{"focus at start", "+focus Buy groceries", true},
		{"focus in middle", "Buy +focus groceries", true},
		{"case insensitive", "Buy groceries +Focus", true},
		{"upper case", "Buy groceries +FOCUS", true},
		{"focus substring", "Buy groceries +focused", false},
		{"focus in word", "unfocused task", false},
		{"empty text", "", false},
		{"only focus", "+focus", true},
		{"focus with other tags", "+focus @deep-work #technical", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{Text: tt.text}
			got := HasFocusTag(task)
			if got != tt.want {
				t.Errorf("HasFocusTag(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestGetFocusTasks(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()

	t1 := &Task{ID: "1", Text: "Task one +focus", Status: StatusTodo}
	t2 := &Task{ID: "2", Text: "Task two", Status: StatusTodo}
	t3 := &Task{ID: "3", Text: "Task three +focus", Status: StatusInProgress}
	t4 := &Task{ID: "4", Text: "Task four", Status: StatusComplete}

	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)
	pool.AddTask(t4)

	got := GetFocusTasks(pool)
	if len(got) != 2 {
		t.Fatalf("GetFocusTasks() returned %d tasks, want 2", len(got))
	}

	ids := make(map[string]bool)
	for _, task := range got {
		ids[task.ID] = true
	}
	if !ids["1"] || !ids["3"] {
		t.Errorf("GetFocusTasks() returned wrong tasks: got IDs %v, want [1, 3]", ids)
	}
}

func TestGetFocusTasks_empty(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()
	pool.AddTask(&Task{ID: "1", Text: "No focus here", Status: StatusTodo})

	got := GetFocusTasks(pool)
	if len(got) != 0 {
		t.Errorf("GetFocusTasks() returned %d tasks, want 0", len(got))
	}
}

func TestClearFocusTags(t *testing.T) {
	t.Parallel()
	pool := NewTaskPool()

	now := time.Now().UTC()
	t1 := &Task{ID: "1", Text: "Task one +focus", Status: StatusTodo, UpdatedAt: now}
	t2 := &Task{ID: "2", Text: "Task two", Status: StatusTodo, UpdatedAt: now}
	t3 := &Task{ID: "3", Text: "+focus Task three +focus", Status: StatusTodo, UpdatedAt: now}

	pool.AddTask(t1)
	pool.AddTask(t2)
	pool.AddTask(t3)

	ClearFocusTags(pool)

	if HasFocusTag(t1) {
		t.Errorf("Task 1 still has focus tag after ClearFocusTags")
	}
	if t1.Text != "Task one" {
		t.Errorf("Task 1 text = %q, want %q", t1.Text, "Task one")
	}
	if t2.Text != "Task two" {
		t.Errorf("Task 2 text = %q, want %q", t2.Text, "Task two")
	}
	if HasFocusTag(t3) {
		t.Errorf("Task 3 still has focus tag after ClearFocusTags (both occurrences)")
	}
	if t3.Text != "Task three" {
		t.Errorf("Task 3 text = %q, want %q", t3.Text, "Task three")
	}
}

func TestIsFocusExpired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		hoursAgo float64
		want     bool
	}{
		{"just now", 0, false},
		{"1 hour ago", 1, false},
		{"15 hours ago", 15, false},
		{"16 hours ago", 16, true},
		{"24 hours ago", 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := time.Now().UTC().Add(-time.Duration(tt.hoursAgo * float64(time.Hour)))
			got := IsFocusExpired(ts)
			if got != tt.want {
				t.Errorf("IsFocusExpired(%v hours ago) = %v, want %v", tt.hoursAgo, got, tt.want)
			}
		})
	}
}

func TestFocusScoreBoost(t *testing.T) {
	t.Parallel()
	recentPlanning := time.Now().UTC().Add(-1 * time.Hour)
	expiredPlanning := time.Now().UTC().Add(-17 * time.Hour)

	tests := []struct {
		name         string
		tasks        []*Task
		planningTime time.Time
		want         float64
	}{
		{
			"no focus tasks",
			[]*Task{{Text: "task a"}, {Text: "task b"}},
			recentPlanning,
			0,
		},
		{
			"one focus task, not expired",
			[]*Task{{Text: "task a +focus"}, {Text: "task b"}},
			recentPlanning,
			FocusBoost,
		},
		{
			"two focus tasks, not expired",
			[]*Task{{Text: "task a +focus"}, {Text: "task b +focus"}},
			recentPlanning,
			FocusBoost * 2,
		},
		{
			"focus tasks but expired",
			[]*Task{{Text: "task a +focus"}, {Text: "task b +focus"}},
			expiredPlanning,
			0,
		},
		{
			"empty tasks",
			nil,
			recentPlanning,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FocusScoreBoost(tt.tasks, tt.planningTime)
			if got != tt.want {
				t.Errorf("FocusScoreBoost() = %v, want %v", got, tt.want)
			}
		})
	}
}
