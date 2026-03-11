package retrospector

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// KillSwitchState tracks the recommendation outcome history and
// determines whether the retrospector should reduce to read-only mode.
type KillSwitchState struct {
	Outcomes            []OutcomeRecord `json:"outcomes"`
	ConsecutiveRejects  int             `json:"consecutive_rejects"`
	ReadOnly            bool            `json:"read_only"`
	ReadOnlySince       *time.Time      `json:"read_only_since,omitempty"`
	RecalibrationNeeded bool            `json:"recalibration_needed"`
}

// OutcomeRecord logs a single recommendation outcome.
type OutcomeRecord struct {
	RecommendationID string    `json:"recommendation_id"`
	Outcome          Outcome   `json:"outcome"`
	Timestamp        time.Time `json:"timestamp"`
}

// KillSwitch manages the kill switch state for the retrospector.
type KillSwitch struct {
	path  string
	state KillSwitchState
}

const consecutiveRejectThreshold = 3

// NewKillSwitch creates a KillSwitch that persists state to the given path.
// If the file exists, state is loaded from it.
func NewKillSwitch(path string) (*KillSwitch, error) {
	ks := &KillSwitch{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ks, nil
		}
		return nil, fmt.Errorf("read kill switch state %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &ks.state); err != nil {
		return nil, fmt.Errorf("parse kill switch state: %w", err)
	}
	return ks, nil
}

// RecordOutcome records the outcome of a recommendation and updates
// the consecutive rejection counter. If 3+ consecutive rejections
// occur, the kill switch triggers read-only mode.
func (ks *KillSwitch) RecordOutcome(recID string, outcome Outcome) error {
	now := time.Now().UTC()
	ks.state.Outcomes = append(ks.state.Outcomes, OutcomeRecord{
		RecommendationID: recID,
		Outcome:          outcome,
		Timestamp:        now,
	})

	switch outcome {
	case OutcomeRejected:
		ks.state.ConsecutiveRejects++
	default:
		ks.state.ConsecutiveRejects = 0
	}

	if ks.state.ConsecutiveRejects >= consecutiveRejectThreshold {
		ks.state.ReadOnly = true
		ks.state.ReadOnlySince = &now
		ks.state.RecalibrationNeeded = true
	}

	return ks.save()
}

// IsReadOnly returns whether the kill switch has been triggered.
func (ks *KillSwitch) IsReadOnly() bool {
	return ks.state.ReadOnly
}

// NeedsRecalibration returns whether recalibration has been requested
// but not yet acknowledged.
func (ks *KillSwitch) NeedsRecalibration() bool {
	return ks.state.RecalibrationNeeded
}

// ConsecutiveRejects returns the current consecutive rejection count.
func (ks *KillSwitch) ConsecutiveRejects() int {
	return ks.state.ConsecutiveRejects
}

// Reset clears the read-only flag and consecutive rejection counter,
// typically called after human recalibration.
func (ks *KillSwitch) Reset() error {
	ks.state.ReadOnly = false
	ks.state.ReadOnlySince = nil
	ks.state.ConsecutiveRejects = 0
	ks.state.RecalibrationNeeded = false
	return ks.save()
}

// State returns a copy of the current kill switch state.
func (ks *KillSwitch) State() KillSwitchState {
	return ks.state
}

func (ks *KillSwitch) save() error {
	data, err := json.MarshalIndent(ks.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal kill switch state: %w", err)
	}
	return os.WriteFile(ks.path, data, 0o600)
}
