package llm

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"backend", cfg.Backend, "ollama"},
		{"ollama endpoint", cfg.Ollama.Endpoint, "http://localhost:11434"},
		{"ollama model", cfg.Ollama.Model, "llama3.2"},
		{"claude model", cfg.Claude.Model, "claude-sonnet-4-20250514"},
		{"branch prefix", cfg.Output.OutputBranchPrefix, "story/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Parallel()

	if ErrBackendUnavailable == nil {
		t.Error("ErrBackendUnavailable should not be nil")
	}
	if ErrEmptyPrompt == nil {
		t.Error("ErrEmptyPrompt should not be nil")
	}
	if ErrEmptyResponse == nil {
		t.Error("ErrEmptyResponse should not be nil")
	}
}
