package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

func TestConflictView_View(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	conflicts := []tasks.Conflict{
		{
			LocalTask:  &tasks.Task{ID: "1", Text: "Buy groceries", Status: tasks.StatusInProgress, UpdatedAt: now},
			RemoteTask: &tasks.Task{ID: "1", Text: "Buy organic groceries", Status: tasks.StatusTodo, UpdatedAt: now},
		},
	}
	cs := tasks.NewConflictSet("Local", conflicts)
	cv := NewConflictView(cs, nil)
	cv.SetWidth(80)

	view := cv.View()

	if !strings.Contains(view, "Sync Conflict") {
		t.Error("view should contain 'Sync Conflict' header")
	}
	if !strings.Contains(view, "1/1") {
		t.Error("view should show conflict count '1/1'")
	}
	if !strings.Contains(view, "Buy groceries") {
		t.Error("view should show local task text")
	}
	if !strings.Contains(view, "Buy organic groceries") {
		t.Error("view should show remote task text")
	}
	if !strings.Contains(view, "LOCAL") {
		t.Error("view should show LOCAL label")
	}
	if !strings.Contains(view, "REMOTE") {
		t.Error("view should show REMOTE label")
	}
	if !strings.Contains(view, "Text differs") {
		t.Error("view should highlight text difference")
	}
	if !strings.Contains(view, "Status:") {
		t.Error("view should highlight status difference")
	}
}

func TestConflictView_AllResolved(t *testing.T) {
	t.Parallel()
	cs := tasks.NewConflictSet("Local", nil)
	cv := NewConflictView(cs, nil)

	view := cv.View()

	if !strings.Contains(view, "All conflicts resolved") {
		t.Error("view should show 'All conflicts resolved' for empty set")
	}
}

func TestConflictView_ResolveLocal(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cs := tasks.NewConflictSet("Test", []tasks.Conflict{
		{
			LocalTask:  &tasks.Task{ID: "1", Text: "local", UpdatedAt: now},
			RemoteTask: &tasks.Task{ID: "1", Text: "remote", UpdatedAt: now},
		},
	})
	cv := NewConflictView(cs, nil)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal("expected command after resolving last conflict")
	}

	msg := cmd()
	if _, ok := msg.(ConflictResolvedMsg); !ok {
		t.Errorf("expected ConflictResolvedMsg, got %T", msg)
	}
}

func TestConflictView_ResolveRemote(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cs := tasks.NewConflictSet("Test", []tasks.Conflict{
		{
			LocalTask:  &tasks.Task{ID: "1", Text: "local", UpdatedAt: now},
			RemoteTask: &tasks.Task{ID: "1", Text: "remote", UpdatedAt: now},
		},
	})
	cv := NewConflictView(cs, nil)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected command after resolving")
	}

	msg := cmd()
	resolved, ok := msg.(ConflictResolvedMsg)
	if !ok {
		t.Fatalf("expected ConflictResolvedMsg, got %T", msg)
	}

	resolutions := resolved.ConflictSet.Resolutions()
	if len(resolutions) != 1 {
		t.Fatalf("expected 1 resolution, got %d", len(resolutions))
	}
	if resolutions[0].Winner != "remote" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "remote")
	}
}

func TestConflictView_ResolveBoth(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cs := tasks.NewConflictSet("Test", []tasks.Conflict{
		{
			LocalTask:  &tasks.Task{ID: "1", Text: "local", UpdatedAt: now},
			RemoteTask: &tasks.Task{ID: "1", Text: "remote", UpdatedAt: now},
		},
	})
	cv := NewConflictView(cs, nil)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if cmd == nil {
		t.Fatal("expected command after resolving")
	}

	msg := cmd()
	resolved, ok := msg.(ConflictResolvedMsg)
	if !ok {
		t.Fatalf("expected ConflictResolvedMsg, got %T", msg)
	}

	resolutions := resolved.ConflictSet.Resolutions()
	if resolutions[0].Winner != "both" {
		t.Errorf("Winner = %q, want %q", resolutions[0].Winner, "both")
	}
}

func TestConflictView_EscReturns(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	cs := tasks.NewConflictSet("Test", []tasks.Conflict{
		{
			LocalTask:  &tasks.Task{ID: "1", Text: "local", UpdatedAt: now},
			RemoteTask: &tasks.Task{ID: "1", Text: "remote", UpdatedAt: now},
		},
	})
	cv := NewConflictView(cs, nil)

	cmd := cv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on Esc")
	}

	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}
