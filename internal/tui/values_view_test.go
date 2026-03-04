package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewValuesSetupView(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	if vv.mode != ValuesSetupMode {
		t.Errorf("expected setup mode, got %d", vv.mode)
	}
}

func TestNewValuesEditView(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesEditView(cfg)
	if vv.mode != ValuesEditMode {
		t.Errorf("expected edit mode, got %d", vv.mode)
	}
}

func TestValuesSetupView_AddValue(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	vv.textInput.SetValue("Health and fitness")

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd after adding value (not at max yet)")
	}
	if len(cfg.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(cfg.Values))
	}
	if cfg.Values[0] != "Health and fitness" {
		t.Errorf("expected 'Health and fitness', got '%s'", cfg.Values[0])
	}
}

func TestValuesSetupView_EmptyEnterWithValues(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesSetupView(cfg)
	vv.textInput.SetValue("")

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected save cmd when pressing enter with empty input and existing values")
	}
	msg := cmd()
	if _, ok := msg.(ValuesSavedMsg); !ok {
		t.Errorf("expected ValuesSavedMsg, got %T", msg)
	}
}

func TestValuesSetupView_EmptyEnterNoValues(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	vv.textInput.SetValue("")

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when pressing enter with no values")
	}
}

func TestValuesSetupView_EscWithValues(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesSetupView(cfg)

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected cmd on Esc with values")
	}
	msg := cmd()
	if _, ok := msg.(ValuesSavedMsg); !ok {
		t.Errorf("expected ValuesSavedMsg, got %T", msg)
	}
}

func TestValuesSetupView_EscWithoutValues(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected cmd on Esc without values")
	}
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("expected ReturnToDoorsMsg, got %T", msg)
	}
}

func TestValuesEditView_DeleteValue(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"A", "B", "C"}}
	vv := NewValuesEditView(cfg)
	vv.selectedIndex = 1

	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if len(cfg.Values) != 2 {
		t.Fatalf("expected 2 values after delete, got %d", len(cfg.Values))
	}
	if cfg.Values[0] != "A" || cfg.Values[1] != "C" {
		t.Errorf("unexpected values: %v", cfg.Values)
	}
}

func TestValuesEditView_Navigate(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"A", "B", "C"}}
	vv := NewValuesEditView(cfg)
	vv.selectedIndex = 0

	vv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if vv.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1, got %d", vv.selectedIndex)
	}

	vv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if vv.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", vv.selectedIndex)
	}
}

func TestValuesEditView_Reorder(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"A", "B", "C"}}
	vv := NewValuesEditView(cfg)
	vv.selectedIndex = 0

	// Move down with J
	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	if cfg.Values[0] != "B" || cfg.Values[1] != "A" {
		t.Errorf("unexpected order after J: %v", cfg.Values)
	}
	if vv.selectedIndex != 1 {
		t.Errorf("expected selectedIndex 1 after J, got %d", vv.selectedIndex)
	}

	// Move up with K
	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	if cfg.Values[0] != "A" || cfg.Values[1] != "B" {
		t.Errorf("unexpected order after K: %v", cfg.Values)
	}
}

func TestValuesEditView_SaveOnEsc(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"A"}}
	vv := NewValuesEditView(cfg)

	cmd := vv.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected cmd on Esc")
	}
	msg := cmd()
	if _, ok := msg.(ValuesSavedMsg); !ok {
		t.Errorf("expected ValuesSavedMsg, got %T", msg)
	}
}

func TestRenderValuesFooter(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *core.ValuesConfig
		contains string
		empty    bool
	}{
		{"nil config", nil, "", true},
		{"empty values", &core.ValuesConfig{}, "", true},
		{"single value", &core.ValuesConfig{Values: []string{"Health"}}, "Health", false},
		{"multiple values", &core.ValuesConfig{Values: []string{"Health", "Family"}}, "·", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderValuesFooter(tt.cfg)
			if tt.empty && result != "" {
				t.Errorf("expected empty footer, got '%s'", result)
			}
			if !tt.empty && !strings.Contains(result, tt.contains) {
				t.Errorf("expected footer to contain '%s', got '%s'", tt.contains, result)
			}
		})
	}
}

func TestValuesSetupView_View(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	view := vv.View()

	if !strings.Contains(view, "Values & Goals Setup") {
		t.Error("expected setup header in view")
	}
	if !strings.Contains(view, "Define what matters most") {
		t.Error("expected setup instructions in view")
	}
}

func TestValuesEditView_View(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health", "Family"}}
	vv := NewValuesEditView(cfg)
	view := vv.View()

	if !strings.Contains(view, "Edit Values & Goals") {
		t.Error("expected edit header in view")
	}
	if !strings.Contains(view, "Health") {
		t.Error("expected values in view")
	}
}

func TestValuesView_SetWidth(t *testing.T) {
	t.Parallel()
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	vv.SetWidth(100)
	if vv.width != 100 {
		t.Errorf("expected width 100, got %d", vv.width)
	}
}

func TestValuesView_SetWidth_Small(t *testing.T) {
	t.Parallel()
	cfg := &core.ValuesConfig{}
	vv := NewValuesSetupView(cfg)
	vv.SetWidth(3)
	if vv.width != 3 {
		t.Errorf("expected width 3, got %d", vv.width)
	}
}

func TestValuesEditView_AddMode(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesEditView(cfg)

	// Press 'a' to enter add mode
	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !vv.textInput.Focused() {
		t.Error("expected textInput to be focused after 'a'")
	}

	// Type a value and press Enter
	vv.textInput.SetValue("Family")
	vv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(cfg.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(cfg.Values))
	}
	if vv.textInput.Focused() {
		t.Error("expected textInput to blur after enter")
	}
}

func TestValuesEditView_AddMode_Escape(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesEditView(cfg)

	// Press 'a' to enter add mode, then Escape
	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	vv.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if vv.textInput.Focused() {
		t.Error("expected textInput to blur after escape")
	}
	if len(cfg.Values) != 1 {
		t.Errorf("expected 1 value unchanged, got %d", len(cfg.Values))
	}
}

func TestValuesEditView_AddMode_EmptyEnter(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesEditView(cfg)

	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	vv.textInput.SetValue("")
	vv.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if len(cfg.Values) != 1 {
		t.Errorf("expected 1 value unchanged, got %d", len(cfg.Values))
	}
}

func TestValuesEditView_View_NoValues(t *testing.T) {
	cfg := &core.ValuesConfig{}
	vv := NewValuesEditView(cfg)
	view := vv.View()

	if !strings.Contains(view, "No values defined") {
		t.Error("expected 'No values defined' message")
	}
}

func TestValuesEditView_View_Focused(t *testing.T) {
	cfg := &core.ValuesConfig{Values: []string{"Health"}}
	vv := NewValuesEditView(cfg)
	vv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	view := vv.View()
	if !strings.Contains(view, "Enter to add") {
		t.Error("expected add mode help text")
	}
}
