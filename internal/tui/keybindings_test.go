package tui

import (
	"fmt"
	"testing"
)

// allViewModes returns every ViewMode constant for exhaustive testing.
func allViewModes() []ViewMode {
	return []ViewMode{
		ViewDoors, ViewDetail, ViewMood, ViewSearch, ViewHealth,
		ViewAddTask, ViewValuesGoals, ViewFeedback,
		ViewNextSteps, ViewAvoidancePrompt, ViewInsights, ViewOnboarding,
		ViewConflict, ViewSyncLog, ViewThemePicker, ViewDevQueue, ViewProposals,
		ViewHelp, ViewDeferred, ViewSnooze, ViewSyncLogDetail,
	}
}

func TestViewKeyBindings_AllModesRegistered(t *testing.T) {
	t.Parallel()
	for _, mode := range allViewModes() {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			t.Parallel()
			groups := viewKeyBindings(mode, false)
			if len(groups) == 0 {
				t.Errorf("ViewMode %d has no binding groups", mode)
			}
			total := 0
			for _, g := range groups {
				total += len(g.Bindings)
			}
			if total == 0 {
				t.Errorf("ViewMode %d has groups but no bindings", mode)
			}
		})
	}
}

func TestViewKeyBindings_AllModesHaveHelp(t *testing.T) {
	t.Parallel()
	modes := allViewModes()
	// Also test doors-selected variant.
	type testCase struct {
		name         string
		mode         ViewMode
		doorSelected bool
	}
	var tests []testCase
	for _, m := range modes {
		tests = append(tests, testCase{
			name:         fmt.Sprintf("mode_%d", m),
			mode:         m,
			doorSelected: false,
		})
	}
	tests = append(tests, testCase{
		name:         "doors_selected",
		mode:         ViewDoors,
		doorSelected: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups := viewKeyBindings(tt.mode, tt.doorSelected)
			found := false
			for _, g := range groups {
				for _, b := range g.Bindings {
					if b.Key == "?" {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				t.Errorf("ViewMode %d (doorSelected=%v) missing ? binding", tt.mode, tt.doorSelected)
			}
		})
	}
}

func TestBarBindings_MaxEight(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name         string
		mode         ViewMode
		doorSelected bool
	}
	var tests []testCase
	for _, m := range allViewModes() {
		tests = append(tests, testCase{
			name:         fmt.Sprintf("mode_%d", m),
			mode:         m,
			doorSelected: false,
		})
	}
	tests = append(tests, testCase{
		name:         "doors_selected",
		mode:         ViewDoors,
		doorSelected: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bindings := barBindings(tt.mode, tt.doorSelected)
			if len(bindings) > 8 {
				t.Errorf("ViewMode %d (doorSelected=%v) has %d priority-1 bindings, max is 8",
					tt.mode, tt.doorSelected, len(bindings))
			}
		})
	}
}

func TestBarBindings_OnlyPriorityOne(t *testing.T) {
	t.Parallel()
	for _, mode := range allViewModes() {
		t.Run(fmt.Sprintf("mode_%d", mode), func(t *testing.T) {
			t.Parallel()
			for _, b := range barBindings(mode, false) {
				if b.Priority != PriorityAlways {
					t.Errorf("barBindings returned non-priority-1 binding: %s (%s) with priority %d",
						b.Key, b.Description, b.Priority)
				}
			}
		})
	}
}

func TestViewKeyBindings_NoDuplicateKeysPerView(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name         string
		mode         ViewMode
		doorSelected bool
	}
	var tests []testCase
	for _, m := range allViewModes() {
		tests = append(tests, testCase{
			name:         fmt.Sprintf("mode_%d", m),
			mode:         m,
			doorSelected: false,
		})
	}
	tests = append(tests, testCase{
		name:         "doors_selected",
		mode:         ViewDoors,
		doorSelected: true,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			seen := make(map[string]bool)
			for _, g := range viewKeyBindings(tt.mode, tt.doorSelected) {
				for _, b := range g.Bindings {
					if seen[b.Key] {
						t.Errorf("duplicate key %q in ViewMode %d (doorSelected=%v)",
							b.Key, tt.mode, tt.doorSelected)
					}
					seen[b.Key] = true
				}
			}
		})
	}
}

func TestAllKeyBindingGroups_CoversAllViews(t *testing.T) {
	t.Parallel()
	allGroups := allKeyBindingGroups()
	if len(allGroups) == 0 {
		t.Fatal("allKeyBindingGroups returned no groups")
	}

	// Collect all key-description pairs from allKeyBindingGroups.
	allPairs := make(map[string]bool)
	for _, g := range allGroups {
		for _, b := range g.Bindings {
			allPairs[b.Key+":"+b.Description] = true
		}
	}

	// Verify every per-view binding appears in the global set.
	for _, mode := range allViewModes() {
		for _, doorSelected := range []bool{false, true} {
			for _, g := range viewKeyBindings(mode, doorSelected) {
				for _, b := range g.Bindings {
					pair := b.Key + ":" + b.Description
					if !allPairs[pair] {
						t.Errorf("binding %s (%s) from ViewMode %d not in allKeyBindingGroups",
							b.Key, b.Description, mode)
					}
				}
			}
		}
	}
}

func TestAllKeyBindingGroups_NoDuplicatesWithinGroup(t *testing.T) {
	t.Parallel()
	for _, g := range allKeyBindingGroups() {
		seen := make(map[string]bool)
		for _, b := range g.Bindings {
			pair := b.Key + ":" + b.Description
			if seen[pair] {
				t.Errorf("duplicate key-description %q in group %q", pair, g.Name)
			}
			seen[pair] = true
		}
	}
}

func TestAllKeyBindingGroups_HasExpectedCategories(t *testing.T) {
	t.Parallel()
	groups := allKeyBindingGroups()
	expected := map[string]bool{
		"Navigation": false,
		"Actions":    false,
		"Display":    false,
		"Commands":   false,
	}
	for _, g := range groups {
		if _, ok := expected[g.Name]; ok {
			expected[g.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected category %q not found in allKeyBindingGroups", name)
		}
	}
}

func TestDoorsView_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		doorSelected bool
		wantKeys     []string
	}{
		{
			name:         "no selection",
			doorSelected: false,
			wantKeys:     []string{"a/w/d", "s", "n", ":", "?"},
		},
		{
			name:         "selected",
			doorSelected: true,
			wantKeys:     []string{"enter", "a/w/d", "esc", "?"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups := viewKeyBindings(ViewDoors, tt.doorSelected)
			keys := collectKeys(groups)
			for _, want := range tt.wantKeys {
				if !keys[want] {
					t.Errorf("DoorsView (selected=%v) missing key %q", tt.doorSelected, want)
				}
			}
		})
	}
}

func TestDetailView_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	groups := viewKeyBindings(ViewDetail, false)
	keys := collectKeys(groups)
	for _, want := range []string{"q/esc/space/enter", "c", "b", "e", "f", "?"} {
		if !keys[want] {
			t.Errorf("DetailView missing key %q", want)
		}
	}
}

func TestTextInputViews_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		mode ViewMode
	}{
		{"AddTask", ViewAddTask},
		{"Search", ViewSearch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groups := viewKeyBindings(tt.mode, false)
			keys := collectKeys(groups)
			for _, want := range []string{"esc", "?"} {
				if !keys[want] {
					t.Errorf("%s missing key %q", tt.name, want)
				}
			}
			// AddTask uses "enter"/"submit", Search uses "enter"/"select".
			if !keys["enter"] {
				t.Errorf("%s missing key %q", tt.name, "enter")
			}
		})
	}
}

func TestPriorityConstants(t *testing.T) {
	t.Parallel()
	if PriorityAlways != 1 {
		t.Errorf("PriorityAlways = %d, want 1", PriorityAlways)
	}
	if PriorityIfSpace != 2 {
		t.Errorf("PriorityIfSpace = %d, want 2", PriorityIfSpace)
	}
	if PriorityOverlay != 3 {
		t.Errorf("PriorityOverlay = %d, want 3", PriorityOverlay)
	}
}

// collectKeys extracts all key strings from binding groups into a set.
func collectKeys(groups []KeyBindingGroup) map[string]bool {
	keys := make(map[string]bool)
	for _, g := range groups {
		for _, b := range g.Bindings {
			keys[b.Key] = true
		}
	}
	return keys
}

func TestDetailView_HasAllKeyHandlerBindings(t *testing.T) {
	t.Parallel()
	groups := viewKeyBindings(ViewDetail, false)
	keys := collectKeys(groups)
	// All keys from handleDetailKeys must be registered.
	for _, want := range []string{
		"q/esc/space/enter", "c", "b", "i", "e", "f", "p", "r",
		"m", "l", "x", "z", "g", "+", "-", "u", "d", "y", "?",
	} {
		if !keys[want] {
			t.Errorf("DetailView missing key %q", want)
		}
	}
}

func TestHelpView_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	groups := viewKeyBindings(ViewHelp, false)
	keys := collectKeys(groups)
	for _, want := range []string{"q/esc", "j/k", "?"} {
		if !keys[want] {
			t.Errorf("HelpView missing key %q", want)
		}
	}
}

func TestDeferredView_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	groups := viewKeyBindings(ViewDeferred, false)
	keys := collectKeys(groups)
	for _, want := range []string{"q/esc", "j/k", "u", "?"} {
		if !keys[want] {
			t.Errorf("DeferredView missing key %q", want)
		}
	}
}

func TestSnoozeView_HasRequiredBindings(t *testing.T) {
	t.Parallel()
	groups := viewKeyBindings(ViewSnooze, false)
	keys := collectKeys(groups)
	for _, want := range []string{"↑/↓", "enter", "esc", "?"} {
		if !keys[want] {
			t.Errorf("SnoozeView missing key %q", want)
		}
	}
}

func TestContextBarBindings_DetailSubModes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		mode     DetailViewMode
		wantKeys []string
	}{
		{"normal", DetailModeView, []string{"q/esc/space/enter", "c", "e", "?"}},
		{"blocker input", DetailModeBlockerInput, []string{"enter", "esc"}},
		{"expand input", DetailModeExpandInput, []string{"enter", "esc"}},
		{"dispatch confirm", DetailModeDispatchConfirm, []string{"y", "n"}},
		{"link select", DetailModeLinkSelect, []string{"↑/↓", "enter", "esc"}},
		{"dep browse", DetailModeDepBrowse, []string{"↑/↓", "esc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := BarContext{Mode: ViewDetail, DetailMode: tt.mode}
			bindings := contextBarBindings(ctx)
			keys := make(map[string]bool)
			for _, b := range bindings {
				keys[b.Key] = true
			}
			for _, want := range tt.wantKeys {
				if !keys[want] {
					t.Errorf("DetailMode %d missing key %q in bar", tt.mode, want)
				}
			}
		})
	}
}

func TestContextBarBindings_CommandMode(t *testing.T) {
	t.Parallel()
	ctx := BarContext{Mode: ViewSearch, CommandMode: true}
	bindings := contextBarBindings(ctx)
	keys := make(map[string]bool)
	for _, b := range bindings {
		keys[b.Key] = true
	}
	if !keys["enter"] {
		t.Error("command mode bar missing 'enter'")
	}
	if !keys["esc"] {
		t.Error("command mode bar missing 'esc'")
	}
	// Should NOT show search navigation keys
	if keys["↑/↓"] {
		t.Error("command mode bar should not show search navigation keys")
	}
}

func TestContextBarBindings_FallbackToNormal(t *testing.T) {
	t.Parallel()
	// Non-sub-mode contexts should fall back to normal barBindings.
	ctx := BarContext{Mode: ViewDoors, DoorSelected: false}
	bindings := contextBarBindings(ctx)
	normalBindings := barBindings(ViewDoors, false)
	if len(bindings) != len(normalBindings) {
		t.Errorf("fallback binding count mismatch: got %d, want %d", len(bindings), len(normalBindings))
	}
}

func TestCommandBindingGroup_HasAllCommands(t *testing.T) {
	t.Parallel()
	group := commandBindingGroup()
	if group.Name != "Commands" {
		t.Errorf("command group name = %q, want %q", group.Name, "Commands")
	}
	wantCommands := []string{
		":add", ":mood", ":stats", ":health", ":dashboard",
		":goals", ":synclog", ":theme", ":deferred", ":devqueue",
		":suggestions", ":help", ":quit",
	}
	keys := make(map[string]bool)
	for _, b := range group.Bindings {
		keys[b.Key] = true
	}
	for _, want := range wantCommands {
		if !keys[want] {
			t.Errorf("command group missing %q", want)
		}
	}
}

func TestAllKeyBindingGroups_CommandsSectionPresent(t *testing.T) {
	t.Parallel()
	groups := allKeyBindingGroups()
	found := false
	for _, g := range groups {
		if g.Name == "Commands" {
			found = true
			if len(g.Bindings) == 0 {
				t.Error("Commands group has no bindings")
			}
			break
		}
	}
	if !found {
		t.Error("Commands section not found in allKeyBindingGroups")
	}
}

func TestAllKeyBindingGroups_CommandsIsLast(t *testing.T) {
	t.Parallel()
	groups := allKeyBindingGroups()
	if len(groups) == 0 {
		t.Fatal("no groups")
	}
	last := groups[len(groups)-1]
	if last.Name != "Commands" {
		t.Errorf("last group = %q, want Commands", last.Name)
	}
}
