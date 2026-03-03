package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOnboardingView_StepProgression(t *testing.T) {
	t.Parallel()

	ov := NewOnboardingView()

	// Initial state should be welcome step
	view := ov.View()
	if !strings.Contains(view, "Welcome to ThreeDoors") {
		t.Error("expected welcome step on init")
	}
	if !strings.Contains(view, "Step 1 of 3") {
		t.Error("expected step 1 indicator")
	}

	// Press Enter to advance to keybindings
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "Key Bindings") {
		t.Error("expected keybindings step after Enter")
	}
	if !strings.Contains(view, "Step 2 of 3") {
		t.Error("expected step 2 indicator")
	}

	// Press Enter to advance to done
	ov.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view = ov.View()
	if !strings.Contains(view, "You're All Set") {
		t.Error("expected done step after Enter")
	}
	if !strings.Contains(view, "Step 3 of 3") {
		t.Error("expected step 3 indicator")
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
