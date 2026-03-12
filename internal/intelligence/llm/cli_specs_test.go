package llm

import (
	"testing"
	"time"
)

func TestClaudeCLISpec(t *testing.T) {
	t.Parallel()

	spec := ClaudeCLISpec()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"name", spec.Name, "claude-cli"},
		{"command", spec.Command, "claude"},
		{"system prompt flag", spec.SystemPrompt.Flag, "--system-prompt"},
		{"output format flag", spec.OutputFormat.Flag, "--output-format"},
		{"output format value", spec.OutputFormat.Value, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}

	if len(spec.BaseArgs) != 1 || spec.BaseArgs[0] != "--print" {
		t.Errorf("BaseArgs = %v, want [--print]", spec.BaseArgs)
	}
	if spec.InputMethod != InputStdin {
		t.Errorf("InputMethod = %d, want InputStdin (%d)", spec.InputMethod, InputStdin)
	}
	if spec.Timeout != 120*time.Second {
		t.Errorf("Timeout = %v, want 120s", spec.Timeout)
	}
	if spec.SystemPrompt.Enabled {
		t.Error("SystemPrompt should be disabled by default")
	}
}

func TestGeminiCLISpec(t *testing.T) {
	t.Parallel()

	spec := GeminiCLISpec()

	if spec.Name != "gemini-cli" {
		t.Errorf("Name = %q, want %q", spec.Name, "gemini-cli")
	}
	if spec.Command != "gemini" {
		t.Errorf("Command = %q, want %q", spec.Command, "gemini")
	}
	if spec.InputMethod != InputStdin {
		t.Errorf("InputMethod = %d, want InputStdin", spec.InputMethod)
	}
	if spec.OutputFormat.Flag != "--output-format" {
		t.Errorf("OutputFormat.Flag = %q, want %q", spec.OutputFormat.Flag, "--output-format")
	}
	if spec.OutputFormat.Value != "json" {
		t.Errorf("OutputFormat.Value = %q, want %q", spec.OutputFormat.Value, "json")
	}
	if len(spec.BaseArgs) != 0 {
		t.Errorf("BaseArgs = %v, want empty", spec.BaseArgs)
	}
}

func TestOllamaCLISpec(t *testing.T) {
	t.Parallel()

	spec := OllamaCLISpec("mistral")

	if spec.Name != "ollama-cli" {
		t.Errorf("Name = %q, want %q", spec.Name, "ollama-cli")
	}
	if spec.Command != "ollama" {
		t.Errorf("Command = %q, want %q", spec.Command, "ollama")
	}
	if len(spec.BaseArgs) != 2 || spec.BaseArgs[0] != "run" || spec.BaseArgs[1] != "mistral" {
		t.Errorf("BaseArgs = %v, want [run mistral]", spec.BaseArgs)
	}
	if spec.InputMethod != InputArg {
		t.Errorf("InputMethod = %d, want InputArg (%d)", spec.InputMethod, InputArg)
	}
	if spec.SystemPrompt.Flag != "--system" {
		t.Errorf("SystemPrompt.Flag = %q, want %q", spec.SystemPrompt.Flag, "--system")
	}
}

func TestCustomCLISpec(t *testing.T) {
	t.Parallel()

	spec := CustomCLISpec("my-llm", []string{"--fast", "--model", "v2"})

	if spec.Name != "custom" {
		t.Errorf("Name = %q, want %q", spec.Name, "custom")
	}
	if spec.Command != "my-llm" {
		t.Errorf("Command = %q, want %q", spec.Command, "my-llm")
	}
	if len(spec.BaseArgs) != 3 {
		t.Fatalf("BaseArgs len = %d, want 3", len(spec.BaseArgs))
	}
	if spec.BaseArgs[0] != "--fast" || spec.BaseArgs[1] != "--model" || spec.BaseArgs[2] != "v2" {
		t.Errorf("BaseArgs = %v, want [--fast --model v2]", spec.BaseArgs)
	}
	if spec.InputMethod != InputStdin {
		t.Errorf("InputMethod = %d, want InputStdin", spec.InputMethod)
	}
}
