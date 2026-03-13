package main

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters/textfile"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) *tui.MainModel {
	t.Helper()
	pool := core.NewTaskPool()
	pool.AddTask(core.NewTask("Test task 1"))
	pool.AddTask(core.NewTask("Test task 2"))
	pool.AddTask(core.NewTask("Test task 3"))
	tracker := core.NewSessionTracker()
	return tui.NewMainModel(pool, tracker, textfile.NewTextFileProvider(), nil, false, nil)
}

func TestQuitKey(t *testing.T) {
	m := newTestModel(t)

	// 'q' sends RequestQuitMsg, which (with no completions and <5min) becomes tea.Quit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("'q' key should trigger a command")
		return
	}

	// First step: RequestQuitMsg
	result := cmd()
	updated, cmd = updated.Update(result)

	if cmd == nil {
		t.Error("RequestQuitMsg should trigger tea.Quit command")
		return
	}

	quitResult := cmd()
	if _, ok := quitResult.(tea.QuitMsg); !ok {
		t.Error("'q' key should ultimately return a tea.QuitMsg")
	}
	_ = updated
}

func TestRegisterBuiltinAdapters_PopulatesRegistry(t *testing.T) {
	t.Parallel()

	reg := core.NewRegistry()
	registerBuiltinAdapters(reg)

	// All built-in adapters must be registered.
	expected := []string{
		"textfile",
		"applenotes",
		"jira",
		"github",
		"obsidian",
		"reminders",
		"todoist",
		"linear",
		"clickup",
	}
	for _, name := range expected {
		if !reg.IsRegistered(name) {
			t.Errorf("adapter %q not registered after registerBuiltinAdapters", name)
		}
	}
}

func TestRegisterBuiltinAdapters_CalledBeforeCLIRouting(t *testing.T) {
	// Regression test for Story 66.1: adapters must be registered before
	// the CLI routing branch so that CLI commands can resolve providers.
	// Verify the global default registry has adapters populated — this
	// mirrors what main() does before the CLI routing if-block.
	reg := core.DefaultRegistry()
	registerBuiltinAdapters(reg)

	if !reg.IsRegistered("textfile") {
		t.Error("DefaultRegistry missing textfile after registerBuiltinAdapters — CLI commands would fail")
	}
	if !reg.IsRegistered("jira") {
		t.Error("DefaultRegistry missing jira after registerBuiltinAdapters — non-textfile CLI commands would fail")
	}
}

func TestIsSubcommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"known command task", "task", true},
		{"known command doors", "doors", true},
		{"known command version", "version", true},
		{"unknown command foo", "foo", false},
		{"empty string", "", false},
		{"plan is known", "plan", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isSubcommand(tt.arg); got != tt.want {
				t.Errorf("isSubcommand(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestCtrlCKey(t *testing.T) {
	m := newTestModel(t)

	// 'ctrl+c' sends RequestQuitMsg, which (with no completions and <5min) becomes tea.Quit
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updated, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("'ctrl+c' should trigger a command")
		return
	}

	// First step: RequestQuitMsg
	result := cmd()
	updated, cmd = updated.Update(result)

	if cmd == nil {
		t.Error("RequestQuitMsg should trigger tea.Quit command")
		return
	}

	quitResult := cmd()
	if _, ok := quitResult.(tea.QuitMsg); !ok {
		t.Error("'ctrl+c' should ultimately return a tea.QuitMsg")
	}
	_ = updated
}
