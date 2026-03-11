package saga

import (
	"fmt"
	"strings"
	"time"
)

// SagaType classifies the type of saga detected.
type SagaType string

const (
	// SagaTypeOverlap indicates 2+ workers dispatched for same branch within the window.
	SagaTypeOverlap SagaType = "overlap"
	// SagaTypeEscalationTrap indicates a sequential fix-break chain pattern.
	SagaTypeEscalationTrap SagaType = "escalation_trap"
)

// FailureRelation classifies whether failures across workers are related or independent.
type FailureRelation string

const (
	FailureRelated     FailureRelation = "related"
	FailureIndependent FailureRelation = "independent"
	FailureUnknown     FailureRelation = "unknown"
)

// Recommendation is the suggested action for a detected saga.
type Recommendation string

const (
	RecommendTargetedFix Recommendation = "targeted_fix"
	RecommendRevert      Recommendation = "revert_and_reimplement"
	RecommendEscalate    Recommendation = "escalate"
	RecommendRootCause   Recommendation = "root_cause_analysis"
	RecommendCodingStd   Recommendation = "coding_standard_proposal"
)

// CIFailure represents a single CI failure observation.
type CIFailure struct {
	WorkerName  string   `json:"worker_name"`
	RunID       int64    `json:"run_id"`
	Categories  []string `json:"categories"`
	FilesFixed  []string `json:"files_fixed,omitempty"`
	FilesBroken []string `json:"files_broken,omitempty"`
}

// SagaAlert is the structured alert sent to the supervisor.
type SagaAlert struct {
	Type            SagaType         `json:"type"`
	Branch          string           `json:"branch"`
	Workers         []WorkerRecord   `json:"workers"`
	FailureChain    []CIFailure      `json:"failure_chain,omitempty"`
	FailureRelation FailureRelation  `json:"failure_relation"`
	Recommendations []Recommendation `json:"recommendations"`
	Timestamp       time.Time        `json:"timestamp"`
	Summary         string           `json:"summary"`
}

// Detector analyzes branch overlaps and CI failure chains to detect sagas.
type Detector struct {
	escalationThreshold int
}

// NewDetector creates a saga detector. escalationThreshold is the number of
// workers that triggers an escalation trap warning (typically 3).
func NewDetector(escalationThreshold int) *Detector {
	return &Detector{
		escalationThreshold: escalationThreshold,
	}
}

// Analyze takes a branch overlap and optional CI failure data, and produces a SagaAlert.
func (d *Detector) Analyze(overlap BranchOverlap, failures []CIFailure) SagaAlert {
	alert := SagaAlert{
		Branch:    overlap.Branch,
		Workers:   overlap.Workers,
		Timestamp: time.Now().UTC(),
	}

	alert.FailureChain = failures
	alert.FailureRelation = classifyFailures(failures)

	if d.isEscalationTrap(failures) {
		alert.Type = SagaTypeEscalationTrap
		alert.Recommendations = []Recommendation{
			RecommendRootCause,
			RecommendRevert,
			RecommendCodingStd,
		}
		alert.Summary = fmt.Sprintf(
			"Escalation trap on %s: %d workers in sequential fix-break chain. "+
				"Recommend root cause analysis before further fix attempts.",
			overlap.Branch, len(overlap.Workers),
		)
	} else {
		alert.Type = SagaTypeOverlap
		recs := []Recommendation{RecommendTargetedFix}
		if len(overlap.Workers) >= d.escalationThreshold {
			recs = append(recs, RecommendEscalate)
		}
		alert.Recommendations = recs
		alert.Summary = fmt.Sprintf(
			"Saga detected on %s: %d workers dispatched for same fix within window. "+
				"Failures are %s.",
			overlap.Branch, len(overlap.Workers), alert.FailureRelation,
		)
	}

	return alert
}

// isEscalationTrap detects the pattern: Worker 1 fails → Worker 2 fixes A breaks B →
// Worker 3 fixes B breaks C. This requires the failure chain to show sequential
// fix-break relationships where a file fixed by one worker is broken by the next.
func (d *Detector) isEscalationTrap(failures []CIFailure) bool {
	if len(failures) < 2 {
		return false
	}

	chainLen := 0
	for i := 1; i < len(failures); i++ {
		prev := failures[i-1]
		curr := failures[i]

		// Check if current worker's fixed files overlap with previous worker's broken files,
		// OR if current worker broke new files while fixing old ones.
		if hasOverlap(curr.FilesFixed, prev.FilesBroken) && len(curr.FilesBroken) > 0 {
			chainLen++
		}
	}

	// A chain of 1+ fix-break transitions constitutes an escalation trap.
	return chainLen > 0
}

// classifyFailures determines if CI failures across workers are related or independent.
func classifyFailures(failures []CIFailure) FailureRelation {
	if len(failures) < 2 {
		return FailureUnknown
	}

	// Check if failure categories overlap across workers.
	catSets := make([]map[string]bool, len(failures))
	for i, f := range failures {
		catSets[i] = make(map[string]bool)
		for _, c := range f.Categories {
			catSets[i][c] = true
		}
	}

	// Check file overlap between failures.
	allFiles := make([]map[string]bool, len(failures))
	for i, f := range failures {
		allFiles[i] = make(map[string]bool)
		for _, file := range f.FilesFixed {
			allFiles[i][file] = true
		}
		for _, file := range f.FilesBroken {
			allFiles[i][file] = true
		}
	}

	// If failure categories or files overlap between any pair, failures are related.
	for i := 0; i < len(failures); i++ {
		for j := i + 1; j < len(failures); j++ {
			if setsOverlap(catSets[i], catSets[j]) || setsOverlap(allFiles[i], allFiles[j]) {
				return FailureRelated
			}
		}
	}

	return FailureIndependent
}

// hasOverlap returns true if any element in a appears in b.
func hasOverlap(a, b []string) bool {
	set := make(map[string]bool, len(b))
	for _, s := range b {
		set[s] = true
	}
	for _, s := range a {
		if set[s] {
			return true
		}
	}
	return false
}

// setsOverlap returns true if the two sets share any element.
func setsOverlap(a, b map[string]bool) bool {
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

// FormatAlert formats a SagaAlert as a human-readable message for the supervisor.
func FormatAlert(alert SagaAlert) string {
	var b strings.Builder
	fmt.Fprintf(&b, "SAGA DETECTED: %s\n", alert.Type)
	fmt.Fprintf(&b, "Branch: %s\n", alert.Branch)
	fmt.Fprintf(&b, "Workers: ")
	for i, w := range alert.Workers {
		if i > 0 {
			fmt.Fprintf(&b, ", ")
		}
		fmt.Fprintf(&b, "%s (%s)", w.Name, w.Timestamp.Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "\n")

	if len(alert.FailureChain) > 0 {
		fmt.Fprintf(&b, "Failure Chain:\n")
		for i, f := range alert.FailureChain {
			fmt.Fprintf(&b, "  %d. %s: categories=%v", i+1, f.WorkerName, f.Categories)
			if len(f.FilesFixed) > 0 {
				fmt.Fprintf(&b, " fixed=%v", f.FilesFixed)
			}
			if len(f.FilesBroken) > 0 {
				fmt.Fprintf(&b, " broke=%v", f.FilesBroken)
			}
			fmt.Fprintf(&b, "\n")
		}
	}

	fmt.Fprintf(&b, "Failure Relation: %s\n", alert.FailureRelation)
	fmt.Fprintf(&b, "Recommendations: %v\n", alert.Recommendations)
	fmt.Fprintf(&b, "Summary: %s\n", alert.Summary)
	return b.String()
}
