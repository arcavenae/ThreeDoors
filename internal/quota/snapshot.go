// Package quota provides usage monitoring data capture for Claude Code
// token consumption tracking. It records usage snapshots to JSONL files
// for retrospective analysis and trend identification.
package quota

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// SnapshotFileName is the default filename for quota usage snapshots.
	SnapshotFileName = "quota-usage.jsonl"
)

// ThresholdTier indicates the current quota consumption severity level.
type ThresholdTier string

const (
	TierNormal    ThresholdTier = "normal"    // <70% usage
	TierElevated  ThresholdTier = "elevated"  // 70-85% usage
	TierCritical  ThresholdTier = "critical"  // 85-95% usage
	TierEmergency ThresholdTier = "emergency" // >95% usage
)

// TierFromUsage returns the threshold tier for a given usage percentage.
func TierFromUsage(pct float64) ThresholdTier {
	switch {
	case pct >= 95:
		return TierEmergency
	case pct >= 85:
		return TierCritical
	case pct >= 70:
		return TierElevated
	default:
		return TierNormal
	}
}

// SnapshotAgentUsage records token consumption for a single agent within a snapshot.
type SnapshotAgentUsage struct {
	Name         string `json:"name"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	BilledTokens int64  `json:"billed_tokens"`
}

// QuotaSnapshot captures a point-in-time quota usage reading.
// One snapshot is recorded per quota check. Fields are designed for
// JSONL storage and jq-based analysis by retrospector.
//
// OTEL mapping (Phase 3 Marvel — see docs/operations/otel-snapshot-mapping.md):
//
//	Timestamp       → span timestamp
//	WindowUsagePct  → dark_factory.phase.token_usage (gauge)
//	AgentBreakdown  → dark_factory.phase.token_usage{agent=<name>}
//	ThresholdTier   → dark_factory.quota.tier (attribute)
//	PeakHours       → dark_factory.quota.peak_hours (attribute)
type QuotaSnapshot struct {
	// Type discriminator for mixed-type JSONL files.
	Type string `json:"type"`

	// Timestamp is when this snapshot was recorded (ISO 8601 UTC).
	Timestamp time.Time `json:"timestamp"`

	// WindowUsagePct is the estimated usage percentage within the current
	// rolling 5-hour window (0-100+, can exceed 100 if over quota).
	WindowUsagePct float64 `json:"window_usage_pct"`

	// WindowStartTime is the start of the current rolling 5-hour window.
	WindowStartTime time.Time `json:"window_start_time"`

	// EstimatedResetTime is the estimated time when the window resets.
	EstimatedResetTime time.Time `json:"estimated_reset_time"`

	// AgentBreakdown lists per-agent token consumption in this window.
	AgentBreakdown []SnapshotAgentUsage `json:"agent_breakdown"`

	// TotalBilledTokens is the sum of all agents' billed tokens.
	TotalBilledTokens int64 `json:"total_billed_tokens"`

	// ThresholdTier is the active consumption severity tier.
	ThresholdTier ThresholdTier `json:"threshold_tier"`

	// PeakHours indicates whether this snapshot was taken during
	// peak hours (05:00-11:00 PT / 13:00-19:00 UTC).
	PeakHours bool `json:"peak_hours"`
}

// SnapshotWriter appends usage snapshots to a JSONL file.
type SnapshotWriter struct {
	path string
}

// NewSnapshotWriter creates a SnapshotWriter for the given directory.
// Snapshots are written to <dir>/quota-usage.jsonl.
func NewSnapshotWriter(dir string) *SnapshotWriter {
	return &SnapshotWriter{
		path: filepath.Join(dir, SnapshotFileName),
	}
}

// Write appends a usage snapshot to the JSONL file.
// The snapshot's Type field is set to "quota_snapshot" before writing.
// Creates the file if it does not exist. Uses atomic append via
// O_APPEND to prevent partial writes from concurrent processes.
func (w *SnapshotWriter) Write(snap QuotaSnapshot) error {
	snap.Type = "quota_snapshot"

	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open snapshot file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on append

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	return nil
}

// Path returns the path to the snapshot JSONL file.
func (w *SnapshotWriter) Path() string {
	return w.path
}

// SnapshotReader reads usage snapshots from a JSONL file.
type SnapshotReader struct {
	path string
}

// NewSnapshotReader creates a SnapshotReader for the given directory.
func NewSnapshotReader(dir string) *SnapshotReader {
	return &SnapshotReader{
		path: filepath.Join(dir, SnapshotFileName),
	}
}

// ReadAll returns all snapshots from the file.
// Returns nil, nil for missing or empty files.
// Corrupted lines are silently skipped.
func (r *SnapshotReader) ReadAll() ([]QuotaSnapshot, error) {
	return r.ReadSince(time.Time{})
}

// ReadSince returns snapshots with Timestamp at or after the given time.
// Returns nil, nil for missing or empty files.
func (r *SnapshotReader) ReadSince(since time.Time) ([]QuotaSnapshot, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open snapshot file: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only

	var snapshots []QuotaSnapshot
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var snap QuotaSnapshot
		if err := json.Unmarshal(line, &snap); err != nil {
			continue // skip corrupted lines
		}
		if snap.Type != "quota_snapshot" {
			continue
		}
		if !snap.Timestamp.Before(since) {
			snapshots = append(snapshots, snap)
		}
	}
	if err := scanner.Err(); err != nil {
		return snapshots, fmt.Errorf("scan snapshot file: %w", err)
	}

	return snapshots, nil
}
