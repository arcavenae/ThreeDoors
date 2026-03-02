package tasks

import "math/rand"

// SelectRandomDoors picks up to count random tasks from allTasks.
// If exclude is provided, tries to avoid those tasks (best effort).
// Returns a new slice; does not modify input.
func SelectRandomDoors(allTasks []Task, count int, exclude []Task) []Task {
	if len(allTasks) == 0 || count <= 0 {
		return nil
	}

	excludeSet := make(map[string]bool, len(exclude))
	for _, t := range exclude {
		excludeSet[t.Text] = true
	}

	// Try to pick from non-excluded tasks first
	var candidates []Task
	for _, t := range allTasks {
		if !excludeSet[t.Text] {
			candidates = append(candidates, t)
		}
	}

	// If not enough non-excluded candidates, fall back to all tasks
	if len(candidates) < count {
		candidates = make([]Task, len(allTasks))
		copy(candidates, allTasks)
	}

	// Shuffle and pick
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if count > len(candidates) {
		count = len(candidates)
	}

	result := make([]Task, count)
	copy(result, candidates[:count])
	return result
}
