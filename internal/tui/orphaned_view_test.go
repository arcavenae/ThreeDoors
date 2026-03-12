package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestOrphanedView_EmptyPool(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	view := NewOrphanedView(pool)
	view.SetWidth(80)

	output := view.View()
	if !strings.Contains(output, "No orphaned tasks") {
		t.Error("expected empty message when no orphaned tasks")
	}
}

func TestOrphanedView_ShowsOrphanedTasks(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("Buy groceries")
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now
	task.SourceProvider = "todoist"
	pool.AddTask(task)

	view := NewOrphanedView(pool)
	view.SetWidth(80)

	output := view.View()
	if !strings.Contains(output, "Buy groceries") {
		t.Error("expected task text in orphaned view")
	}
	if !strings.Contains(output, "todoist") {
		t.Error("expected source provider in orphaned view")
	}
}

func TestOrphanedView_EscReturns(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	view := NewOrphanedView(pool)

	cmd := view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestOrphanedView_KeepAction(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("Keep me")
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now
	pool.AddTask(task)

	view := NewOrphanedView(pool)

	// Press 'K' to keep
	cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	if cmd == nil {
		t.Fatal("expected command from K key")
	}
	msg := cmd()
	action, ok := msg.(OrphanedTaskActionMsg)
	if !ok {
		t.Fatalf("expected OrphanedTaskActionMsg, got %T", msg)
	}
	if action.Action != "keep" {
		t.Errorf("expected keep action, got %q", action.Action)
	}
	if action.TaskID != task.ID {
		t.Errorf("expected task ID %q, got %q", task.ID, action.TaskID)
	}
}

func TestOrphanedView_DeleteAction(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("Delete me")
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now
	pool.AddTask(task)

	view := NewOrphanedView(pool)

	// Press 'd' to delete
	cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected command from d key")
	}
	msg := cmd()
	action, ok := msg.(OrphanedTaskActionMsg)
	if !ok {
		t.Fatalf("expected OrphanedTaskActionMsg, got %T", msg)
	}
	if action.Action != "delete" {
		t.Errorf("expected delete action, got %q", action.Action)
	}
}

func TestOrphanedView_Navigation(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		task := core.NewTask("Task")
		task.Orphaned = true
		task.OrphanedAt = &now
		pool.AddTask(task)
	}

	view := NewOrphanedView(pool)

	if view.selectedIndex != 0 {
		t.Fatalf("expected initial index 0, got %d", view.selectedIndex)
	}

	// Navigate down with 'j'
	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if view.selectedIndex != 1 {
		t.Errorf("expected index 1 after j, got %d", view.selectedIndex)
	}

	// Navigate down again
	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if view.selectedIndex != 2 {
		t.Errorf("expected index 2 after j, got %d", view.selectedIndex)
	}

	// Navigate up with 'k'
	view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if view.selectedIndex != 1 {
		t.Errorf("expected index 1 after k, got %d", view.selectedIndex)
	}
}

func TestOrphanedView_EnterKeeps(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("Enter to keep")
	now := time.Now().UTC()
	task.Orphaned = true
	task.OrphanedAt = &now
	pool.AddTask(task)

	view := NewOrphanedView(pool)

	cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter")
	}
	msg := cmd()
	action, ok := msg.(OrphanedTaskActionMsg)
	if !ok {
		t.Fatalf("expected OrphanedTaskActionMsg, got %T", msg)
	}
	if action.Action != "keep" {
		t.Errorf("expected keep action from Enter, got %q", action.Action)
	}
}

func TestOrphanedView_NoActionsOnEmpty(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	view := NewOrphanedView(pool)

	// All action keys should be no-ops on empty list
	for _, r := range []rune{'K', 'd', 'x', 'e'} {
		cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if cmd != nil {
			t.Errorf("expected nil command for %c on empty list", r)
		}
	}

	cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil command for Enter on empty list")
	}
}

func TestFormatTimeAgo(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"just now", now.Add(-10 * time.Second), "just now"},
		{"minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"1 min ago", now.Add(-90 * time.Second), "1 min ago"},
		{"hours ago", now.Add(-3 * time.Hour), "3h ago"},
		{"1 hour ago", now.Add(-90 * time.Minute), "1 hour ago"},
		{"days ago", now.Add(-48 * time.Hour), "2d ago"},
		{"1 day ago", now.Add(-36 * time.Hour), "1 day ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatTimeAgo(tt.time)
			if got != tt.want {
				t.Errorf("formatTimeAgo() = %q, want %q", got, tt.want)
			}
		})
	}
}
