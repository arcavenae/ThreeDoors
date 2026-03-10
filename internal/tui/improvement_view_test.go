package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestImprovementView_Render(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)
	view := iv.View()

	if !strings.Contains(view, "Session Reflection") {
		t.Error("expected 'Session Reflection' header")
	}
	if !strings.Contains(view, "What's one thing you could improve") {
		t.Error("expected prompt text")
	}
	if !strings.Contains(view, "Enter to save") {
		t.Error("expected help text")
	}
}

func TestImprovementView_TextInput(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	// Type characters
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	view := iv.View()
	if !strings.Contains(view, "hi") {
		t.Errorf("expected typed text in view, got: %s", view)
	}
}

func TestImprovementView_Backspace(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if iv.input != "a" {
		t.Errorf("expected 'a' after backspace, got: %s", iv.input)
	}
}

func TestImprovementView_BackspaceEmpty(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	// Backspace on empty should not panic
	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd != nil {
		t.Error("backspace on empty should not produce a command")
	}
	if iv.input != "" {
		t.Errorf("expected empty input, got: %s", iv.input)
	}
}

func TestImprovementView_SubmitWithText(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter with text should produce a command")
		return
	}

	msg := cmd()
	submitted, ok := msg.(ImprovementSubmittedMsg)
	if !ok {
		t.Fatalf("expected ImprovementSubmittedMsg, got %T", msg)
	}
	if submitted.Text != "test" {
		t.Errorf("expected 'test', got: %s", submitted.Text)
	}
}

func TestImprovementView_SubmitEmpty(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on empty should produce a skip command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ImprovementSkippedMsg); !ok {
		t.Fatalf("expected ImprovementSkippedMsg, got %T", msg)
	}
}

func TestImprovementView_EscSkips(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should produce a skip command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ImprovementSkippedMsg); !ok {
		t.Fatalf("expected ImprovementSkippedMsg, got %T", msg)
	}
}

func TestImprovementView_CtrlCSkips(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("Ctrl+C should produce a skip command")
		return
	}

	msg := cmd()
	if _, ok := msg.(ImprovementSkippedMsg); !ok {
		t.Fatalf("expected ImprovementSkippedMsg, got %T", msg)
	}
}

func TestImprovementView_CharLimit(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	// Fill to max
	for i := 0; i < 500; i++ {
		iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	}

	if len(iv.input) != 500 {
		t.Errorf("expected 500 chars, got %d", len(iv.input))
	}

	// Should not grow past 500
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if len(iv.input) != 500 {
		t.Errorf("expected 500 chars after limit, got %d", len(iv.input))
	}
}

func TestImprovementView_TrimsWhitespace(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(80)

	// Type spaces then text
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	iv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should produce a command")
		return
	}

	msg := cmd()
	submitted, ok := msg.(ImprovementSubmittedMsg)
	if !ok {
		t.Fatalf("expected ImprovementSubmittedMsg, got %T", msg)
	}
	if submitted.Text != "x" {
		t.Errorf("expected trimmed 'x', got: '%s'", submitted.Text)
	}
}

func TestImprovementView_MinWidth(t *testing.T) {
	iv := NewImprovementView()
	iv.SetWidth(10) // Very narrow
	view := iv.View()

	// Should not panic with narrow width
	if view == "" {
		t.Error("expected non-empty view even at narrow width")
	}
}
