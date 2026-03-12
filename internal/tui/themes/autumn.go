package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewAutumnTheme creates a layered autumn theme with block elements (▒▓)
// and angular textures. Uses warm amber/rust tones to evoke falling leaves.
func NewAutumnTheme() *DoorTheme {
	frameColor := lipgloss.CompleteColor{TrueColor: "#cc7722", ANSI256: "172", ANSI: "3"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#eebb55", ANSI256: "179", ANSI: "11"}

	return &DoorTheme{
		Name:        "autumn",
		Description: "Autumn harvest — layered block elements, angular textures",
		Render:      autumnRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#1a0f00", ANSI256: "52", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#dd9933", ANSI256: "178", ANSI: "3"},
			Selected: selectedColor,

			StatsAccent:        "#CC7722",
			StatsGradientStart: "#8B4513",
			StatsGradientEnd:   "#CC7722",
		},
		MinWidth:  15,
		MinHeight: 12,

		Season:      "autumn",
		SeasonStart: MonthDay{9, 1},
		SeasonEnd:   MonthDay{11, 30},
	}
}

func autumnRender(frameColor, selectedColor lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
		color := frameColor
		hChar := "─"
		vChar := "│"
		tl, tr := "┌", "┐"
		bl, br := "└", "┘"
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
			return autumnCompact(content, inner, hChar, vChar, tl, tr, bl, br, style, hint)
		}

		return autumnDoor(content, width, height, inner, hChar, vChar, tl, tr, bl, br, style, selected, hint, emphasis)
	}
}

func autumnCompact(content string, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, hint string) string {
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Angular top border
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

	// Handle: filled diamond for angular feel
	knobPad := inner - 4
	if knobPad < 1 {
		knobPad = 1
	}
	knobLine := renderHandleWithHint(inner, knobPad, "●", hint)
	fmt.Fprintf(&b, "%s%s%s\n",
		style.Render(vChar),
		knobLine,
		style.Render(vChar),
	)

	fmt.Fprintf(&b, "%s\n", blankLine)

	// Angular bottom border
	fmt.Fprintf(&b, "%s", style.Render(bl+hBar+br))

	return b.String()
}

func autumnDoor(content string, width, height, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, selected bool, hint string, emphasis float64) string {
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
	blankLine := style.Render(hingeV) + strings.Repeat(" ", inner) + style.Render(openV)

	shade := ""
	if cracked {
		shade = crackShade
		blankLine += crackShade
	}

	// Layered block texture rows using ▒ and ▓
	textureTopRow := anatomy.LintelRow + 1
	textureBotRow := anatomy.ThresholdRow - 1
	if textureBotRow <= anatomy.HandleRow {
		textureBotRow = -1
	}

	textureRow := func(block string) string {
		pattern := strings.Repeat(block, inner)
		return style.Render(hingeV) + pattern + style.Render(openV)
	}

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeTL+hBar+openTR), shade)

		case row == textureTopRow:
			fmt.Fprintf(&b, "%s%s", textureRow("▒"), shade)

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeTee+hBar+openTee), shade)

		case row == anatomy.HandleRow:
			knobPad := inner - 1
			if knobPad < 1 {
				knobPad = 1
			}
			handleChar := HandleCharForEmphasis(emphasis, selected, RoundKnobFrames)
			knobLine := renderHandleWithHint(inner, knobPad, handleChar, hint)
			fmt.Fprintf(&b, "%s%s%s%s", style.Render(hingeV), knobLine, style.Render(openV), shade)

		case row == textureBotRow:
			fmt.Fprintf(&b, "%s%s", textureRow("▓"), shade)

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s%s", style.Render(hingeBL+hBar+openBR), shade)

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				lineWidth := ansi.StringWidth(line)
				padding := inner - 3 - lineWidth
				if padding < 0 {
					padding = 0
				}
				fmt.Fprintf(&b, "%s%s%s%s",
					style.Render(hingeV),
					"   "+line+strings.Repeat(" ", padding),
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

	// Threshold: layered block pattern below the door
	threshold := strings.Repeat("▒", width)
	fmt.Fprintf(&b, "\n%s", style.Render(threshold))

	if cracked {
		return ApplyShadowWithCrack(b.String(), width, 15, selected)
	}
	return ApplyShadow(b.String(), width, 15, selected)
}
