package saga

import (
	"fmt"
	"strings"
)

// RecurrenceThreshold is the number of times a saga pattern must recur
// across different PRs before a BOARD.md recommendation is triggered.
const RecurrenceThreshold = 3

// Confidence represents the confidence level of a recommendation.
type Confidence string

const (
	ConfidenceHigh   Confidence = "High"
	ConfidenceMedium Confidence = "Medium"
	ConfidenceLow    Confidence = "Low"
)

// RecurrencePattern represents a saga pattern that recurs across multiple PRs.
type RecurrencePattern struct {
	SagaType    SagaType   `json:"saga_type"`
	BranchCount int        `json:"branch_count"`
	TotalSagas  int        `json:"total_sagas"`
	Branches    []string   `json:"branches"`
	Confidence  Confidence `json:"confidence"`
}

// BoardRecommendation is a structured recommendation for BOARD.md.
type BoardRecommendation struct {
	Pattern  RecurrencePattern
	Proposal string
	Evidence string
}

// RecurrenceTracker analyzes saga findings across PRs to detect recurring patterns.
type RecurrenceTracker struct {
	threshold int
}

// NewRecurrenceTracker creates a tracker with the given recurrence threshold.
func NewRecurrenceTracker(threshold int) *RecurrenceTracker {
	return &RecurrenceTracker{threshold: threshold}
}

// Analyze examines saga findings and returns patterns that exceed the recurrence threshold.
func (rt *RecurrenceTracker) Analyze(findings []SagaFinding) []RecurrencePattern {
	// Group by saga type.
	byType := make(map[SagaType][]SagaFinding)
	for _, f := range findings {
		if f.Type != FindingSagaDetected {
			continue
		}
		byType[f.SagaType] = append(byType[f.SagaType], f)
	}

	var patterns []RecurrencePattern
	for sagaType, typedFindings := range byType {
		// Count distinct branches.
		branches := make(map[string]bool)
		for _, f := range typedFindings {
			branches[f.Branch] = true
		}

		if len(branches) < rt.threshold {
			continue
		}

		branchList := make([]string, 0, len(branches))
		for b := range branches {
			branchList = append(branchList, b)
		}

		patterns = append(patterns, RecurrencePattern{
			SagaType:    sagaType,
			BranchCount: len(branches),
			TotalSagas:  len(typedFindings),
			Branches:    branchList,
			Confidence:  scoreConfidence(len(typedFindings)),
		})
	}

	return patterns
}

// ProduceRecommendation generates a BOARD.md recommendation for a recurring pattern.
func (rt *RecurrenceTracker) ProduceRecommendation(pattern RecurrencePattern) BoardRecommendation {
	var proposal string
	switch pattern.SagaType {
	case SagaTypeEscalationTrap:
		proposal = "Implement mandatory root cause analysis before dispatching follow-up workers for CI fixes. " +
			"When a worker fails, require the supervisor to analyze the full failure chain before dispatching another worker."
	case SagaTypeOverlap:
		proposal = "Implement dispatch deduplication: before dispatching a new worker for a branch, " +
			"check if another worker is already active on that branch. If so, wait for its result."
	default:
		proposal = fmt.Sprintf("Investigate recurring %s saga pattern and propose a systemic fix.", pattern.SagaType)
	}

	evidence := fmt.Sprintf(
		"%d saga events across %d distinct branches. Branches: %s",
		pattern.TotalSagas, pattern.BranchCount, strings.Join(pattern.Branches, ", "),
	)

	return BoardRecommendation{
		Pattern:  pattern,
		Proposal: proposal,
		Evidence: evidence,
	}
}

// FormatBoardEntry formats a recommendation as a BOARD.md table row.
func FormatBoardEntry(id string, rec BoardRecommendation, date string) string {
	return fmt.Sprintf(
		"| %s | %s | %s | retrospector (%s) | %s | Supervisor review |",
		id, rec.Proposal, date, rec.Pattern.Confidence, rec.Evidence,
	)
}

// scoreConfidence assigns a confidence level based on evidence count.
func scoreConfidence(evidenceCount int) Confidence {
	switch {
	case evidenceCount >= 5:
		return ConfidenceHigh
	case evidenceCount >= 3:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}
