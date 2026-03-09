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
	frameColor := lipgloss.CompleteColor{TrueColor: "#d7af87", ANSI256: "180", ANSI: "3"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#ffd7af", ANSI256: "223", ANSI: "11"}

	return &DoorTheme{
		Name:        "shoji",
		Description: "Japanese shoji — wooden lattice grid with paper panes",
		Render:      shojiRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#af8700", ANSI256: "137", ANSI: "3"},
			Selected: selectedColor,

			StatsAccent:        "#92400E", // earth brown
			StatsGradientStart: "#92400E", // clay
			StatsGradientEnd:   "#D4A574", // sand
		},
		MinWidth:  19,
		MinHeight: 14,
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

func shojiRender(frameColor, selectedColor lipgloss.TerminalColor) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
		// Compact mode: use existing fixed layout
		if height < 14 {
			return shojiCompactRender(content, width, selected, frameColor, selectedColor, hint)
		}

		return shojiDoorRender(content, width, height, selected, frameColor, selectedColor, hint)
	}
}

// shojiCompactRender preserves the original fixed-layout rendering for compact mode.
func shojiCompactRender(content string, width int, selected bool, frameColor, selectedColor lipgloss.TerminalColor, _ string) string {
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

	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}

	contentW := innerW - 2
	if contentW < 1 {
		contentW = 1
	}

	wrapped := ansi.Wordwrap(content, contentW, "")
	contentLines := strings.Split(wrapped, "\n")

	crossPos := innerW / 2

	hBar := shojiHBar(ch, innerW, style)
	crossBar := shojiCrossBar(ch, innerW, crossPos, style)
	emptyRow := shojiEmptyRow(ch, innerW, style)

	var b strings.Builder

	fmt.Fprintf(&b, "%s\n", style.Render(ch.tTop+strings.Repeat(ch.h, innerW)+ch.tTop))
	fmt.Fprintf(&b, "%s\n", emptyRow)
	fmt.Fprintf(&b, "%s\n", hBar)
	fmt.Fprintf(&b, "%s\n", emptyRow)
	for _, line := range contentLines {
		fmt.Fprintf(&b, "%s\n", shojiContentRow(ch, innerW, line, style))
	}
	fmt.Fprintf(&b, "%s\n", emptyRow)
	fmt.Fprintf(&b, "%s\n", crossBar)
	fmt.Fprintf(&b, "%s\n", emptyRow)
	fmt.Fprintf(&b, "%s\n", emptyRow)
	fmt.Fprintf(&b, "%s\n", hBar)
	fmt.Fprintf(&b, "%s\n", emptyRow)
	fmt.Fprintf(&b, "%s", style.Render(ch.tBot+strings.Repeat(ch.h, innerW)+ch.tBot))

	return b.String()
}

// shojiDoorRender renders the Shoji theme with door-like proportions using DoorAnatomy.
func shojiDoorRender(content string, width, height int, selected bool, frameColor, selectedColor lipgloss.TerminalColor, hint string) string {
	anatomy := NewDoorAnatomy(height)

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

	innerW := width - 2
	if innerW < 1 {
		innerW = 1
	}

	contentW := innerW - 2
	if contentW < 1 {
		contentW = 1
	}

	wrapped := ansi.Wordwrap(content, contentW, "")
	contentLines := strings.Split(wrapped, "\n")

	crossPos := innerW / 2
	emptyRow := shojiEmptyRow(ch, innerW, style)

	// Place a lattice bar one row after ContentStart (gives a pane above content)
	latticeBarRow := anatomy.ContentStart - 1
	if latticeBarRow <= anatomy.LintelRow {
		latticeBarRow = anatomy.LintelRow + 1
	}

	// Place a second lattice bar between HandleRow and ThresholdRow
	latticeBar2Row := anatomy.HandleRow + 1
	if latticeBar2Row >= anatomy.ThresholdRow {
		latticeBar2Row = anatomy.ThresholdRow - 1
	}

	var b strings.Builder

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			// Top rail
			fmt.Fprintf(&b, "%s", style.Render(ch.tTop+strings.Repeat(ch.h, innerW)+ch.tTop))

		case row == latticeBarRow:
			// Lattice bar above content
			fmt.Fprintf(&b, "%s", shojiHBar(ch, innerW, style))

		case row == anatomy.PanelDivider:
			// Mid-cross bar at panel divider
			fmt.Fprintf(&b, "%s", shojiCrossBar(ch, innerW, crossPos, style))

		case row == anatomy.HandleRow:
			// Handle row: ○ on the right side, with optional hint
			knobPad := innerW - 3
			if knobPad < 1 {
				knobPad = 1
			}
			knobLine := renderHandleWithHint(innerW, knobPad, "○", hint)
			fmt.Fprintf(&b, "%s%s%s", style.Render(ch.v), knobLine, style.Render(ch.v))

		case row == latticeBar2Row:
			// Lattice bar below handle
			fmt.Fprintf(&b, "%s", shojiHBar(ch, innerW, style))

		case row == anatomy.ThresholdRow:
			// Bottom rail
			fmt.Fprintf(&b, "%s", style.Render(ch.tBot+strings.Repeat(ch.h, innerW)+ch.tBot))

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			// Content area
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				fmt.Fprintf(&b, "%s", shojiContentRow(ch, innerW, contentLines[lineIdx], style))
			} else {
				fmt.Fprintf(&b, "%s", emptyRow)
			}

		default:
			// Empty pane row
			fmt.Fprintf(&b, "%s", emptyRow)
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Threshold line below the door
	fmt.Fprintf(&b, "\n%s", style.Render(strings.Repeat("▔", width)))

	return ApplyShadow(b.String(), width, 19, selected)
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
func shojiEmptyRow(ch shojiChars, innerW int, style lipgloss.Style) string {
	return style.Render(ch.v) + strings.Repeat(" ", innerW) + style.Render(ch.v)
}

// shojiContentRow renders a content row with padded text: │   Fix login bug        │
func shojiContentRow(ch shojiChars, innerW int, text string, style lipgloss.Style) string {
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
