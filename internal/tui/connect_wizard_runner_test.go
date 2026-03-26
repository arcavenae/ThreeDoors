package tui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

// stubHealthChecker implements connection.HealthChecker for runner tests.
type stubRunnerHealthChecker struct {
	result connection.HealthCheckResult
	err    error
}

func (s *stubRunnerHealthChecker) CheckHealth(_ *connection.Connection, _ string) (connection.HealthCheckResult, error) {
	return s.result, s.err
}

// stubRunnerCredentialStore implements connection.CredentialStore for runner tests.
type stubRunnerCredentialStore struct{}

func (s *stubRunnerCredentialStore) Get(_ string) (string, error) { return "", nil }
func (s *stubRunnerCredentialStore) Set(_, _ string) error        { return nil }
func (s *stubRunnerCredentialStore) Delete(_ string) error        { return nil }

func newRunnerTestService(t *testing.T, manager *connection.ConnectionManager, checker connection.HealthChecker) *connection.ConnectionService {
	t.Helper()
	dir := t.TempDir()
	svc, err := connection.NewConnectionService(connection.ServiceConfig{
		Manager:    manager,
		Creds:      &stubRunnerCredentialStore{},
		ConfigPath: dir + "/config.yaml",
		Checker:    checker,
	})
	if err != nil {
		t.Fatalf("NewConnectionService: %v", err)
	}
	return svc
}

func TestConnectWizardRunner_CompleteMsg(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	checker := &stubRunnerHealthChecker{result: connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
		TaskCount:    5,
	}}
	svc := newRunnerTestService(t, manager, checker)

	runner := NewConnectWizardRunner("todoist", svc, manager)

	msg := ConnectWizardCompleteMsg{
		ProviderName: "todoist",
		Label:        "My Todoist",
		Settings:     map[string]string{},
		SyncMode:     "readonly",
		PollInterval: 5 * time.Minute,
		Token:        "tok_abc",
	}

	model, cmd := runner.Update(msg)
	r := model.(*ConnectWizardRunner)

	if !r.done {
		t.Fatal("runner should be done after complete msg")
	}
	if r.Err() != nil {
		t.Fatalf("unexpected error: %v", r.Err())
	}
	if r.result == nil {
		t.Fatal("result should be non-nil")
	}
	if r.result.conn.Label != "My Todoist" {
		t.Errorf("label = %q, want %q", r.result.conn.Label, "My Todoist")
	}
	if r.result.conn.ProviderName != "todoist" {
		t.Errorf("provider = %q, want %q", r.result.conn.ProviderName, "todoist")
	}
	if r.result.testResult == nil {
		t.Fatal("test result should be non-nil")
	}
	if !r.result.testResult.Healthy {
		t.Error("expected healthy connection")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd")
	}

	// Verify PrintResult output.
	var buf bytes.Buffer
	r.PrintResult(&buf)
	out := buf.String()
	for _, want := range []string{
		"Connection created:",
		"Name:     My Todoist",
		"Provider: todoist",
		"Connection test:",
		"✓ DNS resolution",
		"✓ Authentication",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestConnectWizardRunner_CancelMsg(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	svc := newRunnerTestService(t, manager, nil)

	runner := NewConnectWizardRunner("", svc, manager)

	model, cmd := runner.Update(ConnectWizardCancelMsg{})
	r := model.(*ConnectWizardRunner)

	if !r.done {
		t.Fatal("runner should be done after cancel msg")
	}
	if r.Err() != nil {
		t.Fatalf("unexpected error: %v", r.Err())
	}
	if r.result != nil {
		t.Error("result should be nil on cancel")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd")
	}
}

func TestConnectWizardRunner_DelegatesToWizard(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	svc := newRunnerTestService(t, manager, nil)

	runner := NewConnectWizardRunner("", svc, manager)

	model, _ := runner.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	r := model.(*ConnectWizardRunner)

	if r.done {
		t.Error("runner should not be done after regular key")
	}

	view := r.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestConnectWizardRunner_PrintResult_Nil(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	svc := newRunnerTestService(t, manager, nil)

	runner := NewConnectWizardRunner("", svc, manager)

	var buf bytes.Buffer
	runner.PrintResult(&buf)
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestConnectWizardRunner_ProviderPreSelection(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	svc := newRunnerTestService(t, manager, nil)

	runner := NewConnectWizardRunner("github", svc, manager)

	// The wizard should have pre-selected github and be at StepProviderConfig.
	if runner.wizard.Step() != StepProviderConfig {
		t.Errorf("step = %v, want StepProviderConfig after pre-selection", runner.wizard.Step())
	}
	if runner.wizard.SelectedProvider() != "github" {
		t.Errorf("selectedProvider = %q, want %q", runner.wizard.SelectedProvider(), "github")
	}
}

func TestConnectWizardRunner_InvalidProviderStaysAtStep1(t *testing.T) {
	t.Parallel()

	manager := connection.NewConnectionManager(nil)
	svc := newRunnerTestService(t, manager, nil)

	runner := NewConnectWizardRunner("nonexistent", svc, manager)

	if runner.wizard.Step() != StepProviderSelect {
		t.Errorf("step = %v, want StepProviderSelect for invalid provider", runner.wizard.Step())
	}
}

func TestFormatPollDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"zero", 0, "5m"},
		{"30 seconds", 30 * time.Second, "30s"},
		{"1 minute", time.Minute, "1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"1 hour", time.Hour, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatPollDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatPollDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatHealthResult(t *testing.T) {
	t.Parallel()

	result := connection.HealthCheckResult{
		APIReachable: true,
		TokenValid:   true,
		RateLimitOK:  true,
	}
	tr := formatHealthResult(result)

	if !tr.Healthy {
		t.Error("expected healthy=true")
	}
	if len(tr.Checks) != 5 {
		t.Errorf("len(checks) = %d, want 5", len(tr.Checks))
	}
	for _, c := range tr.Checks {
		if !c.Passed {
			t.Errorf("check %q should pass", c.Name)
		}
	}
}
