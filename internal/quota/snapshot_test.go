package quota

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTierFromUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pct  float64
		want ThresholdTier
	}{
		{"zero usage", 0, TierNormal},
		{"low usage", 50, TierNormal},
		{"below elevated boundary", 69.9, TierNormal},
		{"elevated boundary", 70, TierElevated},
		{"mid elevated", 80, TierElevated},
		{"below critical boundary", 84.9, TierElevated},
		{"critical boundary", 85, TierCritical},
		{"mid critical", 90, TierCritical},
		{"below emergency boundary", 94.9, TierCritical},
		{"emergency boundary", 95, TierEmergency},
		{"over quota", 110, TierEmergency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TierFromUsage(tt.pct)
			if got != tt.want {
				t.Errorf("TierFromUsage(%v) = %v, want %v", tt.pct, got, tt.want)
			}
		})
	}
}

func makeSnapshot(ts time.Time, pct float64, agents []SnapshotAgentUsage) QuotaSnapshot {
	var total int64
	for _, a := range agents {
		total += a.BilledTokens
	}
	windowStart := ts.Add(-3 * time.Hour)
	return QuotaSnapshot{
		Timestamp:          ts,
		WindowUsagePct:     pct,
		WindowStartTime:    windowStart,
		EstimatedResetTime: windowStart.Add(5 * time.Hour),
		AgentBreakdown:     agents,
		TotalBilledTokens:  total,
		ThresholdTier:      TierFromUsage(pct),
		PeakHours:          false,
	}
}

func TestSnapshotWriterWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	ts := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	snap := makeSnapshot(ts, 65.5, []SnapshotAgentUsage{
		{Name: "supervisor", InputTokens: 1000, OutputTokens: 500, BilledTokens: 1500},
		{Name: "worker-1", InputTokens: 2000, OutputTokens: 1000, BilledTokens: 3000},
	})

	if err := w.Write(snap); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, SnapshotFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Verify it's valid JSON
	var decoded QuotaSnapshot
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &decoded); err != nil {
		t.Fatalf("Unmarshal written data: %v", err)
	}

	// Type must be set automatically
	if decoded.Type != "quota_snapshot" {
		t.Errorf("Type = %q, want %q", decoded.Type, "quota_snapshot")
	}

	if decoded.WindowUsagePct != 65.5 {
		t.Errorf("WindowUsagePct = %v, want 65.5", decoded.WindowUsagePct)
	}

	if decoded.ThresholdTier != TierNormal {
		t.Errorf("ThresholdTier = %v, want %v", decoded.ThresholdTier, TierNormal)
	}

	if len(decoded.AgentBreakdown) != 2 {
		t.Fatalf("AgentBreakdown len = %d, want 2", len(decoded.AgentBreakdown))
	}

	if decoded.TotalBilledTokens != 4500 {
		t.Errorf("TotalBilledTokens = %d, want 4500", decoded.TotalBilledTokens)
	}
}

func TestSnapshotWriterAppendsMultiple(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	ts := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	for i := range 3 {
		snap := makeSnapshot(ts.Add(time.Duration(i)*time.Hour), float64(50+i*10), nil)
		if err := w.Write(snap); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, SnapshotFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3", len(lines))
	}

	// Each line must be valid JSON
	for i, line := range lines {
		var snap QuotaSnapshot
		if err := json.Unmarshal([]byte(line), &snap); err != nil {
			t.Errorf("line %d: invalid JSON: %v", i, err)
		}
	}
}

func TestSnapshotWriterSetsType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	snap := makeSnapshot(time.Now().UTC(), 42, nil)
	snap.Type = "should_be_overwritten"

	if err := w.Write(snap); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, SnapshotFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var decoded QuotaSnapshot
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Type != "quota_snapshot" {
		t.Errorf("Type = %q, want %q", decoded.Type, "quota_snapshot")
	}
}

func TestSnapshotReaderReadAll(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)
	r := NewSnapshotReader(dir)

	ts := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	agents := []SnapshotAgentUsage{
		{Name: "merge-queue", InputTokens: 500, OutputTokens: 200, BilledTokens: 700},
	}

	for i := range 5 {
		snap := makeSnapshot(ts.Add(time.Duration(i)*time.Hour), float64(40+i*12), agents)
		if err := w.Write(snap); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}

	snapshots, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(snapshots) != 5 {
		t.Fatalf("got %d snapshots, want 5", len(snapshots))
	}

	// Verify ordering preserved
	if snapshots[0].WindowUsagePct != 40 {
		t.Errorf("first snapshot pct = %v, want 40", snapshots[0].WindowUsagePct)
	}
	if snapshots[4].WindowUsagePct != 88 {
		t.Errorf("last snapshot pct = %v, want 88", snapshots[4].WindowUsagePct)
	}
}

func TestSnapshotReaderReadSince(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)
	r := NewSnapshotReader(dir)

	base := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	for i := range 5 {
		snap := makeSnapshot(base.Add(time.Duration(i)*time.Hour), float64(50), nil)
		if err := w.Write(snap); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}

	since := base.Add(2 * time.Hour)
	snapshots, err := r.ReadSince(since)
	if err != nil {
		t.Fatalf("ReadSince: %v", err)
	}

	if len(snapshots) != 3 {
		t.Fatalf("got %d snapshots, want 3 (hours 2,3,4)", len(snapshots))
	}
}

func TestSnapshotReaderMissingFile(t *testing.T) {
	t.Parallel()

	r := NewSnapshotReader(t.TempDir())
	snapshots, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll on missing file: %v", err)
	}
	if snapshots != nil {
		t.Errorf("got %v, want nil for missing file", snapshots)
	}
}

func TestSnapshotReaderCorruptedLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	ts := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	if err := w.Write(makeSnapshot(ts, 50, nil)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Append a corrupted line
	f, err := os.OpenFile(filepath.Join(dir, SnapshotFileName), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open for corruption: %v", err)
	}
	if _, err := f.Write([]byte("{invalid json\n")); err != nil {
		t.Fatalf("write corruption: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close after corruption: %v", err)
	}

	// Append a valid line after the corrupted one
	if err := w.Write(makeSnapshot(ts.Add(time.Hour), 60, nil)); err != nil {
		t.Fatalf("Write after corruption: %v", err)
	}

	r := NewSnapshotReader(dir)
	snapshots, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll with corrupted data: %v", err)
	}

	if len(snapshots) != 2 {
		t.Errorf("got %d snapshots, want 2 (skipping corrupted line)", len(snapshots))
	}
}

func TestSnapshotReaderIgnoresNonSnapshotTypes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Write a line with a different type discriminator
	f, err := os.OpenFile(filepath.Join(dir, SnapshotFileName), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	other := map[string]any{"type": "other_event", "timestamp": "2026-03-30T12:00:00Z"}
	data, _ := json.Marshal(other)
	if _, err := f.Write(append(data, '\n')); err != nil {
		t.Fatalf("write other: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close after write: %v", err)
	}

	// Write a valid snapshot after it
	w := NewSnapshotWriter(dir)
	ts := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	if err := w.Write(makeSnapshot(ts, 50, nil)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	r := NewSnapshotReader(dir)
	snapshots, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("got %d snapshots, want 1 (ignoring non-snapshot type)", len(snapshots))
	}
}

func TestSnapshotJSONLParsableWithJQ(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	ts := time.Date(2026, 3, 30, 15, 30, 0, 0, time.UTC)
	snap := makeSnapshot(ts, 72.3, []SnapshotAgentUsage{
		{Name: "supervisor", InputTokens: 5000, OutputTokens: 2000, BilledTokens: 7000},
		{Name: "merge-queue", InputTokens: 1000, OutputTokens: 300, BilledTokens: 1300},
		{Name: "worker-abc", InputTokens: 8000, OutputTokens: 4000, BilledTokens: 12000},
	})
	snap.PeakHours = true

	if err := w.Write(snap); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, SnapshotFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	line := strings.TrimSpace(string(data))

	// Verify all required AC2 fields are present
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	requiredFields := []string{
		"type",
		"timestamp",
		"window_usage_pct",
		"window_start_time",
		"estimated_reset_time",
		"agent_breakdown",
		"total_billed_tokens",
		"threshold_tier",
		"peak_hours",
	}

	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing required field %q in JSON output", field)
		}
	}

	// Verify timestamp is ISO 8601
	var snap2 QuotaSnapshot
	if err := json.Unmarshal([]byte(line), &snap2); err != nil {
		t.Fatalf("Unmarshal to QuotaSnapshot: %v", err)
	}
	if snap2.Timestamp.IsZero() {
		t.Error("timestamp parsed as zero")
	}
}

func TestSnapshotWriterPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	want := filepath.Join(dir, SnapshotFileName)
	if got := w.Path(); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestSnapshotWriterCreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewSnapshotWriter(dir)

	snap := makeSnapshot(time.Now().UTC(), 50, nil)
	if err := w.Write(snap); err != nil {
		t.Fatalf("Write: %v", err)
	}

	info, err := os.Stat(w.Path())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file permissions = %o, want 600", info.Mode().Perm())
	}
}
