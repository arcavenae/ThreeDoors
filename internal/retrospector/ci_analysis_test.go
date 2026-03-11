package retrospector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeCIFailures_ClassifiesByTaxonomy(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"lint", "gofumpt"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"WARNING: DATA RACE"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
		{PR: 103, CIFirstPass: false, CIFailures: []string{"build failed"}, Timestamp: time.Date(2026, 3, 10, 3, 0, 0, 0, time.UTC)},
		{PR: 104, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 4, 0, 0, 0, time.UTC)}, // clean
	}

	result := AnalyzeCIFailures(findings)

	if len(result.CategoryBreakdown) == 0 {
		t.Fatal("expected non-empty category breakdown")
	}
	if result.CategoryBreakdown[CategoryLint] != 3 {
		t.Errorf("lint count = %d, want 3", result.CategoryBreakdown[CategoryLint])
	}
	if result.CategoryBreakdown[CategoryRace] != 1 {
		t.Errorf("race count = %d, want 1", result.CategoryBreakdown[CategoryRace])
	}
	if result.CategoryBreakdown[CategoryBuild] != 1 {
		t.Errorf("build count = %d, want 1", result.CategoryBreakdown[CategoryBuild])
	}
	if result.TotalFailures != 4 {
		t.Errorf("total failures = %d, want 4", result.TotalFailures)
	}
	if result.TotalPRs != 5 {
		t.Errorf("total PRs = %d, want 5", result.TotalPRs)
	}
}

func TestAnalyzeCIFailures_RecurringPattern(t *testing.T) {
	t.Parallel()

	// Same failure category across 3+ PRs should be detected as recurring
	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"golangci-lint found issues"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
		{PR: 103, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 3, 0, 0, 0, time.UTC)},
	}

	result := AnalyzeCIFailures(findings)

	if len(result.RecurringPatterns) == 0 {
		t.Fatal("expected at least one recurring pattern")
	}

	found := false
	for _, rp := range result.RecurringPatterns {
		if rp.Category == CategoryLint {
			found = true
			if rp.Occurrences < 3 {
				t.Errorf("lint occurrences = %d, want >= 3", rp.Occurrences)
			}
			if rp.FixLayer != LayerCLAUDEMD {
				t.Errorf("lint fix layer = %q, want %q", rp.FixLayer, LayerCLAUDEMD)
			}
			if rp.FixProposal == "" {
				t.Error("expected non-empty fix proposal for lint pattern")
			}
			if len(rp.AffectedPRs) < 3 {
				t.Errorf("affected PRs = %d, want >= 3", len(rp.AffectedPRs))
			}
		}
	}
	if !found {
		t.Error("lint recurring pattern not found")
	}
}

func TestAnalyzeCIFailures_NoRecurringBelowThreshold(t *testing.T) {
	t.Parallel()

	// Only 2 PRs with same failure — below threshold of 3
	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
	}

	result := AnalyzeCIFailures(findings)

	for _, rp := range result.RecurringPatterns {
		if rp.Category == CategoryLint {
			t.Error("should not detect recurring pattern with only 2 PRs")
		}
	}
}

func TestAnalyzeCIFailures_UnclassifiedFlagged(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"some unknown error"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"another mystery"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: false, CIFailures: []string{"what is this"}, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
	}

	result := AnalyzeCIFailures(findings)

	if !result.HasUnclassified {
		t.Error("expected HasUnclassified to be true")
	}
	if result.UnclassifiedCount != 3 {
		t.Errorf("unclassified count = %d, want 3", result.UnclassifiedCount)
	}
}

func TestAnalyzeCIFailures_NoFailures(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 100, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
	}

	result := AnalyzeCIFailures(findings)

	if result.TotalFailures != 0 {
		t.Errorf("total failures = %d, want 0", result.TotalFailures)
	}
	if len(result.RecurringPatterns) != 0 {
		t.Errorf("recurring patterns = %d, want 0", len(result.RecurringPatterns))
	}
}

func TestCIAnalysisToRecommendations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	boardPath := filepath.Join(dir, "BOARD.md")
	ksPath := filepath.Join(dir, "killswitch.json")

	boardContent := `# Board

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|

## Decided
`
	if err := os.WriteFile(boardPath, []byte(boardContent), 0o600); err != nil {
		t.Fatalf("write board: %v", err)
	}

	bw := NewBoardWriter(boardPath)
	ks, err := NewKillSwitch(ksPath)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// Create findings with recurring lint failures across 4 PRs
	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"golangci-lint found issues"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: false, CIFailures: []string{"lint"}, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
		{PR: 103, CIFirstPass: false, CIFailures: []string{"gofumpt"}, Timestamp: time.Date(2026, 3, 10, 3, 0, 0, 0, time.UTC)},
		{PR: 104, CIFirstPass: true, Timestamp: time.Date(2026, 3, 10, 4, 0, 0, 0, time.UTC)},
	}

	analysis := AnalyzeCIFailures(findings)
	recs, err := FileCIAnalysisRecommendations(bw, ks, analysis)
	if err != nil {
		t.Fatalf("FileCIAnalysisRecommendations: %v", err)
	}

	if len(recs) == 0 {
		t.Fatal("expected at least 1 recommendation")
	}

	// Verify recommendation mentions the fix layer
	foundLayerRef := false
	for _, rec := range recs {
		if strings.Contains(rec.Text, string(LayerCLAUDEMD)) ||
			strings.Contains(rec.Text, "CLAUDE.md") {
			foundLayerRef = true
		}
	}
	if !foundLayerRef {
		t.Error("expected recommendation to reference the fix layer (CLAUDE.md)")
	}

	// Verify recommendation was written to BOARD.md
	data, err := os.ReadFile(boardPath)
	if err != nil {
		t.Fatalf("read board: %v", err)
	}
	if !strings.Contains(string(data), "P-001") {
		t.Error("P-001 not found in BOARD.md")
	}
}

func TestCIAnalysisToRecommendations_KillSwitchActive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	boardPath := filepath.Join(dir, "BOARD.md")
	ksPath := filepath.Join(dir, "killswitch.json")

	boardContent := `# Board

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|

## Decided
`
	if err := os.WriteFile(boardPath, []byte(boardContent), 0o600); err != nil {
		t.Fatalf("write board: %v", err)
	}

	bw := NewBoardWriter(boardPath)
	ks, err := NewKillSwitch(ksPath)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	// Trigger kill switch
	for i := 0; i < 3; i++ {
		if err := ks.RecordOutcome("P-001", OutcomeRejected); err != nil {
			t.Fatalf("RecordOutcome: %v", err)
		}
	}

	analysis := CIAnalysisResult{
		RecurringPatterns: []RecurringCIPattern{
			{Category: CategoryLint, Occurrences: 5, FixLayer: LayerCLAUDEMD, FixProposal: "test", AffectedPRs: []int{1, 2, 3, 4, 5}},
		},
	}

	recs, err := FileCIAnalysisRecommendations(bw, ks, analysis)
	if err != nil {
		t.Fatalf("FileCIAnalysisRecommendations: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected 0 recs in read-only mode, got %d", len(recs))
	}
}

func TestCIAnalysisToRecommendations_UnclassifiedRecommendation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	boardPath := filepath.Join(dir, "BOARD.md")
	ksPath := filepath.Join(dir, "killswitch.json")

	boardContent := `# Board

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|

## Decided
`
	if err := os.WriteFile(boardPath, []byte(boardContent), 0o600); err != nil {
		t.Fatalf("write board: %v", err)
	}

	bw := NewBoardWriter(boardPath)
	ks, err := NewKillSwitch(ksPath)
	if err != nil {
		t.Fatalf("NewKillSwitch: %v", err)
	}

	findings := []Finding{
		{PR: 100, CIFirstPass: false, CIFailures: []string{"mystery error"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 101, CIFirstPass: false, CIFailures: []string{"unknown thing"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
		{PR: 102, CIFirstPass: false, CIFailures: []string{"weird failure"}, Timestamp: time.Date(2026, 3, 10, 2, 0, 0, 0, time.UTC)},
	}

	analysis := AnalyzeCIFailures(findings)
	recs, err := FileCIAnalysisRecommendations(bw, ks, analysis)
	if err != nil {
		t.Fatalf("FileCIAnalysisRecommendations: %v", err)
	}

	// Should file a recommendation about unclassified failures
	foundUnclassified := false
	for _, rec := range recs {
		if strings.Contains(strings.ToLower(rec.Text), "unclassified") ||
			strings.Contains(strings.ToLower(rec.Text), "human review") {
			foundUnclassified = true
		}
	}
	if !foundUnclassified {
		t.Error("expected recommendation flagging unclassified failures for human review")
	}
}
