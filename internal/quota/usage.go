package quota

import "time"

// WindowUsage represents aggregated token usage within a time window.
// This is the minimal type needed by the warning engine (Story 76.3).
// Story 76.1 will extend this with JSONL parsing and session-level breakdowns.
type WindowUsage struct {
	TotalTokens  int64     // Total tokens consumed in the window
	InputTokens  int64     // Input tokens consumed
	OutputTokens int64     // Output tokens consumed
	WindowStart  time.Time // Start of the rolling window (UTC)
	WindowEnd    time.Time // End of the rolling window (UTC)
	PlanBudget   int64     // Estimated token budget for the plan window
	SessionCount int       // Number of sessions included
}

// UsagePercent returns the usage as a percentage of the plan budget.
// Returns 0 if PlanBudget is zero to avoid division by zero.
func (w WindowUsage) UsagePercent() float64 {
	if w.PlanBudget == 0 {
		return 0
	}
	return float64(w.TotalTokens) / float64(w.PlanBudget) * 100
}

// RemainingTokens returns the estimated remaining tokens in the budget.
// Returns 0 if usage exceeds budget.
func (w WindowUsage) RemainingTokens() int64 {
	remaining := w.PlanBudget - w.TotalTokens
	if remaining < 0 {
		return 0
	}
	return remaining
}

// TimeUntilReset returns the duration until the window resets.
// Returns 0 if the window has already passed.
func (w WindowUsage) TimeUntilReset(now time.Time) time.Duration {
	d := w.WindowEnd.Sub(now)
	if d < 0 {
		return 0
	}
	return d
}
