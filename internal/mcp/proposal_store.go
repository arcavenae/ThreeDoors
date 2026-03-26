package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

// ErrProposalNotFound is returned when a proposal ID doesn't exist.
var ErrProposalNotFound = fmt.Errorf("proposal not found")

// ErrPerTaskCapReached is returned when a task has too many pending proposals.
var ErrPerTaskCapReached = fmt.Errorf("per-task pending proposal cap reached")

// ErrDuplicateProposal is returned when a proposal duplicates an existing one.
var ErrDuplicateProposal = fmt.Errorf("duplicate proposal")

// ProposalStore manages proposals backed by an append-only JSONL file.
type ProposalStore struct {
	mu        sync.RWMutex
	path      string
	proposals map[string]*Proposal
	pool      *core.TaskPool
	nowFunc   func() time.Time
}

// NewProposalStore creates a store backed by the given JSONL file path.
func NewProposalStore(path string, pool *core.TaskPool) (*ProposalStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create proposal store directory: %w", err)
	}

	s := &ProposalStore{
		path:      path,
		proposals: make(map[string]*Proposal),
		pool:      pool,
		nowFunc:   func() time.Time { return time.Now().UTC() },
	}

	if err := s.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("load proposals: %w", err)
	}

	return s, nil
}

// Create adds a new proposal after validation, dedup, and cap checks.
func (s *ProposalStore) Create(proposal *Proposal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check per-task cap.
	pendingCount := 0
	for _, p := range s.proposals {
		if p.TaskID == proposal.TaskID && p.Status == ProposalPending {
			pendingCount++
		}
	}
	if pendingCount >= MaxPendingPerTask {
		return fmt.Errorf("task %s: %w (%d/%d)", proposal.TaskID, ErrPerTaskCapReached, pendingCount, MaxPendingPerTask)
	}

	// Dedup: check against existing pending proposals for same task.
	for _, p := range s.proposals {
		if p.TaskID != proposal.TaskID {
			continue
		}
		if p.Status == ProposalPending {
			sim := core.TextSimilarity(payloadText(p.Payload), payloadText(proposal.Payload))
			if sim >= DuplicateSimilarityThreshold {
				return fmt.Errorf("similar to pending proposal %s (%.0f%% match): %w", p.ID, sim*100, ErrDuplicateProposal)
			}
		}
		// Check recently rejected (within 7 days).
		if p.Status == ProposalRejected && p.ReviewedAt != nil {
			if s.nowFunc().Sub(*p.ReviewedAt) < DefaultExpirationDuration {
				sim := core.TextSimilarity(payloadText(p.Payload), payloadText(proposal.Payload))
				if sim >= DuplicateSimilarityThreshold {
					return fmt.Errorf("similar to recently rejected proposal %s: %w", p.ID, ErrDuplicateProposal)
				}
			}
		}
	}

	// Dedup: check against existing tasks in pool.
	if s.pool != nil {
		tasks := s.pool.GetAllTasks()
		proposalText := payloadText(proposal.Payload)
		for _, t := range tasks {
			sim := core.TextSimilarity(t.Text, proposalText)
			if sim >= DuplicateSimilarityThreshold {
				return fmt.Errorf("similar to existing task %s (%.0f%% match): %w", t.ID, sim*100, ErrDuplicateProposal)
			}
		}
		// Check completed tasks.
		for _, t := range tasks {
			if t.Status == core.StatusComplete && t.CompletedAt != nil {
				if s.nowFunc().Sub(*t.CompletedAt) < DefaultExpirationDuration {
					sim := core.TextSimilarity(t.Text, proposalText)
					if sim >= DuplicateSimilarityThreshold {
						return fmt.Errorf("similar to recently completed task %s: %w", t.ID, ErrDuplicateProposal)
					}
				}
			}
		}
	}

	s.proposals[proposal.ID] = proposal
	return s.appendToDisk(proposal)
}

// Get retrieves a proposal by ID.
func (s *ProposalStore) Get(id string) (*Proposal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.proposals[id]
	if !ok {
		return nil, fmt.Errorf("proposal %s: %w", id, ErrProposalNotFound)
	}
	return p, nil
}

// List returns proposals matching the given filter.
func (s *ProposalStore) List(filter ProposalFilter) []*Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Proposal
	for _, p := range s.proposals {
		if filter.TaskID != "" && p.TaskID != filter.TaskID {
			continue
		}
		if filter.Status != "" && p.Status != filter.Status {
			continue
		}
		if filter.Type != "" && p.Type != filter.Type {
			continue
		}
		if filter.Source != "" && p.Source != filter.Source {
			continue
		}
		result = append(result, p)
	}
	return result
}

// UpdateStatus changes a proposal's status and sets the ReviewedAt timestamp.
// Performs optimistic concurrency check when transitioning from pending.
func (s *ProposalStore) UpdateStatus(id string, status ProposalStatus, reviewedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.proposals[id]
	if !ok {
		return fmt.Errorf("proposal %s: %w", id, ErrProposalNotFound)
	}

	if p.IsTerminal() {
		return fmt.Errorf("proposal %s is already in terminal state %s", id, p.Status)
	}

	// Optimistic concurrency: check if the task has been modified since the proposal was created.
	if status == ProposalApproved && s.pool != nil {
		task := s.pool.GetTask(p.TaskID)
		if task != nil && !task.UpdatedAt.Equal(p.BaseVersion) {
			p.Status = ProposalStale
			p.ReviewedAt = &reviewedAt
			return s.appendToDisk(p)
		}
	}

	p.Status = status
	p.ReviewedAt = &reviewedAt
	return s.appendToDisk(p)
}

// ExpireSweep marks proposals older than their ExpiresAt as expired.
func (s *ProposalStore) ExpireSweep() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.nowFunc()
	expired := 0
	for _, p := range s.proposals {
		if p.Status == ProposalPending && now.After(p.ExpiresAt) {
			p.Status = ProposalExpired
			_ = s.appendToDisk(p)
			expired++
		}
	}
	return expired
}

func (s *ProposalStore) loadFromDisk() error {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open proposals file: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var p Proposal
		if err := json.Unmarshal(scanner.Bytes(), &p); err != nil {
			continue
		}
		s.proposals[p.ID] = &p
	}
	return scanner.Err()
}

func (s *ProposalStore) appendToDisk(p *Proposal) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal proposal %s: %w", p.ID, err)
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open proposals file for append: %w", err)
	}
	defer func() { _ = f.Close() }()

	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write proposal %s: %w", p.ID, err)
	}
	return nil
}

// payloadText extracts a text field from a JSON payload for similarity comparison.
func payloadText(payload json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return string(payload)
	}
	// Try common text fields.
	for _, key := range []string{"text", "description", "note", "context", "blocker"} {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return string(payload)
}
