package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewClassicTheme creates the Classic theme with door-like proportions.
// When height >= MinHeight, renders with panel divider, doorknob, and threshold.
// When height < MinHeight (or 0), falls back to the original compact card style.
func NewClassicTheme() *DoorTheme {
	frameColor := lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "5"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#eeeeee", ANSI256: "255", ANSI: "15"}

	unselectedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(frameColor).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(selectedColor).
		Padding(1, 2)

	fillColor := lipgloss.CompleteColor{TrueColor: "#0d0d2a", ANSI256: "17", ANSI: "0"}
	fillLowerColor := lipgloss.CompleteColor{TrueColor: "#080820", ANSI256: "17", ANSI: "0"}
	shadowNear := lipgloss.CompleteColor{TrueColor: "#6868b8", ANSI256: "104", ANSI: "13"}
	shadowFar := lipgloss.CompleteColor{TrueColor: "#15152a", ANSI256: "235", ANSI: "0"}

	return &DoorTheme{
		Name:        "classic",
		Description: "Classic rounded border — the original ThreeDoors look",
		Render:      classicRender(frameColor, selectedColor, unselectedStyle, selectedStyle, fillColor, fillLowerColor, shadowNear, shadowFar),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     fillColor,
			Accent:   frameColor,
			Selected: selectedColor,

			FillLower:  fillLowerColor,
			Highlight:  lipgloss.CompleteColor{TrueColor: "#7070ff", ANSI256: "63", ANSI: "5"},
			ShadowEdge: lipgloss.CompleteColor{TrueColor: "#3a3a8f", ANSI256: "61", ANSI: "4"},
			ShadowNear: shadowNear,
			ShadowFar:  shadowFar,

			StatsAccent:        "#D97706", // warm amber
			StatsGradientStart: "#D97706", // amber
			StatsGradientEnd:   "#FCD34D", // gold
		},
		MinWidth:  15,
		MinHeight: 10,
	}
}

func classicRender(frameColor, selectedColor lipgloss.TerminalColor, unselectedStyle, selectedStyle lipgloss.Style, fill, fillLower, shadowNear, shadowFar lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	highlightColor := lipgloss.CompleteColor{TrueColor: "#7070ff", ANSI256: "63", ANSI: "5"}
	shadowEdgeColor := lipgloss.CompleteColor{TrueColor: "#3a3a8f", ANSI256: "61", ANSI: "4"}

	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
		// Compact mode: use original Lipgloss card style
		if height < 10 {
			if selected {
				return selectedStyle.Width(width).Render(content)
			}
			return unselectedStyle.Width(width).Render(content)
		}

		// Door-like proportions using DoorAnatomy
		anatomy := NewDoorAnatomy(height)

		// Bevel lighting: top/left edges use Highlight, bottom/right use ShadowEdge.
		// When selected, both use Selected color to maintain selection visual.
		var hlStyle, shStyle lipgloss.Style
		if selected {
			hlStyle = lipgloss.NewStyle().Foreground(selectedColor)
			shStyle = lipgloss.NewStyle().Foreground(selectedColor)
		} else {
			hlStyle = lipgloss.NewStyle().Foreground(highlightColor)
			shStyle = lipgloss.NewStyle().Foreground(shadowEdgeColor)
		}
		// Interior width: total width minus 2 border characters
		inner := width - 2
		if inner < 1 {
			inner = 1
		}

		// Word-wrap content with 2-char padding on each side
		contentWidth := inner - 4
		if contentWidth < 1 {
			contentWidth = 1
		}
		wrapped := ansi.Wordwrap(content, contentWidth, "")
		contentLines := strings.Split(wrapped, "\n")

		var b strings.Builder

		// Border characters: left (hinge) uses heavier weight, right (opening) uses standard
		var hingeTL, openTR, hingeBL, openBR string
		var hingeV, openV string
		var hingeTee, openTee string
		hChar := "─"

		cracked := isCracked(selected, emphasis)

		if selected {
			hingeTL, openTR = "┏", "┐"
			hingeBL, openBR = "┗", "┘"
			hingeV, openV = "┃", "│"
			hingeTee, openTee = "┣", "┤"
			hChar = "━"
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
			inner-- // reduce content area by 1 for shade column
		}

		hBar := strings.Repeat(hChar, inner)

		shade := ""
		if cracked {
			shade = crackShade
		}

		for row := 0; row < height; row++ {
			bg := panelBg(row, anatomy.PanelDivider, fill, fillLower)

			switch {
			case row == anatomy.LintelRow:
				fmt.Fprintf(&b, "%s%s", hlStyle.Render(hingeTL+hBar+openTR), shade)

			case row == anatomy.PanelDivider:
				fmt.Fprintf(&b, "%s%s", shStyle.Render(hingeTee+hBar+openTee), shade)

			case row == anatomy.HandleRow:
				knobPad := inner - 1
				if knobPad < 1 {
					knobPad = 1
				}
				handleChar := HandleCharForEmphasis(emphasis, selected, RoundKnobFrames)
				knobLine := renderHandleWithHint(inner, knobPad, handleChar, hint)
				bgKnob := bgFillContent(knobLine, inner, 0, bg)
				fmt.Fprintf(&b, "%s%s%s%s", hlStyle.Render(hingeTee), bgKnob, shStyle.Render(openV), shade)

			case row == anatomy.ThresholdRow:
				fmt.Fprintf(&b, "%s%s", shStyle.Render(hingeBL+hBar+openBR), shade)

			case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
				lineIdx := row - anatomy.ContentStart
				if lineIdx < len(contentLines) {
					line := contentLines[lineIdx]
					fmt.Fprintf(&b, "%s%s%s%s",
						hlStyle.Render(hingeV),
						bgFillContent(line, inner, 2, bg),
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

		// Threshold line below the door
		threshold := strings.Repeat("▔", width)
		fmt.Fprintf(&b, "\n%s", shStyle.Render(threshold))

		if cracked {
			return ApplyShadowWithCrack(b.String(), width, 15, selected, shadowNear, shadowFar)
		}
		return ApplyShadow(b.String(), width, 15, selected, shadowNear, shadowFar)
	}
}
