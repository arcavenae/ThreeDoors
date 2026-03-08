package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewModernTheme creates the Modern/Minimalist theme: clean single-line
// box-drawing frame, generous whitespace, and a minimalist handle.
// When height >= MinHeight (12), renders with door-like proportions: heavy bars,
// thin panel divider, and open circle handle.
// When height < MinHeight (or 0), falls back to the compact card style.
func NewModernTheme() *DoorTheme {
	frameColor := lipgloss.Color("238")
	selectedColor := lipgloss.Color("255")

	return &DoorTheme{
		Name:        "modern",
		Description: "Modern minimalist — clean lines, generous whitespace",
		Render:      modernRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("0"),
			Accent:   frameColor,
			Selected: selectedColor,
		},
		MinWidth:  15,
		MinHeight: 12,
	}
}

func modernRender(frameColor, selectedColor lipgloss.Color) func(string, int, int, bool) string {
	return func(content string, width int, height int, selected bool) string {
		color := frameColor
		hChar := "─"
		vChar := "│"
		if selected {
			color = selectedColor
			hChar = "━"
			vChar = "┃"
		}
		style := lipgloss.NewStyle().Foreground(color)

		// Interior width: total width minus 2 border characters
		inner := width - 2
		if inner < 1 {
			inner = 1
		}

		// Compact mode: use original card style
		if height < 12 {
			return modernCompact(content, inner, hChar, vChar, style)
		}

		// Door-like proportions using DoorAnatomy
		return modernDoor(content, width, height, inner, hChar, vChar, style)
	}
}

func modernCompact(content string, inner int, hChar, vChar string, style lipgloss.Style) string {
	// Word-wrap content with 3-char left padding, 3-char right padding
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)

	// Top border
	fmt.Fprintf(&b, "%s\n", style.Render(hChar+hBar+hChar))

	// Upper padding (1 blank line)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Content lines (left-padded with 3 spaces)
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

	// Doorknob line: ● placed near the right side
	knobPad := inner - 4
	if knobPad < 1 {
		knobPad = 1
	}
	knobLine := strings.Repeat(" ", knobPad) + "●" + strings.Repeat(" ", inner-knobPad-1)
	fmt.Fprintf(&b, "%s%s%s\n",
		style.Render(vChar),
		knobLine,
		style.Render(vChar),
	)

	// Lower padding (1 blank line)
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Bottom border
	fmt.Fprintf(&b, "%s", style.Render(hChar+hBar+hChar))

	return b.String()
}

func modernDoor(content string, width, height, inner int, hChar, vChar string, style lipgloss.Style) string {
	anatomy := NewDoorAnatomy(height)

	// Word-wrap content with 3-char padding on each side
	contentWidth := inner - 6
	if contentWidth < 1 {
		contentWidth = 1
	}
	wrapped := ansi.Wordwrap(content, contentWidth, "")
	contentLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(vChar) + strings.Repeat(" ", inner) + style.Render(vChar)

	// Panel divider always uses thin line for minimalist look
	thinH := "─"

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			// Top border: heavy bar across full width (no corners)
			fmt.Fprintf(&b, "%s", style.Render(hChar+hBar+hChar))

		case row == anatomy.PanelDivider:
			// Minimalist panel divider: thin line regardless of selection
			divBar := strings.Repeat(thinH, inner)
			fmt.Fprintf(&b, "%s", style.Render(vChar+divBar+vChar))

		case row == anatomy.HandleRow:
			// Minimalist handle: ○ (open circle) on the right side
			knobPad := inner - 4
			if knobPad < 1 {
				knobPad = 1
			}
			knobLine := strings.Repeat(" ", knobPad) + "○" + strings.Repeat(" ", inner-knobPad-1)
			fmt.Fprintf(&b, "%s%s%s", style.Render(vChar), knobLine, style.Render(vChar))

		case row == anatomy.ThresholdRow:
			// Bottom border: heavy bar across full width (no corners)
			fmt.Fprintf(&b, "%s", style.Render(hChar+hBar+hChar))

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			// Content area with 3-char left padding
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

	return b.String()
}
