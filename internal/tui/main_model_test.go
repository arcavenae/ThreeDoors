package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arcaven/ThreeDoors/internal/tasks"
)

func sampleTasks() []tasks.Task {
	return []tasks.Task{
		{Text: "Task Alpha"},
		{Text: "Task Beta"},
		{Text: "Task Gamma"},
		{Text: "Task Delta"},
		{Text: "Task Epsilon"},
	}
}

func TestModel_Init_ReturnsCmd(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() = nil, want non-nil tea.Cmd for task loading")
	}
}

func TestModel_Init_ReturnsNilWhenNoLoader(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() with no loader should return nil")
	}
}

func TestModel_Update_QuitOnQ(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", result)
	}
}

func TestModel_Update_QuitOnCtrlC(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", result)
	}
}

func TestModel_Update_IgnoresOtherKeys(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"enter key", tea.KeyMsg{Type: tea.KeyEnter}},
		{"space key", tea.KeyMsg{Type: tea.KeySpace}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			_, cmd := m.Update(tt.msg)
			if cmd != nil {
				t.Errorf("expected nil command for key %q, got non-nil", tt.name)
			}
		})
	}
}

func TestModel_NoDoorSelectedInitially(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	if m.selectedIdx != -1 {
		t.Errorf("selectedIdx = %d, want -1", m.selectedIdx)
	}
}

func TestModel_HasThreeDoorsAfterCreation(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	if len(m.doors) != 3 {
		t.Errorf("doors count = %d, want 3", len(m.doors))
	}
}

func TestModel_SelectLeftDoor(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"a key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}},
		{"left arrow", tea.KeyMsg{Type: tea.KeyLeft}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			updated, _ := m.Update(tt.msg)
			um := updated.(Model)
			if um.selectedIdx != 0 {
				t.Errorf("selectedIdx = %d, want 0", um.selectedIdx)
			}
		})
	}
}

func TestModel_SelectCenterDoor(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"w key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w")}},
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			updated, _ := m.Update(tt.msg)
			um := updated.(Model)
			if um.selectedIdx != 1 {
				t.Errorf("selectedIdx = %d, want 1", um.selectedIdx)
			}
		})
	}
}

func TestModel_SelectRightDoor(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"d key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}},
		{"right arrow", tea.KeyMsg{Type: tea.KeyRight}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			updated, _ := m.Update(tt.msg)
			um := updated.(Model)
			if um.selectedIdx != 2 {
				t.Errorf("selectedIdx = %d, want 2", um.selectedIdx)
			}
		})
	}
}

func TestModel_RerollDoors(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"s key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}},
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			// Select a door first
			m.selectedIdx = 1

			updated, _ := m.Update(tt.msg)
			um := updated.(Model)
			if um.selectedIdx != -1 {
				t.Errorf("selectedIdx = %d, want -1 after reroll", um.selectedIdx)
			}
			if len(um.doors) != 3 {
				t.Errorf("doors count = %d, want 3 after reroll", len(um.doors))
			}
		})
	}
}

func TestModel_TaskManagementKeysAreNoOp(t *testing.T) {
	keys := []string{"c", "b", "i", "e", "f", "p"}
	for _, key := range keys {
		t.Run(key+" key", func(t *testing.T) {
			m := NewModelWithTasks(sampleTasks())
			m.selectedIdx = 0 // Select a door
			origDoors := make([]tasks.Task, len(m.doors))
			copy(origDoors, m.doors)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
			updated, cmd := m.Update(msg)
			um := updated.(Model)
			if cmd != nil {
				t.Errorf("key %q should return nil cmd", key)
			}
			if um.selectedIdx != 0 {
				t.Errorf("key %q changed selectedIdx to %d", key, um.selectedIdx)
			}
		})
	}
}

func TestModel_View_ContainsHeader(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	view := m.View()

	if !strings.Contains(view, "ThreeDoors - Technical Demo") {
		t.Errorf("View() missing header, got: %q", view)
	}
}

func TestModel_View_ContainsQuitHint(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	view := m.View()

	if !strings.Contains(view, "quit") {
		t.Errorf("View() missing quit hint, got: %q", view)
	}
}

func TestModel_View_ContainsDoorTexts(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	view := m.View()

	for _, door := range m.doors {
		if !strings.Contains(view, door.Text) {
			t.Errorf("View() missing door text %q", door.Text)
		}
	}
}

func TestModel_View_ShowsErrorOnLoadFailure(t *testing.T) {
	m := NewModelWithTasks(nil)
	m.err = fmt.Errorf("permission denied")
	view := m.View()

	if !strings.Contains(view, "Error loading tasks:") {
		t.Errorf("View() missing error prefix, got: %q", view)
	}
	if !strings.Contains(view, "permission denied") {
		t.Errorf("View() missing error message, got: %q", view)
	}
}

func TestModel_View_ShowsNoTasksMessage(t *testing.T) {
	m := NewModelWithTasks(nil)
	view := m.View()

	if !strings.Contains(view, "No tasks found") {
		t.Errorf("View() missing no-tasks message, got: %q", view)
	}
}

func TestModel_View_HandlesFewTasks(t *testing.T) {
	twoTasks := []tasks.Task{{Text: "Only One"}, {Text: "Only Two"}}
	m := NewModelWithTasks(twoTasks)

	if len(m.doors) != 2 {
		t.Errorf("doors count = %d, want 2", len(m.doors))
	}

	view := m.View()
	if !strings.Contains(view, "Only One") || !strings.Contains(view, "Only Two") {
		t.Errorf("View() missing task texts with 2 tasks, got: %q", view)
	}
}

func TestModel_WindowSizeMsg(t *testing.T) {
	m := NewModelWithTasks(sampleTasks())
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.width != 120 {
		t.Errorf("width = %d, want 120", um.width)
	}
	if um.height != 40 {
		t.Errorf("height = %d, want 40", um.height)
	}
}

func TestModel_TasksLoadedMsg(t *testing.T) {
	m := Model{selectedIdx: -1, width: defaultWidth, height: defaultHeight}
	loadedTasks := sampleTasks()
	msg := tasksLoadedMsg{tasks: loadedTasks}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if len(um.tasks) != len(loadedTasks) {
		t.Errorf("tasks count = %d, want %d", len(um.tasks), len(loadedTasks))
	}
	if len(um.doors) != 3 {
		t.Errorf("doors count = %d, want 3", len(um.doors))
	}
	if um.selectedIdx != -1 {
		t.Errorf("selectedIdx = %d, want -1", um.selectedIdx)
	}
}

func TestModel_TasksLoadErrorMsg(t *testing.T) {
	m := Model{selectedIdx: -1, width: defaultWidth, height: defaultHeight}
	msg := tasksLoadErrorMsg{err: fmt.Errorf("test error")}

	updated, _ := m.Update(msg)
	um := updated.(Model)

	if um.err == nil {
		t.Error("err should be set after tasksLoadErrorMsg")
	}
}
