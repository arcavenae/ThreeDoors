package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewModernTheme creates the Modern/Minimalist theme: clean single-line
// box-drawing frame, generous whitespace, and a single ● doorknob.
// When selected, uses heavy box-drawing characters (━┃) instead of thin (─│).
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
		MinWidth: 15,
	}
}

func modernRender(frameColor, selectedColor lipgloss.Color) func(string, int, bool) string {
	return func(content string, width int, selected bool) string {
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
}
