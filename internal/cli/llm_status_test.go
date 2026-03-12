package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
)

// testLookPath creates a LookPathFunc that reports the given commands as available.
func testLookPath(available ...string) llm.LookPathFunc {
	set := make(map[string]bool, len(available))
	for _, cmd := range available {
		set[cmd] = true
	}
	return func(file string) (string, error) {
		if set[file] {
			return "/usr/bin/" + file, nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
}

// testGetenv creates a getenv func from a map.
func testGetenv(vars map[string]string) func(string) string {
	return func(key string) string {
		if vars == nil {
			return ""
		}
		return vars[key]
	}
}

func TestRunLLMStatusWith_HumanOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		availableCmds []string
		env           map[string]string
		wantContains  []string
	}{
		{
			name:          "claude available",
			availableCmds: []string{"claude"},
			wantContains:  []string{"Active Backend", "claude-cli", "CLI", "/usr/bin/claude", "reachable"},
		},
		{
			name:          "no backends available",
			availableCmds: nil,
			wantContains:  []string{"No LLM backends available", "claude", "gemini", "ollama"},
		},
		{
			name:          "multiple backends show fallbacks",
			availableCmds: []string{"claude", "gemini", "ollama"},
			wantContains:  []string{"claude-cli", "Fallback Backends", "gemini-cli", "ollama-cli"},
		},
		{
			name:          "services shown when backend available",
			availableCmds: []string{"claude"},
			wantContains:  []string{"Services", "decompose", "enrich", "breakdown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, false)
			cfg := llm.DefaultConfig()
			cfg.Backend = ""
			cfg.Ollama.Endpoint = "http://127.0.0.1:1"

			err := runLLMStatusWith(context.Background(), formatter, cfg, testLookPath(tt.availableCmds...), testGetenv(tt.env))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q:\n%s", want, output)
				}
			}
		})
	}
}

func TestRunLLMStatusWith_JSONOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		availableCmds []string
		env           map[string]string
		checkJSON     func(t *testing.T, data llmStatusJSON)
	}{
		{
			name:          "claude available returns structured JSON",
			availableCmds: []string{"claude"},
			checkJSON: func(t *testing.T, data llmStatusJSON) {
				t.Helper()
				if data.Backend == nil {
					t.Fatal("expected backend, got nil")
				}
				if data.Backend.Name != "claude-cli" {
					t.Errorf("got backend name %q, want %q", data.Backend.Name, "claude-cli")
				}
				if data.Backend.Type != "CLI" {
					t.Errorf("got type %q, want %q", data.Backend.Type, "CLI")
				}
				if data.Backend.CommandPath != "/usr/bin/claude" {
					t.Errorf("got command path %q, want %q", data.Backend.CommandPath, "/usr/bin/claude")
				}
				if !data.Backend.Available {
					t.Error("expected backend to be available")
				}
				if len(data.Services) == 0 {
					t.Error("expected services, got none")
				}
				for _, svc := range data.Services {
					if !svc.Ready {
						t.Errorf("expected service %q to be ready", svc.Name)
					}
				}
			},
		},
		{
			name:          "no backends returns null backend",
			availableCmds: nil,
			checkJSON: func(t *testing.T, data llmStatusJSON) {
				t.Helper()
				if data.Backend != nil {
					t.Errorf("expected nil backend, got %+v", data.Backend)
				}
				for _, svc := range data.Services {
					if svc.Ready {
						t.Errorf("expected service %q to be not ready", svc.Name)
					}
				}
			},
		},
		{
			name:          "multiple backends shows fallbacks",
			availableCmds: []string{"claude", "gemini", "ollama"},
			checkJSON: func(t *testing.T, data llmStatusJSON) {
				t.Helper()
				if len(data.Fallbacks) < 2 {
					t.Errorf("expected at least 2 fallbacks, got %d", len(data.Fallbacks))
				}
				// The selected backend (claude-cli) should not appear in fallbacks.
				for _, fb := range data.Fallbacks {
					if fb.Name == "claude-cli" {
						t.Error("selected backend should not appear in fallbacks")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			cfg := llm.DefaultConfig()
			cfg.Backend = ""
			cfg.Ollama.Endpoint = "http://127.0.0.1:1"

			err := runLLMStatusWith(context.Background(), formatter, cfg, testLookPath(tt.availableCmds...), testGetenv(tt.env))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var envelope JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
				t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
			}
			if envelope.Command != "llm status" {
				t.Errorf("got command %q, want %q", envelope.Command, "llm status")
			}

			// Re-marshal and unmarshal data field to get typed struct.
			dataBytes, err := json.Marshal(envelope.Data)
			if err != nil {
				t.Fatalf("marshal data: %v", err)
			}
			var data llmStatusJSON
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				t.Fatalf("unmarshal data: %v", err)
			}

			tt.checkJSON(t, data)
		})
	}
}

func TestClassifyBackendType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		wantType string
	}{
		{"claude-cli", "CLI"},
		{"gemini-cli", "CLI"},
		{"ollama-cli", "CLI"},
		{"custom", "CLI"},
		{"claude", "HTTP"},
		{"ollama", "HTTP"},
		{"unknown-thing", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyBackendType(tt.name)
			if got != tt.wantType {
				t.Errorf("classifyBackendType(%q) = %q, want %q", tt.name, got, tt.wantType)
			}
		})
	}
}

func TestResolveCommandPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		backend  string
		wantPath string
	}{
		{"claude-cli resolves", "claude-cli", "/usr/bin/claude"},
		{"gemini-cli resolves", "gemini-cli", "/usr/bin/gemini"},
		{"ollama-cli resolves", "ollama-cli", "/usr/bin/ollama"},
		{"HTTP backend returns empty", "claude", ""},
		{"unknown returns empty", "unknown", ""},
	}

	lp := testLookPath("claude", "gemini", "ollama")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveCommandPath(tt.backend, lp)
			if got != tt.wantPath {
				t.Errorf("resolveCommandPath(%q) = %q, want %q", tt.backend, got, tt.wantPath)
			}
		})
	}
}

func TestBuildFallbackNames(t *testing.T) {
	t.Parallel()

	discovery := &llm.DiscoveryResult{
		Selected:    "claude-cli",
		Available:   []string{"claude-cli", "gemini-cli"},
		Unavailable: []string{"ollama-cli"},
	}

	names := buildFallbackNames(discovery)
	if len(names) != 2 {
		t.Fatalf("expected 2 fallbacks, got %d: %v", len(names), names)
	}
	for _, name := range names {
		if name == "claude-cli" {
			t.Error("selected backend should not appear in fallback names")
		}
	}
}

func TestKnownServices(t *testing.T) {
	t.Parallel()

	svcs := knownServices()
	if len(svcs) == 0 {
		t.Fatal("expected at least one service")
	}
	expected := map[string]bool{"decompose": true, "enrich": true, "breakdown": true}
	for _, svc := range svcs {
		if !expected[svc] {
			t.Errorf("unexpected service %q", svc)
		}
	}
}
