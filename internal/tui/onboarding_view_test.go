package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnboardingView_StepProgression(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.SetWidth(80)

	// Initial state should be welcome step
	view := ov.View()
	if !strings.Contains(view, "Welcome to ThreeDoors") {
		t.Error("expected welcome step on init")
	}
	if !strings.Contains(view, "Step 1 of 5") {
		t.Error("expected step 1 of 5 indicator")
	}

	// Press Enter to advance to keybindings
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "Key Bindings") {
		t.Error("expected keybindings step after Enter")
	}
	if !strings.Contains(view, "Step 2 of 5") {
		t.Error("expected step 2 of 5 indicator")
	}

	// Press Enter to advance to values
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "Values & Goals") {
		t.Error("expected values step after Enter")
	}
	if !strings.Contains(view, "Step 3 of 5") {
		t.Error("expected step 3 of 5 indicator")
	}

	// Press Enter with empty to advance to import
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "Import Tasks") {
		t.Error("expected import step after empty Enter on values")
	}
	if !strings.Contains(view, "Step 4 of 5") {
		t.Error("expected step 4 of 5 indicator")
	}

	// Press Enter with empty to advance to done (skip import)
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "You're All Set") {
		t.Error("expected done step after empty Enter on import")
	}
	if !strings.Contains(view, "Step 5 of 5") {
		t.Error("expected step 5 of 5 indicator")
	}

	// Press Enter to complete onboarding
	cmd := ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on final Enter")
	}
	msg := cmd()
	if _, ok := msg.(OnboardingCompletedMsg); !ok {
		t.Errorf("expected OnboardingCompletedMsg, got %T", msg)
	}
}

func TestOnboardingView_SkipAtEveryStep(t *testing.T) {
	t.Parallel()

	steps := []struct {
		name string
		step onboardingStep
	}{
		{"welcome", stepWelcome},
		{"keybindings", stepKeybindings},
		{"done", stepDone},
	}

	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ov := NewOnboardingView()
			ov.step = tt.step

			cmd := ov.Update(tea.KeyMsg{Type: tea.KeyEscape})
			if cmd == nil {
				t.Fatal("expected command on Esc")
			}
			msg := cmd()
			if _, ok := msg.(OnboardingCompletedMsg); !ok {
				t.Errorf("expected OnboardingCompletedMsg on Esc at step %s, got %T", tt.name, msg)
			}
		})
	}
}

func TestOnboardingView_ValuesSkipOnEsc(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepValues
	ov.textInput.Focus()

	// Esc should skip to import step
	ov.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ov.step != stepImport {
		t.Errorf("step = %d, want stepImport after Esc on values", ov.step)
	}
}

func TestOnboardingView_ImportSkipOnEsc(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepImport
	ov.textInput.Focus()

	// Esc should skip to done step
	ov.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if ov.step != stepDone {
		t.Errorf("step = %d, want stepDone after Esc on import", ov.step)
	}
}

func TestOnboardingView_KeybindingsInteractive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		key        tea.KeyMsg
		triedKey   string
		lastAction string
	}{
		{
			name:       "left door with A",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			triedKey:   "left",
			lastAction: "Left door selected!",
		},
		{
			name:       "center door with W",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}},
			triedKey:   "up",
			lastAction: "Center door selected!",
		},
		{
			name:       "right door with D",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			triedKey:   "right",
			lastAction: "Right door selected!",
		},
		{
			name:       "reroll with S",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			triedKey:   "reroll",
			lastAction: "Doors re-rolled!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ov := NewOnboardingView()
			ov.step = stepKeybindings

			ov.Update(tt.key)

			if !ov.triedKeys[tt.triedKey] {
				t.Errorf("expected triedKeys[%q] = true", tt.triedKey)
			}
			if ov.lastAction != tt.lastAction {
				t.Errorf("lastAction = %q, want %q", ov.lastAction, tt.lastAction)
			}
		})
	}
}

func TestOnboardingView_TriedKeysCount(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepKeybindings
	ov.SetWidth(80)

	// Try all four keys
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeyRunes, Runes: []rune{'w'}},
		{Type: tea.KeyRunes, Runes: []rune{'d'}},
		{Type: tea.KeyRunes, Runes: []rune{'s'}},
	}

	for _, k := range keys {
		ov.Update(k)
	}

	if len(ov.triedKeys) != 4 {
		t.Errorf("triedKeys count = %d, want 4", len(ov.triedKeys))
	}

	view := ov.View()
	if !strings.Contains(view, "Tried 4 of 4") {
		t.Error("expected 'Tried 4 of 4' in view")
	}
}

func TestOnboardingView_SetWidth(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.SetWidth(120)

	if ov.width != 120 {
		t.Errorf("width = %d, want 120", ov.width)
	}
}

func TestOnboardingView_WelcomeContent(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.SetWidth(80)
	view := ov.View()

	expectedPhrases := []string{
		"Welcome to ThreeDoors",
		"three tasks",
		"no wrong answers",
		"Esc to skip",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(view, phrase) {
			t.Errorf("welcome view missing phrase: %q", phrase)
		}
	}
}

func TestOnboardingView_DoneContent(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepDone
	ov.SetWidth(80)
	view := ov.View()

	expectedPhrases := []string{
		"You're All Set",
		"Enter",
		"Search tasks",
		"Command palette",
		"progress over perfection",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(view, phrase) {
			t.Errorf("done view missing phrase: %q", phrase)
		}
	}
}

func TestOnboardingView_SpaceAdvances(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()

	// Space should advance from welcome to keybindings
	ov.Update(tea.KeyMsg{Type: tea.KeySpace})
	if ov.step != stepKeybindings {
		t.Errorf("step = %d, want stepKeybindings after space", ov.step)
	}
}

func TestOnboardingView_ValuesEntry(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepValues
	ov.textInput.Focus()
	ov.SetWidth(80)

	// Type a value and press Enter
	ov.textInput.SetValue("Be kind")
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if len(ov.values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(ov.values))
	}
	if ov.values[0] != "Be kind" {
		t.Errorf("value = %q, want %q", ov.values[0], "Be kind")
	}

	// Should still be on values step
	if ov.step != stepValues {
		t.Errorf("step = %d, want stepValues after adding value", ov.step)
	}

	// View should show the value
	view := ov.View()
	if !strings.Contains(view, "Be kind") {
		t.Error("values view should show entered value")
	}
}

func TestOnboardingView_ValuesMaxFive(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepValues
	ov.textInput.Focus()

	for i := 0; i < 5; i++ {
		ov.textInput.SetValue("value")
		ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}

	if len(ov.values) != 5 {
		t.Fatalf("expected 5 values, got %d", len(ov.values))
	}
	// Should auto-advance to import
	if ov.step != stepImport {
		t.Errorf("step = %d, want stepImport after 5 values", ov.step)
	}
}

func TestOnboardingView_ImportWithFile(t *testing.T) {
	t.Parallel()

	// Create a temp file with tasks
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(path, []byte("Task one\nTask two\nTask three"), 0o644); err != nil {
		t.Fatal(err)
	}

	ov := NewOnboardingView()
	ov.step = stepImport
	ov.textInput.Focus()
	ov.SetWidth(80)

	// Enter the file path
	ov.textInput.SetValue(path)
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should advance to preview
	if ov.step != stepImportPreview {
		t.Fatalf("step = %d, want stepImportPreview", ov.step)
	}
	if ov.importResult == nil {
		t.Fatal("expected importResult to be set")
	}
	if len(ov.importResult.Tasks) != 3 {
		t.Errorf("imported %d tasks, want 3", len(ov.importResult.Tasks))
	}

	// View should show preview
	view := ov.View()
	if !strings.Contains(view, "Import Preview") {
		t.Error("expected import preview view")
	}
	if !strings.Contains(view, "3 tasks") {
		t.Error("expected '3 tasks' in preview")
	}
}

func TestOnboardingView_ImportPreviewAccept(t *testing.T) {
	t.Parallel()

	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(path, []byte("Task one\nTask two"), 0o644); err != nil {
		t.Fatal(err)
	}

	ov := NewOnboardingView()
	ov.step = stepImport
	ov.textInput.Focus()
	ov.textInput.SetValue(path)
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Accept import
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if ov.step != stepDone {
		t.Errorf("step = %d, want stepDone after accepting import", ov.step)
	}
	if ov.importResult == nil {
		t.Error("expected importResult to remain after accept")
	}
}

func TestOnboardingView_ImportPreviewReject(t *testing.T) {
	t.Parallel()

	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(path, []byte("Task one\nTask two"), 0o644); err != nil {
		t.Fatal(err)
	}

	ov := NewOnboardingView()
	ov.step = stepImport
	ov.textInput.Focus()
	ov.textInput.SetValue(path)
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Reject import
	ov.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if ov.step != stepDone {
		t.Errorf("step = %d, want stepDone after rejecting import", ov.step)
	}
	if ov.importResult != nil {
		t.Error("expected importResult to be nil after reject")
	}
}

func TestOnboardingView_ImportInvalidPath(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepImport
	ov.textInput.Focus()
	ov.SetWidth(80)

	ov.textInput.SetValue("/nonexistent/path/tasks.txt")
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should stay on import step with error
	if ov.step != stepImport {
		t.Errorf("step = %d, want stepImport after invalid path", ov.step)
	}
	if ov.importError == "" {
		t.Error("expected importError to be set")
	}

	// Error should show in view (importError is rendered with styling,
	// so check for a substring of the error message)
	view := ov.View()
	if !strings.Contains(view, "nonexistent") {
		t.Error("expected error message in view")
	}
}

func TestOnboardingView_CompletedMsgCarriesData(t *testing.T) {
	t.Parallel()

	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(path, []byte("Task one\nTask two"), 0o644); err != nil {
		t.Fatal(err)
	}

	ov := NewOnboardingView()
	ov.step = stepValues
	ov.textInput.Focus()

	// Add a value
	ov.textInput.SetValue("Focus")
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Skip to import
	ov.textInput.SetValue("")
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Import tasks
	ov.textInput.SetValue(path)
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Accept
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Complete onboarding
	cmd := ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on final Enter")
	}

	msg := cmd()
	completed, ok := msg.(OnboardingCompletedMsg)
	if !ok {
		t.Fatalf("expected OnboardingCompletedMsg, got %T", msg)
	}

	if len(completed.Values) != 1 || completed.Values[0] != "Focus" {
		t.Errorf("Values = %v, want [Focus]", completed.Values)
	}
	if len(completed.ImportedTasks) != 2 {
		t.Errorf("ImportedTasks = %d, want 2", len(completed.ImportedTasks))
	}
}

func TestOnboardingView_DoneSummary(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()
	ov.step = stepDone
	ov.values = []string{"Be kind", "Stay focused"}
	ov.SetWidth(80)

	view := ov.View()
	if !strings.Contains(view, "2 values/goals saved") {
		t.Error("done view should show values summary")
	}
}
