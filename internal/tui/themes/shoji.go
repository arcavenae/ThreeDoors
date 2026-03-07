package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewShojiTheme creates the Japanese Shoji theme: thin wooden frame with large
// paper panes. The lattice feel comes from a few horizontal bars and a single
// mid-cross junction, not from many small cells.
// When selected, uses heavy grid characters (╋━┃) instead of light (┼─│).
func NewShojiTheme() *DoorTheme {
	frameColor := lipgloss.Color("180")
	selectedColor := lipgloss.Color("223")

	return &DoorTheme{
		Name:        "shoji",
		Description: "Japanese shoji — wooden lattice grid with paper panes",
		Render:      shojiRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("0"),
			Accent:   lipgloss.Color("137"),
			Selected: selectedColor,
		},
		MinWidth: 19,
	}
}

// shojiChars holds the box-drawing characters for a shoji frame.
type shojiChars struct {
	h     string // horizontal line segment
	v     string // vertical line
	cross string // interior cross junction
	tTop  string // top T-junction
	tBot  string // bottom T-junction
	tLeft string // left T-junction
	tRght string // right T-junction
}

func shojiRender(frameColor, selectedColor lipgloss.Color) func(string, int, bool) string {
	return func(content string, width int, selected bool) string {
		color := frameColor
		ch := shojiChars{
			h: "─", v: "│", cross: "┼",
			tTop: "┬", tBot: "┴", tLeft: "├", tRght: "┤",
		}
		if selected {
			color = selectedColor
			ch = shojiChars{
				h: "━", v: "┃", cross: "╋",
				tTop: "┳", tBot: "┻", tLeft: "┣", tRght: "┫",
			}
		}
		style := lipgloss.NewStyle().Foreground(color)

		// Interior width between the two frame verticals
		innerW := width - 2
		if innerW < 1 {
			innerW = 1
		}

		// Content text area: interior minus 2 padding spaces
		contentW := innerW - 2
		if contentW < 1 {
			contentW = 1
		}

		// Word-wrap content
		wrapped := ansi.Wordwrap(content, contentW, "")
		contentLines := strings.Split(wrapped, "\n")

		// Layout rows (excluding top/bottom rails which are separate):
		//   row 0: empty pane
		//   row 1: ├───┤ lattice bar
		//   row 2: empty pane
		//   rows 3..3+N-1: content lines
		//   row 3+N: empty pane
		//   row 3+N+1: ├─────┼─────┤ mid-cross bar
		//   row 3+N+2: empty pane
		//   row 3+N+3: empty pane
		//   row 3+N+4: ├───┤ lattice bar
		//   row 3+N+5: empty pane
		//
		// Total rendered width = width (frame v + innerW + frame v)

		totalW := innerW + 2

		// Position of the single cross junction on the mid-cross bar
		crossPos := innerW / 2

		hBar := shojiHBar(ch, innerW, style)
		crossBar := shojiCrossBar(ch, innerW, crossPos, style)
		emptyRow := shojiEmptyRow(ch, innerW, totalW, style)

		var b strings.Builder

		// Top rail
		fmt.Fprintf(&b, "%s\n", style.Render(ch.tTop+strings.Repeat(ch.h, innerW)+ch.tTop))
		// Row 0: empty
		fmt.Fprintf(&b, "%s\n", emptyRow)
		// Lattice bar 1
		fmt.Fprintf(&b, "%s\n", hBar)
		// Row 2: empty
		fmt.Fprintf(&b, "%s\n", emptyRow)
		// Content lines
		for _, line := range contentLines {
			fmt.Fprintf(&b, "%s\n", shojiContentRow(ch, innerW, totalW, line, style))
		}
		// Row after content: empty
		fmt.Fprintf(&b, "%s\n", emptyRow)
		// Mid-cross bar
		fmt.Fprintf(&b, "%s\n", crossBar)
		// Two empty rows
		fmt.Fprintf(&b, "%s\n", emptyRow)
		fmt.Fprintf(&b, "%s\n", emptyRow)
		// Lattice bar 2
		fmt.Fprintf(&b, "%s\n", hBar)
		// Bottom empty row
		fmt.Fprintf(&b, "%s\n", emptyRow)
		// Bottom rail
		fmt.Fprintf(&b, "%s", style.Render(ch.tBot+strings.Repeat(ch.h, innerW)+ch.tBot))

		return b.String()
	}
}

// shojiHBar builds a horizontal lattice bar: ├────────────────────────┤
func shojiHBar(ch shojiChars, innerW int, style lipgloss.Style) string {
	return style.Render(ch.tLeft + strings.Repeat(ch.h, innerW) + ch.tRght)
}

// shojiCrossBar builds a mid-cross bar: ├──────────┼─────────────┤
func shojiCrossBar(ch shojiChars, innerW, crossPos int, style lipgloss.Style) string {
	left := crossPos
	right := innerW - crossPos - 1
	if right < 0 {
		right = 0
	}
	return style.Render(ch.tLeft + strings.Repeat(ch.h, left) + ch.cross + strings.Repeat(ch.h, right) + ch.tRght)
}

// shojiEmptyRow renders an empty pane row: │                        │
func shojiEmptyRow(ch shojiChars, innerW, _ int, style lipgloss.Style) string {
	return style.Render(ch.v) + strings.Repeat(" ", innerW) + style.Render(ch.v)
}

// shojiContentRow renders a content row with padded text: │   Fix login bug        │
func shojiContentRow(ch shojiChars, innerW, _ int, text string, style lipgloss.Style) string {
	textWidth := ansi.StringWidth(text)
	contentW := innerW - 2
	if contentW < 1 {
		contentW = 1
	}
	if textWidth > contentW {
		text = ansi.Truncate(text, contentW, "")
		textWidth = ansi.StringWidth(text)
	}
	rightPad := contentW - textWidth
	if rightPad < 0 {
		rightPad = 0
	}
	return style.Render(ch.v) + " " + text + strings.Repeat(" ", rightPad) + " " + style.Render(ch.v)
}
