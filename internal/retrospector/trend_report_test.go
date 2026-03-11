package retrospector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestComputeWeeklyMetrics(t *testing.T) {
	t.Parallel()

	t.Run("empty findings", func(t *testing.T) {
		t.Parallel()
		m := ComputeWeeklyMetrics(nil, nil)
		if m.PRsMerged != 0 {
			t.Errorf("PRsMerged = %d, want 0", m.PRsMerged)
		}
	})

	t.Run("basic metrics", func(t *testing.T) {
		t.Parallel()
		findings := []Finding{
			{PR: 1, CIFirstPass: true, RebaseCount: 1, ACMatch: ACMatchFull, Conflicts: 0},
			{PR: 2, CIFirstPass: false, RebaseCount: 3, ACMatch: ACMatchPartial, Conflicts: 1},
			{PR: 3, CIFirstPass: true, RebaseCount: 0, ACMatch: ACMatchFull, Conflicts: 0},
			{PR: 4, CIFirstPass: true, RebaseCount: 2, ACMatch: ACMatchFull, Conflicts: 0},
		}
		outcomes := []OutcomeRecord{
			{Outcome: OutcomeAccepted},
			{Outcome: OutcomeRejected},
			{Outcome: OutcomeAccepted},
		}

		m := ComputeWeeklyMetrics(findings, outcomes)

		if m.PRsMerged != 4 {
			t.Errorf("PRsMerged = %d, want 4", m.PRsMerged)
		}
		// 3/4 CI pass
		if m.CIFirstPassRate != 0.75 {
			t.Errorf("CIFirstPassRate = %f, want 0.75", m.CIFirstPassRate)
		}
		// (1+3+0+2)/4 = 1.5
		if m.AvgRebaseCount != 1.5 {
			t.Errorf("AvgRebaseCount = %f, want 1.5", m.AvgRebaseCount)
		}
		// 1/4 conflict PRs
		if m.ConflictRate != 0.25 {
			t.Errorf("ConflictRate = %f, want 0.25", m.ConflictRate)
		}
		// 3/4 AC full match
		if m.ACFullMatchRate != 0.75 {
			t.Errorf("ACFullMatchRate = %f, want 0.75", m.ACFullMatchRate)
		}
		// 2/3 acceptance rate
		wantAcceptance := 2.0 / 3.0
		if m.AcceptanceRate < wantAcceptance-0.01 || m.AcceptanceRate > wantAcceptance+0.01 {
			t.Errorf("AcceptanceRate = %f, want ~%f", m.AcceptanceRate, wantAcceptance)
		}
	})

	t.Run("no outcomes", func(t *testing.T) {
		t.Parallel()
		findings := []Finding{
			{PR: 1, CIFirstPass: true},
		}
		m := ComputeWeeklyMetrics(findings, nil)
		if m.AcceptanceRate != 0 {
			t.Errorf("AcceptanceRate should be 0 with no outcomes, got %f", m.AcceptanceRate)
		}
	})
}

func TestFilterFindingsByWeek(t *testing.T) {
	t.Parallel()

	// 2026-W10 starts on Monday March 2
	findings := []Finding{
		{PR: 1, Timestamp: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)},  // Week 10
		{PR: 2, Timestamp: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC)},  // Week 10
		{PR: 3, Timestamp: time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)},  // Week 11
		{PR: 4, Timestamp: time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)}, // Week 11
	}

	week10 := FilterFindingsByWeek(findings, 2026, 10)
	if len(week10) != 2 {
		t.Fatalf("Week 10 should have 2 findings, got %d", len(week10))
	}
	if week10[0].PR != 1 || week10[1].PR != 2 {
		t.Errorf("Week 10 PRs = %d,%d, want 1,2", week10[0].PR, week10[1].PR)
	}

	week11 := FilterFindingsByWeek(findings, 2026, 11)
	if len(week11) != 2 {
		t.Fatalf("Week 11 should have 2 findings, got %d", len(week11))
	}
}

func TestFilterOutcomesByWeek(t *testing.T) {
	t.Parallel()

	outcomes := []OutcomeRecord{
		{Outcome: OutcomeAccepted, Timestamp: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)},
		{Outcome: OutcomeRejected, Timestamp: time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)},
	}

	week10 := FilterOutcomesByWeek(outcomes, 2026, 10)
	if len(week10) != 1 {
		t.Fatalf("Week 10 should have 1 outcome, got %d", len(week10))
	}
}

func TestDetectRegressions(t *testing.T) {
	t.Parallel()

	t.Run("no history", func(t *testing.T) {
		t.Parallel()
		current := WeeklyMetrics{CIFirstPassRate: 0.5}
		regs := DetectRegressions(current, nil)
		if len(regs) != 0 {
			t.Errorf("expected no regressions with no history, got %d", len(regs))
		}
	})

	t.Run("CI rate regression", func(t *testing.T) {
		t.Parallel()
		current := WeeklyMetrics{
			CIFirstPassRate: 0.5,
			AvgRebaseCount:  1.0,
			ConflictRate:    0.1,
			ACFullMatchRate: 0.8,
			AcceptanceRate:  0.7,
		}
		history := []WeeklyMetrics{
			{CIFirstPassRate: 0.8, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
			{CIFirstPassRate: 0.9, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
			{CIFirstPassRate: 0.7, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
			{CIFirstPassRate: 0.8, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
		}

		regs := DetectRegressions(current, history)
		found := false
		for _, r := range regs {
			if r.Name == "CI First-Pass Rate" {
				found = true
				if r.CurrentValue != 0.5 {
					t.Errorf("CurrentValue = %f, want 0.5", r.CurrentValue)
				}
				if r.FourWeekAvg != 0.8 {
					t.Errorf("FourWeekAvg = %f, want 0.8", r.FourWeekAvg)
				}
				if len(r.PossibleCauses) == 0 {
					t.Error("should have possible causes")
				}
			}
		}
		if !found {
			t.Error("expected CI First-Pass Rate regression to be detected")
		}
	})

	t.Run("rebase count regression (higher is worse)", func(t *testing.T) {
		t.Parallel()
		current := WeeklyMetrics{
			CIFirstPassRate: 0.8,
			AvgRebaseCount:  5.0,
			ConflictRate:    0.1,
			ACFullMatchRate: 0.8,
			AcceptanceRate:  0.7,
		}
		history := []WeeklyMetrics{
			{CIFirstPassRate: 0.8, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
			{CIFirstPassRate: 0.8, AvgRebaseCount: 2.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
		}

		regs := DetectRegressions(current, history)
		found := false
		for _, r := range regs {
			if r.Name == "Average Rebase Count" {
				found = true
				if r.CurrentValue != 5.0 {
					t.Errorf("CurrentValue = %f, want 5.0", r.CurrentValue)
				}
			}
		}
		if !found {
			t.Error("expected Average Rebase Count regression")
		}
	})

	t.Run("no regression when metrics are good", func(t *testing.T) {
		t.Parallel()
		current := WeeklyMetrics{
			CIFirstPassRate: 0.9,
			AvgRebaseCount:  0.5,
			ConflictRate:    0.0,
			ACFullMatchRate: 1.0,
			AcceptanceRate:  1.0,
		}
		history := []WeeklyMetrics{
			{CIFirstPassRate: 0.8, AvgRebaseCount: 1.0, ConflictRate: 0.1, ACFullMatchRate: 0.8, AcceptanceRate: 0.7},
		}

		regs := DetectRegressions(current, history)
		if len(regs) != 0 {
			t.Errorf("expected no regressions, got %d: %+v", len(regs), regs)
		}
	})
}

func TestFormatTrendReport(t *testing.T) {
	t.Parallel()

	report := TrendReport{
		Week:   "2026-W10",
		Period: "2026-03-02 to 2026-03-08",
		Metrics: WeeklyMetrics{
			PRsMerged:       10,
			CIFirstPassRate: 0.8,
			AvgRebaseCount:  1.5,
			SagaCount:       2,
			ConflictRate:    0.1,
			ACFullMatchRate: 0.9,
			AcceptanceRate:  0.75,
		},
		Regressions: []MetricRegression{
			{
				Name:           "CI First-Pass Rate",
				CurrentValue:   0.8,
				FourWeekAvg:    0.9,
				Delta:          -0.1,
				PossibleCauses: []string{"flaky tests"},
			},
		},
		GeneratedAt: time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC),
	}

	md := FormatTrendReport(report)

	checks := []string{
		"2026-W10",
		"PRs Merged | 10",
		"CI First-Pass Rate | 80.0%",
		"Average Rebase Count | 1.5",
		"Saga Count | 2",
		"Recommendation Acceptance Rate | 75.0%",
		"Conflict Rate | 10.0%",
		"AC Full Match Rate | 90.0%",
		"Metric Regressions",
		"CI First-Pass Rate decreased",
		"4-Week Average",
		"flaky tests",
		"SLAES retrospector",
	}

	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("report should contain %q", check)
		}
	}
}

func TestFormatTrendReport_NoRegressions(t *testing.T) {
	t.Parallel()

	report := TrendReport{
		Week:   "2026-W10",
		Period: "2026-03-02 to 2026-03-08",
		Metrics: WeeklyMetrics{
			PRsMerged: 5,
		},
		GeneratedAt: time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC),
	}

	md := FormatTrendReport(report)
	if !strings.Contains(md, "No regressions detected") {
		t.Error("report should contain no-regression message")
	}
}

func TestSaveReport(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "operations")
	fl := NewFindingsLog(tmpDir)
	tr := NewTrendReporter(fl, reportsDir)

	report := TrendReport{
		Week:   "2026-W10",
		Period: "2026-03-02 to 2026-03-08",
		Metrics: WeeklyMetrics{
			PRsMerged: 5,
		},
		GeneratedAt: time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC),
	}

	path, err := tr.SaveReport(report)
	if err != nil {
		t.Fatalf("SaveReport() error: %v", err)
	}

	expectedPath := filepath.Join(reportsDir, "slaes-report-2026-W10.md")
	if path != expectedPath {
		t.Errorf("path = %q, want %q", path, expectedPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved report: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "2026-W10") {
		t.Error("saved report should contain week identifier")
	}
	if !strings.Contains(content, "PRs Merged | 5") {
		t.Error("saved report should contain metrics")
	}
}

func TestGenerateReport(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	fl := NewFindingsLog(tmpDir)
	tr := NewTrendReporter(fl, filepath.Join(tmpDir, "operations"))

	// Create findings spanning multiple weeks
	findings := []Finding{
		// Week 10 (2026-03-02 to 2026-03-08)
		{PR: 1, CIFirstPass: true, RebaseCount: 1, ACMatch: ACMatchFull, Timestamp: time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC)},
		{PR: 2, CIFirstPass: false, RebaseCount: 2, ACMatch: ACMatchPartial, Timestamp: time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC)},
		// Week 9 (2026-02-23 to 2026-03-01) — history
		{PR: 3, CIFirstPass: true, RebaseCount: 0, ACMatch: ACMatchFull, Timestamp: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)},
		{PR: 4, CIFirstPass: true, RebaseCount: 1, ACMatch: ACMatchFull, Timestamp: time.Date(2026, 2, 25, 10, 0, 0, 0, time.UTC)},
	}

	report, err := tr.GenerateReport(findings, nil, 2026, 10)
	if err != nil {
		t.Fatalf("GenerateReport() error: %v", err)
	}

	if report.Week != "2026-W10" {
		t.Errorf("Week = %q, want %q", report.Week, "2026-W10")
	}
	if report.Metrics.PRsMerged != 2 {
		t.Errorf("PRsMerged = %d, want 2", report.Metrics.PRsMerged)
	}
	if report.Metrics.CIFirstPassRate != 0.5 {
		t.Errorf("CIFirstPassRate = %f, want 0.5", report.Metrics.CIFirstPassRate)
	}
}

func TestSubtractWeeks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		year     int
		week     int
		n        int
		wantYear int
		wantWeek int
	}{
		{"same year", 2026, 10, 1, 2026, 9},
		{"same year 4 weeks", 2026, 10, 4, 2026, 6},
		{"cross year boundary", 2026, 1, 1, 2025, 52},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotYear, gotWeek := subtractWeeks(tt.year, tt.week, tt.n)
			if gotYear != tt.wantYear || gotWeek != tt.wantWeek {
				t.Errorf("subtractWeeks(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.year, tt.week, tt.n, gotYear, gotWeek, tt.wantYear, tt.wantWeek)
			}
		})
	}
}

func TestISOWeekStart(t *testing.T) {
	t.Parallel()

	// 2026-W10 should start on Monday March 2, 2026
	got := isoWeekStart(2026, 10)
	want := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("isoWeekStart(2026, 10) = %v, want %v", got, want)
	}
}
