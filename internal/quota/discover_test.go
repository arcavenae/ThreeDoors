package quota

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSessionFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantCount int
		wantErr   bool
	}{
		{
			name: "finds jsonl files under projects",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				projDir := filepath.Join(dir, "projects", "proj-hash")
				if err := os.MkdirAll(projDir, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"a.jsonl", "b.jsonl", "readme.txt"} {
					if err := os.WriteFile(filepath.Join(projDir, name), []byte("{}"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			wantCount: 2,
		},
		{
			name: "no projects directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantCount: 0,
		},
		{
			name: "empty projects directory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.MkdirAll(filepath.Join(dir, "projects"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantCount: 0,
		},
		{
			name: "nested project directories",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				for _, sub := range []string{"proj-a", "proj-b"} {
					subDir := filepath.Join(dir, "projects", sub)
					if err := os.MkdirAll(subDir, 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(subDir, "session.jsonl"), []byte("{}"), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := tt.setup(t)
			files, err := DiscoverSessionFiles(dir)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(files) != tt.wantCount {
				t.Errorf("got %d files, want %d", len(files), tt.wantCount)
			}
		})
	}
}

func TestDefaultConfigDir(t *testing.T) {
	// Cannot use t.Parallel — t.Setenv mutates process environment.
	t.Run("respects CLAUDE_CONFIG_DIR", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "/custom/path")
		got := DefaultConfigDir()
		if got != "/custom/path" {
			t.Errorf("got %q, want /custom/path", got)
		}
	})

	t.Run("falls back to home dir", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "")
		got := DefaultConfigDir()
		if got == "" {
			t.Error("expected non-empty default config dir")
		}
	})
}
