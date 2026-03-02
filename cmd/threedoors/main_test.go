package main

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/arcaven/ThreeDoors/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T) *tui.MainModel {
	t.Helper()
	pool := tasks.NewTaskPool()
	pool.AddTask(tasks.NewTask("Test task 1"))
	pool.AddTask(tasks.NewTask("Test task 2"))
	pool.AddTask(tasks.NewTask("Test task 3"))
	tracker := tasks.NewSessionTracker()
	return tui.NewMainModel(pool, tracker, tasks.NewTextFileProvider(), nil)
}

func TestQuitKey(t *testing.T) {
	m := newTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("'q' key should trigger tea.Quit command")
		return
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Error("'q' key should return a tea.QuitMsg")
	}
}

func TestCtrlCKey(t *testing.T) {
	m := newTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("'ctrl+c' should trigger tea.Quit command")
		return
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Error("'ctrl+c' should return a tea.QuitMsg")
	}
}
