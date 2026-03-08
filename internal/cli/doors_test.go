package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
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
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
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
	}

	wantSubstr := "load tasks"
	if !bytes.Contains([]byte(err.Error()), []byte(wantSubstr)) {
		t.Errorf("error = %q, want substring %q", err.Error(), wantSubstr)
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
