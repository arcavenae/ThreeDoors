package dispatch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDevDispatchConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultDevDispatchConfig()

	if cfg.Enabled {
		t.Error("Enabled should default to false")
	}
	if cfg.MaxConcurrent != 2 {
		t.Errorf("MaxConcurrent = %d, want 2", cfg.MaxConcurrent)
	}
	if cfg.AutoDispatch {
		t.Error("AutoDispatch should default to false")
	}
	if cfg.CooldownMinutes != 5 {
		t.Errorf("CooldownMinutes = %d, want 5", cfg.CooldownMinutes)
	}
	if cfg.DailyLimit != 10 {
		t.Errorf("DailyLimit = %d, want 10", cfg.DailyLimit)
	}
	if cfg.RequireStory {
		t.Error("RequireStory should default to false")
	}
}

func TestLoadDevDispatchConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    DevDispatchConfig
	}{
		{
			name:    "full config",
			content: "dev_dispatch:\n  enabled: true\n  max_concurrent: 4\n  auto_dispatch: true\n  cooldown_minutes: 10\n  daily_limit: 20\n  require_story: true\n",
			want: DevDispatchConfig{
				Enabled:         true,
				MaxConcurrent:   4,
				AutoDispatch:    true,
				CooldownMinutes: 10,
				DailyLimit:      20,
				RequireStory:    true,
			},
		},
		{
			name:    "partial config fills defaults",
			content: "dev_dispatch:\n  enabled: true\n",
			want: DevDispatchConfig{
				Enabled:         true,
				MaxConcurrent:   2,
				AutoDispatch:    false,
				CooldownMinutes: 5,
				DailyLimit:      10,
				RequireStory:    false,
			},
		},
		{
			name:    "no dev_dispatch section returns defaults",
			content: "provider: textfile\n",
			want:    DefaultDevDispatchConfig(),
		},
		{
			name:    "empty file returns defaults",
			content: "",
			want:    DefaultDevDispatchConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("write config: %v", err)
			}

			got, err := LoadDevDispatchConfig(path)
			if err != nil {
				t.Fatalf("LoadDevDispatchConfig: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadDevDispatchConfigFileNotFound(t *testing.T) {
	t.Parallel()
	cfg, err := LoadDevDispatchConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("should return defaults for missing file: %v", err)
	}
	if cfg != DefaultDevDispatchConfig() {
		t.Errorf("got %+v, want defaults", cfg)
	}
}

func TestLoadDevDispatchConfigInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadDevDispatchConfig(path)
	if err == nil {
		t.Fatal("should return error for invalid YAML")
	}
}

func TestApplyDefaults(t *testing.T) {
	t.Parallel()
	cfg := &DevDispatchConfig{}
	applyDefaults(cfg)

	if cfg.MaxConcurrent != 2 {
		t.Errorf("MaxConcurrent = %d, want 2", cfg.MaxConcurrent)
	}
	if cfg.CooldownMinutes != 5 {
		t.Errorf("CooldownMinutes = %d, want 5", cfg.CooldownMinutes)
	}
	if cfg.DailyLimit != 10 {
		t.Errorf("DailyLimit = %d, want 10", cfg.DailyLimit)
	}
}
