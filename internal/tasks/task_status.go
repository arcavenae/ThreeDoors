package tasks

import "fmt"

// TaskStatus defines the lifecycle states a task can be in.
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusBlocked    TaskStatus = "blocked"
	StatusInProgress TaskStatus = "in-progress"
	StatusInReview   TaskStatus = "in-review"
	StatusComplete   TaskStatus = "complete"
)

// validTransitions maps each status to its allowed next states.
var validTransitions = map[TaskStatus][]TaskStatus{
	StatusTodo:       {StatusTodo, StatusInProgress, StatusBlocked, StatusComplete},
	StatusBlocked:    {StatusBlocked, StatusTodo, StatusInProgress, StatusComplete},
	StatusInProgress: {StatusInProgress, StatusBlocked, StatusInReview, StatusComplete},
	StatusInReview:   {StatusInReview, StatusInProgress, StatusComplete},
	StatusComplete:   {},
}

// IsValidTransition checks if transitioning from one status to another is allowed.
func IsValidTransition(from, to TaskStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// GetValidTransitions returns the list of valid next states for a given status.
func GetValidTransitions(current TaskStatus) []TaskStatus {
	transitions, ok := validTransitions[current]
	if !ok {
		return nil
	}
	return transitions
}

// ValidateStatus checks if a string is a valid TaskStatus.
func ValidateStatus(s string) error {
	switch TaskStatus(s) {
	case StatusTodo, StatusBlocked, StatusInProgress, StatusInReview, StatusComplete:
		return nil
	default:
		return fmt.Errorf("invalid task status: %q", s)
	}
}
