package themes

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Shadow characters from Unicode block elements (U+2580-U+259F).
const (
	shadowRight         = "▐" // Right half block — single-column shadow
	shadowRightSelected = "█" // Full block — enhanced near-shadow for selected doors
	shadowBottom        = "▄" // Lower half block — bottom shadow row
)

// Shadow gradient characters for multi-column shadows.
const (
	shadowGradientNear = "▓" // Dense shade — nearest to door
	shadowGradientMid  = "▒" // Medium shade — middle gradient
	shadowGradientFar  = "░" // Light shade — farthest from door
)

// shadowColorFallback is used when a theme doesn't provide shadow colors.
var shadowColorFallback lipgloss.TerminalColor = lipgloss.CompleteColor{TrueColor: "#707070", ANSI256: "242", ANSI: "8"}

// ShadowColumns returns the number of right-edge shadow columns to render
// based on available width beyond the minimum. Selected doors get 1 extra
// column for enhanced depth.
//
// Width thresholds:
//   - width < minWidth+2: 0 columns (no shadow)
//   - width >= minWidth+2: 1 column (unselected), 2 columns (selected)
//   - width >= minWidth+4: 2 columns (unselected), 3 columns (selected)
//   - width >= minWidth+6: 3 columns (both — maximum gradient)
func ShadowColumns(width, minWidth int, selected bool) int {
	extra := width - minWidth
	var cols int
	switch {
	case extra < 2:
		cols = 0
	case extra < 4:
		cols = 1
	case extra < 6:
		cols = 2
	default:
		cols = 3
	}
	if selected && cols > 0 && cols < 3 {
		cols++
	}
	return cols
}

// ApplyShadowWithCrack adds a gradient shadow with the first column rendered
// as a crack-of-light shade character (░), simulating light leaking through
// a cracked-open door.
func ApplyShadowWithCrack(rendered string, width, minWidth int, selected bool, shadowNear, shadowFar lipgloss.TerminalColor) string {
	cols := ShadowColumns(width, minWidth, selected)
	if cols == 0 {
		return rendered
	}

	near, far := resolveShadowColors(shadowNear, shadowFar)
	chars := gradientChars(cols, selected)
	colors := gradientColors(cols, near, far)
	chars[0] = crackShade

	return applyShadowGradient(rendered, width, cols, chars, colors)
}

// ApplyShadow adds a width-adaptive gradient shadow to a rendered door string.
// Shadow columns use gradient shade characters (▓▒░) with per-theme colors
// that fade from shadowNear (door-adjacent) to shadowFar (background-adjacent).
func ApplyShadow(rendered string, width, minWidth int, selected bool, shadowNear, shadowFar lipgloss.TerminalColor) string {
	cols := ShadowColumns(width, minWidth, selected)
	if cols == 0 {
		return rendered
	}

	near, far := resolveShadowColors(shadowNear, shadowFar)
	chars := gradientChars(cols, selected)
	colors := gradientColors(cols, near, far)

	return applyShadowGradient(rendered, width, cols, chars, colors)
}

func resolveShadowColors(near, far lipgloss.TerminalColor) (lipgloss.TerminalColor, lipgloss.TerminalColor) {
	if near == nil {
		near = shadowColorFallback
	}
	if far == nil {
		far = shadowColorFallback
	}
	return near, far
}

// gradientChars returns the shade characters for each shadow column.
func gradientChars(cols int, selected bool) []string {
	switch cols {
	case 1:
		if selected {
			return []string{shadowRightSelected}
		}
		return []string{shadowRight}
	case 2:
		if selected {
			return []string{shadowRightSelected, shadowGradientFar}
		}
		return []string{shadowGradientNear, shadowGradientFar}
	case 3:
		if selected {
			return []string{shadowRightSelected, shadowGradientMid, shadowGradientFar}
		}
		return []string{shadowGradientNear, shadowGradientMid, shadowGradientFar}
	default:
		return nil
	}
}

// gradientColors returns the color for each shadow column.
func gradientColors(cols int, near, far lipgloss.TerminalColor) []lipgloss.TerminalColor {
	switch cols {
	case 1:
		return []lipgloss.TerminalColor{near}
	case 2:
		return []lipgloss.TerminalColor{near, far}
	case 3:
		return []lipgloss.TerminalColor{near, near, far}
	default:
		return nil
	}
}

func applyShadowGradient(rendered string, width, cols int, chars []string, colors []lipgloss.TerminalColor) string {
	lines := strings.Split(rendered, "\n")
	var b strings.Builder

	styles := make([]lipgloss.Style, cols)
	for i := range cols {
		styles[i] = lipgloss.NewStyle().Foreground(colors[i])
	}

	for i, line := range lines {
		if i == 0 {
			// First row: no right shadow (offset effect) — pad to match width
			fmt.Fprintf(&b, "%s%s", line, strings.Repeat(" ", cols))
		} else {
			fmt.Fprintf(&b, "%s", line)
			for c := range cols {
				fmt.Fprintf(&b, "%s", styles[c].Render(chars[c]))
			}
		}
		if i < len(lines)-1 {
			fmt.Fprintf(&b, "\n")
		}
	}

	// Bottom shadow row: offset 1 char right, spans width + shadow columns
	fmt.Fprintf(&b, "\n %s", styles[0].Render(strings.Repeat(shadowBottom, width+cols-1)))

	return b.String()
}
