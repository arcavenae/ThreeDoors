package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// writeSessionsJSONL writes session metrics to a JSONL file for test use.
func writeSessionsJSONL(t *testing.T, dir string, sessions []core.SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	var buf bytes.Buffer
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal session: %v", err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write sessions.jsonl: %v", err)
	}
	return path
}

func TestCalculateStreak(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		sessions []core.SessionMetrics
		want     int
	}{
		{
			name:     "no sessions",
			sessions: nil,
			want:     0,
		},
		{
			name: "single day with completion",
			sessions: []core.SessionMetrics{
				{StartTime: now, TasksCompleted: 1},
			},
			want: 1,
		},
		{
			name: "three consecutive days",
			sessions: []core.SessionMetrics{
				{StartTime: now, TasksCompleted: 2},
				{StartTime: now.AddDate(0, 0, -1), TasksCompleted: 1},
				{StartTime: now.AddDate(0, 0, -2), TasksCompleted: 3},
			},
			want: 3,
		},
		{
			name: "streak broken by zero completion day",
			sessions: []core.SessionMetrics{
				{StartTime: now, TasksCompleted: 1},
				{StartTime: now.AddDate(0, 0, -1), TasksCompleted: 0},
				{StartTime: now.AddDate(0, 0, -2), TasksCompleted: 5},
			},
			want: 1,
		},
		{
			name: "no completions",
			sessions: []core.SessionMetrics{
				{StartTime: now, TasksCompleted: 0},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })

			if len(tt.sessions) > 0 {
				path := writeSessionsJSONL(t, dir, tt.sessions)
				if err := analyzer.LoadSessions(path); err != nil {
					t.Fatalf("LoadSessions: %v", err)
				}
			}

			got := calculateStreak(analyzer)
			if got != tt.want {
				t.Errorf("calculateStreak() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestStatsSummary_EmptyData(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	sessions := []core.SessionMetrics{}

	err := runStatsSummary(formatter, analyzer, sessions)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Dashboard") {
		t.Error("expected Dashboard header")
	}
	if !strings.Contains(output, "Total sessions:        0") {
		t.Errorf("expected 0 total sessions, got:\n%s", output)
	}
}

func TestStatsSummary_JSON_EmptyData(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	sessions := []core.SessionMetrics{}

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

	dataBytes, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var summary statsSummary
	if err := json.Unmarshal(dataBytes, &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if summary.TotalSessions != 0 {
		t.Errorf("total_sessions = %d, want 0", summary.TotalSessions)
	}
	if summary.CompletionRate != 0 {
		t.Errorf("completion_rate = %f, want 0", summary.CompletionRate)
	}
}

func TestRunStatsPatterns_SufficientData(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)
	sessions := make([]core.SessionMetrics, 6)
	for i := range sessions {
		sessions[i] = core.SessionMetrics{
			StartTime:      now.Add(-time.Duration(i) * time.Hour),
			TasksCompleted: i + 1,
		}
	}

	dir := t.TempDir()
	path := writeSessionsJSONL(t, dir, sessions)

	analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })
	if err := analyzer.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}

	tests := []struct {
		name   string
		isJSON bool
	}{
		{"human", false},
		{"json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, tt.isJSON)
			err := runStatsPatterns(formatter, analyzer, sessions, tt.isJSON)
			if err != nil {
				t.Fatalf("runStatsPatterns: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("expected non-empty output")
			}

			if tt.isJSON {
				var env JSONEnvelope
				if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if env.Command != "stats.patterns" {
					t.Errorf("command = %q, want %q", env.Command, "stats.patterns")
				}
			}
		})
	}
}

func TestStatsSummary_WithSessions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		{StartTime: now, TasksCompleted: 3},
		{StartTime: now.Add(-1 * time.Hour), TasksCompleted: 2},
		{StartTime: now.AddDate(0, 0, -1), TasksCompleted: 0},
	}

	dir := t.TempDir()
	path := writeSessionsJSONL(t, dir, sessions)

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })
	if err := analyzer.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}

	err := runStatsSummary(formatter, analyzer, sessions)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Total sessions:        3") {
		t.Errorf("expected 3 total sessions, got:\n%s", output)
	}
}

func TestCalculateCompletionRate_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sessions []core.SessionMetrics
		want     float64
	}{
		{"nil sessions", nil, 0},
		{"empty sessions", []core.SessionMetrics{}, 0},
		{"all zero", []core.SessionMetrics{
			{TasksCompleted: 0},
			{TasksCompleted: 0},
			{TasksCompleted: 0},
		}, 0},
		{"all completed", []core.SessionMetrics{
			{TasksCompleted: 1},
			{TasksCompleted: 5},
			{TasksCompleted: 2},
		}, 100},
		{"one of three", []core.SessionMetrics{
			{TasksCompleted: 1},
			{TasksCompleted: 0},
			{TasksCompleted: 0},
		}, 100.0 / 3.0},
		{"two of three", []core.SessionMetrics{
			{TasksCompleted: 1},
			{TasksCompleted: 1},
			{TasksCompleted: 0},
		}, 200.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := calculateCompletionRate(tt.sessions)
			diff := got - tt.want
			if diff < -0.01 || diff > 0.01 {
				t.Errorf("calculateCompletionRate() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestStatsDaily_WithData(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		{StartTime: now, TasksCompleted: 5},
		{StartTime: now.AddDate(0, 0, -1), TasksCompleted: 3},
	}

	dir := t.TempDir()
	path := writeSessionsJSONL(t, dir, sessions)

	analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })
	if err := analyzer.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}

	tests := []struct {
		name   string
		isJSON bool
	}{
		{"human", false},
		{"json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, tt.isJSON)
			err := runStatsDaily(formatter, analyzer, tt.isJSON)
			if err != nil {
				t.Fatalf("runStatsDaily: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("expected non-empty output")
			}

			if tt.isJSON {
				var env JSONEnvelope
				if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if env.Command != "stats.daily" {
					t.Errorf("command = %q, want %q", env.Command, "stats.daily")
				}
			} else {
				if !strings.Contains(output, "DATE") || !strings.Contains(output, "COMPLETED") {
					t.Errorf("missing table headers, got:\n%s", output)
				}
			}
		})
	}
}

func TestStatsWeekly_WithData(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		{StartTime: now, TasksCompleted: 5},
		{StartTime: now.AddDate(0, 0, -7), TasksCompleted: 3},
	}

	dir := t.TempDir()
	path := writeSessionsJSONL(t, dir, sessions)

	analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })
	if err := analyzer.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	err := runStatsWeekly(formatter, analyzer, false)
	if err != nil {
		t.Fatalf("runStatsWeekly: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"Week-over-Week", "This week:", "Last week:", "Change:"} {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q, got:\n%s", want, output)
		}
	}
}

func TestStatsSummary_JSON_WithSessions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 13, 15, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		{StartTime: now, TasksCompleted: 3},
		{StartTime: now.Add(-1 * time.Hour), TasksCompleted: 0},
	}

	dir := t.TempDir()
	path := writeSessionsJSONL(t, dir, sessions)

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(func() time.Time { return now })
	if err := analyzer.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}

	err := runStatsSummary(formatter, analyzer, sessions)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	dataBytes, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var summary statsSummary
	if err := json.Unmarshal(dataBytes, &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if summary.TotalSessions != 2 {
		t.Errorf("total_sessions = %d, want 2", summary.TotalSessions)
	}
	if summary.CompletionRate != 50 {
		t.Errorf("completion_rate = %f, want 50", summary.CompletionRate)
	}
}

func TestStatsCmd_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := newStatsCmd()

	tests := []struct {
		flag    string
		wantDef string
	}{
		{"daily", "false"},
		{"weekly", "false"},
		{"patterns", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			t.Parallel()
			f := cmd.Flags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("missing flag %q", tt.flag)
			}
			if f.DefValue != tt.wantDef {
				t.Errorf("flag %q default = %q, want %q", tt.flag, f.DefValue, tt.wantDef)
			}
		})
	}
}

func TestStatsSummary_OutputFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)

	err := runStatsSummary(formatter, analyzer, nil)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	output := buf.String()
	for _, label := range []string{
		"Dashboard",
		"Tasks completed today:",
		"Streak:",
		"Completion rate:",
		"Total sessions:",
	} {
		if !strings.Contains(output, label) {
			t.Errorf("missing label %q, got:\n%s", label, output)
		}
	}
}

func TestStatsSummary_StreakFormatting(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)

	err := runStatsSummary(formatter, analyzer, nil)
	if err != nil {
		t.Fatalf("runStatsSummary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "day(s)") {
		t.Errorf("streak should show 'day(s)', got:\n%s", output)
	}
}

func TestStatsPatterns_InsufficientData_JSON_Envelope(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)

	analyzer := core.NewPatternAnalyzerWithNow(time.Now)
	sessions := []core.SessionMetrics{{TasksCompleted: 1}}

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

	// With insufficient data, metadata should contain a message
	if env.Metadata != nil {
		meta, ok := env.Metadata.(map[string]interface{})
		if ok {
			if msg, ok := meta["message"].(string); ok {
				if !strings.Contains(msg, "not enough") {
					t.Errorf("message = %q, want containing 'not enough'", msg)
				}
			}
		}
	}
}

func TestNewStatsCmd_Help(t *testing.T) {
	t.Parallel()

	cmd := newStatsCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("stats --help: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"stats", "daily", "weekly", "patterns"} {
		if !strings.Contains(output, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

func TestVersionData_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
		commit  string
		build   string
	}{
		{"release", "1.0.0", "abc1234", "2026-03-13T00:00:00Z"},
		{"dev", "dev", "unknown", "unknown"},
		{"prerelease", "2.0.0-beta.1", "def5678", "2026-01-01T12:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := versionData{
				Version:   tt.version,
				Commit:    tt.commit,
				BuildDate: tt.build,
				GoVersion: "go1.25.4",
			}

			jsonBytes, err := json.Marshal(data)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var decoded versionData
			if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Version != tt.version {
				t.Errorf("Version = %q, want %q", decoded.Version, tt.version)
			}
			if decoded.Commit != tt.commit {
				t.Errorf("Commit = %q, want %q", decoded.Commit, tt.commit)
			}
		})
	}
}

func TestVersionHumanOutput_ChannelField(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate, oldChannel := Version, Commit, BuildDate, Channel
	Version, Commit, BuildDate, Channel = "1.0.0", "abc1234", "2026-01-15T10:00:00Z", "stable"
	t.Cleanup(func() {
		Version, Commit, BuildDate, Channel = oldVersion, oldCommit, oldBuildDate, oldChannel
	})

	var buf bytes.Buffer
	err := writeVersion(&buf, false)
	if err != nil {
		t.Fatalf("writeVersion: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("missing version in output: %s", output)
	}
}

func TestPlanCmd_Structure_Detailed(t *testing.T) {
	t.Parallel()

	cmd := newPlanCmd()
	if cmd.Use != "plan" {
		t.Errorf("Use = %q, want %q", cmd.Use, "plan")
	}
	if !strings.Contains(cmd.Long, "planning") {
		t.Errorf("Long description should mention planning, got: %q", cmd.Long)
	}

	// RunE should not return error
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Errorf("RunE returned error: %v", err)
	}
}

func TestPlanCmd_EmptyArgs(t *testing.T) {
	t.Parallel()

	// Plan command accepts NoArgs — verify it works with empty args
	cmd := newPlanCmd()
	if cmd.Args != nil {
		// If Args is set, verify it accepts empty
		if err := cmd.Args(cmd, nil); err != nil {
			t.Errorf("plan should accept nil args: %v", err)
		}
	}
}

func TestIsPlanCommand_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"plan as second arg", []string{"threedoors", "plan"}, true},
		{"plan with flags", []string{"threedoors", "plan", "--verbose"}, true},
		{"other command", []string{"threedoors", "doors"}, false},
		{"plan as third arg", []string{"threedoors", "config", "plan"}, false},
		{"empty", []string{}, false},
		{"nil", nil, false},
		{"only binary", []string{"threedoors"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := planCommandArgs
			t.Cleanup(func() { planCommandArgs = original })

			SetPlanCommandArgs(tt.args)
			got := IsPlanCommand()
			if got != tt.want {
				t.Errorf("IsPlanCommand() = %v, want %v (args=%v)", got, tt.want, tt.args)
			}
		})
	}
}

func TestConfigShowJSON_Envelope(t *testing.T) {
	t.Parallel()

	cfg := &core.ProviderConfig{
		SchemaVersion:      3,
		Provider:           "applenotes",
		NoteTitle:          "Work",
		Theme:              "scifi",
		DevDispatchEnabled: true,
	}

	m := configToMap(cfg)
	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	if err := formatter.WriteJSON("config show", m, nil); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
	}
	if env.Error != nil {
		t.Error("unexpected error in envelope")
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data not a map: %T", env.Data)
	}
	for _, key := range []string{"provider", "note_title", "theme", "schema_version", "dev_dispatch_enabled"} {
		if _, ok := data[key]; !ok {
			t.Errorf("JSON data missing key %q", key)
		}
	}
}

func TestHealthResultJSON_Serialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		result  healthResultJSON
		wantLen int
	}{
		{
			name: "healthy with all checks passing",
			result: healthResultJSON{
				Overall:    "ok",
				DurationMs: 42,
				Checks: []healthCheckJSON{
					{Name: "Database", Status: "ok", Message: "Connected"},
					{Name: "Config", Status: "ok", Message: "Valid"},
					{Name: "Sessions", Status: "ok", Message: "Writable"},
				},
			},
			wantLen: 3,
		},
		{
			name: "unhealthy with warnings",
			result: healthResultJSON{
				Overall:    "warn",
				DurationMs: 150,
				Checks: []healthCheckJSON{
					{Name: "Database", Status: "ok", Message: "Connected"},
					{Name: "Config", Status: "warn", Message: "Theme not set"},
				},
			},
			wantLen: 2,
		},
		{
			name: "unhealthy with errors",
			result: healthResultJSON{
				Overall:    "fail",
				DurationMs: 300,
				Checks: []healthCheckJSON{
					{Name: "Database", Status: "fail", Message: "Cannot connect"},
				},
			},
			wantLen: 1,
		},
		{
			name: "empty checks",
			result: healthResultJSON{
				Overall:    "ok",
				DurationMs: 0,
				Checks:     []healthCheckJSON{},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var decoded healthResultJSON
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			if decoded.Overall != tt.result.Overall {
				t.Errorf("Overall = %q, want %q", decoded.Overall, tt.result.Overall)
			}
			if decoded.DurationMs != tt.result.DurationMs {
				t.Errorf("DurationMs = %d, want %d", decoded.DurationMs, tt.result.DurationMs)
			}
			if len(decoded.Checks) != tt.wantLen {
				t.Errorf("got %d checks, want %d", len(decoded.Checks), tt.wantLen)
			}
		})
	}
}

func TestHealthResultJSON_InEnvelope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		overall string
		checks  int
	}{
		{"healthy", "ok", 3},
		{"warning", "warn", 2},
		{"failing", "fail", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checks := make([]healthCheckJSON, tt.checks)
			for i := range checks {
				checks[i] = healthCheckJSON{
					Name:    fmt.Sprintf("check_%d", i),
					Status:  "ok",
					Message: "passed",
				}
			}
			if tt.overall == "warn" && len(checks) > 0 {
				checks[len(checks)-1].Status = "warn"
			}
			if tt.overall == "fail" && len(checks) > 0 {
				checks[len(checks)-1].Status = "fail"
			}

			result := healthResultJSON{
				Overall:    tt.overall,
				DurationMs: 100,
				Checks:     checks,
			}

			var buf bytes.Buffer
			formatter := NewOutputFormatter(&buf, true)
			if err := formatter.WriteJSON("health", result, nil); err != nil {
				t.Fatalf("WriteJSON: %v", err)
			}

			var env JSONEnvelope
			if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if env.Command != "health" {
				t.Errorf("command = %q, want %q", env.Command, "health")
			}

			dataMap, ok := env.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data not a map: %T", env.Data)
			}
			if dataMap["overall"] != tt.overall {
				t.Errorf("overall = %v, want %q", dataMap["overall"], tt.overall)
			}
		})
	}
}
