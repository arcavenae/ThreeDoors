package retrospector

import (
	"encoding/json"
	"fmt"
	"os"
)

// State tracks the last processed PR number to avoid reprocessing.
type State struct {
	LastProcessedPR int `json:"last_processed_pr"`
	path            string
}

// LoadState reads the retrospector state from disk.
// Returns a zero state (LastProcessedPR=0) if the file doesn't exist.
func LoadState(path string) (*State, error) {
	s := &State{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	return s, nil
}

// Save persists the state to disk using atomic write (tmp+rename).
func (s *State) Save() error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write state tmp: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename state: %w", err)
	}

	return nil
}
