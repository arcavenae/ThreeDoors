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
		Render:      shojiRender(frameColor, selectedColor, lipgloss.CompleteColor{TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"}),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#af8700", ANSI256: "137", ANSI: "3"},
			Selected: selectedColor,

			FillLower:  lipgloss.CompleteColor{TrueColor: "#141005", ANSI256: "233", ANSI: "0"},
			Highlight:  lipgloss.CompleteColor{TrueColor: "#e8c888", ANSI256: "186", ANSI: "11"},
			ShadowEdge: lipgloss.CompleteColor{TrueColor: "#8f7540", ANSI256: "137", ANSI: "3"},
			ShadowNear: lipgloss.CompleteColor{TrueColor: "#6a5530", ANSI256: "94", ANSI: "3"},
			ShadowFar:  lipgloss.CompleteColor{TrueColor: "#3a2a18", ANSI256: "236", ANSI: "0"},

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

func shojiRender(frameColor, selectedColor, fill lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
		// Compact mode: use existing fixed layout
		if height < 14 {
			return shojiCompactRender(content, width, selected, frameColor, selectedColor, hint)
		}

		return shojiDoorRender(content, width, height, selected, frameColor, selectedColor, hint, emphasis, fill)
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
// Hinge asymmetry: left uses heavier junctions (double-vert unselected, heavy selected),
// right uses standard-weight junctions.
func shojiDoorRender(content string, width, height int, selected bool, frameColor, selectedColor lipgloss.TerminalColor, hint string, emphasis float64, fill lipgloss.TerminalColor) string {
	anatomy := NewDoorAnatomy(height)
	cracked := isCracked(selected, emphasis)

	color := frameColor
	// Left (hinge) and right (opening) character sets
	var hingeTop, openTop string
	var hingeBot, openBot string
	var hingeV, openV string
	var hingeTee, openTee string
	var cross, h string

	if selected {
		color = selectedColor
		hingeTop, openTop = "┳", "┬"
		hingeBot, openBot = "┻", "┴"
		hingeV, openV = "┃", "│"
		hingeTee, openTee = "┣", "┤"
		cross = "╋"
		h = "━"
	} else {
		hingeTop, openTop = "╥", "┬"
		hingeBot, openBot = "╨", "┴"
		hingeV, openV = "║", "│"
		hingeTee, openTee = "╟", "┤"
		cross = "┼"
		h = "─"
	}

	if cracked {
		openTop = crackTR
		openBot = crackBR
		openV = crackV
	}
	style := lipgloss.NewStyle().Foreground(color)

	innerW := width - 2
	if cracked {
		innerW = width - 3 // reduce for shade column
	}
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

	shade := ""
	if cracked {
		shade = crackShade
	}

	// Helper for empty row with hinge asymmetry
	emptyRow := style.Render(hingeV) + bgFillLine(innerW, fill) + style.Render(openV) + shade

	// Helper for content row with hinge asymmetry
	contentRow := func(text string) string {
		cw := innerW - 2
		if cw < 1 {
			cw = 1
		}
		textWidth := ansi.StringWidth(text)
		if textWidth > cw {
			text = ansi.Truncate(text, cw, "")
		}
		return style.Render(hingeV) + bgFillContent(text, innerW, 1, fill) + style.Render(openV) + shade
	}

	// Helper for horizontal bar with hinge asymmetry
	hBar := func() string {
		return style.Render(hingeTee + strings.Repeat(h, innerW) + openTee)
	}

	// Helper for cross bar with hinge asymmetry
	crossBar := func() string {
		left := crossPos
		right := innerW - crossPos - 1
		if right < 0 {
			right = 0
		}
		return style.Render(hingeTee + strings.Repeat(h, left) + cross + strings.Repeat(h, right) + openTee)
	}

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
			fmt.Fprintf(&b, "%s%s", style.Render(hingeTop+strings.Repeat(h, innerW)+openTop), shade)

		case row == latticeBarRow:
			fmt.Fprintf(&b, "%s%s", hBar(), shade)

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s%s", crossBar(), shade)

		case row == anatomy.HandleRow:
			knobPad := innerW - 1
			if knobPad < 1 {
				knobPad = 1
			}
			handleChar := HandleCharForEmphasis(emphasis, selected, OpenKnobFrames)
			knobLine := renderHandleWithHint(innerW, knobPad, handleChar, hint)
			fmt.Fprintf(&b, "%s%s%s%s", style.Render(hingeV), bgFillContent(knobLine, innerW, 0, fill), style.Render(openV), shade)

		case row == latticeBar2Row:
			fmt.Fprintf(&b, "%s%s", hBar(), shade)

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeBot+strings.Repeat(h, innerW)+openBot), shade)

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				fmt.Fprintf(&b, "%s", contentRow(contentLines[lineIdx]))
			} else {
				fmt.Fprintf(&b, "%s", emptyRow)
			}

		default:
			fmt.Fprintf(&b, "%s", emptyRow)
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Threshold line below the door
	fmt.Fprintf(&b, "\n%s", style.Render(strings.Repeat("▔", width)))

	if cracked {
		return ApplyShadowWithCrack(b.String(), width, 19, selected)
	}
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
