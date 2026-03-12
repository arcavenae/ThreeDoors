package retrospector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TrendMetric represents a single metric value for a time period.
type TrendMetric struct {
	Name  string
	Value float64
	Unit  string
}

// MetricRegression flags a metric that dropped below its 4-week average.
type MetricRegression struct {
	Name           string
	CurrentValue   float64
	FourWeekAvg    float64
	Delta          float64
	PossibleCauses []string
}

// WeeklyMetrics holds the computed metrics for one week.
type WeeklyMetrics struct {
	PRsMerged             int
	CIFirstPassRate       float64
	AvgRebaseCount        float64
	SagaCount             int
	DocConsistencyScore   float64
	AcceptanceRate        float64
	TotalFindings         int
	ConflictRate          float64
	ACFullMatchRate       float64
	HighConfidenceRecRate float64
}

// TrendReport is a weekly continuous improvement report.
type TrendReport struct {
	Week        string // YYYY-WNN format
	Period      string // human-readable date range
	Metrics     WeeklyMetrics
	Regressions []MetricRegression
	GeneratedAt time.Time
}

// TrendReporter generates periodic trend reports from JSONL findings data.
type TrendReporter struct {
	findingsLog *FindingsLog
	reportsDir  string
}

// NewTrendReporter creates a reporter that reads findings from the given log
// and writes reports to the given directory.
func NewTrendReporter(fl *FindingsLog, reportsDir string) *TrendReporter {
	return &TrendReporter{
		findingsLog: fl,
		reportsDir:  reportsDir,
	}
}

// ComputeWeeklyMetrics calculates metrics for findings within a time window.
func ComputeWeeklyMetrics(findings []Finding, outcomes []OutcomeRecord) WeeklyMetrics {
	if len(findings) == 0 {
		return WeeklyMetrics{}
	}

	m := WeeklyMetrics{
		PRsMerged:     len(findings),
		TotalFindings: len(findings),
	}

	// CI first-pass rate
	ciPass := 0
	for _, f := range findings {
		if f.CIFirstPass {
			ciPass++
		}
	}
	m.CIFirstPassRate = float64(ciPass) / float64(len(findings))

	// Average rebase count
	totalRebase := 0
	for _, f := range findings {
		totalRebase += f.RebaseCount
	}
	m.AvgRebaseCount = float64(totalRebase) / float64(len(findings))

	// Conflict rate
	conflictPRs := 0
	for _, f := range findings {
		if f.Conflicts > 0 {
			conflictPRs++
		}
	}
	m.ConflictRate = float64(conflictPRs) / float64(len(findings))

	// AC full match rate
	fullMatch := 0
	for _, f := range findings {
		if f.ACMatch == ACMatchFull {
			fullMatch++
		}
	}
	m.ACFullMatchRate = float64(fullMatch) / float64(len(findings))

	// Recommendation acceptance rate
	if len(outcomes) > 0 {
		accepted := 0
		for _, o := range outcomes {
			if o.Outcome == OutcomeAccepted {
				accepted++
			}
		}
		m.AcceptanceRate = float64(accepted) / float64(len(outcomes))
	}

	return m
}

// FilterFindingsByWeek returns findings whose timestamp falls within the
// ISO week identified by year and week number.
func FilterFindingsByWeek(findings []Finding, year int, week int) []Finding {
	var filtered []Finding
	for _, f := range findings {
		fy, fw := f.Timestamp.ISOWeek()
		if fy == year && fw == week {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterOutcomesByWeek returns outcomes whose timestamp falls within the
// ISO week identified by year and week number.
func FilterOutcomesByWeek(outcomes []OutcomeRecord, year int, week int) []OutcomeRecord {
	var filtered []OutcomeRecord
	for _, o := range outcomes {
		oy, ow := o.Timestamp.ISOWeek()
		if oy == year && ow == week {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

// DetectRegressions compares current week metrics against a 4-week average
// and returns any metrics that regressed.
func DetectRegressions(current WeeklyMetrics, history []WeeklyMetrics) []MetricRegression {
	if len(history) == 0 {
		return nil
	}

	var regressions []MetricRegression

	type metricCheck struct {
		name       string
		current    float64
		histValues func(m WeeklyMetrics) float64
		higherGood bool // true if higher is better
		causes     []string
	}

	checks := []metricCheck{
		{
			name:       "CI First-Pass Rate",
			current:    current.CIFirstPassRate,
			histValues: func(m WeeklyMetrics) float64 { return m.CIFirstPassRate },
			higherGood: true,
			causes:     []string{"increased complexity of changes", "inadequate local testing", "flaky tests"},
		},
		{
			name:       "Average Rebase Count",
			current:    current.AvgRebaseCount,
			histValues: func(m WeeklyMetrics) float64 { return m.AvgRebaseCount },
			higherGood: false,
			causes:     []string{"high concurrent PR count", "long-lived branches", "hot file contention"},
		},
		{
			name:       "Conflict Rate",
			current:    current.ConflictRate,
			histValues: func(m WeeklyMetrics) float64 { return m.ConflictRate },
			higherGood: false,
			causes:     []string{"concurrent epics touching same files", "insufficient PR sequencing"},
		},
		{
			name:       "AC Full Match Rate",
			current:    current.ACFullMatchRate,
			histValues: func(m WeeklyMetrics) float64 { return m.ACFullMatchRate },
			higherGood: true,
			causes:     []string{"story specs lack specificity", "scope creep in PRs", "missing task breakdown"},
		},
		{
			name:       "Recommendation Acceptance Rate",
			current:    current.AcceptanceRate,
			histValues: func(m WeeklyMetrics) float64 { return m.AcceptanceRate },
			higherGood: true,
			causes:     []string{"recommendations not actionable", "insufficient evidence", "misaligned priorities"},
		},
	}

	for _, check := range checks {
		avg := computeAverage(history, check.histValues)
		if avg == 0 {
			continue
		}

		regressed := false
		if check.higherGood {
			regressed = check.current < avg
		} else {
			regressed = check.current > avg
		}

		if regressed {
			delta := check.current - avg
			regressions = append(regressions, MetricRegression{
				Name:           check.name,
				CurrentValue:   check.current,
				FourWeekAvg:    avg,
				Delta:          delta,
				PossibleCauses: check.causes,
			})
		}
	}

	return regressions
}

func computeAverage(history []WeeklyMetrics, extract func(WeeklyMetrics) float64) float64 {
	if len(history) == 0 {
		return 0
	}
	sum := 0.0
	for _, m := range history {
		sum += extract(m)
	}
	return sum / float64(len(history))
}

// GenerateReport creates a trend report for the given ISO week, comparing
// against up to 4 prior weeks of history.
func (tr *TrendReporter) GenerateReport(findings []Finding, outcomes []OutcomeRecord, year int, week int) (TrendReport, error) {
	weekFindings := FilterFindingsByWeek(findings, year, week)
	weekOutcomes := FilterOutcomesByWeek(outcomes, year, week)

	current := ComputeWeeklyMetrics(weekFindings, weekOutcomes)

	// Compute 4-week history
	var history []WeeklyMetrics
	for i := 1; i <= 4; i++ {
		hy, hw := subtractWeeks(year, week, i)
		hFindings := FilterFindingsByWeek(findings, hy, hw)
		hOutcomes := FilterOutcomesByWeek(outcomes, hy, hw)
		hMetrics := ComputeWeeklyMetrics(hFindings, hOutcomes)
		if hMetrics.TotalFindings > 0 {
			history = append(history, hMetrics)
		}
	}

	regressions := DetectRegressions(current, history)

	weekStart := isoWeekStart(year, week)
	weekEnd := weekStart.AddDate(0, 0, 6)

	report := TrendReport{
		Week:        fmt.Sprintf("%d-W%02d", year, week),
		Period:      fmt.Sprintf("%s to %s", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02")),
		Metrics:     current,
		Regressions: regressions,
		GeneratedAt: time.Now().UTC(),
	}

	return report, nil
}

// SaveReport writes the trend report as a markdown file.
func (tr *TrendReporter) SaveReport(report TrendReport) (string, error) {
	if err := os.MkdirAll(tr.reportsDir, 0o700); err != nil {
		return "", fmt.Errorf("create reports dir: %w", err)
	}

	filename := fmt.Sprintf("slaes-report-%s.md", report.Week)
	path := filepath.Join(tr.reportsDir, filename)

	content := FormatTrendReport(report)

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("write report %s: %w", path, err)
	}

	return path, nil
}

// FormatTrendReport renders a TrendReport as markdown.
func FormatTrendReport(report TrendReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# SLAES Continuous Improvement Report — %s\n\n", report.Week)
	fmt.Fprintf(&b, "**Period:** %s\n", report.Period)
	fmt.Fprintf(&b, "**Generated:** %s\n\n", report.GeneratedAt.Format(time.RFC3339))

	fmt.Fprintf(&b, "## Key Metrics\n\n")
	fmt.Fprintf(&b, "| Metric | Value |\n")
	fmt.Fprintf(&b, "|--------|-------|\n")
	fmt.Fprintf(&b, "| PRs Merged | %d |\n", report.Metrics.PRsMerged)
	fmt.Fprintf(&b, "| CI First-Pass Rate | %.1f%% |\n", report.Metrics.CIFirstPassRate*100)
	fmt.Fprintf(&b, "| Average Rebase Count | %.1f |\n", report.Metrics.AvgRebaseCount)
	fmt.Fprintf(&b, "| Saga Count | %d |\n", report.Metrics.SagaCount)
	fmt.Fprintf(&b, "| Doc Consistency Score | %.1f%% |\n", report.Metrics.DocConsistencyScore*100)
	fmt.Fprintf(&b, "| Recommendation Acceptance Rate | %.1f%% |\n", report.Metrics.AcceptanceRate*100)
	fmt.Fprintf(&b, "| Conflict Rate | %.1f%% |\n", report.Metrics.ConflictRate*100)
	fmt.Fprintf(&b, "| AC Full Match Rate | %.1f%% |\n", report.Metrics.ACFullMatchRate*100)

	if len(report.Regressions) > 0 {
		fmt.Fprintf(&b, "\n## Metric Regressions\n\n")
		for _, r := range report.Regressions {
			direction := "decreased"
			if r.Delta > 0 {
				direction = "increased"
			}
			fmt.Fprintf(&b, "### %s %s\n\n", r.Name, direction)
			fmt.Fprintf(&b, "- **Current:** %.2f\n", r.CurrentValue)
			fmt.Fprintf(&b, "- **4-Week Average:** %.2f\n", r.FourWeekAvg)
			fmt.Fprintf(&b, "- **Delta:** %+.2f\n", r.Delta)
			fmt.Fprintf(&b, "- **Possible Causes:**\n")
			for _, c := range r.PossibleCauses {
				fmt.Fprintf(&b, "  - %s\n", c)
			}
			fmt.Fprintf(&b, "\n")
		}
	} else {
		fmt.Fprintf(&b, "\n## Metric Regressions\n\nNo regressions detected — all metrics are at or above their 4-week average.\n")
	}

	fmt.Fprintf(&b, "\n---\n\n*Report generated by SLAES retrospector.*\n")

	return b.String()
}

// subtractWeeks returns the ISO year and week N weeks before the given week.
func subtractWeeks(year, week, n int) (int, int) {
	t := isoWeekStart(year, week)
	t = t.AddDate(0, 0, -7*n)
	y, w := t.ISOWeek()
	return y, w
}

// isoWeekStart returns the Monday of the given ISO week.
func isoWeekStart(year, week int) time.Time {
	// January 4 is always in week 1 of the ISO year.
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	// Find the Monday of week 1
	daysSinceMonday := (int(jan4.Weekday()) + 6) % 7
	week1Monday := jan4.AddDate(0, 0, -daysSinceMonday)
	// Add weeks
	return week1Monday.AddDate(0, 0, (week-1)*7)
}
