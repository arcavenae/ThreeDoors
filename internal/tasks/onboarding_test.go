package tasks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsFirstRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configYAML string
		noFile     bool
		want       bool
	}{
		{
			name:   "no config directory",
			noFile: true,
			want:   true,
		},
		{
			name:       "empty config file",
			configYAML: "",
			want:       true,
		},
		{
			name:       "config without onboarding field",
			configYAML: "provider: textfile\n",
			want:       true,
		},
		{
			name:       "onboarding_complete false",
			configYAML: "onboarding_complete: false\n",
			want:       true,
		},
		{
			name:       "onboarding_complete true",
			configYAML: "onboarding_complete: true\n",
			want:       false,
		},
		{
			name:       "onboarding complete with other fields",
			configYAML: "provider: textfile\nonboarding_complete: true\nvalues:\n  - focus\n",
			want:       false,
		},
		{
			name:       "invalid yaml",
			configYAML: ":::invalid",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if !tt.noFile {
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(tt.configYAML), 0o644); err != nil {
					t.Fatalf("write config: %v", err)
				}
			}

			got := IsFirstRun(dir)
			if got != tt.want {
				t.Errorf("IsFirstRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkOnboardingComplete(t *testing.T) {
	t.Parallel()

	t.Run("creates config when none exists", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := MarkOnboardingComplete(dir); err != nil {
			t.Fatalf("MarkOnboardingComplete() error: %v", err)
		}

		if IsFirstRun(dir) {
			t.Error("expected IsFirstRun() = false after marking complete")
		}
	})

	t.Run("preserves existing config fields", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		existing := "provider: applenotes\nnote_title: My Tasks\nvalues:\n    - focus\n"
		if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(existing), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		if err := MarkOnboardingComplete(dir); err != nil {
			t.Fatalf("MarkOnboardingComplete() error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
		if err != nil {
			t.Fatalf("read config: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "onboarding_complete: true") {
			t.Errorf("config missing onboarding_complete: true, got:\n%s", content)
		}
		if !strings.Contains(content, "provider: applenotes") {
			t.Errorf("config lost provider field, got:\n%s", content)
		}
		if !strings.Contains(content, "note_title: My Tasks") {
			t.Errorf("config lost note_title field, got:\n%s", content)
		}
	})

	t.Run("idempotent when already complete", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("onboarding_complete: true\n"), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		if err := MarkOnboardingComplete(dir); err != nil {
			t.Fatalf("MarkOnboardingComplete() error: %v", err)
		}

		if IsFirstRun(dir) {
			t.Error("expected IsFirstRun() = false")
		}
	})
}
