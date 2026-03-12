package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
	tea "github.com/charmbracelet/bubbletea"
)

// runLLMStatusCmd returns a tea.Cmd that runs LLM backend discovery
// and produces a LLMStatusResultMsg with a formatted summary.
func runLLMStatusCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := loadTUILLMConfig()
		if err != nil {
			return LLMStatusResultMsg{Text: fmt.Sprintf("LLM status: config error: %v", err)}
		}

		ctx := context.Background()
		_, discovery, _ := llm.DiscoverBackend(ctx, cfg)
		if discovery == nil {
			discovery = &llm.DiscoveryResult{}
		}

		return LLMStatusResultMsg{Text: formatLLMStatus(discovery)}
	}
}

// formatLLMStatus builds a concise human-readable status string for flash display.
func formatLLMStatus(d *llm.DiscoveryResult) string {
	if d.Selected == "" {
		return "LLM: No backends available. Install claude, gemini, or ollama."
	}

	var s strings.Builder
	fmt.Fprintf(&s, "LLM: %s", d.Selected)

	// Show fallback count.
	fallbackCount := len(d.Available) - 1
	if fallbackCount < 0 {
		fallbackCount = 0
	}
	if fallbackCount > 0 {
		fmt.Fprintf(&s, " | %d fallback(s)", fallbackCount)
	}

	// Show unavailable count.
	if len(d.Unavailable) > 0 {
		fmt.Fprintf(&s, " | %d unavailable", len(d.Unavailable))
	}

	return s.String()
}

// loadTUILLMConfig loads the LLM config from the standard config path.
func loadTUILLMConfig() (llm.Config, error) {
	configDir, err := core.GetConfigDirPath()
	if err != nil {
		return llm.Config{}, fmt.Errorf("config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	cfg, err := core.LoadProviderConfig(configPath)
	if err != nil {
		return llm.Config{}, fmt.Errorf("load config: %w", err)
	}

	return cfg.LLM, nil
}
