package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestObsidianAdapter_ImplementsTaskProvider(t *testing.T) {
	var _ core.TaskProvider = (*ObsidianAdapter)(nil)
}

func TestParseCheckboxLineObsidian(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantText   string
		wantStatus core.TaskStatus
		wantID     string
		wantCheck  bool
	}{
		{"todo dash", "- [ ] Buy groceries", "Buy groceries", core.StatusTodo, "", true},
		{"complete dash lowercase", "- [x] Done task", "Done task", core.StatusComplete, "", true},
		{"complete dash uppercase", "- [X] Done task", "Done task", core.StatusComplete, "", true},
		{"in-progress dash", "- [/] Working on it", "Working on it", core.StatusInProgress, "", true},
		{"todo star", "* [ ] Star todo", "Star todo", core.StatusTodo, "", true},
		{"complete star", "* [x] Star done", "Star done", core.StatusComplete, "", true},
		{"in-progress star", "* [/] Star wip", "Star wip", core.StatusInProgress, "", true},
		{"indented todo", "  - [ ] Indented task", "Indented task", core.StatusTodo, "", true},
		{"tab indented", "\t- [ ] Tab indented", "Tab indented", core.StatusTodo, "", true},
		{"not checkbox", "regular text", "regular text", core.StatusTodo, "", false},
		{"heading", "# My Heading", "# My Heading", core.StatusTodo, "", false},
		{"empty checkbox", "- [ ] ", "", core.StatusTodo, "", true},
		{"dash only", "- text", "- text", core.StatusTodo, "", false},
		{"with embedded id", "- [ ] My task <!-- td:abc-123 -->", "My task", core.StatusTodo, "abc-123", true},
		{"complete with id", "- [x] Done <!-- td:def-456 -->", "Done", core.StatusComplete, "def-456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, status, id, isCheckbox := parseCheckboxLineObsidian(tt.line)
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
			if isCheckbox != tt.wantCheck {
				t.Errorf("isCheckbox = %v, want %v", isCheckbox, tt.wantCheck)
			}
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantText   string
		wantTags   int
		wantDue    string
		wantEffort core.TaskEffort
	}{
		{"plain text", "Buy groceries", "Buy groceries", 0, "", ""},
		{"with due date", "Buy groceries 📅 2026-03-15", "Buy groceries", 0, "2026-03-15", ""},
		{"with tag", "Buy groceries #shopping", "Buy groceries #shopping", 1, "", ""},
		{"with high priority", "Important task ⏫", "Important task", 0, "", core.EffortDeepWork},
		{"with medium priority", "Normal task 🔼", "Normal task", 0, "", core.EffortMedium},
		{"with low priority", "Easy task 🔽", "Easy task", 0, "", core.EffortQuickWin},
		{"mixed metadata", "Task #work 📅 2026-01-01 ⏫", "Task #work", 1, "2026-01-01", core.EffortDeepWork},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleaned, tags, dueDate, effort := extractMetadata(tt.text)
			if cleaned != tt.wantText {
				t.Errorf("cleaned = %q, want %q", cleaned, tt.wantText)
			}
			if len(tags) != tt.wantTags {
				t.Errorf("tags count = %d, want %d", len(tags), tt.wantTags)
			}
			if dueDate != tt.wantDue {
				t.Errorf("dueDate = %q, want %q", dueDate, tt.wantDue)
			}
			if effort != tt.wantEffort {
				t.Errorf("effort = %q, want %q", effort, tt.wantEffort)
			}
		})
	}
}

func TestObsidianAdapter_LoadTasks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := `# My Tasks

- [ ] Buy groceries <!-- td:id-1 -->
- [x] Write tests <!-- td:id-2 -->
- [/] Review PR <!-- td:id-3 -->
Some regular text
- [ ] Read a book <!-- td:id-4 -->
`
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 4 {
		t.Fatalf("got %d tasks, want 4", len(tasks))
	}

	wantStatuses := []core.TaskStatus{core.StatusTodo, core.StatusComplete, core.StatusInProgress, core.StatusTodo}
	wantIDs := []string{"id-1", "id-2", "id-3", "id-4"}
	for i, task := range tasks {
		if task.Status != wantStatuses[i] {
			t.Errorf("task %d status = %q, want %q", i, task.Status, wantStatuses[i])
		}
		if task.ID != wantIDs[i] {
			t.Errorf("task %d ID = %q, want %q", i, task.ID, wantIDs[i])
		}
	}
}

func TestObsidianAdapter_LoadTasks_NoEmbeddedID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "- [ ] Task without ID\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID == "" {
		t.Error("task should get a generated UUID when no embedded ID exists")
	}
}

func TestObsidianAdapter_LoadTasks_EmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks from empty dir, want 0", len(tasks))
	}
}

func TestObsidianAdapter_LoadTasks_NonExistentDir(t *testing.T) {
	t.Parallel()

	adapter := NewObsidianAdapter("/nonexistent/path", "", "")
	_, err := adapter.LoadTasks()
	if err == nil {
		t.Error("LoadTasks() should return error for nonexistent dir")
	}
}

func TestObsidianAdapter_LoadTasks_SubFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "tasks")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "- [ ] Subfolder task <!-- td:sub-1 -->\n"
	if err := os.WriteFile(filepath.Join(subDir, "todo.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "tasks", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].Text != "Subfolder task" {
		t.Errorf("task text = %q, want %q", tasks[0].Text, "Subfolder task")
	}
}

func TestObsidianAdapter_LoadTasks_MultipleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("- [ ] Task A <!-- td:a1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("- [ ] Task B <!-- td:b1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
}

func TestObsidianAdapter_LoadTasks_WithMetadata(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := "- [ ] Important task ⏫ #work 📅 2026-03-15 <!-- td:meta-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}

	task := tasks[0]
	if task.Effort != core.EffortDeepWork {
		t.Errorf("effort = %q, want %q", task.Effort, core.EffortDeepWork)
	}
	if task.Context != "work" {
		t.Errorf("context = %q, want %q", task.Context, "work")
	}
}

func TestObsidianAdapter_SaveTask_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	content := "# Tasks\n- [ ] Original task <!-- td:upd-1 -->\n"
	filePath := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}

	tasks[0].Text = "Updated task"
	if err := adapter.SaveTask(tasks[0]); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if got != "# Tasks\n- [ ] Updated task <!-- td:upd-1 -->\n" {
		t.Errorf("file content = %q", got)
	}
}

func TestObsidianAdapter_SaveTask_AppendNew(t *testing.T) {
	dir := t.TempDir()
	content := "- [ ] Existing task <!-- td:exist-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	newTask := core.NewTask("Brand new task")
	if err := adapter.SaveTask(newTask); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	foundExisting := false
	foundNew := false
	for _, task := range tasks {
		if task.Text == "Existing task" {
			foundExisting = true
		}
		if task.Text == "Brand new task" {
			foundNew = true
		}
	}
	if !foundExisting {
		t.Error("existing task not found after save")
	}
	if !foundNew {
		t.Error("new task not found after save")
	}
}

func TestObsidianAdapter_SaveTask_CreateFile(t *testing.T) {
	dir := t.TempDir()

	adapter := NewObsidianAdapter(dir, "", "")
	newTask := core.NewTask("First task")
	if err := adapter.SaveTask(newTask); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "tasks.md"))
	if err != nil {
		t.Fatalf("read tasks.md: %v", err)
	}
	if got := string(data); got == "" {
		t.Error("tasks.md is empty after save")
	}
}

func TestObsidianAdapter_DeleteTask(t *testing.T) {
	dir := t.TempDir()
	content := "- [ ] Keep this <!-- td:keep-1 -->\n- [ ] Delete this <!-- td:del-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	if err := adapter.DeleteTask("del-1"); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() after delete error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks after delete, want 1", len(tasks))
	}
	if tasks[0].Text != "Keep this" {
		t.Errorf("remaining task = %q, want %q", tasks[0].Text, "Keep this")
	}
}

func TestObsidianAdapter_DeleteTask_NonExistent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] Task <!-- td:t1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	err := adapter.DeleteTask("nonexistent-id")
	if err != nil {
		t.Errorf("DeleteTask() for nonexistent ID returned error: %v", err)
	}
}

func TestObsidianAdapter_MarkComplete(t *testing.T) {
	dir := t.TempDir()
	content := "- [ ] Complete me <!-- td:comp-1 -->\n"
	filePath := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	if err := adapter.MarkComplete("comp-1"); err != nil {
		t.Fatalf("MarkComplete() error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "- [x] Complete me <!-- td:comp-1 -->\n" {
		t.Errorf("file content = %q", got)
	}
}

func TestObsidianAdapter_MarkComplete_NonExistent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] Task <!-- td:t1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	err := adapter.MarkComplete("nonexistent-id")
	if err == nil {
		t.Error("MarkComplete() on nonexistent task should return error")
	}
}

func TestObsidianAdapter_MarkComplete_AlreadyComplete(t *testing.T) {
	dir := t.TempDir()
	content := "- [x] Already done <!-- td:done-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	err := adapter.MarkComplete("done-1")
	if err == nil {
		t.Error("MarkComplete() on already complete task should return error")
	}
}

func TestObsidianAdapter_SaveTasks_Batch(t *testing.T) {
	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	batch := make([]*core.Task, 5)
	for i := range batch {
		batch[i] = core.NewTask("Batch task")
	}

	if err := adapter.SaveTasks(batch); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	loaded, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(loaded) != 5 {
		t.Errorf("got %d tasks, want 5", len(loaded))
	}
}

func TestObsidianAdapter_SaveTasks_Empty(t *testing.T) {
	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	if err := adapter.SaveTasks([]*core.Task{}); err != nil {
		t.Fatalf("SaveTasks() with empty list error: %v", err)
	}
}

func TestObsidianAdapter_PreservesNonCheckboxContent(t *testing.T) {
	dir := t.TempDir()
	content := "# My Project\n\nSome notes here.\n\n- [ ] A task <!-- td:pres-1 -->\n\n## Section 2\n"
	filePath := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}

	tasks[0].Text = "Updated task"
	if err := adapter.SaveTask(tasks[0]); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	want := "# My Project\n\nSome notes here.\n\n- [ ] Updated task <!-- td:pres-1 -->\n\n## Section 2\n"
	if got != want {
		t.Errorf("non-checkbox content not preserved.\ngot:  %q\nwant: %q", got, want)
	}
}

func TestObsidianAdapter_RoundTrip_IDPreservation(t *testing.T) {
	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	original := []*core.Task{
		core.NewTask("Task Alpha"),
		core.NewTask("Task Beta"),
	}

	if err := adapter.SaveTasks(original); err != nil {
		t.Fatalf("SaveTasks() error: %v", err)
	}

	loaded, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(loaded) != len(original) {
		t.Fatalf("got %d tasks, want %d", len(loaded), len(original))
	}

	tasksByID := make(map[string]*core.Task, len(loaded))
	for _, task := range loaded {
		tasksByID[task.ID] = task
	}

	for _, orig := range original {
		lt, ok := tasksByID[orig.ID]
		if !ok {
			t.Errorf("task %q not found after save/load", orig.ID)
			continue
		}
		if lt.Text != orig.Text {
			t.Errorf("task %q: Text = %q, want %q", orig.ID, lt.Text, orig.Text)
		}
	}
}

// Story 8.3: Vault Configuration Tests

func TestValidateVaultPath_ValidDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := ValidateVaultPath(dir); err != nil {
		t.Errorf("ValidateVaultPath(%q) unexpected error: %v", dir, err)
	}
}

func TestValidateVaultPath_NonExistent(t *testing.T) {
	t.Parallel()
	err := ValidateVaultPath("/nonexistent/vault/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
		return
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

func TestValidateVaultPath_NotADirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notadir.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ValidateVaultPath(filePath)
	if err == nil {
		t.Fatal("expected error for file path")
		return
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' in error, got: %v", err)
	}
}

func TestValidateVaultPath_NotWritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping: running as root")
	}
	t.Parallel()
	dir := t.TempDir()
	readonlyDir := filepath.Join(dir, "readonly")
	if err := os.Mkdir(readonlyDir, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(readonlyDir, 0o755)
	})
	err := ValidateVaultPath(readonlyDir)
	if err == nil {
		t.Fatal("expected error for read-only directory")
		return
	}
	if !strings.Contains(err.Error(), "not writable") {
		t.Errorf("expected 'not writable' in error, got: %v", err)
	}
}

func TestObsidianAdapter_FilePattern(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create files with different extensions
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] Task A <!-- td:a1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("- [ ] Task B <!-- td:b2 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "project.md"), []byte("- [ ] Task C <!-- td:c3 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		pattern     string
		wantTaskIDs []string
	}{
		{"default pattern matches all md", "", []string{"a1", "c3"}},
		{"specific file pattern", "tasks.md", []string{"a1"}},
		{"star md pattern", "*.md", []string{"a1", "c3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			adapter := NewObsidianAdapter(dir, "", tt.pattern)
			tasks, err := adapter.LoadTasks()
			if err != nil {
				t.Fatalf("LoadTasks() error: %v", err)
			}
			gotIDs := make(map[string]bool)
			for _, task := range tasks {
				gotIDs[task.ID] = true
			}
			for _, wantID := range tt.wantTaskIDs {
				if !gotIDs[wantID] {
					t.Errorf("expected task ID %q not found", wantID)
				}
			}
			if len(tasks) != len(tt.wantTaskIDs) {
				t.Errorf("got %d tasks, want %d", len(tasks), len(tt.wantTaskIDs))
			}
		})
	}
}

func TestObsidianAdapter_FilePatternSubfolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subdir := filepath.Join(dir, "tasks")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Only task files in subfolder match
	if err := os.WriteFile(filepath.Join(subdir, "todo.md"), []byte("- [ ] Sub task <!-- td:s1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.md"), []byte("- [ ] Root task <!-- td:r1 -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "tasks", "*.md")
	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task from subfolder, got %d", len(tasks))
	}
	if tasks[0].ID != "s1" {
		t.Errorf("expected task ID 's1', got %q", tasks[0].ID)
	}
}

func TestObsidianAdapter_DefaultFilePattern(t *testing.T) {
	t.Parallel()

	adapter := NewObsidianAdapter("/tmp", "", "")
	if adapter.filePattern != "*.md" {
		t.Errorf("expected default file pattern '*.md', got %q", adapter.filePattern)
	}
}

func TestTaskToObsidianLine(t *testing.T) {
	tests := []struct {
		name   string
		status core.TaskStatus
		text   string
		id     string
		want   string
	}{
		{"todo", core.StatusTodo, "Buy groceries", "abc", "- [ ] Buy groceries <!-- td:abc -->"},
		{"complete", core.StatusComplete, "Done task", "def", "- [x] Done task <!-- td:def -->"},
		{"in-progress", core.StatusInProgress, "Working", "ghi", "- [/] Working <!-- td:ghi -->"},
		{"blocked maps to todo", core.StatusBlocked, "Blocked", "jkl", "- [ ] Blocked <!-- td:jkl -->"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &core.Task{ID: tt.id, Status: tt.status, Text: tt.text}
			got := taskToObsidianLine(task)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// AC-Q6: Input sanitization tests for special characters in filenames and task text.
func TestObsidianAdapter_InputSanitization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "special chars in filename",
			filename: "tasks (copy).md",
			content:  "- [ ] Normal task <!-- td:san-1 -->\n",
		},
		{
			name:     "unicode in filename",
			filename: "задачи.md",
			content:  "- [ ] Unicode filename task <!-- td:san-2 -->\n",
		},
		{
			name:     "quotes and HTML in task text",
			filename: "test.md",
			content:  "- [ ] Task with 'quotes' & \"double\" <html> <!-- td:san-3 -->\n",
		},
		{
			name:     "emoji in task text",
			filename: "test.md",
			content:  "- [ ] 🚀 Launch the rocket 🎯 <!-- td:san-4 -->\n",
		},
		{
			name:     "newline-like content in task",
			filename: "test.md",
			content:  "- [ ] Task with \\n escape chars <!-- td:san-5 -->\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, tt.filename), []byte(tt.content), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}

			adapter := NewObsidianAdapter(dir, "", "")
			loaded, err := adapter.LoadTasks()
			if err != nil {
				t.Fatalf("LoadTasks() error: %v", err)
			}
			if len(loaded) == 0 {
				t.Fatal("expected at least one task")
			}

			// Verify round-trip: save and reload
			if err := adapter.SaveTask(loaded[0]); err != nil {
				t.Fatalf("SaveTask() round-trip error: %v", err)
			}
			reloaded, err := adapter.LoadTasks()
			if err != nil {
				t.Fatalf("LoadTasks() after save error: %v", err)
			}
			if len(reloaded) == 0 {
				t.Fatal("no tasks after round-trip")
			}
			if reloaded[0].ID != loaded[0].ID {
				t.Errorf("ID changed after round-trip: got %q, want %q", reloaded[0].ID, loaded[0].ID)
			}
		})
	}
}
