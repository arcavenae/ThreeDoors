package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- View Rendering ---

func TestMoodView_RendersAllOptions(t *testing.T) {
	mv := NewMoodView()
	mv.SetWidth(80)
	view := mv.View()

	expectedMoods := []string{"Focused", "Tired", "Stressed", "Energized", "Distracted", "Calm", "Other"}
	for _, mood := range expectedMoods {
		if !strings.Contains(view, mood) {
			t.Errorf("MoodView should contain %q", mood)
		}
	}
}

func TestMoodView_RendersHeader(t *testing.T) {
	mv := NewMoodView()
	mv.SetWidth(80)
	view := mv.View()
	if !strings.Contains(view, "How are you feeling") {
		t.Error("MoodView should contain 'How are you feeling' header")
	}
}

func TestMoodView_RendersHelpText(t *testing.T) {
	mv := NewMoodView()
	mv.SetWidth(80)
	view := mv.View()
	if !strings.Contains(view, "1-7") {
		t.Error("MoodView should contain '1-7' in help text")
	}
}

// --- Mood Selection via Number Keys ---

func TestMoodView_NumberKeySelects(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"1", "Focused"},
		{"2", "Tired"},
		{"3", "Stressed"},
		{"4", "Energized"},
		{"5", "Distracted"},
		{"6", "Calm"},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			mv := NewMoodView()
			cmd := mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				t.Fatalf("key %q should return a command", tt.key)
				return
			}
			msg := cmd()
			mcm, ok := msg.(MoodCapturedMsg)
			if !ok {
				t.Fatalf("expected MoodCapturedMsg, got %T", msg)
			}
			if mcm.Mood != tt.expected {
				t.Errorf("expected mood %q, got %q", tt.expected, mcm.Mood)
			}
			if mcm.CustomText != "" {
				t.Errorf("expected empty custom text for predefined mood, got %q", mcm.CustomText)
			}
		})
	}
}

// --- Esc Cancels ---

func TestMoodView_EscCancels(t *testing.T) {
	mv := NewMoodView()
	cmd := mv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
		return
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

// --- Other / Custom Input ---

func TestMoodView_OtherKey_EntersCustomMode(t *testing.T) {
	mv := NewMoodView()
	cmd := mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	// Should enter custom input mode, not return a command
	if cmd != nil {
		t.Error("key '7' (Other) should enter custom mode, not return a command")
	}
	if !mv.isCustom {
		t.Error("isCustom should be true after pressing '7'")
	}
}

func TestMoodView_CustomInput_EnterSubmits(t *testing.T) {
	mv := NewMoodView()
	// Enter custom mode
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})

	// Type custom mood
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	// Submit
	cmd := mv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command after typing custom mood")
		return
	}
	msg := cmd()
	mcm, ok := msg.(MoodCapturedMsg)
	if !ok {
		t.Fatalf("expected MoodCapturedMsg, got %T", msg)
	}
	if mcm.Mood != "Other" {
		t.Errorf("expected mood 'Other', got %q", mcm.Mood)
	}
	if mcm.CustomText != "good" {
		t.Errorf("expected custom text 'good', got %q", mcm.CustomText)
	}
}

func TestMoodView_CustomInput_EmptyEnterIgnored(t *testing.T) {
	mv := NewMoodView()
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})

	// Press Enter with empty input
	cmd := mv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter with empty custom input should not return a command")
	}
}

func TestMoodView_CustomInput_EscCancels(t *testing.T) {
	mv := NewMoodView()
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	// Esc cancels custom mode
	cmd := mv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Error("Esc in custom mode should return to mood selection, not return a command")
	}
	if mv.isCustom {
		t.Error("isCustom should be false after Esc")
	}
	if mv.customInput != "" {
		t.Error("customInput should be cleared after Esc")
	}
}

func TestMoodView_CustomInput_BackspaceWorks(t *testing.T) {
	mv := NewMoodView()
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	mv.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if mv.customInput != "a" {
		t.Errorf("expected 'a' after backspace, got %q", mv.customInput)
	}
}

func TestMoodView_CustomMode_ShowsInputPrompt(t *testing.T) {
	mv := NewMoodView()
	mv.SetWidth(80)
	mv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	view := mv.View()
	if !strings.Contains(view, "Enter your mood") {
		t.Error("View should show input prompt in custom mode")
	}
}
