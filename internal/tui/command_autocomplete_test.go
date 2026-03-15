package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Command Registry Tests ---

func TestFilterCommands_ColonAloneShowsAll(t *testing.T) {
	t.Parallel()
	results := filterCommands("")
	if len(results) != len(commandRegistry) {
		t.Errorf("expected %d commands for empty prefix, got %d", len(commandRegistry), len(results))
	}
}

func TestFilterCommands_PrefixFiltering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		prefix   string
		wantCmds []string
	}{
		{
			name:     "a prefix matches add and add-ctx",
			prefix:   "a",
			wantCmds: []string{"add", "add-ctx"},
		},
		{
			name:     "d prefix matches dashboard deferred devqueue dispatch",
			prefix:   "d",
			wantCmds: []string{"dashboard", "deferred", "devqueue", "dispatch"},
		},
		{
			name:     "xyz matches nothing",
			prefix:   "xyz",
			wantCmds: nil,
		},
		{
			name:     "h prefix matches health help history",
			prefix:   "h",
			wantCmds: []string{"health", "help", "history"},
		},
		{
			name:     "exact match for quit",
			prefix:   "quit",
			wantCmds: []string{"quit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := filterCommands(tt.prefix)
			if len(got) != len(tt.wantCmds) {
				var gotNames []string
				for _, c := range got {
					gotNames = append(gotNames, c.Name)
				}
				t.Fatalf("filterCommands(%q) returned %d commands %v, want %d %v",
					tt.prefix, len(got), gotNames, len(tt.wantCmds), tt.wantCmds)
			}
			for i, want := range tt.wantCmds {
				if got[i].Name != want {
					t.Errorf("filterCommands(%q)[%d].Name = %q, want %q", tt.prefix, i, got[i].Name, want)
				}
			}
		})
	}
}

func TestFilterCommands_CaseInsensitive(t *testing.T) {
	t.Parallel()
	lower := filterCommands("a")
	upper := filterCommands("A")
	if len(lower) != len(upper) {
		t.Errorf("case-insensitive mismatch: 'a' got %d, 'A' got %d", len(lower), len(upper))
	}
}

func TestFilterCommands_AllHaveDescriptions(t *testing.T) {
	t.Parallel()
	for _, cmd := range commandRegistry {
		if cmd.Desc == "" {
			t.Errorf("command %q has empty description", cmd.Name)
		}
	}
}

// --- Autocomplete Navigation Tests ---

func TestSearchView_CommandAutocomplete_ArrowNavigation(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	// Should start at index 0
	if sv.commandSelectedIndex != 0 {
		t.Errorf("expected initial commandSelectedIndex 0, got %d", sv.commandSelectedIndex)
	}

	// Down moves to next
	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.commandSelectedIndex != 1 {
		t.Errorf("expected commandSelectedIndex 1 after down, got %d", sv.commandSelectedIndex)
	}

	// Up moves back
	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.commandSelectedIndex != 0 {
		t.Errorf("expected commandSelectedIndex 0 after up, got %d", sv.commandSelectedIndex)
	}
}

func TestSearchView_CommandAutocomplete_WrapNavigation(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":q") // Only "quit" matches
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	if len(sv.commandSuggestions) != 1 {
		t.Fatalf("expected 1 suggestion for :q, got %d", len(sv.commandSuggestions))
	}

	// Down from last wraps to first
	sv.commandSelectedIndex = 0
	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.commandSelectedIndex != 0 {
		t.Errorf("expected wrap to 0, got %d", sv.commandSelectedIndex)
	}

	// Up from first wraps to last
	sv.commandSelectedIndex = 0
	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.commandSelectedIndex != 0 {
		t.Errorf("expected wrap to 0 (single item), got %d", sv.commandSelectedIndex)
	}
}

func TestSearchView_CommandAutocomplete_WrapDown(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":a") // "add" and "add-ctx"
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	if len(sv.commandSuggestions) != 2 {
		t.Fatalf("expected 2 suggestions for :a, got %d", len(sv.commandSuggestions))
	}

	// Move to last
	sv.commandSelectedIndex = 1
	// Down from last wraps to first
	sv.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sv.commandSelectedIndex != 0 {
		t.Errorf("expected wrap to 0, got %d", sv.commandSelectedIndex)
	}
}

func TestSearchView_CommandAutocomplete_WrapUp(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":a") // "add" and "add-ctx"
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	// Up from first wraps to last
	sv.commandSelectedIndex = 0
	sv.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sv.commandSelectedIndex != 1 {
		t.Errorf("expected wrap to 1, got %d", sv.commandSelectedIndex)
	}
}

// --- Tab Completion Tests ---

func TestSearchView_CommandAutocomplete_TabCompletes(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":h")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	// First suggestion should be "health"
	if len(sv.commandSuggestions) < 1 {
		t.Fatal("expected suggestions for :h")
	}
	sv.commandSelectedIndex = 0
	expected := sv.commandSuggestions[0].Name

	sv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if sv.textInput.Value() != ":"+expected {
		t.Errorf("expected tab to complete to :%s, got %q", expected, sv.textInput.Value())
	}
}

func TestSearchView_CommandAutocomplete_TabSelectsHighlighted(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":h")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	// Navigate to second suggestion (help)
	sv.commandSelectedIndex = 1
	if sv.commandSelectedIndex >= len(sv.commandSuggestions) {
		t.Skip("only one suggestion for :h")
	}
	expected := sv.commandSuggestions[1].Name

	sv.Update(tea.KeyMsg{Type: tea.KeyTab})
	if sv.textInput.Value() != ":"+expected {
		t.Errorf("expected tab to complete to :%s, got %q", expected, sv.textInput.Value())
	}
}

// --- Backspace Updates Suggestions ---

func TestSearchView_CommandAutocomplete_BackspaceUpdatesSuggestions(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")

	// Type ":he" — should get health, help
	sv.textInput.SetValue(":he")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()
	narrowCount := len(sv.commandSuggestions)

	// Simulate backspace to ":h" — should still match health, help
	sv.textInput.SetValue(":h")
	sv.updateCommandSuggestions()
	broadCount := len(sv.commandSuggestions)

	if broadCount < narrowCount {
		t.Errorf("broader prefix should have >= suggestions: broad=%d, narrow=%d", broadCount, narrowCount)
	}

	// Backspace to ":" — should show all
	sv.textInput.SetValue(":")
	sv.updateCommandSuggestions()
	if len(sv.commandSuggestions) != len(commandRegistry) {
		t.Errorf("expected all %d commands for ':', got %d", len(commandRegistry), len(sv.commandSuggestions))
	}
}

// --- Enter Executes Highlighted Suggestion ---

func TestSearchView_CommandAutocomplete_EnterExecutesHighlighted(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.textInput.SetValue(":he")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	// Navigate to "help" (second suggestion)
	for i, s := range sv.commandSuggestions {
		if s.Name == "help" {
			sv.commandSelectedIndex = i
			break
		}
	}

	cmd := sv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on highlighted suggestion should return a command")
		return
	}
	msg := cmd()
	if _, ok := msg.(ShowHelpMsg); !ok {
		t.Errorf("expected ShowHelpMsg, got %T", msg)
	}
}

// --- View Rendering Tests ---

func TestSearchView_CommandAutocomplete_ViewShowsSuggestions(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	view := sv.View()

	// Should contain "Command mode"
	if !strings.Contains(view, "Command mode") {
		t.Error("View should show 'Command mode' indicator")
	}

	// Should contain at least some command names
	if !strings.Contains(view, ":add") {
		t.Error("View should show :add command")
	}
	if !strings.Contains(view, ":help") {
		t.Error("View should show :help command")
	}
}

func TestSearchView_CommandAutocomplete_ViewShowsDescriptions(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.textInput.SetValue(":")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	view := sv.View()

	// Should contain command descriptions
	if !strings.Contains(view, "Create a new task") {
		t.Error("View should show command descriptions")
	}
}

func TestSearchView_CommandAutocomplete_ViewShowsNoMatchingCommands(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.textInput.SetValue(":xyz")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	view := sv.View()
	if !strings.Contains(view, "No matching commands") {
		t.Error("View should show 'No matching commands' when no matches")
	}
}

func TestSearchView_CommandAutocomplete_FilteredViewShowsSubset(t *testing.T) {
	t.Parallel()
	sv := newTestSearchView("task1")
	sv.SetWidth(80)
	sv.textInput.SetValue(":a")
	sv.isCommandMode = true
	sv.updateCommandSuggestions()

	view := sv.View()

	// Should contain add commands
	if !strings.Contains(view, ":add") {
		t.Error("View should show :add for prefix 'a'")
	}
	// Should NOT contain non-matching commands
	if strings.Contains(view, ":quit") {
		t.Error("View should not show :quit for prefix 'a'")
	}
}

// --- Command Registry DRY Check ---

func TestCommandRegistry_CoversAllCommands(t *testing.T) {
	t.Parallel()
	// These are the commands from executeCommand that should be in the registry.
	// "exit" and "snoozed" are aliases and don't need separate entries.
	expectedCmds := []string{
		"add", "add-ctx", "mood", "stats", "health", "dashboard",
		"insights", "goals", "synclog", "tag", "theme", "deferred",
		"devqueue", "suggestions", "dispatch", "help", "quit",
	}

	registryNames := make(map[string]bool)
	for _, cmd := range commandRegistry {
		registryNames[cmd.Name] = true
	}

	for _, cmd := range expectedCmds {
		if !registryNames[cmd] {
			t.Errorf("command %q is in executeCommand but missing from commandRegistry", cmd)
		}
	}
}
