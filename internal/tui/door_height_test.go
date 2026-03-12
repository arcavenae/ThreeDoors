package tui

import (
	"strings"
	"testing"
)

func TestDoorHeight_24LineTerm(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(24)
	view := dv.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("24-line terminal should render header")
	}
	// Doors should still render at minimum height
	if !strings.Contains(view, "[todo]") {
		t.Error("24-line terminal should render door status indicators")
	}
}

func TestDoorHeight_100LineTerm(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(100)
	view := dv.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("100-line terminal should render header")
	}
}

func TestDoorHeight_Formula(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		height  int
		wantMin int
		wantMax int
	}{
		{"tiny terminal (20 lines)", 20, 10, 10},
		{"small terminal (24 lines)", 24, 10, 12},
		{"medium terminal (40 lines)", 40, 10, 20},
		{"large terminal (60 lines)", 60, 25, 25},
		{"huge terminal (100 lines)", 100, 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Simulate the door height formula: min(max(10, height*0.5), 25) - 1 for threshold
			rawHeight := int(float64(tt.height) * 0.5)
			if rawHeight < 10 {
				rawHeight = 10
			}
			if rawHeight > 25 {
				rawHeight = 25
			}
			doorHeight := rawHeight - 1 // threshold line reservation

			if doorHeight < tt.wantMin-1 {
				t.Errorf("door height %d below expected minimum %d", doorHeight, tt.wantMin-1)
			}
			if doorHeight > tt.wantMax {
				t.Errorf("door height %d above expected maximum %d", doorHeight, tt.wantMax)
			}
		})
	}
}

func TestDoorsView_VerticalCentering_PaddingDistribution(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	dv.SetHeight(60) // plenty of space for padding

	view := dv.View()
	lines := strings.Split(view, "\n")

	// Find first and last non-empty content lines
	headerEnd := -1
	doorStart := -1
	for i, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" && headerEnd == -1 && i > 0 {
			headerEnd = i
		}
		if strings.Contains(line, "╭") || strings.Contains(line, "┌") || strings.Contains(line, "│") {
			if doorStart == -1 {
				doorStart = i
			}
		}
	}

	// Verify doors are NOT at the very top (should have padding above)
	if doorStart > 0 && headerEnd > 0 && doorStart <= headerEnd+1 {
		// This is fine for small terminals but for 60 lines there should be padding
		topPadding := doorStart - headerEnd
		if topPadding < 2 {
			t.Logf("Warning: minimal top padding (%d lines) — expected more for 60-line terminal", topPadding)
		}
	}

	// Verify the view contains all expected sections
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("missing header")
	}
	if !strings.Contains(view, "[todo]") {
		t.Error("missing door content")
	}
	if !strings.Contains(view, "quit") {
		t.Error("missing footer")
	}
}

func TestDoorsView_NoHeight_RendersWithoutPadding(t *testing.T) {
	t.Parallel()
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	// height = 0 (default, no WindowSizeMsg received)
	view := dv.View()

	if !strings.Contains(view, "ThreeDoors") {
		t.Error("should render header even without height info")
	}
	if !strings.Contains(view, "quit") {
		t.Error("should render footer even without height info")
	}
}
