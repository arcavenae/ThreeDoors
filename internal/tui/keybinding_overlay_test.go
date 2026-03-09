package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

func setOverlayAsciiProfile(t *testing.T) {
	t.Helper()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })
}

func newTestOverlay(state OverlayState, width, height int) *KeybindingOverlay {
	return NewKeybindingOverlay(state, width, height)
}

func TestKeybindingOverlay_RendersAllGroups(t *testing.T) {
	setOverlayAsciiProfile(t)

	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 200)
	out := ko.View()

	allGroups := allKeyBindingGroups()
	for _, g := range allGroups {
		if !strings.Contains(out, g.Name) {
			t.Errorf("overlay missing group %q", g.Name)
		}
		for _, b := range g.Bindings {
			if !strings.Contains(out, b.Description) {
				t.Errorf("overlay missing binding description %q from group %q", b.Description, g.Name)
			}
		}
	}
}

func TestKeybindingOverlay_CurrentViewFirst(t *testing.T) {
	setOverlayAsciiProfile(t)

	tests := []struct {
		name     string
		mode     ViewMode
		wantName string
	}{
		{"doors view", ViewDoors, "Navigation"},
		{"detail view", ViewDetail, "Navigation"},
		{"conflict view", ViewConflict, "Actions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ko := newTestOverlay(OverlayState{ViewMode: tt.mode}, 80, 100)
			out := ko.View()

			if !strings.Contains(out, "(current)") {
				t.Error("overlay missing (current) marker")
			}

			expectedMarker := tt.wantName + " (current)"
			if !strings.Contains(out, expectedMarker) {
				t.Errorf("expected %q marker in overlay", expectedMarker)
			}
		})
	}
}

func TestKeybindingOverlay_ViewportScrolling(t *testing.T) {
	setOverlayAsciiProfile(t)

	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 15)
	initialView := ko.View()

	// Scroll down via viewport
	ko.Update(tea.KeyMsg{Type: tea.KeyDown})
	ko.Update(tea.KeyMsg{Type: tea.KeyDown})
	ko.Update(tea.KeyMsg{Type: tea.KeyDown})
	scrolledView := ko.View()

	if initialView == scrolledView {
		t.Error("scrolling did not change output")
	}
}

func TestKeybindingOverlay_MouseWheelEnabled(t *testing.T) {
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)
	if !ko.viewport.MouseWheelEnabled {
		t.Error("viewport should have mouse wheel enabled")
	}
}

func TestKeybindingOverlay_FooterAlwaysPresent(t *testing.T) {
	setOverlayAsciiProfile(t)

	tests := []struct {
		name   string
		height int
	}{
		{"normal", 60},
		{"small", 15},
		{"tiny", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, tt.height)
			out := ko.View()
			if !strings.Contains(out, "Press ? or esc to close") {
				t.Error("footer not present in overlay output")
			}
		})
	}
}

func TestKeybindingOverlay_ScrollIndicator(t *testing.T) {
	setOverlayAsciiProfile(t)

	// Small height forces scroll — should show "▼ more".
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 15)
	out := ko.View()
	if !strings.Contains(out, "▼ more") {
		t.Error("expected scroll-down indicator when content exceeds height")
	}

	// Scroll down partway — should show both indicators.
	for range 3 {
		ko.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	outMid := ko.View()
	if !strings.Contains(outMid, "▲") {
		t.Error("expected scroll-up indicator when scrolled down")
	}
}

func TestKeybindingOverlay_TooSmallTerminal(t *testing.T) {
	setOverlayAsciiProfile(t)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"too narrow", 15, 24},
		{"too short", 80, 3},
		{"both too small", 10, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, tt.width, tt.height)
			out := ko.View()
			if out != "" {
				t.Errorf("expected empty string for small terminal %dx%d, got output", tt.width, tt.height)
			}
		})
	}
}

func TestKeybindingOverlay_Title(t *testing.T) {
	setOverlayAsciiProfile(t)

	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 60)
	out := ko.View()
	if !strings.Contains(out, "KEYBINDING REFERENCE") {
		t.Error("overlay missing title")
	}
}

func TestKeybindingOverlay_BorderPresent(t *testing.T) {
	setOverlayAsciiProfile(t)

	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 60)
	out := ko.View()
	if !strings.Contains(out, "╔") || !strings.Contains(out, "╗") {
		t.Error("overlay missing double-line border")
	}
}

func TestReorderGroupsForView(t *testing.T) {
	t.Parallel()

	groups := []KeyBindingGroup{
		{Name: "Actions", Bindings: []KeyBinding{{Key: "x", Description: "test"}}},
		{Name: "Navigation", Bindings: []KeyBinding{{Key: "y", Description: "test"}}},
		{Name: "Display", Bindings: []KeyBinding{{Key: "z", Description: "test"}}},
	}

	reordered := reorderGroupsForView(groups, ViewDoors)
	if reordered[0].Name != "Navigation" {
		t.Errorf("expected Navigation first for ViewDoors, got %q", reordered[0].Name)
	}

	if len(reordered) != len(groups) {
		t.Errorf("expected %d groups, got %d", len(groups), len(reordered))
	}
}

func TestFormatBinding(t *testing.T) {
	setOverlayAsciiProfile(t)

	b := KeyBinding{Key: "esc", Description: "back", Priority: PriorityAlways}
	result := formatBinding(b)
	if !strings.Contains(result, "esc") {
		t.Error("formatted binding missing key")
	}
	if !strings.Contains(result, "back") {
		t.Error("formatted binding missing description")
	}
	if !strings.HasPrefix(result, "  ") {
		t.Error("formatted binding missing leading indent")
	}
}

// Golden file tests — deterministic output snapshots.

func TestGolden_Overlay_80x24(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)
	out := ko.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_Overlay_Scrolled(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 15)
	// Scroll down 5 lines via viewport
	for range 5 {
		ko.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	out := ko.View()
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_Overlay_Narrow(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 50, 24)
	out := ko.View()
	golden.RequireEqual(t, []byte(out))
}

func TestKeybindingOverlay_CommandsSectionPresent(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 200)
	out := ko.View()
	if !strings.Contains(out, "Commands") {
		t.Error("overlay missing Commands section")
	}
	// Verify some key commands are listed.
	for _, cmd := range []string{":add", ":mood", ":stats", ":health", ":deferred"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("overlay missing command %q", cmd)
		}
	}
}

func TestKeybindingOverlay_GroupOrdering(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 200)
	out := ko.View()

	// Navigation should appear first (marked as current for DoorsView).
	navIdx := strings.Index(out, "Navigation")
	actIdx := strings.Index(out, "Actions")
	dispIdx := strings.Index(out, "Display")
	cmdIdx := strings.Index(out, "Commands")

	if navIdx < 0 || actIdx < 0 || dispIdx < 0 || cmdIdx < 0 {
		t.Fatal("not all expected groups found in overlay output")
	}

	// Verify ordering: Navigation (current) > Actions > Display > Commands
	if navIdx >= actIdx {
		t.Error("Navigation should appear before Actions")
	}
	if actIdx >= dispIdx {
		t.Error("Actions should appear before Display")
	}
	if dispIdx >= cmdIdx {
		t.Error("Display should appear before Commands")
	}
}

func TestKeybindingOverlay_CommandsSectionSeparate(t *testing.T) {
	setOverlayAsciiProfile(t)
	ko := newTestOverlay(OverlayState{ViewMode: ViewDoors}, 80, 200)
	out := ko.View()

	// Commands section should be separate from key bindings sections.
	navIdx := strings.Index(out, "Navigation")
	cmdIdx := strings.Index(out, "Commands")
	if cmdIdx <= navIdx {
		t.Error("Commands should appear after Navigation (key bindings)")
	}
}

func TestKeybindingOverlay_ContextFromDifferentViews(t *testing.T) {
	setOverlayAsciiProfile(t)

	tests := []struct {
		name     string
		mode     ViewMode
		wantName string
	}{
		{"doors", ViewDoors, "Navigation"},
		{"detail", ViewDetail, "Navigation"},
		{"search", ViewSearch, "Navigation"},
		{"conflict", ViewConflict, "Actions"},
		{"help", ViewHelp, "Navigation"},
		{"deferred", ViewDeferred, "Navigation"},
		{"snooze", ViewSnooze, "Navigation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ko := newTestOverlay(OverlayState{ViewMode: tt.mode}, 80, 200)
			out := ko.View()
			marker := tt.wantName + " (current)"
			if !strings.Contains(out, marker) {
				t.Errorf("expected %q marker in overlay for %s", marker, tt.name)
			}
		})
	}
}
