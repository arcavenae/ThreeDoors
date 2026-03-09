package core

import "time"

// EnergyLevel represents the user's current energy level.
type EnergyLevel string

const (
	EnergyHigh   EnergyLevel = "high"
	EnergyMedium EnergyLevel = "medium"
	EnergyLow    EnergyLevel = "low"
)

// InferEnergyFromTime maps time of day to an energy level.
//   - 06:00–11:59 → High
//   - 12:00–16:59 → Medium
//   - 17:00–05:59 → Low
func InferEnergyFromTime(t time.Time) EnergyLevel {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour <= 11:
		return EnergyHigh
	case hour >= 12 && hour <= 16:
		return EnergyMedium
	default:
		return EnergyLow
	}
}

// MatchesEnergy returns true if a task's effort level aligns with the given energy level.
//   - High energy → prefer deep-work tasks
//   - Medium energy → prefer medium-effort tasks
//   - Low energy → prefer quick-win tasks
//
// Tasks with no effort tag match any energy level.
func MatchesEnergy(task *Task, energy EnergyLevel) bool {
	if task.Effort == "" {
		return true
	}
	switch energy {
	case EnergyHigh:
		return task.Effort == EffortDeepWork
	case EnergyMedium:
		return task.Effort == EffortMedium
	case EnergyLow:
		return task.Effort == EffortQuickWin
	default:
		return true
	}
}
