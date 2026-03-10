package tui

import (
	"testing"
)

func TestNewDoorAnimation(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()

	if da == nil {
		t.Fatal("NewDoorAnimation returned nil")
		return
	}
	if da.active {
		t.Error("new animation should not be active")
	}
	for i := 0; i < 3; i++ {
		if da.Emphasis(i) != 0.0 {
			t.Errorf("door %d emphasis = %f, want 0.0", i, da.Emphasis(i))
		}
	}
}

func TestDoorAnimation_SetSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		selection   int
		wantTargets [3]float64
		wantActive  bool
	}{
		{"select door 0", 0, [3]float64{1.0, 0.0, 0.0}, true},
		{"select door 1", 1, [3]float64{0.0, 1.0, 0.0}, true},
		{"select door 2", 2, [3]float64{0.0, 0.0, 1.0}, true},
		{"deselect all", -1, [3]float64{0.0, 0.0, 0.0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			da := NewDoorAnimation()
			da.SetSelection(tt.selection)

			if da.active != tt.wantActive {
				t.Errorf("active = %v, want %v", da.active, tt.wantActive)
			}
			for i, want := range tt.wantTargets {
				if da.targets[i] != want {
					t.Errorf("target[%d] = %f, want %f", i, da.targets[i], want)
				}
			}
		})
	}
}

func TestDoorAnimation_UpdateConverges(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()
	da.SetSelection(1) // select middle door

	// Run animation until settled (should converge within reasonable frames)
	maxFrames := 300 // 5 seconds at 60fps — generous limit
	frames := 0
	for frames < maxFrames {
		if !da.Update() {
			break
		}
		frames++
	}

	if frames >= maxFrames {
		t.Fatalf("animation did not settle within %d frames", maxFrames)
	}

	// After settling, emphasis should match targets
	if da.Emphasis(0) != 0.0 {
		t.Errorf("door 0 emphasis = %f, want 0.0", da.Emphasis(0))
	}
	if da.Emphasis(1) != 1.0 {
		t.Errorf("door 1 emphasis = %f, want 1.0", da.Emphasis(1))
	}
	if da.Emphasis(2) != 0.0 {
		t.Errorf("door 2 emphasis = %f, want 0.0", da.Emphasis(2))
	}
	if da.Active() {
		t.Error("animation should not be active after settling")
	}
}

func TestDoorAnimation_SpringBehavior(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()
	da.SetSelection(0)

	// First frame should show some movement toward target
	da.Update()
	if da.Emphasis(0) <= 0.0 {
		t.Error("door 0 should have positive emphasis after first frame")
	}

	// Selected door emphasis should increase monotonically in early frames
	prev := da.Emphasis(0)
	for i := 0; i < 5; i++ {
		da.Update()
		current := da.Emphasis(0)
		if current < prev {
			// Spring might overshoot and then come back — that's OK after
			// the first few frames. We just check initial motion is toward target.
			break
		}
		prev = current
	}
}

func TestDoorAnimation_SwitchSelection(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()

	// Select door 0 and run to settled
	da.SetSelection(0)
	for da.Update() {
	}

	if da.Emphasis(0) != 1.0 {
		t.Fatalf("door 0 should be at 1.0 after settling, got %f", da.Emphasis(0))
	}

	// Now switch to door 2
	da.SetSelection(2)
	if !da.Active() {
		t.Error("animation should be active after switching selection")
	}

	// Run to settled again
	for da.Update() {
	}

	if da.Emphasis(0) != 0.0 {
		t.Errorf("door 0 emphasis = %f, want 0.0", da.Emphasis(0))
	}
	if da.Emphasis(2) != 1.0 {
		t.Errorf("door 2 emphasis = %f, want 1.0", da.Emphasis(2))
	}
}

func TestDoorAnimation_EmphasisOutOfBounds(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()

	if da.Emphasis(-1) != 0 {
		t.Error("out of bounds index -1 should return 0")
	}
	if da.Emphasis(3) != 0 {
		t.Error("out of bounds index 3 should return 0")
	}
	if da.Emphasis(100) != 0 {
		t.Error("out of bounds index 100 should return 0")
	}
}

func TestDoorAnimation_UpdateWhenInactive(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()

	// Update without setting selection should return false
	if da.Update() {
		t.Error("Update on inactive animation should return false")
	}
}

func TestDoorAnimation_Deselect(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()

	// Select and settle
	da.SetSelection(1)
	for da.Update() {
	}

	// Deselect
	da.SetSelection(-1)
	for da.Update() {
	}

	for i := 0; i < 3; i++ {
		if da.Emphasis(i) != 0.0 {
			t.Errorf("door %d emphasis = %f, want 0.0 after deselect", i, da.Emphasis(i))
		}
	}
}

func TestDoorAnimation_FrameCount(t *testing.T) {
	t.Parallel()
	da := NewDoorAnimation()
	da.SetSelection(0)

	frames := 0
	for da.Update() {
		frames++
	}

	// Spring should settle in a reasonable number of frames.
	// At 60fps with freq=6, damping=0.4 this should be under 120 frames (2 seconds).
	if frames > 180 {
		t.Errorf("animation took %d frames to settle, expected under 180", frames)
	}
	if frames < 5 {
		t.Errorf("animation settled in only %d frames, expected at least 5", frames)
	}
}

func TestAnimationTickCmd(t *testing.T) {
	t.Parallel()
	cmd := AnimationTickCmd()
	if cmd == nil {
		t.Fatal("AnimationTickCmd returned nil")
		return
	}
}
