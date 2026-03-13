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
	frameColor := lipgloss.CompleteColor{TrueColor: "#00afff", ANSI256: "39", ANSI: "4"}
	selectedColor := lipgloss.CompleteColor{TrueColor: "#00ffff", ANSI256: "51", ANSI: "14"}

	shadowNear := lipgloss.CompleteColor{TrueColor: "#2080a0", ANSI256: "31", ANSI: "6"}
	shadowFar := lipgloss.CompleteColor{TrueColor: "#001a2f", ANSI256: "17", ANSI: "0"}

	return &DoorTheme{
		Name:        "scifi",
		Description: "Sci-fi spaceship — double-line frame, shade panels, ACCESS label",
		Render:      scifiRender(frameColor, selectedColor, lipgloss.CompleteColor{TrueColor: "#0a1a2e", ANSI256: "17", ANSI: "0"}, lipgloss.CompleteColor{TrueColor: "#061425", ANSI256: "17", ANSI: "0"}, shadowNear, shadowFar),
		Colors: ThemeColors{
			Frame:    frameColor,
			Fill:     lipgloss.CompleteColor{TrueColor: "#0a1a2e", ANSI256: "17", ANSI: "0"},
			Accent:   lipgloss.CompleteColor{TrueColor: "#00d7ff", ANSI256: "45", ANSI: "14"},
			Selected: selectedColor,

			FillLower:  lipgloss.CompleteColor{TrueColor: "#061425", ANSI256: "17", ANSI: "0"},
			Highlight:  lipgloss.CompleteColor{TrueColor: "#00d7ff", ANSI256: "45", ANSI: "14"},
			ShadowEdge: lipgloss.CompleteColor{TrueColor: "#005f7f", ANSI256: "24", ANSI: "4"},
			ShadowNear: shadowNear,
			ShadowFar:  shadowFar,

			StatsAccent:        "#22C55E", // neon green
			StatsGradientStart: "#22C55E", // green
			StatsGradientEnd:   "#06B6D4", // cyan
		},
		MinWidth:  16,
		MinHeight: 14,
	}
}

func scifiRender(frameColor, selectedColor, fill, fillLower, shadowNear, shadowFar lipgloss.TerminalColor) func(string, int, int, bool, string, float64) string {
	highlightColor := lipgloss.CompleteColor{TrueColor: "#00d7ff", ANSI256: "45", ANSI: "14"}
	shadowEdgeColor := lipgloss.CompleteColor{TrueColor: "#005f7f", ANSI256: "24", ANSI: "4"}

	return func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
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

		// Bevel lighting for door mode
		var hlStyle, shStyle lipgloss.Style
		if selected {
			hlStyle = lipgloss.NewStyle().Foreground(selectedColor)
			shStyle = lipgloss.NewStyle().Foreground(selectedColor)
		} else {
			hlStyle = lipgloss.NewStyle().Foreground(highlightColor)
			shStyle = lipgloss.NewStyle().Foreground(shadowEdgeColor)
		}

		// Door-like proportions using DoorAnatomy
		return scifiRenderDoor(hlStyle, shStyle, contentLines, width, contentW, rail, shadeChar, railW, height, selected, hint, emphasis, fill, fillLower, shadowNear, shadowFar)
	}
}

// scifiRenderCompact renders the original fixed-height Sci-Fi card style.
// Hinge asymmetry: outer left stays double (╔║╚), outer right uses lighter (╕│╛).
func scifiRenderCompact(style lipgloss.Style, _ string, contentLines []string, width, contentW int, rail, _ string, _ bool, hint string) string {
	railW := 1
	blankContent := strings.Repeat(" ", contentW)

	var b strings.Builder

	// Top border: ╔═╤══════════════════════╤═╕ (hinge left, lighter right)
	fmt.Fprintf(&b, "%s\n", style.Render(
		"╔"+strings.Repeat("═", railW)+"╤"+strings.Repeat("═", contentW)+"╤"+strings.Repeat("═", railW)+"╕"))

	// Blank line
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("│"))

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
			style.Render("│"), rail, style.Render("│"))
	}

	// Blank lines after content
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("│"))
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("│"))

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
		style.Render("│"), rail, style.Render("│"))

	// Blank line after ACCESS
	fmt.Fprintf(&b, "%s%s%s%s%s%s%s\n",
		style.Render("║"), rail, style.Render("│"), blankContent, style.Render("│"), rail, style.Render("│"))

	// Bottom border: ╚═╧══════════════════════╧═╛ (hinge left, lighter right)
	fmt.Fprintf(&b, "%s", style.Render(
		"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+"╛"))

	return b.String()
}

// scifiRenderDoor renders the Sci-Fi theme with door-like proportions using DoorAnatomy.
// Hinge asymmetry: outer left border stays double-line (╔║╚), outer right uses
// single-vertical with double-horizontal connections (╕│╛) for lighter weight.
func scifiRenderDoor(hlStyle, shStyle lipgloss.Style, contentLines []string, width, contentW int, rail, shadeChar string, railW, height int, selected bool, hint string, emphasis float64, fill, fillLower, shadowNear, shadowFar lipgloss.TerminalColor) string {
	anatomy := NewDoorAnatomy(height)
	cracked := isCracked(selected, emphasis)

	// Find the row for the ACCESS label: midpoint between HandleRow and ThresholdRow
	accessRow := anatomy.HandleRow + (anatomy.ThresholdRow-anatomy.HandleRow)/2
	if accessRow <= anatomy.HandleRow {
		accessRow = anatomy.HandleRow + 1
	}
	if accessRow >= anatomy.ThresholdRow {
		accessRow = anatomy.ThresholdRow - 1
	}

	var b strings.Builder

	// Scifi crack: outer right border │ → ╎, corners ╕ → ╮, ╛ → ╯
	outerR := "│"
	outerTR := "╕"
	outerBR := "╛"
	if cracked {
		outerR = crackV
		outerTR = crackTR
		outerBR = crackBR
	}

	// Rail strings styled for bevel: left rail = highlight, right rail = shadow
	hlRail := hlStyle.Render(rail)
	shRail := shStyle.Render(rail)

	for row := 0; row < height; row++ {
		bg := panelBg(row, anatomy.PanelDivider, fill, fillLower)

		switch {
		case row == anatomy.LintelRow:
			fmt.Fprintf(&b, "%s", hlStyle.Render(
				"╔"+strings.Repeat("═", railW)+"╤"+strings.Repeat("═", contentW)+"╤"+strings.Repeat("═", railW)+outerTR))

		case row == anatomy.PanelDivider:
			fmt.Fprintf(&b, "%s%s%s%s%s",
				hlStyle.Render("║"), hlRail,
				shStyle.Render("╞"+strings.Repeat("═", contentW)+"╡"),
				shRail, shStyle.Render(outerR))

		case row == anatomy.HandleRow:
			handleChar := HandleCharForEmphasis(emphasis, selected, SciFiHandleFrames)
			handleStr := handleChar + "──┤"
			handleWidth := ansi.StringWidth(handleStr)
			leftPad := contentW - handleWidth
			if leftPad < 0 {
				leftPad = 0
			}
			handleLine := strings.Repeat(" ", leftPad) + handleStr
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				hlStyle.Render("║"), hlRail, hlStyle.Render("│"),
				bgFillContent(handleLine, contentW, 0, bg),
				shStyle.Render("│"), shRail, shStyle.Render(outerR))

		case row == accessRow:
			label := "[ACCESS]"
			if hint != "" {
				label = hint + " " + label
			}
			labelWidth := ansi.StringWidth(label)
			leftPad := contentW - labelWidth - 2
			if leftPad < 0 {
				leftPad = 0
			}
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				hlStyle.Render("║"), hlRail, hlStyle.Render("│"),
				bgFillContent(label, contentW, leftPad, bg),
				shStyle.Render("│"), shRail, shStyle.Render(outerR))

		case row == anatomy.ThresholdRow:
			fmt.Fprintf(&b, "%s", shStyle.Render(
				"╚"+strings.Repeat("═", railW)+"╧"+strings.Repeat("═", contentW)+"╧"+strings.Repeat("═", railW)+outerBR))

		case row >= anatomy.ContentStart && row < anatomy.PanelDivider:
			lineIdx := row - anatomy.ContentStart
			if lineIdx < len(contentLines) {
				line := contentLines[lineIdx]
				fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
					hlStyle.Render("║"), hlRail, hlStyle.Render("│"),
					bgFillContent(line, contentW, 2, bg),
					shStyle.Render("│"), shRail, shStyle.Render(outerR))
			} else {
				fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
					hlStyle.Render("║"), hlRail, hlStyle.Render("│"), bgFillLine(contentW, bg), shStyle.Render("│"), shRail, shStyle.Render(outerR))
			}

		default:
			fmt.Fprintf(&b, "%s%s%s%s%s%s%s",
				hlStyle.Render("║"), hlRail, hlStyle.Render("│"), bgFillLine(contentW, bg), shStyle.Render("│"), shRail, shStyle.Render(outerR))
		}

		if row < height-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Floor grating line below the door
	gratingChar := "▓"
	grating := strings.Repeat(gratingChar, width)
	fmt.Fprintf(&b, "\n%s", shStyle.Render(grating))

	if cracked {
		return ApplyShadowWithCrack(b.String(), width, 16, selected, shadowNear, shadowFar)
	}
	return ApplyShadow(b.String(), width, 16, selected, shadowNear, shadowFar)
}
