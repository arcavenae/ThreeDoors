package cli

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestCalculateCompletionRate_SingleSession(t *testing.T) {
	t.Parallel()

	got := calculateCompletionRate([]core.SessionMetrics{{TasksCompleted: 1}})
	if got != 100 {
		t.Errorf("calculateCompletionRate() = %f, want 100", got)
	}
}

func TestRunStatsDaily_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	err := runStatsDaily(formatter, analyzer, true)
	if err != nil {
		t.Fatalf("runStatsDaily: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "stats.daily" {
		t.Errorf("command = %q, want %q", env.Command, "stats.daily")
	}
}

func TestRunStatsWeekly_JSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	err := runStatsWeekly(formatter, analyzer, true)
	if err != nil {
		t.Fatalf("runStatsWeekly: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "stats.weekly" {
		t.Errorf("command = %q, want %q", env.Command, "stats.weekly")
	}
}

func TestRunStatsPatterns_JSON_InsufficientData(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	sessions := []core.SessionMetrics{
		{TasksCompleted: 1},
	}

	err := runStatsPatterns(formatter, analyzer, sessions, true)
	if err != nil {
		t.Fatalf("runStatsPatterns: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Command != "stats.patterns" {
		t.Errorf("command = %q, want %q", env.Command, "stats.patterns")
	}
}
