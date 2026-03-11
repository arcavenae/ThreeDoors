package retrospector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func makeFindings(count int, opts ...func(*Finding)) []Finding {
	findings := make([]Finding, count)
	for i := range findings {
		findings[i] = Finding{
			PR:          100 + i,
			Story:       "43.1",
			ACMatch:     "full",
			CIFirstPass: true,
			Conflicts:   0,
			RebaseCount: 0,
			Timestamp:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			Repo:        "ThreeDoors",
		}
		for _, opt := range opts {
			opt(&findings[i])
		}
	}
	return findings
}

func TestDetectPatternsNone(t *testing.T) {
	t.Parallel()

	// All clean findings — no patterns
	findings := makeFindings(5)
	patterns := DetectPatterns(findings)
	if len(patterns) != 0 {
		t.Errorf("got %d patterns, want 0 for clean findings", len(patterns))
	}
}

func TestDetectCIFailurePattern(t *testing.T) {
	t.Parallel()

	findings := makeFindings(6, func(f *Finding) {
		// Make most fail CI
		if f.PR%2 == 0 {
			f.CIFirstPass = false
			f.CIFailures = []string{"lint"}
		}
	})

	patterns := DetectPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternCIFailure {
			found = true
			if p.DataPoints != 3 {
				t.Errorf("CI failure pattern data points = %d, want 3", p.DataPoints)
			}
			if !strings.Contains(p.Summary, "lint") {
				t.Errorf("summary should mention lint: %s", p.Summary)
			}
		}
	}
	if !found {
		t.Error("CI failure pattern not detected")
	}
}

func TestDetectMergeConflictPattern(t *testing.T) {
	t.Parallel()

	findings := makeFindings(4, func(f *Finding) {
		if f.PR <= 101 {
			f.Conflicts = 3
		}
	})

	patterns := DetectPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternMergeConflict {
			found = true
			if p.DataPoints != 2 {
				t.Errorf("merge conflict pattern data points = %d, want 2", p.DataPoints)
			}
		}
	}
	if !found {
		t.Error("merge conflict pattern not detected")
	}
}

func TestDetectACMismatchPattern(t *testing.T) {
	t.Parallel()

	findings := makeFindings(5, func(f *Finding) {
		if f.PR >= 103 {
			f.ACMatch = "partial"
		}
	})

	patterns := DetectPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternACMismatch {
			found = true
			if p.DataPoints != 2 {
				t.Errorf("AC mismatch pattern data points = %d, want 2", p.DataPoints)
			}
		}
	}
	if !found {
		t.Error("AC mismatch pattern not detected")
	}
}

func TestDetectExcessiveRebasePattern(t *testing.T) {
	t.Parallel()

	findings := makeFindings(4, func(f *Finding) {
		if f.PR <= 101 {
			f.RebaseCount = 5
		}
	})

	patterns := DetectPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternExcessiveRebase {
			found = true
			if p.DataPoints != 2 {
				t.Errorf("excessive rebase pattern data points = %d, want 2", p.DataPoints)
			}
		}
	}
	if !found {
		t.Error("excessive rebase pattern not detected")
	}
}

func TestDetectPatternsSortedByConfidence(t *testing.T) {
	t.Parallel()

	// Create findings that trigger multiple patterns with different data point counts
	findings := make([]Finding, 10)
	for i := range findings {
		findings[i] = Finding{
			PR:          100 + i,
			ACMatch:     "full",
			CIFirstPass: true,
			Timestamp:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			Repo:        "ThreeDoors",
		}
	}
	// 6 CI failures (High confidence)
	for i := 0; i < 6; i++ {
		findings[i].CIFirstPass = false
		findings[i].CIFailures = []string{"test"}
	}
	// 3 merge conflicts (Medium confidence)
	for i := 0; i < 3; i++ {
		findings[i].Conflicts = 2
	}

	patterns := DetectPatterns(findings)
	if len(patterns) < 2 {
		t.Fatalf("expected at least 2 patterns, got %d", len(patterns))
	}

	// First pattern should be highest confidence
	if patterns[0].Confidence != ConfidenceHigh {
		t.Errorf("first pattern confidence = %q, want High", patterns[0].Confidence)
	}
}

func TestDetectPatternsInsufficientData(t *testing.T) {
	t.Parallel()

	// Only 1 CI failure — below threshold of 2
	findings := makeFindings(5)
	findings[0].CIFirstPass = false
	findings[0].CIFailures = []string{"lint"}

	patterns := DetectPatterns(findings)
	for _, p := range patterns {
		if p.Type == PatternCIFailure {
			t.Error("should not detect CI failure pattern with only 1 data point")
		}
	}
}

func TestPipelineProcessBatch(t *testing.T) {
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

	pipeline := NewPipeline(bw, ks)

	// Create findings with CI failures
	findings := makeFindings(6, func(f *Finding) {
		f.CIFirstPass = false
		f.CIFailures = []string{"lint"}
	})

	recs, err := pipeline.ProcessBatch(findings)
	if err != nil {
		t.Fatalf("ProcessBatch: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected at least 1 recommendation")
	}

	// Verify recommendation was written to board
	data, err := os.ReadFile(boardPath)
	if err != nil {
		t.Fatalf("read board: %v", err)
	}
	if !strings.Contains(string(data), "P-001") {
		t.Error("P-001 not found in board after ProcessBatch")
	}
}

func TestPipelineRateLimit(t *testing.T) {
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

	pipeline := NewPipeline(bw, ks)

	// Create findings that trigger all 4 pattern types
	findings := make([]Finding, 10)
	for i := range findings {
		findings[i] = Finding{
			PR:          100 + i,
			ACMatch:     "partial",
			CIFirstPass: false,
			CIFailures:  []string{"lint"},
			Conflicts:   3,
			RebaseCount: 5,
			Timestamp:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			Repo:        "ThreeDoors",
		}
	}

	recs, err := pipeline.ProcessBatch(findings)
	if err != nil {
		t.Fatalf("ProcessBatch: %v", err)
	}
	if len(recs) > maxRecommendationsPerBatch {
		t.Errorf("got %d recs, want <= %d (rate limit)", len(recs), maxRecommendationsPerBatch)
	}
}

func TestPipelineSkipsWhenReadOnly(t *testing.T) {
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

	pipeline := NewPipeline(bw, ks)

	findings := makeFindings(6, func(f *Finding) {
		f.CIFirstPass = false
		f.CIFailures = []string{"lint"}
	})

	recs, err := pipeline.ProcessBatch(findings)
	if err != nil {
		t.Fatalf("ProcessBatch: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected 0 recs in read-only mode, got %d", len(recs))
	}
}

func TestPipelineNoPatterns(t *testing.T) {
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

	pipeline := NewPipeline(bw, ks)

	// All clean — no patterns
	findings := makeFindings(5)
	recs, err := pipeline.ProcessBatch(findings)
	if err != nil {
		t.Fatalf("ProcessBatch: %v", err)
	}
	if len(recs) != 0 {
		t.Errorf("expected 0 recs for clean findings, got %d", len(recs))
	}
}

func TestJoinWithCommaAnd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		items []string
		want  string
	}{
		{"empty", nil, ""},
		{"one", []string{"PR #1"}, "PR #1"},
		{"two", []string{"PR #1", "PR #2"}, "PR #1 and PR #2"},
		{"three", []string{"PR #1", "PR #2", "PR #3"}, "PR #1, PR #2, and PR #3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := joinWithCommaAnd(tt.items)
			if got != tt.want {
				t.Errorf("joinWithCommaAnd(%v) = %q, want %q", tt.items, got, tt.want)
			}
		})
	}
}
