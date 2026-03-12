package llm

import (
	"testing"
)

func TestNewCLIBackend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      CLIConfig
		wantName string
		wantErr  bool
	}{
		{
			name:     "claude-cli provider",
			cfg:      CLIConfig{Provider: "claude-cli"},
			wantName: "claude-cli",
		},
		{
			name:     "gemini-cli provider",
			cfg:      CLIConfig{Provider: "gemini-cli"},
			wantName: "gemini-cli",
		},
		{
			name:     "ollama-cli provider",
			cfg:      CLIConfig{Provider: "ollama-cli", Model: "mistral"},
			wantName: "ollama-cli",
		},
		{
			name:     "ollama-cli default model",
			cfg:      CLIConfig{Provider: "ollama-cli"},
			wantName: "ollama-cli",
		},
		{
			name:     "custom provider",
			cfg:      CLIConfig{Provider: "custom", Command: "my-llm", Args: []string{"--fast"}},
			wantName: "custom",
		},
		{
			name:    "custom without command errors",
			cfg:     CLIConfig{Provider: "custom"},
			wantErr: true,
		},
		{
			name:    "unknown provider errors",
			cfg:     CLIConfig{Provider: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			backend, err := NewCLIBackend(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCLIBackend() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if backend.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", backend.Name(), tt.wantName)
			}
		})
	}
}

func TestNewCLIBackendCustomCommand(t *testing.T) {
	t.Parallel()

	// Command override should take effect
	backend, err := NewCLIBackend(CLIConfig{
		Provider: "claude-cli",
		Command:  "/usr/local/bin/claude",
	})
	if err != nil {
		t.Fatalf("NewCLIBackend() error = %v", err)
	}
	if backend.spec.Command != "/usr/local/bin/claude" {
		t.Errorf("Command = %q, want %q", backend.spec.Command, "/usr/local/bin/claude")
	}
}

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
