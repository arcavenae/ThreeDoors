package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestDailyNotePath(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		folder     string
		dateFormat string
		wantSuffix string
	}{
		{"default format", "", "", filepath.Join("2026-03-15.md")},
		{"custom folder", "Daily", "", filepath.Join("Daily", "2026-03-15.md")},
		{"custom format", "", "01-02-2006.md", "03-15-2026.md"},
		{"folder and format", "Notes/Daily", "2006-01-02.md", filepath.Join("Notes/Daily", "2026-03-15.md")},
		{"year-month subfolder format", "Daily", "2006/01/02.md", filepath.Join("Daily", "2026/03/15.md")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			adapter := NewObsidianAdapter(dir, "", "")
			adapter.SetDailyNotes(&DailyNotesConfig{
				Enabled:    true,
				Folder:     tt.folder,
				DateFormat: tt.dateFormat,
			})

			got, err := adapter.dailyNotePath(date)
			if err != nil {
				t.Fatalf("dailyNotePath() error: %v", err)
			}
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("dailyNotePath() = %q, want suffix %q", got, tt.wantSuffix)
			}
			if !strings.HasPrefix(got, dir) {
				t.Errorf("dailyNotePath() = %q, should start with vault path %q", got, dir)
			}
		})
	}
}

func TestDailyNotePath_DisabledReturnsError(t *testing.T) {
	t.Parallel()
	adapter := NewObsidianAdapter(t.TempDir(), "", "")
	_, err := adapter.dailyNotePath(time.Now().UTC())
	if err == nil {
		t.Error("expected error when daily notes not enabled")
	}
}

func TestLoadDailyNoteTasks_MissingFileReturnsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	tasks, err := adapter.loadDailyNoteTasks(time.Now().UTC())
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0 for missing daily note", len(tasks))
	}
}

func TestLoadDailyNoteTasks_ParsesTasks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	content := `# 2026-03-15

## Tasks

- [ ] Morning standup <!-- td:dn-1 -->
- [x] Code review <!-- td:dn-2 -->
- [/] Write docs <!-- td:dn-3 -->

## Notes

Some notes here.
`
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	tasks, err := adapter.loadDailyNoteTasks(date)
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	wantStatuses := []core.TaskStatus{core.StatusTodo, core.StatusComplete, core.StatusInProgress}
	wantIDs := []string{"dn-1", "dn-2", "dn-3"}
	for i, task := range tasks {
		if task.Status != wantStatuses[i] {
			t.Errorf("task %d status = %q, want %q", i, task.Status, wantStatuses[i])
		}
		if task.ID != wantIDs[i] {
			t.Errorf("task %d ID = %q, want %q", i, task.ID, wantIDs[i])
		}
	}
}

func TestLoadDailyNoteTasks_HeadingScopedReading(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	content := `# 2026-03-15

## Meetings

- [ ] Meeting task outside heading <!-- td:mt-1 -->

## Tasks

- [ ] Under heading task <!-- td:ht-1 -->
- [x] Another heading task <!-- td:ht-2 -->

## Notes

- [ ] Stray checkbox in notes <!-- td:nt-1 -->
`
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	tasks, err := adapter.loadDailyNoteTasks(date)
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2 (only under ## Tasks)", len(tasks))
	}
	if tasks[0].ID != "ht-1" {
		t.Errorf("task 0 ID = %q, want %q", tasks[0].ID, "ht-1")
	}
	if tasks[1].ID != "ht-2" {
		t.Errorf("task 1 ID = %q, want %q", tasks[1].ID, "ht-2")
	}
}

func TestLoadDailyNoteTasks_NoHeadingReturnsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	content := `# 2026-03-15

- [ ] Task with no heading section <!-- td:nh-1 -->
`
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	tasks, err := adapter.loadDailyNoteTasks(date)
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0 when heading not present", len(tasks))
	}
}

func TestLoadDailyNoteTasks_SubheadingIncluded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	// ### subheading (level 3) under ## Tasks (level 2) should be included
	content := `## Tasks

- [ ] Top level task <!-- td:tl-1 -->

### Subtasks

- [ ] Subtask under heading <!-- td:st-1 -->

## Notes

- [ ] Not included <!-- td:ni-1 -->
`
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	tasks, err := adapter.loadDailyNoteTasks(date)
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2 (includes subheading tasks)", len(tasks))
	}
	if tasks[0].ID != "tl-1" {
		t.Errorf("task 0 ID = %q, want %q", tasks[0].ID, "tl-1")
	}
	if tasks[1].ID != "st-1" {
		t.Errorf("task 1 ID = %q, want %q", tasks[1].ID, "st-1")
	}
}

func TestLoadDailyNoteTasks_WithFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dailyDir := filepath.Join(dir, "Daily")
	if err := os.MkdirAll(dailyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	content := "## Tasks\n\n- [ ] Daily task <!-- td:df-1 -->\n"
	if err := os.WriteFile(filepath.Join(dailyDir, "2026-01-05.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Folder:  "Daily",
	})

	tasks, err := adapter.loadDailyNoteTasks(date)
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "df-1" {
		t.Errorf("task ID = %q, want %q", tasks[0].ID, "df-1")
	}
}

func TestAppendTaskToDailyNote_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	task := core.NewTask("New daily task")
	if err := adapter.appendTaskToDailyNote(task, date); err != nil {
		t.Fatalf("appendTaskToDailyNote() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "2026-03-15.md"))
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	if !strings.Contains(got, "## Tasks") {
		t.Error("file should contain heading")
	}
	if !strings.Contains(got, "New daily task") {
		t.Error("file should contain task text")
	}
	if !strings.Contains(got, "<!-- td:") {
		t.Error("file should contain embedded task ID")
	}
}

func TestAppendTaskToDailyNote_AppendsUnderHeading(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	existing := `# 2026-03-15

## Tasks

- [ ] Existing task <!-- td:ex-1 -->

## Notes

Some notes.
`
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	task := core.NewTask("Appended task")
	if err := adapter.appendTaskToDailyNote(task, date); err != nil {
		t.Fatalf("appendTaskToDailyNote() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "2026-03-15.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)

	// Should contain both tasks
	if !strings.Contains(got, "Existing task") {
		t.Error("should preserve existing task")
	}
	if !strings.Contains(got, "Appended task") {
		t.Error("should contain appended task")
	}

	// Notes section should still be present
	if !strings.Contains(got, "## Notes") {
		t.Error("should preserve other sections")
	}
	if !strings.Contains(got, "Some notes.") {
		t.Error("should preserve notes content")
	}

	// Appended task should appear before ## Notes
	taskIdx := strings.Index(got, "Appended task")
	notesIdx := strings.Index(got, "## Notes")
	if taskIdx > notesIdx {
		t.Error("appended task should appear before ## Notes section")
	}
}

func TestAppendTaskToDailyNote_NoHeadingInFile(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	existing := "# 2026-03-15\n\nSome notes here.\n"
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	task := core.NewTask("New task")
	if err := adapter.appendTaskToDailyNote(task, date); err != nil {
		t.Fatalf("appendTaskToDailyNote() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "2026-03-15.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)

	if !strings.Contains(got, "## Tasks") {
		t.Error("should add heading when not present")
	}
	if !strings.Contains(got, "New task") {
		t.Error("should add task")
	}
	if !strings.Contains(got, "Some notes here.") {
		t.Error("should preserve existing content")
	}
}

func TestAppendTaskToDailyNote_CreatesDirAndFile(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Folder:  "Daily",
	})

	task := core.NewTask("Task in new dir")
	if err := adapter.appendTaskToDailyNote(task, date); err != nil {
		t.Fatalf("appendTaskToDailyNote() error: %v", err)
	}

	path := filepath.Join(dir, "Daily", "2026-03-15.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("daily note file should be created")
	}
}

func TestLoadTasks_IncludesDailyNotes(t *testing.T) {
	dir := t.TempDir()

	// Create regular tasks file
	tasksContent := "- [ ] Vault task <!-- td:vt-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(tasksContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create today's daily note with heading
	today := time.Now().UTC()
	dailyFilename := today.Format("2006-01-02.md")
	dailyContent := "## Tasks\n\n- [ ] Daily task <!-- td:dt-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, dailyFilename), []byte(dailyContent), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	foundVault := false
	foundDaily := false
	for _, task := range tasks {
		if task.ID == "vt-1" {
			foundVault = true
		}
		if task.ID == "dt-1" {
			foundDaily = true
		}
	}

	if !foundVault {
		t.Error("should include vault tasks")
	}
	if !foundDaily {
		t.Error("should include daily note tasks")
	}
}

func TestLoadTasks_DeduplicatesDailyNotes(t *testing.T) {
	dir := t.TempDir()

	// Same task ID in both vault and daily note (daily note is in vault dir)
	today := time.Now().UTC()
	dailyFilename := today.Format("2006-01-02.md")
	content := "## Tasks\n\n- [ ] Duplicate task <!-- td:dup-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, dailyFilename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	tasks, err := adapter.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	// When daily note is in the same folder as vault, the task appears in vault load.
	// Dedup should ensure it only appears once.
	count := 0
	for _, task := range tasks {
		if task.ID == "dup-1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("duplicate task appeared %d times, want 1", count)
	}
}

func TestSaveTask_RoutesToDailyNote(t *testing.T) {
	dir := t.TempDir()

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	task := core.NewTask("Quick add task")
	if err := adapter.SaveTask(task); err != nil {
		t.Fatalf("SaveTask() error: %v", err)
	}

	// Task should be in today's daily note, not tasks.md
	today := time.Now().UTC()
	dailyFilename := today.Format("2006-01-02.md")
	data, err := os.ReadFile(filepath.Join(dir, dailyFilename))
	if err != nil {
		t.Fatalf("daily note should exist: %v", err)
	}
	if !strings.Contains(string(data), "Quick add task") {
		t.Error("task should be in daily note")
	}

	// tasks.md should not exist
	if _, err := os.Stat(filepath.Join(dir, "tasks.md")); !os.IsNotExist(err) {
		t.Error("tasks.md should not be created when daily notes enabled")
	}
}

func TestSanitizeDailyNotePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"normal date", "2026-03-15.md", "2026-03-15.md", false},
		{"with spaces", "2026 03 15.md", "2026 03 15.md", false},
		{"null byte", "2026\x00-03-15.md", "", true},
		{"path traversal dots", "../evil.md", "", true},
		{"subdirectory allowed", "2026/03/15.md", "2026/03/15.md", false},
		{"just dot", ".", "", true},
		{"double dot", "..", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := sanitizeDailyNotePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitizeDailyNotePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("sanitizeDailyNotePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// AC-Q6: Input sanitization tests for special characters in date formats and heading names.
func TestDailyNotes_InputSanitization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		heading    string
		dateFormat string
		taskText   string
	}{
		{
			name:       "unicode heading",
			heading:    "## 任务列表",
			dateFormat: "2006-01-02.md",
			taskText:   "Normal task",
		},
		{
			name:       "heading with special chars",
			heading:    "## Tasks & Notes (Daily)",
			dateFormat: "2006-01-02.md",
			taskText:   "Task with 'quotes' & \"doubles\"",
		},
		{
			name:       "emoji heading",
			heading:    "## 📋 Tasks",
			dateFormat: "2006-01-02.md",
			taskText:   "🚀 Launch prep",
		},
		{
			name:       "heading with angle brackets",
			heading:    "## Tasks <important>",
			dateFormat: "2006-01-02.md",
			taskText:   "Task with <html> entities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

			adapter := NewObsidianAdapter(dir, "", "")
			adapter.SetDailyNotes(&DailyNotesConfig{
				Enabled:    true,
				Heading:    tt.heading,
				DateFormat: tt.dateFormat,
			})

			task := core.NewTask(tt.taskText)
			if err := adapter.appendTaskToDailyNote(task, date); err != nil {
				t.Fatalf("appendTaskToDailyNote() error: %v", err)
			}

			// Verify round-trip: load tasks back
			tasks, err := adapter.loadDailyNoteTasks(date)
			if err != nil {
				t.Fatalf("loadDailyNoteTasks() error: %v", err)
			}
			if len(tasks) == 0 {
				t.Fatal("expected at least one task after round-trip")
			}
			if tasks[0].ID != task.ID {
				t.Errorf("ID mismatch: got %q, want %q", tasks[0].ID, task.ID)
			}
		})
	}
}

func TestDailyNotes_DateFormatSanitization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dateFormat string
		wantErr    bool
	}{
		{"standard YYYY-MM-DD", "2006-01-02.md", false},
		{"US format MM-DD-YYYY", "01-02-2006.md", false},
		{"dot separator", "2006.01.02.md", false},
		{"underscore separator", "2006_01_02.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

			adapter := NewObsidianAdapter(dir, "", "")
			adapter.SetDailyNotes(&DailyNotesConfig{
				Enabled:    true,
				DateFormat: tt.dateFormat,
			})

			_, err := adapter.dailyNotePath(date)
			if (err != nil) != tt.wantErr {
				t.Errorf("dailyNotePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppendTaskToDailyNote_DefaultHeading(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		// Heading left empty — should use default
	})

	task := core.NewTask("Default heading task")
	if err := adapter.appendTaskToDailyNote(task, date); err != nil {
		t.Fatalf("appendTaskToDailyNote() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "2026-03-15.md"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), defaultDailyNotesHeading) {
		t.Errorf("should use default heading %q", defaultDailyNotesHeading)
	}
}

// Story 8.4: Common date format auto-detection tests

func TestResolveDailyNotePath_AutoDetectsFormat(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		configFormat   string
		actualFilename string
	}{
		{"US format detected", "2006-01-02.md", "03-15-2026.md"},
		{"dot separator detected", "2006-01-02.md", "2026.03.15.md"},
		{"underscore detected", "2006-01-02.md", "2026_03_15.md"},
		{"compact detected", "2006-01-02.md", "20260315.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			// Create file with non-configured format
			content := "## Tasks\n\n- [ ] Auto-detected task <!-- td:ad-1 -->\n"
			if err := os.WriteFile(filepath.Join(dir, tt.actualFilename), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}

			adapter := NewObsidianAdapter(dir, "", "")
			adapter.SetDailyNotes(&DailyNotesConfig{
				Enabled:    true,
				DateFormat: tt.configFormat,
			})

			path, err := adapter.resolveDailyNotePath(date)
			if err != nil {
				t.Fatalf("resolveDailyNotePath() error: %v", err)
			}

			if filepath.Base(path) != tt.actualFilename {
				t.Errorf("resolved to %q, want filename %q", filepath.Base(path), tt.actualFilename)
			}
		})
	}
}

func TestResolveDailyNotePath_PrefersConfiguredFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	// Create files in both configured and auto-detected formats
	content := "## Tasks\n\n- [ ] Task <!-- td:t-1 -->\n"
	if err := os.WriteFile(filepath.Join(dir, "03-15-2026.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "2026-03-15.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled:    true,
		DateFormat: "2006-01-02.md",
	})

	path, err := adapter.resolveDailyNotePath(date)
	if err != nil {
		t.Fatalf("resolveDailyNotePath() error: %v", err)
	}

	if filepath.Base(path) != "2026-03-15.md" {
		t.Errorf("should prefer configured format, got %q", filepath.Base(path))
	}
}

func TestResolveDailyNotePath_WithFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	// Create daily note in folder with non-default format
	dailyDir := filepath.Join(dir, "Daily")
	if err := os.MkdirAll(dailyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "## Tasks\n\n- [ ] Task <!-- td:t-1 -->\n"
	if err := os.WriteFile(filepath.Join(dailyDir, "03-15-2026.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled:    true,
		Folder:     "Daily",
		DateFormat: "2006-01-02.md",
	})

	path, err := adapter.resolveDailyNotePath(date)
	if err != nil {
		t.Fatalf("resolveDailyNotePath() error: %v", err)
	}

	if filepath.Base(path) != "03-15-2026.md" {
		t.Errorf("should auto-detect format in folder, got %q", filepath.Base(path))
	}
}

func TestResolveDailyNotePath_NoFileReturnsConfiguredPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled:    true,
		DateFormat: "2006-01-02.md",
	})

	path, err := adapter.resolveDailyNotePath(date)
	if err != nil {
		t.Fatalf("resolveDailyNotePath() error: %v", err)
	}

	if filepath.Base(path) != "2026-03-15.md" {
		t.Errorf("should return configured format path, got %q", filepath.Base(path))
	}
}

// Story 8.4: ValidateDateFormat tests

func TestValidateDateFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"standard ISO", "2006-01-02.md", false},
		{"US format", "01-02-2006.md", false},
		{"dot separator", "2006.01.02.md", false},
		{"underscore", "2006_01_02.md", false},
		{"compact", "20060102.md", false},
		{"subdirectory", "2006/01/02.md", false},
		{"long month", "Jan 2, 2006.md", false},
		{"empty string", "", true},
		{"no .md extension", "2006-01-02", true},
		{"no .md extension txt", "2006-01-02.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDateFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateFormat(%q) error = %v, wantErr %v", tt.format, err, tt.wantErr)
			}
		})
	}
}

// Story 8.4: QuickAddToDailyNote tests

func TestQuickAddToDailyNote(t *testing.T) {
	dir := t.TempDir()

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	task, err := adapter.QuickAddToDailyNote("Buy groceries")
	if err != nil {
		t.Fatalf("QuickAddToDailyNote() error: %v", err)
	}

	if task == nil {
		t.Fatal("expected non-nil task")
	}
	if task.Text != "Buy groceries" {
		t.Errorf("task text = %q, want %q", task.Text, "Buy groceries")
	}
	if task.ID == "" {
		t.Error("task should have an ID")
	}

	// Verify it was written to today's daily note
	today := time.Now().UTC()
	dailyFilename := today.Format("2006-01-02.md")
	data, err := os.ReadFile(filepath.Join(dir, dailyFilename))
	if err != nil {
		t.Fatalf("daily note should exist: %v", err)
	}
	if !strings.Contains(string(data), "Buy groceries") {
		t.Error("daily note should contain task")
	}
	if !strings.Contains(string(data), "## Tasks") {
		t.Error("daily note should contain heading")
	}
}

func TestQuickAddToDailyNote_DisabledReturnsError(t *testing.T) {
	t.Parallel()
	adapter := NewObsidianAdapter(t.TempDir(), "", "")
	_, err := adapter.QuickAddToDailyNote("some task")
	if err == nil {
		t.Error("expected error when daily notes not enabled")
	}
}

func TestQuickAddToDailyNote_EmptyTextReturnsError(t *testing.T) {
	adapter := NewObsidianAdapter(t.TempDir(), "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	_, err := adapter.QuickAddToDailyNote("")
	if err == nil {
		t.Error("expected error for empty text")
	}
	_, err = adapter.QuickAddToDailyNote("   ")
	if err == nil {
		t.Error("expected error for whitespace-only text")
	}
}

func TestQuickAddToDailyNote_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	adapter := NewObsidianAdapter(dir, "", "")
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
		Heading: "## Tasks",
	})

	// Quick-add three tasks
	task1, err := adapter.QuickAddToDailyNote("First task")
	if err != nil {
		t.Fatalf("QuickAddToDailyNote() error: %v", err)
	}
	task2, err := adapter.QuickAddToDailyNote("Second task")
	if err != nil {
		t.Fatalf("QuickAddToDailyNote() error: %v", err)
	}

	// Load them back
	tasks, err := adapter.loadDailyNoteTasks(time.Now().UTC())
	if err != nil {
		t.Fatalf("loadDailyNoteTasks() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}

	if tasks[0].ID != task1.ID {
		t.Errorf("task 0 ID = %q, want %q", tasks[0].ID, task1.ID)
	}
	if tasks[1].ID != task2.ID {
		t.Errorf("task 1 ID = %q, want %q", tasks[1].ID, task2.ID)
	}
}

// Story 8.4: headingMarkdownLevel tests

func TestHeadingMarkdownLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		line  string
		level int
	}{
		{"h1", "# Title", 1},
		{"h2", "## Subtitle", 2},
		{"h3", "### Sub-subtitle", 3},
		{"h4", "#### Level 4", 4},
		{"h5", "##### Level 5", 5},
		{"h6", "###### Level 6", 6},
		{"not heading no space", "#NoSpace", 0},
		{"not heading empty", "", 0},
		{"not heading text", "Just text", 0},
		{"too many hashes", "####### Seven", 0},
		{"hash only", "#", 0},
		{"with leading whitespace", "  ## Tasks", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := headingMarkdownLevel(tt.line)
			if got != tt.level {
				t.Errorf("headingMarkdownLevel(%q) = %d, want %d", tt.line, got, tt.level)
			}
		})
	}
}

// AC-Q6: Additional input sanitization for daily note paths and headings

func TestDailyNotes_PathTraversalInDateFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	adapter := NewObsidianAdapter(dir, "", "")

	// Attempt to use a date format that would produce path traversal
	adapter.SetDailyNotes(&DailyNotesConfig{
		Enabled: true,
	})

	// sanitizeDailyNotePath catches path traversal attempts
	_, err := sanitizeDailyNotePath("../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal in daily note path")
	}

	_, err = sanitizeDailyNotePath("valid/2026-03-15.md")
	if err != nil {
		t.Errorf("should allow valid subdirectory path: %v", err)
	}
}
