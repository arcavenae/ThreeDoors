package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func newDisconnectTestConnection(t *testing.T) *connection.Connection {
	t.Helper()
	mgr := connection.NewConnectionManager(nil)
	conn, err := mgr.Add("todoist", "Work Todoist", nil)
	if err != nil {
		t.Fatal(err)
	}
	conn.TaskCount = 15
	return conn
}

func TestDisconnectDialog_NewDisconnectDialog(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)
	if d == nil {
		t.Fatal("expected non-nil DisconnectDialog")
	}
	if d.conn != conn {
		t.Error("expected conn to be set")
	}
	if d.taskAction != "keep" {
		t.Errorf("expected default taskAction 'keep', got %q", d.taskAction)
	}
	if d.confirmed {
		t.Error("expected confirmed to be false initially")
	}
}

func TestDisconnectDialog_ViewRendersConnectionName(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)
	d.SetWidth(80)
	d.SetHeight(24)

	view := d.View()
	if !strings.Contains(view, "Disconnect") {
		t.Error("expected 'Disconnect' in header")
	}
	if !strings.Contains(view, "Work Todoist") {
		t.Error("expected connection label 'Work Todoist' in view")
	}
	if !strings.Contains(view, "todoist") {
		t.Error("expected provider name 'todoist' in view")
	}
}

func TestDisconnectDialog_ViewShowsTaskOptions(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)
	d.SetWidth(80)
	d.SetHeight(24)

	view := d.View()
	if !strings.Contains(view, "Keep tasks locally") {
		t.Error("expected 'Keep tasks locally' option in view")
	}
	if !strings.Contains(view, "Remove synced tasks") {
		t.Error("expected 'Remove synced tasks' option in view")
	}
}

func TestDisconnectDialog_EscCancels(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)

	cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	if _, ok := msg.(DisconnectCancelledMsg); !ok {
		t.Errorf("expected DisconnectCancelledMsg, got %T", msg)
	}
}

func TestDisconnectDialog_SetDimensions(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)
	d.SetWidth(100)
	d.SetHeight(40)
	if d.width != 100 {
		t.Errorf("expected width 100, got %d", d.width)
	}
	if d.height != 40 {
		t.Errorf("expected height 40, got %d", d.height)
	}
}

func TestDisconnectDialog_ViewShowsEscHint(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)
	d.SetWidth(80)

	view := d.View()
	if !strings.Contains(view, "esc:cancel") {
		t.Error("expected 'esc:cancel' hint in view")
	}
}

func TestDisconnectDialog_CompleteDialogKeepTasks(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)

	// Simulate confirmed with keep
	d.confirmed = true
	d.taskAction = "keep"

	cmd := d.completeDialog()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	confirmed, ok := msg.(DisconnectConfirmedMsg)
	if !ok {
		t.Fatalf("expected DisconnectConfirmedMsg, got %T", msg)
	}
	if confirmed.ConnectionID != conn.ID {
		t.Errorf("expected ConnectionID %s, got %s", conn.ID, confirmed.ConnectionID)
	}
	if !confirmed.KeepTasks {
		t.Error("expected KeepTasks to be true")
	}
}

func TestDisconnectDialog_CompleteDialogRemoveTasks(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)

	// Simulate confirmed with remove
	d.confirmed = true
	d.taskAction = "remove"

	cmd := d.completeDialog()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	confirmed, ok := msg.(DisconnectConfirmedMsg)
	if !ok {
		t.Fatalf("expected DisconnectConfirmedMsg, got %T", msg)
	}
	if confirmed.KeepTasks {
		t.Error("expected KeepTasks to be false")
	}
}

func TestDisconnectDialog_CompleteDialogNotConfirmedCancels(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)

	d.confirmed = false
	d.taskAction = "keep"

	cmd := d.completeDialog()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(DisconnectCancelledMsg); !ok {
		t.Errorf("expected DisconnectCancelledMsg, got %T", msg)
	}
}

func TestDisconnectDialog_NonKeyMsgPassedToForm(t *testing.T) {
	t.Parallel()
	conn := newDisconnectTestConnection(t)
	d := NewDisconnectDialog(conn)

	// A non-key message should be passed to the form without error
	cmd := d.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	// We just verify it doesn't panic; cmd may or may not be nil
	_ = cmd
}

func TestConnectionManager_Disconnect(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn, err := mgr.Add("todoist", "Test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Disconnect with keepTasks=true
	if err := mgr.Disconnect(conn.ID, true); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	// Connection should be gone
	if _, err := mgr.Get(conn.ID); err == nil {
		t.Error("expected connection to be removed after disconnect")
	}
}

func TestConnectionManager_DisconnectRemoveTasks(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn, err := mgr.Add("jira", "Work Jira", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := mgr.Disconnect(conn.ID, false); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if _, err := mgr.Get(conn.ID); err == nil {
		t.Error("expected connection to be removed after disconnect")
	}
}

func TestConnectionManager_DisconnectNotFound(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)

	err := mgr.Disconnect("nonexistent", true)
	if err == nil {
		t.Error("expected error for nonexistent connection")
	}
}
