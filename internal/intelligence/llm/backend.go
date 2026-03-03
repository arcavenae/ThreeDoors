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
	Backend string       `yaml:"backend"`
	Ollama  OllamaConfig `yaml:"ollama"`
	Claude  ClaudeConfig `yaml:"claude"`
	Output  OutputConfig `yaml:"decomposition"`
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
