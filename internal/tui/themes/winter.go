package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewWinterTheme creates a crystalline winter theme with angular frames and
// dense dot patterns. Uses icy blue tones and sharp box-drawing characters
// to evoke frost and cold geometry.
func NewWinterTheme() *DoorTheme {
	frameColor := lipgloss.CompleteColor{TrueColor: "#87ceeb", ANSI256: "117", ANSI: "14"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#e0f0ff", ANSI256: "195", ANSI: "15"}

	return &DoorTheme{
		Name:        "winter",
		Description: "Winter crystalline — angular frames, dense dot patterns",
		Render:      winterRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#1a1a2e", ANSI256: "17", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#a0d2db", ANSI256: "152", ANSI: "14"},
			Selected: selectedColor,

			StatsAccent:        "#87CEEB",
			StatsGradientStart: "#1E3A5F",
			StatsGradientEnd:   "#87CEEB",
		},
		MinWidth:  15,
		MinHeight: 12,

		Season:      "winter",
		SeasonStart: MonthDay{12, 1},
		SeasonEnd:   MonthDay{2, 29},
	}
}

func winterRender(frameColor, selectedColor lipgloss.TerminalColor) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
		color := frameColor
		hChar := "─"
		vChar := "│"
		cornerTL, cornerTR := "┌", "┐"
		cornerBL, cornerBR := "└", "┘"
		if selected {
			color = selectedColor
			hChar = "━"
			vChar = "┃"
			cornerTL, cornerTR = "┏", "┓"
			cornerBL, cornerBR = "┗", "┛"
		}
		style := lipgloss.NewStyle().Foreground(color)

		inner := width - 2
		if inner < 1 {
			inner = 1
		}

		if height < 12 {
			return winterCompact(content, inner, hChar, vChar, cornerTL, cornerTR, cornerBL, cornerBR, style, hint)
		}

		return winterDoor(content, width, height, inner, hChar, vChar, cornerTL, cornerTR, cornerBL, cornerBR, style, selected, hint)
	}
}

func winterCompact(content string, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, hint string) string {
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Top border with angular corners
	fmt.Fprintf(&b, "%s\n", style.Render(tl+hBar+tr))

	// Upper padding
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Content lines
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

	// Blank line after content
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Doorknob line with dense dot pattern: ·· ◆
	knobPad := inner - 4
	if knobPad < 1 {
		knobPad = 1
	}
	knobLine := renderHandleWithHint(inner, knobPad, "◆", hint)
	fmt.Fprintf(&b, "%s%s%s\n",
		style.Render(vChar),
		knobLine,
		style.Render(vChar),
	)

	// Lower padding
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Bottom border
	fmt.Fprintf(&b, "%s", style.Render(bl+hBar+br))

	return b.String()
}

func winterDoor(content string, width, height, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, selected bool, hint string) string {
	anatomy := NewDoorAnatomy(height)

	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Dense dot fill character for frost texture
	dotChar := "·"

	// Frost texture row: dense dots between borders
	frostRow := func() string {
		pattern := strings.Repeat(dotChar+" ", inner/2)
		if len(pattern) > inner {
			pattern = pattern[:inner]
		}
		padLen := inner - ansi.StringWidth(pattern)
		if padLen < 0 {
			padLen = 0
		}
		return style.Render(vChar) + pattern + strings.Repeat(" ", padLen) + style.Render(vChar)
	}

	// Row just below lintel gets frost
	frostTopRow := anatomy.LintelRow + 1

	// Row just above threshold gets frost
	frostBotRow := anatomy.ThresholdRow - 1
	if frostBotRow <= anatomy.HandleRow {
		frostBotRow = -1 // skip if no room
	}

	// Panel divider uses tee junctions
	divLeft, divRight := "├", "┤"
	if selected {
		divLeft, divRight = "┣", "┫"
	}

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			fmt.Fprintf(&b, "%s", style.Render(tl+hBar+tr))

		case row == frostTopRow:
			fmt.Fprintf(&b, "%s", frostRow())

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s", style.Render(divLeft+hBar+divRight))

		case row == anatomy.HandleRow:
			knobPad := inner - 4
			if knobPad < 1 {
				knobPad = 1
			}
			knobLine := renderHandleWithHint(inner, knobPad, "◆", hint)
			fmt.Fprintf(&b, "%s%s%s", style.Render(vChar), knobLine, style.Render(vChar))

		case row == frostBotRow:
			fmt.Fprintf(&b, "%s", frostRow())

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s", style.Render(bl+hBar+br))

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
					style.Render(vChar),
					"   "+line+strings.Repeat(" ", padding),
					style.Render(vChar),
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

	// Threshold line below the door: dense dot pattern
	threshold := strings.Repeat("▔", width)
	fmt.Fprintf(&b, "\n%s", style.Render(threshold))

	return ApplyShadow(b.String(), width, 15, selected)
}
