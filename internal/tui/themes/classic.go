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
	frameColor := lipgloss.Color("63")
	selectedColor := lipgloss.Color("255")

	unselectedStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(frameColor).
		Padding(1, 2)

	selectedStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(selectedColor).
		Padding(1, 2)

	return &DoorTheme{
		Name:        "classic",
		Description: "Classic rounded border — the original ThreeDoors look",
		Render:      classicRender(frameColor, selectedColor, unselectedStyle, selectedStyle),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("0"),
			Accent:   frameColor,
			Selected: selectedColor,

			StatsAccent:        "#D97706", // warm amber
			StatsGradientStart: "#D97706", // amber
			StatsGradientEnd:   "#FCD34D", // gold
		},
		MinWidth:  15,
		MinHeight: 10,
	}
}

func classicRender(frameColor, selectedColor lipgloss.Color, unselectedStyle, selectedStyle lipgloss.Style) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
		// Compact mode: use original Lipgloss card style
		if height < 10 {
			if selected {
				return selectedStyle.Width(width).Render(content)
			}
			return unselectedStyle.Width(width).Render(content)
		}

		// Door-like proportions using DoorAnatomy
		anatomy := NewDoorAnatomy(height)
		color := frameColor
		if selected {
			color = selectedColor
		}
		style := lipgloss.NewStyle().Foreground(color)

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

		// Border characters
		topLeft, topRight := "╭", "╮"
		botLeft, botRight := "╰", "╯"
		vChar := "│"
		hChar := "─"
		if selected {
			topLeft, topRight = "┏", "┓"
			botLeft, botRight = "┗", "┛"
			vChar = "┃"
			hChar = "━"
		}

		hBar := strings.Repeat(hChar, inner)
		blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

		for row := 0; row < height; row++ {
			switch {
			case row == anatomy.LintelRow:
				// Top border (lintel)
				fmt.Fprintf(&b, "%s", style.Render(topLeft+hBar+topRight))

			case row == anatomy.PanelDivider:
				// Panel divider: ├─────────────┤ (or ┣━━━━━━━━━━━━━┫ selected)
				divLeft, divRight := "├", "┤"
				if selected {
					divLeft, divRight = "┣", "┫"
				}
				fmt.Fprintf(&b, "%s", style.Render(divLeft+hBar+divRight))

			case row == anatomy.HandleRow:
				// Doorknob row: ● on the right side
				knobPad := inner - 3
				if knobPad < 1 {
					knobPad = 1
				}
				knobLine := renderHandleWithHint(inner, knobPad, "●", hint)
				fmt.Fprintf(&b, "%s%s%s", style.Render(vChar), knobLine, style.Render(vChar))

			case row == anatomy.ThresholdRow:
				// Bottom border
				fmt.Fprintf(&b, "%s", style.Render(botLeft+hBar+botRight))

			case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
				// Content area
				lineIdx := row - anatomy.ContentStart
				if lineIdx < len(contentLines) {
					line := contentLines[lineIdx]
					lineWidth := ansi.StringWidth(line)
					padding := inner - 2 - lineWidth
					if padding < 0 {
						padding = 0
					}
					fmt.Fprintf(&b, "%s%s%s",
						style.Render(vChar),
						"  "+line+strings.Repeat(" ", padding),
						style.Render(vChar),
					)
				} else {
					fmt.Fprintf(&b, "%s", blankLine)
				}

			default:
				// Blank interior row
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
}
