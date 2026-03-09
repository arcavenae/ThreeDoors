package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Milestone represents a single milestone that can be shown to the user.
type Milestone struct {
	ID      string
	Message string
}

// MilestoneData is the JSON structure persisted to milestones.json.
type MilestoneData struct {
	Shown []string `json:"shown"`
}

// MilestoneChecker checks whether any milestone should be displayed
// and persists which milestones have already been shown.
type MilestoneChecker struct {
	configDir string
	data      MilestoneData
}

// milestoneDefinitions lists the four milestones in priority order (highest first).
// Per SOUL.md: observations, not achievements. No gamification language.
var milestoneDefinitions = []struct {
	id        string
	message   string
	checkFunc func(totalTasks, currentStreak, sessionCount int) bool
}{
	{
		id:      "first-session",
		message: "Welcome! Your journey starts here.",
		checkFunc: func(_, _, sessionCount int) bool {
			return sessionCount >= 1
		},
	},
	{
		id:      "100-tasks",
		message: "Triple digits! 100 tasks completed.",
		checkFunc: func(totalTasks, _, _ int) bool {
			return totalTasks >= 100
		},
	},
	{
		id:      "50-tasks",
		message: "50 tasks done — half a century of getting things done!",
		checkFunc: func(totalTasks, _, _ int) bool {
			return totalTasks >= 50
		},
	},
	{
		id:      "10-day-streak",
		message: "10 days in a row — double digits!",
		checkFunc: func(_, currentStreak, _ int) bool {
			return currentStreak >= 10
		},
	},
}

// NewMilestoneChecker creates a MilestoneChecker that persists to the given config directory.
func NewMilestoneChecker(configDir string) *MilestoneChecker {
	return &MilestoneChecker{
		configDir: configDir,
	}
}

// Load reads milestones.json from the config directory.
// Returns nil if the file doesn't exist (new user).
func (mc *MilestoneChecker) Load() error {
	path := mc.filePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			mc.data = MilestoneData{}
			return nil
		}
		return fmt.Errorf("read milestones file: %w", err)
	}

	var md MilestoneData
	if err := json.Unmarshal(data, &md); err != nil {
		return fmt.Errorf("parse milestones file: %w", err)
	}
	mc.data = md
	return nil
}

// CheckMilestones returns the highest-priority unshown milestone, or nil if none qualify.
func (mc *MilestoneChecker) CheckMilestones(totalTasks, currentStreak, sessionCount int) *Milestone {
	shown := make(map[string]bool, len(mc.data.Shown))
	for _, id := range mc.data.Shown {
		shown[id] = true
	}

	for _, def := range milestoneDefinitions {
		if shown[def.id] {
			continue
		}
		if def.checkFunc(totalTasks, currentStreak, sessionCount) {
			return &Milestone{
				ID:      def.id,
				Message: def.message,
			}
		}
	}
	return nil
}

// MarkShown records a milestone as shown and persists to milestones.json.
// Uses atomic write: write to .tmp, fsync, rename.
func (mc *MilestoneChecker) MarkShown(id string) error {
	// Don't duplicate
	for _, shown := range mc.data.Shown {
		if shown == id {
			return nil
		}
	}

	mc.data.Shown = append(mc.data.Shown, id)

	data, err := json.Marshal(mc.data)
	if err != nil {
		return fmt.Errorf("marshal milestones: %w", err)
	}

	if err := os.MkdirAll(mc.configDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmpPath := mc.filePath() + ".tmp"
	if err := writeFileAtomic(tmpPath, mc.filePath(), data); err != nil {
		return err
	}

	return nil
}

// IsShown returns whether a milestone has already been shown.
func (mc *MilestoneChecker) IsShown(id string) bool {
	for _, shown := range mc.data.Shown {
		if shown == id {
			return true
		}
	}
	return false
}

func (mc *MilestoneChecker) filePath() string {
	return filepath.Join(mc.configDir, "milestones.json")
}

// writeFileAtomic writes data to a file atomically: write .tmp, fsync, rename.
func writeFileAtomic(tmpPath, finalPath string, data []byte) error {
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create tmp file %s: %w", tmpPath, err)
	}

	_, writeErr := f.Write(data)
	if writeErr != nil {
		f.Close()          //nolint:errcheck // best-effort close on write error
		os.Remove(tmpPath) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("write tmp file %s: %w", tmpPath, writeErr)
	}

	if syncErr := f.Sync(); syncErr != nil {
		f.Close()          //nolint:errcheck // best-effort close on sync error
		os.Remove(tmpPath) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("sync tmp file %s: %w", tmpPath, syncErr)
	}

	if closeErr := f.Close(); closeErr != nil {
		os.Remove(tmpPath) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("close tmp file %s: %w", tmpPath, closeErr)
	}

	if renameErr := os.Rename(tmpPath, finalPath); renameErr != nil {
		return fmt.Errorf("rename %s to %s: %w", tmpPath, finalPath, renameErr)
	}

	return nil
}
