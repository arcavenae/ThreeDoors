package themes

// DoorAnatomy holds the structural row positions for door-like rendering.
// All positions are zero-indexed row numbers within a door of the given height.
type DoorAnatomy struct {
	LintelRow    int // Top border / header area (always 0)
	ContentStart int // First row available for content text
	PanelDivider int // Horizontal division between upper/lower panels (~45% height)
	HandleRow    int // Doorknob/handle placement (~60% height)
	ThresholdRow int // Bottom border / floor line (height-1)
	HingeCol     int // Hinge placement column (always 0 — leftmost edge)
}

// NewDoorAnatomy calculates structural row positions for a door of the given height.
// All positions are within bounds [0, height-1] and monotonically increasing.
func NewDoorAnatomy(height int) DoorAnatomy {
	if height < 5 {
		height = 5
	}

	lintel := 0
	contentStart := 2
	threshold := height - 1

	panelDivider := height * 45 / 100
	handleRow := height * 60 / 100

	// Ensure monotonically increasing: lintel < contentStart < panelDivider < handleRow < threshold
	if panelDivider <= contentStart {
		panelDivider = contentStart + 1
	}
	if handleRow <= panelDivider {
		handleRow = panelDivider + 1
	}
	if handleRow >= threshold {
		handleRow = threshold - 1
	}
	if panelDivider >= handleRow {
		panelDivider = handleRow - 1
	}
	if contentStart >= panelDivider {
		contentStart = panelDivider - 1
	}

	return DoorAnatomy{
		LintelRow:    lintel,
		ContentStart: contentStart,
		PanelDivider: panelDivider,
		HandleRow:    handleRow,
		ThresholdRow: threshold,
		HingeCol:     0,
	}
}
