package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewSummerTheme creates a bold summer theme with radiating lines and
// bold geometric shapes. Uses warm golden/orange tones to evoke bright sun.
func NewSummerTheme() *DoorTheme {
	frameColor := lipgloss.CompleteColor{TrueColor: "#ffaa00", ANSI256: "214", ANSI: "11"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#ffdd55", ANSI256: "221", ANSI: "11"}

	shadowNear := lipgloss.CompleteColor{TrueColor: "#a08828", ANSI256: "136", ANSI: "3"}
	shadowFar := lipgloss.CompleteColor{TrueColor: "#3a2a08", ANSI256: "236", ANSI: "0"}

	return &DoorTheme{
		Name:        "summer",
		Description: "Summer radiance — bold geometric shapes, radiating lines",
		Render:      summerRender(frameColor, selectedColor, lipgloss.CompleteColor{TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"}, lipgloss.CompleteColor{TrueColor: "#141005", ANSI256: "233", ANSI: "0"}, shadowNear, shadowFar),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#ffcc33", ANSI256: "220", ANSI: "11"},
			Selected: selectedColor,

			FillLower:  lipgloss.CompleteColor{TrueColor: "#141005", ANSI256: "233", ANSI: "0"},
			Highlight:  lipgloss.CompleteColor{TrueColor: "#ffd060", ANSI256: "221", ANSI: "11"},
			ShadowEdge: lipgloss.CompleteColor{TrueColor: "#8f7020", ANSI256: "136", ANSI: "3"},
			ShadowNear: shadowNear,
			ShadowFar:  shadowFar,

			StatsAccent:        "#FFAA00",
			StatsGradientStart: "#CC5500",
			StatsGradientEnd:   "#FFAA00",
		},
		MinWidth:  15,
		MinHeight: 12,

		Season:      "summer",
		SeasonStart: MonthDay{6, 1},
		SeasonEnd:   MonthDay{8, 31},
	}
}

func summerRender(frameColor, selectedColor, fill, fillLower, shadowNear, shadowFar lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	highlightColor := lipgloss.CompleteColor{TrueColor: "#ffd060", ANSI256: "221", ANSI: "11"}
	shadowEdgeColor := lipgloss.CompleteColor{TrueColor: "#8f7020", ANSI256: "136", ANSI: "3"}

	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
		color := frameColor
		hChar := "═"
		vChar := "║"
		tl, tr := "╔", "╗"
		bl, br := "╚", "╝"
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
			return summerCompact(content, inner, hChar, vChar, tl, tr, bl, br, style, hint)
		}

		// Bevel lighting for door mode
		var hlStyle, shStyle lipgloss.Style
		if selected {
			hlStyle = lipgloss.NewStyle().Foreground(selectedColor)
			shStyle = lipgloss.NewStyle().Foreground(selectedColor)
		} else {
			hlStyle = lipgloss.NewStyle().Foreground(highlightColor)
			shStyle = lipgloss.NewStyle().Foreground(shadowEdgeColor)
		}

		return summerDoor(content, width, height, inner, hChar, vChar, tl, tr, bl, br, hlStyle, shStyle, selected, hint, emphasis, fill, fillLower, shadowNear, shadowFar)
	}
}

func summerCompact(content string, inner int, hChar, vChar, tl, tr, bl, br string, style lipgloss.Style, hint string) string {
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Bold double-line top
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

	// Bold geometric handle: ■
	knobPad := inner - 4
	if knobPad < 1 {
		knobPad = 1
	}
	knobLine := renderHandleWithHint(inner, knobPad, "■", hint)
	fmt.Fprintf(&b, "%s%s%s\n",
		style.Render(vChar),
		knobLine,
		style.Render(vChar),
	)

	fmt.Fprintf(&b, "%s\n", blankLine)

	// Bold double-line bottom
	fmt.Fprintf(&b, "%s", style.Render(bl+hBar+br))

	return b.String()
}

func summerDoor(content string, width, height, inner int, hChar, vChar, tl, tr, bl, br string, hlStyle, shStyle lipgloss.Style, selected bool, hint string, emphasis float64, fill, fillLower, shadowNear, shadowFar lipgloss.TerminalColor) string {
	anatomy := NewDoorAnatomy(height)
	cracked := isCracked(selected, emphasis)

	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	// Summer already uses double-line (╔║╚) — hinge keeps double left,
	// opening (right) uses single-vertical with double-horizontal connections
	var hingeTL, openTR, hingeBL, openBR string
	var hingeV, openV string
	var hingeTee, openTee string
	var divH string

	if selected {
		hingeTL, openTR = "┏", "┐"
		hingeBL, openBR = "┗", "┘"
		hingeV, openV = "┃", "│"
		hingeTee, openTee = "┣", "┤"
		divH = "━"
	} else {
		hingeTL, openTR = "╔", "╕"
		hingeBL, openBR = "╚", "╛"
		hingeV, openV = "║", "│"
		hingeTee, openTee = "╠", "╡"
		divH = "═"
	}

	if cracked {
		openTR = crackTR
		openBR = crackBR
		openV = crackV
		inner--
	}

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)

	shade := ""
	if cracked {
		shade = crackShade
	}

	// Radiating accent row: geometric triangles
	accentChar := "▲"
	accentRow := anatomy.LintelRow + 1

	for row := 0; row < height; row++ {
		bg := panelBg(row, anatomy.PanelDivider, fill, fillLower)

		switch {
		case row == anatomy.LintelRow:
			fmt.Fprintf(&b, "%s%s", hlStyle.Render(hingeTL+hBar+openTR), shade)

		case row == accentRow:
			pattern := buildRadiatingPattern(accentChar, inner)
			fmt.Fprintf(&b, "%s%s%s%s", hlStyle.Render(hingeV), bgFillContent(pattern, inner, 0, bg), shStyle.Render(openV), shade)

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s%s", shStyle.Render(hingeTee+strings.Repeat(divH, inner)+openTee), shade)

		case row == anatomy.HandleRow:
			knobPad := inner - 1
			if knobPad < 1 {
				knobPad = 1
			}
			handleChar := HandleCharForEmphasis(emphasis, selected, SquareHandleFrames)
			knobLine := renderHandleWithHint(inner, knobPad, handleChar, hint)
			fmt.Fprintf(&b, "%s%s%s%s", hlStyle.Render(hingeV), bgFillContent(knobLine, inner, 0, bg), shStyle.Render(openV), shade)

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s%s", shStyle.Render(hingeBL+hBar+openBR), shade)

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				fmt.Fprintf(&b, "%s%s%s%s",
					hlStyle.Render(hingeV),
					bgFillContent(line, inner, 3, bg),
					shStyle.Render(openV),
					shade,
				)
			} else {
				fmt.Fprintf(&b, "%s%s%s%s", hlStyle.Render(hingeV), bgFillLine(inner, bg), shStyle.Render(openV), shade)
			}

		default:
			fmt.Fprintf(&b, "%s%s%s%s", hlStyle.Render(hingeV), bgFillLine(inner, bg), shStyle.Render(openV), shade)
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Bold threshold below the door
	threshold := strings.Repeat("▀", width)
	fmt.Fprintf(&b, "\n%s", shStyle.Render(threshold))

	if cracked {
		return ApplyShadowWithCrack(b.String(), width, 15, selected, shadowNear, shadowFar)
	}
	return ApplyShadow(b.String(), width, 15, selected, shadowNear, shadowFar)
}

// buildRadiatingPattern creates a spaced geometric pattern centered in the given width.
func buildRadiatingPattern(char string, innerWidth int) string {
	// Place geometric shapes every 4 characters for radiating effect
	var pattern strings.Builder
	for i := 0; i < innerWidth; i++ {
		if i%4 == 2 {
			fmt.Fprintf(&pattern, "%s", char)
		} else {
			fmt.Fprintf(&pattern, " ")
		}
	}
	result := pattern.String()
	padLen := innerWidth - ansi.StringWidth(result)
	if padLen < 0 {
		padLen = 0
	}
	return result + strings.Repeat(" ", padLen)
}
