package llm

import (
	"context"
	"fmt"
	"time"
)

// LLMBackend abstracts LLM provider communication for task decomposition.
type LLMBackend interface {
	// Name returns the provider identifier (e.g., "ollama", "claude").
	Name() string

	// Complete sends a prompt to the LLM and returns the response text.
	Complete(ctx context.Context, prompt string) (string, error)

	// Available reports whether the backend is reachable and ready.
	Available(ctx context.Context) bool
}

// Config holds LLM backend configuration loaded from config.yaml.
type Config struct {
	Backend   string       `yaml:"backend"`
	Ollama    OllamaConfig `yaml:"ollama"`
	Claude    ClaudeConfig `yaml:"claude"`
	ClaudeCLI CLIConfig    `yaml:"claude_cli"`
	GeminiCLI CLIConfig    `yaml:"gemini_cli"`
	OllamaCLI CLIConfig    `yaml:"ollama_cli"`
	Custom    CLIConfig    `yaml:"custom"`
	Output    OutputConfig `yaml:"decomposition"`
}

// CLIConfig holds settings for a CLI-based LLM backend.
type CLIConfig struct {
	Provider string        `yaml:"provider"`
	Command  string        `yaml:"command"`
	Args     []string      `yaml:"args"`
	Model    string        `yaml:"model"`
	Timeout  time.Duration `yaml:"timeout"`
}

// OllamaConfig holds Ollama-specific settings.
type OllamaConfig struct {
	Endpoint string `yaml:"endpoint"`
	Model    string `yaml:"model"`
}

// ClaudeConfig holds Anthropic Claude API settings.
type ClaudeConfig struct {
	Model  string `yaml:"model"`
	APIKey string `yaml:"-"` // loaded from ANTHROPIC_API_KEY env var, never serialized
}

// OutputConfig holds settings for writing decomposed stories to git repos.
type OutputConfig struct {
	OutputRepo         string `yaml:"output_repo"`
	OutputBranchPrefix string `yaml:"output_branch_prefix"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Backend: "ollama",
		Ollama: OllamaConfig{
			Endpoint: "http://localhost:11434",
			Model:    "llama3.2",
		},
		Claude: ClaudeConfig{
			Model: "claude-sonnet-4-20250514",
		},
		Output: OutputConfig{
			OutputBranchPrefix: "story/",
		},
	}
}

// NewCLIBackend creates a CLIProvider from a CLIConfig.
// Supported providers: "claude-cli", "gemini-cli", "ollama-cli", "custom".
func NewCLIBackend(cfg CLIConfig) (*CLIProvider, error) {
	var spec CLISpec
	switch cfg.Provider {
	case "claude-cli":
		spec = ClaudeCLISpec()
	case "gemini-cli":
		spec = GeminiCLISpec()
	case "ollama-cli":
		model := cfg.Model
		if model == "" {
			model = "llama3.2"
		}
		spec = OllamaCLISpec(model)
	case "custom":
		if cfg.Command == "" {
			return nil, fmt.Errorf("custom CLI backend requires a command: %w", ErrBackendUnavailable)
		}
		spec = CustomCLISpec(cfg.Command, cfg.Args)
	default:
		return nil, fmt.Errorf("unknown CLI provider %q: %w", cfg.Provider, ErrBackendUnavailable)
	}

	if cfg.Command != "" {
		spec.Command = cfg.Command
	}
	if cfg.Timeout > 0 {
		spec.Timeout = cfg.Timeout
	}

	return NewCLIProvider(spec, &ExecRunner{}), nil
}

// ErrBackendUnavailable is returned when an LLM backend cannot be reached.
var ErrBackendUnavailable = fmt.Errorf("llm backend unavailable")

// ErrEmptyPrompt is returned when Complete is called with an empty prompt.
var ErrEmptyPrompt = fmt.Errorf("prompt must not be empty")

// ErrEmptyResponse is returned when the LLM returns an empty response.
var ErrEmptyResponse = fmt.Errorf("llm returned empty response")

// CompletionResult captures a response along with metadata for logging.
type CompletionResult struct {
	Text      string
	Backend   string
	Model     string
	Duration  time.Duration
	Timestamp time.Time
}
