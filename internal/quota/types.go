// Package quota provides a read-only parser for Claude Code JSONL session
// files, extracting token usage data for quota monitoring.
package quota

import "time"

// TokenCount holds the breakdown of tokens for a single interaction.
type TokenCount struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

// Total returns the sum of all token fields.
func (tc TokenCount) Total() int64 {
	return tc.InputTokens + tc.OutputTokens + tc.CacheCreationInputTokens + tc.CacheReadInputTokens
}

// Interaction represents a single assistant response with token usage.
type Interaction struct {
	SessionID  string     `json:"session_id"`
	Timestamp  time.Time  `json:"timestamp"`
	Model      string     `json:"model"`
	Tokens     TokenCount `json:"tokens"`
	HasToolUse bool       `json:"has_tool_use"`
	SourcePath string     `json:"source_path"`
}

// SessionUsage aggregates token usage for a single session.
type SessionUsage struct {
	SessionID    string     `json:"session_id"`
	Interactions int        `json:"interactions"`
	Tokens       TokenCount `json:"tokens"`
	FirstSeen    time.Time  `json:"first_seen"`
	LastSeen     time.Time  `json:"last_seen"`
}

// WindowUsage aggregates token usage over a rolling time window.
type WindowUsage struct {
	WindowStart time.Time      `json:"window_start"`
	WindowEnd   time.Time      `json:"window_end"`
	Tokens      TokenCount     `json:"tokens"`
	Sessions    []SessionUsage `json:"sessions"`
}

// PlanBudget defines the estimated token budget for a plan tier.
type PlanBudget struct {
	Name        string `json:"name"`
	TokenBudget int64  `json:"token_budget"`
}

// UsageSnapshot is the top-level result returned by the parser, combining
// windowed usage with budget estimation.
type UsageSnapshot struct {
	Window         WindowUsage `json:"window"`
	Budget         PlanBudget  `json:"budget"`
	UsagePercent   float64     `json:"usage_percent"`
	TokensConsumed int64       `json:"tokens_consumed"`
	TokenBudget    int64       `json:"token_budget"`
}

// Predefined plan budgets based on community estimates (Q-006).
var (
	PlanMax5x  = PlanBudget{Name: "Max 5x", TokenBudget: 88_000}
	PlanMax20x = PlanBudget{Name: "Max 20x", TokenBudget: 220_000}
)
