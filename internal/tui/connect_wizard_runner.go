package tui

import (
	"fmt"
	"io"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

// ConnectWizardRunner wraps ConnectWizard as a standalone tea.Model for CLI use.
// It handles wizard completion/cancellation messages and bridges to ConnectionService.
type ConnectWizardRunner struct {
	wizard *ConnectWizard
	svc    *connection.ConnectionService
	done   bool
	err    error
	result *connectWizardResult
}

// connectWizardResult holds the outcome of a successful wizard run.
type connectWizardResult struct {
	conn       *connection.Connection
	testResult *connectTestResult
	testErr    error
}

// connectTestResult mirrors the health check output for display.
type connectTestResult struct {
	Healthy bool
	Checks  []connectTestCheck
}

// connectTestCheck is a single health check item.
type connectTestCheck struct {
	Name   string
	Passed bool
}

// NewConnectWizardRunner creates a runner that wraps the given wizard.
func NewConnectWizardRunner(
	provider string,
	svc *connection.ConnectionService,
	manager *connection.ConnectionManager,
) *ConnectWizardRunner {
	specs := DefaultProviderSpecs()
	wizard := NewConnectWizard(specs, manager)
	if provider != "" {
		wizard.SetProvider(provider)
	}

	return &ConnectWizardRunner{
		wizard: wizard,
		svc:    svc,
	}
}

// Init satisfies tea.Model.
func (r *ConnectWizardRunner) Init() tea.Cmd {
	return r.wizard.Init()
}

// Update satisfies tea.Model.
func (r *ConnectWizardRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ConnectWizardCompleteMsg:
		conn, err := r.svc.Add(msg.ProviderName, msg.Label, msg.Settings, msg.Token)
		if err != nil {
			r.done = true
			r.err = fmt.Errorf("create connection: %w", err)
			return r, tea.Quit
		}

		conn.SyncMode = msg.SyncMode
		conn.PollInterval = msg.PollInterval

		// Run health test (best-effort).
		var testResult *connectTestResult
		var testErr error
		healthResult, err := r.svc.TestConnection(conn.ID)
		if err != nil {
			testErr = err
		} else {
			testResult = formatHealthResult(healthResult)
		}

		r.done = true
		r.result = &connectWizardResult{
			conn:       conn,
			testResult: testResult,
			testErr:    testErr,
		}
		return r, tea.Quit

	case ConnectWizardCancelMsg:
		r.done = true
		return r, tea.Quit
	}

	// Delegate to wizard.
	cmds := r.wizard.Update(msg)
	return r, cmds
}

// View satisfies tea.Model.
func (r *ConnectWizardRunner) View() string {
	return r.wizard.View()
}

// Err returns any error from the wizard run.
func (r *ConnectWizardRunner) Err() error {
	return r.err
}

// PrintResult writes the connection result to the writer in human-readable format.
func (r *ConnectWizardRunner) PrintResult(w io.Writer) {
	if r.result == nil {
		return
	}

	conn := r.result.conn
	_, _ = fmt.Fprintf(w, "\nConnection created:\n")
	_, _ = fmt.Fprintf(w, "  Name:     %s\n", conn.Label)
	_, _ = fmt.Fprintf(w, "  Provider: %s\n", conn.ProviderName)
	_, _ = fmt.Fprintf(w, "  ID:       %s\n", conn.ID)
	_, _ = fmt.Fprintf(w, "  Sync:     %s\n", conn.SyncMode)
	_, _ = fmt.Fprintf(w, "  Poll:     %s\n", formatPollDuration(conn.PollInterval))

	if r.result.testResult != nil {
		_, _ = fmt.Fprintf(w, "\nConnection test:\n")
		for _, c := range r.result.testResult.Checks {
			icon := "✓"
			if !c.Passed {
				icon = "✗"
			}
			_, _ = fmt.Fprintf(w, "  %s %s\n", icon, c.Name)
		}
	} else if r.result.testErr != nil {
		_, _ = fmt.Fprintf(w, "\nConnection test: skipped (%v)\n", r.result.testErr)
	}
}

// formatHealthResult converts a HealthCheckResult into a connectTestResult.
func formatHealthResult(result connection.HealthCheckResult) *connectTestResult {
	return &connectTestResult{
		Healthy: result.Healthy(),
		Checks: []connectTestCheck{
			{Name: "DNS resolution", Passed: result.APIReachable},
			{Name: "TLS", Passed: result.APIReachable},
			{Name: "Authentication", Passed: result.TokenValid},
			{Name: "Authorization", Passed: result.TokenValid},
			{Name: "Rate limit", Passed: result.RateLimitOK},
		},
	}
}

// formatPollDuration formats a duration for human display.
func formatPollDuration(d time.Duration) string {
	if d == 0 {
		return "5m"
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
