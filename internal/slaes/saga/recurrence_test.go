package saga

import (
	"strings"
	"testing"
	"time"
)

func TestRecurrenceTracker_Analyze_BelowThreshold(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(3)

	findings := []SagaFinding{
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/b", SagaType: SagaTypeOverlap},
	}

	patterns := rt.Analyze(findings)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns below threshold, got %d", len(patterns))
	}
}

func TestRecurrenceTracker_Analyze_MeetsThreshold(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(3)

	findings := []SagaFinding{
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/b", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/c", SagaType: SagaTypeOverlap},
	}

	patterns := rt.Analyze(findings)
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}

	p := patterns[0]
	if p.SagaType != SagaTypeOverlap {
		t.Errorf("saga type: got %q, want %q", p.SagaType, SagaTypeOverlap)
	}
	if p.BranchCount != 3 {
		t.Errorf("branch count: got %d, want 3", p.BranchCount)
	}
	if p.TotalSagas != 3 {
		t.Errorf("total sagas: got %d, want 3", p.TotalSagas)
	}
}

func TestRecurrenceTracker_Analyze_MultipleSagaTypes(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(2)

	findings := []SagaFinding{
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/b", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/c", SagaType: SagaTypeEscalationTrap},
		{Type: FindingSagaDetected, Branch: "fix/d", SagaType: SagaTypeEscalationTrap},
	}

	patterns := rt.Analyze(findings)
	if len(patterns) != 2 {
		t.Errorf("expected 2 patterns (one per saga type), got %d", len(patterns))
	}
}

func TestRecurrenceTracker_Analyze_SkipsNonDetectedFindings(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(2)

	findings := []SagaFinding{
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/b", SagaType: SagaTypeOverlap},
		{Type: FindingSagaRecur, Branch: "fix/c", SagaType: SagaTypeOverlap},
	}

	patterns := rt.Analyze(findings)
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].BranchCount != 2 {
		t.Errorf("branch count: got %d, want 2 (should skip recurrence entries)", patterns[0].BranchCount)
	}
}

func TestRecurrenceTracker_Analyze_SameBranchCountsOnce(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(2)

	findings := []SagaFinding{
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
		{Type: FindingSagaDetected, Branch: "fix/a", SagaType: SagaTypeOverlap},
	}

	patterns := rt.Analyze(findings)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns (same branch repeated, only 1 distinct), got %d", len(patterns))
	}
}

func TestScoreConfidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		count int
		want  Confidence
	}{
		{1, ConfidenceLow},
		{2, ConfidenceLow},
		{3, ConfidenceMedium},
		{4, ConfidenceMedium},
		{5, ConfidenceHigh},
		{10, ConfidenceHigh},
	}

	for _, tt := range tests {
		got := scoreConfidence(tt.count)
		if got != tt.want {
			t.Errorf("scoreConfidence(%d): got %q, want %q", tt.count, got, tt.want)
		}
	}
}

func TestProduceRecommendation_EscalationTrap(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(3)

	pattern := RecurrencePattern{
		SagaType:    SagaTypeEscalationTrap,
		BranchCount: 4,
		TotalSagas:  6,
		Branches:    []string{"fix/a", "fix/b", "fix/c", "fix/d"},
		Confidence:  ConfidenceHigh,
	}

	rec := rt.ProduceRecommendation(pattern)

	if !strings.Contains(rec.Proposal, "root cause analysis") {
		t.Errorf("proposal should mention root cause analysis: %q", rec.Proposal)
	}
	if !strings.Contains(rec.Evidence, "6 saga events") {
		t.Errorf("evidence should mention count: %q", rec.Evidence)
	}
	if !strings.Contains(rec.Evidence, "4 distinct branches") {
		t.Errorf("evidence should mention branch count: %q", rec.Evidence)
	}
}

func TestProduceRecommendation_Overlap(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(3)

	pattern := RecurrencePattern{
		SagaType:    SagaTypeOverlap,
		BranchCount: 3,
		TotalSagas:  3,
		Branches:    []string{"fix/a", "fix/b", "fix/c"},
		Confidence:  ConfidenceMedium,
	}

	rec := rt.ProduceRecommendation(pattern)

	if !strings.Contains(rec.Proposal, "dispatch deduplication") {
		t.Errorf("proposal should mention dispatch deduplication: %q", rec.Proposal)
	}
}

func TestFormatBoardEntry(t *testing.T) {
	t.Parallel()

	rec := BoardRecommendation{
		Pattern: RecurrencePattern{
			SagaType:   SagaTypeOverlap,
			Confidence: ConfidenceMedium,
		},
		Proposal: "Test proposal",
		Evidence: "3 events across 3 branches",
	}

	entry := FormatBoardEntry("REC-100", rec, "2026-03-10")

	if !strings.Contains(entry, "REC-100") {
		t.Error("expected recommendation ID in entry")
	}
	if !strings.Contains(entry, "Test proposal") {
		t.Error("expected proposal in entry")
	}
	if !strings.Contains(entry, "Medium") {
		t.Error("expected confidence level in entry")
	}
	if !strings.Contains(entry, "2026-03-10") {
		t.Error("expected date in entry")
	}
}

func TestRecurrenceTracker_Analyze_EmptyFindings(t *testing.T) {
	t.Parallel()
	rt := NewRecurrenceTracker(3)

	patterns := rt.Analyze(nil)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for nil findings, got %d", len(patterns))
	}

	patterns = rt.Analyze([]SagaFinding{})
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for empty findings, got %d", len(patterns))
	}
}

func TestFindingsLog_AppendThenRecurrenceAnalysis(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := dir + "/findings.jsonl"

	fl := NewFindingsLog(path, 200)
	rt := NewRecurrenceTracker(3)

	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	branches := []string{"fix/a", "fix/b", "fix/c", "fix/d"}

	for i, branch := range branches {
		finding := SagaFinding{
			Type:        FindingSagaDetected,
			Branch:      branch,
			PR:          500 + i,
			SagaType:    SagaTypeOverlap,
			WorkerCount: 2,
			Timestamp:   now.Add(time.Duration(i) * time.Hour),
			Repo:        "ThreeDoors",
		}
		if err := fl.Append(finding); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	entries, err := fl.ReadAll()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}

	patterns := rt.Analyze(entries)
	if len(patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(patterns))
	}
	if patterns[0].BranchCount != 4 {
		t.Errorf("branch count: got %d, want 4", patterns[0].BranchCount)
	}
	if patterns[0].Confidence != ConfidenceMedium {
		t.Errorf("confidence: got %q, want %q", patterns[0].Confidence, ConfidenceMedium)
	}
}
