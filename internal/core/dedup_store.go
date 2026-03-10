package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// DedupDecision records a user's decision about a potential duplicate pair.
type DedupDecision struct {
	TaskIDA   string `yaml:"task_id_a"`
	TaskIDB   string `yaml:"task_id_b"`
	Decision  string `yaml:"decision"` // "duplicate" or "distinct"
	DecidedAt string `yaml:"decided_at"`
}

// DedupStore persists duplicate detection decisions to a YAML file.
type DedupStore struct {
	mu        sync.RWMutex
	path      string
	decisions []DedupDecision
	index     map[string]string // "idA|idB" → decision
}

// NewDedupStore creates or loads a DedupStore from the given file path.
func NewDedupStore(path string) (*DedupStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create dedup store directory: %w", err)
	}

	store := &DedupStore{
		path:  path,
		index: make(map[string]string),
	}

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read dedup store: %w", err)
	}

	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &store.decisions); err != nil {
			return nil, fmt.Errorf("parse dedup store: %w", err)
		}
		for _, d := range store.decisions {
			store.index[pairKey(d.TaskIDA, d.TaskIDB)] = d.Decision
		}
	}

	return store, nil
}

// RecordDecision saves a duplicate/distinct decision for a task pair.
func (s *DedupStore) RecordDecision(taskIDA, taskIDB, decision string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := DedupDecision{
		TaskIDA:   taskIDA,
		TaskIDB:   taskIDB,
		Decision:  decision,
		DecidedAt: time.Now().UTC().Format(time.RFC3339),
	}

	s.decisions = append(s.decisions, d)
	s.index[pairKey(taskIDA, taskIDB)] = decision

	return s.save()
}

// HasDecision checks whether a decision exists for the given task pair (symmetric).
func (s *DedupStore) HasDecision(taskIDA, taskIDB string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.index[pairKey(taskIDA, taskIDB)]
	return ok
}

// GetDecision returns the decision for a task pair, if one exists.
func (s *DedupStore) GetDecision(taskIDA, taskIDB string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.index[pairKey(taskIDA, taskIDB)]
	return d, ok
}

// FilterUndecided returns only pairs that have no recorded decision.
func (s *DedupStore) FilterUndecided(pairs []DuplicatePair) []DuplicatePair {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var undecided []DuplicatePair
	for _, p := range pairs {
		if _, ok := s.index[pairKey(p.TaskA.ID, p.TaskB.ID)]; !ok {
			undecided = append(undecided, p)
		}
	}
	return undecided
}

// save writes the decisions to disk. Caller must hold s.mu.
func (s *DedupStore) save() error {
	data, err := yaml.Marshal(s.decisions)
	if err != nil {
		return fmt.Errorf("marshal dedup decisions: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write dedup store tmp: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("rename dedup store: %w", err)
	}
	return nil
}

// pairKey creates a canonical key for a task pair (always orders IDs lexicographically
// so lookups are symmetric).
func pairKey(idA, idB string) string {
	if idA > idB {
		idA, idB = idB, idA
	}
	return idA + "|" + idB
}
