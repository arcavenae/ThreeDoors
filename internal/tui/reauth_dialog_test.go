package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func TestReauthDialog_View(t *testing.T) {
	t.Parallel()

	conn := &connection.Connection{
		ID:           "conn-1",
		ProviderName: "todoist",
		Label:        "Personal",
		State:        connection.StateAuthExpired,
	}

	d := NewReauthDialog(conn, "Settings → Integrations → API token")
	d.SetWidth(80)
	d.SetHeight(24)

	view := d.View()

	if !strings.Contains(view, "Re-authenticate") {
		t.Error("view should contain 'Re-authenticate'")
	}
	if !strings.Contains(view, "Personal") {
		t.Error("view should contain connection label")
	}
	if !strings.Contains(view, "todoist") {
		t.Error("view should contain provider name")
	}
	if !strings.Contains(view, "esc:cancel") {
		t.Error("view should contain cancel hint")
	}
}

func TestReauthDialog_EscCancels(t *testing.T) {
	t.Parallel()

	conn := &connection.Connection{
		ID:           "conn-1",
		ProviderName: "todoist",
		Label:        "Personal",
		State:        connection.StateAuthExpired,
	}

	d := NewReauthDialog(conn, "")
	d.SetWidth(80)
	d.SetHeight(24)

	cmd := d.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}

	msg := cmd()
	if _, ok := msg.(ReauthCancelledMsg); !ok {
		t.Errorf("expected ReauthCancelledMsg, got %T", msg)
	}
}

func TestReauthDialog_Init(t *testing.T) {
	t.Parallel()

	conn := &connection.Connection{
		ID:           "conn-1",
		ProviderName: "todoist",
		Label:        "Personal",
	}

	d := NewReauthDialog(conn, "Token help text")

	// Init should return a command (from the huh form).
	cmd := d.Init()
	// The cmd may be nil or a batch cmd — just verify it doesn't panic.
	_ = cmd
}

func TestReauthDialog_ViewWithoutTokenHelp(t *testing.T) {
	t.Parallel()

	conn := &connection.Connection{
		ID:           "conn-2",
		ProviderName: "linear",
		Label:        "Work",
		State:        connection.StateAuthExpired,
	}

	d := NewReauthDialog(conn, "")
	d.SetWidth(80)
	d.SetHeight(24)

	view := d.View()
	if !strings.Contains(view, "Work") {
		t.Error("view should contain connection label")
	}
	if !strings.Contains(view, "linear") {
		t.Error("view should contain provider name")
	}
}
