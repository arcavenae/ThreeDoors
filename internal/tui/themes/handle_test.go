package themes

import (
	"testing"
)

func TestHandleCharForEmphasis_ForwardSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		emphasis float64
		selected bool
		frames   HandleFrames
		want     string
	}{
		{"rest at 0.0", 0.0, true, RoundKnobFrames, "●"},
		{"rest at 0.1", 0.1, true, RoundKnobFrames, "●"},
		{"rest at 0.29", 0.29, true, RoundKnobFrames, "●"},
		{"turning at 0.3", 0.3, true, RoundKnobFrames, "◐"},
		{"turning at 0.45", 0.45, true, RoundKnobFrames, "◐"},
		{"turning at 0.59", 0.59, true, RoundKnobFrames, "◐"},
		{"turned at 0.6", 0.6, true, RoundKnobFrames, "○"},
		{"turned at 0.8", 0.8, true, RoundKnobFrames, "○"},
		{"turned at 1.0", 1.0, true, RoundKnobFrames, "○"},
		{"turned at overshoot 1.1", 1.1, true, RoundKnobFrames, "○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HandleCharForEmphasis(tt.emphasis, tt.selected, tt.frames)
			if got != tt.want {
				t.Errorf("HandleCharForEmphasis(%f, %v, RoundKnobFrames) = %q, want %q",
					tt.emphasis, tt.selected, got, tt.want)
			}
		})
	}
}

func TestHandleCharForEmphasis_ReverseSequence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		emphasis float64
		selected bool
		frames   HandleFrames
		want     string
	}{
		{"turned at 1.0", 1.0, false, RoundKnobFrames, "○"},
		{"turned at 0.7", 0.7, false, RoundKnobFrames, "○"},
		{"turned at 0.6", 0.6, false, RoundKnobFrames, "○"},
		{"springback at 0.59", 0.59, false, RoundKnobFrames, "◑"},
		{"springback at 0.45", 0.45, false, RoundKnobFrames, "◑"},
		{"springback at 0.3", 0.3, false, RoundKnobFrames, "◑"},
		{"rest at 0.29", 0.29, false, RoundKnobFrames, "●"},
		{"rest at 0.0", 0.0, false, RoundKnobFrames, "●"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HandleCharForEmphasis(tt.emphasis, tt.selected, tt.frames)
			if got != tt.want {
				t.Errorf("HandleCharForEmphasis(%f, %v, RoundKnobFrames) = %q, want %q",
					tt.emphasis, tt.selected, got, tt.want)
			}
		})
	}
}

func TestHandleCharForEmphasis_Deterministic(t *testing.T) {
	t.Parallel()

	// Same emphasis and direction must always produce the same character.
	for i := 0; i < 100; i++ {
		got := HandleCharForEmphasis(0.45, true, RoundKnobFrames)
		if got != "◐" {
			t.Fatalf("iteration %d: expected ◐, got %q", i, got)
		}
	}
}

func TestHandleCharForEmphasis_PerThemeFrames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		frames   HandleFrames
		restChar string
		turnChar string
	}{
		{"round knob", RoundKnobFrames, "●", "○"},
		{"open knob", OpenKnobFrames, "○", "●"},
		{"square handle", SquareHandleFrames, "■", "□"},
		{"diamond handle", DiamondHandleFrames, "◆", "○"},
		{"scifi handle", SciFiHandleFrames, "◈", "○"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rest := HandleCharForEmphasis(0.0, false, tt.frames)
			if rest != tt.restChar {
				t.Errorf("rest = %q, want %q", rest, tt.restChar)
			}
			turned := HandleCharForEmphasis(1.0, true, tt.frames)
			if turned != tt.turnChar {
				t.Errorf("turned = %q, want %q", turned, tt.turnChar)
			}
		})
	}
}

func TestHandleCharForEmphasis_DirectionDifference(t *testing.T) {
	t.Parallel()

	// At emphasis 0.45, forward and reverse should produce different characters.
	forward := HandleCharForEmphasis(0.45, true, RoundKnobFrames)
	reverse := HandleCharForEmphasis(0.45, false, RoundKnobFrames)

	if forward == reverse {
		t.Errorf("forward and reverse at emphasis 0.45 should differ, both are %q", forward)
	}
	if forward != "◐" {
		t.Errorf("forward at 0.45 = %q, want ◐", forward)
	}
	if reverse != "◑" {
		t.Errorf("reverse at 0.45 = %q, want ◑", reverse)
	}
}

func TestHandleCharForEmphasis_NegativeEmphasis(t *testing.T) {
	t.Parallel()

	// Negative emphasis (spring undershoot) should return rest char.
	got := HandleCharForEmphasis(-0.1, true, RoundKnobFrames)
	if got != "●" {
		t.Errorf("negative emphasis = %q, want ●", got)
	}
}
