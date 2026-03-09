package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLogPlanningSession(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sessionsPath := filepath.Join(dir, "sessions.jsonl")

	event := PlanningSessionEvent{
		Timestamp:        time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC),
		DurationSeconds:  120.5,
		TasksReviewed:    10,
		TasksContinued:   5,
		TasksDeferred:    3,
		TasksDropped:     2,
		FocusTaskCount:   3,
		EnergyLevel:      EnergyHigh,
		EnergyOverridden: false,
	}

	err := LogPlanningSession(sessionsPath, event)
	if err != nil {
		t.Fatalf("LogPlanningSession() error = %v", err)
	}

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	lines := strings.TrimSpace(string(data))
	if strings.Count(lines, "\n") != 0 {
		t.Fatalf("expected exactly one line, got %d", strings.Count(lines, "\n")+1)
	}

	var decoded PlanningSessionEvent
	if err := json.Unmarshal([]byte(lines), &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Type != "planning_session" {
		t.Errorf("Type = %q, want %q", decoded.Type, "planning_session")
	}
	if decoded.DurationSeconds != 120.5 {
		t.Errorf("DurationSeconds = %v, want 120.5", decoded.DurationSeconds)
	}
	if decoded.TasksReviewed != 10 {
		t.Errorf("TasksReviewed = %d, want 10", decoded.TasksReviewed)
	}
	if decoded.TasksContinued != 5 {
		t.Errorf("TasksContinued = %d, want 5", decoded.TasksContinued)
	}
	if decoded.TasksDeferred != 3 {
		t.Errorf("TasksDeferred = %d, want 3", decoded.TasksDeferred)
	}
	if decoded.TasksDropped != 2 {
		t.Errorf("TasksDropped = %d, want 2", decoded.TasksDropped)
	}
	if decoded.FocusTaskCount != 3 {
		t.Errorf("FocusTaskCount = %d, want 3", decoded.FocusTaskCount)
	}
	if decoded.EnergyLevel != EnergyHigh {
		t.Errorf("EnergyLevel = %q, want %q", decoded.EnergyLevel, EnergyHigh)
	}
	if decoded.EnergyOverridden {
		t.Errorf("EnergyOverridden = true, want false")
	}
}

func TestLogPlanningSession_appends(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sessionsPath := filepath.Join(dir, "sessions.jsonl")

	event1 := PlanningSessionEvent{
		Timestamp:       time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC),
		DurationSeconds: 60,
		TasksReviewed:   5,
		EnergyLevel:     EnergyHigh,
	}
	event2 := PlanningSessionEvent{
		Timestamp:       time.Date(2026, 3, 9, 9, 0, 0, 0, time.UTC),
		DurationSeconds: 90,
		TasksReviewed:   8,
		EnergyLevel:     EnergyMedium,
	}

	if err := LogPlanningSession(sessionsPath, event1); err != nil {
		t.Fatalf("LogPlanningSession(1) error = %v", err)
	}
	if err := LogPlanningSession(sessionsPath, event2); err != nil {
		t.Fatalf("LogPlanningSession(2) error = %v", err)
	}

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var d1, d2 PlanningSessionEvent
	if err := json.Unmarshal([]byte(lines[0]), &d1); err != nil {
		t.Fatalf("unmarshal line 1: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &d2); err != nil {
		t.Fatalf("unmarshal line 2: %v", err)
	}

	if d1.Type != "planning_session" || d2.Type != "planning_session" {
		t.Errorf("both events should have type planning_session")
	}
	if d1.TasksReviewed != 5 || d2.TasksReviewed != 8 {
		t.Errorf("event data mismatch")
	}
}

func TestLogPlanningSession_setsType(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sessionsPath := filepath.Join(dir, "sessions.jsonl")

	// Even if caller sets a different type, it should be overridden
	event := PlanningSessionEvent{
		Type:        "wrong_type",
		Timestamp:   time.Date(2026, 3, 8, 9, 0, 0, 0, time.UTC),
		EnergyLevel: EnergyLow,
	}

	if err := LogPlanningSession(sessionsPath, event); err != nil {
		t.Fatalf("LogPlanningSession() error = %v", err)
	}

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var decoded PlanningSessionEvent
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Type != "planning_session" {
		t.Errorf("Type = %q, want %q (should override caller value)", decoded.Type, "planning_session")
	}
}
