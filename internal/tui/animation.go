package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

const (
	// animationFPS is the target frame rate for spring animations.
	animationFPS = 60

	// animationFrameInterval is the time between animation frames.
	animationFrameInterval = time.Second / animationFPS

	// springFrequency controls how fast the spring oscillates (Hz).
	springFrequency = 6.0

	// springDamping controls how quickly the spring settles (0–1).
	// Higher values settle faster with less bounce.
	springDamping = 0.4

	// selectionEmphasisTarget is the target value for a fully selected door.
	selectionEmphasisTarget = 1.0

	// selectionEmphasisRest is the rest value for an unselected door.
	selectionEmphasisRest = 0.0

	// emphasisSettledThreshold is the distance from target below which
	// the animation is considered settled and stops ticking.
	emphasisSettledThreshold = 0.01
)

// animationFrameMsg is an internal tick message that drives animation frames.
type animationFrameMsg time.Time

// DoorAnimation tracks spring-physics animation state for door selection emphasis.
// Each door has an emphasis value that springs between 0 (unselected) and 1 (selected).
type DoorAnimation struct {
	springs  [3]harmonica.Spring
	emphasis [3]float64 // current emphasis values (0.0–1.0)
	velocity [3]float64 // current velocity per door
	targets  [3]float64 // target emphasis per door
	active   bool       // whether animation is currently running
}

// NewDoorAnimation creates a DoorAnimation with spring-physics parameters.
func NewDoorAnimation() *DoorAnimation {
	da := &DoorAnimation{}
	for i := range da.springs {
		da.springs[i] = harmonica.NewSpring(harmonica.FPS(animationFPS), springFrequency, springDamping)
	}
	return da
}

// SetSelection updates the animation targets based on which door is selected.
// Pass -1 for no selection (all doors return to rest).
func (da *DoorAnimation) SetSelection(index int) {
	for i := range da.targets {
		if i == index {
			da.targets[i] = selectionEmphasisTarget
		} else {
			da.targets[i] = selectionEmphasisRest
		}
	}
	da.active = true
}

// Update advances the animation by one frame. Returns true if the animation
// is still active (not yet settled).
func (da *DoorAnimation) Update() bool {
	if !da.active {
		return false
	}

	settled := true
	for i := range da.springs {
		da.emphasis[i], da.velocity[i] = da.springs[i].Update(da.emphasis[i], da.velocity[i], da.targets[i])

		dist := da.targets[i] - da.emphasis[i]
		if dist < 0 {
			dist = -dist
		}
		velAbs := da.velocity[i]
		if velAbs < 0 {
			velAbs = -velAbs
		}
		if dist > emphasisSettledThreshold || velAbs > emphasisSettledThreshold {
			settled = false
		}
	}

	if settled {
		// Snap to targets and stop
		for i := range da.emphasis {
			da.emphasis[i] = da.targets[i]
			da.velocity[i] = 0
		}
		da.active = false
		return false
	}
	return true
}

// Emphasis returns the current emphasis value for door i (0.0–1.0).
// Values may slightly overshoot 1.0 due to spring physics (bounce).
func (da *DoorAnimation) Emphasis(i int) float64 {
	if i < 0 || i >= 3 {
		return 0
	}
	return da.emphasis[i]
}

// Active returns whether the animation is still running.
func (da *DoorAnimation) Active() bool {
	return da.active
}

// Settled returns true when all emphasis values have reached their targets.
func (da *DoorAnimation) Settled() bool {
	return !da.active
}

// TickCmd returns a tea.Cmd that sends an animationFrameMsg after the frame interval.
func AnimationTickCmd() tea.Cmd {
	return tea.Tick(animationFrameInterval, func(t time.Time) tea.Msg {
		return animationFrameMsg(t)
	})
}
