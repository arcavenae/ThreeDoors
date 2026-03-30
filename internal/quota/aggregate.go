package quota

import "time"

// DefaultWindow is the rolling window duration for token aggregation.
const DefaultWindow = 5 * time.Hour

// AggregateWindow filters interactions to a rolling time window ending at
// windowEnd and groups them by session.
func AggregateWindow(interactions []Interaction, windowEnd time.Time, window time.Duration) WindowUsage {
	windowStart := windowEnd.Add(-window)

	sessionMap := make(map[string]*SessionUsage)
	var total TokenCount

	for _, ix := range interactions {
		if ix.Timestamp.Before(windowStart) || ix.Timestamp.After(windowEnd) {
			continue
		}

		total.InputTokens += ix.Tokens.InputTokens
		total.OutputTokens += ix.Tokens.OutputTokens
		total.CacheCreationInputTokens += ix.Tokens.CacheCreationInputTokens
		total.CacheReadInputTokens += ix.Tokens.CacheReadInputTokens

		su, ok := sessionMap[ix.SessionID]
		if !ok {
			su = &SessionUsage{
				SessionID: ix.SessionID,
				FirstSeen: ix.Timestamp,
				LastSeen:  ix.Timestamp,
			}
			sessionMap[ix.SessionID] = su
		}
		su.Interactions++
		su.Tokens.InputTokens += ix.Tokens.InputTokens
		su.Tokens.OutputTokens += ix.Tokens.OutputTokens
		su.Tokens.CacheCreationInputTokens += ix.Tokens.CacheCreationInputTokens
		su.Tokens.CacheReadInputTokens += ix.Tokens.CacheReadInputTokens

		if ix.Timestamp.Before(su.FirstSeen) {
			su.FirstSeen = ix.Timestamp
		}
		if ix.Timestamp.After(su.LastSeen) {
			su.LastSeen = ix.Timestamp
		}
	}

	sessions := make([]SessionUsage, 0, len(sessionMap))
	for _, su := range sessionMap {
		sessions = append(sessions, *su)
	}

	return WindowUsage{
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		Tokens:      total,
		Sessions:    sessions,
	}
}

// Snapshot produces a full UsageSnapshot by aggregating interactions within
// the window and computing budget percentage against the given plan.
func Snapshot(interactions []Interaction, windowEnd time.Time, window time.Duration, budget PlanBudget) UsageSnapshot {
	wu := AggregateWindow(interactions, windowEnd, window)
	consumed := wu.Tokens.Total()

	var pct float64
	if budget.TokenBudget > 0 {
		pct = float64(consumed) / float64(budget.TokenBudget) * 100
	}

	return UsageSnapshot{
		Window:         wu,
		Budget:         budget,
		UsagePercent:   pct,
		TokensConsumed: consumed,
		TokenBudget:    budget.TokenBudget,
	}
}
