package core

import (
	"fmt"
	"sort"
	"time"
)

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

// EnergyMatchScore returns a gradient score (0.0–1.0) for how well a task's
// effort matches the given energy level.
//
//   - 1.0: perfect match (e.g., high energy + deep-work)
//   - 0.5: adjacent match (e.g., high energy + medium effort)
//   - 0.25: mismatch (e.g., high energy + quick-win)
//   - 0.75: no effort tag (neutral — matches anything, slightly below perfect)
func EnergyMatchScore(task *Task, energy EnergyLevel) float64 {
	if task.Effort == "" {
		return 0.75
	}

	// Map effort to ordinal: deep-work=2, medium=1, quick-win=0
	effortOrd := effortOrdinal(task.Effort)
	// Map energy to ordinal: high=2, medium=1, low=0
	energyOrd := energyOrdinal(energy)

	diff := effortOrd - energyOrd
	if diff < 0 {
		diff = -diff
	}

	switch diff {
	case 0:
		return 1.0
	case 1:
		return 0.5
	default:
		return 0.25
	}
}

func effortOrdinal(e TaskEffort) int {
	switch e {
	case EffortDeepWork:
		return 2
	case EffortMedium:
		return 1
	case EffortQuickWin:
		return 0
	default:
		return 1
	}
}

func energyOrdinal(e EnergyLevel) int {
	switch e {
	case EnergyHigh:
		return 2
	case EnergyMedium:
		return 1
	case EnergyLow:
		return 0
	default:
		return 1
	}
}

// SortByEnergyMatch sorts tasks in-place by energy match score (best first).
// Uses stable sort to preserve original order for equal scores.
func SortByEnergyMatch(tasks []*Task, energy EnergyLevel) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return EnergyMatchScore(tasks[i], energy) > EnergyMatchScore(tasks[j], energy)
	})
}

// FilterByEnergy returns tasks whose effort matches the given energy level.
func FilterByEnergy(tasks []*Task, energy EnergyLevel) []*Task {
	if len(tasks) == 0 {
		return nil
	}
	var result []*Task
	for _, t := range tasks {
		if MatchesEnergy(t, energy) {
			result = append(result, t)
		}
	}
	return result
}

// EnergySetScore returns the sum of energy match scores for a set of tasks.
func EnergySetScore(tasks []*Task, energy EnergyLevel) float64 {
	var total float64
	for _, t := range tasks {
		total += EnergyMatchScore(t, energy)
	}
	return total
}

// NextEnergyLevel cycles through energy levels: High → Medium → Low → High.
func NextEnergyLevel(current EnergyLevel) EnergyLevel {
	switch current {
	case EnergyHigh:
		return EnergyMedium
	case EnergyMedium:
		return EnergyLow
	case EnergyLow:
		return EnergyHigh
	default:
		return EnergyHigh
	}
}

// EnergyLabel returns a human-readable label for the energy level.
func EnergyLabel(energy EnergyLevel) string {
	switch energy {
	case EnergyHigh:
		return "High"
	case EnergyMedium:
		return "Medium"
	case EnergyLow:
		return "Low"
	default:
		return "Unknown"
	}
}

// TimeOfDayLabel returns a time-of-day name based on the hour.
func TimeOfDayLabel(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour <= 11:
		return "morning"
	case hour >= 12 && hour <= 16:
		return "afternoon"
	default:
		return "evening"
	}
}

// EnergyDisplayString returns a formatted energy display string.
// Auto-inferred: "Energy: High (morning)"
// Overridden: "Energy: Medium (override)"
func EnergyDisplayString(energy EnergyLevel, t time.Time, overridden bool) string {
	if overridden {
		return fmt.Sprintf("Energy: %s (override)", EnergyLabel(energy))
	}
	return fmt.Sprintf("Energy: %s (%s)", EnergyLabel(energy), TimeOfDayLabel(t))
}
