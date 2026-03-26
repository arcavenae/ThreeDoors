package tui

import (
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/intelligence/services"
	tea "github.com/charmbracelet/bubbletea"
)

func makeExtractedTasks() []services.ExtractedTask {
	return []services.ExtractedTask{
		{Text: "Buy groceries", Effort: 2, Tags: []string{"errands"}, Source: "clipboard"},
		{Text: "Write unit tests", Effort: 3, Tags: []string{"dev"}, Source: "clipboard"},
		{Text: "Call dentist", Effort: 1, Tags: nil, Source: "clipboard"},
	}
}

func TestNewExtractView(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	if ev.step != extractStepSourceSelect {
		t.Errorf("initial step = %d, want %d (sourceSelect)", ev.step, extractStepSourceSelect)
	}
	if len(ev.tasks) != 0 {
		t.Errorf("initial tasks = %d, want 0", len(ev.tasks))
	}
}

func TestExtractViewSourceSelectFile(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if ev.step != extractStepFileInput {
		t.Errorf("after 'f', step = %d, want %d (fileInput)", ev.step, extractStepFileInput)
	}
	if ev.source != "file" {
		t.Errorf("source = %q, want %q", ev.source, "file")
	}
}

func TestExtractViewSourceSelectClipboard(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if ev.step != extractStepLoading {
		t.Errorf("after 'c', step = %d, want %d (loading)", ev.step, extractStepLoading)
	}
	if cmd == nil {
		t.Fatal("clipboard selection should return a command")
	}
	msg := cmd()
	startMsg, ok := msg.(ExtractStartMsg)
	if !ok {
		t.Fatalf("expected ExtractStartMsg, got %T", msg)
	}
	if startMsg.Source != "clipboard" {
		t.Errorf("source = %q, want %q", startMsg.Source, "clipboard")
	}
}

func TestExtractViewSourceSelectPaste(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if ev.step != extractStepPasteInput {
		t.Errorf("after 'p', step = %d, want %d (pasteInput)", ev.step, extractStepPasteInput)
	}
}

func TestExtractViewSourceSelectCancel(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewFileInputSubmit(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	// Navigate to file input
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	// Type a path
	for _, r := range "/tmp/notes.txt" {
		ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ev.step != extractStepLoading {
		t.Errorf("after enter, step = %d, want %d (loading)", ev.step, extractStepLoading)
	}
	if cmd == nil {
		t.Fatal("enter should return a command")
	}
	msg := cmd()
	startMsg, ok := msg.(ExtractStartMsg)
	if !ok {
		t.Fatalf("expected ExtractStartMsg, got %T", msg)
	}
	if startMsg.Source != "file" {
		t.Errorf("source = %q, want %q", startMsg.Source, "file")
	}
	if startMsg.Input != "/tmp/notes.txt" {
		t.Errorf("input = %q, want %q", startMsg.Input, "/tmp/notes.txt")
	}
}

func TestExtractViewFileInputBack(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	ev.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ev.step != extractStepSourceSelect {
		t.Errorf("after esc, step = %d, want %d (sourceSelect)", ev.step, extractStepSourceSelect)
	}
}

func TestExtractViewFileInputEmptySubmit(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("enter with empty path should return nil")
	}
	if ev.step != extractStepFileInput {
		t.Errorf("should stay on fileInput, got step = %d", ev.step)
	}
}

func TestExtractViewPasteInputSubmit(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})

	// Type some text
	for _, r := range "Buy milk" {
		ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if ev.step != extractStepLoading {
		t.Errorf("after ctrl+d, step = %d, want %d (loading)", ev.step, extractStepLoading)
	}
	if cmd == nil {
		t.Fatal("ctrl+d should return a command")
	}
	msg := cmd()
	startMsg, ok := msg.(ExtractStartMsg)
	if !ok {
		t.Fatalf("expected ExtractStartMsg, got %T", msg)
	}
	if startMsg.Source != "text" {
		t.Errorf("source = %q, want %q", startMsg.Source, "text")
	}
}

func TestExtractViewPasteInputBack(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	ev.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ev.step != extractStepSourceSelect {
		t.Errorf("after esc, step = %d, want %d (sourceSelect)", ev.step, extractStepSourceSelect)
	}
}

func TestExtractViewLoadingCancel(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	// Navigate to loading via clipboard
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc during loading should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewSetResult(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	tasks := makeExtractedTasks()

	ev.SetResult(tasks, "ollama")

	if ev.step != extractStepReview {
		t.Errorf("step = %d, want %d (review)", ev.step, extractStepReview)
	}
	if len(ev.tasks) != 3 {
		t.Errorf("tasks = %d, want 3", len(ev.tasks))
	}
	if ev.SelectedCount() != 3 {
		t.Errorf("all non-duplicate tasks should be selected, got %d", ev.SelectedCount())
	}
	if ev.backendName != "ollama" {
		t.Errorf("backendName = %q, want %q", ev.backendName, "ollama")
	}
}

func TestExtractViewSetResultWithDuplicates(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	tasks := makeExtractedTasks()
	tasks[1].Duplicate = true

	ev.SetResult(tasks, "test")

	if ev.SelectedCount() != 2 {
		t.Errorf("duplicate tasks should not be selected, got %d selected", ev.SelectedCount())
	}
	if ev.selected[1] {
		t.Error("task[1] is duplicate, should not be selected")
	}
}

func TestExtractViewSetError(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()

	ev.SetError("LLM timeout")

	if ev.step != extractStepReview {
		t.Errorf("step = %d, want %d (review)", ev.step, extractStepReview)
	}
	if ev.errorMsg != "LLM timeout" {
		t.Errorf("errorMsg = %q, want %q", ev.errorMsg, "LLM timeout")
	}
}

func TestExtractViewReviewNavigation(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	if ev.cursorIndex != 0 {
		t.Errorf("initial cursor = %d, want 0", ev.cursorIndex)
	}

	ev.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ev.cursorIndex != 1 {
		t.Errorf("after down, cursor = %d, want 1", ev.cursorIndex)
	}

	ev.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ev.cursorIndex != 2 {
		t.Errorf("after second down, cursor = %d, want 2", ev.cursorIndex)
	}

	// Clamp at end
	ev.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ev.cursorIndex != 2 {
		t.Errorf("should clamp at end, cursor = %d, want 2", ev.cursorIndex)
	}

	ev.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ev.cursorIndex != 1 {
		t.Errorf("after up, cursor = %d, want 1", ev.cursorIndex)
	}
}

func TestExtractViewReviewViNavigation(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ev.cursorIndex != 1 {
		t.Errorf("j should move down, got cursor=%d", ev.cursorIndex)
	}

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if ev.cursorIndex != 0 {
		t.Errorf("k should move up, got cursor=%d", ev.cursorIndex)
	}
}

func TestExtractViewReviewToggle(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	// Toggle first off
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if ev.SelectedCount() != 2 {
		t.Errorf("after toggle off, got %d selected, want 2", ev.SelectedCount())
	}
	if ev.selected[0] {
		t.Error("first item should be deselected")
	}

	// Toggle back on
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if ev.SelectedCount() != 3 {
		t.Errorf("after toggle on, got %d selected, want 3", ev.SelectedCount())
	}
}

func TestExtractViewReviewSelectAll(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	// All selected — pressing 'a' should deselect all
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if ev.SelectedCount() != 0 {
		t.Errorf("after toggle all off, got %d selected, want 0", ev.SelectedCount())
	}

	// None selected — pressing 'a' should select all
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if ev.SelectedCount() != 3 {
		t.Errorf("after toggle all on, got %d selected, want 3", ev.SelectedCount())
	}
}

func TestExtractViewReviewSelectNone(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if ev.SelectedCount() != 0 {
		t.Errorf("after 'n', got %d selected, want 0", ev.SelectedCount())
	}
}

func TestExtractViewReviewImport(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")
	ev.sourceLabel = "clipboard"

	// Deselect first item
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should return a command")
	}
	msg := cmd()
	importMsg, ok := msg.(ExtractImportMsg)
	if !ok {
		t.Fatalf("expected ExtractImportMsg, got %T", msg)
	}
	if len(importMsg.Tasks) != 2 {
		t.Errorf("imported %d tasks, want 2", len(importMsg.Tasks))
	}
	if importMsg.Source != "clipboard" {
		t.Errorf("source = %q, want %q", importMsg.Source, "clipboard")
	}
}

func TestExtractViewReviewImportNoneSelected(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	// Deselect all
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter with nothing selected should return nil")
	}
}

func TestExtractViewReviewCancel(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewReviewQuit(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewErrorDismiss(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetError("network failure")

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("any key during error should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewEmptyResultDismiss(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(nil, "test")

	cmd := ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("any key during empty result should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ExtractCancelMsg); !ok {
		t.Errorf("expected ExtractCancelMsg, got %T", msg)
	}
}

func TestExtractViewEditing(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	// Enter edit mode for first task
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if ev.step != extractStepEditing {
		t.Errorf("step = %d, want %d (editing)", ev.step, extractStepEditing)
	}
	if ev.editIndex != 0 {
		t.Errorf("editIndex = %d, want 0", ev.editIndex)
	}
	if ev.editField != 0 {
		t.Errorf("editField = %d, want 0 (text)", ev.editField)
	}
}

func TestExtractViewEditingEsc(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	ev.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if ev.step != extractStepReview {
		t.Errorf("after esc in editing, step = %d, want %d (review)", ev.step, extractStepReview)
	}
}

func TestExtractViewEditingCycleFields(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Tab to next field (effort)
	ev.Update(tea.KeyMsg{Type: tea.KeyTab})
	if ev.editField != 1 {
		t.Errorf("after tab, editField = %d, want 1 (effort)", ev.editField)
	}

	// Tab to next field (tags)
	ev.Update(tea.KeyMsg{Type: tea.KeyTab})
	if ev.editField != 2 {
		t.Errorf("after second tab, editField = %d, want 2 (tags)", ev.editField)
	}

	// Tab wraps around to text
	ev.Update(tea.KeyMsg{Type: tea.KeyTab})
	if ev.editField != 0 {
		t.Errorf("after third tab, editField = %d, want 0 (text)", ev.editField)
	}
}

func TestExtractViewEditingEnterAdvances(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	// Enter advances from text to effort
	ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ev.editField != 1 {
		t.Errorf("after enter, editField = %d, want 1", ev.editField)
	}

	// Enter advances from effort to tags
	ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ev.editField != 2 {
		t.Errorf("after second enter, editField = %d, want 2", ev.editField)
	}

	// Enter from tags exits editing
	ev.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ev.step != extractStepReview {
		t.Errorf("after third enter, step = %d, want %d (review)", ev.step, extractStepReview)
	}
}

func TestExtractViewSelectedTasks(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetResult(makeExtractedTasks(), "test")

	// Deselect middle item
	ev.Update(tea.KeyMsg{Type: tea.KeyDown})
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	selected := ev.SelectedTasks()
	if len(selected) != 2 {
		t.Fatalf("got %d selected, want 2", len(selected))
	}
	if selected[0].Text != "Buy groceries" {
		t.Errorf("first selected = %q", selected[0].Text)
	}
	if selected[1].Text != "Call dentist" {
		t.Errorf("second selected = %q", selected[1].Text)
	}
}

// Rendering tests

func TestExtractViewRenderSourceSelect(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)

	view := ev.View()
	if !strings.Contains(view, "EXTRACT TASKS") {
		t.Error("view should contain header")
	}
	if !strings.Contains(view, "[f]") {
		t.Error("view should contain file option")
	}
	if !strings.Contains(view, "[c]") {
		t.Error("view should contain clipboard option")
	}
	if !strings.Contains(view, "[p]") {
		t.Error("view should contain paste option")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("view should contain cancel hint")
	}
}

func TestExtractViewRenderLoading(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)

	// Navigate to loading
	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	view := ev.View()
	if !strings.Contains(view, "Extracting tasks") {
		t.Error("loading view should show extracting message")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("loading view should show cancel hint")
	}
}

func TestExtractViewRenderReview(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)
	ev.SetResult(makeExtractedTasks(), "ollama")
	ev.sourceLabel = "clipboard"

	view := ev.View()
	if !strings.Contains(view, "Buy groceries") {
		t.Error("view should contain task text")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("view should contain selected checkbox")
	}
	if !strings.Contains(view, "3/3 selected") {
		t.Error("view should show selection count")
	}
	if !strings.Contains(view, "clipboard") {
		t.Error("view should show source label")
	}
	if !strings.Contains(view, "ollama") {
		t.Error("view should show backend name")
	}
	if !strings.Contains(view, "Import") {
		t.Error("view should contain import hint")
	}
}

func TestExtractViewRenderError(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)
	ev.SetError("connection refused")

	view := ev.View()
	if !strings.Contains(view, "connection refused") {
		t.Error("view should show error message")
	}
}

func TestExtractViewRenderNoTasks(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)
	ev.SetResult(nil, "test")

	view := ev.View()
	if !strings.Contains(view, "No actionable tasks found") {
		t.Error("view should show no tasks message")
	}
}

func TestExtractViewRenderDuplicateBadge(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)
	tasks := makeExtractedTasks()
	tasks[0].Duplicate = true
	ev.SetResult(tasks, "test")

	view := ev.View()
	if !strings.Contains(view, "dup?") {
		t.Error("view should show duplicate badge")
	}
}

func TestExtractViewRenderEffortBadge(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)
	ev.SetResult(makeExtractedTasks(), "test")

	view := ev.View()
	if !strings.Contains(view, "E2") {
		t.Error("view should show effort badge E2")
	}
	if !strings.Contains(view, "E3") {
		t.Error("view should show effort badge E3")
	}
}

func TestExtractViewRenderFileInput(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	view := ev.View()
	if !strings.Contains(view, "Enter file path") {
		t.Error("view should show file path prompt")
	}
}

func TestExtractViewRenderPasteInput(t *testing.T) {
	t.Parallel()
	ev := NewExtractView()
	ev.SetWidth(80)

	ev.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := ev.View()
	if !strings.Contains(view, "Paste your text") {
		t.Error("view should show paste prompt")
	}
}

// Command registration test

func TestExtractCommandRegistered(t *testing.T) {
	t.Parallel()
	found := false
	for _, cmd := range commandRegistry {
		if cmd.Name == "extract" {
			found = true
			break
		}
	}
	if !found {
		t.Error("extract command should be registered in commandRegistry")
	}
}

func TestExtractCommandFiltered(t *testing.T) {
	t.Parallel()
	matches := filterCommands("ext")
	found := false
	for _, m := range matches {
		if m.Name == "extract" {
			found = true
			break
		}
	}
	if !found {
		t.Error("filterCommands('ext') should match extract")
	}
}
