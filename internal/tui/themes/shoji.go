package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewShojiTheme creates the Japanese Shoji theme: a lattice grid pattern
// using cross junctions (┼) with task text overlaid on central cells.
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
		MinWidth: 15,
	}
}

// shojiChars holds the box-drawing characters for a shoji grid.
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

		// Grid cell width is 3 runes: 2 horizontal chars + 1 junction
		cellW := 3
		numCols := (width - 1) / cellW
		if numCols < 3 {
			numCols = 3
		}
		// Actual rendered width in runes
		totalW := numCols*cellW + 1

		// Content area: skip 1 col on each side for margin
		contentCols := numCols - 2
		if contentCols < 1 {
			contentCols = 1
		}
		// Content char width = (contentCols * cellW) - 1 for the leading space
		// Each content cell is cellW-1 spaces + junction, but text overlays them
		contentCharW := contentCols*cellW - 1
		if contentCharW < 1 {
			contentCharW = 1
		}

		// Word-wrap content
		wrapped := ansi.Wordwrap(content, contentCharW, "")
		contentLines := strings.Split(wrapped, "\n")

		// Grid layout: we want enough rows for content
		numGridRows := 5
		contentStartRow := 2
		minNeeded := contentStartRow + len(contentLines) + 1
		if numGridRows < minNeeded {
			numGridRows = minNeeded
		}

		var b strings.Builder

		// Top rail: ┬──┬──┬──┬
		fmt.Fprintf(&b, "%s\n", style.Render(buildHLine(ch, numCols, ch.tTop, ch.tTop, ch.tTop)))

		contentIdx := 0
		for row := 0; row < numGridRows; row++ {
			// Cell row
			if row >= contentStartRow && contentIdx < len(contentLines) {
				fmt.Fprintf(&b, "%s\n", buildContentRow(style, ch, numCols, cellW, totalW, contentLines[contentIdx]))
				contentIdx++
			} else {
				fmt.Fprintf(&b, "%s\n", buildEmptyRow(style, ch, numCols, cellW, totalW))
			}

			// Grid separator line (except after last row)
			if row < numGridRows-1 {
				fmt.Fprintf(&b, "%s\n", style.Render(buildHLine(ch, numCols, ch.tLeft, ch.cross, ch.tRght)))
			}
		}

		// Bottom rail: ┴──┴──┴──┴
		fmt.Fprintf(&b, "%s", style.Render(buildHLine(ch, numCols, ch.tBot, ch.tBot, ch.tBot)))

		return b.String()
	}
}

// buildHLine builds a horizontal grid line like ├──┼──┼──┤.
func buildHLine(ch shojiChars, numCols int, left, mid, right string) string {
	var b strings.Builder
	b.WriteString(left)
	for col := 0; col < numCols; col++ {
		b.WriteString(strings.Repeat(ch.h, 2))
		if col < numCols-1 {
			b.WriteString(mid)
		} else {
			b.WriteString(right)
		}
	}
	return b.String()
}

// buildEmptyRow renders a row of empty cells: │  │  │  │
func buildEmptyRow(style lipgloss.Style, ch shojiChars, numCols, cellW, totalW int) string {
	var b strings.Builder
	b.WriteString(style.Render(ch.v))
	for col := 0; col < numCols; col++ {
		b.WriteString(strings.Repeat(" ", cellW-1))
		b.WriteString(style.Render(ch.v))
	}
	return b.String()
}

// buildContentRow renders a cell row with text overlaid on the central columns.
// The total rune width must match totalW exactly.
func buildContentRow(style lipgloss.Style, ch shojiChars, numCols, cellW, totalW int, text string) string {
	var b strings.Builder

	// First margin column: │  │
	b.WriteString(style.Render(ch.v))
	b.WriteString(strings.Repeat(" ", cellW-1))
	b.WriteString(style.Render(ch.v))

	// Central area: overlay text on the middle columns
	centralCols := numCols - 2
	if centralCols < 1 {
		centralCols = 1
	}
	// Central char width between the two margin │ chars
	// = centralCols cells worth of space minus 1 for the leading space
	// Actually: centralCols * cellW - 1 chars (cellW-1 spaces per cell + junction, minus trailing junction which is the margin │)
	centralCharW := centralCols*cellW - 1

	textWidth := ansi.StringWidth(text)
	if textWidth > centralCharW {
		text = ansi.Truncate(text, centralCharW, "")
		textWidth = ansi.StringWidth(text)
	}

	// Write: space + text + padding to fill centralCharW
	rightPad := centralCharW - 1 - textWidth
	if rightPad < 0 {
		rightPad = 0
	}
	b.WriteString(" ")
	b.WriteString(text)
	b.WriteString(strings.Repeat(" ", rightPad))

	// Last margin column: │  │
	b.WriteString(style.Render(ch.v))
	b.WriteString(strings.Repeat(" ", cellW-1))
	b.WriteString(style.Render(ch.v))

	return b.String()
}
