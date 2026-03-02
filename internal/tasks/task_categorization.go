package tasks

import "fmt"

// TaskType categorizes a task by its nature.
type TaskType string

const (
	TypeCreative       TaskType = "creative"
	TypeAdministrative TaskType = "administrative"
	TypeTechnical      TaskType = "technical"
	TypePhysical       TaskType = "physical"
)

// ValidateTaskType checks if a TaskType value is valid. Empty is valid (uncategorized).
func ValidateTaskType(t TaskType) error {
	switch t {
	case "", TypeCreative, TypeAdministrative, TypeTechnical, TypePhysical:
		return nil
	default:
		return fmt.Errorf("invalid task type: %q", t)
	}
}

// TaskEffort categorizes a task by its effort level.
type TaskEffort string

const (
	EffortQuickWin TaskEffort = "quick-win"
	EffortMedium   TaskEffort = "medium"
	EffortDeepWork TaskEffort = "deep-work"
)

// ValidateTaskEffort checks if a TaskEffort value is valid. Empty is valid (uncategorized).
func ValidateTaskEffort(e TaskEffort) error {
	switch e {
	case "", EffortQuickWin, EffortMedium, EffortDeepWork:
		return nil
	default:
		return fmt.Errorf("invalid task effort: %q", e)
	}
}

// TaskLocation categorizes a task by where it can be done.
type TaskLocation string

const (
	LocationHome     TaskLocation = "home"
	LocationWork     TaskLocation = "work"
	LocationErrands  TaskLocation = "errands"
	LocationAnywhere TaskLocation = "anywhere"
)

// ValidateTaskLocation checks if a TaskLocation value is valid. Empty is valid (uncategorized).
func ValidateTaskLocation(l TaskLocation) error {
	switch l {
	case "", LocationHome, LocationWork, LocationErrands, LocationAnywhere:
		return nil
	default:
		return fmt.Errorf("invalid task location: %q", l)
	}
}
