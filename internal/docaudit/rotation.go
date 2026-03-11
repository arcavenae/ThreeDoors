package docaudit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const rotationInterval = 4 * time.Hour

// LoadRotationState reads the rotation state from a JSON file.
// Returns a zero-value state if the file does not exist.
func LoadRotationState(path string) (RotationState, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RotationState{}, nil
	}
	if err != nil {
		return RotationState{}, fmt.Errorf("read rotation state %s: %w", path, err)
	}

	var state RotationState
	if err := json.Unmarshal(data, &state); err != nil {
		return RotationState{}, fmt.Errorf("unmarshal rotation state: %w", err)
	}
	return state, nil
}

// SaveRotationState writes the rotation state to a JSON file atomically.
func SaveRotationState(path string, state RotationState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rotation state: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write rotation state tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename rotation state: %w", err)
	}
	return nil
}

// NextMode determines which deep analysis mode should run next, based on the
// current rotation state. Returns the next mode and whether it's time to run.
func NextMode(state RotationState, now time.Time) (RotationMode, bool) {
	modes := AllModes()

	// If no mode has ever run, start with the first one.
	if state.LastMode == "" {
		return modes[0], true
	}

	// Check if enough time has passed since the last run.
	if now.Sub(state.LastRun) < rotationInterval {
		return "", false
	}

	// Find the current mode's index and advance to the next.
	currentIdx := -1
	for i, m := range modes {
		if m == state.LastMode {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(modes)
	return modes[nextIdx], true
}

// RecordRun updates the rotation state after a mode has been executed.
func RecordRun(state *RotationState, mode RotationMode, now time.Time, clean bool) {
	state.LastMode = mode
	state.LastRun = now
	state.CycleCount++

	run := ModeRun{
		Mode:      mode,
		Timestamp: now,
		Clean:     clean,
	}
	state.ModeHistory = append(state.ModeHistory, run)

	// Keep only the last 20 history entries.
	if len(state.ModeHistory) > 20 {
		state.ModeHistory = state.ModeHistory[len(state.ModeHistory)-20:]
	}
}
