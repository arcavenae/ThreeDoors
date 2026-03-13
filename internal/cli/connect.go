package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// connectResultJSON is the JSON representation of a connect command result.
type connectResultJSON struct {
	Name     string           `json:"name"`
	Provider string           `json:"provider"`
	ID       string           `json:"id"`
	Status   string           `json:"status"`
	Test     *connectTestJSON `json:"test,omitempty"`
}

// connectTestJSON is the JSON representation of a connection test result.
type connectTestJSON struct {
	Healthy bool                   `json:"healthy"`
	Checks  []sourcesTestCheckJSON `json:"checks"`
}

// providerFlagSpec defines flag requirements and settings mapping for a provider.
type providerFlagSpec struct {
	tokenKey      string            // settings key for token (e.g., "api_token"); empty if no token
	required      []string          // flag names required beyond --label
	flagToSetting map[string]string // flag name → settings key
}

// knownProviderSpecs maps provider names to their flag specifications.
var knownProviderSpecs = map[string]providerFlagSpec{
	"todoist": {
		tokenKey: "api_token",
		required: nil,
		flagToSetting: map[string]string{
			"project-ids": "project_ids",
			"filter":      "filter",
		},
	},
	"github": {
		tokenKey: "token",
		required: []string{"repos"},
		flagToSetting: map[string]string{
			"repos": "repos",
		},
	},
	"jira": {
		tokenKey: "api_token",
		required: []string{"server"},
		flagToSetting: map[string]string{
			"server": "url",
		},
	},
	"textfile": {
		tokenKey: "",
		required: []string{"path"},
		flagToSetting: map[string]string{
			"path": "path",
		},
	},
}

func newConnectCmd() *cobra.Command {
	var (
		label      string
		token      string
		repos      string
		server     string
		path       string
		projectIDs string
		filter     string
	)

	cmd := &cobra.Command{
		Use:   "connect <provider>",
		Short: "Configure a data source connection",
		Long: `Configure a data source connection. Supported providers: todoist, github, jira, textfile.

When called with --label and other flags, runs in non-interactive mode.
When called without flags in a terminal, launches the interactive setup wizard.`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"todoist", "github", "jira", "textfile", "applenotes", "obsidian", "reminders", "linear"},
		RunE: func(cmd *cobra.Command, args []string) error {
			provider := args[0]
			jsonMode := isJSONOutput(cmd)

			if label == "" {
				formatter := NewOutputFormatter(os.Stdout, jsonMode)
				if jsonMode {
					return formatter.WriteJSONError("connect", ExitValidation,
						"interactive mode not supported with --json",
						"provide --label and other required flags")
				}
				return runConnectWizard(provider, os.Stdout)
			}

			// Build flag values map from provided flags.
			flagValues := map[string]string{
				"repos":       repos,
				"server":      server,
				"path":        path,
				"project-ids": projectIDs,
				"filter":      filter,
			}

			configPath, svc, manager, err := bootstrapForConnect()
			if err != nil {
				formatter := NewOutputFormatter(os.Stdout, jsonMode)
				if jsonMode {
					return formatter.WriteJSONError("connect", ExitGeneralError, err.Error(), "")
				}
				return err
			}

			return runConnectTo(provider, label, token, flagValues, svc, manager, os.Stdout, jsonMode, configPath)
		},
	}

	cmd.Flags().StringVar(&label, "label", "", "user-friendly connection name (required for non-interactive)")
	cmd.Flags().StringVar(&token, "token", "", "authentication token (or set via environment variable)")
	cmd.Flags().StringVar(&repos, "repos", "", "comma-separated repository list (github)")
	cmd.Flags().StringVar(&server, "server", "", "server URL (jira)")
	cmd.Flags().StringVar(&path, "path", "", "file path (textfile)")
	cmd.Flags().StringVar(&projectIDs, "project-ids", "", "comma-separated project IDs (todoist)")
	cmd.Flags().StringVar(&filter, "filter", "", "task filter expression (todoist)")

	return cmd
}

// bootstrapForConnect sets up the ConnectionService and Manager for adding
// a new connection. Unlike bootstrap(), it does not initialize providers or
// load tasks — only what's needed for connection management.
func bootstrapForConnect() (string, *connection.ConnectionService, *connection.ConnectionManager, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return "", nil, nil, fmt.Errorf("config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return "", nil, nil, fmt.Errorf("load config: %w", err)
	}

	// Try to resolve existing connections so they're preserved when persisting.
	resolved, _ := connection.ResolveFromConfig(cfg, core.DefaultRegistry(), configPath, nil)
	if resolved != nil {
		return configPath, resolved.Service, resolved.Manager, nil
	}

	// No existing connections or all failed to resolve — create fresh service.
	manager := connection.NewConnectionManager(nil)
	creds := connection.NewEnvCredentialStore()
	svc, err := connection.NewConnectionService(connection.ServiceConfig{
		Manager:    manager,
		Creds:      creds,
		ConfigPath: configPath,
	})
	if err != nil {
		return "", nil, nil, fmt.Errorf("create connection service: %w", err)
	}

	return configPath, svc, manager, nil
}

// runConnectTo validates flags, creates a connection, runs a health test, and
// writes the result. Extracted for testability.
func runConnectTo(
	provider, label, token string,
	flagValues map[string]string,
	svc *connection.ConnectionService,
	manager *connection.ConnectionManager,
	w io.Writer,
	jsonMode bool,
	configPath string,
) error {
	formatter := NewOutputFormatter(w, jsonMode)

	// Build settings map from flags.
	settings, err := buildConnectSettings(provider, token, flagValues)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("connect", ExitValidation, err.Error(), "")
		}
		return err
	}

	// Create the connection.
	conn, err := svc.Add(provider, label, settings, token)
	if err != nil {
		if jsonMode {
			return formatter.WriteJSONError("connect", ExitGeneralError,
				fmt.Sprintf("create connection: %v", err), "")
		}
		return fmt.Errorf("create connection: %w", err)
	}

	// Run connection test (best-effort — report result but don't fail the command).
	var testResult *connectTestJSON
	healthResult, testErr := svc.TestConnection(conn.ID)
	if testErr == nil {
		testResult = formatConnectTest(healthResult)
	}

	if jsonMode {
		data := connectResultJSON{
			Name:     conn.Label,
			Provider: conn.ProviderName,
			ID:       conn.ID,
			Status:   conn.State.String(),
			Test:     testResult,
		}
		return formatter.WriteJSON("connect", data, nil)
	}

	// Human-readable output.
	_ = formatter.Writef("Connection created:\n")
	_ = formatter.Writef("  Name:     %s\n", conn.Label)
	_ = formatter.Writef("  Provider: %s\n", conn.ProviderName)
	_ = formatter.Writef("  ID:       %s\n", conn.ID)
	_ = formatter.Writef("  Status:   %s\n", conn.State.String())

	if testResult != nil {
		_ = formatter.Writef("\nConnection test:\n")
		for _, c := range testResult.Checks {
			icon := "✓"
			if !c.Passed {
				icon = "✗"
			}
			_ = formatter.Writef("  %s %s\n", icon, c.Name)
		}
	} else if testErr != nil {
		_ = formatter.Writef("\nConnection test: skipped (%v)\n", testErr)
	}

	return nil
}

// buildConnectSettings validates required flags and maps flag values to
// provider settings.
func buildConnectSettings(provider, token string, flagValues map[string]string) (map[string]string, error) {
	spec, known := knownProviderSpecs[provider]
	settings := make(map[string]string)

	if known {
		// Check required flags.
		var missing []string
		for _, req := range spec.required {
			if flagValues[req] == "" {
				missing = append(missing, "--"+req)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("missing required flags for %s: %s", provider, strings.Join(missing, ", "))
		}

		// Map flags to settings.
		for flag, settingKey := range spec.flagToSetting {
			if v := flagValues[flag]; v != "" {
				settings[settingKey] = v
			}
		}

		// Store token in provider-appropriate settings key.
		if token != "" && spec.tokenKey != "" {
			settings[spec.tokenKey] = token
		}
	} else {
		// Unknown provider — pass all non-empty flag values as settings.
		for flag, v := range flagValues {
			if v != "" {
				settings[flag] = v
			}
		}
		if token != "" {
			settings["token"] = token
		}
	}

	return settings, nil
}

// WizardRunnerFactory creates a tea.Model that wraps the connect wizard for
// standalone CLI use. The provider arg pre-selects a provider. The returned
// model should handle ConnectWizardCompleteMsg/CancelMsg internally.
// The resultWriter receives post-program human-readable output.
// The errorFunc returns any error from the wizard run.
type WizardRunnerFactory func(
	provider string,
	svc *connection.ConnectionService,
	manager *connection.ConnectionManager,
) (model tea.Model, resultWriter func(io.Writer), errorFunc func() error)

// NewConnectWizardRunner is set by cmd/threedoors to wire the tui package
// into the CLI without creating an import cycle.
var NewConnectWizardRunner WizardRunnerFactory

// runConnectWizard launches the interactive connect wizard in a standalone
// Bubbletea program. It detects TTY, bootstraps services, and runs the wizard.
func runConnectWizard(provider string, w io.Writer) error {
	if !isTerminal(os.Stdin.Fd()) {
		return fmt.Errorf("interactive wizard requires a terminal; use --label and other flags in non-interactive mode")
	}

	if NewConnectWizardRunner == nil {
		return fmt.Errorf("interactive wizard not available")
	}

	_, svc, manager, err := bootstrapForConnect()
	if err != nil {
		return err
	}

	model, resultWriter, errorFunc := NewConnectWizardRunner(provider, svc, manager)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("wizard: %w", err)
	}

	if err := errorFunc(); err != nil {
		return err
	}

	resultWriter(w)
	return nil
}

// formatConnectTest converts a HealthCheckResult into a connectTestJSON.
func formatConnectTest(result connection.HealthCheckResult) *connectTestJSON {
	checks := []sourcesTestCheckJSON{
		{Name: "DNS resolution", Passed: result.APIReachable},
		{Name: "TLS", Passed: result.APIReachable},
		{Name: "Authentication", Passed: result.TokenValid},
		{Name: "Authorization", Passed: result.TokenValid},
		{Name: "Rate limit", Passed: result.RateLimitOK},
	}
	return &connectTestJSON{
		Healthy: result.Healthy(),
		Checks:  checks,
	}
}
