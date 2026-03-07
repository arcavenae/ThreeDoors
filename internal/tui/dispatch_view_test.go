package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/dispatch"
	tea "github.com/charmbracelet/bubbletea"
)

func newDispatchDetailView(text string) *DetailView {
	task := core.NewTask(text)
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetDevDispatchInfo(true, true)
	dv.SetWidth(80)
	return dv
}

// --- Dispatch Key Binding ---

func TestDetailView_XKey_DispatchEnabled_EntersConfirmMode(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("implement feature X")
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Error("'x' entering confirm mode should not return a command")
	}
	if dv.mode != DetailModeDispatchConfirm {
		t.Errorf("expected DetailModeDispatchConfirm, got %d", dv.mode)
	}
}

func TestDetailView_XKey_DispatchDisabled_NoAction(t *testing.T) {
	t.Parallel()
	task := core.NewTask("test task")
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)
	// dispatch not enabled, no cross-refs either
	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Error("'x' with dispatch disabled and no cross-refs should return nil")
	}
	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView, got %d", dv.mode)
	}
}

func TestDetailView_XKey_AlreadyDispatched_ShowsFlash(t *testing.T) {
	t.Parallel()
	task := core.NewTask("already dispatched task")
	now := time.Now().UTC()
	task.DevDispatch = &dispatch.DevDispatch{
		Queued:   true,
		QueuedAt: &now,
	}
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetDevDispatchInfo(true, true)
	dv.SetWidth(80)

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd == nil {
		t.Fatal("'x' on already-dispatched task should return a flash command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "already dispatched") {
		t.Errorf("expected 'already dispatched' message, got %q", fm.Text)
	}
}

// --- Confirmation Flow ---

func TestDetailView_DispatchConfirm_YKey_SendsDispatchRequest(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("implement feature X")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if cmd == nil {
		t.Fatal("'y' in confirm mode should return a command")
	}
	msg := cmd()
	drm, ok := msg.(DevDispatchRequestMsg)
	if !ok {
		t.Fatalf("expected DevDispatchRequestMsg, got %T", msg)
	}
	if drm.Task.Text != "implement feature X" {
		t.Errorf("expected task text 'implement feature X', got %q", drm.Task.Text)
	}
	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after confirm, got %d", dv.mode)
	}
}

func TestDetailView_DispatchConfirm_NKey_Cancels(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("implement feature X")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd != nil {
		t.Error("'n' in confirm mode should not return a command")
	}
	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after cancel, got %d", dv.mode)
	}
}

func TestDetailView_DispatchConfirm_EscKey_Cancels(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("implement feature X")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	cmd := dv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Error("Esc in confirm mode should not return a command")
	}
	if dv.mode != DetailModeView {
		t.Errorf("expected DetailModeView after escape, got %d", dv.mode)
	}
}

// --- View Rendering ---

func TestDetailView_DispatchConfirm_ShowsPrompt(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("implement feature X")
	dv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	view := dv.View()
	if !strings.Contains(view, "Dispatch") {
		t.Error("confirm mode should show 'Dispatch' prompt")
	}
	if !strings.Contains(view, "implement feature X") {
		t.Error("confirm mode should show task text")
	}
	if !strings.Contains(view, "y/n") {
		t.Error("confirm mode should show y/n options")
	}
}

func TestDetailView_DispatchEnabled_ShowsHint(t *testing.T) {
	t.Parallel()
	dv := newDispatchDetailView("test task")
	view := dv.View()
	if !strings.Contains(view, "[X]dispatch") {
		t.Error("help text should contain [X]dispatch when dispatch is enabled")
	}
}

func TestDetailView_DispatchDisabled_NoHint(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "[X]dispatch") {
		t.Error("help text should NOT contain [X]dispatch when dispatch is disabled")
	}
}

func TestDetailView_DevBadge_ShownWhenDispatched(t *testing.T) {
	t.Parallel()
	task := core.NewTask("dispatched task")
	now := time.Now().UTC()
	task.DevDispatch = &dispatch.DevDispatch{
		Queued:   true,
		QueuedAt: &now,
	}
	dv := NewDetailView(task, nil, nil, nil)
	dv.SetWidth(80)

	view := dv.View()
	if !strings.Contains(view, "QUEUED") {
		t.Error("detail view should show [QUEUED] badge for dispatched task")
	}
}

func TestDetailView_DevBadge_NotShownWhenNotDispatched(t *testing.T) {
	t.Parallel()
	dv := newTestDetailView("test task")
	dv.SetWidth(80)
	view := dv.View()
	if strings.Contains(view, "DEV") {
		t.Error("detail view should NOT show [DEV] badge for non-dispatched task")
	}
}

// --- Badge Rendering ---

func TestDevDispatchBadge_QueuedTask(t *testing.T) {
	t.Parallel()
	task := core.NewTask("test task")
	task.DevDispatch = &dispatch.DevDispatch{Queued: true}
	badge := DevDispatchBadge(task)
	if !strings.Contains(badge, "QUEUED") {
		t.Errorf("DevDispatchBadge() = %q, want to contain 'QUEUED'", badge)
	}
}

func TestDoorsView_ShowsDevBadge(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	task := core.NewTask("dispatched task")
	now := time.Now().UTC()
	task.DevDispatch = &dispatch.DevDispatch{
		Queued:   true,
		QueuedAt: &now,
	}
	pool.AddTask(task)

	// Add filler tasks
	for i := 0; i < 2; i++ {
		pool.AddTask(core.NewTask("filler task"))
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	view := dv.View()

	if !strings.Contains(view, "QUEUED") {
		t.Error("doors view should show [QUEUED] badge for dispatched task")
	}
}

func TestDoorsView_NoDevBadge_WhenNotDispatched(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	for i := 0; i < 3; i++ {
		pool.AddTask(core.NewTask("regular task"))
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	view := dv.View()

	if strings.Contains(view, "QUEUED") {
		t.Error("doors view should NOT show [QUEUED] badge for non-dispatched tasks")
	}
}

// --- Search View :dispatch command ---

func TestSearchView_DispatchCommand_EnabledShowsHint(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	pool.AddTask(core.NewTask("task"))
	sv := NewSearchView(pool, nil, nil, nil, nil)
	sv.SetDevDispatchEnabled(true)

	sv.textInput.SetValue(":dispatch")
	sv.checkCommandMode()

	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":dispatch should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "detail view") {
		t.Errorf("expected hint about detail view, got %q", fm.Text)
	}
}

func TestSearchView_DispatchCommand_DisabledShowsError(t *testing.T) {
	t.Parallel()
	pool := core.NewTaskPool()
	pool.AddTask(core.NewTask("task"))
	sv := NewSearchView(pool, nil, nil, nil, nil)
	// devDispatchEnabled defaults to false

	sv.textInput.SetValue(":dispatch")
	sv.checkCommandMode()

	cmd := sv.executeCommand()
	if cmd == nil {
		t.Fatal(":dispatch should return a command")
	}
	msg := cmd()
	fm, ok := msg.(FlashMsg)
	if !ok {
		t.Fatalf("expected FlashMsg, got %T", msg)
	}
	if !strings.Contains(fm.Text, "not enabled") {
		t.Errorf("expected 'not enabled' message, got %q", fm.Text)
	}
}

// --- Config ---

func TestDevDispatchEnabled_InConfig(t *testing.T) {
	t.Parallel()
	cfg := &core.ProviderConfig{
		DevDispatchEnabled: true,
	}
	if !cfg.DevDispatchEnabled {
		t.Error("DevDispatchEnabled should be true")
	}
}
