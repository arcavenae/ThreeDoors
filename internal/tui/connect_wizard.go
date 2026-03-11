package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// wizardPhase tracks which phase the connect wizard is in.
type wizardPhase int

const (
	wizardPhaseForm    wizardPhase = iota // huh form collecting Steps 1-3
	wizardPhaseTesting                    // Step 4: running connection test
	wizardPhaseResult                     // Step 4: showing test results, confirm/cancel
)

// ShowConnectWizardMsg opens the connect wizard.
type ShowConnectWizardMsg struct{}

// ConnectWizardCompleteMsg is sent when the wizard finishes successfully.
type ConnectWizardCompleteMsg struct {
	Connection *connection.Connection
}

// ConnectWizardCancelMsg is sent when the wizard is cancelled.
type ConnectWizardCancelMsg struct{}

// connectTestResultMsg carries the async health check result.
type connectTestResultMsg struct {
	result connection.HealthCheckResult
	err    error
}

// providerChoice pairs a provider name with its description for the select list.
type providerChoice struct {
	Name        string
	Description string
	AuthType    connection.AuthType
}

// ConnectWizard is a Bubbletea model for the interactive connection setup wizard.
type ConnectWizard struct {
	phase     wizardPhase
	form      *huh.Form
	providers []providerChoice
	formSpecs map[string]connection.FormSpec
	registry  *core.Registry

	// Form field bindings (set by huh form).
	selectedProvider string
	label            string
	apiToken         string
	localPath        string
	syncMode         string
	pollInterval     string
	filterValue      string

	// Step 4 state.
	testResult *connection.HealthCheckResult
	testErr    error

	// Service for creating connections.
	connService ConnectServicer

	width  int
	height int
}

// ConnectServicer abstracts the connection creation for testability.
type ConnectServicer interface {
	Add(providerName, label string, settings map[string]string, credential string) (*connection.Connection, error)
	TestConnection(id string) (connection.HealthCheckResult, error)
}

// NewConnectWizard creates a ConnectWizard with providers from the registry.
func NewConnectWizard(reg *core.Registry, svc ConnectServicer) *ConnectWizard {
	w := &ConnectWizard{
		phase:        wizardPhaseForm,
		registry:     reg,
		formSpecs:    make(map[string]connection.FormSpec),
		connService:  svc,
		syncMode:     "readonly",
		pollInterval: "5m",
	}

	w.loadProviders()
	w.buildForm()

	return w
}

// loadProviders discovers registered providers and their FormSpecs.
func (w *ConnectWizard) loadProviders() {
	names := w.registry.ListProviders()
	for _, name := range names {
		choice := providerChoice{
			Name:        name,
			Description: name + " provider",
			AuthType:    connection.AuthNone,
		}

		// Check if the provider factory produces a FormSpecProvider.
		// We can't instantiate just to check, so use a default description.
		// Providers register their FormSpec via the formSpecs map populated externally.
		w.providers = append(w.providers, choice)
	}
}

// SetFormSpec registers a FormSpec for a provider name.
// Call this before the wizard is shown to populate provider-specific fields.
func (w *ConnectWizard) SetFormSpec(providerName string, spec connection.FormSpec) {
	w.formSpecs[providerName] = spec

	// Update the provider choice with the spec's description.
	for i := range w.providers {
		if w.providers[i].Name == providerName {
			if spec.Description != "" {
				w.providers[i].Description = spec.Description
			}
			w.providers[i].AuthType = spec.AuthType
			break
		}
	}

	// Rebuild the form with updated provider info.
	w.buildForm()
}

// buildForm constructs the huh form with groups for each wizard step.
// Step 2 uses separate groups per auth type with WithHideFunc for dynamic switching.
func (w *ConnectWizard) buildForm() {
	// Step 1: Provider selection.
	options := make([]huh.Option[string], 0, len(w.providers))
	for _, p := range w.providers {
		label := fmt.Sprintf("%s — %s", p.Name, p.Description)
		options = append(options, huh.NewOption(label, p.Name))
	}

	providerSelect := huh.NewSelect[string]().
		Title("Select a data source").
		Description("Choose which service to connect").
		Options(options...).
		Value(&w.selectedProvider)

	// Step 2 common: label input (shown in all auth groups).
	makeLabelInput := func() *huh.Input {
		return huh.NewInput().
			Title("Connection label").
			Description("A friendly name for this connection (e.g., 'Work Jira')").
			Value(&w.label).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("label is required")
				}
				return nil
			})
	}

	// Step 2a: API token auth group.
	tokenHelp := "Enter your API token"
	apiTokenInput := huh.NewInput().
		Title("API Token").
		Description(tokenHelp).
		Value(&w.apiToken).
		EchoMode(huh.EchoModePassword)

	apiTokenGroup := huh.NewGroup(makeLabelInput(), apiTokenInput).
		Title("Step 2: Configure Connection").
		WithHideFunc(func() bool {
			return w.authTypeForSelected() != connection.AuthAPIToken
		})

	// Step 2b: Local path auth group.
	localPathInput := huh.NewInput().
		Title("File path").
		Description("Path to the task file or directory").
		Value(&w.localPath)

	localPathGroup := huh.NewGroup(makeLabelInput(), localPathInput).
		Title("Step 2: Configure Connection").
		WithHideFunc(func() bool {
			return w.authTypeForSelected() != connection.AuthLocalPath
		})

	// Step 2c: OAuth auth group.
	oauthNote := huh.NewNote().
		Title("OAuth Authentication").
		Description("OAuth device code flow will be initiated after setup.\nYou'll be prompted to authorize in your browser.")

	oauthGroup := huh.NewGroup(makeLabelInput(), oauthNote).
		Title("Step 2: Configure Connection").
		WithHideFunc(func() bool {
			return w.authTypeForSelected() != connection.AuthOAuth
		})

	// Step 2d: No-auth group (just label).
	noAuthGroup := huh.NewGroup(makeLabelInput()).
		Title("Step 2: Configure Connection").
		WithHideFunc(func() bool {
			return w.authTypeForSelected() != connection.AuthNone
		})

	// Step 3: Sync configuration.
	syncModeSelect := huh.NewSelect[string]().
		Title("Sync mode").
		Description("How should tasks be synchronized?").
		Options(
			huh.NewOption("Read-only (import tasks, no write-back)", "readonly"),
			huh.NewOption("Bidirectional (sync completions back)", "bidirectional"),
		).
		Value(&w.syncMode)

	pollIntervalSelect := huh.NewSelect[string]().
		Title("Poll interval").
		Description("How often to check for new tasks").
		Options(
			huh.NewOption("1 minute", "1m"),
			huh.NewOption("5 minutes", "5m"),
			huh.NewOption("15 minutes", "15m"),
			huh.NewOption("30 minutes", "30m"),
			huh.NewOption("1 hour", "1h"),
		).
		Value(&w.pollInterval)

	filterInput := huh.NewInput().
		Title("Filter").
		Description("Optional: filter expression (projects, repos, JQL, etc.)").
		Value(&w.filterValue)

	w.form = huh.NewForm(
		huh.NewGroup(providerSelect).Title("Step 1: Choose Provider"),
		apiTokenGroup,
		localPathGroup,
		oauthGroup,
		noAuthGroup,
		huh.NewGroup(syncModeSelect, pollIntervalSelect, filterInput).Title("Step 3: Sync Settings"),
	).WithShowHelp(true).WithShowErrors(true)
}

// authTypeForSelected returns the AuthType of the currently selected provider.
func (w *ConnectWizard) authTypeForSelected() connection.AuthType {
	if spec, ok := w.formSpecs[w.selectedProvider]; ok {
		return spec.AuthType
	}
	// Default: check known provider patterns.
	switch w.selectedProvider {
	case "todoist", "jira":
		return connection.AuthAPIToken
	case "github":
		return connection.AuthOAuth
	case "textfile", "obsidian":
		return connection.AuthLocalPath
	default:
		return connection.AuthNone
	}
}

// SetWidth sets the available width.
func (w *ConnectWizard) SetWidth(width int) {
	w.width = width
}

// SetHeight sets the available height.
func (w *ConnectWizard) SetHeight(height int) {
	w.height = height
}

// Init initializes the wizard.
func (w *ConnectWizard) Init() tea.Cmd {
	return w.form.Init()
}

// Update handles messages for the connect wizard.
func (w *ConnectWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch w.phase {
		case wizardPhaseForm:
			if msg.String() == "esc" {
				return w, func() tea.Msg { return ConnectWizardCancelMsg{} }
			}
		case wizardPhaseTesting:
			if msg.String() == "esc" {
				return w, func() tea.Msg { return ConnectWizardCancelMsg{} }
			}
		case wizardPhaseResult:
			switch msg.String() {
			case "esc", "n":
				return w, func() tea.Msg { return ConnectWizardCancelMsg{} }
			case "enter", "y":
				return w, w.finalizeConnection()
			}
		}

	case connectTestResultMsg:
		w.phase = wizardPhaseResult
		if msg.err != nil {
			w.testErr = msg.err
		} else {
			w.testResult = &msg.result
		}
		return w, nil
	}

	// Delegate to huh form during form phase.
	if w.phase == wizardPhaseForm {
		form, cmd := w.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			w.form = f
		}

		// Check if form completed.
		if w.form.State == huh.StateCompleted {
			w.phase = wizardPhaseTesting
			return w, w.runConnectionTest()
		}

		return w, cmd
	}

	return w, nil
}

// runConnectionTest creates a temporary connection and tests it.
func (w *ConnectWizard) runConnectionTest() tea.Cmd {
	return func() tea.Msg {
		if w.connService == nil {
			return connectTestResultMsg{
				err: fmt.Errorf("connection service not available"),
			}
		}

		settings := w.buildSettings()
		credential := w.apiToken

		conn, err := w.connService.Add(w.selectedProvider, w.label, settings, credential)
		if err != nil {
			return connectTestResultMsg{err: fmt.Errorf("create connection: %w", err)}
		}

		result, err := w.connService.TestConnection(conn.ID)
		if err != nil {
			return connectTestResultMsg{
				result: connection.HealthCheckResult{},
				err:    fmt.Errorf("test connection: %w", err),
			}
		}

		return connectTestResultMsg{result: result}
	}
}

// buildSettings assembles the settings map from form values.
func (w *ConnectWizard) buildSettings() map[string]string {
	settings := make(map[string]string)

	settings["sync_mode"] = w.syncMode
	settings["poll_interval"] = w.pollInterval

	if w.filterValue != "" {
		settings["filter"] = w.filterValue
	}

	if w.localPath != "" {
		settings["path"] = w.localPath
	}

	return settings
}

// finalizeConnection completes the wizard by signaling success.
func (w *ConnectWizard) finalizeConnection() tea.Cmd {
	return func() tea.Msg {
		// The connection was already created during the test phase.
		// Signal completion.
		return ConnectWizardCompleteMsg{}
	}
}

// View renders the wizard based on the current phase.
func (w *ConnectWizard) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	fmt.Fprintf(&s, "%s\n", titleStyle.Render("Connect Data Source"))

	switch w.phase {
	case wizardPhaseForm:
		fmt.Fprintf(&s, "%s", w.form.View())

	case wizardPhaseTesting:
		fmt.Fprintf(&s, "\nTesting connection to %s...\n", w.selectedProvider)
		fmt.Fprintf(&s, "Checking API reachability, token validity, and task count.\n")

	case wizardPhaseResult:
		fmt.Fprintf(&s, "\n")
		w.renderTestResults(&s)
	}

	return s.String()
}

// renderTestResults shows the health check outcome and confirm/cancel prompt.
func (w *ConnectWizard) renderTestResults(s *strings.Builder) {
	if w.testErr != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		fmt.Fprintf(s, "%s\n\n", errStyle.Render("Connection test failed:"))
		fmt.Fprintf(s, "  %s\n\n", w.testErr.Error())
		fmt.Fprintf(s, "Press [enter] to save anyway, [esc] to cancel.\n")
		return
	}

	if w.testResult != nil {
		okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

		statusIcon := func(ok bool) string {
			if ok {
				return okStyle.Render("✓")
			}
			return failStyle.Render("✗")
		}

		fmt.Fprintf(s, "Connection Test Results:\n\n")
		fmt.Fprintf(s, "  %s API Reachable\n", statusIcon(w.testResult.APIReachable))
		fmt.Fprintf(s, "  %s Token Valid\n", statusIcon(w.testResult.TokenValid))
		fmt.Fprintf(s, "  %s Rate Limit OK\n", statusIcon(w.testResult.RateLimitOK))
		fmt.Fprintf(s, "  Tasks found: %d\n\n", w.testResult.TaskCount)

		if w.testResult.Healthy() {
			fmt.Fprintf(s, "%s\n\n", okStyle.Render("All checks passed!"))
		} else {
			fmt.Fprintf(s, "%s\n\n", failStyle.Render("Some checks failed."))
		}
	}

	fmt.Fprintf(s, "Provider: %s\n", w.selectedProvider)
	fmt.Fprintf(s, "Label:    %s\n", w.label)
	fmt.Fprintf(s, "Sync:     %s every %s\n\n", w.syncMode, w.pollInterval)
	fmt.Fprintf(s, "Press [enter/y] to connect, [esc/n] to cancel.\n")
}

// parsePollInterval converts a poll interval string to time.Duration.
func parsePollInterval(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}
