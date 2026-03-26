package todoist

import "github.com/arcavenae/ThreeDoors/internal/core"

// MapPriorityToEffort converts a Todoist priority (0-4) to a ThreeDoors TaskEffort.
// Todoist uses an inverted scale: 1 is normal, 4 is urgent.
// Priority 0 means "no priority set".
//
// Mapping:
//
//	0 (none)     -> quick-win
//	1 (normal)   -> quick-win
//	2 (high)     -> medium
//	3 (urgent)   -> deep-work
//	4 (critical) -> deep-work
func MapPriorityToEffort(priority int) core.TaskEffort {
	switch priority {
	case 0, 1:
		return core.EffortQuickWin
	case 2:
		return core.EffortMedium
	case 3, 4:
		return core.EffortDeepWork
	default:
		return core.EffortQuickWin
	}
}

// MapStatus converts Todoist is_completed to a ThreeDoors TaskStatus.
func MapStatus(isCompleted bool) core.TaskStatus {
	if isCompleted {
		return core.StatusComplete
	}
	return core.StatusTodo
}
