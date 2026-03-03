package core

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

// TimeContext holds calendar-derived time information for door selection.
// A nil or zero-value TimeContext means no calendar data is available.
type TimeContext struct {
	NextEventIn   time.Duration // Duration until next calendar event
	AvailableTime time.Duration // Duration of current free block
	NextEventName string        // Title of next event (for TUI display)
	HasCalendar   bool          // Whether calendar data is available
}

// TimeContextProvider supplies time context for door selection.
type TimeContextProvider interface {
	GetTimeContext(ctx context.Context) (*TimeContext, error)
}

// Default effort-to-time thresholds.
const (
	QuickWinThreshold = 30 * time.Minute
	MediumThreshold   = 90 * time.Minute
)

// TimeContextScore scores a set of tasks based on how well their effort levels
// match the available time block. Returns 0 when no calendar data is available.
//
// Scoring per task:
//   - Available ≤ 30 min: quick-win +2, medium +1, deep-work +0
//   - Available 30–90 min: quick-win +1, medium +2, deep-work +0
//   - Available > 90 min: quick-win +0, medium +1, deep-work +2
//   - Uncategorized effort: +1 always (neutral)
func TimeContextScore(tasks []*Task, timeCtx *TimeContext) int {
	if timeCtx == nil || !timeCtx.HasCalendar {
		return 0
	}

	score := 0
	for _, t := range tasks {
		score += effortTimeScore(t.Effort, timeCtx.AvailableTime)
	}
	return score
}

// SelectDoorsWithTimeContext picks up to count tasks from the pool, considering
// time context from calendar data in addition to diversity.
// Falls back to diversity-only selection when timeCtx is nil or has no calendar data.
func SelectDoorsWithTimeContext(pool *TaskPool, count int, timeCtx *TimeContext) []*Task {
	rng := rand.New(rand.NewPCG(uint64(time.Now().UTC().UnixNano()), 0))
	return selectDoorsWithTimeContextAndRand(pool, count, timeCtx, rng)
}

// selectDoorsWithTimeContextAndRand picks tasks with time awareness using a deterministic RNG.
func selectDoorsWithTimeContextAndRand(pool *TaskPool, count int, timeCtx *TimeContext, rng *rand.Rand) []*Task {
	// Fallback: no calendar data → standard diversity selection
	if timeCtx == nil || !timeCtx.HasCalendar {
		return selectDoorsWithRand(pool, count, rng)
	}

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
		perm := make([]*Task, len(available))
		copy(perm, available)
		for j := range count {
			k := j + rng.IntN(len(perm)-j)
			perm[j], perm[k] = perm[k], perm[j]
		}
		candidate := perm[:count]

		// Combined score: diversity + time context
		score := DiversityScore(candidate) + TimeContextScore(candidate, timeCtx)

		if score > bestScore {
			bestScore = score
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		} else if score == bestScore && rng.IntN(i+1) == 0 {
			bestSet = make([]*Task, count)
			copy(bestSet, candidate)
		}
	}

	for _, t := range bestSet {
		pool.MarkRecentlyShown(t.ID)
	}
	return bestSet
}

// FormatTimeContext returns a human-readable string for TUI display.
// Returns empty string when no calendar data or event is more than 4 hours away.
func FormatTimeContext(ctx *TimeContext) string {
	if ctx == nil || !ctx.HasCalendar || ctx.NextEventIn <= 0 {
		return ""
	}

	const maxDisplayDuration = 4 * time.Hour
	if ctx.NextEventIn > maxDisplayDuration {
		return ""
	}

	minutes := int(ctx.NextEventIn.Minutes())
	var timeStr string
	if minutes < 60 {
		timeStr = fmt.Sprintf("%d min", minutes)
	} else {
		h := minutes / 60
		m := minutes % 60
		timeStr = fmt.Sprintf("%dh %dmin", h, m)
	}

	if ctx.NextEventName != "" {
		return fmt.Sprintf("Next event in %s — %s", timeStr, ctx.NextEventName)
	}
	return fmt.Sprintf("Next event in %s", timeStr)
}

// effortTimeScore returns the score for a single task's effort given available time.
func effortTimeScore(effort TaskEffort, available time.Duration) int {
	switch {
	case available <= QuickWinThreshold:
		// Short block: prefer quick-win
		switch effort {
		case EffortQuickWin:
			return 2
		case EffortMedium:
			return 1
		case EffortDeepWork:
			return 0
		default:
			return 1
		}
	case available <= MediumThreshold:
		// Medium block: prefer medium
		switch effort {
		case EffortQuickWin:
			return 1
		case EffortMedium:
			return 2
		case EffortDeepWork:
			return 0
		default:
			return 1
		}
	default:
		// Long block: prefer deep-work
		switch effort {
		case EffortQuickWin:
			return 0
		case EffortMedium:
			return 1
		case EffortDeepWork:
			return 2
		default:
			return 1
		}
	}
}
