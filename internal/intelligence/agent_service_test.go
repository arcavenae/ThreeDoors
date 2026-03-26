package intelligence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/intelligence/llm"
)

// mockBackend implements llm.LLMBackend for testing.
type mockBackend struct {
	name      string
	response  string
	err       error
	available bool
}

func (m *mockBackend) Name() string { return m.name }
func (m *mockBackend) Complete(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}
func (m *mockBackend) Available(_ context.Context) bool { return m.available }

// mockCommandRunner implements llm.CommandRunner for testing.
type mockCommandRunner struct {
	calls []mockCall
	err   error
}

type mockCall struct {
	Dir  string
	Name string
	Args []string
}

func (m *mockCommandRunner) Run(_ context.Context, dir string, name string, args ...string) (string, error) {
	m.calls = append(m.calls, mockCall{Dir: dir, Name: name, Args: args})
	if m.err != nil {
		return "", m.err
	}
	return "", nil
}

func TestNewAgentService(t *testing.T) {
	tests := []struct {
		name    string
		cfg     llm.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ollama config",
			cfg: llm.Config{
				Backend: "ollama",
				Ollama:  llm.OllamaConfig{Endpoint: "http://localhost:11434", Model: "llama3.2"},
				Output:  llm.OutputConfig{OutputRepo: "/tmp/test-repo"},
			},
			wantErr: false,
		},
		{
			name: "missing output repo",
			cfg: llm.Config{
				Backend: "ollama",
			},
			wantErr: true,
			errMsg:  "output_repo is required",
		},
		{
			name: "unknown backend",
			cfg: llm.Config{
				Backend: "gpt4",
				Output:  llm.OutputConfig{OutputRepo: "/tmp/test-repo"},
			},
			wantErr: true,
			errMsg:  "unknown backend",
		},
		{
			name: "claude without api key",
			cfg: llm.Config{
				Backend: "claude",
				Output:  llm.OutputConfig{OutputRepo: "/tmp/test-repo"},
			},
			wantErr: true,
			errMsg:  "ANTHROPIC_API_KEY",
		},
		// Auto-discovery with empty Backend is environment-dependent
		// (probes PATH for CLI tools). It's tested via discoverBackendWith
		// in the llm package with mock LookPath. Omitted here to avoid
		// flaky results across CI vs local environments.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv
			t.Setenv("ANTHROPIC_API_KEY", "")

			svc, err := NewAgentService(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}
			if svc == nil {
				t.Fatal("expected non-nil service")
				return
			}
		})
	}
}

func TestDecomposeAndWrite(t *testing.T) {
	t.Parallel()

	validJSON := `[{
		"epic": "14",
		"story_id": "14.3",
		"title": "Test Story",
		"user_story": "As a developer, I want tests, So that quality is ensured",
		"acceptance_criteria": ["AC1: Tests pass"],
		"tasks": ["Write tests"],
		"dev_notes": "Use table-driven tests"
	}]`

	tests := []struct {
		name        string
		taskDesc    string
		backend     *mockBackend
		runner      *mockCommandRunner
		initGit     bool
		wantErr     bool
		errContains string
		wantStories int
	}{
		{
			name:     "successful decomposition and write",
			taskDesc: "Build a login form",
			backend: &mockBackend{
				name:      "test",
				response:  validJSON,
				available: true,
			},
			runner:      &mockCommandRunner{},
			initGit:     true,
			wantErr:     false,
			wantStories: 1,
		},
		{
			name:     "empty task description",
			taskDesc: "",
			backend: &mockBackend{
				name:      "test",
				response:  validJSON,
				available: true,
			},
			runner:      &mockCommandRunner{},
			wantErr:     true,
			errContains: "must not be empty",
		},
		{
			name:     "whitespace-only task description",
			taskDesc: "   \n\t  ",
			backend: &mockBackend{
				name:      "test",
				response:  validJSON,
				available: true,
			},
			runner:      &mockCommandRunner{},
			wantErr:     true,
			errContains: "must not be empty",
		},
		{
			name:     "LLM returns error",
			taskDesc: "Build a login form",
			backend: &mockBackend{
				name:      "test",
				err:       fmt.Errorf("connection refused"),
				available: false,
			},
			runner:      &mockCommandRunner{},
			wantErr:     true,
			errContains: "decompose task",
		},
		{
			name:     "LLM returns empty stories",
			taskDesc: "Build a login form",
			backend: &mockBackend{
				name:      "test",
				response:  `[]`,
				available: true,
			},
			runner:      &mockCommandRunner{},
			wantErr:     true,
			errContains: "could not parse",
		},
		{
			name:     "git write failure",
			taskDesc: "Build a login form",
			backend: &mockBackend{
				name:      "test",
				response:  validJSON,
				available: true,
			},
			runner:      &mockCommandRunner{err: fmt.Errorf("git error")},
			wantErr:     true,
			errContains: "write stories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decomposer := llm.NewTaskDecomposer(tt.backend)
			repoPath := t.TempDir()
			if tt.initGit {
				if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755); err != nil {
					t.Fatalf("failed to create .git dir: %v", err)
				}
			}
			writer := llm.NewGitOutputWriter(repoPath, "story/", tt.runner)

			svc := NewAgentServiceWithDeps(decomposer, writer)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			result, err := svc.DecomposeAndWrite(ctx, tt.taskDesc)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}
			if result == nil {
				t.Fatal("expected non-nil result")
				return
			}
			if len(result.Stories) != tt.wantStories {
				t.Errorf("got %d stories, want %d", len(result.Stories), tt.wantStories)
			}
		})
	}
}

func TestDecomposeAndWriteContextCancellation(t *testing.T) {
	t.Parallel()

	backend := &mockBackend{
		name:      "test",
		err:       context.Canceled,
		available: true,
	}
	decomposer := llm.NewTaskDecomposer(backend)
	writer := llm.NewGitOutputWriter(t.TempDir(), "story/", &mockCommandRunner{})
	svc := NewAgentServiceWithDeps(decomposer, writer)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := svc.DecomposeAndWrite(ctx, "Some task")
	if err == nil {
		t.Fatal("expected error from cancelled context")
		return
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstr(s, substr)
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
