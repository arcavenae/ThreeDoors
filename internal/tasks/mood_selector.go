package tasks

import (
	"math/rand/v2"
	"strings"
	"time"
)

// MoodAlignmentScore scores a set of tasks for mood alignment.
// +2 for each task matching preferredType, +1 for each matching preferredEffort.
// These stack: a task matching both gets +3.
// Returns 0 for nil or empty input.
func MoodAlignmentScore(tasks []*Task, preferredType TaskType, preferredEffort TaskEffort) int {
	score := 0
	for _, t := range tasks {
		if t.Type == preferredType && preferredType != "" {
			score += 2
		}
		if t.Effort == preferredEffort && preferredEffort != "" {
			score++
		}
	}
	return score
}

// SelectDoorsWithMood picks up to count tasks from the pool, considering mood preferences.
// Falls back to diversity-only selection when mood data is unavailable.
func SelectDoorsWithMood(pool *TaskPool, count int, currentMood string, patterns *PatternReport) []*Task {
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	return selectDoorsWithMoodAndRand(pool, count, currentMood, patterns, rng)
}

// selectDoorsWithMoodAndRand picks tasks with mood awareness using a deterministic RNG.
func selectDoorsWithMoodAndRand(pool *TaskPool, count int, currentMood string, patterns *PatternReport, rng *rand.Rand) []*Task {
	// Fallback check: no mood, no patterns, or no matching correlation
	if currentMood == "" || patterns == nil {
		return selectDoorsWithRand(pool, count, rng)
	}

	// Find matching mood correlation (case-insensitive — moods are lowercased in analysis)
	normalizedMood := strings.ToLower(strings.TrimSpace(currentMood))
	var correlation *MoodCorrelation
	for i := range patterns.MoodCorrelations {
		if patterns.MoodCorrelations[i].Mood == normalizedMood {
			correlation = &patterns.MoodCorrelations[i]
			break
		}
	}
	if correlation == nil || correlation.PreferredType == "" {
		return selectDoorsWithRand(pool, count, rng)
	}

	preferredType := TaskType(correlation.PreferredType)
	preferredEffort := TaskEffort(correlation.PreferredEffort)

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

		// Combined score: diversity + mood alignment
		score := DiversityScore(candidate) + MoodAlignmentScore(candidate, preferredType, preferredEffort)

		if score > bestScore {
			bestScore = score
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		} else if score == bestScore && rng.IntN(i+1) == 0 {
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		}
	}

	// Diversity floor enforcement: if ALL tasks match preferred type, swap one out
	matchCount := 0
	for _, t := range bestSet {
		if t.Type == preferredType {
			matchCount++
		}
	}
	if matchCount == count {
		// Find a non-matching task from pool to swap in (excluding tasks already in bestSet)
		bestSetIDs := make(map[string]bool, count)
		for _, t := range bestSet {
			bestSetIDs[t.ID] = true
		}
		var nonMatching []*Task
		for _, t := range available {
			if t.Type != preferredType && !bestSetIDs[t.ID] {
				nonMatching = append(nonMatching, t)
			}
		}
		if len(nonMatching) > 0 {
			// Replace the last task with a random non-matching one
			replacement := nonMatching[rng.IntN(len(nonMatching))]
			bestSet[count-1] = replacement
		}
		// If no non-matching tasks exist, accept the all-matching set
	}

	for _, t := range bestSet {
		pool.MarkRecentlyShown(t.ID)
	}
	return bestSet
}
