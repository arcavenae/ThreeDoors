package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// NewSciFiTheme creates the Sci-Fi/Spaceship theme: double-line outer frame,
// shade-filled side rails, single content panel with inline ACCESS label.
// When selected, uses bright shade (▓) instead of light (░).
// When height >= MinHeight, renders door-like proportions with bulkhead divider,
// access panel handle, and floor grating.
func NewSciFiTheme() *DoorTheme {
	frameColor := lipgloss.Color("39")
	selectedColor := lipgloss.Color("51")

	return &DoorTheme{
		Name:        "scifi",
		Description: "Sci-fi spaceship — double-line frame, shade panels, ACCESS label",
		Render:      scifiRender(frameColor, selectedColor),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.Color("236"),
			Accent:   lipgloss.Color("45"),
			Selected: selectedColor,

			StatsAccent:        "#22C55E", // neon green
			StatsGradientStart: "#22C55E", // green
			StatsGradientEnd:   "#06B6D4", // cyan
		},
		MinWidth:  16,
		MinHeight: 14,
	}
}

func scifiRender(frameColor, selectedColor lipgloss.Color) func(string, int, int, bool, string) string {
	return func(content string, width int, height int, selected bool, hint string) string {
		color := frameColor
		shadeChar := "░"
		if selected {
			color = selectedColor
			shadeChar = "▓"
		}
		style := lipgloss.NewStyle().Foreground(color)

		// Layout: ║░│ content │░║
		// Rail width: 1 shade char on each side
		// Total border overhead: 2 (║) + 2 (░ x2) + 2 (│) = 6
		railW := 1
		innerBorder := 6
		contentW := width - innerBorder
		if contentW < 1 {
			contentW = 1
		}

		// Word-wrap content with 2-char padding on each side
		textW := contentW - 4
		if textW < 1 {
			textW = 1
		}
		wrapped := ansi.Wordwrap(content, textW, "")
		contentLines := strings.Split(wrapped, "\n")

		rail := strings.Repeat(shadeChar, railW)

		// Compact mode: use existing fixed-layout rendering
		if height < 14 {
			return scifiRenderCompact(style, content, contentLines, width, contentW, rail, shadeChar, selected, hint)
		}

		// Door-like proportions using DoorAnatomy
		return scifiRenderDoor(style, contentLines, width, contentW, rail, shadeChar, railW, height, selected, hint)
	}
}

// scifiRenderCompact renders the original fixed-height Sci-Fi card style.
func scifiRenderCompact(style lipgloss.Style, _ string, contentLines []string, width, contentW int, rail, _ string, _ bool, hint string) string {
	railW := 1
	blankContent := strings.Repeat(" ", contentW)

	var b strings.Builder

	// Top border: ╔═╤══════════════════════╤═╗
	fmt.Fprintf(&b, "%s\n", style.Render(
		"╔"+strings.Repeat("═", railW)+"╤"+strings.Repeat("═", contentW)+"╤"+strings.Repeat("═", railW)+"╗"))

	// Blank line
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

	// Content lines with 2-char padding
	for _, line := range contentLines {
		lineWidth := ansi.StringWidth(line)
		pad := contentW - 2 - lineWidth
		if pad < 0 {
			pad = 0
		}
		fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
			style.Render("║"), rail, style.Render("│"),
			"  "+line+strings.Repeat(" ", pad),
			style.Render("│"), rail, style.Render("║"))
	}

	// Blank lines after content
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

	// ACCESS label right-aligned with 2-char padding, with optional hint
	label := "[ACCESS]"
	if hint != "" {
		label = hint + " " + label
	}
	leftPad := contentW - ansi.StringWidth(label) - 2
	if leftPad < 0 {
		leftPad = 0
	}
	labelRight := contentW - leftPad - ansi.StringWidth(label)
	if labelRight < 0 {
		labelRight = 0
	}
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"),
		strings.Repeat(" ", leftPad)+label+strings.Repeat(" ", labelRight),
		style.Render("│"), rail, style.Render("║"))

	// Blank line after ACCESS
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))

	// Bottom border: ╚═╧══════════════════════╧═╝
	fmt.Fprintf(&b, "%s", style.Render(
		"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+"╝"))

	return b.String()
}

// scifiRenderDoor renders the Sci-Fi theme with door-like proportions using DoorAnatomy.
func scifiRenderDoor(style lipgloss.Style, contentLines []string, width, contentW int, rail, shadeChar string, railW, height int, selected bool, hint string) string {
	anatomy := NewDoorAnatomy(height)
	blankContent := strings.Repeat(" ", contentW)

	// Find the row for the ACCESS label: midpoint between HandleRow and ThresholdRow
	accessRow := anatomy.HandleRow + (anatomy.ThresholdRow-anatomy.HandleRow)/2
	if accessRow <= anatomy.HandleRow {
		accessRow = anatomy.HandleRow + 1
	}
	if accessRow >= anatomy.ThresholdRow {
		accessRow = anatomy.ThresholdRow - 1
	}

	var b strings.Builder

	for row := 0; row < height; row++ {
		switch {
		case row == anatomy.LintelRow:
			// Top border: ╔═╤══════════════════════╤═╗
			fmt.Fprintf(&b, "%s", style.Render(
				"╔"+strings.Repeat("═", railW)+"╤"+strings.Repeat("═", contentW)+"╤"+strings.Repeat("═", railW)+"╗"))

		case row == anatomy.PanelDivider:
			// Bulkhead divider: ║░╞═════════════════════╡░║
			fmt.Fprintf(&b, "%s%s%s%s%s%s",
				style.Render("║"), rail,
				style.Render("╞"+strings.Repeat("═", contentW)+"╡"),
				rail, style.Render("║"), "")

		case row == anatomy.HandleRow:
			// Access panel handle: ◈──┤ on the right side
			handleStr := "◈──┤"
			handleWidth := ansi.StringWidth(handleStr)
			leftPad := contentW - handleWidth - 1
			if leftPad < 0 {
				leftPad = 0
			}
			rightPad := contentW - leftPad - handleWidth
			if rightPad < 0 {
				rightPad = 0
			}
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				style.Render("║"), rail, style.Render("│"),
				strings.Repeat(" ", leftPad)+handleStr+strings.Repeat(" ", rightPad),
				style.Render("│"), rail, style.Render("║"))

		case row == accessRow:
			// ACCESS label right-aligned with 2-char padding, with optional hint
			label := "[ACCESS]"
			if hint != "" {
				label = hint + " " + label
			}
			labelWidth := ansi.StringWidth(label)
			leftPad := contentW - labelWidth - 2
			if leftPad < 0 {
				leftPad = 0
			}
			rightPad := contentW - leftPad - labelWidth
			if rightPad < 0 {
				rightPad = 0
			}
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				style.Render("║"), rail, style.Render("│"),
				strings.Repeat(" ", leftPad)+label+strings.Repeat(" ", rightPad),
				style.Render("│"), rail, style.Render("║"))

		case row == anatomy.ThresholdRow:
			// Bottom border: ╚═╧══════════════════════╧═╝
			fmt.Fprintf(&b, "%s", style.Render(
				"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+"╝"))

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			// Content area with 2-char padding
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				lineWidth := ansi.StringWidth(line)
				pad := contentW - 2 - lineWidth
				if pad < 0 {
					pad = 0
				}
				fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
					style.Render("║"), rail, style.Render("│"),
					"  "+line+strings.Repeat(" ", pad),
					style.Render("│"), rail, style.Render("║"))
			} else {
				fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
					style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))
			}

		default:
			// Blank interior row
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("║"))
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Floor grating line below the door
	gratingChar := "▓"
	grating := strings.Repeat(gratingChar, width)
	fmt.Fprintf(&b, "\n%s", style.Render(grating))

	return ApplyShadow(b.String(), width, 16, selected)
}
