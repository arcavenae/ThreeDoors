package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDisconnectDialog_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		label    string
		contains []string
	}{
		{
			name:  "shows connection label",
			label: "Work Jira",
			contains: []string{
				"Work Jira",
				"Keep tasks locally",
				"Remove synced tasks",
				"[Enter] Disconnect",
				"[Esc] Cancel",
			},
		},
		{
			name:  "truncates long label",
			label: strings.Repeat("A", 60),
			contains: []string{
				"AAA...",
				"Keep tasks locally",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := NewDisconnectDialog("conn-1", tt.label)
			d.SetWidth(80)
			view := d.View()
			for _, want := range tt.contains {
				if !strings.Contains(view, want) {
					t.Errorf("View() missing %q", want)
				}
			}
		})
	}
}

func TestDisconnectDialog_CursorNavigation(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test")

	// Initial cursor at 0 (keep tasks)
	if d.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", d.cursor)
	}

	// Move down
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", d.cursor)
	}

	// Don't go past end
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 1 {
		t.Errorf("cursor after extra j = %d, want 1", d.cursor)
	}

	// Move up
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", d.cursor)
	}

	// Don't go before start
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 0 {
		t.Errorf("cursor after extra k = %d, want 0", d.cursor)
	}

	// Arrow keys
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", d.cursor)
	}
	d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", d.cursor)
	}
}

func TestDisconnectDialog_ConfirmKeepTasks(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test Source")
	// cursor starts at 0 = keep tasks

	cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Enter")
	}
	msg := cmd()
	confirmed, ok := msg.(DisconnectConfirmedMsg)
	if !ok {
		t.Fatalf("expected DisconnectConfirmedMsg, got %T", msg)
	}
	if confirmed.ConnectionID != "conn-1" {
		t.Errorf("ConnectionID = %q, want %q", confirmed.ConnectionID, "conn-1")
	}
	if !confirmed.KeepTasks {
		t.Error("KeepTasks = false, want true (cursor at keep)")
	}
}

func TestDisconnectDialog_ConfirmRemoveTasks(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test Source")
	// Move to "Remove synced tasks"
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Enter")
	}
	msg := cmd()
	confirmed, ok := msg.(DisconnectConfirmedMsg)
	if !ok {
		t.Fatalf("expected DisconnectConfirmedMsg, got %T", msg)
	}
	if confirmed.KeepTasks {
		t.Error("KeepTasks = true, want false (cursor at remove)")
	}
}

func TestDisconnectDialog_Cancel(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test Source")

	cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	if _, ok := msg.(DisconnectCancelledMsg); !ok {
		t.Fatalf("expected DisconnectCancelledMsg, got %T", msg)
	}
}

func TestDisconnectDialog_NonKeyMsg(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test")
	cmd := d.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

func TestDisconnectDialog_CursorIndicator(t *testing.T) {
	t.Parallel()

	d := NewDisconnectDialog("conn-1", "Test")
	d.SetWidth(80)

	// Cursor at 0 — "Keep tasks locally" should be highlighted
	view := d.View()
	if !strings.Contains(view, "> Keep tasks locally") {
		t.Error("View() missing cursor indicator on Keep tasks locally")
	}

	// Move down
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	view = d.View()
	if !strings.Contains(view, "> Remove synced tasks") {
		t.Error("View() missing cursor indicator on Remove synced tasks")
	}
}
