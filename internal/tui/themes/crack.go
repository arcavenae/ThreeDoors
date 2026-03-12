package themes

// CrackEmphasisThreshold is the minimum spring emphasis value at which
// the crack-of-light effect appears on the right border of a selected door.
// Below this threshold, the door renders with standard border characters
// even when selected, enabling smooth synchronization with the spring animation.
const CrackEmphasisThreshold = 0.3

// Crack-of-light characters: used to replace right border when cracked open.
const (
	crackV  = "╎" // Replaces right vertical border │
	crackTR = "╮" // Replaces top-right corner (rounded, suggesting door swung away)
	crackBR = "╯" // Replaces bottom-right corner (rounded)
)

// crackShade is the shade character that appears to the right of the crack,
// simulating light leaking through the gap.
const crackShade = "░"

// isCracked returns true when the crack-of-light effect should be rendered.
// Requires both selection (selected=true) and sufficient spring emphasis.
func isCracked(selected bool, emphasis float64) bool {
	return selected && emphasis >= CrackEmphasisThreshold
}
