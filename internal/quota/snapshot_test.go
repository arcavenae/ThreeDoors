package quota

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildSnapshotRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 30, 14, 0, 0, 0, time.UTC)
	windowStart := now.Add(-5 * time.Hour)
	windowEnd := now

	snap := UsageSnapshot{
		Window: WindowUsage{
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
		},
		Budget:         PlanMax5x,
		UsagePercent:   85.5,
		TokensConsumed: 75_240,
		TokenBudget:    88_000,
	}

	eval := EvaluationResult{
		Triggered: true,
		ActiveTier: &Tier{
			Percent:    80,
			Label:      "yellow",
			Suggestion: "reduce heartbeat",
		},
		IsPeak: true,
	}

	attr := Attribution{
		TotalTokens: 75_240,
		Agents: []AgentUsage{
			{
				AgentName:    "merge-queue",
				TierLabel:    "P1",
				TotalTokens:  40_000,
				UsagePercent: 53.2,
			},
			{
				AgentName:    "worker-alpha",
				TierLabel:    "P0",
				TotalTokens:  35_240,
				UsagePercent: 46.8,
			},
		},
	}

	record := BuildSnapshotRecord(snap, eval, attr, now)

	if record.Timestamp != "2026-03-30T14:00:00Z" {
		t.Errorf("Timestamp = %q, want 2026-03-30T14:00:00Z", record.Timestamp)
	}
	if record.UsagePercent != 85.5 {
		t.Errorf("UsagePercent = %f, want 85.5", record.UsagePercent)
	}
	if record.TokensConsumed != 75_240 {
		t.Errorf("TokensConsumed = %d, want 75240", record.TokensConsumed)
	}
	if record.TokenBudget != 88_000 {
		t.Errorf("TokenBudget = %d, want 88000", record.TokenBudget)
	}
	if record.ActiveTierLabel != "yellow" {
		t.Errorf("ActiveTierLabel = %q, want yellow", record.ActiveTierLabel)
	}
	if !record.IsPeak {
		t.Error("IsPeak = false, want true")
	}
	if record.WindowStart != "2026-03-30T09:00:00Z" {
		t.Errorf("WindowStart = %q, want 2026-03-30T09:00:00Z", record.WindowStart)
	}
	if record.EstimatedReset != "2026-03-30T14:00:00Z" {
		t.Errorf("EstimatedReset = %q, want 2026-03-30T14:00:00Z", record.EstimatedReset)
	}
	if len(record.AgentBreakdown) != 2 {
		t.Fatalf("AgentBreakdown len = %d, want 2", len(record.AgentBreakdown))
	}
	if record.AgentBreakdown[0].AgentName != "merge-queue" {
		t.Errorf("AgentBreakdown[0].AgentName = %q, want merge-queue", record.AgentBreakdown[0].AgentName)
	}
	if record.AgentBreakdown[1].TotalTokens != 35_240 {
		t.Errorf("AgentBreakdown[1].TotalTokens = %d, want 35240", record.AgentBreakdown[1].TotalTokens)
	}
}

func TestBuildSnapshotRecord_NoTierTriggered(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 30, 22, 0, 0, 0, time.UTC)
	snap := UsageSnapshot{
		Window: WindowUsage{
			WindowStart: now.Add(-5 * time.Hour),
			WindowEnd:   now,
		},
		UsagePercent:   30.0,
		TokensConsumed: 26_400,
		TokenBudget:    88_000,
	}
	eval := EvaluationResult{Triggered: false, IsPeak: false}
	attr := Attribution{TotalTokens: 26_400}

	record := BuildSnapshotRecord(snap, eval, attr, now)

	if record.ActiveTierLabel != "" {
		t.Errorf("ActiveTierLabel = %q, want empty", record.ActiveTierLabel)
	}
	if record.IsPeak {
		t.Error("IsPeak = true, want false")
	}
	if len(record.AgentBreakdown) != 0 {
		t.Errorf("AgentBreakdown len = %d, want 0", len(record.AgentBreakdown))
	}
}

func TestSnapshotRecord_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	record := SnapshotRecord{
		Timestamp:       "2026-03-30T14:00:00Z",
		UsagePercent:    85.5,
		TokensConsumed:  75_240,
		TokenBudget:     88_000,
		ActiveTierLabel: "yellow",
		IsPeak:          true,
		WindowStart:     "2026-03-30T09:00:00Z",
		EstimatedReset:  "2026-03-30T14:00:00Z",
		AgentBreakdown: []AgentSnapshotEntry{
			{AgentName: "merge-queue", TierLabel: "P1", TotalTokens: 40_000, UsagePercent: 53.2},
		},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify it's valid single-line JSON (JSONL compatible).
	if strings.Contains(string(data), "\n") {
		t.Error("JSON output contains newlines — not JSONL compatible")
	}

	var decoded SnapshotRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Timestamp != record.Timestamp {
		t.Errorf("Timestamp = %q, want %q", decoded.Timestamp, record.Timestamp)
	}
	if decoded.UsagePercent != record.UsagePercent {
		t.Errorf("UsagePercent = %f, want %f", decoded.UsagePercent, record.UsagePercent)
	}
	if decoded.ActiveTierLabel != record.ActiveTierLabel {
		t.Errorf("ActiveTierLabel = %q, want %q", decoded.ActiveTierLabel, record.ActiveTierLabel)
	}
	if len(decoded.AgentBreakdown) != 1 {
		t.Fatalf("AgentBreakdown len = %d, want 1", len(decoded.AgentBreakdown))
	}
	if decoded.AgentBreakdown[0].AgentName != "merge-queue" {
		t.Errorf("AgentBreakdown[0].AgentName = %q, want merge-queue", decoded.AgentBreakdown[0].AgentName)
	}
}

func TestSnapshotWriter_Write(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.jsonl")
	writer := NewSnapshotWriter(path)

	record := SnapshotRecord{
		Timestamp:       "2026-03-30T14:00:00Z",
		UsagePercent:    85.5,
		TokensConsumed:  75_240,
		TokenBudget:     88_000,
		ActiveTierLabel: "yellow",
		IsPeak:          true,
		WindowStart:     "2026-03-30T09:00:00Z",
		EstimatedReset:  "2026-03-30T14:00:00Z",
		AgentBreakdown:  []AgentSnapshotEntry{},
	}

	if err := writer.Write(record); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}

	var decoded SnapshotRecord
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Unmarshal line: %v", err)
	}
	if decoded.UsagePercent != 85.5 {
		t.Errorf("UsagePercent = %f, want 85.5", decoded.UsagePercent)
	}
}

func TestSnapshotWriter_MultipleWrites(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "snapshots.jsonl")
	writer := NewSnapshotWriter(path)

	for i := range 3 {
		record := SnapshotRecord{
			Timestamp:      time.Date(2026, 3, 30, 14+i, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UsagePercent:   float64(30 + i*20),
			TokensConsumed: int64(26_400 + i*17_600),
			TokenBudget:    88_000,
			AgentBreakdown: []AgentSnapshotEntry{},
		}
		if err := writer.Write(record); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}

	// Verify each line is valid JSON.
	for i, line := range lines {
		var record SnapshotRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("line %d: Unmarshal: %v", i, err)
		}
	}
}

func TestSnapshotWriter_CreatesDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "snapshots.jsonl")
	writer := NewSnapshotWriter(path)

	record := SnapshotRecord{
		Timestamp:      "2026-03-30T14:00:00Z",
		UsagePercent:   50.0,
		TokensConsumed: 44_000,
		TokenBudget:    88_000,
		AgentBreakdown: []AgentSnapshotEntry{},
	}

	if err := writer.Write(record); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSnapshotRecord_JqCompatibility(t *testing.T) {
	t.Parallel()

	// Simulate multiple snapshots as JSONL and verify each line
	// parses independently (jq processes line-by-line).
	records := []SnapshotRecord{
		{
			Timestamp:       "2026-03-30T14:00:00Z",
			UsagePercent:    85.5,
			TokensConsumed:  75_240,
			TokenBudget:     88_000,
			ActiveTierLabel: "yellow",
			IsPeak:          true,
			WindowStart:     "2026-03-30T09:00:00Z",
			EstimatedReset:  "2026-03-30T14:00:00Z",
			AgentBreakdown: []AgentSnapshotEntry{
				{AgentName: "merge-queue", TierLabel: "P1", TotalTokens: 40_000, UsagePercent: 53.2},
			},
		},
		{
			Timestamp:       "2026-03-30T14:05:00Z",
			UsagePercent:    87.0,
			TokensConsumed:  76_560,
			TokenBudget:     88_000,
			ActiveTierLabel: "yellow",
			IsPeak:          true,
			WindowStart:     "2026-03-30T09:05:00Z",
			EstimatedReset:  "2026-03-30T14:05:00Z",
			AgentBreakdown:  []AgentSnapshotEntry{},
		},
	}

	var jsonl strings.Builder
	for _, r := range records {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		jsonl.Write(data)
		jsonl.WriteByte('\n')
	}

	// Parse back line-by-line (simulating jq behavior).
	lines := strings.Split(strings.TrimSpace(jsonl.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	for i, line := range lines {
		var decoded SnapshotRecord
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Errorf("line %d: Unmarshal: %v", i, err)
		}
		// Verify all AC2 fields are present.
		if decoded.Timestamp == "" {
			t.Errorf("line %d: missing timestamp", i)
		}
		if decoded.WindowStart == "" {
			t.Errorf("line %d: missing window_start", i)
		}
		if decoded.EstimatedReset == "" {
			t.Errorf("line %d: missing estimated_reset", i)
		}
	}
}
