package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

// stubConnectService implements ConnectServicer for testing.
type stubConnectService struct {
	addCalled  bool
	testCalled bool
	addErr     error
	testErr    error
	testResult connection.HealthCheckResult
	lastAdd    struct {
		provider   string
		label      string
		settings   map[string]string
		credential string
	}
}

func (s *stubConnectService) Add(providerName, label string, settings map[string]string, credential string) (*connection.Connection, error) {
	s.addCalled = true
	s.lastAdd.provider = providerName
	s.lastAdd.label = label
	s.lastAdd.settings = settings
	s.lastAdd.credential = credential
	if s.addErr != nil {
		return nil, s.addErr
	}
	conn, err := connection.NewConnection(providerName, label, settings)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *stubConnectService) TestConnection(id string) (connection.HealthCheckResult, error) {
	s.testCalled = true
	return s.testResult, s.testErr
}

func newConnectTestRegistry() *core.Registry {
	reg := core.NewRegistry()
	_ = reg.Register("todoist", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return nil, fmt.Errorf("not implemented")
	})
	_ = reg.Register("github", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return nil, fmt.Errorf("not implemented")
	})
	_ = reg.Register("textfile", func(config *core.ProviderConfig) (core.TaskProvider, error) {
		return nil, fmt.Errorf("not implemented")
	})
	return reg
}

func TestNewConnectWizard(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)

	if w == nil {
		t.Fatal("NewConnectWizard returned nil")
	}
	if w.phase != wizardPhaseForm {
		t.Errorf("initial phase = %d, want %d (wizardPhaseForm)", w.phase, wizardPhaseForm)
	}
	if len(w.providers) != 3 {
		t.Errorf("providers count = %d, want 3", len(w.providers))
	}
	if w.form == nil {
		t.Error("form is nil")
	}
	if w.syncMode != "readonly" {
		t.Errorf("default syncMode = %q, want %q", w.syncMode, "readonly")
	}
	if w.pollInterval != "5m" {
		t.Errorf("default pollInterval = %q, want %q", w.pollInterval, "5m")
	}
}

func TestConnectWizard_SetFormSpec(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)

	spec := connection.FormSpec{
		AuthType:    connection.AuthAPIToken,
		Description: "Todoist task manager",
		TokenHelp:   "Settings > Integrations > API token",
	}

	w.SetFormSpec("todoist", spec)

	stored, ok := w.formSpecs["todoist"]
	if !ok {
		t.Fatal("FormSpec not stored for todoist")
	}
	if stored.AuthType != connection.AuthAPIToken {
		t.Errorf("stored AuthType = %d, want %d", stored.AuthType, connection.AuthAPIToken)
	}

	// Verify provider choice was updated.
	for _, p := range w.providers {
		if p.Name == "todoist" {
			if p.Description != "Todoist task manager" {
				t.Errorf("provider description = %q, want %q", p.Description, "Todoist task manager")
			}
			if p.AuthType != connection.AuthAPIToken {
				t.Errorf("provider AuthType = %d, want %d", p.AuthType, connection.AuthAPIToken)
			}
		}
	}
}

func TestConnectWizard_CancelFromForm(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.Init()

	// Press Esc to cancel.
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected command from Esc, got nil")
	}

	msg := cmd()
	if _, ok := msg.(ConnectWizardCancelMsg); !ok {
		t.Errorf("expected ConnectWizardCancelMsg, got %T", msg)
	}

	// Verify no connection was created.
	if svc.addCalled {
		t.Error("Add was called after cancel")
	}
}

func TestConnectWizard_CancelFromTestPhase(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseTesting

	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected command from Esc during testing, got nil")
	}

	msg := cmd()
	if _, ok := msg.(ConnectWizardCancelMsg); !ok {
		t.Errorf("expected ConnectWizardCancelMsg, got %T", msg)
	}
}

func TestConnectWizard_CancelFromResultPhase(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseResult

	// Both 'n' and Esc should cancel.
	for _, key := range []string{"n", "esc"} {
		w.phase = wizardPhaseResult
		var msg tea.Msg
		if key == "esc" {
			_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEsc})
			if cmd == nil {
				t.Fatalf("expected command from %s, got nil", key)
			}
			msg = cmd()
		} else {
			_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
			if cmd == nil {
				t.Fatalf("expected command from %s, got nil", key)
			}
			msg = cmd()
		}
		if _, ok := msg.(ConnectWizardCancelMsg); !ok {
			t.Errorf("key %q: expected ConnectWizardCancelMsg, got %T", key, msg)
		}
	}
}

func TestConnectWizard_ConfirmFromResultPhase(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseResult

	// Enter should finalize.
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected command from Enter, got nil")
	}

	msg := cmd()
	if _, ok := msg.(ConnectWizardCompleteMsg); !ok {
		t.Errorf("expected ConnectWizardCompleteMsg, got %T", msg)
	}
}

func TestConnectWizard_TestResultMsg_Success(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseTesting

	result := connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
		TaskCount:    42,
	}

	w.Update(connectTestResultMsg{result: result})

	if w.phase != wizardPhaseResult {
		t.Errorf("phase = %d, want %d (wizardPhaseResult)", w.phase, wizardPhaseResult)
	}
	if w.testResult == nil {
		t.Fatal("testResult is nil")
	}
	if !w.testResult.Healthy() {
		t.Error("testResult.Healthy() = false, want true")
	}
	if w.testResult.TaskCount != 42 {
		t.Errorf("TaskCount = %d, want 42", w.testResult.TaskCount)
	}
	if w.testErr != nil {
		t.Errorf("testErr = %v, want nil", w.testErr)
	}
}

func TestConnectWizard_TestResultMsg_Error(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseTesting

	w.Update(connectTestResultMsg{err: fmt.Errorf("connection refused")})

	if w.phase != wizardPhaseResult {
		t.Errorf("phase = %d, want %d (wizardPhaseResult)", w.phase, wizardPhaseResult)
	}
	if w.testErr == nil {
		t.Fatal("testErr is nil, want error")
	}
	if !strings.Contains(w.testErr.Error(), "connection refused") {
		t.Errorf("testErr = %q, want to contain 'connection refused'", w.testErr.Error())
	}
}

func TestConnectWizard_View_FormPhase(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.Init()

	view := w.View()

	if !strings.Contains(view, "Connect Data Source") {
		t.Error("view missing title 'Connect Data Source'")
	}
}

func TestConnectWizard_View_TestingPhase(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseTesting
	w.selectedProvider = "todoist"

	view := w.View()

	if !strings.Contains(view, "Testing connection") {
		t.Error("view missing 'Testing connection'")
	}
	if !strings.Contains(view, "todoist") {
		t.Error("view missing provider name")
	}
}

func TestConnectWizard_View_ResultPhase_Success(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseResult
	w.selectedProvider = "todoist"
	w.label = "My Tasks"
	w.syncMode = "readonly"
	w.pollInterval = "5m"
	w.testResult = &connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
		TaskCount:    10,
	}

	view := w.View()

	if !strings.Contains(view, "API Reachable") {
		t.Error("view missing 'API Reachable'")
	}
	if !strings.Contains(view, "Token Valid") {
		t.Error("view missing 'Token Valid'")
	}
	if !strings.Contains(view, "All checks passed") {
		t.Error("view missing 'All checks passed'")
	}
	if !strings.Contains(view, "todoist") {
		t.Error("view missing provider name")
	}
	if !strings.Contains(view, "My Tasks") {
		t.Error("view missing label")
	}
}

func TestConnectWizard_View_ResultPhase_Error(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseResult
	w.testErr = fmt.Errorf("auth failed")

	view := w.View()

	if !strings.Contains(view, "Connection test failed") {
		t.Error("view missing 'Connection test failed'")
	}
	if !strings.Contains(view, "auth failed") {
		t.Error("view missing error message")
	}
}

func TestConnectWizard_AuthTypeForSelected(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}

	tests := []struct {
		name     string
		provider string
		spec     *connection.FormSpec
		want     connection.AuthType
	}{
		{
			name:     "known API token provider",
			provider: "todoist",
			want:     connection.AuthAPIToken,
		},
		{
			name:     "known OAuth provider",
			provider: "github",
			want:     connection.AuthOAuth,
		},
		{
			name:     "known local path provider",
			provider: "textfile",
			want:     connection.AuthLocalPath,
		},
		{
			name:     "unknown provider defaults to AuthNone",
			provider: "unknown",
			want:     connection.AuthNone,
		},
		{
			name:     "FormSpec overrides default",
			provider: "todoist",
			spec:     &connection.FormSpec{AuthType: connection.AuthOAuth},
			want:     connection.AuthOAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w2 := NewConnectWizard(reg, svc)
			if tt.spec != nil {
				w2.formSpecs[tt.provider] = *tt.spec
			}
			w2.selectedProvider = tt.provider
			got := w2.authTypeForSelected()
			if got != tt.want {
				t.Errorf("authTypeForSelected() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestConnectWizard_BuildSettings(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)

	w.syncMode = "bidirectional"
	w.pollInterval = "15m"
	w.filterValue = "project:work"
	w.localPath = "/tmp/tasks.yaml"

	settings := w.buildSettings()

	if settings["sync_mode"] != "bidirectional" {
		t.Errorf("sync_mode = %q, want %q", settings["sync_mode"], "bidirectional")
	}
	if settings["poll_interval"] != "15m" {
		t.Errorf("poll_interval = %q, want %q", settings["poll_interval"], "15m")
	}
	if settings["filter"] != "project:work" {
		t.Errorf("filter = %q, want %q", settings["filter"], "project:work")
	}
	if settings["path"] != "/tmp/tasks.yaml" {
		t.Errorf("path = %q, want %q", settings["path"], "/tmp/tasks.yaml")
	}
}

func TestConnectWizard_BuildSettings_EmptyOptionals(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)

	w.syncMode = "readonly"
	w.pollInterval = "5m"
	// Leave filterValue and localPath empty.

	settings := w.buildSettings()

	if _, ok := settings["filter"]; ok {
		t.Error("filter should be absent when empty")
	}
	if _, ok := settings["path"]; ok {
		t.Error("path should be absent when empty")
	}
}

func TestConnectWizard_RunConnectionTest_NoService(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	w := NewConnectWizard(reg, nil)

	cmd := w.runConnectionTest()
	msg := cmd()

	result, ok := msg.(connectTestResultMsg)
	if !ok {
		t.Fatalf("expected connectTestResultMsg, got %T", msg)
	}
	if result.err == nil {
		t.Error("expected error when service is nil")
	}
}

func TestConnectWizard_RunConnectionTest_AddError(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{addErr: fmt.Errorf("storage full")}
	w := NewConnectWizard(reg, svc)
	w.selectedProvider = "todoist"
	w.label = "Test"

	cmd := w.runConnectionTest()
	msg := cmd()

	result, ok := msg.(connectTestResultMsg)
	if !ok {
		t.Fatalf("expected connectTestResultMsg, got %T", msg)
	}
	if result.err == nil {
		t.Error("expected error from Add failure")
	}
	if !strings.Contains(result.err.Error(), "storage full") {
		t.Errorf("error = %q, want to contain 'storage full'", result.err.Error())
	}
}

func TestConnectWizard_RunConnectionTest_Success(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{
		testResult: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
			TaskCount:    5,
		},
	}
	w := NewConnectWizard(reg, svc)
	w.selectedProvider = "todoist"
	w.label = "Work"
	w.apiToken = "test-token-123"
	w.syncMode = "readonly"
	w.pollInterval = "5m"

	cmd := w.runConnectionTest()
	msg := cmd()

	result, ok := msg.(connectTestResultMsg)
	if !ok {
		t.Fatalf("expected connectTestResultMsg, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("unexpected error: %v", result.err)
	}
	if !result.result.Healthy() {
		t.Error("expected healthy result")
	}
	if result.result.TaskCount != 5 {
		t.Errorf("TaskCount = %d, want 5", result.result.TaskCount)
	}

	// Verify Add was called with correct args.
	if !svc.addCalled {
		t.Error("Add was not called")
	}
	if svc.lastAdd.provider != "todoist" {
		t.Errorf("Add provider = %q, want %q", svc.lastAdd.provider, "todoist")
	}
	if svc.lastAdd.label != "Work" {
		t.Errorf("Add label = %q, want %q", svc.lastAdd.label, "Work")
	}
	if svc.lastAdd.credential != "test-token-123" {
		t.Errorf("Add credential = %q, want %q", svc.lastAdd.credential, "test-token-123")
	}
}

func TestParsePollInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantMins int
	}{
		{"1m", "1m", 1},
		{"5m", "5m", 5},
		{"15m", "15m", 15},
		{"30m", "30m", 30},
		{"1h", "1h", 60},
		{"invalid defaults to 5m", "invalid", 5},
		{"empty defaults to 5m", "", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parsePollInterval(tt.input)
			gotMins := int(got.Minutes())
			if gotMins != tt.wantMins {
				t.Errorf("parsePollInterval(%q) = %d min, want %d min", tt.input, gotMins, tt.wantMins)
			}
		})
	}
}

func TestConnectWizard_CommandRegistryHasConnect(t *testing.T) {
	t.Parallel()

	found := false
	for _, cmd := range commandRegistry {
		if cmd.Name == "connect" {
			found = true
			if cmd.Desc == "" {
				t.Error(":connect command has empty description")
			}
			break
		}
	}
	if !found {
		t.Error(":connect command not found in commandRegistry")
	}
}

func TestConnectWizard_SetDimensions(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	w := NewConnectWizard(reg, nil)
	w.SetWidth(80)
	w.SetHeight(24)

	if w.width != 80 {
		t.Errorf("width = %d, want 80", w.width)
	}
	if w.height != 24 {
		t.Errorf("height = %d, want 24", w.height)
	}
}

func TestConnectWizard_View_ResultPhase_PartialFailure(t *testing.T) {
	t.Parallel()

	reg := newConnectTestRegistry()
	svc := &stubConnectService{}
	w := NewConnectWizard(reg, svc)
	w.phase = wizardPhaseResult
	w.selectedProvider = "todoist"
	w.label = "Work"
	w.syncMode = "readonly"
	w.pollInterval = "5m"
	w.testResult = &connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   false, // token invalid
		RateLimitOK:  true,
		TaskCount:    0,
	}

	view := w.View()

	if !strings.Contains(view, "Some checks failed") {
		t.Error("view should show 'Some checks failed' for partial failure")
	}
}

func TestViewMode_ConnectWizard(t *testing.T) {
	t.Parallel()

	if ViewConnectWizard.String() != "ConnectWizard" {
		t.Errorf("ViewConnectWizard.String() = %q, want %q", ViewConnectWizard.String(), "ConnectWizard")
	}
}
