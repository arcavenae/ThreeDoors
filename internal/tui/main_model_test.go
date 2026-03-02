package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_Init_ReturnsNil(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Init() = %v, want nil", cmd)
	}
}

func TestModel_Update_QuitOnQ(t *testing.T) {
	m := NewModel()
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
	m := NewModel()
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
		{"letter a", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}},
		{"letter z", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}},
		{"enter key", tea.KeyMsg{Type: tea.KeyEnter}},
		{"space key", tea.KeyMsg{Type: tea.KeySpace}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			_, cmd := m.Update(tt.msg)
			if cmd != nil {
				t.Errorf("expected nil command for key %q, got non-nil", tt.name)
			}
		})
	}
}

func TestModel_View_ContainsHeader(t *testing.T) {
	m := NewModel()
	view := m.View()

	if !strings.Contains(view, "ThreeDoors - Technical Demo") {
		t.Errorf("View() = %q, want it to contain %q", view, "ThreeDoors - Technical Demo")
	}
}

func TestModel_View_ContainsQuitHint(t *testing.T) {
	m := NewModel()
	view := m.View()

	if !strings.Contains(view, "Press q to quit") {
		t.Errorf("View() = %q, want it to contain %q", view, "Press q to quit")
	}
}
