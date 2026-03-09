package tui

import "github.com/charmbracelet/bubbles/viewport"

// NewScrollableView creates a pre-configured viewport with ThreeDoors defaults.
// All scrollable views should use this factory to ensure consistent scrolling
// behavior (mouse wheel enabled, same key bindings, same speed).
func NewScrollableView(width, height int) viewport.Model {
	vp := viewport.New(width, height)
	vp.MouseWheelEnabled = true
	return vp
}
