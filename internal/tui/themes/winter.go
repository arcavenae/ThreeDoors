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
		Render:      winterRender(frameColor, selectedColor, lipgloss.CompleteColor{TrueColor: "#0a0f1a", ANSI256: "233", ANSI: "0"}),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#0a0f1a", ANSI256: "233", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#a0d2db", ANSI256: "152", ANSI: "14"},
			Selected: selectedColor,

			FillLower:  lipgloss.CompleteColor{TrueColor: "#060a14", ANSI256: "232", ANSI: "0"},
			Highlight:  lipgloss.CompleteColor{TrueColor: "#a0d2e8", ANSI256: "152", ANSI: "14"},
			ShadowEdge: lipgloss.CompleteColor{TrueColor: "#4a6a80", ANSI256: "66", ANSI: "4"},
			ShadowNear: lipgloss.CompleteColor{TrueColor: "#354f60", ANSI256: "59", ANSI: "8"},
			ShadowFar:  lipgloss.CompleteColor{TrueColor: "#1a2a38", ANSI256: "236", ANSI: "0"},

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

func winterRender(frameColor, selectedColor, fill lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
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

		return winterDoor(content, width, height, inner, hChar, vChar, cornerTL, cornerTR, cornerBL, cornerBR, style, selected, hint, emphasis, fill)
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

func winterDoor(content string, width, height, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, selected bool, hint string, emphasis float64, fill lipgloss.TerminalColor) string {
	anatomy := NewDoorAnatomy(height)
	cracked := isCracked(selected, emphasis)

	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	// Hinge (left) uses heavier weight, opening (right) uses standard
	var hingeTL, openTR, hingeBL, openBR string
	var hingeV, openV string
	var hingeTee, openTee string

	if selected {
		hingeTL, openTR = "┏", "┐"
		hingeBL, openBR = "┗", "┘"
		hingeV, openV = "┃", "│"
		hingeTee, openTee = "┣", "┤"
	} else {
		hingeTL, openTR = "╓", "┐"
		hingeBL, openBR = "╙", "┘"
		hingeV, openV = "║", "│"
		hingeTee, openTee = "╟", "┤"
	}

	if cracked {
		openTR = crackTR
		openBR = crackBR
		openV = crackV
		inner--
	}

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(hingeV) + bgFillLine(inner, fill) + style.Render(openV)

	shade := ""
	if cracked {
		shade = crackShade
		blankLine += crackShade
	}

	// Dense dot fill character for frost texture
	dotChar := "·"

	// Frost texture row: dense dots between hinge borders
	frostRow := func() string {
		pattern := strings.Repeat(dotChar+" ", inner/2)
		if len(pattern) > inner {
			pattern = pattern[:inner]
		}
		return style.Render(hingeV) + bgFillContent(pattern, inner, 0, fill) + style.Render(openV)
	}

	// Row just below lintel gets frost
	frostTopRow := anatomy.LintelRow + 1

	// Row just above threshold gets frost
	frostBotRow := anatomy.ThresholdRow - 1
	if frostBotRow <= anatomy.HandleRow {
		frostBotRow = -1 // skip if no room
	}

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeTL+hBar+openTR), shade)

		case row == frostTopRow:
			fmt.Fprintf(&b, "%s%s", frostRow(), shade)

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeTee+hBar+openTee), shade)

		case row == anatomy.HandleRow:
			knobPad := inner - 1
			if knobPad < 1 {
				knobPad = 1
			}
			handleChar := HandleCharForEmphasis(emphasis, selected, DiamondHandleFrames)
			knobLine := renderHandleWithHint(inner, knobPad, handleChar, hint)
			fmt.Fprintf(&b, "%s%s%s%s", style.Render(hingeV), bgFillContent(knobLine, inner, 0, fill), style.Render(openV), shade)

		case row == frostBotRow:
			fmt.Fprintf(&b, "%s%s", frostRow(), shade)

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeBL+hBar+openBR), shade)

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				fmt.Fprintf(&b, "%s%s%s%s",
					style.Render(hingeV),
					bgFillContent(line, inner, 3, fill),
					style.Render(openV),
					shade,
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

	if cracked {
		return ApplyShadowWithCrack(b.String(), width, 15, selected)
	}
	return ApplyShadow(b.String(), width, 15, selected)
}
