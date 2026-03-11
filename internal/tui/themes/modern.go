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
	frameColor := lipgloss.CompleteColor{TrueColor: "#444444", ANSI256: "238", ANSI: "8"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#eeeeee", ANSI256: "255", ANSI: "15"}

	return &DoorTheme{
		Name:        "modern",
		Description: "Modern minimalist — clean lines, generous whitespace",
		Render:      modernRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"},
			Accent:   frameColor,
			Selected: selectedColor,

			StatsAccent:        "#0D9488", // cool teal
			StatsGradientStart: "#2563EB", // blue
			StatsGradientEnd:   "#0D9488", // teal
		},
		MinWidth:  15,
		MinHeight: 12,
	}
}

func modernRender(frameColor, selectedColor lipgloss.TerminalColor) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
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
			return modernCompact(content, inner, hChar, vChar, style, hint)
		}

		// Door-like proportions using DoorAnatomy
		return modernDoor(content, width, height, inner, hChar, vChar, style, selected, hint)
	}
}

func modernCompact(content string, inner int, hChar, vChar string, style lipgloss.Style, hint string) string {
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

	// Doorknob line: ● placed near the right side, with optional hint
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

	// Lower padding (1 blank line)
	fmt.Fprintf(&b, "%s\n", blankLine)

	// Bottom border
	fmt.Fprintf(&b, "%s", style.Render(hChar+hBar+hChar))

	return b.String()
}

func modernDoor(content string, width, height, inner int, hChar, vChar string, style lipgloss.Style, selected bool, hint string) string {
	anatomy := NewDoorAnatomy(height)

	// Word-wrap content with 3-char padding on each side
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
		hingeTL, openTR = "┏", "━"
		hingeBL, openBR = "┗", "━"
		hingeV, openV = "┃", "│"
		hingeTee, openTee = "┣", "┤"
	} else {
		hingeTL, openTR = "╓", "─"
		hingeBL, openBR = "╙", "─"
		hingeV, openV = "║", "│"
		hingeTee, openTee = "╟", "┤"
	}

	var b strings.Builder

	hBar := strings.Repeat(hChar, inner)
	blankLine := style.Render(hingeV) + strings.Repeat(" ", inner) + style.Render(openV)

	// Panel divider always uses thin line for minimalist look
	thinH := "─"

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			// Top border: hinge corner left, minimalist right
			fmt.Fprintf(&b, "%s", style.Render(hingeTL+hBar+openTR))

		case row == anatomy.PanelDivider:
			// Minimalist panel divider: thin line regardless of selection
			divBar := strings.Repeat(thinH, inner)
			fmt.Fprintf(&b, "%s", style.Render(hingeTee+divBar+openTee))

		case row == anatomy.HandleRow:
			// Minimalist handle: ○ at rightmost content column
			knobPad := inner - 1
			if knobPad < 1 {
				knobPad = 1
			}
			knobLine := renderHandleWithHint(inner, knobPad, "○", hint)
			fmt.Fprintf(&b, "%s%s%s", style.Render(hingeV), knobLine, style.Render(openV))

		case row == anatomy.ThresholdRow:
			// Bottom border: hinge corner left, minimalist right
			fmt.Fprintf(&b, "%s", style.Render(hingeBL+hBar+openBR))

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
					style.Render(hingeV),
					"   "+line+strings.Repeat(" ", padding),
					style.Render(openV),
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
