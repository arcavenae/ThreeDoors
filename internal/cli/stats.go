package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/metrics"
	"github.com/spf13/cobra"
)

// newStatsCmd creates the "stats" command.
func newStatsCmd() *cobra.Command {
	var (
		daily    bool
		weekly   bool
		patterns bool
	)

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Display productivity statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStats(cmd, daily, weekly, patterns)
		},
	}

	cmd.Flags().BoolVar(&daily, "daily", false, "show daily completions for the last 7 days")
	cmd.Flags().BoolVar(&weekly, "weekly", false, "show week-over-week comparison")
	cmd.Flags().BoolVar(&patterns, "patterns", false, "show full pattern analysis")

	return cmd
}

// statsSummary holds the dashboard summary for JSON output.
type statsSummary struct {
	TodayCompleted int     `json:"today_completed"`
	Streak         int     `json:"streak_days"`
	CompletionRate float64 `json:"completion_rate"`
	TotalSessions  int     `json:"total_sessions"`
}

func runStats(cmd *cobra.Command, daily, weekly, patterns bool) error {
	isJSON := isJSONOutput(cmd)
	formatter := NewOutputFormatter(os.Stdout, isJSON)

	configDir, err := core.GetConfigDirPath()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("stats", ExitGeneralError, fmt.Sprintf("config dir: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: config dir: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	sessionsPath := filepath.Join(configDir, "sessions.jsonl")
	reader := metrics.NewReader(sessionsPath)
	sessions, err := reader.ReadAll()
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("stats", ExitGeneralError, fmt.Sprintf("read sessions: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: read sessions: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	analyzer := core.NewPatternAnalyzer()
	if err := analyzer.LoadSessions(sessionsPath); err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("stats", ExitGeneralError, fmt.Sprintf("load sessions: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: load sessions: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	if daily {
		return runStatsDaily(formatter, analyzer, isJSON)
	}
	if weekly {
		return runStatsWeekly(formatter, analyzer, isJSON)
	}
	if patterns {
		return runStatsPatterns(formatter, analyzer, sessions, isJSON)
	}

	return runStatsSummary(formatter, analyzer, sessions)
}

func runStatsSummary(formatter *OutputFormatter, analyzer *core.PatternAnalyzer, sessions []core.SessionMetrics) error {
	dailyMap := analyzer.GetDailyCompletions(1)
	todayCompleted := 0
	for _, v := range dailyMap {
		todayCompleted = v
	}

	streak := calculateStreak(analyzer)
	completionRate := calculateCompletionRate(sessions)

	summary := statsSummary{
		TodayCompleted: todayCompleted,
		Streak:         streak,
		CompletionRate: completionRate,
		TotalSessions:  len(sessions),
	}

	if formatter.IsJSON() {
		return formatter.WriteJSON("stats", summary, nil)
	}

	_ = formatter.Writef("Dashboard\n")
	_ = formatter.Writef("  Tasks completed today: %d\n", summary.TodayCompleted)
	_ = formatter.Writef("  Streak:                %d day(s)\n", summary.Streak)
	_ = formatter.Writef("  Completion rate:       %.1f%%\n", summary.CompletionRate)
	_ = formatter.Writef("  Total sessions:        %d\n", summary.TotalSessions)
	return nil
}

func runStatsDaily(formatter *OutputFormatter, analyzer *core.PatternAnalyzer, isJSON bool) error {
	dailyMap := analyzer.GetDailyCompletions(7)

	if isJSON {
		return formatter.WriteJSON("stats.daily", dailyMap, nil)
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(dailyMap))
	for k := range dailyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tw := formatter.TableWriter()
	_, _ = fmt.Fprintf(tw, "DATE\tCOMPLETED\n")
	for _, date := range keys {
		_, _ = fmt.Fprintf(tw, "%s\t%d\n", date, dailyMap[date])
	}
	_ = tw.Flush()
	return nil
}

func runStatsWeekly(formatter *OutputFormatter, analyzer *core.PatternAnalyzer, isJSON bool) error {
	wk := analyzer.GetWeekOverWeek()

	if isJSON {
		return formatter.WriteJSON("stats.weekly", wk, nil)
	}

	_ = formatter.Writef("Week-over-Week\n")
	_ = formatter.Writef("  This week:  %d tasks\n", wk.ThisWeekTotal)
	_ = formatter.Writef("  Last week:  %d tasks\n", wk.LastWeekTotal)
	_ = formatter.Writef("  Change:     %.1f%% (%s)\n", wk.PercentChange, wk.Direction)
	return nil
}

func runStatsPatterns(formatter *OutputFormatter, analyzer *core.PatternAnalyzer, sessions []core.SessionMetrics, isJSON bool) error {
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		if isJSON {
			_ = formatter.WriteJSONError("stats.patterns", ExitGeneralError, fmt.Sprintf("analyze: %v", err), "")
		} else {
			fmt.Fprintf(os.Stderr, "Error: analyze: %v\n", err)
		}
		os.Exit(ExitGeneralError)
	}

	if report == nil {
		msg := "not enough sessions for pattern analysis (need at least 5)"
		if isJSON {
			return formatter.WriteJSON("stats.patterns", nil, map[string]string{"message": msg})
		}
		return formatter.Writef("%s\n", msg)
	}

	if isJSON {
		return formatter.WriteJSON("stats.patterns", report, nil)
	}

	_ = formatter.Writef("Pattern Analysis (%d sessions)\n\n", report.SessionCount)

	_ = formatter.Writef("Door Position Bias:\n")
	_ = formatter.Writef("  Left: %d  Center: %d  Right: %d  (preferred: %s)\n",
		report.DoorPositionBias.LeftCount,
		report.DoorPositionBias.CenterCount,
		report.DoorPositionBias.RightCount,
		report.DoorPositionBias.PreferredPosition,
	)

	if len(report.TimeOfDayPatterns) > 0 {
		_ = formatter.Writef("\nTime of Day:\n")
		for _, p := range report.TimeOfDayPatterns {
			_ = formatter.Writef("  %s: %d sessions, avg %.1f tasks, avg %.1f min\n",
				p.Period, p.SessionCount, p.AvgTasksCompleted, p.AvgDuration)
		}
	}

	if len(report.MoodCorrelations) > 0 {
		_ = formatter.Writef("\nMood Correlations:\n")
		for _, mc := range report.MoodCorrelations {
			_ = formatter.Writef("  %s: %.1f avg tasks (%d sessions)\n",
				mc.Mood, mc.AvgTasksCompleted, mc.SessionCount)
		}
	}

	if len(report.AvoidanceList) > 0 {
		_ = formatter.Writef("\nAvoidance List:\n")
		for _, a := range report.AvoidanceList {
			_ = formatter.Writef("  %q: bypassed %d/%d times\n",
				a.TaskText, a.TimesBypassed, a.TimesShown)
		}
	}

	return nil
}

// calculateStreak returns consecutive days (including today) with at least one completion.
func calculateStreak(analyzer *core.PatternAnalyzer) int {
	// Check last 30 days for streak
	dailyMap := analyzer.GetDailyCompletions(30)

	keys := make([]string, 0, len(dailyMap))
	for k := range dailyMap {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	streak := 0
	for _, date := range keys {
		if dailyMap[date] > 0 {
			streak++
		} else {
			break
		}
	}
	return streak
}

// calculateCompletionRate returns the percentage of sessions with at least one completion.
func calculateCompletionRate(sessions []core.SessionMetrics) float64 {
	if len(sessions) == 0 {
		return 0
	}
	completed := 0
	for _, s := range sessions {
		if s.TasksCompleted > 0 {
			completed++
		}
	}
	return float64(completed) / float64(len(sessions)) * 100
}
