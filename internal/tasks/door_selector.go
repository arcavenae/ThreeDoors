package tasks

import "math/rand/v2"

// SelectDoors picks up to count random tasks from the pool for door display.
func SelectDoors(pool *TaskPool, count int) []*Task {
	available := pool.GetAvailableForDoors()
	if len(available) == 0 {
		return nil
	}
	if len(available) <= count {
		for _, t := range available {
			pool.MarkRecentlyShown(t.ID)
		}
		return available
	}

	// Fisher-Yates shuffle and take first `count`
	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	selected := available[:count]
	for _, t := range selected {
		pool.MarkRecentlyShown(t.ID)
	}
	return selected
}
