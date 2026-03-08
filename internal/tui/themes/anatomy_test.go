package themes

import "testing"

func TestNewDoorAnatomy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"minimum height 5", 5},
		{"small height 8", 8},
		{"min theme height 10", 10},
		{"standard height 16", 16},
		{"tall height 24", 24},
		{"very tall 40", 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := NewDoorAnatomy(tt.height)

			// All within bounds [0, height-1]
			if a.LintelRow < 0 || a.LintelRow >= tt.height {
				t.Errorf("LintelRow %d out of bounds [0, %d)", a.LintelRow, tt.height)
			}
			if a.ContentStart < 0 || a.ContentStart >= tt.height {
				t.Errorf("ContentStart %d out of bounds [0, %d)", a.ContentStart, tt.height)
			}
			if a.PanelDivider < 0 || a.PanelDivider >= tt.height {
				t.Errorf("PanelDivider %d out of bounds [0, %d)", a.PanelDivider, tt.height)
			}
			if a.HandleRow < 0 || a.HandleRow >= tt.height {
				t.Errorf("HandleRow %d out of bounds [0, %d)", a.HandleRow, tt.height)
			}
			if a.ThresholdRow < 0 || a.ThresholdRow >= tt.height {
				t.Errorf("ThresholdRow %d out of bounds [0, %d)", a.ThresholdRow, tt.height)
			}

			// Monotonically increasing
			if a.LintelRow >= a.ContentStart {
				t.Errorf("LintelRow (%d) must be < ContentStart (%d)", a.LintelRow, a.ContentStart)
			}
			if a.ContentStart >= a.PanelDivider {
				t.Errorf("ContentStart (%d) must be < PanelDivider (%d)", a.ContentStart, a.PanelDivider)
			}
			if a.PanelDivider >= a.HandleRow {
				t.Errorf("PanelDivider (%d) must be < HandleRow (%d)", a.PanelDivider, a.HandleRow)
			}
			if a.HandleRow >= a.ThresholdRow {
				t.Errorf("HandleRow (%d) must be < ThresholdRow (%d)", a.HandleRow, a.ThresholdRow)
			}
		})
	}
}

func TestNewDoorAnatomy_FixedPositions(t *testing.T) {
	t.Parallel()

	// LintelRow is always 0
	a := NewDoorAnatomy(16)
	if a.LintelRow != 0 {
		t.Errorf("LintelRow should be 0, got %d", a.LintelRow)
	}

	// ThresholdRow is always height-1
	if a.ThresholdRow != 15 {
		t.Errorf("ThresholdRow should be 15 for height=16, got %d", a.ThresholdRow)
	}

	// ContentStart is 2
	if a.ContentStart != 2 {
		t.Errorf("ContentStart should be 2, got %d", a.ContentStart)
	}
}

func TestNewDoorAnatomy_ProportionalPositions(t *testing.T) {
	t.Parallel()

	a := NewDoorAnatomy(20)
	// PanelDivider at ~45% of 20 = 9
	if a.PanelDivider != 9 {
		t.Errorf("PanelDivider for height=20 should be 9, got %d", a.PanelDivider)
	}
	// HandleRow at ~60% of 20 = 12
	if a.HandleRow != 12 {
		t.Errorf("HandleRow for height=20 should be 12, got %d", a.HandleRow)
	}
}

func TestNewDoorAnatomy_BelowMinimum(t *testing.T) {
	t.Parallel()

	// Heights below 5 are clamped to 5
	a := NewDoorAnatomy(3)
	if a.ThresholdRow != 4 {
		t.Errorf("ThresholdRow should be 4 for clamped height=5, got %d", a.ThresholdRow)
	}
	// Still monotonically increasing
	if a.LintelRow >= a.ContentStart || a.ContentStart >= a.PanelDivider ||
		a.PanelDivider >= a.HandleRow || a.HandleRow >= a.ThresholdRow {
		t.Errorf("positions not monotonically increasing: %+v", a)
	}
}
