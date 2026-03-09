package core

import (
	"math/rand/v2"
	"time"
)

// DiversityScore scores a set of tasks for category diversity.
// +1 per unique Type value + +1 per unique Effort value + +1 per unique Location value.
// Uncategorized ("") counts as its own distinct value. Max score = 3 * len(tasks).
// Returns 0 for nil or empty input.
func DiversityScore(tasks []*Task) int {
	if len(tasks) == 0 {
		return 0
	}
	types := make(map[TaskType]bool)
	efforts := make(map[TaskEffort]bool)
	locations := make(map[TaskLocation]bool)

	for _, t := range tasks {
		types[t.Type] = true
		efforts[t.Effort] = true
		locations[t.Location] = true
	}

	return len(types) + len(efforts) + len(locations)
}

// SelectDoors picks up to count tasks from the pool, preferring diversity.
func SelectDoors(pool *TaskPool, count int) []*Task {
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	return selectDoorsWithRand(pool, count, rng, nil)
}

// SelectDoorsWithFocus picks up to count tasks, boosting focus-tagged tasks.
// planningTimestamp controls focus expiry; if nil, no focus boost is applied.
func SelectDoorsWithFocus(pool *TaskPool, count int, planningTimestamp *time.Time) []*Task {
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	return selectDoorsWithRand(pool, count, rng, planningTimestamp)
}

// selectDoorsWithRand picks up to count tasks from the pool using the provided RNG.
// It generates N=10 random candidate sets and picks the one with the highest diversity score.
// If planningTimestamp is non-nil, focus-tagged tasks receive a scoring boost.
func selectDoorsWithRand(pool *TaskPool, count int, rng *rand.Rand, planningTimestamp *time.Time) []*Task {
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

	const numCandidates = 10

	bestScore := -1
	var bestSet []*Task

	for i := range numCandidates {
		// Generate a random candidate set via Fisher-Yates partial shuffle
		perm := make([]*Task, len(available))
		copy(perm, available)
		for j := range count {
			k := j + rng.IntN(len(perm)-j)
			perm[j], perm[k] = perm[k], perm[j]
		}
		candidate := perm[:count]
		score := DiversityScore(candidate)
		if planningTimestamp != nil {
			score += int(FocusScoreBoost(candidate, *planningTimestamp))
		}

		if score > bestScore {
			bestScore = score
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		} else if score == bestScore && rng.IntN(i+1) == 0 {
			// Random tiebreak: replace with probability 1/(i+1)
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		}
	}

	for _, t := range bestSet {
		pool.MarkRecentlyShown(t.ID)
	}
	return bestSet
}
