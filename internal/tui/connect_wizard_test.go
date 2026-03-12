package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func testProviderSpecs() []ProviderFormSpec {
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
			OAuthClientID: "test-oauth-client-id",
			EnvTokenFunc:  func() (string, string) { return "", "" }, // no env token in tests
		},
		{
			Name:        "textfile",
			DisplayName: "Plain text files",
			Description: "Local YAML/text task files",
			AuthType:    AuthNone,
			NeedsPath:   true,
			PathHelp:    "Path to your task file",
		},
		{
			Name:        "reminders",
			DisplayName: "Apple Reminders",
			Description: "macOS Reminders app",
			AuthType:    AuthNone,
		},
	}
}

func TestNewConnectWizard(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := testProviderSpecs()
	w := NewConnectWizard(specs, connMgr)

	if w == nil {
		t.Fatal("NewConnectWizard returned nil")
	}

	if w.Step() != StepProviderSelect {
		t.Errorf("initial step = %v, want StepProviderSelect", w.Step())
	}

	if w.form == nil {
		t.Fatal("form should not be nil after creation")
	}
}

func TestConnectWizard_SetDimensions(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.SetWidth(80)
	w.SetHeight(24)

	if w.width != 80 {
		t.Errorf("width = %d, want 80", w.width)
	}
	if w.height != 24 {
		t.Errorf("height = %d, want 24", w.height)
	}
}

func TestConnectWizard_ViewRendersHeader(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)
	w.SetWidth(80)
	w.SetHeight(24)

	view := w.View()

	if !strings.Contains(view, "Connect a Data Source") {
		t.Error("view should contain 'Connect a Data Source' header")
	}
	if !strings.Contains(view, "1/4") {
		t.Error("view should contain step indicator '1/4'")
	}
	if !strings.Contains(view, "Select Provider") {
		t.Error("view should contain step name 'Select Provider'")
	}
	if !strings.Contains(view, "esc:cancel") {
		t.Error("view should contain 'esc:cancel' hint")
	}
}

func TestConnectWizard_ViewShowsProviders(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)
	w.SetWidth(80)

	view := w.View()

	if !strings.Contains(view, "Todoist") {
		t.Error("view should contain Todoist provider option")
	}
	if !strings.Contains(view, "GitHub Issues") {
		t.Error("view should contain GitHub Issues provider option")
	}
	if !strings.Contains(view, "Plain text files") {
		t.Error("view should contain Plain text files provider option")
	}
}

func TestConnectWizard_CancelOnEsc(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	cmd := w.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("cancel should produce a command")
	}

	msg := cmd()
	if _, ok := msg.(ConnectWizardCancelMsg); !ok {
		t.Errorf("expected ConnectWizardCancelMsg, got %T", msg)
	}

	if !w.cancelled {
		t.Error("wizard should be marked as cancelled")
	}
}

func TestConnectWizard_DefaultProviderSpecs(t *testing.T) {
	t.Parallel()

	specs := DefaultProviderSpecs()
	if len(specs) == 0 {
		t.Fatal("DefaultProviderSpecs returned empty list")
	}

	names := make(map[string]bool)
	for _, spec := range specs {
		if spec.Name == "" {
			t.Error("spec has empty Name")
		}
		if spec.DisplayName == "" {
			t.Errorf("spec %q has empty DisplayName", spec.Name)
		}
		if spec.Description == "" {
			t.Errorf("spec %q has empty Description", spec.Name)
		}
		if names[spec.Name] {
			t.Errorf("duplicate spec name: %s", spec.Name)
		}
		names[spec.Name] = true
	}

	// Verify known providers are present
	expectedProviders := []string{"todoist", "github", "jira", "textfile", "obsidian", "reminders"}
	for _, name := range expectedProviders {
		if !names[name] {
			t.Errorf("expected provider %q not found in DefaultProviderSpecs", name)
		}
	}
}

func TestConnectWizard_DefaultProviderSpecsAuthTypes(t *testing.T) {
	t.Parallel()

	specs := DefaultProviderSpecs()

	authTypes := make(map[string]ProviderAuthType)
	for _, spec := range specs {
		authTypes[spec.Name] = spec.AuthType
	}

	tests := []struct {
		name     string
		provider string
		want     ProviderAuthType
	}{
		{"todoist needs API token", "todoist", AuthAPIToken},
		{"github needs OAuth", "github", AuthOAuth},
		{"jira needs API token", "jira", AuthAPIToken},
		{"textfile needs no auth", "textfile", AuthNone},
		{"reminders needs no auth", "reminders", AuthNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := authTypes[tt.provider]
			if !ok {
				t.Fatalf("provider %q not found", tt.provider)
			}
			if got != tt.want {
				t.Errorf("auth type for %q = %v, want %v", tt.provider, got, tt.want)
			}
		})
	}
}

func TestConnectWizard_NeedsPathProviders(t *testing.T) {
	t.Parallel()

	specs := DefaultProviderSpecs()

	for _, spec := range specs {
		switch spec.Name {
		case "textfile", "obsidian":
			if !spec.NeedsPath {
				t.Errorf("provider %q should have NeedsPath=true", spec.Name)
			}
			if spec.PathHelp == "" {
				t.Errorf("provider %q should have PathHelp", spec.Name)
			}
		default:
			if spec.NeedsPath {
				t.Errorf("provider %q should have NeedsPath=false", spec.Name)
			}
		}
	}
}

func TestConnectWizard_TokenHelpForAPITokenProviders(t *testing.T) {
	t.Parallel()

	specs := DefaultProviderSpecs()
	for _, spec := range specs {
		if spec.AuthType == AuthAPIToken && spec.TokenHelp == "" {
			t.Errorf("provider %q with AuthAPIToken should have TokenHelp", spec.Name)
		}
	}
}

func TestConnectWizard_StepProgression(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	// Step 1: Select provider
	if w.Step() != StepProviderSelect {
		t.Fatalf("step = %v, want StepProviderSelect", w.Step())
	}

	// Simulate provider selection by setting value and marking form complete
	w.selectedProvider = "reminders"
	w.form = huh.NewForm(huh.NewGroup(huh.NewNote().Title("done")))
	w.form.Init()
	// Force the form state to completed to simulate user completing the form
	advanceCmd := w.advanceStep()
	if w.Step() != StepProviderConfig {
		t.Fatalf("after advancing from step 1, step = %v, want StepProviderConfig", w.Step())
	}
	if advanceCmd == nil {
		t.Error("advancing should return a command for form init")
	}

	// Step 2 → Step 3
	w.label = "My Reminders"
	advanceCmd = w.advanceStep()
	if w.Step() != StepSyncConfig {
		t.Fatalf("after advancing from step 2, step = %v, want StepSyncConfig", w.Step())
	}
	if advanceCmd == nil {
		t.Error("advancing should return a command for form init")
	}

	// Step 3 → Step 4
	w.syncMode = "readonly"
	w.pollInterval = "5m"
	advanceCmd = w.advanceStep()
	if w.Step() != StepTestConfirm {
		t.Fatalf("after advancing from step 3, step = %v, want StepTestConfirm", w.Step())
	}
	if advanceCmd == nil {
		t.Error("advancing should return a command for form init")
	}

	// Step 4 → Complete
	advanceCmd = w.advanceStep()
	if !w.finished {
		t.Error("wizard should be finished after step 4")
	}
	if advanceCmd == nil {
		t.Fatal("completion should produce a command")
	}

	msg := advanceCmd()
	completeMsg, ok := msg.(ConnectWizardCompleteMsg)
	if !ok {
		t.Fatalf("expected ConnectWizardCompleteMsg, got %T", msg)
	}
	if completeMsg.ProviderName != "reminders" {
		t.Errorf("provider = %q, want %q", completeMsg.ProviderName, "reminders")
	}
	if completeMsg.Label != "My Reminders" {
		t.Errorf("label = %q, want %q", completeMsg.Label, "My Reminders")
	}
	if completeMsg.SyncMode != "readonly" {
		t.Errorf("sync mode = %q, want %q", completeMsg.SyncMode, "readonly")
	}
	if completeMsg.PollInterval != 5*time.Minute {
		t.Errorf("poll interval = %v, want 5m", completeMsg.PollInterval)
	}
}

func TestConnectWizard_DefaultLabelUsesDisplayName(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "todoist"
	w.label = "" // empty label

	// Advance from step 1
	w.advanceStep()

	// Advance from step 2 — should default label
	w.apiToken = "test-token"
	w.advanceStep()

	if w.label != "Todoist" {
		t.Errorf("label = %q, want %q (should default to DisplayName)", w.label, "Todoist")
	}
}

func TestConnectWizard_BuildSummary(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "todoist"
	w.label = "Work Todoist"
	w.apiToken = "secret-token"
	w.syncMode = "bidirectional"
	w.pollInterval = "1m"

	summary := w.buildSummary()

	if !strings.Contains(summary, "todoist") {
		t.Error("summary should contain provider name")
	}
	if !strings.Contains(summary, "Work Todoist") {
		t.Error("summary should contain label")
	}
	if !strings.Contains(summary, "••••••••") {
		t.Error("summary should mask the token")
	}
	if strings.Contains(summary, "secret-token") {
		t.Error("summary must NOT contain the actual token")
	}
	if !strings.Contains(summary, "bidirectional") {
		t.Error("summary should contain sync mode")
	}
	if !strings.Contains(summary, "1m") {
		t.Error("summary should contain poll interval")
	}
}

func TestConnectWizard_BuildSummaryWithPath(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "textfile"
	w.label = "My Tasks"
	w.filePath = "/tmp/tasks.yaml"
	w.syncMode = "readonly"
	w.pollInterval = "5m"

	summary := w.buildSummary()

	if !strings.Contains(summary, "/tmp/tasks.yaml") {
		t.Error("summary should contain file path")
	}
}

func TestConnectWizard_CompleteMsgSettings(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "textfile"
	w.label = "My Tasks"
	w.filePath = "/tmp/tasks.yaml"
	w.syncMode = "bidirectional"
	w.pollInterval = "30s"
	w.step = StepTestConfirm

	cmd := w.advanceStep()
	if cmd == nil {
		t.Fatal("expected completion command")
	}

	msg := cmd()
	completeMsg, ok := msg.(ConnectWizardCompleteMsg)
	if !ok {
		t.Fatalf("expected ConnectWizardCompleteMsg, got %T", msg)
	}

	if completeMsg.Settings["path"] != "/tmp/tasks.yaml" {
		t.Errorf("settings path = %q, want %q", completeMsg.Settings["path"], "/tmp/tasks.yaml")
	}
	if completeMsg.PollInterval != 30*time.Second {
		t.Errorf("poll interval = %v, want 30s", completeMsg.PollInterval)
	}
}

func TestConnectWizard_Step2FormForAPIToken(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "todoist"
	w.buildStep2Form()

	view := w.form.View()
	if !strings.Contains(view, "Give this connection a name") {
		t.Error("step 2 should contain label field")
	}
	if !strings.Contains(view, "API Token") {
		t.Error("step 2 for API token provider should contain API Token field")
	}
}

func TestConnectWizard_Step2FormForLocalProvider(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "textfile"
	w.buildStep2Form()

	view := w.form.View()
	if !strings.Contains(view, "Give this connection a name") {
		t.Error("step 2 should contain label field")
	}
	if !strings.Contains(view, "Path") {
		t.Error("step 2 for local provider should contain path field")
	}
}

func TestConnectWizard_Step2FormForOAuth(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "github"
	w.buildStep2Form()

	view := w.form.View()
	if !strings.Contains(view, "Authentication method") {
		t.Error("step 2 for OAuth provider should contain auth method selection")
	}
	if !strings.Contains(view, "Personal Access Token") {
		t.Error("step 2 for OAuth provider should contain PAT option")
	}
}

func TestConnectWizard_Step2FormForNoAuth(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "reminders"
	w.buildStep2Form()

	view := w.form.View()
	if !strings.Contains(view, "Give this connection a name") {
		t.Error("step 2 should contain label field even for no-auth providers")
	}
	// Should NOT have API token or path fields
	if strings.Contains(view, "API Token") {
		t.Error("step 2 for no-auth provider should not contain API Token field")
	}
}

func TestConnectWizard_Step3FormSyncOptions(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.buildStep3Form()

	view := w.form.View()
	if !strings.Contains(view, "Sync mode") {
		t.Error("step 3 should contain sync mode selection")
	}
	if !strings.Contains(view, "Poll interval") {
		t.Error("step 3 should contain poll interval selection")
	}
}

func TestConnectWizard_Step4FormCreation(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "todoist"
	w.label = "Work Tasks"
	w.apiToken = "tok_123"
	w.syncMode = "bidirectional"
	w.pollInterval = "1m"

	// Verify buildStep4Form doesn't panic and creates a form
	w.buildStep4Form()
	if w.form == nil {
		t.Fatal("step 4 form should not be nil")
	}

	// Verify the form renders without error (huh output depends on terminal state,
	// so we just check non-empty and non-panic)
	view := w.form.View()
	if view == "" {
		t.Error("step 4 form view should not be empty")
	}
}

func TestExpandPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string // just check prefix for ~ expansion
	}{
		{"absolute path unchanged", "/tmp/tasks.yaml", "/tmp/tasks.yaml"},
		{"relative path unchanged", "tasks.yaml", "tasks.yaml"},
		{"tilde expands", "~/tasks.yaml", "/"}, // just verify it starts with /
		{"tilde alone not expanded", "~tasks", "~tasks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := expandPath(tt.path)
			if tt.path == "~/tasks.yaml" {
				if !strings.HasPrefix(result, "/") {
					t.Errorf("expandPath(%q) = %q, expected to start with /", tt.path, result)
				}
				if !strings.HasSuffix(result, "/tasks.yaml") {
					t.Errorf("expandPath(%q) = %q, expected to end with /tasks.yaml", tt.path, result)
				}
			} else {
				if result != tt.want {
					t.Errorf("expandPath(%q) = %q, want %q", tt.path, result, tt.want)
				}
			}
		})
	}
}

func TestWizardStepString(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	tests := []struct {
		step WizardStep
		want string
	}{
		{StepProviderSelect, "Select Provider"},
		{StepProviderConfig, "Configure"},
		{StepOAuthFlow, "Authenticate"},
		{StepSyncConfig, "Sync Settings"},
		{StepTestConfirm, "Confirm"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			w := *w // copy
			w.step = tt.step
			if got := w.stepDisplayName(); got != tt.want {
				t.Errorf("stepDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConnectWizard_ViewModeEnum(t *testing.T) {
	t.Parallel()

	if ViewConnectWizard.String() != "ConnectWizard" {
		t.Errorf("ViewConnectWizard.String() = %q, want %q", ViewConnectWizard.String(), "ConnectWizard")
	}
}

func TestConnectWizard_CommandRegistered(t *testing.T) {
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

func TestProviderAuthTypeConstants(t *testing.T) {
	t.Parallel()

	if AuthNone != 0 {
		t.Errorf("AuthNone = %d, want 0", AuthNone)
	}
	if AuthAPIToken != 1 {
		t.Errorf("AuthAPIToken = %d, want 1", AuthAPIToken)
	}
	if AuthOAuth != 2 {
		t.Errorf("AuthOAuth = %d, want 2", AuthOAuth)
	}
}

func TestConnectWizard_GetSelectedSpec(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	// No selection yet
	w.selectedProvider = ""
	if spec := w.getSelectedSpec(); spec != nil {
		t.Error("getSelectedSpec should return nil when no provider selected")
	}

	// Valid selection
	w.selectedProvider = "todoist"
	spec := w.getSelectedSpec()
	if spec == nil {
		t.Fatal("getSelectedSpec returned nil for valid provider")
	}
	if spec.Name != "todoist" {
		t.Errorf("spec.Name = %q, want %q", spec.Name, "todoist")
	}

	// Invalid selection
	w.selectedProvider = "nonexistent"
	if spec := w.getSelectedSpec(); spec != nil {
		t.Error("getSelectedSpec should return nil for unknown provider")
	}
}

func TestConnectWizard_Step2FormForOAuth_WithEnvToken(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := []ProviderFormSpec{
		{
			Name:          "github",
			DisplayName:   "GitHub Issues",
			Description:   "GitHub repository issues",
			AuthType:      AuthOAuth,
			OAuthClientID: "test-client-id",
			EnvTokenFunc:  func() (string, string) { return "gho_env_token", "GH_TOKEN" },
		},
	}
	w := NewConnectWizard(specs, connMgr)

	w.selectedProvider = "github"
	w.buildStep2Form()

	view := w.form.View()
	if !strings.Contains(view, "Authentication method") {
		t.Error("should contain auth method selection")
	}
	if !strings.Contains(view, "GH_TOKEN") {
		t.Error("should show env var name when env token detected")
	}
	// Default auth method should be env when env token is available.
	if w.authMethod != "env" {
		t.Errorf("authMethod = %q, want %q when env token detected", w.authMethod, "env")
	}
}

func TestConnectWizard_Step2FormForOAuth_NoOAuthClientID(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := []ProviderFormSpec{
		{
			Name:        "github",
			DisplayName: "GitHub Issues",
			Description: "GitHub repository issues",
			AuthType:    AuthOAuth,
			// No OAuthClientID — falls back to PAT only
			EnvTokenFunc: func() (string, string) { return "", "" },
		},
	}
	w := NewConnectWizard(specs, connMgr)

	w.selectedProvider = "github"
	w.buildStep2Form()

	// When no OAuth client ID, only PAT should be available.
	if w.authMethod != "pat" {
		t.Errorf("authMethod = %q, want %q when no OAuth client ID", w.authMethod, "pat")
	}
}

func TestConnectWizard_AdvanceStep_EnvTokenUsed(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := []ProviderFormSpec{
		{
			Name:          "github",
			DisplayName:   "GitHub Issues",
			Description:   "GitHub repository issues",
			AuthType:      AuthOAuth,
			OAuthClientID: "test-client-id",
			EnvTokenFunc:  func() (string, string) { return "gho_env_token_123", "GH_TOKEN" },
		},
	}
	w := NewConnectWizard(specs, connMgr)

	// Simulate provider selection done.
	w.selectedProvider = "github"
	w.step = StepProviderConfig
	w.buildStep2Form()

	// User chose "env" auth method.
	w.authMethod = "env"
	w.label = "My GitHub"

	// Advance from config step.
	w.advanceStep()

	// Should skip OAuth flow and go to sync config.
	if w.step != StepSyncConfig {
		t.Errorf("step = %v, want StepSyncConfig (should skip OAuth flow)", w.step)
	}
	if w.apiToken != "gho_env_token_123" {
		t.Errorf("apiToken = %q, want %q (env token should be used)", w.apiToken, "gho_env_token_123")
	}
}

func TestConnectWizard_AdvanceStep_OAuthSelected(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := []ProviderFormSpec{
		{
			Name:          "github",
			DisplayName:   "GitHub Issues",
			Description:   "GitHub repository issues",
			AuthType:      AuthOAuth,
			OAuthClientID: "test-client-id",
			EnvTokenFunc:  func() (string, string) { return "", "" },
		},
	}
	w := NewConnectWizard(specs, connMgr)

	w.selectedProvider = "github"
	w.step = StepProviderConfig
	w.buildStep2Form()
	w.authMethod = "oauth"
	w.label = "My GitHub"

	// Advance from config step — should enter OAuth flow.
	cmd := w.advanceStep()

	if w.step != StepOAuthFlow {
		t.Errorf("step = %v, want StepOAuthFlow", w.step)
	}
	if cmd == nil {
		t.Error("OAuth flow should return a command to start device code flow")
	}
}

func TestConnectWizard_AdvanceStep_PATSelected(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	specs := []ProviderFormSpec{
		{
			Name:          "github",
			DisplayName:   "GitHub Issues",
			Description:   "GitHub repository issues",
			AuthType:      AuthOAuth,
			OAuthClientID: "test-client-id",
			EnvTokenFunc:  func() (string, string) { return "", "" },
		},
	}
	w := NewConnectWizard(specs, connMgr)

	w.selectedProvider = "github"
	w.step = StepProviderConfig
	w.buildStep2Form()
	w.authMethod = "pat"
	w.apiToken = "ghp_manual_token"
	w.label = "My GitHub"

	// Advance from config step — should skip OAuth flow.
	w.advanceStep()

	if w.step != StepSyncConfig {
		t.Errorf("step = %v, want StepSyncConfig (PAT should skip OAuth)", w.step)
	}
	if w.apiToken != "ghp_manual_token" {
		t.Errorf("apiToken = %q, want %q", w.apiToken, "ghp_manual_token")
	}
}

func TestConnectWizard_OAuthFlow_TokenResult(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "github"
	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{
		userCode:        "ABCD-1234",
		verificationURI: "https://github.com/login/device",
		status:          "waiting",
	}
	w.label = "My GitHub"

	// Simulate successful token result.
	cmd := w.updateOAuthFlow(oauthTokenResultMsg{token: "gho_oauth_token"})

	if w.apiToken != "gho_oauth_token" {
		t.Errorf("apiToken = %q, want %q", w.apiToken, "gho_oauth_token")
	}
	if w.oauthState.status != "success" {
		t.Errorf("status = %q, want %q", w.oauthState.status, "success")
	}
	// Should have returned advanceStep command.
	if cmd == nil {
		t.Error("successful token result should return a command to advance")
	}
}

func TestConnectWizard_OAuthFlow_TokenError(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "github"
	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{status: "waiting"}

	// Simulate token error.
	cmd := w.updateOAuthFlow(oauthTokenResultMsg{err: fmt.Errorf("access denied")})

	if w.oauthState.status != "error" {
		t.Errorf("status = %q, want %q", w.oauthState.status, "error")
	}
	if !strings.Contains(w.oauthState.errorMsg, "access denied") {
		t.Errorf("errorMsg = %q, should contain 'access denied'", w.oauthState.errorMsg)
	}
	if cmd != nil {
		t.Error("error should not return a command")
	}
}

func TestConnectWizard_OAuthFlow_DeviceCodeError(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "github"
	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{status: "waiting"}

	// Simulate device code request error.
	cmd := w.updateOAuthFlow(oauthDeviceCodeMsg{err: fmt.Errorf("network error")})

	if w.oauthState.status != "error" {
		t.Errorf("status = %q, want %q", w.oauthState.status, "error")
	}
	if !strings.Contains(w.oauthState.errorMsg, "network error") {
		t.Errorf("errorMsg = %q, should contain 'network error'", w.oauthState.errorMsg)
	}
	if cmd != nil {
		t.Error("error should not return a command")
	}
}

func TestConnectWizard_OAuthFlowView_Waiting(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{
		userCode:        "ABCD-1234",
		verificationURI: "https://github.com/login/device",
		status:          "waiting",
		browserOpened:   true,
	}

	view := w.viewOAuthFlow()
	if !strings.Contains(view, "ABCD-1234") {
		t.Error("should display user code")
	}
	if !strings.Contains(view, "github.com/login/device") {
		t.Error("should display verification URL")
	}
	if !strings.Contains(view, "Browser opened") {
		t.Error("should indicate browser was opened")
	}
	if !strings.Contains(view, "Waiting for authorization") {
		t.Error("should show waiting message")
	}
}

func TestConnectWizard_OAuthFlowView_Error(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{
		status:   "error",
		errorMsg: "token expired",
	}

	view := w.viewOAuthFlow()
	if !strings.Contains(view, "token expired") {
		t.Error("should display error message")
	}
	if !strings.Contains(view, "failed") {
		t.Error("should indicate authentication failed")
	}
}

func TestConnectWizard_OAuthFlowView_Success(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{status: "success"}

	view := w.viewOAuthFlow()
	if !strings.Contains(view, "successful") {
		t.Error("should display success message")
	}
}

func TestConnectWizard_OAuthFlowView_NilState(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.step = StepOAuthFlow
	w.oauthState = nil

	view := w.viewOAuthFlow()
	if !strings.Contains(view, "Initializing") {
		t.Error("should display initializing message when state is nil")
	}
}

func TestConnectWizard_StepDisplayNumber(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	tests := []struct {
		step      WizardStep
		wantNum   int
		wantTotal int
	}{
		{StepProviderSelect, 1, 4},
		{StepProviderConfig, 2, 4},
		{StepOAuthFlow, 2, 4}, // OAuth is part of configure step
		{StepSyncConfig, 3, 4},
		{StepTestConfirm, 4, 4},
	}

	for _, tt := range tests {
		w.step = tt.step
		num, total := w.stepDisplayNumber()
		if num != tt.wantNum {
			t.Errorf("step %v: num = %d, want %d", tt.step, num, tt.wantNum)
		}
		if total != tt.wantTotal {
			t.Errorf("step %v: total = %d, want %d", tt.step, total, tt.wantTotal)
		}
	}
}

func TestConnectWizard_CancelDuringOAuthFlow(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.step = StepOAuthFlow
	w.oauthState = &oauthFlowState{status: "waiting"}

	// Send Esc.
	cmd := w.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if !w.cancelled {
		t.Error("wizard should be cancelled")
	}
	if w.oauthState != nil {
		t.Error("OAuth state should be cleaned up on cancel")
	}
	if cmd == nil {
		t.Fatal("cancel should produce a command")
	}
	msg := cmd()
	if _, ok := msg.(ConnectWizardCancelMsg); !ok {
		t.Errorf("expected ConnectWizardCancelMsg, got %T", msg)
	}
}

func TestConnectWizard_CompleteMsgIncludesToken(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	w.selectedProvider = "github"
	w.label = "My GitHub"
	w.apiToken = "gho_test_token"
	w.syncMode = "readonly"
	w.pollInterval = "5m"
	w.step = StepTestConfirm

	cmd := w.advanceStep()
	if cmd == nil {
		t.Fatal("expected completion command")
	}

	msg := cmd()
	completeMsg, ok := msg.(ConnectWizardCompleteMsg)
	if !ok {
		t.Fatalf("expected ConnectWizardCompleteMsg, got %T", msg)
	}
	if completeMsg.Token != "gho_test_token" {
		t.Errorf("Token = %q, want %q", completeMsg.Token, "gho_test_token")
	}
}

func TestDefaultGitHubEnvTokenFunc(t *testing.T) {
	// This is not parallel — it modifies env vars.
	t.Setenv("GH_TOKEN", "gho_test_env")
	t.Setenv("GITHUB_TOKEN", "")

	token, envVar := defaultGitHubEnvTokenFunc()
	if token != "gho_test_env" {
		t.Errorf("token = %q, want %q", token, "gho_test_env")
	}
	if envVar != "GH_TOKEN" {
		t.Errorf("envVar = %q, want %q", envVar, "GH_TOKEN")
	}
}

func TestDefaultGitHubEnvTokenFunc_GitHubToken(t *testing.T) {
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "ghp_test_env")

	token, envVar := defaultGitHubEnvTokenFunc()
	if token != "ghp_test_env" {
		t.Errorf("token = %q, want %q", token, "ghp_test_env")
	}
	if envVar != "GITHUB_TOKEN" {
		t.Errorf("envVar = %q, want %q", envVar, "GITHUB_TOKEN")
	}
}

func TestDefaultGitHubEnvTokenFunc_Neither(t *testing.T) {
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	token, envVar := defaultGitHubEnvTokenFunc()
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
	if envVar != "" {
		t.Errorf("envVar = %q, want empty", envVar)
	}
}

func TestConnectWizard_BuildOAuthConfig_GitHub(t *testing.T) {
	t.Parallel()

	connMgr := connection.NewConnectionManager(nil)
	w := NewConnectWizard(testProviderSpecs(), connMgr)

	spec := &ProviderFormSpec{
		Name:          "github",
		OAuthClientID: "test-oauth-id",
	}

	config := w.buildOAuthConfig(spec)

	if config.ClientID != "test-oauth-id" {
		t.Errorf("ClientID = %q, want %q", config.ClientID, "test-oauth-id")
	}
	if config.AuthEndpoint != "https://github.com/login/device/code" {
		t.Errorf("AuthEndpoint = %q", config.AuthEndpoint)
	}
	if config.TokenEndpoint != "https://github.com/login/oauth/access_token" {
		t.Errorf("TokenEndpoint = %q", config.TokenEndpoint)
	}
	if len(config.Scopes) != 1 || config.Scopes[0] != "repo" {
		t.Errorf("Scopes = %v, want [repo]", config.Scopes)
	}
}
