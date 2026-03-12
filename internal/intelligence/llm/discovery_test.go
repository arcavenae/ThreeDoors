package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockLookPath creates a LookPathFunc that reports the given commands as available.
func mockLookPath(available ...string) LookPathFunc {
	set := make(map[string]bool, len(available))
	for _, cmd := range available {
		set[cmd] = true
	}
	return func(file string) (string, error) {
		if set[file] {
			return "/usr/bin/" + file, nil
		}
		return "", fmt.Errorf("executable file not found in $PATH")
	}
}

// mockGetenv creates a getenv func from a map.
func mockGetenv(vars map[string]string) func(string) string {
	return func(key string) string {
		return vars[key]
	}
}

func TestDiscoverBackend_UserConfigured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		backend     string
		env         map[string]string
		wantName    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "ollama configured explicitly",
			backend:  "ollama",
			wantName: "ollama",
		},
		{
			name:     "claude configured with API key",
			backend:  "claude",
			env:      map[string]string{"ANTHROPIC_API_KEY": "sk-test-key"},
			wantName: "claude",
		},
		{
			name:        "claude configured without API key",
			backend:     "claude",
			env:         map[string]string{},
			wantErr:     true,
			errContains: "ANTHROPIC_API_KEY",
		},
		{
			name:     "claude-cli configured",
			backend:  "claude-cli",
			wantName: "claude-cli",
		},
		{
			name:     "gemini-cli configured",
			backend:  "gemini-cli",
			wantName: "gemini-cli",
		},
		{
			name:     "ollama-cli configured",
			backend:  "ollama-cli",
			wantName: "ollama-cli",
		},
		{
			name:        "unknown backend",
			backend:     "not-a-thing",
			wantErr:     true,
			errContains: "unknown backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			cfg := DefaultConfig()
			cfg.Backend = tt.backend

			env := tt.env
			if env == nil {
				env = map[string]string{}
			}

			backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath(), mockGetenv(env))

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" {
					if got := err.Error(); !contains(got, tt.errContains) {
						t.Errorf("error %q does not contain %q", got, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if backend.Name() != tt.wantName {
				t.Errorf("got backend name %q, want %q", backend.Name(), tt.wantName)
			}
			if result.Selected != tt.wantName {
				t.Errorf("got result.Selected %q, want %q", result.Selected, tt.wantName)
			}
		})
	}
}

func TestDiscoverBackend_CLIPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		availableCmds   []string
		wantSelected    string
		wantAvailable   []string
		wantUnavailable []string
	}{
		{
			name:            "all CLIs available, selects claude",
			availableCmds:   []string{"claude", "gemini", "ollama"},
			wantSelected:    "claude-cli",
			wantAvailable:   []string{"claude-cli", "gemini-cli", "ollama-cli"},
			wantUnavailable: nil,
		},
		{
			name:            "only gemini and ollama, selects gemini",
			availableCmds:   []string{"gemini", "ollama"},
			wantSelected:    "gemini-cli",
			wantAvailable:   []string{"gemini-cli", "ollama-cli"},
			wantUnavailable: []string{"claude-cli"},
		},
		{
			name:            "only ollama CLI, selects ollama-cli",
			availableCmds:   []string{"ollama"},
			wantSelected:    "ollama-cli",
			wantAvailable:   []string{"ollama-cli"},
			wantUnavailable: []string{"claude-cli", "gemini-cli"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			cfg := DefaultConfig()
			cfg.Backend = "" // trigger auto-discovery

			backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath(tt.availableCmds...), mockGetenv(nil))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if backend.Name() != tt.wantSelected {
				t.Errorf("got backend %q, want %q", backend.Name(), tt.wantSelected)
			}
			if result.Selected != tt.wantSelected {
				t.Errorf("got result.Selected %q, want %q", result.Selected, tt.wantSelected)
			}
			if !stringSliceEqual(result.Available, tt.wantAvailable) {
				t.Errorf("got Available %v, want %v", result.Available, tt.wantAvailable)
			}
			if len(tt.wantUnavailable) > 0 && !stringSliceEqual(result.Unavailable[:len(tt.wantUnavailable)], tt.wantUnavailable) {
				t.Errorf("got Unavailable %v, want prefix %v", result.Unavailable, tt.wantUnavailable)
			}
		})
	}
}

func TestDiscoverBackend_HTTPFallback_OllamaReachable(t *testing.T) {
	t.Parallel()

	// Mock Ollama HTTP server that responds to /api/tags.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Backend = ""
	cfg.Ollama.Endpoint = srv.URL

	backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath(), mockGetenv(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend.Name() != "ollama" {
		t.Errorf("got backend %q, want %q", backend.Name(), "ollama")
	}
	if result.Selected != "ollama" {
		t.Errorf("got result.Selected %q, want %q", result.Selected, "ollama")
	}
}

func TestDiscoverBackend_HTTPFallback_ClaudeAPIKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Backend = ""
	// Make Ollama HTTP unreachable by pointing to a closed server.
	cfg.Ollama.Endpoint = "http://127.0.0.1:1" // should fail to connect

	env := map[string]string{"ANTHROPIC_API_KEY": "sk-test-key"}
	backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath(), mockGetenv(env))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend.Name() != "claude" {
		t.Errorf("got backend %q, want %q", backend.Name(), "claude")
	}
	if result.Selected != "claude" {
		t.Errorf("got result.Selected %q, want %q", result.Selected, "claude")
	}
}

func TestDiscoverBackend_GracefulDegradation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Backend = ""
	cfg.Ollama.Endpoint = "http://127.0.0.1:1" // unreachable

	backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath(), mockGetenv(nil))
	if !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("got err=%v, want ErrBackendUnavailable", err)
	}
	if backend != nil {
		t.Errorf("expected nil backend, got %v", backend)
	}

	// All backends should be unavailable.
	expectedUnavailable := []string{"claude-cli", "gemini-cli", "ollama-cli", "ollama", "claude"}
	if !stringSliceEqual(result.Unavailable, expectedUnavailable) {
		t.Errorf("got Unavailable %v, want %v", result.Unavailable, expectedUnavailable)
	}
}

func TestDiscoverBackend_DiscoveryResultFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Backend = ""

	_, result, err := discoverBackendWith(ctx, cfg, mockLookPath("claude", "ollama"), mockGetenv(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Selected != "claude-cli" {
		t.Errorf("got Selected %q, want %q", result.Selected, "claude-cli")
	}

	wantAvailable := []string{"claude-cli", "ollama-cli"}
	if !stringSliceEqual(result.Available, wantAvailable) {
		t.Errorf("got Available %v, want %v", result.Available, wantAvailable)
	}

	// gemini-cli should be in unavailable (at minimum).
	foundGemini := false
	for _, name := range result.Unavailable {
		if name == "gemini-cli" {
			foundGemini = true
			break
		}
	}
	if !foundGemini {
		t.Errorf("expected gemini-cli in Unavailable, got %v", result.Unavailable)
	}
}

func TestDiscoverAvailableCLIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		available []string
		want      []string
	}{
		{
			name:      "none available",
			available: nil,
			want:      nil,
		},
		{
			name:      "all available",
			available: []string{"claude", "gemini", "ollama"},
			want:      []string{"claude-cli", "gemini-cli", "ollama-cli"},
		},
		{
			name:      "only gemini",
			available: []string{"gemini"},
			want:      []string{"gemini-cli"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := discoverAvailableCLIsWith(mockLookPath(tt.available...))
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverBackend_UserConfigOverridesDiscovery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := DefaultConfig()
	cfg.Backend = "ollama-cli"

	// Even though claude is available, user config takes precedence.
	backend, result, err := discoverBackendWith(ctx, cfg, mockLookPath("claude", "gemini", "ollama"), mockGetenv(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backend.Name() != "ollama-cli" {
		t.Errorf("got backend %q, want %q", backend.Name(), "ollama-cli")
	}
	if result.Selected != "ollama-cli" {
		t.Errorf("got result.Selected %q, want %q", result.Selected, "ollama-cli")
	}
}

// stringSliceEqual compares two string slices for equality.
func stringSliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
