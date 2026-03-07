package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestNewStatsCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := newStatsCmd()
	if cmd.Use != "stats" {
		t.Errorf("Use = %q, want %q", cmd.Use, "stats")
	}

	for _, name := range []string{"daily", "weekly", "patterns"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestCalculateCompletionRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sessions []core.SessionMetrics
		want     float64
	}{
		{"empty", nil, 0},
		{"all completed", []core.SessionMetrics{
			{TasksCompleted: 3},
			{TasksCompleted: 1},
		}, 100},
		{"half completed", []core.SessionMetrics{
			{TasksCompleted: 2},
			{TasksCompleted: 0},
		}, 50},
		{"none completed", []core.SessionMetrics{
			{TasksCompleted: 0},
			{TasksCompleted: 0},
		}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := calculateCompletionRate(tt.sessions)
			if got != tt.want {
				t.Errorf("calculateCompletionRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatsSummary_JSON(t *testing.T) {
	t.Parallel()

	summary := statsSummary{
		TodayCompleted: 5,
		Streak:         3,
		CompletionRate: 75.0,
		TotalSessions:  10,
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded["today_completed"] != float64(5) {
		t.Errorf("today_completed = %v, want 5", decoded["today_completed"])
	}
	if decoded["streak_days"] != float64(3) {
		t.Errorf("streak_days = %v, want 3", decoded["streak_days"])
	}
	if decoded["completion_rate"] != float64(75) {
		t.Errorf("completion_rate = %v, want 75", decoded["completion_rate"])
	}
	if decoded["total_sessions"] != float64(10) {
		t.Errorf("total_sessions = %v, want 10", decoded["total_sessions"])
	}
}

func TestRunStatsSummary_HumanReadable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	sessions := []core.SessionMetrics{
		{StartTime: time.Now().UTC(), TasksCompleted: 3},
		{StartTime: time.Now().UTC(), TasksCompleted: 0},
	}

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	// Load sessions directly isn't possible without a file, so test the summary renderer
	err := runStatsSummary(formatter, analyzer, sessions)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRunStatsSummary_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	// Override package-level jsonOutput for JSON mode detection inside formatter
	sessions := []core.SessionMetrics{
		{StartTime: time.Now().UTC(), TasksCompleted: 2},
	}

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	err := runStatsSummary(formatter, analyzer, sessions)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "stats" {
		t.Errorf("command = %q, want %q", env.Command, "stats")
	}
}

func TestRunStatsDaily_HumanReadable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	err := runStatsDaily(formatter, analyzer, false)
	if err != nil {
		t.Fatalf("runStatsDaily: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRunStatsWeekly_HumanReadable(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	err := runStatsWeekly(formatter, analyzer, false)
	if err != nil {
		t.Fatalf("runStatsWeekly: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRunStatsPatterns_InsufficientData(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	sessions := []core.SessionMetrics{
		{TasksCompleted: 1},
		{TasksCompleted: 2},
	}

	err := runStatsPatterns(formatter, analyzer, sessions, false)
	if err != nil {
		t.Fatalf("runStatsPatterns: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected message about insufficient data")
	}
}

func TestRootCmd_IncludesMoodAndStats(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	names := make(map[string]bool)
	for _, cmd := range root.Commands() {
		names[cmd.Name()] = true
	}

	for _, want := range []string{"mood", "stats"} {
		if !names[want] {
			t.Errorf("root command missing %q subcommand", want)
		}
	}
}
