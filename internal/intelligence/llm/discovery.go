package llm

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// DiscoveryResult captures which backends were found during auto-discovery.
type DiscoveryResult struct {
	Selected    string   `json:"selected"`
	Available   []string `json:"available"`
	Unavailable []string `json:"unavailable"`
}

// LookPathFunc abstracts exec.LookPath for testing.
type LookPathFunc func(file string) (string, error)

// cliCandidate describes a CLI tool to probe during discovery.
type cliCandidate struct {
	name    string // provider name (e.g., "claude-cli")
	command string // binary name to look up
}

// cliPriority defines the discovery order per S2-D5:
// claude CLI → gemini CLI → ollama CLI.
var cliPriority = []cliCandidate{
	{"claude-cli", "claude"},
	{"gemini-cli", "gemini"},
	{"ollama-cli", "ollama"},
}

// DiscoverBackend selects the best available LLM backend.
//
// Priority order (S2-D5):
//  1. User-configured backend (explicit in config.yaml)
//  2. Claude CLI
//  3. Gemini CLI
//  4. Ollama CLI
//  5. Ollama HTTP API (if reachable)
//  6. Claude HTTP API (if ANTHROPIC_API_KEY is set)
//  7. ErrBackendUnavailable (graceful degradation)
func DiscoverBackend(ctx context.Context, cfg Config) (LLMBackend, *DiscoveryResult, error) {
	return discoverBackendWith(ctx, cfg, exec.LookPath, os.Getenv)
}

// discoverBackendWith is the testable core of DiscoverBackend.
func discoverBackendWith(ctx context.Context, cfg Config, lookPath LookPathFunc, getenv func(string) string) (LLMBackend, *DiscoveryResult, error) {
	result := &DiscoveryResult{}

	// Priority 1: User-configured backend takes precedence.
	if cfg.Backend != "" {
		backend, err := newExplicitBackend(cfg, getenv)
		if err != nil {
			return nil, result, fmt.Errorf("discover backend: configured %q: %w", cfg.Backend, err)
		}
		result.Selected = backend.Name()
		slog.Info("using configured LLM backend", "backend", result.Selected)
		return backend, result, nil
	}

	// Priority 2-4: CLI tools in priority order.
	for _, c := range cliPriority {
		_, err := lookPath(c.command)
		if err == nil {
			result.Available = append(result.Available, c.name)
		} else {
			result.Unavailable = append(result.Unavailable, c.name)
		}
	}

	// Select first available CLI.
	if len(result.Available) > 0 {
		selected := result.Available[0]
		backend, err := newCLIBackendByName(selected, cfg)
		if err != nil {
			return nil, result, fmt.Errorf("discover backend: create %s: %w", selected, err)
		}
		result.Selected = selected
		logDiscovery(result)
		return backend, result, nil
	}

	// Priority 5: Ollama HTTP API.
	ollamaHTTP := NewOllamaBackend(cfg.Ollama)
	if ollamaHTTP.Available(ctx) {
		result.Selected = "ollama"
		logDiscovery(result)
		return ollamaHTTP, result, nil
	}
	result.Unavailable = append(result.Unavailable, "ollama")

	// Priority 6: Claude HTTP API.
	apiKey := getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		claudeCfg := cfg.Claude
		claudeCfg.APIKey = apiKey
		claudeHTTP := NewClaudeBackend(claudeCfg)
		result.Selected = "claude"
		logDiscovery(result)
		return claudeHTTP, result, nil
	}
	result.Unavailable = append(result.Unavailable, "claude")

	// Priority 7: Nothing available.
	logDiscovery(result)
	return nil, result, ErrBackendUnavailable
}

// DiscoverAvailableCLIs returns the names of all CLI tools found on PATH.
func DiscoverAvailableCLIs() []string {
	return discoverAvailableCLIsWith(exec.LookPath)
}

// discoverAvailableCLIsWith is the testable core of DiscoverAvailableCLIs.
func discoverAvailableCLIsWith(lookPath LookPathFunc) []string {
	var available []string
	for _, c := range cliPriority {
		if _, err := lookPath(c.command); err == nil {
			available = append(available, c.name)
		}
	}
	return available
}

// newExplicitBackend creates a backend from an explicit Backend config value.
func newExplicitBackend(cfg Config, getenv func(string) string) (LLMBackend, error) {
	switch cfg.Backend {
	case "ollama":
		return NewOllamaBackend(cfg.Ollama), nil
	case "claude":
		claudeCfg := cfg.Claude
		if claudeCfg.APIKey == "" {
			claudeCfg.APIKey = getenv("ANTHROPIC_API_KEY")
		}
		if claudeCfg.APIKey == "" {
			return nil, fmt.Errorf("claude backend requires ANTHROPIC_API_KEY: %w", ErrBackendUnavailable)
		}
		return NewClaudeBackend(claudeCfg), nil
	case "claude-cli":
		return NewCLIBackend(CLIConfig{Provider: "claude-cli", Command: cfg.ClaudeCLI.Command, Timeout: cfg.ClaudeCLI.Timeout})
	case "gemini-cli":
		return NewCLIBackend(CLIConfig{Provider: "gemini-cli", Command: cfg.GeminiCLI.Command, Timeout: cfg.GeminiCLI.Timeout})
	case "ollama-cli":
		return NewCLIBackend(CLIConfig{Provider: "ollama-cli", Command: cfg.OllamaCLI.Command, Model: cfg.OllamaCLI.Model, Timeout: cfg.OllamaCLI.Timeout})
	case "custom":
		return NewCLIBackend(CLIConfig{Provider: "custom", Command: cfg.Custom.Command, Args: cfg.Custom.Args, Timeout: cfg.Custom.Timeout})
	default:
		return nil, fmt.Errorf("unknown backend %q: %w", cfg.Backend, ErrBackendUnavailable)
	}
}

// newCLIBackendByName creates a CLI backend by provider name using config defaults.
func newCLIBackendByName(name string, cfg Config) (LLMBackend, error) {
	switch name {
	case "claude-cli":
		return NewCLIBackend(CLIConfig{Provider: "claude-cli", Command: cfg.ClaudeCLI.Command, Timeout: cfg.ClaudeCLI.Timeout})
	case "gemini-cli":
		return NewCLIBackend(CLIConfig{Provider: "gemini-cli", Command: cfg.GeminiCLI.Command, Timeout: cfg.GeminiCLI.Timeout})
	case "ollama-cli":
		return NewCLIBackend(CLIConfig{Provider: "ollama-cli", Command: cfg.OllamaCLI.Command, Model: cfg.OllamaCLI.Model, Timeout: cfg.OllamaCLI.Timeout})
	default:
		return nil, fmt.Errorf("unknown CLI provider %q: %w", name, ErrBackendUnavailable)
	}
}

// logDiscovery logs the discovery result at appropriate levels.
func logDiscovery(r *DiscoveryResult) {
	slog.Info("LLM backend discovery complete",
		"selected", r.Selected,
		"available", r.Available,
	)
	for _, name := range r.Unavailable {
		slog.Debug("LLM backend unavailable", "backend", name)
	}
}
