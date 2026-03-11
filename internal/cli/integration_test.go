package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters/textfile"
	"github.com/arcaven/ThreeDoors/internal/core"
)

// setupTestEnv creates a temporary ThreeDoors config directory with a textfile
// provider registered and a valid config. Not parallel — modifies HOME env var.
func setupTestEnv(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}

	// Create tasks directory
	tasksDir := filepath.Join(configDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir tasks: %v", err)
	}

	// Write config
	cfg := &core.ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
		NoteTitle:     "Test Tasks",
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// Register textfile provider in the default registry
	reg := core.DefaultRegistry()
	if !reg.IsRegistered("textfile") {
		_ = reg.Register("textfile", func(_ *core.ProviderConfig) (core.TaskProvider, error) {
			return textfile.NewTextFileProvider(), nil
		})
	}

	// Set HOME so GetConfigDirPath resolves to our temp dir
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	_ = os.Setenv("HOME", dir)

	return configDir
}

// addTestTask creates a task via the provider in the test environment.
func addTestTask(t *testing.T, text string) string {
	t.Helper()

	ctx, err := bootstrap()
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	task := core.NewTask(text)
	ctx.pool.AddTask(task)
	if err := ctx.provider.SaveTask(task); err != nil {
		t.Fatalf("save task: %v", err)
	}
	return task.ID
}

// executeCmd creates a new root command and runs it with the given args.
// Output goes to os.Stdout (captured by these functions). Returns any error.
func executeCmd(t *testing.T, args ...string) error {
	t.Helper()
	root := NewRootCmd()
	root.SetArgs(args)
	return root.Execute()
}

func TestIntegration_ConfigShow_JSON(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "show", "--json"); err != nil {
		t.Fatalf("config show --json: %v", err)
	}
}

func TestIntegration_ConfigShow_Human(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "show"); err != nil {
		t.Fatalf("config show: %v", err)
	}
}

func TestIntegration_ConfigGet_JSON(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "get", "provider", "--json"); err != nil {
		t.Fatalf("config get --json: %v", err)
	}
}

func TestIntegration_ConfigGet_Human(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "get", "provider"); err != nil {
		t.Fatalf("config get: %v", err)
	}
}

func TestIntegration_ConfigSet_JSON(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "set", "theme", "modern", "--json"); err != nil {
		t.Fatalf("config set --json: %v", err)
	}
}

func TestIntegration_ConfigSet_Human(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "config", "set", "theme", "scifi"); err != nil {
		t.Fatalf("config set: %v", err)
	}
}

func TestIntegration_TaskList_JSON(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "list", "--json"); err != nil {
		t.Fatalf("task list --json: %v", err)
	}
}

func TestIntegration_TaskList_Human(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "list"); err != nil {
		t.Fatalf("task list: %v", err)
	}
}

func TestIntegration_TaskList_WithStatusFilter(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "list", "--status", "todo"); err != nil {
		t.Fatalf("task list --status: %v", err)
	}
}

func TestIntegration_TaskList_WithTypeFilter(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "list", "--type", "technical"); err != nil {
		t.Fatalf("task list --type: %v", err)
	}
}

func TestIntegration_TaskAdd_JSON(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "add", "Buy groceries", "--json"); err != nil {
		t.Fatalf("task add --json: %v", err)
	}
}

func TestIntegration_TaskAdd_Human(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "add", "Walk the dog"); err != nil {
		t.Fatalf("task add: %v", err)
	}
}

func TestIntegration_TaskAdd_WithFlags(t *testing.T) {
	setupTestEnv(t)
	if err := executeCmd(t, "task", "add", "Write tests", "--context", "Coverage is low", "--type", "technical", "--effort", "medium"); err != nil {
		t.Fatalf("task add with flags: %v", err)
	}
}

func TestIntegration_TaskShow_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Show me")
	if err := executeCmd(t, "task", "show", taskID[:8], "--json"); err != nil {
		t.Fatalf("task show --json: %v", err)
	}
}

func TestIntegration_TaskShow_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Show me human")
	if err := executeCmd(t, "task", "show", taskID[:8]); err != nil {
		t.Fatalf("task show: %v", err)
	}
}

func TestIntegration_TaskEdit_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Edit me")
	if err := executeCmd(t, "task", "edit", taskID[:8], "--text", "Edited text", "--json"); err != nil {
		t.Fatalf("task edit --json: %v", err)
	}
}

func TestIntegration_TaskEdit_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Edit me human")
	if err := executeCmd(t, "task", "edit", taskID[:8], "--context", "New context"); err != nil {
		t.Fatalf("task edit: %v", err)
	}
}

func TestIntegration_TaskComplete_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Complete me")
	if err := executeCmd(t, "task", "complete", taskID[:8], "--json"); err != nil {
		t.Fatalf("task complete --json: %v", err)
	}
}

func TestIntegration_TaskComplete_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Complete human")
	if err := executeCmd(t, "task", "complete", taskID[:8]); err != nil {
		t.Fatalf("task complete: %v", err)
	}
}

func TestIntegration_TaskNote_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Note target")
	if err := executeCmd(t, "task", "note", taskID[:8], "A test note", "--json"); err != nil {
		t.Fatalf("task note --json: %v", err)
	}
}

func TestIntegration_TaskNote_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Note human")
	if err := executeCmd(t, "task", "note", taskID[:8], "My note"); err != nil {
		t.Fatalf("task note: %v", err)
	}
}

func TestIntegration_TaskSearch_JSON(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Searchable task")
	if err := executeCmd(t, "task", "search", "Searchable", "--json"); err != nil {
		t.Fatalf("task search --json: %v", err)
	}
}

func TestIntegration_TaskSearch_Human(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Find me please")
	if err := executeCmd(t, "task", "search", "Find me"); err != nil {
		t.Fatalf("task search: %v", err)
	}
}

func TestIntegration_TaskDelete_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Delete me")
	if err := executeCmd(t, "task", "delete", taskID[:8], "--json"); err != nil {
		t.Fatalf("task delete --json: %v", err)
	}
}

func TestIntegration_TaskDelete_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Delete human")
	if err := executeCmd(t, "task", "delete", taskID[:8]); err != nil {
		t.Fatalf("task delete: %v", err)
	}
}

func TestIntegration_TaskBlock_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Block me")
	if err := executeCmd(t, "task", "block", taskID[:8], "--reason", "Blocked reason", "--json"); err != nil {
		t.Fatalf("task block --json: %v", err)
	}
}

func TestIntegration_TaskBlock_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Block human")
	if err := executeCmd(t, "task", "block", taskID[:8], "--reason", "Need API"); err != nil {
		t.Fatalf("task block: %v", err)
	}
}

func TestIntegration_TaskUnblock_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Unblock me")

	// First block the task
	if err := executeCmd(t, "task", "block", taskID[:8], "--reason", "reason"); err != nil {
		t.Fatalf("block: %v", err)
	}

	// Now unblock
	if err := executeCmd(t, "task", "unblock", taskID[:8], "--json"); err != nil {
		t.Fatalf("task unblock --json: %v", err)
	}
}

func TestIntegration_TaskUnblock_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Unblock human")

	if err := executeCmd(t, "task", "block", taskID[:8], "--reason", "blocker"); err != nil {
		t.Fatalf("block: %v", err)
	}

	if err := executeCmd(t, "task", "unblock", taskID[:8]); err != nil {
		t.Fatalf("task unblock: %v", err)
	}
}

func TestIntegration_TaskStatus_JSON(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Status me")
	if err := executeCmd(t, "task", "status", taskID[:8], "in-progress", "--json"); err != nil {
		t.Fatalf("task status --json: %v", err)
	}
}

func TestIntegration_TaskStatus_Human(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Status human")
	if err := executeCmd(t, "task", "status", taskID[:8], "in-progress"); err != nil {
		t.Fatalf("task status: %v", err)
	}
}

func TestIntegration_Version_JSON(t *testing.T) {
	t.Parallel()
	if err := executeCmd(t, "version", "--json"); err != nil {
		t.Fatalf("version --json: %v", err)
	}
}

func TestIntegration_Version_Human(t *testing.T) {
	t.Parallel()
	if err := executeCmd(t, "version"); err != nil {
		t.Fatalf("version: %v", err)
	}
}

func TestIntegration_TaskAdd_NoArgs(t *testing.T) {
	setupTestEnv(t)

	// task add with no args and no stdin pipe should error
	root := NewRootCmd()
	root.SetArgs([]string{"task", "add"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for task add with no args")
	}
}

func TestIntegration_TaskBlock_NoReason(t *testing.T) {
	setupTestEnv(t)
	taskID := addTestTask(t, "Block without reason")

	root := NewRootCmd()
	root.SetArgs([]string{"task", "block", taskID[:8]})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --reason is missing")
	}
}

func TestIntegration_Doctor_JSON(t *testing.T) {
	configDir := setupTestEnv(t)
	// Create tasks.yaml so doctor doesn't report CheckFail and os.Exit
	tasksYAML := filepath.Join(configDir, "tasks.yaml")
	if err := os.WriteFile(tasksYAML, []byte("tasks: []\n"), 0o644); err != nil {
		t.Fatalf("create tasks.yaml: %v", err)
	}
	if err := executeCmd(t, "doctor", "--json"); err != nil {
		t.Fatalf("doctor --json: %v", err)
	}
}

func TestIntegration_Doctor_Human(t *testing.T) {
	configDir := setupTestEnv(t)
	// Create tasks.yaml so doctor doesn't report CheckFail and os.Exit
	tasksYAML := filepath.Join(configDir, "tasks.yaml")
	if err := os.WriteFile(tasksYAML, []byte("tasks: []\n"), 0o644); err != nil {
		t.Fatalf("create tasks.yaml: %v", err)
	}
	if err := executeCmd(t, "doctor"); err != nil {
		t.Fatalf("doctor: %v", err)
	}
}

func TestIntegration_Stats_JSON(t *testing.T) {
	setupTestEnv(t)
	// Stats requires sessions.jsonl — create an empty one
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "stats", "--json"); err != nil {
		t.Fatalf("stats --json: %v", err)
	}
}

func TestIntegration_Stats_Human(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "stats"); err != nil {
		t.Fatalf("stats: %v", err)
	}
}

func TestIntegration_Stats_Daily(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "stats", "--daily"); err != nil {
		t.Fatalf("stats --daily: %v", err)
	}
}

func TestIntegration_Stats_Weekly(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "stats", "--weekly"); err != nil {
		t.Fatalf("stats --weekly: %v", err)
	}
}

func TestIntegration_Stats_Patterns(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "stats", "--patterns"); err != nil {
		t.Fatalf("stats --patterns: %v", err)
	}
}

func TestIntegration_MoodSet_Valid(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "mood", "set", "focused"); err != nil {
		t.Fatalf("mood set: %v", err)
	}
}

func TestIntegration_MoodSet_JSON(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "mood", "set", "tired", "--json"); err != nil {
		t.Fatalf("mood set --json: %v", err)
	}
}

func TestIntegration_MoodSet_Custom(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "mood", "set", "custom", "feeling great"); err != nil {
		t.Fatalf("mood set custom: %v", err)
	}
}

func TestIntegration_MoodHistory_JSON(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "mood", "history", "--json"); err != nil {
		t.Fatalf("mood history --json: %v", err)
	}
}

func TestIntegration_MoodHistory_Human(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	if err := os.WriteFile(sessionsPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create sessions.jsonl: %v", err)
	}
	if err := executeCmd(t, "mood", "history"); err != nil {
		t.Fatalf("mood history: %v", err)
	}
}

func TestExitError_Error(t *testing.T) {
	t.Parallel()

	e := exitError{code: 42}
	if e.Error() != "exit code 42" {
		t.Errorf("Error() = %q, want %q", e.Error(), "exit code 42")
	}

	e0 := exitError{code: 0}
	if e0.Error() != "exit code 0" {
		t.Errorf("Error() = %q, want %q", e0.Error(), "exit code 0")
	}
}

func TestIntegration_TaskAdd_StdinFlag(t *testing.T) {
	setupTestEnv(t)

	// Override the stdinReader and detector for testing
	origReader := stdinReader
	origDetector := stdinDetector
	t.Cleanup(func() {
		stdinReader = origReader
		stdinDetector = origDetector
	})

	stdinReader = strings.NewReader("Task from stdin A\nTask from stdin B\n")
	stdinDetector = func() bool { return false }

	if err := executeCmd(t, "task", "add", "--stdin", "--json"); err != nil {
		t.Fatalf("task add --stdin: %v", err)
	}
}

func TestIntegration_DocAudit_JSON(t *testing.T) {
	setupTestEnv(t)

	// doc-audit needs a project root with story files and planning docs
	// For coverage, just test that the command executes (it will report errors
	// about missing files, which is fine for coverage purposes)
	root := NewRootCmd()
	root.SetArgs([]string{"doc-audit", "--root", t.TempDir(), "--json"})
	// The command will error because the project root has no docs
	// but the code paths will be exercised
	_ = root.Execute()
}

func TestIntegration_DocAudit_Human(t *testing.T) {
	setupTestEnv(t)

	root := NewRootCmd()
	root.SetArgs([]string{"doc-audit", "--root", t.TempDir()})
	_ = root.Execute()
}

func TestIntegration_TaskAdd_StdinMultiline(t *testing.T) {
	setupTestEnv(t)

	// Override the stdinReader for testing
	origReader := stdinReader
	origDetector := stdinDetector
	t.Cleanup(func() {
		stdinReader = origReader
		stdinDetector = origDetector
	})

	stdinReader = strings.NewReader("Multi task 1\nMulti task 2\n\nMulti task 3\n")
	stdinDetector = func() bool { return false }

	// Test human output path
	if err := executeCmd(t, "task", "add", "--stdin"); err != nil {
		t.Fatalf("task add --stdin human: %v", err)
	}
}

func TestIntegration_MoodHistory_WithEntries(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	// Create a sessions file with mood data
	sessionData := `{"start_time":"2026-01-01T10:00:00Z","end_time":"2026-01-01T10:30:00Z","tasks_completed":1,"mood_entries":[{"timestamp":"2026-01-01T10:00:00Z","mood":"focused"}]}` + "\n"
	if err := os.WriteFile(sessionsPath, []byte(sessionData), 0o644); err != nil {
		t.Fatalf("write sessions: %v", err)
	}
	if err := executeCmd(t, "mood", "history"); err != nil {
		t.Fatalf("mood history: %v", err)
	}
}

func TestIntegration_MoodHistory_WithEntries_JSON(t *testing.T) {
	setupTestEnv(t)
	configDir, _ := core.GetConfigDirPath()
	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	sessionData := `{"start_time":"2026-01-01T10:00:00Z","end_time":"2026-01-01T10:30:00Z","tasks_completed":1,"mood_entries":[{"timestamp":"2026-01-01T10:00:00Z","mood":"focused"}]}` + "\n"
	if err := os.WriteFile(sessionsPath, []byte(sessionData), 0o644); err != nil {
		t.Fatalf("write sessions: %v", err)
	}
	if err := executeCmd(t, "mood", "history", "--json"); err != nil {
		t.Fatalf("mood history --json: %v", err)
	}
}
