package retrospector

import (
	"time"
)

// Confidence represents the confidence level of a recommendation.
type Confidence string

const (
	// ConfidenceHigh indicates 5+ supporting data points across multiple PRs.
	ConfidenceHigh Confidence = "High"
	// ConfidenceMedium indicates 2-4 supporting data points.
	ConfidenceMedium Confidence = "Medium"
	// ConfidenceLow indicates 1 data point or extrapolation from limited evidence.
	ConfidenceLow Confidence = "Low"
)

// Recommendation represents a structured improvement recommendation
// filed to BOARD.md by the retrospector agent.
type Recommendation struct {
	ID         string
	Text       string
	Date       time.Time
	Source     string
	Confidence Confidence
	Evidence   []string
	Link       string
	Awaiting   string
}

// Outcome tracks the resolution of a recommendation.
type Outcome string

const (
	OutcomeAccepted Outcome = "accepted"
	OutcomeRejected Outcome = "rejected"
	OutcomeDeferred Outcome = "deferred"
)
