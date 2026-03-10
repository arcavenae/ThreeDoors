package core

import (
	"testing"
)

func TestDefaultInlineHintsConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultInlineHintsConfig()
	if !cfg.ShowInlineHints {
		t.Error("expected ShowInlineHints to default to true")
	}
}

func TestResolveInlineHintState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		show        bool
		wantEnabled bool
		wantFade    bool
	}{
		{
			name:        "enabled returns true",
			show:        true,
			wantEnabled: true,
			wantFade:    false,
		},
		{
			name:        "disabled returns false",
			show:        false,
			wantEnabled: false,
			wantFade:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &InlineHintsConfig{ShowInlineHints: tt.show}
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

func TestShowKeyHintsMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		showKeybindingBar *bool
		showKeyHints      *bool
		wantShowKeyHints  *bool
	}{
		{
			name:              "migrates show_keybinding_bar true to show_key_hints true",
			showKeybindingBar: boolPtr(true),
			showKeyHints:      nil,
			wantShowKeyHints:  boolPtr(true),
		},
		{
			name:              "migrates show_keybinding_bar false to show_key_hints false",
			showKeybindingBar: boolPtr(false),
			showKeyHints:      nil,
			wantShowKeyHints:  boolPtr(false),
		},
		{
			name:              "does not overwrite existing show_key_hints",
			showKeybindingBar: boolPtr(false),
			showKeyHints:      boolPtr(true),
			wantShowKeyHints:  boolPtr(true),
		},
		{
			name:              "both nil stays nil",
			showKeybindingBar: nil,
			showKeyHints:      nil,
			wantShowKeyHints:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &ProviderConfig{
				ShowKeybindingBar: tt.showKeybindingBar,
				ShowKeyHints:      tt.showKeyHints,
			}
			MigrateConfig(cfg)
			if tt.wantShowKeyHints == nil {
				if cfg.ShowKeyHints != nil {
					t.Errorf("expected ShowKeyHints nil, got %v", *cfg.ShowKeyHints)
				}
			} else {
				if cfg.ShowKeyHints == nil {
					t.Fatal("expected ShowKeyHints non-nil, got nil")
					return
				}
				if *cfg.ShowKeyHints != *tt.wantShowKeyHints {
					t.Errorf("ShowKeyHints: got %v, want %v", *cfg.ShowKeyHints, *tt.wantShowKeyHints)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestShowKeyHintsPersistenceRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.yaml"

	// Save config with show_key_hints
	show := true
	cfg := &ProviderConfig{
		ShowKeyHints: &show,
	}
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load it back
	loaded, err := LoadProviderConfig(configPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
		return
	}

	if loaded.ShowKeyHints == nil {
		t.Fatal("expected ShowKeyHints non-nil after load")
		return
	}
	if *loaded.ShowKeyHints != true {
		t.Errorf("ShowKeyHints: got %v, want true", *loaded.ShowKeyHints)
	}
}
