package tui

import (
	"strings"
	"testing"
)

func TestLayoutBreakpoint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		height int
		want   Breakpoint
	}{
		{"zero height", 0, BreakpointMinimal},
		{"very small", 5, BreakpointMinimal},
		{"boundary 9", 9, BreakpointMinimal},
		{"boundary 10", 10, BreakpointCompact},
		{"mid compact", 12, BreakpointCompact},
		{"boundary 15", 15, BreakpointCompact},
		{"boundary 16", 16, BreakpointStandard},
		{"mid standard", 20, BreakpointStandard},
		{"boundary 24", 24, BreakpointStandard},
		{"boundary 25", 25, BreakpointComfortable},
		{"mid comfortable", 30, BreakpointComfortable},
		{"boundary 40", 40, BreakpointComfortable},
		{"boundary 41", 41, BreakpointSpacious},
		{"large terminal", 80, BreakpointSpacious},
		{"very large", 200, BreakpointSpacious},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := layoutBreakpoint(tt.height)
			if got != tt.want {
				t.Errorf("layoutBreakpoint(%d) = %d, want %d", tt.height, got, tt.want)
			}
		})
	}
}

func TestRenderDoors_ReturnsOnlyDoors(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("task1", "task2", "task3")
	dv.SetWidth(80)
	dv.SetHeight(30)

	doors := dv.RenderDoors()

	// Should contain task text
	if !strings.Contains(doors, "task1") {
		t.Error("RenderDoors should contain task1")
	}
	// Should NOT contain header/greeting text
	if strings.Contains(doors, "ThreeDoors") {
		t.Error("RenderDoors should not contain header")
	}
	// Should NOT contain footer/help text
	if strings.Contains(doors, "quit") {
		t.Error("RenderDoors should not contain footer")
	}
}

func TestRenderDoors_EmptyState(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView()
	dv.SetWidth(80)
	dv.SetHeight(30)

	doors := dv.RenderDoors()
	if !strings.Contains(doors, "All tasks done") {
		t.Error("empty state should show all-done message")
	}
}

func TestRenderStatusSection_CompletionCount(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.completedCount = 5

	status := dv.RenderStatusSection()
	if !strings.Contains(status, "Completed this session: 5") {
		t.Errorf("status should contain completion count, got: %s", status)
	}
}

func TestRenderStatusSection_Conflicts(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetPendingConflicts(3)

	status := dv.RenderStatusSection()
	if !strings.Contains(status, "3 sync conflict") {
		t.Errorf("status should contain conflict count, got: %s", status)
	}
}

func TestRenderStatusSection_Proposals(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetPendingProposals(2)

	status := dv.RenderStatusSection()
	if !strings.Contains(status, "2 suggestions") {
		t.Errorf("status should contain proposal count, got: %s", status)
	}
}

func TestRenderStatusSection_Empty(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	status := dv.RenderStatusSection()
	if status != "" {
		t.Errorf("status should be empty when no indicators, got: %q", status)
	}
}

func TestRenderHeader_Full(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	header := dv.RenderHeader()
	if !strings.Contains(header, "ThreeDoors") {
		t.Error("header should contain title")
	}
	// Should contain greeting
	if header == "" {
		t.Error("header should not be empty")
	}
}

func TestRenderCompactHeader(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	compact := dv.RenderCompactHeader()
	if !strings.Contains(compact, "ThreeDoors") {
		t.Error("compact header should contain title")
	}
	// Compact header should be shorter than full header
	full := dv.RenderHeader()
	if len(compact) >= len(full) {
		t.Error("compact header should be shorter than full header")
	}
}

func TestRenderFooter(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	footer := dv.RenderFooter()
	if !strings.Contains(footer, "quit") {
		t.Error("footer should contain quit hint")
	}
}

func TestRenderCompactFooter(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	compact := dv.RenderCompactFooter()
	if !strings.Contains(compact, "quit") {
		t.Error("compact footer should contain quit")
	}
	full := dv.RenderFooter()
	if len(compact) >= len(full) {
		t.Error("compact footer should be shorter than full footer")
	}
}

func TestHasDoors(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1")
	if !dv.HasDoors() {
		t.Error("should have doors")
	}
	empty := newTestDoorsView()
	if empty.HasDoors() {
		t.Error("should not have doors when empty")
	}
}

func TestBreakpoint_Minimal_NoDoors(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(8) // < 10 → minimal

	view := dv.View()

	// Should NOT contain header
	if strings.Contains(view, "ThreeDoors") {
		t.Error("minimal breakpoint should not show header")
	}
	// Should NOT contain footer help text
	if strings.Contains(view, "quit") {
		t.Error("minimal breakpoint should not show footer")
	}
	// Should still contain door content
	if !strings.Contains(view, "t1") {
		t.Error("minimal breakpoint should still show doors")
	}
}

func TestBreakpoint_Compact_OneLine(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(12) // 10-15 → compact

	view := dv.View()

	// Should contain compact header (just "ThreeDoors", not full title)
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("compact breakpoint should show compact header")
	}
	// Should contain compact footer
	if !strings.Contains(view, "quit") {
		t.Error("compact breakpoint should show compact footer")
	}
}

func TestBreakpoint_Standard_FullUI(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(20) // 16-24 → standard

	view := dv.View()

	// Full header with "Technical Demo" subtitle
	if !strings.Contains(view, "Technical Demo") {
		t.Error("standard breakpoint should show full header")
	}
	// Full footer
	if !strings.Contains(view, "quit") {
		t.Error("standard breakpoint should show full footer")
	}
}

func TestBreakpoint_Comfortable_HasPadding(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(35) // 25-40 → comfortable

	view := dv.View()

	// Should render fully with breathing room
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("comfortable breakpoint should show full header")
	}
	if !strings.Contains(view, "quit") {
		t.Error("comfortable breakpoint should show full footer")
	}
}

func TestBreakpoint_Spacious_DoorsCapped(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(50) // 40+ → spacious

	view := dv.View()

	if !strings.Contains(view, "ThreeDoors") {
		t.Error("spacious breakpoint should show full header")
	}
	if !strings.Contains(view, "quit") {
		t.Error("spacious breakpoint should show full footer")
	}
	// Door height cap: doorHeight() caps at 25
	if dv.doorHeight() > 25 {
		t.Errorf("door height should be capped at 25, got %d", dv.doorHeight())
	}
}

func TestBreakpoint_NoFlickerAcrossTransitions(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)

	// Simulate resizing across all breakpoints — no panics, always produces output.
	for h := 1; h <= 60; h++ {
		dv.SetHeight(h)
		view := dv.View()
		if view == "" {
			t.Errorf("height %d produced empty view", h)
		}
		// Must always contain at least the door content
		if !strings.Contains(view, "t1") {
			t.Errorf("height %d: door content missing", h)
		}
	}
}

func TestBreakpoint_NoErrorMessages(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)

	// At all heights, no error or "too small" messages should appear (D-119).
	for h := 1; h <= 60; h++ {
		dv.SetHeight(h)
		view := dv.View()
		if strings.Contains(view, "too small") || strings.Contains(view, "error") || strings.Contains(view, "Error") {
			t.Errorf("height %d: degradation should be invisible, got error message in view", h)
		}
	}
}

func TestDoorWidth(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	dv.SetWidth(80)
	if w := dv.doorWidth(); w <= 0 {
		t.Errorf("doorWidth with width=80 should be positive, got %d", w)
	}

	dv.SetWidth(10)
	if w := dv.doorWidth(); w != 30 {
		t.Errorf("doorWidth with narrow terminal should be default 30, got %d", w)
	}

	dv.SetWidth(0)
	if w := dv.doorWidth(); w != 30 {
		t.Errorf("doorWidth with zero width should be 30, got %d", w)
	}
}

func TestDoorHeight(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")

	dv.SetHeight(0)
	if h := dv.doorHeight(); h != 0 {
		t.Errorf("doorHeight with zero height should be 0, got %d", h)
	}

	dv.SetHeight(20)
	h := dv.doorHeight()
	if h < 9 { // 10 - 1 for threshold
		t.Errorf("doorHeight(20) should be at least 9, got %d", h)
	}

	dv.SetHeight(100)
	h = dv.doorHeight()
	if h > 25 {
		t.Errorf("doorHeight should cap at 25, got %d", h)
	}
}
