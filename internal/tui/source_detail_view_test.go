package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestConnection(t *testing.T, mgr *connection.ConnectionManager, provider, label string, settings map[string]string) *connection.Connection {
	t.Helper()
	conn, err := mgr.Add(provider, label, settings)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func connectedTestConnection(t *testing.T, mgr *connection.ConnectionManager, provider, label string, settings map[string]string) *connection.Connection {
	t.Helper()
	conn := newTestConnection(t, mgr, provider, label, settings)
	if err := mgr.Transition(conn.ID, connection.StateConnecting); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Transition(conn.ID, connection.StateConnected); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestSourceDetailView_New(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := newTestConnection(t, mgr, "todoist", "My Todoist", nil)

	dv := NewSourceDetailView(conn, mgr)
	if dv == nil {
		t.Fatal("expected non-nil SourceDetailView")
	}
	if dv.conn != conn {
		t.Error("expected conn to be set")
	}
	if dv.connMgr != mgr {
		t.Error("expected connMgr to be set")
	}
}

func TestSourceDetailView_SetDimensions(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := newTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	dv.SetWidth(100)
	dv.SetHeight(40)
	if dv.width != 100 {
		t.Errorf("expected width 100, got %d", dv.width)
	}
	if dv.height != 40 {
		t.Errorf("expected height 40, got %d", dv.height)
	}
}

func TestSourceDetailView_ViewDisconnected(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := newTestConnection(t, mgr, "todoist", "Personal Todoist", map[string]string{
		"workspace": "personal",
	})

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()

	// Should show connection label
	if !strings.Contains(view, "Personal Todoist") {
		t.Error("expected connection label in view")
	}
	// Should show provider
	if !strings.Contains(view, "todoist") {
		t.Error("expected provider name in view")
	}
	// Should show state
	if !strings.Contains(view, "disconnected") && !strings.Contains(view, "Disconnected") {
		t.Error("expected disconnected state in view")
	}
	// Should show sync mode
	if !strings.Contains(view, "readonly") {
		t.Error("expected sync mode in view")
	}
	// Should show settings
	if !strings.Contains(view, "workspace") {
		t.Error("expected settings key in view")
	}
}

func TestSourceDetailView_ViewConnected(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "jira", "Work Jira", map[string]string{
		"project": "PROJ-1",
	})
	conn.TaskCount = 15
	conn.LastSync = time.Now().UTC().Add(-3 * time.Minute)

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()

	if !strings.Contains(view, "Work Jira") {
		t.Error("expected connection label in view")
	}
	if !strings.Contains(view, "jira") {
		t.Error("expected provider name in view")
	}
	if !strings.Contains(view, "15") {
		t.Error("expected task count in view")
	}
}

func TestSourceDetailView_ViewWithHealthCheck(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test Todoist", nil)

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	// Simulate receiving health check results
	dv.healthResult = &connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
		TaskCount:    42,
	}

	view := dv.View()

	if !strings.Contains(view, "API") {
		t.Error("expected health check API label in view")
	}
	if !strings.Contains(view, "Token") {
		t.Error("expected health check Token label in view")
	}
	if !strings.Contains(view, "Rate") {
		t.Error("expected health check Rate limit label in view")
	}
	// Check indicators — passing checks show ✓
	if !strings.Contains(view, "✓") {
		t.Error("expected ✓ indicator for passing health checks")
	}
}

func TestSourceDetailView_ViewWithFailedHealthCheck(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Broken Source", nil)

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	dv.healthResult = &connection.HealthCheckResult{
		APIReachable: false,
		TokenValid:   false,
		RateLimitOK:  true,
		TaskCount:    0,
	}

	view := dv.View()

	if !strings.Contains(view, "✗") {
		t.Error("expected ✗ indicator for failing health checks")
	}
	if !strings.Contains(view, "✓") {
		t.Error("expected ✓ for passing rate limit check")
	}
}

func TestSourceDetailView_KeybindingEsc(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := newTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for Esc")
	}
	msg := cmd()
	if _, ok := msg.(ShowSourcesMsg); !ok {
		t.Errorf("expected ShowSourcesMsg, got %T", msg)
	}
}

func TestSourceDetailView_KeybindingP(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'p'")
	}
	msg := cmd()
	m, ok := msg.(SourceActionMsg)
	if !ok {
		t.Fatalf("expected SourceActionMsg, got %T", msg)
	}
	if m.Action != "toggle_pause" {
		t.Errorf("expected toggle_pause, got %s", m.Action)
	}
	if m.ConnectionID != conn.ID {
		t.Errorf("expected conn ID %s, got %s", conn.ID, m.ConnectionID)
	}
}

func TestSourceDetailView_KeybindingD(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'd'")
	}
	msg := cmd()
	m, ok := msg.(SourceActionMsg)
	if !ok {
		t.Fatalf("expected SourceActionMsg, got %T", msg)
	}
	if m.Action != "disconnect" {
		t.Errorf("expected disconnect, got %s", m.Action)
	}
}

func TestSourceDetailView_KeybindingE(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'e'")
	}
	msg := cmd()
	m, ok := msg.(SourceActionMsg)
	if !ok {
		t.Fatalf("expected SourceActionMsg, got %T", msg)
	}
	if m.Action != "edit" {
		t.Errorf("expected edit, got %s", m.Action)
	}
}

func TestSourceDetailView_KeybindingR(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'r'")
	}
	msg := cmd()
	m, ok := msg.(SourceActionMsg)
	if !ok {
		t.Fatalf("expected SourceActionMsg, got %T", msg)
	}
	if m.Action != "reauth" {
		t.Errorf("expected reauth, got %s", m.Action)
	}
}

func TestSourceDetailView_KeybindingL(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for 'l'")
	}
	msg := cmd()
	m, ok := msg.(SourceActionMsg)
	if !ok {
		t.Fatalf("expected SourceActionMsg, got %T", msg)
	}
	if m.Action != "sync_log" {
		t.Errorf("expected sync_log, got %s", m.Action)
	}
}

func TestSourceDetailView_HealthCheckResultMsg(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	result := connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  false,
		TaskCount:    10,
	}

	cmd := dv.Update(SourceHealthCheckResultMsg{
		ConnectionID: conn.ID,
		Result:       result,
	})
	_ = cmd

	if dv.healthResult == nil {
		t.Fatal("expected healthResult to be set")
	}
	if dv.healthResult.RateLimitOK {
		t.Error("expected RateLimitOK to be false")
	}
	if dv.healthResult.TaskCount != 10 {
		t.Errorf("expected TaskCount 10, got %d", dv.healthResult.TaskCount)
	}
}

func TestSourceDetailView_HealthCheckErrorMsg(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	dv.Update(SourceHealthCheckResultMsg{
		ConnectionID: conn.ID,
		Err:          "connection timed out",
	})

	if dv.healthError == "" {
		t.Error("expected healthError to be set")
	}

	dv.SetWidth(100)
	dv.SetHeight(40)
	view := dv.View()
	if !strings.Contains(view, "timed out") {
		t.Errorf("expected error message in view, got:\n%s", view)
	}
}

func TestSourceDetailView_VariousWidths(t *testing.T) {
	t.Parallel()
	widths := []int{30, 60, 80, 120}

	for _, w := range widths {
		t.Run("width", func(t *testing.T) {
			t.Parallel()
			mgr := connection.NewConnectionManager(nil)
			conn := connectedTestConnection(t, mgr, "todoist", "Test Source", nil)
			conn.TaskCount = 5

			dv := NewSourceDetailView(conn, mgr)
			dv.SetWidth(w)
			dv.SetHeight(24)

			view := dv.View()
			if view == "" {
				t.Errorf("expected non-empty view at width %d", w)
			}
			if !strings.Contains(view, "Test Source") {
				t.Errorf("expected label at width %d", w)
			}
		})
	}
}

func TestSourceDetailView_PausedState(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Paused Source", nil)
	if err := mgr.Transition(conn.ID, connection.StatePaused); err != nil {
		t.Fatal(err)
	}

	// Re-fetch connection after transition
	conn, err := mgr.Get(conn.ID)
	if err != nil {
		t.Fatal(err)
	}

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()
	if !strings.Contains(view, "paused") && !strings.Contains(view, "Paused") {
		t.Error("expected paused state in view")
	}
}

func TestSourceDetailView_ErrorState(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "jira", "Broken Jira", nil)
	if err := mgr.TransitionWithError(conn.ID, connection.StateError, "API timeout"); err != nil {
		t.Fatal(err)
	}

	conn, err := mgr.Get(conn.ID)
	if err != nil {
		t.Fatal(err)
	}

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()
	if !strings.Contains(view, "API timeout") {
		t.Error("expected error message in view")
	}
}

func TestSourceDetailView_SettingsDisplay(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "jira", "Work Jira", map[string]string{
		"project": "PROJ-1",
		"jql":     "assignee = me",
	})

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()
	if !strings.Contains(view, "project") {
		t.Error("expected setting key 'project' in view")
	}
	if !strings.Contains(view, "PROJ-1") {
		t.Error("expected setting value 'PROJ-1' in view")
	}
	if !strings.Contains(view, "jql") {
		t.Error("expected setting key 'jql' in view")
	}
}

func TestSourceDetailView_FooterKeybindings(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Test", nil)

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	view := dv.View()
	// Should show key hints
	if !strings.Contains(view, "esc") {
		t.Error("expected esc hint in footer")
	}
}

func TestSourceDetailView_NonKeyMsgReturnsNil(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := newTestConnection(t, mgr, "todoist", "Test", nil)
	dv := NewSourceDetailView(conn, mgr)

	cmd := dv.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Error("expected nil cmd for non-key message")
	}
}

func TestSourceDetailView_ConnectionRefresh(t *testing.T) {
	t.Parallel()
	mgr := connection.NewConnectionManager(nil)
	conn := connectedTestConnection(t, mgr, "todoist", "Refresh Test", nil)

	dv := NewSourceDetailView(conn, mgr)
	dv.SetWidth(100)
	dv.SetHeight(40)

	// Change connection task count externally
	conn.TaskCount = 99

	// View should reflect the current connection state
	view := dv.View()
	if !strings.Contains(view, "99") {
		t.Error("expected updated task count in view")
	}
}
