package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

func TestFormatLLMStatus_NoBackend(t *testing.T) {
	t.Parallel()

	d := &llm.DiscoveryResult{}
	got := formatLLMStatus(d)

	if !strings.Contains(got, "No backends available") {
		t.Errorf("expected no-backend message, got: %q", got)
	}
}

func TestFormatLLMStatus_WithBackend(t *testing.T) {
	t.Parallel()

	d := &llm.DiscoveryResult{
		Selected:    "claude-cli",
		Available:   []string{"claude-cli", "gemini-cli"},
		Unavailable: []string{"ollama-cli"},
	}
	got := formatLLMStatus(d)

	if !strings.Contains(got, "claude-cli") {
		t.Errorf("expected backend name in output, got: %q", got)
	}
	if !strings.Contains(got, "1 fallback") {
		t.Errorf("expected fallback count, got: %q", got)
	}
	if !strings.Contains(got, "1 unavailable") {
		t.Errorf("expected unavailable count, got: %q", got)
	}
}

func TestFormatLLMStatus_SingleBackendNoFallbacks(t *testing.T) {
	t.Parallel()

	d := &llm.DiscoveryResult{
		Selected:    "ollama",
		Available:   []string{"ollama"},
		Unavailable: []string{"claude-cli", "gemini-cli", "ollama-cli"},
	}
	got := formatLLMStatus(d)

	if !strings.Contains(got, "ollama") {
		t.Errorf("expected backend name, got: %q", got)
	}
	// With only 1 available (the selected one), fallback count should be 0 → not shown.
	if strings.Contains(got, "fallback") {
		t.Errorf("should not show fallback count when zero, got: %q", got)
	}
}

func TestCommandRegistryContainsLLMStatus(t *testing.T) {
	t.Parallel()

	found := false
	for _, cmd := range commandRegistry {
		if cmd.Name == "llm-status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'llm-status' in commandRegistry")
	}
}
