package quota

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		yaml       string
		wantTiers  int
		wantCLI    bool
		wantPeakSF float64
		wantErr    bool
	}{
		{
			name:       "empty config uses defaults",
			yaml:       "",
			wantTiers:  4,
			wantCLI:    false,
			wantPeakSF: 0.8,
		},
		{
			name: "custom tiers",
			yaml: `
tiers:
  - percent: 60
    label: "info"
    suggestion: "Just a heads up"
  - percent: 85
    label: "warn"
    suggestion: "Getting high"
`,
			wantTiers:  2,
			wantPeakSF: 0.8,
		},
		{
			name: "cli notify mode",
			yaml: `
notify: cli
`,
			wantTiers: 4,
			wantCLI:   true,
		},
		{
			name: "custom peak config",
			yaml: `
peak:
  start_hour: 6
  end_hour: 12
  shift_factor: 0.75
`,
			wantTiers:  4,
			wantPeakSF: 0.75,
		},
		{
			name:    "invalid yaml",
			yaml:    "{{{{",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := ParseConfig([]byte(tt.yaml))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(cfg.Tiers) != tt.wantTiers {
				t.Errorf("tiers = %d, want %d", len(cfg.Tiers), tt.wantTiers)
			}
			if cfg.NotifyViaCLI != tt.wantCLI {
				t.Errorf("NotifyViaCLI = %v, want %v", cfg.NotifyViaCLI, tt.wantCLI)
			}
			if tt.wantPeakSF > 0 {
				if diff := cfg.PeakShiftFactor - tt.wantPeakSF; diff > 0.001 || diff < -0.001 {
					t.Errorf("PeakShiftFactor = %f, want %f", cfg.PeakShiftFactor, tt.wantPeakSF)
				}
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "quota.yaml")
		content := []byte(`
notify: cli
peak:
  shift_factor: 0.9
`)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("write test file: %v", err)
		}

		cfg, err := LoadConfigFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.NotifyViaCLI {
			t.Error("expected CLI mode")
		}
		if diff := cfg.PeakShiftFactor - 0.9; diff > 0.001 || diff < -0.001 {
			t.Errorf("PeakShiftFactor = %f, want 0.9", cfg.PeakShiftFactor)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		_, err := LoadConfigFromFile("/nonexistent/quota.yaml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Not parallel — modifies environment

	t.Run("default when no env set", func(t *testing.T) {
		cfg := LoadConfigFromEnv()
		if cfg.NotifyViaCLI {
			t.Error("expected multiclaude mode by default")
		}
		if cfg.PeakShiftFactor != 0.8 {
			t.Errorf("PeakShiftFactor = %f, want 0.8", cfg.PeakShiftFactor)
		}
	})

	t.Run("cli notify from env", func(t *testing.T) {
		t.Setenv("QUOTA_NOTIFY", "cli")
		cfg := LoadConfigFromEnv()
		if !cfg.NotifyViaCLI {
			t.Error("expected CLI mode from env")
		}
	})

	t.Run("peak shift from env", func(t *testing.T) {
		t.Setenv("QUOTA_PEAK_SHIFT", "0.7")
		cfg := LoadConfigFromEnv()
		if diff := cfg.PeakShiftFactor - 0.7; diff > 0.001 || diff < -0.001 {
			t.Errorf("PeakShiftFactor = %f, want 0.7", cfg.PeakShiftFactor)
		}
	})

	t.Run("invalid peak shift ignored", func(t *testing.T) {
		t.Setenv("QUOTA_PEAK_SHIFT", "notanumber")
		cfg := LoadConfigFromEnv()
		if cfg.PeakShiftFactor != 0.8 {
			t.Errorf("PeakShiftFactor = %f, want 0.8 (default)", cfg.PeakShiftFactor)
		}
	})

	t.Run("out of range peak shift ignored", func(t *testing.T) {
		t.Setenv("QUOTA_PEAK_SHIFT", "1.5")
		cfg := LoadConfigFromEnv()
		if cfg.PeakShiftFactor != 0.8 {
			t.Errorf("PeakShiftFactor = %f, want 0.8 (default)", cfg.PeakShiftFactor)
		}
	})
}
