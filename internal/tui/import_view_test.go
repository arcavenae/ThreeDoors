package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestImportView_NewImportView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		prefilledPath string
		wantStep      importStep
	}{
		{"empty path starts at path input", "", importStepPath},
		{"invalid prefilled path stays at path input", "/nonexistent/file.txt", importStepPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			iv := NewImportView(tt.prefilledPath)
			if iv.step != tt.wantStep {
				t.Errorf("step = %d, want %d", iv.step, tt.wantStep)
			}
		})
	}
}

func TestImportView_PrefilledPathValid(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "Buy milk\nWalk dog\nRead book\n")

	iv := NewImportView(path)

	if iv.step != importStepPreview {
		t.Fatalf("step = %d, want importStepPreview after valid prefilled path", iv.step)
	}
	if iv.importResult == nil {
		t.Fatal("expected importResult to be set")
	}
	if len(iv.importResult.Tasks) != 3 {
		t.Errorf("got %d tasks, want 3", len(iv.importResult.Tasks))
	}
}

func TestImportView_PathInput_EnterParses(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "Task one\nTask two\n")

	iv := NewImportView("")
	iv.textInput.SetValue(path)
	iv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if iv.step != importStepPreview {
		t.Fatalf("step = %d, want importStepPreview after enter", iv.step)
	}
	if iv.importResult == nil {
		t.Fatal("expected importResult to be set")
	}
	if len(iv.importResult.Tasks) != 2 {
		t.Errorf("got %d tasks, want 2", len(iv.importResult.Tasks))
	}
}

func TestImportView_PathInput_InvalidPath(t *testing.T) {
	t.Parallel()

	iv := NewImportView("")
	iv.SetWidth(80)
	iv.textInput.SetValue("/nonexistent/path/tasks.txt")
	iv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if iv.step != importStepPath {
		t.Errorf("step = %d, want importStepPath after invalid path", iv.step)
	}
	if iv.importError == "" {
		t.Error("expected importError to be set")
	}
}

func TestImportView_PathInput_EmptyFile(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "# just a comment\n")

	iv := NewImportView("")
	iv.textInput.SetValue(path)
	iv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if iv.step != importStepPath {
		t.Errorf("step = %d, want importStepPath after empty file", iv.step)
	}
	if iv.importError != "No tasks found in file" {
		t.Errorf("importError = %q, want 'No tasks found in file'", iv.importError)
	}
}

func TestImportView_EscapeCancels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step importStep
	}{
		{"escape from path input", importStepPath},
		{"escape from preview", importStepPreview},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			iv := NewImportView("")
			iv.step = tt.step
			if tt.step == importStepPreview {
				iv.importResult = &core.ImportResult{
					Tasks:      []*core.Task{core.NewTask("test")},
					SourcePath: "/tmp/test.txt",
					Format:     "text",
				}
			}

			cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEscape})
			if cmd == nil {
				t.Fatal("expected cmd, got nil")
			}
			msg := cmd()
			if _, ok := msg.(ReturnToDoorsMsg); !ok {
				t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
			}
		})
	}
}

func TestImportView_PreviewConfirm(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "Task A\nTask B\nTask C\n")

	iv := NewImportView(path)
	if iv.step != importStepPreview {
		t.Fatalf("step = %d, want importStepPreview", iv.step)
	}

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	confirmed, ok := msg.(ImportConfirmedMsg)
	if !ok {
		t.Fatalf("expected ImportConfirmedMsg, got %T", msg)
	}
	if len(confirmed.Tasks) != 3 {
		t.Errorf("got %d tasks, want 3", len(confirmed.Tasks))
	}
	if confirmed.Source == "" {
		t.Error("expected Source to be set")
	}
}

func TestImportView_PreviewConfirmY(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "Do stuff\n")

	iv := NewImportView(path)
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(ImportConfirmedMsg); !ok {
		t.Fatalf("expected ImportConfirmedMsg, got %T", msg)
	}
}

func TestImportView_PreviewRejectN(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "Some task\n")

	iv := NewImportView(path)
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestImportView_PreviewFiltersCompletedTasks(t *testing.T) {
	t.Parallel()

	path := writeTempTaskFile(t, "- [ ] Todo task\n- [x] Done task\n- [ ] Another todo\n")

	iv := NewImportView(path)
	if iv.step != importStepPreview {
		t.Fatalf("step = %d, want importStepPreview", iv.step)
	}

	// All 3 should be in importResult
	if len(iv.importResult.Tasks) != 3 {
		t.Fatalf("got %d tasks in result, want 3", len(iv.importResult.Tasks))
	}

	// Only TODO tasks should be in ImportConfirmedMsg
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	confirmed := msg.(ImportConfirmedMsg)
	if len(confirmed.Tasks) != 2 {
		t.Errorf("got %d confirmed tasks, want 2 (only todo)", len(confirmed.Tasks))
	}
}

func TestImportView_EmptyEnterCancels(t *testing.T) {
	t.Parallel()

	iv := NewImportView("")
	// Leave text input empty and press enter
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Fatalf("expected ReturnToDoorsMsg from empty enter, got %T", msg)
	}
}

func TestImportView_ViewRendering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		step         importStep
		importResult *core.ImportResult
		importError  string
		wantContains []string
	}{
		{
			name:         "path input view",
			step:         importStepPath,
			wantContains: []string{"Import Tasks", "plain text", "Markdown"},
		},
		{
			name:         "path input with error",
			step:         importStepPath,
			importError:  "file not found",
			wantContains: []string{"Import Tasks", "file not found"},
		},
		{
			name: "preview view",
			step: importStepPreview,
			importResult: &core.ImportResult{
				Tasks:      []*core.Task{core.NewTask("Buy milk"), core.NewTask("Walk dog")},
				SourcePath: "/tmp/tasks.txt",
				Format:     "text",
			},
			wantContains: []string{"Import Preview", "2 tasks", "text", "Buy milk", "Walk dog"},
		},
		{
			name: "preview with many tasks truncates",
			step: importStepPreview,
			importResult: &core.ImportResult{
				Tasks: []*core.Task{
					core.NewTask("T1"), core.NewTask("T2"), core.NewTask("T3"),
					core.NewTask("T4"), core.NewTask("T5"), core.NewTask("T6"),
					core.NewTask("T7"),
				},
				SourcePath: "/tmp/tasks.txt",
				Format:     "text",
			},
			wantContains: []string{"Import Preview", "7 tasks", "T1", "T5", "and 2 more"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			iv := NewImportView("")
			iv.SetWidth(80)
			iv.step = tt.step
			iv.importResult = tt.importResult
			iv.importError = tt.importError

			output := iv.View()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("View() missing %q", want)
				}
			}
		})
	}
}

func TestImportView_CommandRegistered(t *testing.T) {
	t.Parallel()

	matches := filterCommands("imp")
	found := false
	for _, cmd := range matches {
		if cmd.Name == "import" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'import' command not found in command registry autocomplete for prefix 'imp'")
	}
}

func TestImportView_CommandExecution(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	sv := NewSearchView(pool, nil, nil, nil, nil)

	tests := []struct {
		name     string
		input    string
		wantType string
		wantPath string
	}{
		{"import no args", ":import", "ShowImportMsg", ""},
		{"import with path", ":import ~/tasks.txt", "ShowImportMsg", "~/tasks.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svCopy := *sv
			svCopy.textInput.SetValue(tt.input)
			svCopy.isCommandMode = true
			cmd := svCopy.executeCommand()
			if cmd == nil {
				t.Fatal("expected cmd, got nil")
			}
			msg := cmd()
			switch m := msg.(type) {
			case ShowImportMsg:
				if tt.wantType != "ShowImportMsg" {
					t.Errorf("got ShowImportMsg, want %s", tt.wantType)
				}
				if m.PrefilledPath != tt.wantPath {
					t.Errorf("PrefilledPath = %q, want %q", m.PrefilledPath, tt.wantPath)
				}
			default:
				t.Errorf("got %T, want %s", msg, tt.wantType)
			}
		})
	}
}

func TestImportView_MarkdownImport(t *testing.T) {
	t.Parallel()

	content := "# Tasks\n- [ ] Buy groceries\n- [x] Send email\n- [ ] Clean house\n"
	path := writeTempTaskFile(t, content)

	// Rename to .md for format detection
	mdPath := path + ".md"
	if err := os.Rename(path, mdPath); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(mdPath) })

	iv := NewImportView(mdPath)
	if iv.step != importStepPreview {
		t.Fatalf("step = %d, want importStepPreview", iv.step)
	}
	if iv.importResult.Format != "markdown" {
		t.Errorf("format = %q, want 'markdown'", iv.importResult.Format)
	}
	if len(iv.importResult.Tasks) != 3 {
		t.Errorf("got %d tasks, want 3", len(iv.importResult.Tasks))
	}
}

// writeTempTaskFile creates a temporary file with the given content and returns its path.
func writeTempTaskFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
