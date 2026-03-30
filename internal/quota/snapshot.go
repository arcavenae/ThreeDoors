package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SnapshotRecord is the JSONL record written to the operational data store.
// Each record is one line in docs/operations/quota-usage.jsonl and represents
// a point-in-time aggregation of quota usage. Fields are derived/aggregated
// metrics — raw per-interaction data stays in the source JSONL session files.
//
// OTEL Metric Mapping (R-016, Phase 3 Marvel):
//
//	Timestamp        → observation timestamp on all metrics
//	UsagePercent     → dark_factory.quota.usage_ratio (gauge, 0.0–1.0)
//	TokensConsumed   → gen_ai.client.token.usage (counter, token_type=*)
//	TokenBudget      → dark_factory.quota.budget (gauge)
//	ActiveTierLabel  → dark_factory.quota.tier (attribute on usage_ratio)
//	IsPeak           → dark_factory.quota.peak (attribute on usage_ratio)
//	WindowStart      → dark_factory.quota.window_start (attribute)
//	EstimatedReset   → dark_factory.quota.reset_time (attribute)
//	AgentBreakdown   → dark_factory.phase.token_usage (histogram, agent=<name>)
type SnapshotRecord struct {
	// Timestamp is the ISO 8601 time the snapshot was taken.
	Timestamp string `json:"timestamp"`

	// UsagePercent is the current usage as a percentage of budget (0.0–100.0).
	UsagePercent float64 `json:"usage_percent"`

	// TokensConsumed is the total tokens consumed within the window.
	TokensConsumed int64 `json:"tokens_consumed"`

	// TokenBudget is the estimated token budget for the plan tier.
	TokenBudget int64 `json:"token_budget"`

	// ActiveTierLabel is the highest triggered warning tier ("green", "yellow",
	// "orange", "red") or empty if below all thresholds.
	ActiveTierLabel string `json:"active_tier_label"`

	// IsPeak indicates whether the snapshot was taken during Anthropic peak hours.
	IsPeak bool `json:"is_peak"`

	// WindowStart is the ISO 8601 start time of the rolling usage window.
	WindowStart string `json:"window_start"`

	// EstimatedReset is the ISO 8601 estimated time when the window rolls over.
	EstimatedReset string `json:"estimated_reset"`

	// AgentBreakdown contains per-agent token consumption within the window.
	AgentBreakdown []AgentSnapshotEntry `json:"agent_breakdown"`
}

// AgentSnapshotEntry is a per-agent summary within a snapshot.
type AgentSnapshotEntry struct {
	AgentName    string  `json:"agent_name"`
	TierLabel    string  `json:"tier_label"`
	TotalTokens  int64   `json:"total_tokens"`
	UsagePercent float64 `json:"usage_percent"`
}

// BuildSnapshotRecord creates a SnapshotRecord from the components produced by
// the quota pipeline: a UsageSnapshot, a threshold EvaluationResult, and an
// Attribution report.
func BuildSnapshotRecord(
	snap UsageSnapshot,
	eval EvaluationResult,
	attr Attribution,
	now time.Time,
) SnapshotRecord {
	tierLabel := ""
	if eval.Triggered && eval.ActiveTier != nil {
		tierLabel = eval.ActiveTier.Label
	}

	agents := make([]AgentSnapshotEntry, len(attr.Agents))
	for i, a := range attr.Agents {
		agents[i] = AgentSnapshotEntry{
			AgentName:    a.AgentName,
			TierLabel:    a.TierLabel,
			TotalTokens:  a.TotalTokens,
			UsagePercent: a.UsagePercent,
		}
	}

	return SnapshotRecord{
		Timestamp:       now.UTC().Format(time.RFC3339),
		UsagePercent:    snap.UsagePercent,
		TokensConsumed:  snap.TokensConsumed,
		TokenBudget:     snap.TokenBudget,
		ActiveTierLabel: tierLabel,
		IsPeak:          eval.IsPeak,
		WindowStart:     snap.Window.WindowStart.UTC().Format(time.RFC3339),
		EstimatedReset:  snap.Window.WindowEnd.UTC().Format(time.RFC3339),
		AgentBreakdown:  agents,
	}
}

// SnapshotWriter appends usage snapshots to a JSONL file using atomic writes.
type SnapshotWriter struct {
	// Path is the JSONL file to append to.
	Path string
}

// NewSnapshotWriter creates a SnapshotWriter targeting the given file path.
func NewSnapshotWriter(path string) *SnapshotWriter {
	return &SnapshotWriter{Path: path}
}

// Write serializes a SnapshotRecord to JSON and appends it as a single line
// to the JSONL file. The append is followed by an fsync to ensure durability.
func (w *SnapshotWriter) Write(record SnapshotRecord) error {
	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	line = append(line, '\n')

	if err := os.MkdirAll(filepath.Dir(w.Path), 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}

	f, err := os.OpenFile(w.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open snapshot file: %w", err)
	}

	writeErr := func() error {
		if _, err := f.Write(line); err != nil {
			return fmt.Errorf("append snapshot: %w", err)
		}
		return f.Sync()
	}()

	if closeErr := f.Close(); closeErr != nil && writeErr == nil {
		return fmt.Errorf("close snapshot file: %w", closeErr)
	}
	return writeErr
}

// DefaultSnapshotPath returns the default path for quota usage snapshots,
// following the docs/operations/ convention from Epic 67.
func DefaultSnapshotPath() string {
	return "docs/operations/quota-usage.jsonl"
}
