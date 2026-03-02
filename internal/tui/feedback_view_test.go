package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestTask() *tasks.Task {
	return tasks.NewTask("Test task for feedback")
}

// --- View Rendering ---

func TestFeedbackView_RendersAllOptions(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.SetWidth(80)
	view := fv.View()

	expectedOptions := []string{"Blocked", "Not now", "Needs breakdown", "Other comment"}
	for _, opt := range expectedOptions {
		if !strings.Contains(view, opt) {
			t.Errorf("FeedbackView should contain %q", opt)
		}
	}
}

func TestFeedbackView_RendersHeader(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.SetWidth(80)
	view := fv.View()
	if !strings.Contains(view, "Door Feedback") {
		t.Error("FeedbackView should contain 'Door Feedback' header")
	}
}

func TestFeedbackView_RendersTaskText(t *testing.T) {
	task := tasks.NewTask("Buy groceries")
	fv := NewFeedbackView(task)
	fv.SetWidth(80)
	view := fv.View()
	if !strings.Contains(view, "Buy groceries") {
		t.Error("FeedbackView should show the task text")
	}
}

func TestFeedbackView_RendersHelpText(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.SetWidth(80)
	view := fv.View()
	if !strings.Contains(view, "1-4") {
		t.Error("FeedbackView should contain '1-4' in help text")
	}
}

// --- Feedback Selection via Number Keys ---

func TestFeedbackView_NumberKeySelects(t *testing.T) {
	tests := []struct {
		key            string
		expectedType   string
		expectedHasCmd bool
	}{
		{"1", "blocked", true},
		{"2", "not-now", true},
		{"3", "needs-breakdown", true},
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			task := newTestTask()
			fv := NewFeedbackView(task)
			cmd := fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			if cmd == nil {
				t.Fatalf("key %q should return a command", tt.key)
			}
			msg := cmd()
			dfm, ok := msg.(DoorFeedbackMsg)
			if !ok {
				t.Fatalf("expected DoorFeedbackMsg, got %T", msg)
			}
			if dfm.FeedbackType != tt.expectedType {
				t.Errorf("expected feedback type %q, got %q", tt.expectedType, dfm.FeedbackType)
			}
			if dfm.Comment != "" {
				t.Errorf("expected empty comment for predefined option, got %q", dfm.Comment)
			}
			if dfm.Task.ID != task.ID {
				t.Errorf("expected task ID %q, got %q", task.ID, dfm.Task.ID)
			}
		})
	}
}

// --- Esc Cancels ---

func TestFeedbackView_EscCancels(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("Esc should return a command")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

// --- Other / Custom Input ---

func TestFeedbackView_OtherKey_EntersCustomMode(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if cmd != nil {
		t.Error("key '4' (Other) should enter custom mode, not return a command")
	}
	if !fv.isCustom {
		t.Error("isCustom should be true after pressing '4'")
	}
}

func TestFeedbackView_CustomInput_EnterSubmits(t *testing.T) {
	task := newTestTask()
	fv := NewFeedbackView(task)
	// Enter custom mode
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})

	// Type comment
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})

	// Submit
	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command after typing custom comment")
	}
	msg := cmd()
	dfm, ok := msg.(DoorFeedbackMsg)
	if !ok {
		t.Fatalf("expected DoorFeedbackMsg, got %T", msg)
	}
	if dfm.FeedbackType != "other" {
		t.Errorf("expected feedback type 'other', got %q", dfm.FeedbackType)
	}
	if dfm.Comment != "too" {
		t.Errorf("expected comment 'too', got %q", dfm.Comment)
	}
}

func TestFeedbackView_CustomInput_EmptyEnterIgnored(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})

	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter with empty custom input should not return a command")
	}
}

func TestFeedbackView_CustomInput_EscCancels(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	cmd := fv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Error("Esc in custom mode should return to feedback selection, not return a command")
	}
	if fv.isCustom {
		t.Error("isCustom should be false after Esc")
	}
	if fv.customInput != "" {
		t.Error("customInput should be cleared after Esc")
	}
}

func TestFeedbackView_CustomInput_BackspaceWorks(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	fv.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if fv.customInput != "a" {
		t.Errorf("expected 'a' after backspace, got %q", fv.customInput)
	}
}

func TestFeedbackView_CustomMode_ShowsInputPrompt(t *testing.T) {
	fv := NewFeedbackView(newTestTask())
	fv.SetWidth(80)
	fv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	view := fv.View()
	if !strings.Contains(view, "Enter your comment") {
		t.Error("View should show input prompt in custom mode")
	}
}
