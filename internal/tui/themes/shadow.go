package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Shadow characters from Unicode block elements (U+2580-U+259F).
const (
	shadowRight         = "▐" // Right half block — right edge shadow
	shadowRightSelected = "█" // Full block — enhanced shadow for selected doors
	shadowBottom        = "▄" // Lower half block — bottom shadow row
)

// Shadow colors: muted grays that create depth without competing with theme colors.
var (
	shadowColor         lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#585858", ANSI256: "240", ANSI: "8"}
	shadowColorSelected lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#bcbcbc", ANSI256: "250", ANSI: "7"}
)

// ApplyShadow adds a right-edge shadow column and a bottom shadow row to a
// rendered door string. The shadow creates a 3D depth illusion by placing
// dark half-block characters on the right and bottom edges.
//
// Shadow is only applied when width >= minWidth+2, ensuring graceful
// degradation on narrow terminals.
func ApplyShadow(rendered string, width, minWidth int, selected bool) string {
	if width < minWidth+2 {
		return rendered
	}

	rightChar := shadowRight
	color := shadowColor
	if selected {
		rightChar = shadowRightSelected
		color = shadowColorSelected
	}
	style := lipgloss.NewStyle().Foreground(color)

	lines := strings.Split(rendered, "\n")
	var b strings.Builder

	for i, line := range lines {
		if i == 0 {
			// First row: no right shadow (shadow starts one row down for offset effect)
			fmt.Fprintf(&b, "%s %s", line, "")
		} else {
			fmt.Fprintf(&b, "%s%s", line, style.Render(rightChar))
		}
		if i < len(lines)-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Bottom shadow row: offset 1 char to the right, spans width of door
	fmt.Fprintf(&b, "\n %s", style.Render(strings.Repeat(shadowBottom, width)))

	return b.String()
}
