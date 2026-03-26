package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/arcavenae/ThreeDoors/internal/intelligence/llm"
	"github.com/spf13/cobra"
)

// llmStatusJSON is the JSON envelope data for the llm status command.
type llmStatusJSON struct {
	Backend   *llmBackendJSON   `json:"backend"`
	Fallbacks []llmFallbackJSON `json:"fallbacks"`
	Services  []llmServiceJSON  `json:"services"`
}

// llmBackendJSON describes the active backend in JSON output.
type llmBackendJSON struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	CommandPath string `json:"command_path,omitempty"`
	Available   bool   `json:"available"`
}

// llmFallbackJSON describes a fallback backend in JSON output.
type llmFallbackJSON struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// llmServiceJSON describes a service readiness entry in JSON output.
type llmServiceJSON struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
}

// newLLMStatusCmd creates the "llm status" subcommand.
func newLLMStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show LLM backend status",
		Long: `Display which LLM backend is active, fallback availability,
and service readiness for intelligence features.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLLMStatus(cmd)
		},
	}
}

func runLLMStatus(cmd *cobra.Command) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	cfg, err := loadLLMConfig()
	if err != nil {
		if isJSON {
			return formatter.WriteJSONError("llm status", ExitProviderError, err.Error(), "")
		}
		return fmt.Errorf("load config: %w", err)
	}

	return runLLMStatusWith(cmd.Context(), formatter, cfg, exec.LookPath, os.Getenv)
}

// runLLMStatusWith is the testable core of runLLMStatus.
func runLLMStatusWith(ctx context.Context, formatter *OutputFormatter, cfg llm.Config, lookPath llm.LookPathFunc, getenv func(string) string) error {
	_, discovery, _ := llm.DiscoverBackendWith(ctx, cfg, lookPath, getenv)

	if discovery == nil {
		discovery = &llm.DiscoveryResult{}
	}

	// Determine backend type and command path.
	backendType := classifyBackendType(discovery.Selected)
	commandPath := resolveCommandPath(discovery.Selected, lookPath)

	// Check availability of the selected backend.
	available := discovery.Selected != ""

	// Build service readiness based on whether a backend is available.
	svcReady := available

	if formatter.IsJSON() {
		return writeLLMStatusJSON(formatter, discovery, backendType, commandPath, available, svcReady)
	}
	return writeLLMStatusHuman(formatter, discovery, backendType, commandPath, available, svcReady)
}

func writeLLMStatusJSON(formatter *OutputFormatter, discovery *llm.DiscoveryResult, backendType, commandPath string, available, svcReady bool) error {
	var backend *llmBackendJSON
	if discovery.Selected != "" {
		backend = &llmBackendJSON{
			Name:        discovery.Selected,
			Type:        backendType,
			CommandPath: commandPath,
			Available:   available,
		}
	}

	fallbacks := buildFallbacksJSON(discovery)
	services := buildServicesJSON(svcReady)

	data := llmStatusJSON{
		Backend:   backend,
		Fallbacks: fallbacks,
		Services:  services,
	}
	return formatter.WriteJSON("llm status", data, nil)
}

func writeLLMStatusHuman(formatter *OutputFormatter, discovery *llm.DiscoveryResult, backendType, commandPath string, available, svcReady bool) error {
	_ = formatter.Writef("LLM Backend Status\n\n")

	if discovery.Selected == "" {
		_ = formatter.Writef("  No LLM backends available.\n\n")
		_ = formatter.Writef("  Install one of the following to enable LLM features:\n")
		_ = formatter.Writef("    - claude   https://docs.anthropic.com/en/docs/claude-cli\n")
		_ = formatter.Writef("    - gemini   https://github.com/google-gemini/gemini-cli\n")
		_ = formatter.Writef("    - ollama   https://ollama.com/download\n")
		return nil
	}

	statusStr := "reachable"
	if !available {
		statusStr = "unreachable"
	}

	_ = formatter.Writef("  Active Backend\n")
	_ = formatter.Writef("    Name:      %s\n", discovery.Selected)
	_ = formatter.Writef("    Type:      %s\n", backendType)
	if commandPath != "" {
		_ = formatter.Writef("    Command:   %s\n", commandPath)
	}
	_ = formatter.Writef("    Status:    %s\n\n", statusStr)

	// Fallback backends
	fallbackNames := buildFallbackNames(discovery)
	if len(fallbackNames) > 0 {
		_ = formatter.Writef("  Fallback Backends\n")
		for _, name := range fallbackNames {
			avail := sliceContains(discovery.Available, name)
			icon := "✗"
			if avail {
				icon = "✓"
			}
			_ = formatter.Writef("    %s %s\n", icon, name)
		}
		_ = formatter.Writef("\n")
	}

	// Service readiness
	_ = formatter.Writef("  Services\n")
	for _, svc := range knownServices() {
		icon := "✗"
		if svcReady {
			icon = "✓"
		}
		_ = formatter.Writef("    %s %s\n", icon, svc)
	}

	return nil
}

// classifyBackendType returns "CLI" or "HTTP" based on the backend name.
func classifyBackendType(name string) string {
	switch name {
	case "claude-cli", "gemini-cli", "ollama-cli", "custom":
		return "CLI"
	case "claude", "ollama":
		return "HTTP"
	default:
		return "unknown"
	}
}

// resolveCommandPath returns the full path for CLI backends.
func resolveCommandPath(name string, lookPath llm.LookPathFunc) string {
	var cmd string
	switch name {
	case "claude-cli":
		cmd = "claude"
	case "gemini-cli":
		cmd = "gemini"
	case "ollama-cli":
		cmd = "ollama"
	default:
		return ""
	}
	path, err := lookPath(cmd)
	if err != nil {
		return ""
	}
	return path
}

// buildFallbacksJSON returns fallback entries excluding the selected backend.
func buildFallbacksJSON(discovery *llm.DiscoveryResult) []llmFallbackJSON {
	names := buildFallbackNames(discovery)
	fallbacks := make([]llmFallbackJSON, 0, len(names))
	for _, name := range names {
		fallbacks = append(fallbacks, llmFallbackJSON{
			Name:      name,
			Available: sliceContains(discovery.Available, name),
		})
	}
	return fallbacks
}

// buildFallbackNames returns all known backend names except the selected one.
func buildFallbackNames(discovery *llm.DiscoveryResult) []string {
	all := make([]string, 0, len(discovery.Available)+len(discovery.Unavailable))
	all = append(all, discovery.Available...)
	all = append(all, discovery.Unavailable...)

	var fallbacks []string
	for _, name := range all {
		if name != discovery.Selected {
			fallbacks = append(fallbacks, name)
		}
	}
	return fallbacks
}

// buildServicesJSON returns service readiness entries.
func buildServicesJSON(ready bool) []llmServiceJSON {
	svcs := knownServices()
	result := make([]llmServiceJSON, 0, len(svcs))
	for _, name := range svcs {
		result = append(result, llmServiceJSON{
			Name:  name,
			Ready: ready,
		})
	}
	return result
}

// knownServices returns the list of LLM-powered service names.
func knownServices() []string {
	return []string{"decompose", "enrich", "breakdown"}
}

func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// loadLLMConfig loads the LLM config from the standard config path.
// loadLLMConfig is defined in extract.go — shared across CLI commands in this package.
