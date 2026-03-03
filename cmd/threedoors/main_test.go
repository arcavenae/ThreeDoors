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
