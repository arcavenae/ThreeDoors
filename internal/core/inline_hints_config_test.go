package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultInlineHintsConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultInlineHintsConfig()
	if !cfg.ShowInlineHints {
		t.Error("expected ShowInlineHints to default to true")
	}
	if cfg.SessionCount != 0 {
		t.Errorf("expected SessionCount 0, got %d", cfg.SessionCount)
	}
	if cfg.FadeThreshold != 5 {
		t.Errorf("expected FadeThreshold 5, got %d", cfg.FadeThreshold)
	}
}

func TestLoadInlineHintsConfig_NoFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg, err := LoadInlineHintsConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.ShowInlineHints {
		t.Error("expected defaults when file missing")
	}
}

func TestInlineHintsConfigRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	original := &InlineHintsConfig{
		ShowInlineHints: false,
		SessionCount:    3,
		FadeThreshold:   10,
	}

	if err := SaveInlineHintsConfig(dir, original); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadInlineHintsConfig(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.ShowInlineHints != original.ShowInlineHints {
		t.Errorf("ShowInlineHints: got %v, want %v", loaded.ShowInlineHints, original.ShowInlineHints)
	}
	if loaded.SessionCount != original.SessionCount {
		t.Errorf("SessionCount: got %d, want %d", loaded.SessionCount, original.SessionCount)
	}
	if loaded.FadeThreshold != original.FadeThreshold {
		t.Errorf("FadeThreshold: got %d, want %d", loaded.FadeThreshold, original.FadeThreshold)
	}
}

func TestSaveInlineHintsConfig_PreservesExistingFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write a pre-existing config with a theme field
	if err := os.WriteFile(configPath, []byte("theme: classic\n"), 0o644); err != nil {
		t.Fatalf("write pre-existing config: %v", err)
	}

	cfg := &InlineHintsConfig{
		ShowInlineHints: true,
		SessionCount:    2,
		FadeThreshold:   5,
	}
	if err := SaveInlineHintsConfig(dir, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "theme: classic") {
		t.Error("expected theme field to be preserved")
	}
	if !strings.Contains(content, "show_inline_hints: true") {
		t.Error("expected show_inline_hints field")
	}
}

func TestSessionCounterAutoDisable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sessionCount int
		threshold    int
		wantEnabled  bool
		wantFade     bool
	}{
		{
			name:         "below threshold shows hints normally",
			sessionCount: 2,
			threshold:    5,
			wantEnabled:  true,
			wantFade:     false,
		},
		{
			name:         "one before threshold shows fade",
			sessionCount: 4,
			threshold:    5,
			wantEnabled:  true,
			wantFade:     true,
		},
		{
			name:         "at threshold disables hints",
			sessionCount: 5,
			threshold:    5,
			wantEnabled:  false,
			wantFade:     false,
		},
		{
			name:         "above threshold stays disabled",
			sessionCount: 10,
			threshold:    5,
			wantEnabled:  false,
			wantFade:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &InlineHintsConfig{
				ShowInlineHints: true,
				SessionCount:    tt.sessionCount,
				FadeThreshold:   tt.threshold,
			}

			enabled, fade := ResolveInlineHintState(cfg)
			if enabled != tt.wantEnabled {
				t.Errorf("enabled: got %v, want %v", enabled, tt.wantEnabled)
			}
			if fade != tt.wantFade {
				t.Errorf("fade: got %v, want %v", fade, tt.wantFade)
			}
		})
	}
}
