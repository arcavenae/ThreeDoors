package core

import (
	"math"
	"sort"
	"strings"
)

// DuplicatePair represents two tasks flagged as potential duplicates.
type DuplicatePair struct {
	TaskA      *Task
	TaskB      *Task
	Similarity float64
}

// Decision constants for duplicate resolution.
const (
	DecisionDuplicate = "duplicate"
	DecisionDistinct  = "distinct"
)

// LevenshteinDistance computes the edit distance between two strings.
func LevenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	// Guard allocation arithmetic from integer overflow.
	if len(b) > math.MaxInt-1 {
		return len(b)
	}
	if len(a) > math.MaxInt-1 {
		return len(a)
	}

	// Use two rows instead of full matrix for space efficiency.
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,    // insertion
				prev[j]+1,      // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

// TextSimilarity returns a normalized similarity score between 0.0 and 1.0
// for two text strings. Case-insensitive with whitespace normalization.
func TextSimilarity(a, b string) float64 {
	a = normalizeText(a)
	b = normalizeText(b)

	maxLen := max(len(a), len(b))
	if maxLen == 0 {
		return 1.0
	}

	dist := LevenshteinDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// DetectDuplicates finds potential cross-provider duplicate pairs among tasks.
// Only tasks from different providers are compared. Results are sorted by
// similarity score descending, with ties broken by lexicographic task ID.
func DetectDuplicates(tasks []*Task, threshold float64) []DuplicatePair {
	if len(tasks) < 2 {
		return nil
	}

	// Sort tasks by ID for deterministic iteration order.
	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	var pairs []DuplicatePair
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			ta, tb := sorted[i], sorted[j]

			// Only compare tasks from different providers.
			if ta.SourceProvider == tb.SourceProvider {
				continue
			}

			sim := TextSimilarity(ta.Text, tb.Text)
			if sim >= threshold {
				pairs = append(pairs, DuplicatePair{
					TaskA:      ta,
					TaskB:      tb,
					Similarity: sim,
				})
			}
		}
	}

	// Sort by score descending; break ties by task IDs for determinism.
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].Similarity != pairs[j].Similarity {
			return pairs[i].Similarity > pairs[j].Similarity
		}
		if pairs[i].TaskA.ID != pairs[j].TaskA.ID {
			return pairs[i].TaskA.ID < pairs[j].TaskA.ID
		}
		return pairs[i].TaskB.ID < pairs[j].TaskB.ID
	})

	return pairs
}

// normalizeText lowercases and collapses whitespace for comparison.
func normalizeText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
