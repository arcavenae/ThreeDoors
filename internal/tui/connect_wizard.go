package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ProviderAuthType describes what authentication a provider needs.
type ProviderAuthType int

const (
	AuthNone     ProviderAuthType = iota // local providers (textfile, reminders)
	AuthAPIToken                         // API token entry (todoist, linear)
	AuthOAuth                            // OAuth device code flow (github, jira cloud)
)

// ProviderFormSpec describes how a provider appears in the connect wizard.
type ProviderFormSpec struct {
	Name        string           // registry name: "todoist", "github", etc.
	DisplayName string           // human-friendly: "Todoist", "GitHub Issues"
	Description string           // one-line description for provider select
	AuthType    ProviderAuthType // what auth flow to show
	TokenHelp   string           // where to find the token (for AuthAPIToken)
	NeedsPath   bool             // whether the provider needs a file/directory path
	PathHelp    string           // help text for the path field
}

// DefaultProviderSpecs returns the form specs for all built-in providers.
func DefaultProviderSpecs() []ProviderFormSpec {
	return []ProviderFormSpec{
		{
			Name:        "todoist",
			DisplayName: "Todoist",
			Description: "Todoist task manager",
			AuthType:    AuthAPIToken,
			TokenHelp:   "Settings → Integrations → Developer → API token",
		},
		{
			Name:        "github",
			DisplayName: "GitHub Issues",
			Description: "GitHub repository issues",
			AuthType:    AuthOAuth,
		},
		{
			Name:        "jira",
			DisplayName: "Jira",
			Description: "Jira Cloud or Server issues",
			AuthType:    AuthAPIToken,
			TokenHelp:   "Profile → Personal Access Tokens → Create token",
		},
		{
			Name:        "linear",
			DisplayName: "Linear",
			Description: "Linear project issues",
			AuthType:    AuthAPIToken,
			TokenHelp:   "Settings → API → Personal API keys",
		},
		{
			Name:        "obsidian",
			DisplayName: "Obsidian",
			Description: "Obsidian vault tasks",
			AuthType:    AuthNone,
			NeedsPath:   true,
			PathHelp:    "Path to your Obsidian vault directory",
		},
		{
			Name:        "textfile",
			DisplayName: "Plain text files",
			Description: "Local YAML/text task files",
			AuthType:    AuthNone,
			NeedsPath:   true,
			PathHelp:    "Path to your task file (e.g., ~/tasks.yaml)",
		},
		{
			Name:        "reminders",
			DisplayName: "Apple Reminders",
			Description: "macOS Reminders app",
			AuthType:    AuthNone,
		},
		{
			Name:        "applenotes",
			DisplayName: "Apple Notes",
			Description: "macOS Notes app (read-only)",
			AuthType:    AuthNone,
		},
	}
}

// WizardStep tracks which step the wizard is on.
type WizardStep int

const (
	StepProviderSelect WizardStep = iota
	StepProviderConfig
	StepSyncConfig
	StepTestConfirm
)

// ConnectWizardCompleteMsg is sent when the wizard finishes successfully.
type ConnectWizardCompleteMsg struct {
	ProviderName string
	Label        string
	Settings     map[string]string
	SyncMode     string
	PollInterval time.Duration
}

// ConnectWizardCancelMsg is sent when the wizard is cancelled.
type ConnectWizardCancelMsg struct{}

// ConnectWizard implements the 4-step connection setup wizard using huh forms.
type ConnectWizard struct {
	specs     []ProviderFormSpec
	connMgr   *connection.ConnectionManager
	width     int
	height    int
	step      WizardStep
	form      *huh.Form
	finished  bool
	cancelled bool

	// Collected values across steps
	selectedProvider string
	label            string
	apiToken         string
	filePath         string
	syncMode         string
	pollInterval     string
}

// NewConnectWizard creates a new wizard with the given provider specs.
func NewConnectWizard(specs []ProviderFormSpec, connMgr *connection.ConnectionManager) *ConnectWizard {
	w := &ConnectWizard{
		specs:   specs,
		connMgr: connMgr,
	}
	w.buildStep1Form()
	return w
}

// SetWidth sets the terminal width.
func (w *ConnectWizard) SetWidth(width int) {
	w.width = width
}

// SetHeight sets the terminal height.
func (w *ConnectWizard) SetHeight(height int) {
	w.height = height
}

// Step returns the current wizard step.
func (w *ConnectWizard) Step() WizardStep {
	return w.step
}

// SelectedProvider returns the selected provider name.
func (w *ConnectWizard) SelectedProvider() string {
	return w.selectedProvider
}

// buildStep1Form creates the provider selection form.
func (w *ConnectWizard) buildStep1Form() {
	options := make([]huh.Option[string], 0, len(w.specs))
	for _, spec := range w.specs {
		options = append(options, huh.NewOption(
			fmt.Sprintf("%s — %s", spec.DisplayName, spec.Description),
			spec.Name,
		))
	}

	w.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to connect?").
				Options(options...).
				Value(&w.selectedProvider),
		),
	).WithShowHelp(false).WithShowErrors(true)

	w.form.Init()
}

// buildStep2Form creates the provider-specific configuration form.
func (w *ConnectWizard) buildStep2Form() {
	spec := w.getSelectedSpec()
	if spec == nil {
		return
	}

	var fields []huh.Field

	// Label field (always present)
	fields = append(fields, huh.NewInput().
		Title("Give this connection a name").
		Placeholder(spec.DisplayName).
		Value(&w.label).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("name is required")
			}
			return nil
		}))

	switch spec.AuthType {
	case AuthAPIToken:
		fields = append(fields, huh.NewInput().
			Title("API Token").
			EchoMode(huh.EchoModePassword).
			Value(&w.apiToken).
			Description(spec.TokenHelp).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("API token is required")
				}
				return nil
			}))

	case AuthOAuth:
		// OAuth placeholder — device code flow is a future story (46.1)
		fields = append(fields, huh.NewNote().
			Title("OAuth Authentication").
			Description("OAuth device code flow will be available in a future update.\nFor now, please use an API token or personal access token."))

		fields = append(fields, huh.NewInput().
			Title("Personal Access Token").
			EchoMode(huh.EchoModePassword).
			Value(&w.apiToken).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("token is required")
				}
				return nil
			}))
	}

	if spec.NeedsPath {
		fields = append(fields, huh.NewInput().
			Title("File/Directory Path").
			Value(&w.filePath).
			Description(spec.PathHelp).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("path is required")
				}
				expanded := expandPath(s)
				if _, err := os.Stat(expanded); err != nil {
					return fmt.Errorf("path does not exist: %s", expanded)
				}
				return nil
			}))
	}

	w.form = huh.NewForm(
		huh.NewGroup(fields...),
	).WithShowHelp(false).WithShowErrors(true)

	w.form.Init()
}

// buildStep3Form creates the sync configuration form.
func (w *ConnectWizard) buildStep3Form() {
	w.syncMode = "readonly"
	w.pollInterval = "5m"

	w.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Sync mode").
				Options(
					huh.NewOption("Read & write (bidirectional)", "bidirectional"),
					huh.NewOption("Read only (import tasks, don't push)", "readonly"),
				).
				Value(&w.syncMode),

			huh.NewSelect[string]().
				Title("Poll interval").
				Options(
					huh.NewOption("Every 30 seconds", "30s"),
					huh.NewOption("Every 1 minute", "1m"),
					huh.NewOption("Every 5 minutes", "5m"),
					huh.NewOption("Every 15 minutes", "15m"),
					huh.NewOption("Every 30 minutes", "30m"),
				).
				Value(&w.pollInterval),
		),
	).WithShowHelp(false).WithShowErrors(true)

	w.form.Init()
}

// buildStep4Form creates the test & confirm form.
func (w *ConnectWizard) buildStep4Form() {
	spec := w.getSelectedSpec()
	displayName := w.selectedProvider
	if spec != nil {
		displayName = spec.DisplayName
	}

	label := w.label
	if label == "" && spec != nil {
		label = spec.DisplayName
	}

	var confirm bool
	w.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Connect %q (%s)", label, displayName)).
				Description(w.buildSummary()),
			huh.NewConfirm().
				Title("Proceed with connection?").
				Affirmative("Connect").
				Negative("Cancel").
				Value(&confirm),
		),
	).WithShowHelp(false).WithShowErrors(true)

	w.form.Init()
}

// buildSummary returns a text summary of the wizard configuration.
func (w *ConnectWizard) buildSummary() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Provider: %s", w.selectedProvider))
	if w.label != "" {
		parts = append(parts, fmt.Sprintf("Label: %s", w.label))
	}
	if w.apiToken != "" {
		parts = append(parts, "Token: ••••••••")
	}
	if w.filePath != "" {
		parts = append(parts, fmt.Sprintf("Path: %s", w.filePath))
	}
	parts = append(parts, fmt.Sprintf("Sync: %s", w.syncMode))
	parts = append(parts, fmt.Sprintf("Poll: %s", w.pollInterval))
	return strings.Join(parts, "\n")
}

// getSelectedSpec returns the ProviderFormSpec for the currently selected provider.
func (w *ConnectWizard) getSelectedSpec() *ProviderFormSpec {
	for i := range w.specs {
		if w.specs[i].Name == w.selectedProvider {
			return &w.specs[i]
		}
	}
	return nil
}

// Init satisfies the tea.Model interface.
func (w *ConnectWizard) Init() tea.Cmd {
	return w.form.Init()
}

// Update handles messages for the connect wizard.
func (w *ConnectWizard) Update(msg tea.Msg) tea.Cmd {
	// Check for Esc to cancel at any step
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			w.cancelled = true
			return func() tea.Msg { return ConnectWizardCancelMsg{} }
		}
	}

	// Pass to the huh form
	model, cmd := w.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		w.form = f
	}

	// Check if the current form completed
	if w.form.State == huh.StateCompleted {
		return w.advanceStep()
	}

	// Check if the form was aborted (huh uses StateAborted for Esc)
	if w.form.State == huh.StateAborted {
		w.cancelled = true
		return func() tea.Msg { return ConnectWizardCancelMsg{} }
	}

	return cmd
}

// advanceStep moves to the next wizard step when the current form completes.
func (w *ConnectWizard) advanceStep() tea.Cmd {
	switch w.step {
	case StepProviderSelect:
		w.step = StepProviderConfig
		w.buildStep2Form()
		return w.form.Init()

	case StepProviderConfig:
		// Default label if empty
		if w.label == "" {
			spec := w.getSelectedSpec()
			if spec != nil {
				w.label = spec.DisplayName
			}
		}
		w.step = StepSyncConfig
		w.buildStep3Form()
		return w.form.Init()

	case StepSyncConfig:
		w.step = StepTestConfirm
		w.buildStep4Form()
		return w.form.Init()

	case StepTestConfirm:
		w.finished = true
		return w.completeWizard()
	}

	return nil
}

// completeWizard sends the completion message with all collected data.
func (w *ConnectWizard) completeWizard() tea.Cmd {
	settings := make(map[string]string)
	if w.filePath != "" {
		settings["path"] = expandPath(w.filePath)
	}

	pollDuration, err := time.ParseDuration(w.pollInterval)
	if err != nil {
		pollDuration = 5 * time.Minute
	}

	return func() tea.Msg {
		return ConnectWizardCompleteMsg{
			ProviderName: w.selectedProvider,
			Label:        w.label,
			Settings:     settings,
			SyncMode:     w.syncMode,
			PollInterval: pollDuration,
		}
	}
}

// View renders the current wizard step.
func (w *ConnectWizard) View() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	// Step indicator
	stepNames := []string{"Select Provider", "Configure", "Sync Settings", "Confirm"}
	stepNum := int(w.step) + 1
	stepTotal := len(stepNames)

	fmt.Fprintf(&s, "%s\n", headerStyle.Render(
		fmt.Sprintf("Connect a Data Source (%d/%d: %s)", stepNum, stepTotal, stepNames[w.step]),
	))
	fmt.Fprintf(&s, "\n")

	// Render the huh form
	fmt.Fprintf(&s, "%s", w.form.View())

	// Footer hint
	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" esc:cancel"))

	return s.String()
}

// expandPath expands ~ to the home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
