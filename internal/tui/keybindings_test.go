package tui

import (
	"fmt"
	"testing"
)

// allViewModes returns every ViewMode constant for exhaustive testing.
func allViewModes() []ViewMode {
	return []ViewMode{
		ViewDoors, ViewDetail, ViewMood, ViewSearch, ViewHealth,
		ViewAddTask, ViewValuesGoals, ViewFeedback, ViewImprovement,
		ViewNextSteps, ViewAvoidancePrompt, ViewInsights, ViewOnboarding,
		ViewConflict, ViewSyncLog, ViewThemePicker, ViewDevQueue, ViewProposals,
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
	for _, want := range []string{"q/esc", "c", "b", "e", "f", "?"} {
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
