package retrospector

import (
	"strings"
	"testing"
	"time"
)

func TestDetectHotFiles_NoHotFiles(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, FileList: []string{"a.go"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 2, FileList: []string{"b.go"}, Timestamp: time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC)},
	}

	ca := NewConflictAnalyzer(findings)
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) != 0 {
		t.Errorf("got %d hot files, want 0", len(hotFiles))
	}
}

func TestDetectHotFiles_ThreeConcurrentPRs(t *testing.T) {
	t.Parallel()

	// Three PRs all touching the same file, all concurrent (within 24h)
	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, FileList: []string{"internal/tasks/pool.go"}, EpicRef: "42", Timestamp: base},
		{PR: 2, FileList: []string{"internal/tasks/pool.go"}, EpicRef: "43", Timestamp: base.Add(2 * time.Hour)},
		{PR: 3, FileList: []string{"internal/tasks/pool.go"}, EpicRef: "42", Timestamp: base.Add(4 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) != 1 {
		t.Fatalf("got %d hot files, want 1", len(hotFiles))
	}
	if hotFiles[0].Path != "internal/tasks/pool.go" {
		t.Errorf("hot file path = %q, want %q", hotFiles[0].Path, "internal/tasks/pool.go")
	}
	if hotFiles[0].Count != 3 {
		t.Errorf("hot file count = %d, want 3", hotFiles[0].Count)
	}
}

func TestDetectHotFiles_NonConcurrentExcluded(t *testing.T) {
	t.Parallel()

	// Three PRs touching same file but NOT concurrent (>24h apart)
	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, FileList: []string{"a.go"}, Timestamp: base},
		{PR: 2, FileList: []string{"a.go"}, Timestamp: base.Add(48 * time.Hour)},
		{PR: 3, FileList: []string{"a.go"}, Timestamp: base.Add(96 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) != 0 {
		t.Errorf("got %d hot files, want 0 (PRs not concurrent)", len(hotFiles))
	}
}

func TestDetectHotFiles_WithLifecycleData(t *testing.T) {
	t.Parallel()

	// Three PRs with precise lifecycle overlap
	findings := []Finding{
		{
			PR: 1, FileList: []string{"shared.go"},
			CreatedAt: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			MergedAt:  time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
			Timestamp: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			PR: 2, FileList: []string{"shared.go"},
			CreatedAt: time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC),
			MergedAt:  time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
			Timestamp: time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			PR: 3, FileList: []string{"shared.go"},
			CreatedAt: time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC),
			MergedAt:  time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
			Timestamp: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
		},
	}

	ca := NewConflictAnalyzer(findings)
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) != 1 {
		t.Fatalf("got %d hot files, want 1", len(hotFiles))
	}
	if hotFiles[0].Count != 3 {
		t.Errorf("hot file count = %d, want 3", hotFiles[0].Count)
	}
}

func TestDetectHotFiles_SortedByCount(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, FileList: []string{"a.go", "b.go"}, Timestamp: base},
		{PR: 2, FileList: []string{"a.go", "b.go"}, Timestamp: base.Add(1 * time.Hour)},
		{PR: 3, FileList: []string{"a.go", "b.go"}, Timestamp: base.Add(2 * time.Hour)},
		{PR: 4, FileList: []string{"b.go"}, Timestamp: base.Add(3 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) < 2 {
		t.Fatalf("got %d hot files, want >= 2", len(hotFiles))
	}
	// b.go should be first (4 concurrent PRs) before a.go (3)
	if hotFiles[0].Path != "b.go" {
		t.Errorf("first hot file = %q, want b.go (highest count)", hotFiles[0].Path)
	}
}

func TestDetectEpicCollisions_NoCollisions(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, EpicRef: "42", FileList: []string{"a.go"}},
		{PR: 2, EpicRef: "43", FileList: []string{"b.go"}},
	}

	ca := NewConflictAnalyzer(findings)
	collisions := ca.DetectEpicCollisions()
	if len(collisions) != 0 {
		t.Errorf("got %d collisions, want 0", len(collisions))
	}
}

func TestDetectEpicCollisions_SharedFiles(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, EpicRef: "42", FileList: []string{"shared.go", "a.go"}},
		{PR: 2, EpicRef: "43", FileList: []string{"shared.go", "b.go"}},
		{PR: 3, EpicRef: "42", FileList: []string{"shared.go"}},
	}

	ca := NewConflictAnalyzer(findings)
	collisions := ca.DetectEpicCollisions()
	if len(collisions) != 1 {
		t.Fatalf("got %d collisions, want 1", len(collisions))
	}
	if collisions[0].EpicA != "42" || collisions[0].EpicB != "43" {
		t.Errorf("collision epics = %s/%s, want 42/43", collisions[0].EpicA, collisions[0].EpicB)
	}
	if len(collisions[0].SharedFiles) != 1 || collisions[0].SharedFiles[0] != "shared.go" {
		t.Errorf("shared files = %v, want [shared.go]", collisions[0].SharedFiles)
	}
}

func TestDetectEpicCollisions_NoEpicRefs(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, FileList: []string{"shared.go"}},
		{PR: 2, FileList: []string{"shared.go"}},
	}

	ca := NewConflictAnalyzer(findings)
	collisions := ca.DetectEpicCollisions()
	if len(collisions) != 0 {
		t.Errorf("got %d collisions, want 0 (no epic refs)", len(collisions))
	}
}

func TestCalculateDispatchSafety_AllSafe(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, FileList: []string{"a.go"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
		{PR: 2, FileList: []string{"b.go"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
	}

	ca := NewConflictAnalyzer(findings)
	score := ca.CalculateDispatchSafety()
	if score.Score != 1.0 {
		t.Errorf("safety score = %f, want 1.0", score.Score)
	}
	if score.Rating != "safe" {
		t.Errorf("rating = %q, want safe", score.Rating)
	}
}

func TestCalculateDispatchSafety_WithHotFiles(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, FileList: []string{"hot.go"}, Timestamp: base},
		{PR: 2, FileList: []string{"hot.go"}, Timestamp: base.Add(1 * time.Hour)},
		{PR: 3, FileList: []string{"hot.go"}, Timestamp: base.Add(2 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	score := ca.CalculateDispatchSafety()
	if score.Score >= 1.0 {
		t.Errorf("safety score = %f, want < 1.0 with hot files", score.Score)
	}
	if len(score.HotFiles) != 1 {
		t.Errorf("got %d hot files in score, want 1", len(score.HotFiles))
	}
}

func TestAnalyzeRebaseChurn_NoChurn(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, RebaseCount: 1},
		{PR: 2, RebaseCount: 2},
	}

	ca := NewConflictAnalyzer(findings)
	entries := ca.AnalyzeRebaseChurn()
	if len(entries) != 0 {
		t.Errorf("got %d churn entries, want 0", len(entries))
	}
}

func TestAnalyzeRebaseChurn_ConcurrentPRsCause(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{
			PR: 1, RebaseCount: 4, FileList: []string{"shared.go"},
			Timestamp: base,
		},
		{
			PR: 2, RebaseCount: 0, FileList: []string{"shared.go"},
			Timestamp: base.Add(1 * time.Hour),
		},
	}

	ca := NewConflictAnalyzer(findings)
	entries := ca.AnalyzeRebaseChurn()
	if len(entries) != 1 {
		t.Fatalf("got %d churn entries, want 1", len(entries))
	}
	if entries[0].RootCause != RootCauseConcurrentPRs {
		t.Errorf("root cause = %q, want %q", entries[0].RootCause, RootCauseConcurrentPRs)
	}
	if len(entries[0].ConcurrentPRs) != 1 || entries[0].ConcurrentPRs[0] != 2 {
		t.Errorf("concurrent PRs = %v, want [2]", entries[0].ConcurrentPRs)
	}
}

func TestAnalyzeRebaseChurn_LongLivedCause(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{
			PR:          1,
			RebaseCount: 5,
			FileList:    []string{"unique.go"},
			CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			MergedAt:    time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			Timestamp:   time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	ca := NewConflictAnalyzer(findings)
	entries := ca.AnalyzeRebaseChurn()
	if len(entries) != 1 {
		t.Fatalf("got %d churn entries, want 1", len(entries))
	}
	if entries[0].RootCause != RootCauseLongLived {
		t.Errorf("root cause = %q, want %q", entries[0].RootCause, RootCauseLongLived)
	}
}

func TestAnalyzeRebaseChurn_SortedByRebaseCount(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, RebaseCount: 3, Timestamp: base},
		{PR: 2, RebaseCount: 7, Timestamp: base.Add(1 * time.Hour)},
		{PR: 3, RebaseCount: 5, Timestamp: base.Add(2 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	entries := ca.AnalyzeRebaseChurn()
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].PR != 2 {
		t.Errorf("first entry PR = %d, want 2 (highest rebase count)", entries[0].PR)
	}
	if entries[1].PR != 3 {
		t.Errorf("second entry PR = %d, want 3", entries[1].PR)
	}
}

func TestAnalyze_FullReport(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, EpicRef: "42", FileList: []string{"shared.go", "a.go"}, RebaseCount: 4, Timestamp: base},
		{PR: 2, EpicRef: "43", FileList: []string{"shared.go", "b.go"}, RebaseCount: 0, Timestamp: base.Add(1 * time.Hour)},
		{PR: 3, EpicRef: "42", FileList: []string{"shared.go"}, RebaseCount: 3, Timestamp: base.Add(2 * time.Hour)},
	}

	ca := NewConflictAnalyzer(findings)
	report := ca.Analyze()

	// Should have hot files (shared.go in 3 concurrent PRs)
	if len(report.HotFiles) == 0 {
		t.Error("expected hot files in report")
	}

	// Should have epic collisions (42 vs 43 sharing shared.go)
	if len(report.EpicCollisions) == 0 {
		t.Error("expected epic collisions in report")
	}

	// Should have rebase churn entries
	if len(report.RebaseChurn) == 0 {
		t.Error("expected rebase churn entries in report")
	}

	// Should have recommendations
	if len(report.Recommendations) == 0 {
		t.Error("expected recommendations in report")
	}

	// Verify recommendations mention specific files and PRs
	foundFileRec := false
	for _, rec := range report.Recommendations {
		if strings.Contains(rec, "shared.go") {
			foundFileRec = true
		}
	}
	if !foundFileRec {
		t.Error("expected recommendation mentioning shared.go")
	}
}

func TestEpicRefFromStory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		storyRef string
		want     string
	}{
		{"valid ref", "51.3", "51"},
		{"single digit", "1.2", "1"},
		{"empty", "", ""},
		{"invalid format", "abc", ""},
		{"no dot", "51", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EpicRefFromStory(tt.storyRef)
			if got != tt.want {
				t.Errorf("EpicRefFromStory(%q) = %q, want %q", tt.storyRef, got, tt.want)
			}
		})
	}
}

func TestFormatPRList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		prs  []int
		want string
	}{
		{"empty", nil, ""},
		{"single", []int{42}, "#42"},
		{"multiple", []int{1, 2, 3}, "#1, #2, #3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatPRList(tt.prs)
			if got != tt.want {
				t.Errorf("formatPRList(%v) = %q, want %q", tt.prs, got, tt.want)
			}
		})
	}
}

func TestDetectConflictAnalysisPatterns_HotFilePattern(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, FileList: []string{"hot.go"}, Timestamp: base},
		{PR: 2, FileList: []string{"hot.go"}, Timestamp: base.Add(1 * time.Hour)},
		{PR: 3, FileList: []string{"hot.go"}, Timestamp: base.Add(2 * time.Hour)},
	}

	patterns := detectConflictAnalysisPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternHotFile {
			found = true
			if p.DataPoints < 3 {
				t.Errorf("hot file pattern data points = %d, want >= 3", p.DataPoints)
			}
		}
	}
	if !found {
		t.Error("hot file pattern not detected")
	}
}

func TestDetectConflictAnalysisPatterns_EpicCollisionPattern(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, EpicRef: "42", FileList: []string{"shared.go"}},
		{PR: 2, EpicRef: "43", FileList: []string{"shared.go"}},
	}

	patterns := detectConflictAnalysisPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternEpicCollision {
			found = true
		}
	}
	if !found {
		t.Error("epic collision pattern not detected")
	}
}

func TestDetectConflictAnalysisPatterns_RebaseChurnPattern(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	findings := []Finding{
		{PR: 1, RebaseCount: 4, FileList: []string{"a.go"}, Timestamp: base},
		{PR: 2, RebaseCount: 5, FileList: []string{"b.go"}, Timestamp: base.Add(1 * time.Hour)},
	}

	patterns := detectConflictAnalysisPatterns(findings)
	found := false
	for _, p := range patterns {
		if p.Type == PatternRebaseChurn {
			found = true
			if p.DataPoints != 2 {
				t.Errorf("rebase churn data points = %d, want 2", p.DataPoints)
			}
		}
	}
	if !found {
		t.Error("rebase churn pattern not detected")
	}
}

func TestDetectConflictAnalysisPatterns_NoPatterns(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 1, FileList: []string{"a.go"}, Timestamp: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
	}

	patterns := detectConflictAnalysisPatterns(findings)
	if len(patterns) != 0 {
		t.Errorf("got %d patterns, want 0", len(patterns))
	}
}
