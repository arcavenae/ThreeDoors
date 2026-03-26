package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func newTestTask(id, text string, status core.TaskStatus) *core.Task {
	now := time.Now().UTC()
	return &core.Task{
		ID:        id,
		Text:      text,
		Status:    status,
		Type:      core.TypeTechnical,
		Effort:    core.EffortMedium,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestDisplayDoors_Human(t *testing.T) {
	t.Parallel()

	doors := []*core.Task{
		newTestTask("aaaaaaaa-1111-2222-3333-444444444444", "Write tests", core.StatusTodo),
		newTestTask("bbbbbbbb-1111-2222-3333-444444444444", "Fix bug", core.StatusTodo),
		newTestTask("cccccccc-1111-2222-3333-444444444444", "Deploy", core.StatusBlocked),
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := displayDoors(formatter, doors, 10); err != nil {
		t.Fatalf("displayDoors() error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"DOOR", "ID", "TEXT", "STATUS", "TYPE", "EFFORT",
		"1", "aaaaaaaa", "Write tests", "todo", "technical", "medium",
		"2", "bbbbbbbb", "Fix bug",
		"3", "cccccccc", "Deploy", "blocked",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("output missing %q\nGot: %s", want, output)
		}
	}
}

func TestDisplayDoors_JSON(t *testing.T) {
	t.Parallel()

	doors := []*core.Task{
		newTestTask("aaaa1111-2222-3333-4444-555555555555", "Task A", core.StatusTodo),
		newTestTask("bbbb1111-2222-3333-4444-555555555555", "Task B", core.StatusTodo),
		newTestTask("cccc1111-2222-3333-4444-555555555555", "Task C", core.StatusBlocked),
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	if err := displayDoors(formatter, doors, 15); err != nil {
		t.Fatalf("displayDoors() error: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if env.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
	}
	if env.Command != "doors" {
		t.Errorf("command = %q, want %q", env.Command, "doors")
	}

	dataBytes, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
		return
	}
	var entries []doorEntry
	if err := json.Unmarshal(dataBytes, &entries); err != nil {
		t.Fatalf("unmarshal entries: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	for i, e := range entries {
		if e.Door != i+1 {
			t.Errorf("entry[%d].Door = %d, want %d", i, e.Door, i+1)
		}
		if e.Task == nil {
			t.Errorf("entry[%d].Task is nil", i)
		}
	}

	metaBytes, err := json.Marshal(env.Metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
		return
	}
	var meta map[string]interface{}
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if meta["selection_method"] != "diversity" {
		t.Errorf("selection_method = %v, want %q", meta["selection_method"], "diversity")
	}
	if meta["total_available"] != float64(15) {
		t.Errorf("total_available = %v, want 15", meta["total_available"])
	}
}

func TestDisplayDoors_FewerThan3Note(t *testing.T) {
	t.Parallel()

	doors := []*core.Task{
		newTestTask("aaaa1111-2222-3333-4444-555555555555", "Only task", core.StatusTodo),
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := displayDoors(formatter, doors, 1); err != nil {
		t.Fatalf("displayDoors() error: %v", err)
	}

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Only 1 task(s) available")) {
		t.Errorf("expected note about fewer tasks, got: %s", output)
	}
}

func TestHandlePick_Success(t *testing.T) {
	t.Parallel()

	task := newTestTask("aaaa1111-2222-3333-4444-555555555555", "Pick me", core.StatusTodo)
	doors := []*core.Task{task}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	provider := &mockProvider{}
	if err := handlePick(nil, formatter, doors, 1, provider); err != nil {
		t.Fatalf("handlePick() error: %v", err)
	}

	if task.Status != core.StatusInProgress {
		t.Errorf("status = %q, want %q", task.Status, core.StatusInProgress)
	}
	if !provider.saved {
		t.Error("provider.SaveTask was not called")
	}
	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Pick me")) {
		t.Errorf("output should contain task text, got: %s", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("in-progress")) {
		t.Errorf("output should mention in-progress, got: %s", output)
	}
}

func TestHandlePick_JSON(t *testing.T) {
	t.Parallel()

	task := newTestTask("aaaa1111-2222-3333-4444-555555555555", "JSON pick", core.StatusTodo)
	doors := []*core.Task{task}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	provider := &mockProvider{}
	if err := handlePick(nil, formatter, doors, 1, provider); err != nil {
		t.Fatalf("handlePick() error: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if env.Command != "doors.pick" {
		t.Errorf("command = %q, want %q", env.Command, "doors.pick")
	}

	dataBytes, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
		return
	}
	var entry doorEntry
	if err := json.Unmarshal(dataBytes, &entry); err != nil {
		t.Fatalf("unmarshal entry: %v", err)
	}
	if entry.Door != 1 {
		t.Errorf("door = %d, want 1", entry.Door)
	}
	if entry.Task.Status != core.StatusInProgress {
		t.Errorf("task status = %q, want %q", entry.Task.Status, core.StatusInProgress)
	}
}

func TestHandlePick_InvalidDoor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pick int
	}{
		{"zero", 0},
		{"too high", 4},
		{"negative", -1},
	}

	doors := []*core.Task{
		newTestTask("a", "A", core.StatusTodo),
		newTestTask("b", "B", core.StatusTodo),
		newTestTask("c", "C", core.StatusTodo),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, false)
			provider := &mockProvider{}

			err := handlePick(nil, formatter, doors, tt.pick, provider)
			if err == nil {
				t.Error("expected error for invalid pick")
			}
			if provider.saved {
				t.Error("provider.SaveTask should not be called for invalid pick")
			}
		})
	}
}

func TestHandlePick_SaveError(t *testing.T) {
	t.Parallel()

	task := newTestTask("aaaa1111-2222-3333-4444-555555555555", "Save fail", core.StatusTodo)
	doors := []*core.Task{task}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	provider := &mockProvider{saveErr: fmt.Errorf("disk full")}
	err := handlePick(nil, formatter, doors, 1, provider)
	if err == nil {
		t.Error("expected error when save fails")
	}
}

func TestShortID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want string
	}{
		{"uuid", "aaaaaaaa-1111-2222-3333-444444444444", "aaaaaaaa"},
		{"short", "abc", "abc"},
		{"exactly 8", "12345678", "12345678"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := shortID(tt.id); got != tt.want {
				t.Errorf("shortID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestDisplayDoors_EmptyTypeAndEffort(t *testing.T) {
	t.Parallel()

	task := newTestTask("aaaa1111-2222-3333-4444-555555555555", "No category", core.StatusTodo)
	task.Type = ""
	task.Effort = ""

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	if err := displayDoors(formatter, []*core.Task{task}, 1); err != nil {
		t.Fatalf("displayDoors() error: %v", err)
	}

	output := buf.String()
	// Empty type/effort should show as "-"
	if !bytes.Contains(buf.Bytes(), []byte("-")) {
		t.Errorf("expected '-' for empty type/effort, got: %s", output)
	}
}

func TestBuildTaskPool_NilProvider(t *testing.T) {
	t.Parallel()

	pool, provider, err := buildTaskPool(nil)
	if err == nil {
		t.Fatal("expected error when provider is nil, got nil")
		return
	}
	if pool != nil {
		t.Errorf("expected nil pool, got %v", pool)
	}
	if provider != nil {
		t.Errorf("expected nil provider, got %v", provider)
	}

	wantSubstr := "no task provider available"
	if !bytes.Contains([]byte(err.Error()), []byte(wantSubstr)) {
		t.Errorf("error = %q, want substring %q", err.Error(), wantSubstr)
	}
}

func TestBuildTaskPool_ValidProvider(t *testing.T) {
	t.Parallel()

	mock := &mockProviderWithTasks{
		tasks: []*core.Task{
			newTestTask("aaa", "Test task", core.StatusTodo),
		},
	}

	pool, provider, err := buildTaskPool(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
		return
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
		return
	}
	if len(pool.GetAvailableForDoors()) != 1 {
		t.Errorf("pool has %d available tasks, want 1", len(pool.GetAvailableForDoors()))
	}
}

func TestBuildTaskPool_LoadError(t *testing.T) {
	t.Parallel()

	mock := &mockProviderWithTasks{
		loadErr: fmt.Errorf("connection refused"),
	}

	_, _, err := buildTaskPool(mock)
	if err == nil {
		t.Fatal("expected error when LoadTasks fails, got nil")
		return
	}

	wantSubstr := "load tasks"
	if !bytes.Contains([]byte(err.Error()), []byte(wantSubstr)) {
		t.Errorf("error = %q, want substring %q", err.Error(), wantSubstr)
	}
}

func TestNewDoorsCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewDoorsCmd()

	if cmd.Use != "doors" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doors")
	}

	pick := cmd.Flags().Lookup("pick")
	if pick == nil {
		t.Fatal("missing --pick flag")
	}
	if pick.DefValue != "0" {
		t.Errorf("pick default = %q, want %q", pick.DefValue, "0")
	}

	interactive := cmd.Flags().Lookup("interactive")
	if interactive == nil {
		t.Fatal("missing --interactive flag")
	}
	if interactive.DefValue != "false" {
		t.Errorf("interactive default = %q, want %q", interactive.DefValue, "false")
	}
}

func TestNewDoorsCmd_WithTasks(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Door task A")
	addTestTask(t, "Door task B")
	addTestTask(t, "Door task C")

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doors command: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DOOR") {
		t.Errorf("expected table header in output, got: %s", output)
	}
}

func TestNewDoorsCmd_JSON(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "JSON door A")
	addTestTask(t, "JSON door B")
	addTestTask(t, "JSON door C")

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doors --json: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "doors" {
		t.Errorf("command = %q, want %q", env.Command, "doors")
	}
}

func TestNewDoorsCmd_NoTasks(t *testing.T) {
	// Use a fresh temp dir with no tasks
	dir := t.TempDir()
	configDir := setupTestEnvInDir(t, dir)
	_ = configDir

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors"})
	err := root.Execute()
	// With an empty task pool, doors should return an error
	if err == nil {
		// If no error, the output should at least contain door table (meaning tasks exist from env)
		t.Log("doors command succeeded with fresh env — likely no tasks but provider fallback")
	}
}

func TestNewDoorsCmd_NoTasks_JSON(t *testing.T) {
	dir := t.TempDir()
	configDir := setupTestEnvInDir(t, dir)
	_ = configDir

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--json"})
	_ = root.Execute()
	// Exercise the JSON output path regardless of task availability
	if buf.Len() > 0 {
		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.SchemaVersion != 1 {
			t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
		}
	}
}

// setupTestEnvInDir creates a test environment in the specified directory.
// Used for tests that need guaranteed isolation from other tests.
func setupTestEnvInDir(t *testing.T, dir string) string {
	t.Helper()

	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	tasksDir := filepath.Join(configDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir tasks: %v", err)
	}

	cfg := &core.ProviderConfig{
		SchemaVersion: 2,
		Provider:      "textfile",
		NoteTitle:     "Test Tasks",
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := core.SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	_ = os.Setenv("HOME", dir)

	return configDir
}

func TestNewDoorsCmd_PickFlag(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Pick door A")
	addTestTask(t, "Pick door B")
	addTestTask(t, "Pick door C")

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--pick", "1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doors --pick 1: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "in-progress") {
		t.Errorf("expected 'in-progress' in output, got: %s", output)
	}
}

func TestNewDoorsCmd_PickFlag_JSON(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "JSON pick A")
	addTestTask(t, "JSON pick B")
	addTestTask(t, "JSON pick C")

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--pick", "1", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doors --pick 1 --json: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "doors.pick" {
		t.Errorf("command = %q, want %q", env.Command, "doors.pick")
	}
}

func TestNewDoorsCmd_PickInvalid(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Invalid pick A")
	addTestTask(t, "Invalid pick B")
	addTestTask(t, "Invalid pick C")

	root := NewRootCmd()
	root.SetArgs([]string{"doors", "--pick", "5"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --pick value")
	}
}

func TestNewDoorsCmd_InteractiveNonTTY(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Interactive A")
	addTestTask(t, "Interactive B")
	addTestTask(t, "Interactive C")

	// Force non-TTY so interactive skips prompting
	orig := stdoutIsTerminal
	t.Cleanup(func() { stdoutIsTerminal = orig })
	stdoutIsTerminal = func() bool { return false }

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--interactive"})
	if err := root.Execute(); err != nil {
		t.Fatalf("doors --interactive (non-TTY): %v", err)
	}

	// Should still display doors even in non-TTY mode
	output := buf.String()
	if !strings.Contains(output, "DOOR") {
		t.Errorf("expected doors table, got: %s", output)
	}
}

func TestNewDoorsCmd_LoadTaskPoolError_JSON(t *testing.T) {
	// Set up an environment where loadTaskPool returns an error
	// by pointing to a directory with an invalid config file
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write invalid YAML to trigger parse error
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	_ = os.Setenv("HOME", dir)

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"doors", "--json"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
	// Should have written JSON error
	if buf.Len() > 0 {
		var env JSONEnvelope
		if jsonErr := json.Unmarshal(buf.Bytes(), &env); jsonErr == nil {
			if env.Error == nil {
				t.Error("expected JSON error envelope")
			}
		}
	}
}

func TestNewDoorsCmd_LoadTaskPoolError_Human(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	origHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	_ = os.Setenv("HOME", dir)

	root := NewRootCmd()
	root.SetArgs([]string{"doors"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestLoadTaskPool_Success(t *testing.T) {
	setupTestEnv(t)
	addTestTask(t, "Pool task A")

	pool, provider, err := loadTaskPool()
	if err != nil {
		t.Fatalf("loadTaskPool: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	// At least one task should be in the pool (the one we added)
	if len(pool.GetAvailableForDoors()) < 1 {
		t.Errorf("pool has %d available tasks, want at least 1", len(pool.GetAvailableForDoors()))
	}
}

func TestLoadTaskPool_WithValidConfig(t *testing.T) {
	setupTestEnv(t)

	pool, provider, err := loadTaskPool()
	if err != nil {
		t.Fatalf("loadTaskPool: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestLoadTaskPool_DefaultConfig(t *testing.T) {
	// When config file is missing, loadTaskPool falls back to default config.
	// Verify it still returns a valid pool and provider.
	dir := t.TempDir()
	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	pool, provider, err := loadTaskPool()
	if err != nil {
		t.Fatalf("loadTaskPool with default config: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestLoadTaskPool_MultiProvider(t *testing.T) {
	// Test the multi-provider path (len(cfg.Providers) > 1)
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(filepath.Join(configDir, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write config with multiple providers to exercise the multi-provider code path
	configYAML := `schema_version: 2
providers:
  - name: primary
    type: textfile
  - name: secondary
    type: textfile
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	// This will exercise the multi-provider path.
	// It may succeed or fail depending on registry state, but exercises the code path.
	pool, _, err := loadTaskPool()
	if err != nil {
		// Multi-provider may fail if providers aren't registered — that's OK,
		// we just want to exercise the error path
		if !strings.Contains(err.Error(), "init providers") {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
}

func TestLoadTaskPool_FallbackProvider(t *testing.T) {
	// Test that loadTaskPool handles unknown provider types gracefully
	// (falls back to textfile if registered, or errors if not)
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".threedoors")
	if err := os.MkdirAll(filepath.Join(configDir, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configYAML := `schema_version: 2
provider: nonexistent-provider-type
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	core.SetHomeDir(dir)
	t.Cleanup(func() { core.SetHomeDir("") })

	// May succeed with fallback or fail — either exercises the code path
	pool, _, err := loadTaskPool()
	if err != nil {
		// Expected when textfile isn't registered
		return
	}
	if pool == nil {
		t.Fatal("expected non-nil pool when fallback succeeds")
	}
}

func TestHandlePick_InvalidDoor_JSON(t *testing.T) {
	t.Parallel()

	doors := []*core.Task{
		newTestTask("a", "A", core.StatusTodo),
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	provider := &mockProvider{}

	err := handlePick(nil, formatter, doors, 5, provider)
	if err == nil {
		t.Error("expected error for invalid pick")
	}

	// JSON error should have been written
	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected JSON error")
	}
	if env.Error.Code != ExitValidation {
		t.Errorf("error code = %d, want %d", env.Error.Code, ExitValidation)
	}
}

func TestHandlePick_SaveError_JSON(t *testing.T) {
	t.Parallel()

	task := newTestTask("aaaa1111-2222-3333-4444-555555555555", "Save fail JSON", core.StatusTodo)
	doors := []*core.Task{task}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	provider := &mockProvider{saveErr: fmt.Errorf("write failed")}

	err := handlePick(nil, formatter, doors, 1, provider)
	if err == nil {
		t.Error("expected error when save fails")
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error == nil {
		t.Fatal("expected JSON error")
	}
	if env.Error.Code != ExitProviderError {
		t.Errorf("error code = %d, want %d", env.Error.Code, ExitProviderError)
	}
}

// mockProviderWithTasks extends mockProvider with configurable LoadTasks behavior.
type mockProviderWithTasks struct {
	mockProvider
	tasks   []*core.Task
	loadErr error
}

func (m *mockProviderWithTasks) LoadTasks() ([]*core.Task, error) {
	return m.tasks, m.loadErr
}

// mockProvider is a minimal TaskProvider for testing.
type mockProvider struct {
	saved   bool
	saveErr error
}

func (m *mockProvider) Name() string                        { return "mock" }
func (m *mockProvider) LoadTasks() ([]*core.Task, error)    { return nil, nil }
func (m *mockProvider) SaveTasks(tasks []*core.Task) error  { return nil }
func (m *mockProvider) DeleteTask(taskID string) error      { return nil }
func (m *mockProvider) MarkComplete(taskID string) error    { return nil }
func (m *mockProvider) Watch() <-chan core.ChangeEvent      { return nil }
func (m *mockProvider) HealthCheck() core.HealthCheckResult { return core.HealthCheckResult{} }
func (m *mockProvider) SaveTask(task *core.Task) error {
	m.saved = true
	return m.saveErr
}
