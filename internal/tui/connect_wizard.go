package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/arcaven/ThreeDoors/internal/core/connection/oauth"
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
	Name          string                  // registry name: "todoist", "github", etc.
	DisplayName   string                  // human-friendly: "Todoist", "GitHub Issues"
	Description   string                  // one-line description for provider select
	AuthType      ProviderAuthType        // what auth flow to show
	TokenHelp     string                  // where to find the token (for AuthAPIToken)
	NeedsPath     bool                    // whether the provider needs a file/directory path
	PathHelp      string                  // help text for the path field
	OAuthClientID string                  // OAuth App client ID (for AuthOAuth providers)
	EnvTokenFunc  func() (string, string) // returns (token, envVarName) if env token exists
	Detected      bool                    // true if auto-detected on the system
	DetectInfo    string                  // detection detail: "gh CLI authenticated", "vault found at ~/docs"
	PreFill       map[string]string       // settings pre-filled from detection (e.g., path, token_env)
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
			Name:          "github",
			DisplayName:   "GitHub Issues",
			Description:   "GitHub repository issues",
			AuthType:      AuthOAuth,
			OAuthClientID: os.Getenv("THREEDOORS_GITHUB_CLIENT_ID"),
			EnvTokenFunc:  defaultGitHubEnvTokenFunc,
		},
		{
			Name:        "jira",
			DisplayName: "Jira",
			Description: "Jira Cloud or Server issues",
			AuthType:    AuthAPIToken,
			TokenHelp:   "Profile → Personal Access Tokens → Create token",
		},
		{
			Name:         "linear",
			DisplayName:  "Linear",
			Description:  "Linear project issues",
			AuthType:     AuthOAuth,
			EnvTokenFunc: defaultLinearEnvTokenFunc,
		},
		{
			Name:        "clickup",
			DisplayName: "ClickUp",
			Description: "ClickUp workspace tasks",
			AuthType:    AuthAPIToken,
			TokenHelp:   "Settings → Apps → API Token (or set CLICKUP_API_TOKEN env var)",
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

// ApplyDetection takes a list of provider specs and detection results, and returns
// a new list with detected providers annotated and moved to the top.
func ApplyDetection(specs []ProviderFormSpec, results []connection.DetectionResult) []ProviderFormSpec {
	if len(results) == 0 {
		return specs
	}

	// Build a map from provider name to detection result
	detections := make(map[string]connection.DetectionResult, len(results))
	for _, r := range results {
		detections[r.ProviderName] = r
	}

	// Split into detected and undetected, annotating detected ones
	var detected, undetected []ProviderFormSpec
	for _, spec := range specs {
		if r, ok := detections[spec.Name]; ok {
			spec.Detected = true
			spec.DetectInfo = r.Reason
			spec.PreFill = r.PreFill
			detected = append(detected, spec)
		} else {
			undetected = append(undetected, spec)
		}
	}

	// Detected providers appear first, then undetected in original order
	result := make([]ProviderFormSpec, 0, len(specs))
	result = append(result, detected...)
	result = append(result, undetected...)
	return result
}

// WizardStep tracks which step the wizard is on.
type WizardStep int

const (
	StepProviderSelect WizardStep = iota
	StepProviderConfig
	StepOAuthFlow // device code flow: show user code, poll for token
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
	Token        string // OAuth/PAT token for credential store (not persisted to config)
}

// ConnectWizardCancelMsg is sent when the wizard is cancelled.
type ConnectWizardCancelMsg struct{}

// oauthFlowState tracks the state of an in-progress OAuth device code flow.
type oauthFlowState struct {
	userCode        string
	verificationURI string
	status          string // "waiting", "success", "error"
	errorMsg        string
	browserOpened   bool
}

// oauthTokenResultMsg is sent when the OAuth polling completes.
type oauthTokenResultMsg struct {
	token string
	err   error
}

// ConnectWizard implements the connection setup wizard using huh forms.
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

	// OAuth flow state
	authMethod  string             // "oauth", "pat", "env" — chosen auth method for OAuth providers
	oauthState  *oauthFlowState    // non-nil during StepOAuthFlow
	oauthCancel context.CancelFunc // cancel polling
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

// SetProvider pre-selects a provider and skips Step 1 (provider selection),
// jumping directly to Step 2 (provider config). If the provider name is not
// found in the spec list, it is ignored and the wizard starts at Step 1.
func (w *ConnectWizard) SetProvider(name string) {
	for _, spec := range w.specs {
		if spec.Name == name {
			w.selectedProvider = name
			w.step = StepProviderConfig
			w.buildStep2Form()
			return
		}
	}
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
		label := fmt.Sprintf("%s — %s", spec.DisplayName, spec.Description)
		if spec.Detected {
			label = fmt.Sprintf("%s (detected — %s)", spec.DisplayName, spec.DetectInfo)
		}
		options = append(options, huh.NewOption(label, spec.Name))
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
		w.authMethod = "pat" // default

		// Check for existing env token.
		var envToken, envVarName string
		if spec.EnvTokenFunc != nil {
			envToken, envVarName = spec.EnvTokenFunc()
		}

		// Build auth method options.
		var authOptions []huh.Option[string]
		if envToken != "" {
			authOptions = append(authOptions, huh.NewOption(
				fmt.Sprintf("Use existing token from %s", envVarName),
				"env",
			))
			w.authMethod = "env" // default to env if available
		}
		if spec.OAuthClientID != "" {
			authOptions = append(authOptions, huh.NewOption(
				"Authenticate with browser (OAuth)",
				"oauth",
			))
			if envToken == "" {
				w.authMethod = "oauth" // default to OAuth if no env token
			}
		}
		authOptions = append(authOptions, huh.NewOption(
			"Enter a Personal Access Token",
			"pat",
		))

		fields = append(fields, huh.NewSelect[string]().
			Title("Authentication method").
			Options(authOptions...).
			Value(&w.authMethod))

		// PAT input field — shown for all methods but only validated when authMethod is "pat".
		fields = append(fields, huh.NewInput().
			Title("Personal Access Token").
			EchoMode(huh.EchoModePassword).
			Value(&w.apiToken).
			Description("Required if using PAT method").
			Validate(func(s string) error {
				if w.authMethod == "pat" && strings.TrimSpace(s) == "" {
					return fmt.Errorf("token is required for PAT method")
				}
				return nil
			}))
	}

	if spec.NeedsPath {
		// Pre-fill path from detection if available
		if spec.Detected && spec.PreFill != nil {
			if path, ok := spec.PreFill["path"]; ok && path != "" {
				w.filePath = path
			}
		}
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
	// Check for Esc to cancel at any step.
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.Type == tea.KeyEsc {
			w.cancelOAuthFlow()
			w.cancelled = true
			return func() tea.Msg { return ConnectWizardCancelMsg{} }
		}
	}

	// Handle OAuth-specific messages during StepOAuthFlow.
	if w.step == StepOAuthFlow {
		return w.updateOAuthFlow(msg)
	}

	// Pass to the huh form.
	model, cmd := w.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		w.form = f
	}

	// Check if the current form completed.
	if w.form.State == huh.StateCompleted {
		return w.advanceStep()
	}

	// Check if the form was aborted (huh uses StateAborted for Esc).
	if w.form.State == huh.StateAborted {
		w.cancelled = true
		return func() tea.Msg { return ConnectWizardCancelMsg{} }
	}

	return cmd
}

// updateOAuthFlow handles messages during the device code flow step.
func (w *ConnectWizard) updateOAuthFlow(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case oauthDeviceCodeMsg:
		if msg.err != nil {
			w.oauthState = &oauthFlowState{
				status:   "error",
				errorMsg: msg.err.Error(),
			}
			return nil
		}

		// Display user code and verification URI.
		w.oauthState = &oauthFlowState{
			userCode:        msg.resp.UserCode,
			verificationURI: msg.resp.VerificationURI,
			status:          "waiting",
		}

		// Open browser.
		_ = oauth.OpenBrowser(context.Background(), msg.resp.VerificationURI)
		w.oauthState.browserOpened = true

		// Start polling for token.
		ctx, cancel := context.WithCancel(context.Background())
		w.oauthCancel = cancel

		config := msg.config
		dcResp := msg.resp
		return func() tea.Msg {
			client := oauth.NewClient(nil)
			tokenResp, err := client.PollForToken(ctx, config, dcResp)
			if err != nil {
				return oauthTokenResultMsg{err: err}
			}
			return oauthTokenResultMsg{token: tokenResp.AccessToken}
		}

	case oauthTokenResultMsg:
		if msg.err != nil {
			w.oauthState = &oauthFlowState{
				status:   "error",
				errorMsg: msg.err.Error(),
			}
			return nil
		}

		// Success — store token and advance.
		w.apiToken = msg.token
		w.oauthState.status = "success"
		return w.advanceStep()
	}

	return nil
}

// advanceStep moves to the next wizard step when the current form completes.
func (w *ConnectWizard) advanceStep() tea.Cmd {
	switch w.step {
	case StepProviderSelect:
		w.step = StepProviderConfig
		w.buildStep2Form()
		return w.form.Init()

	case StepProviderConfig:
		// Default label if empty.
		if w.label == "" {
			spec := w.getSelectedSpec()
			if spec != nil {
				w.label = spec.DisplayName
			}
		}

		spec := w.getSelectedSpec()

		// Handle env token: use it directly, skip OAuth flow.
		if spec != nil && spec.AuthType == AuthOAuth && w.authMethod == "env" {
			if spec.EnvTokenFunc != nil {
				envToken, _ := spec.EnvTokenFunc()
				w.apiToken = envToken
			}
			w.step = StepSyncConfig
			w.buildStep3Form()
			return w.form.Init()
		}

		// Handle OAuth flow: enter the device code step.
		if spec != nil && spec.AuthType == AuthOAuth && w.authMethod == "oauth" {
			w.step = StepOAuthFlow
			return w.startOAuthFlow()
		}

		// PAT or non-OAuth: proceed to sync config.
		w.step = StepSyncConfig
		w.buildStep3Form()
		return w.form.Init()

	case StepOAuthFlow:
		// OAuth flow completed — token is in w.apiToken.
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
			Token:        w.apiToken,
		}
	}
}

// startOAuthFlow initiates the device code flow for the selected OAuth provider.
func (w *ConnectWizard) startOAuthFlow() tea.Cmd {
	spec := w.getSelectedSpec()
	if spec == nil || spec.OAuthClientID == "" {
		w.oauthState = &oauthFlowState{
			status:   "error",
			errorMsg: "OAuth is not configured for this provider",
		}
		return nil
	}

	w.oauthState = &oauthFlowState{status: "waiting"}

	config := w.buildOAuthConfig(spec)

	return func() tea.Msg {
		client := oauth.NewClient(nil)
		dcResp, err := client.StartDeviceCodeFlow(context.Background(), config)
		if err != nil {
			return oauthDeviceCodeMsg{err: err}
		}
		return oauthDeviceCodeMsg{resp: dcResp, config: config}
	}
}

// oauthDeviceCodeMsg carries the device code response to the Update loop.
type oauthDeviceCodeMsg struct {
	resp   *oauth.DeviceCodeResponse
	config oauth.DeviceCodeConfig
	err    error
}

// buildOAuthConfig constructs a DeviceCodeConfig from the provider spec.
func (w *ConnectWizard) buildOAuthConfig(spec *ProviderFormSpec) oauth.DeviceCodeConfig {
	config := oauth.DeviceCodeConfig{
		ClientID: spec.OAuthClientID,
	}

	switch spec.Name {
	case "github":
		config.AuthEndpoint = "https://github.com/login/device/code"
		config.TokenEndpoint = "https://github.com/login/oauth/access_token"
		config.Scopes = []string{"repo"}
	}

	return config
}

// cancelOAuthFlow cleans up any in-progress OAuth polling.
func (w *ConnectWizard) cancelOAuthFlow() {
	if w.oauthCancel != nil {
		w.oauthCancel()
		w.oauthCancel = nil
	}
	w.oauthState = nil
}

// defaultLinearEnvTokenFunc checks for LINEAR_API_KEY.
func defaultLinearEnvTokenFunc() (string, string) {
	if v := os.Getenv("LINEAR_API_KEY"); v != "" {
		return v, "LINEAR_API_KEY"
	}
	return "", ""
}

// defaultGitHubEnvTokenFunc checks for GH_TOKEN or GITHUB_TOKEN.
func defaultGitHubEnvTokenFunc() (string, string) {
	if v := os.Getenv("GH_TOKEN"); v != "" {
		return v, "GH_TOKEN"
	}
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		return v, "GITHUB_TOKEN"
	}
	return "", ""
}

// stepDisplayName returns the user-visible step name for the current step.
func (w *ConnectWizard) stepDisplayName() string {
	switch w.step {
	case StepProviderSelect:
		return "Select Provider"
	case StepProviderConfig:
		return "Configure"
	case StepOAuthFlow:
		return "Authenticate"
	case StepSyncConfig:
		return "Sync Settings"
	case StepTestConfirm:
		return "Confirm"
	default:
		return "Unknown"
	}
}

// stepDisplayNumber returns the user-visible step number (OAuth flow is folded
// into the configure step so users see a stable 4-step flow).
func (w *ConnectWizard) stepDisplayNumber() (current, total int) {
	total = 4 // always 4 user-visible steps
	switch w.step {
	case StepProviderSelect:
		return 1, total
	case StepProviderConfig:
		return 2, total
	case StepOAuthFlow:
		return 2, total // part of configure step
	case StepSyncConfig:
		return 3, total
	case StepTestConfirm:
		return 4, total
	default:
		return 1, total
	}
}

// View renders the current wizard step.
func (w *ConnectWizard) View() string {
	var s strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	// Step indicator.
	stepNum, stepTotal := w.stepDisplayNumber()
	stepName := w.stepDisplayName()

	fmt.Fprintf(&s, "%s\n", headerStyle.Render(
		fmt.Sprintf("Connect a Data Source (%d/%d: %s)", stepNum, stepTotal, stepName),
	))
	fmt.Fprintf(&s, "\n")

	// Render step content.
	if w.step == StepOAuthFlow {
		fmt.Fprintf(&s, "%s", w.viewOAuthFlow())
	} else {
		fmt.Fprintf(&s, "%s", w.form.View())
	}

	// Footer hint.
	fmt.Fprintf(&s, "\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	fmt.Fprintf(&s, "%s", hintStyle.Render(" esc:cancel"))

	return s.String()
}

// viewOAuthFlow renders the device code flow step.
func (w *ConnectWizard) viewOAuthFlow() string {
	if w.oauthState == nil {
		return "  Initializing OAuth flow...\n"
	}

	var s strings.Builder

	switch w.oauthState.status {
	case "waiting":
		codeStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Padding(0, 1)

		fmt.Fprintf(&s, "  Open your browser and enter this code:\n\n")
		fmt.Fprintf(&s, "  %s\n\n", codeStyle.Render(w.oauthState.userCode))
		fmt.Fprintf(&s, "  Verification URL: %s\n", w.oauthState.verificationURI)
		if w.oauthState.browserOpened {
			fmt.Fprintf(&s, "  (Browser opened automatically)\n")
		}
		fmt.Fprintf(&s, "\n  Waiting for authorization...\n")

	case "success":
		fmt.Fprintf(&s, "  Authentication successful!\n")

	case "error":
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
		fmt.Fprintf(&s, "  %s\n", errorStyle.Render("Authentication failed: "+w.oauthState.errorMsg))
		fmt.Fprintf(&s, "  Press Esc to cancel and try again.\n")
	}

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
