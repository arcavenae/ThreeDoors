package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters/textfile"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/connection"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
)

// TestOption configures how the test app is created.
type TestOption func(*testAppConfig)

type testAppConfig struct {
	width    int
	height   int
	taskFile string
	tasks    []string
	provider core.TaskProvider
	connMgr  *connection.ConnectionManager
}

// WithTermSize sets the virtual terminal dimensions for the test.
func WithTermSize(w, h int) TestOption {
	return func(c *testAppConfig) {
		c.width = w
		c.height = h
	}
}

// WithTaskFile copies a task file into the test's config directory so that
// TextFileProvider can load tasks from it.
func WithTaskFile(path string) TestOption {
	return func(c *testAppConfig) {
		c.taskFile = path
		c.tasks = nil // Don't populate pool manually when using a task file.
	}
}

// WithTasks sets the task texts to populate the pool with.
func WithTasks(texts ...string) TestOption {
	return func(c *testAppConfig) {
		c.tasks = texts
	}
}

// WithProvider sets a custom TaskProvider for the test.
func WithProvider(p core.TaskProvider) TestOption {
	return func(c *testAppConfig) {
		c.provider = p
	}
}

// WithConnMgr sets a ConnectionManager on the test model, enabling source/connect views.
func WithConnMgr(mgr *connection.ConnectionManager) TestOption {
	return func(c *testAppConfig) {
		c.connMgr = mgr
	}
}

// NewTestApp creates a teatest.TestModel wrapping the full TUI application.
// It sets the color profile to ASCII for deterministic output and configures
// a temporary home directory to isolate filesystem side effects.
//
// The returned TestModel can send key messages, retrieve output, and inspect
// the final model state.
func NewTestApp(t *testing.T, opts ...TestOption) *teatest.TestModel {
	t.Helper()

	cfg := testAppConfig{
		width:  80,
		height: 30,
		tasks:  []string{"Task A", "Task B", "Task C"},
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Force ASCII color profile for deterministic, portable output.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.TrueColor)
	})

	// Isolate filesystem access to a temp directory.
	tmpDir := t.TempDir()
	core.SetHomeDir(tmpDir)
	t.Cleanup(func() {
		core.SetHomeDir("")
	})

	// If a task file was provided, copy it into the temp config directory.
	if cfg.taskFile != "" {
		configDir := filepath.Join(tmpDir, ".threedoors")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("create test config dir: %v", err)
		}
		src, err := os.ReadFile(cfg.taskFile)
		if err != nil {
			t.Fatalf("read task file %s: %v", cfg.taskFile, err)
			return nil
		}
		dst := filepath.Join(configDir, "tasks.yaml")
		if err := os.WriteFile(dst, src, 0o644); err != nil {
			t.Fatalf("write task file to %s: %v", dst, err)
		}
	}

	// Build provider.
	var provider core.TaskProvider
	if cfg.provider != nil {
		provider = cfg.provider
	} else {
		provider = &testProvider{}
	}

	// Build pool from explicit task texts.
	pool := core.NewTaskPool()
	for _, text := range cfg.tasks {
		pool.AddTask(core.NewTask(text))
	}

	// If using a task file with no explicit tasks, load from the provider.
	if cfg.taskFile != "" && len(cfg.tasks) == 0 {
		loaded, err := textfile.NewTextFileProvider().LoadTasks()
		if err != nil {
			t.Fatalf("load tasks from file: %v", err)
			return nil
		}
		for _, task := range loaded {
			pool.AddTask(task)
		}
		if pool.Count() == 0 {
			t.Fatal("task file provided but no tasks loaded")
		}
	}

	tracker := core.NewSessionTracker()
	model := NewMainModel(pool, tracker, provider, nil, false, nil)

	if cfg.connMgr != nil {
		model.SetConnectionManager(cfg.connMgr)
	}

	tm := teatest.NewTestModel(
		t,
		model,
		teatest.WithInitialTermSize(cfg.width, cfg.height),
	)
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			// Log but don't fail — program may have already quit.
			fmt.Fprintf(os.Stderr, "quit test model: %v\n", err)
		}
	})

	return tm
}
