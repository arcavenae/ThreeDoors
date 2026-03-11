package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewSpringTheme creates a flowing spring theme with curved lines (╭╮╰╯)
// and light open patterns. Uses warm green tones to evoke new growth.
func NewSpringTheme() *DoorTheme {
	frameColor := lipgloss.CompleteColor{TrueColor: "#77dd77", ANSI256: "114", ANSI: "10"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#c1f0c1", ANSI256: "157", ANSI: "10"}

	return &DoorTheme{
		Name:        "spring",
		Description: "Spring bloom — flowing curves, light open patterns",
		Render:      springRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#0a1f0a", ANSI256: "22", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#98d898", ANSI256: "121", ANSI: "10"},
			Selected: selectedColor,

			StatsAccent:        "#77DD77",
			StatsGradientStart: "#2E7D32",
			StatsGradientEnd:   "#77DD77",
		},
		MinWidth:  15,
		MinHeight: 12,

		Season:      "spring",
		SeasonStart: MonthDay{3, 1},
		SeasonEnd:   MonthDay{5, 31},
	}
}

func springRender(frameColor, selectedColor lipgloss.TerminalColor) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
		color := frameColor
		hChar := "─"
		vChar := "│"
		tl, tr := "╭", "╮"
		bl, br := "╰", "╯"
		if selected {
			color = selectedColor
			hChar = "━"
			vChar = "┃"
			tl, tr = "┏", "┓"
			bl, br = "┗", "┛"
		}
		style := lipgloss.NewStyle().Foreground(color)

		inner := width - 2
		if inner < 1 {
			inner = 1
		}

		if height < 12 {
			return springCompact(content, inner, hChar, vChar, tl, tr, bl, br, style, hint)
		}

		return springDoor(content, width, height, inner, hChar, vChar, tl, tr, bl, br, style, selected, hint)
	}
}

func springCompact(content string, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, hint string) string {
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Curved top corners
	fmt.Fprintf(&b, "%s\n", style.Render(tl+hBar+tr))
	fmt.Fprintf(&b, "%s\n", blankLine)

	for _, line := range contentLines {
		lineWidth := ansi.StringWidth(line)
		padding := inner - 3 - lineWidth
		if padding < 0 {
			padding = 0
		}
		fmt.Fprintf(&b, "%s%s%s\n",
			style.Render(vChar),
			"   "+line+strings.Repeat(" ", padding),
			style.Render(vChar),
		)
	}

	fmt.Fprintf(&b, "%s\n", blankLine)

	// Handle: open circle for organic feel
	knobPad := inner - 4
	if knobPad < 1 {
		knobPad = 1
	}
	knobLine := renderHandleWithHint(inner, knobPad, "○", hint)
	fmt.Fprintf(&b, "%s%s%s\n",
		style.Render(vChar),
		knobLine,
		style.Render(vChar),
	)

	fmt.Fprintf(&b, "%s\n", blankLine)

	// Curved bottom corners
	fmt.Fprintf(&b, "%s", style.Render(bl+hBar+br))

	return b.String()
}

func springDoor(content string, width, height, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, selected bool, hint string) string {
	anatomy := NewDoorAnatomy(height)

	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	// Hinge (left) uses heavier weight, opening (right) uses standard with curved corners
	var hingeTL, openTR, hingeBL, openBR string
	var hingeV, openV string
	var hingeTee, openTee string

	if selected {
		hingeTL, openTR = "┏", "┐"
		hingeBL, openBR = "┗", "┘"
		hingeV, openV = "┃", "│"
		hingeTee, openTee = "┣", "┤"
	} else {
		hingeTL, openTR = "╓", "╮"
		hingeBL, openBR = "╙", "╯"
		hingeV, openV = "║", "│"
		hingeTee, openTee = "╟", "┤"
	}

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(hingeV) + strings.Repeat(" ", inner) + style.Render(openV)

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			// Hinge left, curved right corner
			fmt.Fprintf(&b, "%s", style.Render(hingeTL+hBar+openTR))

		case row == anatomy.PanelDivider:
			// Light divider with hinge junctions
			divH := "─"
			fmt.Fprintf(&b, "%s", style.Render(hingeTee+strings.Repeat(divH, inner)+openTee))

		case row == anatomy.HandleRow:
			// Open circle handle at rightmost content column
			knobPad := inner - 1
			if knobPad < 1 {
				knobPad = 1
			}
			knobLine := renderHandleWithHint(inner, knobPad, "○", hint)
			fmt.Fprintf(&b, "%s%s%s", style.Render(hingeV), knobLine, style.Render(openV))

		case row == anatomy.ThresholdRow:
			// Hinge left, curved right corner
			fmt.Fprintf(&b, "%s", style.Render(hingeBL+hBar+openBR))

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				lineWidth := ansi.StringWidth(line)
				padding := inner - 3 - lineWidth
				if padding < 0 {
					padding = 0
				}
				fmt.Fprintf(&b, "%s%s%s",
					style.Render(hingeV),
					"   "+line+strings.Repeat(" ", padding),
					style.Render(openV),
				)
			} else {
				fmt.Fprintf(&b, "%s", blankLine)
			}

		default:
			fmt.Fprintf(&b, "%s", blankLine)
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Threshold line below the door
	threshold := strings.Repeat("▔", width)
	fmt.Fprintf(&b, "\n%s", style.Render(threshold))

	return ApplyShadow(b.String(), width, 15, selected)
}
