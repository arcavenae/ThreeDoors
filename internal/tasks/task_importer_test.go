package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		content string
		want    string
	}{
		{
			name:    "markdown extension",
			path:    "tasks.md",
			content: "some content",
			want:    "markdown",
		},
		{
			name:    "markdown extension uppercase",
			path:    "tasks.MARKDOWN",
			content: "some content",
			want:    "markdown",
		},
		{
			name:    "text extension",
			path:    "tasks.txt",
			content: "some content",
			want:    "text",
		},
		{
			name:    "no extension with checkbox content",
			path:    "tasks",
			content: "- [ ] Buy groceries\n- [x] Walk the dog",
			want:    "markdown",
		},
		{
			name:    "no extension plain text",
			path:    "tasks",
			content: "Buy groceries\nWalk the dog",
			want:    "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := detectFormat(tt.path, tt.content)
			if got != tt.want {
				t.Errorf("detectFormat(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestParseTextTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		wantCount int
		wantTexts []string
	}{
		{
			name:      "simple lines",
			content:   "Buy groceries\nWalk the dog\nWrite code",
			wantCount: 3,
			wantTexts: []string{"Buy groceries", "Walk the dog", "Write code"},
		},
		{
			name:      "with list markers",
			content:   "- Buy groceries\n* Walk the dog\n1. Write code\n2) Read a book",
			wantCount: 4,
			wantTexts: []string{"Buy groceries", "Walk the dog", "Write code", "Read a book"},
		},
		{
			name:      "skips empty lines and comments",
			content:   "Buy groceries\n\n# Shopping list\nWalk the dog\n\n",
			wantCount: 2,
			wantTexts: []string{"Buy groceries", "Walk the dog"},
		},
		{
			name:      "empty content",
			content:   "",
			wantCount: 0,
		},
		{
			name:      "whitespace only",
			content:   "   \n  \n\n",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseTextTasks(tt.content)
			if len(got) != tt.wantCount {
				t.Fatalf("parseTextTasks() returned %d tasks, want %d", len(got), tt.wantCount)
			}
			for i, wantText := range tt.wantTexts {
				if got[i].Text != wantText {
					t.Errorf("task[%d].Text = %q, want %q", i, got[i].Text, wantText)
				}
			}
		})
	}
}

func TestParseMarkdownTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       string
		wantCount     int
		wantTexts     []string
		wantCompleted []bool
	}{
		{
			name:          "mixed checkboxes",
			content:       "- [ ] Buy groceries\n- [x] Walk the dog\n- [ ] Write code",
			wantCount:     3,
			wantTexts:     []string{"Buy groceries", "Walk the dog", "Write code"},
			wantCompleted: []bool{false, true, false},
		},
		{
			name:          "asterisk markers",
			content:       "* [ ] Task one\n* [X] Task two",
			wantCount:     2,
			wantTexts:     []string{"Task one", "Task two"},
			wantCompleted: []bool{false, true},
		},
		{
			name:      "ignores non-checkbox lines",
			content:   "# My Tasks\n- [ ] Real task\nSome random text\n- [x] Done task",
			wantCount: 2,
			wantTexts: []string{"Real task", "Done task"},
		},
		{
			name:      "empty content",
			content:   "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseMarkdownTasks(tt.content)
			if len(got) != tt.wantCount {
				t.Fatalf("parseMarkdownTasks() returned %d tasks, want %d", len(got), tt.wantCount)
			}
			for i, wantText := range tt.wantTexts {
				if got[i].Text != wantText {
					t.Errorf("task[%d].Text = %q, want %q", i, got[i].Text, wantText)
				}
			}
			for i, wantDone := range tt.wantCompleted {
				isComplete := got[i].Status == StatusComplete
				if isComplete != wantDone {
					t.Errorf("task[%d] completed = %v, want %v", i, isComplete, wantDone)
				}
			}
		})
	}
}

func TestParseCheckbox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		line          string
		wantText      string
		wantCompleted bool
	}{
		{"unchecked dash", "- [ ] Buy milk", "Buy milk", false},
		{"checked dash", "- [x] Walk dog", "Walk dog", true},
		{"checked dash uppercase", "- [X] Code review", "Code review", true},
		{"unchecked asterisk", "* [ ] Read book", "Read book", false},
		{"checked asterisk", "* [x] Send email", "Send email", true},
		{"not a checkbox", "Buy groceries", "", false},
		{"heading", "# Tasks", "", false},
		{"plain list", "- Buy groceries", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			text, completed := parseCheckbox(tt.line)
			if text != tt.wantText {
				t.Errorf("parseCheckbox(%q) text = %q, want %q", tt.line, text, tt.wantText)
			}
			if completed != tt.wantCompleted {
				t.Errorf("parseCheckbox(%q) completed = %v, want %v", tt.line, completed, tt.wantCompleted)
			}
		})
	}
}

func TestStripListMarker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		want string
	}{
		{"dash", "- Buy groceries", "Buy groceries"},
		{"asterisk", "* Walk the dog", "Walk the dog"},
		{"numbered dot", "1. Write code", "Write code"},
		{"numbered paren", "2) Read a book", "Read a book"},
		{"no marker", "Buy groceries", "Buy groceries"},
		{"double digit", "12. Big list item", "Big list item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stripListMarker(tt.line)
			if got != tt.want {
				t.Errorf("stripListMarker(%q) = %q, want %q", tt.line, got, tt.want)
			}
		})
	}
}

func TestImportTasksFromFile(t *testing.T) {
	t.Parallel()

	t.Run("text file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.txt")
		content := "Buy groceries\nWalk the dog\nWrite code"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := ImportTasksFromFile(path)
		if err != nil {
			t.Fatalf("ImportTasksFromFile() error = %v", err)
		}
		if result.Format != "text" {
			t.Errorf("format = %q, want %q", result.Format, "text")
		}
		if len(result.Tasks) != 3 {
			t.Errorf("got %d tasks, want 3", len(result.Tasks))
		}
	})

	t.Run("markdown file", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "tasks.md")
		content := "# My Tasks\n- [ ] Buy groceries\n- [x] Walk the dog\n- [ ] Write code"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := ImportTasksFromFile(path)
		if err != nil {
			t.Fatalf("ImportTasksFromFile() error = %v", err)
		}
		if result.Format != "markdown" {
			t.Errorf("format = %q, want %q", result.Format, "markdown")
		}
		if len(result.Tasks) != 3 {
			t.Errorf("got %d tasks, want 3", len(result.Tasks))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		t.Parallel()
		_, err := ImportTasksFromFile("/nonexistent/path/tasks.txt")
		if err == nil {
			t.Error("ImportTasksFromFile() expected error for nonexistent file")
		}
	})
}

func TestExpandPath(t *testing.T) {
	t.Parallel()

	t.Run("absolute path unchanged", func(t *testing.T) {
		t.Parallel()
		got, err := expandPath("/tmp/tasks.txt")
		if err != nil {
			t.Fatalf("expandPath() error = %v", err)
		}
		if got != "/tmp/tasks.txt" {
			t.Errorf("expandPath() = %q, want %q", got, "/tmp/tasks.txt")
		}
	})

	t.Run("tilde expansion", func(t *testing.T) {
		t.Parallel()
		got, err := expandPath("~/tasks.txt")
		if err != nil {
			t.Fatalf("expandPath() error = %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, "tasks.txt")
		if got != want {
			t.Errorf("expandPath() = %q, want %q", got, want)
		}
	})
}
