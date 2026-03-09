package tui

import "testing"

func TestNewScrollableView(t *testing.T) {
	t.Parallel()

	vp := NewScrollableView(80, 24)

	if vp.Width != 80 {
		t.Errorf("width = %d, want 80", vp.Width)
	}
	if vp.Height != 24 {
		t.Errorf("height = %d, want 24", vp.Height)
	}
	if !vp.MouseWheelEnabled {
		t.Error("mouse wheel should be enabled by default")
	}
}

func TestNewScrollableView_ConsistencyAcrossViews(t *testing.T) {
	t.Parallel()

	vp1 := NewScrollableView(80, 20)
	vp2 := NewScrollableView(80, 20)

	if vp1.MouseWheelEnabled != vp2.MouseWheelEnabled {
		t.Error("all viewports from factory should have same mouse wheel setting")
	}
	if vp1.Width != vp2.Width {
		t.Error("all viewports from factory with same args should have same width")
	}
	if vp1.Height != vp2.Height {
		t.Error("all viewports from factory with same args should have same height")
	}
}
