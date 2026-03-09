package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

func TestDoorsViewHintKeysPassedToRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		hintsEnabled  bool
		selectedDoor  int
		wantHintKeys  []string // substrings expected in output
		wantNoHintKey bool     // when true, none of the hint keys should appear
	}{
		{
			name:         "hints enabled no selection shows all door keys",
			hintsEnabled: true,
			selectedDoor: -1,
			wantHintKeys: []string{"[a]", "[w]", "[d]"},
		},
		{
			name:         "hints enabled door selected shows all door keys",
			hintsEnabled: true,
			selectedDoor: 1,
			wantHintKeys: []string{"[a]", "[w]", "[d]"},
		},
		{
			name:          "hints disabled shows no door keys",
			hintsEnabled:  false,
			selectedDoor:  -1,
			wantNoHintKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dv := newHintsTestDoorsView(t)
			dv.SetShowKeyHints(tt.hintsEnabled)
			dv.selectedDoorIndex = tt.selectedDoor
			dv.SetThemeByName("classic")

			out := dv.View()

			if tt.wantNoHintKey {
				for _, key := range []string{"[a]", "[w]", "[d]"} {
					if strings.Contains(out, key) {
						t.Errorf("expected no hint key %q in output when disabled, but found it", key)
					}
				}
				return
			}

			for _, key := range tt.wantHintKeys {
				if !strings.Contains(out, key) {
					t.Errorf("expected hint key %q in output, but not found", key)
				}
			}
		})
	}
}

func TestDoorsViewHintBrightnessWithSelection(t *testing.T) {
	t.Parallel()

	// Render with no selection
	dvNoSel := newHintsTestDoorsView(t)
	dvNoSel.SetShowKeyHints(true)
	dvNoSel.selectedDoorIndex = -1
	dvNoSel.SetThemeByName("classic")
	outNoSel := dvNoSel.View()

	// Render with door 0 selected
	dvSel := newHintsTestDoorsView(t)
	dvSel.SetShowKeyHints(true)
	dvSel.selectedDoorIndex = 0
	dvSel.SetThemeByName("classic")
	outSel := dvSel.View()

	// Output should differ due to hint brightness changes
	if outNoSel == outSel {
		t.Error("expected different output for selected vs unselected doors due to hint brightness")
	}

	// Both should contain the hint keys
	for _, key := range []string{"[a]", "[w]", "[d]"} {
		if !strings.Contains(outNoSel, key) {
			t.Errorf("no-selection output missing hint key %q", key)
		}
		if !strings.Contains(outSel, key) {
			t.Errorf("selected output missing hint key %q", key)
		}
	}
}

func TestDoorsViewActionHints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hintsEnabled bool
		selectedDoor int
		wantAction   []string
		wantNoAction []string
	}{
		{
			name:         "hints enabled no selection shows re-roll and add task",
			hintsEnabled: true,
			selectedDoor: -1,
			wantAction:   []string{"[s]", "re-roll", "[n]", "add task"},
			wantNoAction: []string{"[enter]", "confirm"},
		},
		{
			name:         "hints enabled with selection shows confirm",
			hintsEnabled: true,
			selectedDoor: 1,
			wantAction:   []string{"[s]", "re-roll", "[n]", "add task", "[enter]", "confirm"},
		},
		{
			name:         "hints disabled shows no action hints",
			hintsEnabled: false,
			selectedDoor: -1,
			wantNoAction: []string{"[s]", "[n]", "[enter]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dv := newHintsTestDoorsView(t)
			dv.SetShowKeyHints(tt.hintsEnabled)
			dv.selectedDoorIndex = tt.selectedDoor

			out := dv.View()

			for _, want := range tt.wantAction {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q in output, not found", want)
				}
			}
			for _, noWant := range tt.wantNoAction {
				if strings.Contains(out, noWant) {
					t.Errorf("expected %q NOT in output, but found it", noWant)
				}
			}
		})
	}
}

func TestDoorsViewHelpTextSimplification(t *testing.T) {
	t.Parallel()

	fullHelpText := "a/left, w/up, d/right to select"
	shortHelpText := ": command"
	hintsIndicator := "hints: on"

	tests := []struct {
		name         string
		hintsEnabled bool
		wantFull     bool
	}{
		{
			name:         "hints enabled shows short help and hints indicator",
			hintsEnabled: true,
			wantFull:     false,
		},
		{
			name:         "hints disabled shows full help text",
			hintsEnabled: false,
			wantFull:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dv := newHintsTestDoorsView(t)
			dv.SetShowKeyHints(tt.hintsEnabled)

			out := dv.View()

			if tt.wantFull {
				if !strings.Contains(out, fullHelpText) {
					t.Error("expected full help text when hints disabled")
				}
				if strings.Contains(out, hintsIndicator) {
					t.Error("expected no hints indicator when hints disabled")
				}
			} else {
				if strings.Contains(out, fullHelpText) {
					t.Error("expected simplified help text when hints enabled")
				}
				if !strings.Contains(out, shortHelpText) {
					t.Error("expected short help text when hints enabled")
				}
				if !strings.Contains(out, hintsIndicator) {
					t.Error("expected hints indicator when hints enabled")
				}
			}
		})
	}
}

func TestRenderDoorHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		key          string
		enabled      bool
		isSelected   bool
		hasSelection bool
		wantEmpty    bool
		wantContains string
	}{
		{
			name:      "disabled returns empty",
			key:       "a",
			enabled:   false,
			wantEmpty: true,
		},
		{
			name:         "no selection normal brightness",
			key:          "a",
			enabled:      true,
			wantContains: "[a]",
		},
		{
			name:         "selected door bright",
			key:          "w",
			enabled:      true,
			isSelected:   true,
			hasSelection: true,
			wantContains: "[w]",
		},
		{
			name:         "unselected door dim",
			key:          "d",
			enabled:      true,
			isSelected:   false,
			hasSelection: true,
			wantContains: "[d]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := renderDoorHint(tt.key, tt.enabled, tt.isSelected, tt.hasSelection)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("expected empty, got %q", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantContains) {
				t.Errorf("expected output to contain %q, got %q", tt.wantContains, got)
			}
		})
	}
}

func TestRenderDoorHintBrightnessLevels(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)

	normal := renderDoorHint("a", true, false, false)
	bright := renderDoorHint("a", true, true, true)
	dim := renderDoorHint("a", true, false, true)

	// All should contain the key text
	for _, hint := range []string{normal, bright, dim} {
		if !strings.Contains(hint, "[a]") {
			t.Errorf("hint missing key text, got %q", hint)
		}
	}

	// All three should produce different styled output
	if normal == bright {
		t.Error("normal and bright should differ")
	}
	if normal == dim {
		t.Error("normal and dim should differ")
	}
	if bright == dim {
		t.Error("bright and dim should differ")
	}
}

// Golden file tests for doors view with inline hints (AC-6).
func TestGolden_DoorsViewHintsEnabled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	dv := newGoldenDoorsView(t, "Buy groceries", "Read Go book", "Exercise for 30 min")
	dv.SetShowKeyHints(true)
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_DoorsViewHintsSelected(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	dv := newGoldenDoorsView(t, "Buy groceries", "Read Go book", "Exercise for 30 min")
	dv.SetShowKeyHints(true)
	dv.selectedDoorIndex = 1
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_DoorsViewHintsDisabled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	dv := newGoldenDoorsView(t, "Buy groceries", "Read Go book", "Exercise for 30 min")
	dv.SetShowKeyHints(false)
	out := dv.View()
	golden.RequireEqual(t, []byte(out))
}

// newHintsTestDoorsView creates a DoorsView with deterministic state for hint unit tests.
// Always sets the classic theme since hints are only rendered by themes.
func newHintsTestDoorsView(t *testing.T) *DoorsView {
	t.Helper()

	pool := core.NewTaskPool()
	tasks := []*core.Task{
		core.NewTask("Task A"),
		core.NewTask("Task B"),
		core.NewTask("Task C"),
	}
	for _, task := range tasks {
		pool.AddTask(task)
	}

	tracker := core.NewSessionTracker()
	dv := NewDoorsView(pool, tracker)
	dv.width = 80
	dv.height = 24
	dv.currentDoors = tasks
	dv.greeting = "Test greeting"
	dv.footerMessage = "Test footer"
	dv.SetThemeByName("classic")
	return dv
}
