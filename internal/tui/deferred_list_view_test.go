package tui

import (
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func newDeferredTask(text string, deferUntil *time.Time) *core.Task {
	t := core.NewTask(text)
	t.Status = core.StatusDeferred
	t.DeferUntil = deferUntil
	return t
}

func TestSortDeferredTasks(t *testing.T) {
	t.Parallel()

	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	nextWeek := time.Now().UTC().Add(7 * 24 * time.Hour)

	tests := []struct {
		name     string
		tasks    []*core.Task
		wantText []string
	}{
		{
			name:     "empty",
			tasks:    nil,
			wantText: nil,
		},
		{
			name: "sorted by date ascending",
			tasks: []*core.Task{
				newDeferredTask("next week", &nextWeek),
				newDeferredTask("tomorrow", &tomorrow),
			},
			wantText: []string{"tomorrow", "next week"},
		},
		{
			name: "nil DeferUntil last",
			tasks: []*core.Task{
				newDeferredTask("someday", nil),
				newDeferredTask("tomorrow", &tomorrow),
				newDeferredTask("next week", &nextWeek),
			},
			wantText: []string{"tomorrow", "next week", "someday"},
		},
		{
			name: "multiple nil DeferUntil last",
			tasks: []*core.Task{
				newDeferredTask("someday A", nil),
				newDeferredTask("tomorrow", &tomorrow),
				newDeferredTask("someday B", nil),
			},
			wantText: []string{"tomorrow", "someday A", "someday B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sortDeferredTasks(tt.tasks)
			for i, task := range tt.tasks {
				if task.Text != tt.wantText[i] {
					t.Errorf("index %d: got %q, want %q", i, task.Text, tt.wantText[i])
				}
			}
		})
	}
}

func TestFormatTimeRemaining(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	// Use midnight-based offsets so partial-day rounding doesn't affect results
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)

	tests := []struct {
		name       string
		deferUntil *time.Time
		want       string
	}{
		{
			name:       "nil returns Someday",
			deferUntil: nil,
			want:       "Someday",
		},
		{
			name:       "past time returns Overdue",
			deferUntil: timePtr(now.Add(-1 * time.Hour)),
			want:       "Overdue",
		},
		{
			name:       "less than 24 hours returns Tomorrow",
			deferUntil: timePtr(now.Add(20 * time.Hour)),
			want:       "Tomorrow",
		},
		{
			name:       "3 days returns 3 days",
			deferUntil: timePtr(midnight.Add(3 * 24 * time.Hour)),
			want:       "3 days",
		},
		{
			name:       "5 days returns 5 days",
			deferUntil: timePtr(midnight.Add(5 * 24 * time.Hour)),
			want:       "5 days",
		},
		{
			name:       "14 days returns 2 weeks",
			deferUntil: timePtr(midnight.Add(14 * 24 * time.Hour)),
			want:       "2 weeks",
		},
		{
			name:       "7 days returns 1 week",
			deferUntil: timePtr(midnight.Add(7 * 24 * time.Hour)),
			want:       "1 week",
		},
		{
			name:       "60 days returns 2 months",
			deferUntil: timePtr(midnight.Add(60 * 24 * time.Hour)),
			want:       "2 months",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatTimeRemaining(tt.deferUntil)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatReturnDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deferUntil *time.Time
		want       string
	}{
		{
			name:       "nil returns Someday",
			deferUntil: nil,
			want:       "Someday",
		},
		{
			name:       "date formatted",
			deferUntil: timePtr(time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC)),
			want:       time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC).Local().Format("Jan 2, 2006"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatReturnDate(tt.deferUntil)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeferredListViewEmptyState(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	view := dv.View()
	if !containsStr(view, "No snoozed tasks") {
		t.Errorf("empty state should show 'No snoozed tasks', got: %s", view)
	}
	if !containsStr(view, "Use Z on a door to snooze") {
		t.Errorf("empty state should show hint about Z key, got: %s", view)
	}
}

func TestDeferredListViewWithTasks(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	pool.AddTask(newDeferredTask("Buy groceries", &tomorrow))
	pool.AddTask(newDeferredTask("Someday task", nil))

	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	view := dv.View()
	if !containsStr(view, "Buy groceries") {
		t.Errorf("view should show task text, got: %s", view)
	}
	if !containsStr(view, "Someday task") {
		t.Errorf("view should show someday task, got: %s", view)
	}
	if !containsStr(view, "2 snoozed task(s)") {
		t.Errorf("view should show task count, got: %s", view)
	}
}

func TestDeferredListViewNavigation(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	nextWeek := time.Now().UTC().Add(7 * 24 * time.Hour)
	pool.AddTask(newDeferredTask("Task A", &tomorrow))
	pool.AddTask(newDeferredTask("Task B", &nextWeek))
	pool.AddTask(newDeferredTask("Task C", nil))

	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	if dv.cursor != 0 {
		t.Fatalf("initial cursor should be 0, got %d", dv.cursor)
	}

	// Move down
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if dv.cursor != 1 {
		t.Errorf("cursor should be 1 after j, got %d", dv.cursor)
	}

	// Move down again
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if dv.cursor != 2 {
		t.Errorf("cursor should be 2 after second j, got %d", dv.cursor)
	}

	// Move down at bottom — stays at 2
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if dv.cursor != 2 {
		t.Errorf("cursor should stay at 2 at bottom, got %d", dv.cursor)
	}

	// Move up
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if dv.cursor != 1 {
		t.Errorf("cursor should be 1 after k, got %d", dv.cursor)
	}
}

func TestDeferredListViewUnsnooze(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	task := newDeferredTask("Unsnoozeme", &tomorrow)
	pool.AddTask(task)

	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	if cmd == nil {
		t.Fatal("u key should produce a command")
		return
	}

	msg := cmd()
	unsnoozedMsg, ok := msg.(UnsnoozeTaskMsg)
	if !ok {
		t.Fatalf("expected UnsnoozeTaskMsg, got %T", msg)
	}
	if unsnoozedMsg.Task.Text != "Unsnoozeme" {
		t.Errorf("expected task text 'Unsnoozeme', got %q", unsnoozedMsg.Task.Text)
	}
}

func TestDeferredListViewEditDate(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	task := newDeferredTask("Edit my date", &tomorrow)
	pool.AddTask(task)

	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("e key should produce a command")
		return
	}

	msg := cmd()
	editMsg, ok := msg.(EditDeferDateMsg)
	if !ok {
		t.Fatalf("expected EditDeferDateMsg, got %T", msg)
	}
	if editMsg.Task.Text != "Edit my date" {
		t.Errorf("expected task text 'Edit my date', got %q", editMsg.Task.Text)
	}
}

func TestDeferredListViewEscReturns(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	dv := NewDeferredListView(pool)
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestDeferredListViewRefresh(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	pool.AddTask(newDeferredTask("Task A", &tomorrow))

	dv := NewDeferredListView(pool)
	if len(dv.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(dv.tasks))
	}

	// Add another deferred task to pool
	nextWeek := time.Now().UTC().Add(7 * 24 * time.Hour)
	pool.AddTask(newDeferredTask("Task B", &nextWeek))

	dv.Refresh()
	if len(dv.tasks) != 2 {
		t.Fatalf("after refresh, expected 2 tasks, got %d", len(dv.tasks))
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
