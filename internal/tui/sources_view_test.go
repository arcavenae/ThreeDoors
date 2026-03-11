package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSourcesView_NewSourcesView(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)
	if sv == nil {
		t.Fatal("expected non-nil SourcesView")
	}
	if sv.connMgr != mgr {
		t.Error("expected connMgr to be set")
	}
	if sv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", sv.selectedIndex)
	}
}

func TestSourcesView_SetWidth(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	if sv.width != 80 {
		t.Errorf("expected width 80, got %d", sv.width)
	}
}

func TestSourcesView_SetHeight(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)
	sv.SetHeight(24)
	if sv.height != 24 {
		t.Errorf("expected height 24, got %d", sv.height)
	}
}

func TestSourcesView_ViewEmpty(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	sv.SetHeight(24)

	view := sv.View()
	if !strings.Contains(view, "No connections") {
		t.Errorf("expected empty state message, got:\n%s", view)
	}
	if !strings.Contains(view, "a") {
		t.Error("expected hint about 'a' key to add connection")
	}
}

func TestSourcesView_ViewWithConnections(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	if _, err := mgr.Add("todoist", "Personal Todoist", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Add("jira", "Work Jira", nil); err != nil {
		t.Fatal(err)
	}

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	sv.SetHeight(24)

	view := sv.View()
	if !strings.Contains(view, "Personal Todoist") {
		t.Error("expected 'Personal Todoist' in view")
	}
	if !strings.Contains(view, "Work Jira") {
		t.Error("expected 'Work Jira' in view")
	}
}

func TestSourcesView_StatusIndicators(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		state     connection.ConnectionState
		indicator string
		text      string
	}{
		{"connected shows green dot", connection.StateConnected, "●", "Synced"},
		{"paused shows gray circle", connection.StatePaused, "○", "Paused"},
		{"auth expired shows warning", connection.StateAuthExpired, "⚠", "Auth expired"},
		{"error shows red dot", connection.StateError, "●", "Error"},
		{"disconnected shows gray dot", connection.StateDisconnected, "○", "Disconnected"},
		{"connecting shows ellipsis", connection.StateConnecting, "…", "Connecting"},
		{"syncing shows sync indicator", connection.StateSyncing, "↻", "Syncing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator, text := statusIndicatorAndText(tt.state, time.Time{}, "")
			if indicator != tt.indicator {
				t.Errorf("expected indicator %q, got %q", tt.indicator, indicator)
			}
			if !strings.Contains(text, tt.text) {
				t.Errorf("expected text containing %q, got %q", tt.text, text)
			}
		})
	}
}

func TestSourcesView_SyncedTimeAgo(t *testing.T) {
	t.Parallel()
	lastSync := time.Now().UTC().Add(-5 * time.Minute)
	indicator, text := statusIndicatorAndText(connection.StateConnected, lastSync, "")
	if indicator != "●" {
		t.Errorf("expected ●, got %q", indicator)
	}
	if !strings.Contains(text, "5m ago") {
		t.Errorf("expected 'Synced 5m ago', got %q", text)
	}
}

func TestSourcesView_TaskCount(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn, err := mgr.Add("todoist", "My Tasks", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnected); err != nil {
		t.Fatal(err)
	}
	conn.TaskCount = 42

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	sv.SetHeight(24)

	view := sv.View()
	if !strings.Contains(view, "42") {
		t.Error("expected task count '42' in view")
	}
}

func TestSourcesView_KeybindingEsc(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestSourcesView_KeybindingNavigation(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	if _, err := mgr.Add("todoist", "Alpha", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Add("jira", "Beta", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Add("github", "Gamma", nil); err != nil {
		t.Fatal(err)
	}

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)

	if sv.selectedIndex != 0 {
		t.Fatalf("expected initial selectedIndex 0, got %d", sv.selectedIndex)
	}

	// j moves down
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sv.selectedIndex != 1 {
		t.Errorf("after j: expected 1, got %d", sv.selectedIndex)
	}

	// k moves up
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sv.selectedIndex != 0 {
		t.Errorf("after k: expected 0, got %d", sv.selectedIndex)
	}

	// k at top stays at 0
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sv.selectedIndex != 0 {
		t.Errorf("k at top: expected 0, got %d", sv.selectedIndex)
	}

	// j to bottom, then j again stays at bottom
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sv.selectedIndex != 2 {
		t.Errorf("j at bottom: expected 2, got %d", sv.selectedIndex)
	}
}

func TestSourcesView_KeybindingA(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'a' key")
	}
	msg := cmd()
	if _, ok := msg.(ShowConnectWizardMsg); !ok {
		t.Errorf("expected ShowConnectWizardMsg, got %T", msg)
	}
}

func addConnectedConnection(t *testing.T, mgr *connection.ConnectionManager, provider, label string) *connection.Connection {
	t.Helper()
	conn, err := mgr.Add(provider, label, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnected); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestSourcesView_KeybindingP(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	addConnectedConnection(t, mgr, "todoist", "Test")

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'p' key")
	}
	msg := cmd()
	if m, ok := msg.(SourceActionMsg); !ok {
		t.Errorf("expected SourceActionMsg, got %T", msg)
	} else if m.Action != "toggle_pause" {
		t.Errorf("expected action toggle_pause, got %s", m.Action)
	}
}

func TestSourcesView_KeybindingR(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	addConnectedConnection(t, mgr, "todoist", "Test")

	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'r' key")
	}
	msg := cmd()
	if m, ok := msg.(SourceActionMsg); !ok {
		t.Errorf("expected SourceActionMsg, got %T", msg)
	} else if m.Action != "resync" {
		t.Errorf("expected action resync, got %s", m.Action)
	}
}

func TestSourcesView_KeybindingT(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	addConnectedConnection(t, mgr, "todoist", "Test")

	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 't' key")
	}
	msg := cmd()
	if m, ok := msg.(SourceActionMsg); !ok {
		t.Errorf("expected SourceActionMsg, got %T", msg)
	} else if m.Action != "test" {
		t.Errorf("expected action test, got %s", m.Action)
	}
}

func TestSourcesView_KeybindingD(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	addConnectedConnection(t, mgr, "todoist", "Test")

	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'd' key")
	}
	msg := cmd()
	if m, ok := msg.(SourceActionMsg); !ok {
		t.Errorf("expected SourceActionMsg, got %T", msg)
	} else if m.Action != "disconnect" {
		t.Errorf("expected action disconnect, got %s", m.Action)
	}
}

func TestSourcesView_KeybindingEnter(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	if _, err := mgr.Add("todoist", "Test", nil); err != nil {
		t.Fatal(err)
	}

	sv := NewSourcesView(mgr)

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Enter")
	}
	msg := cmd()
	if m, ok := msg.(ShowSourceDetailMsg); !ok {
		t.Errorf("expected ShowSourceDetailMsg, got %T", msg)
	} else if m.ConnectionID == "" {
		t.Error("expected non-empty ConnectionID")
	}
}

func TestSourcesView_SelectedHighlighted(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	if _, err := mgr.Add("todoist", "Alpha", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Add("jira", "Beta", nil); err != nil {
		t.Fatal(err)
	}

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	sv.SetHeight(24)
	sv.selectedIndex = 1

	view := sv.View()
	if !strings.Contains(view, "Alpha") || !strings.Contains(view, "Beta") {
		t.Error("expected both connection labels in view")
	}
}

func TestSourcesView_RenderAtVariousWidths(t *testing.T) {
	t.Parallel()
	widths := []int{30, 60, 80, 120}

	for _, w := range widths {
		w := w
		t.Run(fmt.Sprintf("width_%d", w), func(t *testing.T) {
			t.Parallel()
			mgr := connection.NewConnectionManager(nil)
			if _, err := mgr.Add("todoist", "Personal", nil); err != nil {
				t.Fatal(err)
			}
			if _, err := mgr.Add("jira", "Work", nil); err != nil {
				t.Fatal(err)
			}

			sv := NewSourcesView(mgr)
			sv.SetWidth(w)
			sv.SetHeight(24)

			view := sv.View()
			if view == "" {
				t.Errorf("expected non-empty view at width %d", w)
			}
		})
	}
}

func TestSourcesView_NoActionOnEmptyList(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	sv := NewSourcesView(mgr)

	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'p'}},
		{Type: tea.KeyRunes, Runes: []rune{'r'}},
		{Type: tea.KeyRunes, Runes: []rune{'t'}},
		{Type: tea.KeyRunes, Runes: []rune{'d'}},
		{Type: tea.KeyEnter},
	}
	for _, k := range keys {
		cmd := sv.Update(k)
		if cmd != nil {
			t.Errorf("expected nil cmd for key %v on empty list", k)
		}
	}
}

func TestSourcesView_AuthExpiredShowsZeroTasks(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn, err := mgr.Add("jira", "Expired Jira", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatal(err)
	}
	if err := mgr.TransitionWithError(conn.ID, connection.StateAuthExpired, "token expired"); err != nil {
		t.Fatal(err)
	}

	sv := NewSourcesView(mgr)
	sv.SetWidth(80)
	sv.SetHeight(24)

	view := sv.View()
	if !strings.Contains(view, "Auth expired") {
		t.Error("expected 'Auth expired' text")
	}
	if !strings.Contains(view, "⚠") {
		t.Error("expected ⚠ indicator")
	}
}
