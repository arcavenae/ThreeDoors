package core

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// PlanningSessionEvent captures metrics from a daily planning session.
type PlanningSessionEvent struct {
	Type             string      `json:"type"`
	Timestamp        time.Time   `json:"timestamp"`
	DurationSeconds  float64     `json:"duration_seconds"`
	TasksReviewed    int         `json:"tasks_reviewed"`
	TasksContinued   int         `json:"tasks_continued"`
	TasksDeferred    int         `json:"tasks_deferred"`
	TasksDropped     int         `json:"tasks_dropped"`
	FocusTaskCount   int         `json:"focus_task_count"`
	EnergyLevel      EnergyLevel `json:"energy_level"`
	EnergyOverridden bool        `json:"energy_overridden"`
}

// LogPlanningSession appends a planning session event to the JSONL session log.
// The event type is always set to "planning_session".
func LogPlanningSession(sessionsPath string, event PlanningSessionEvent) error {
	event.Type = "planning_session"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal planning session event: %w", err)
	}

	f, err := os.OpenFile(sessionsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open sessions file: %w", err)
	}
	defer f.Close() //nolint:errcheck // best-effort close on append file

	_, err = f.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("write planning session event: %w", err)
	}

	return nil
}
