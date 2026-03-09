package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

func setOverlayAsciiProfile(t *testing.T) {
	t.Helper()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })
}

func TestRenderKeybindingOverlay_RendersAllGroups(t *testing.T) {
	setOverlayAsciiProfile(t)

	// Use large height to ensure all content visible.
	out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 100)

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

func TestRenderKeybindingOverlay_CurrentViewFirst(t *testing.T) {
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
			out := RenderKeybindingOverlay(OverlayState{ViewMode: tt.mode}, 80, 100)

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

func TestRenderKeybindingOverlay_ScrollOffset(t *testing.T) {
	setOverlayAsciiProfile(t)

	noScroll := RenderKeybindingOverlay(OverlayState{ScrollOffset: 0, ViewMode: ViewDoors}, 80, 15)
	scrolled := RenderKeybindingOverlay(OverlayState{ScrollOffset: 3, ViewMode: ViewDoors}, 80, 15)

	if noScroll == scrolled {
		t.Error("scrolling did not change output")
	}
}

func TestRenderKeybindingOverlay_ScrollClamping(t *testing.T) {
	setOverlayAsciiProfile(t)

	out := RenderKeybindingOverlay(OverlayState{ScrollOffset: 9999, ViewMode: ViewDoors}, 80, 60)
	if out == "" {
		t.Error("clamped scroll produced empty output")
	}

	outNeg := RenderKeybindingOverlay(OverlayState{ScrollOffset: -5, ViewMode: ViewDoors}, 80, 60)
	outZero := RenderKeybindingOverlay(OverlayState{ScrollOffset: 0, ViewMode: ViewDoors}, 80, 60)
	if outNeg != outZero {
		t.Error("negative scroll should clamp to 0")
	}
}

func TestRenderKeybindingOverlay_FooterAlwaysPresent(t *testing.T) {
	setOverlayAsciiProfile(t)

	tests := []struct {
		name   string
		scroll int
		height int
	}{
		{"no scroll", 0, 60},
		{"scrolled", 5, 15},
		{"max scroll", 9999, 15},
		{"small terminal", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := RenderKeybindingOverlay(OverlayState{ScrollOffset: tt.scroll, ViewMode: ViewDoors}, 80, tt.height)
			if !strings.Contains(out, "Press ? or esc to close") {
				t.Error("footer not present in overlay output")
			}
		})
	}
}

func TestRenderKeybindingOverlay_ScrollIndicator(t *testing.T) {
	setOverlayAsciiProfile(t)

	// Small height forces scroll — should show "▼ more".
	out := RenderKeybindingOverlay(OverlayState{ScrollOffset: 0, ViewMode: ViewDoors}, 80, 15)
	if !strings.Contains(out, "▼ more") {
		t.Error("expected scroll-down indicator when content exceeds height")
	}

	// Scrolled partway — should show both indicators.
	outMid := RenderKeybindingOverlay(OverlayState{ScrollOffset: 3, ViewMode: ViewDoors}, 80, 15)
	if !strings.Contains(outMid, "▲") {
		t.Error("expected scroll-up indicator when scrolled down")
	}
}

func TestRenderKeybindingOverlay_TooSmallTerminal(t *testing.T) {
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
			out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, tt.width, tt.height)
			if out != "" {
				t.Errorf("expected empty string for small terminal %dx%d, got output", tt.width, tt.height)
			}
		})
	}
}

func TestRenderKeybindingOverlay_Title(t *testing.T) {
	setOverlayAsciiProfile(t)

	out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 60)
	if !strings.Contains(out, "KEYBINDING REFERENCE") {
		t.Error("overlay missing title")
	}
}

func TestRenderKeybindingOverlay_BorderPresent(t *testing.T) {
	setOverlayAsciiProfile(t)

	out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 60)
	if !strings.Contains(out, "╔") || !strings.Contains(out, "╗") {
		t.Error("overlay missing double-line border")
	}
}

func TestClampScrollOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		offset   int
		height   int
		wantZero bool
	}{
		{"negative clamps to zero", -5, 60, true},
		{"zero stays zero", 0, 60, true},
		{"large clamps to max", 9999, 60, false},
		{"small height large offset", 100, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ClampScrollOffset(tt.offset, tt.height)
			if result < 0 {
				t.Errorf("ClampScrollOffset returned negative: %d", result)
			}
			if tt.wantZero && result != 0 {
				t.Errorf("expected 0, got %d", result)
			}
		})
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

func TestCountContentLines(t *testing.T) {
	t.Parallel()

	groups := []KeyBindingGroup{
		{Name: "A", Bindings: []KeyBinding{{Key: "x"}, {Key: "y"}}},
		{Name: "B", Bindings: []KeyBinding{{Key: "z"}}},
	}
	if got := countContentLines(groups); got != 6 {
		t.Errorf("countContentLines = %d, want 6", got)
	}
}

// Golden file tests — deterministic output snapshots.

func TestGolden_Overlay_80x24(t *testing.T) {
	setOverlayAsciiProfile(t)
	out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 80, 24)
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_Overlay_Scrolled(t *testing.T) {
	setOverlayAsciiProfile(t)
	out := RenderKeybindingOverlay(OverlayState{ScrollOffset: 5, ViewMode: ViewDoors}, 80, 15)
	golden.RequireEqual(t, []byte(out))
}

func TestGolden_Overlay_Narrow(t *testing.T) {
	setOverlayAsciiProfile(t)
	out := RenderKeybindingOverlay(OverlayState{ViewMode: ViewDoors}, 50, 24)
	golden.RequireEqual(t, []byte(out))
}
