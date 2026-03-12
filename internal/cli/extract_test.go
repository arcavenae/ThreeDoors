package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/intelligence/llm"
	"github.com/arcaven/ThreeDoors/internal/intelligence/services"
	"github.com/spf13/cobra"
)

// mockExtractBackend implements llm.LLMBackend for testing.
type mockExtractBackend struct {
	response string
	err      error
}

func (m *mockExtractBackend) Name() string { return "mock" }
func (m *mockExtractBackend) Complete(_ context.Context, _ string) (string, error) {
	return m.response, m.err
}
func (m *mockExtractBackend) Available(_ context.Context) bool { return true }

// fakePipeStat simulates os.FileInfo for a pipe (not a terminal device).
type fakePipeStat struct{}

func (f *fakePipeStat) Name() string       { return "stdin" }
func (f *fakePipeStat) Size() int64        { return 0 }
func (f *fakePipeStat) Mode() os.FileMode  { return os.ModeNamedPipe }
func (f *fakePipeStat) ModTime() time.Time { return time.Time{} }
func (f *fakePipeStat) IsDir() bool        { return false }
func (f *fakePipeStat) Sys() interface{}   { return nil }

// fakeTermStat simulates os.FileInfo for a terminal (not piped).
type fakeTermStat struct{}

func (f *fakeTermStat) Name() string       { return "stdin" }
func (f *fakeTermStat) Size() int64        { return 0 }
func (f *fakeTermStat) Mode() os.FileMode  { return os.ModeCharDevice }
func (f *fakeTermStat) ModTime() time.Time { return time.Time{} }
func (f *fakeTermStat) IsDir() bool        { return false }
func (f *fakeTermStat) Sys() interface{}   { return nil }

func newTestDeps(backend llm.LLMBackend, stdin string) extractDeps {
	return extractDeps{
		stdin: strings.NewReader(stdin),
		stdinStatFunc: func() (os.FileInfo, error) {
			return &fakePipeStat{}, nil
		},
		configLoader: func() (llm.Config, error) {
			return llm.DefaultConfig(), nil
		},
		backendFunc: func(_ context.Context, _ llm.Config) (llm.LLMBackend, error) {
			if backend == nil {
				return nil, llm.ErrBackendUnavailable
			}
			return backend, nil
		},
		importFunc: func(tasks []services.ExtractedTask) (int, error) {
			return len(tasks), nil
		},
		promptFunc: func(_ io.Reader, _ string) (string, error) {
			return "y", nil
		},
	}
}

// makeCmd creates a minimal cobra.Command with the --json flag for testing.
func makeCmd(jsonMode bool) *cobra.Command {
	cmd := &cobra.Command{Use: "extract"}
	cmd.Flags().Bool("json", jsonMode, "")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	return cmd
}

func TestExtractCmd_FileInput(t *testing.T) {
	t.Parallel()

	tmpFile := t.TempDir() + "/notes.txt"
	if err := os.WriteFile(tmpFile, []byte("Buy groceries and fix the fence"), 0o644); err != nil {
		t.Fatal(err)
	}

	backend := &mockExtractBackend{
		response: `[{"text":"Buy groceries","effort":1,"tags":["shopping"]},{"text":"Fix the fence","effort":3,"tags":["home"]}]`,
	}

	deps := newTestDeps(backend, "")
	deps.stdinStatFunc = func() (os.FileInfo, error) {
		return &fakeTermStat{}, nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, tmpFile, false, true, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractCmd_StdinInput(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{
		response: `[{"text":"Review PR","effort":2,"tags":["dev"]}]`,
	}

	deps := newTestDeps(backend, "Review the open PR for the frontend refactor")
	cmd := makeCmd(false)

	err := runExtract(cmd, "", false, true, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractCmd_JSONOutput(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{
		response: `[{"text":"Write docs","effort":2,"tags":["docs"]}]`,
	}

	deps := newTestDeps(backend, "Write documentation for the API")
	cmd := makeCmd(true)

	err := runExtract(cmd, "", false, false, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractCmd_AutoYes(t *testing.T) {
	t.Parallel()

	var imported int
	backend := &mockExtractBackend{
		response: `[{"text":"Task A","effort":1,"tags":[]},{"text":"Task B","effort":2,"tags":[]}]`,
	}

	deps := newTestDeps(backend, "Task A and Task B need doing")
	deps.importFunc = func(tasks []services.ExtractedTask) (int, error) {
		imported = len(tasks)
		return imported, nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, true, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if imported != 2 {
		t.Errorf("expected 2 tasks imported, got %d", imported)
	}
}

func TestExtractCmd_NoBackend(t *testing.T) {
	t.Parallel()

	deps := newTestDeps(nil, "some input text")
	cmd := makeCmd(false)

	err := runExtract(cmd, "", false, false, deps)
	if err == nil {
		t.Fatal("expected error when no backend available")
	}

	if !strings.Contains(err.Error(), "no LLM backend available") {
		t.Errorf("error should mention backend unavailability, got: %v", err)
	}
}

func TestExtractCmd_MissingFile(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{response: "[]"}
	deps := newTestDeps(backend, "")
	deps.stdinStatFunc = func() (os.FileInfo, error) {
		return &fakeTermStat{}, nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "/nonexistent/missing.txt", false, false, deps)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtractCmd_NoInput(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{response: "[]"}
	deps := newTestDeps(backend, "")
	deps.stdinStatFunc = func() (os.FileInfo, error) {
		return &fakeTermStat{}, nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, false, deps)
	if err == nil {
		t.Fatal("expected error when no input provided")
	}

	if !strings.Contains(err.Error(), "no input provided") {
		t.Errorf("error should mention no input, got: %v", err)
	}
}

func TestExtractCmd_EmptyResult(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{response: "[]"}
	deps := newTestDeps(backend, "random gibberish with no tasks")

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, false, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractCmd_ImportError(t *testing.T) {
	t.Parallel()

	backend := &mockExtractBackend{
		response: `[{"text":"Task A","effort":1,"tags":[]}]`,
	}

	deps := newTestDeps(backend, "Task A needs doing")
	deps.importFunc = func(_ []services.ExtractedTask) (int, error) {
		return 0, errors.New("disk full")
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, true, deps)
	if err == nil {
		t.Fatal("expected error on import failure")
	}
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestExtractCmd_SelectiveImport(t *testing.T) {
	t.Parallel()

	var importedTasks []services.ExtractedTask
	backend := &mockExtractBackend{
		response: `[{"text":"Task A","effort":1,"tags":[]},{"text":"Task B","effort":2,"tags":[]},{"text":"Task C","effort":3,"tags":[]}]`,
	}

	deps := newTestDeps(backend, "Three tasks here")
	deps.importFunc = func(tasks []services.ExtractedTask) (int, error) {
		importedTasks = tasks
		return len(tasks), nil
	}

	promptResponses := []string{"select", "1,3"}
	promptIdx := 0
	deps.promptFunc = func(_ io.Reader, _ string) (string, error) {
		if promptIdx >= len(promptResponses) {
			return "", io.EOF
		}
		resp := promptResponses[promptIdx]
		promptIdx++
		return resp, nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, false, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(importedTasks) != 2 {
		t.Fatalf("expected 2 tasks imported, got %d", len(importedTasks))
	}
	if importedTasks[0].Text != "Task A" {
		t.Errorf("first imported task = %q, want %q", importedTasks[0].Text, "Task A")
	}
	if importedTasks[1].Text != "Task C" {
		t.Errorf("second imported task = %q, want %q", importedTasks[1].Text, "Task C")
	}
}

func TestExtractCmd_CancelImport(t *testing.T) {
	t.Parallel()

	var importCalled bool
	backend := &mockExtractBackend{
		response: `[{"text":"Task A","effort":1,"tags":[]}]`,
	}

	deps := newTestDeps(backend, "Task A")
	deps.importFunc = func(_ []services.ExtractedTask) (int, error) {
		importCalled = true
		return 1, nil
	}
	deps.promptFunc = func(_ io.Reader, _ string) (string, error) {
		return "n", nil
	}

	cmd := makeCmd(false)
	err := runExtract(cmd, "", false, false, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if importCalled {
		t.Error("import should not have been called when user cancels")
	}
}

func TestExtractCmd_JSONWithAutoYes(t *testing.T) {
	t.Parallel()

	var imported int
	backend := &mockExtractBackend{
		response: `[{"text":"Write tests","effort":2,"tags":["dev"]}]`,
	}

	deps := newTestDeps(backend, "Write tests for the extract command")
	deps.importFunc = func(tasks []services.ExtractedTask) (int, error) {
		imported = len(tasks)
		return imported, nil
	}

	cmd := makeCmd(true)
	err := runExtract(cmd, "", false, true, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if imported != 1 {
		t.Errorf("import func should have been called, got imported=%d", imported)
	}
}

func TestExtractCmd_ClipboardInput(t *testing.T) {
	t.Parallel()

	// Clipboard extraction goes through the extractor which calls pbpaste.
	// We can't easily mock that without injecting a custom runner into the extractor.
	// Instead, test the flag parsing by verifying the code path is reached.
	backend := &mockExtractBackend{
		response: `[{"text":"Clipboard task","effort":1,"tags":[]}]`,
	}

	deps := newTestDeps(backend, "")
	deps.stdinStatFunc = func() (os.FileInfo, error) {
		return &fakeTermStat{}, nil
	}

	cmd := makeCmd(false)
	// Clipboard will fail because pbpaste won't work in test env, but it
	// exercises the clipboard code path.
	err := runExtract(cmd, "", true, false, deps)
	// We expect an error since pbpaste can't run in a test context.
	if err == nil {
		// If it succeeds (e.g., clipboard has content), that's also fine.
		return
	}
	// The error should be about clipboard, not "no input".
	if strings.Contains(err.Error(), "no input provided") {
		t.Errorf("clipboard flag should have been used, got no-input error: %v", err)
	}
}

func TestMapEffort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		effort int
		want   string
	}{
		{1, "quick-win"},
		{2, "quick-win"},
		{3, "medium"},
		{4, "deep-work"},
		{5, "deep-work"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("effort_%d", tt.effort), func(t *testing.T) {
			t.Parallel()
			got := mapEffort(tt.effort)
			if string(got) != tt.want {
				t.Errorf("mapEffort(%d) = %q, want %q", tt.effort, got, tt.want)
			}
		})
	}
}

func TestEffortToLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		effort int
		want   string
	}{
		{1, "trivial"},
		{2, "quick"},
		{3, "medium"},
		{4, "significant"},
		{5, "major"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("effort_%d", tt.effort), func(t *testing.T) {
			t.Parallel()
			got := effortToLabel(tt.effort)
			if got != tt.want {
				t.Errorf("effortToLabel(%d) = %q, want %q", tt.effort, got, tt.want)
			}
		})
	}
}

func TestResolveSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filePath  string
		clipboard bool
		want      string
	}{
		{"file", "notes.txt", false, "file:notes.txt"},
		{"clipboard", "", true, "clipboard"},
		{"stdin", "", false, "stdin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveSource(tt.filePath, tt.clipboard)
			if got != tt.want {
				t.Errorf("resolveSource(%q, %v) = %q, want %q", tt.filePath, tt.clipboard, got, tt.want)
			}
		})
	}
}

func TestToJSONTasks(t *testing.T) {
	t.Parallel()

	tasks := []services.ExtractedTask{
		{Text: "Task A", Effort: 1, Tags: []string{"dev"}, Confidence: 0.9},
		{Text: "Task B", Effort: 3, Tags: nil},
	}

	result := toJSONTasks(tasks)
	if len(result) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result))
	}

	if result[0].Text != "Task A" {
		t.Errorf("task[0].Text = %q, want %q", result[0].Text, "Task A")
	}
	if result[1].Tags == nil {
		t.Error("nil tags should be converted to empty slice")
	}
	if len(result[1].Tags) != 0 {
		t.Errorf("nil tags should be empty slice, got %v", result[1].Tags)
	}

	// Verify JSON serialization works.
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"tags":[]`) {
		t.Errorf("JSON should contain empty tags array, got: %s", data)
	}
}

func TestExtractCmd_Registered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "extract" {
			found = true
			break
		}
	}
	if !found {
		t.Error("extract command should be registered in root command")
	}
}

func TestExtractCmd_InKnownSubcommands(t *testing.T) {
	t.Parallel()

	subs := KnownSubcommands()
	found := false
	for _, name := range subs {
		if name == "extract" {
			found = true
			break
		}
	}
	if !found {
		t.Error("extract should be in KnownSubcommands()")
	}
}

// Suppress unused import warnings — cobra is used in makeCmd.
var _ = (*cobra.Command)(nil)
