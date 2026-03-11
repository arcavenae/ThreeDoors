package docaudit

import "time"

// StoryEntry represents a story's status from one of the planning documents.
type StoryEntry struct {
	ID     string // e.g., "51.5"
	Title  string
	Status string // e.g., "Done (PR #123)", "Not Started", "In Progress"
	Source string // which doc this came from
}

// EpicEntry represents an epic's status from one of the planning documents.
type EpicEntry struct {
	ID          string // e.g., "51"
	Title       string
	Status      string // e.g., "COMPLETE", "Not Started", "IN PROGRESS"
	StoryCount  int    // total stories
	StoriesDone int    // stories marked done
	Source      string
}

// FindingType categorizes what kind of inconsistency was found.
type FindingType string

const (
	FindingStatusMismatch  FindingType = "status_mismatch"
	FindingOrphanedStory   FindingType = "orphaned_story"
	FindingPhantomStory    FindingType = "phantom_story"
	FindingEpicMismatch    FindingType = "epic_status_mismatch"
	FindingMissingStoryRef FindingType = "missing_story_ref"
)

// Finding represents a single doc inconsistency.
type Finding struct {
	Type        FindingType `json:"finding_type"`
	StoryID     string      `json:"story_id,omitempty"`
	EpicID      string      `json:"epic_id,omitempty"`
	Description string      `json:"description"`
	Expected    string      `json:"expected"`
	Actual      string      `json:"actual"`
	Authority   string      `json:"authority"` // which doc is authoritative
	Fix         string      `json:"fix"`       // recommended correction
}

// AuditResult is the full output of a doc consistency audit.
type AuditResult struct {
	Timestamp time.Time `json:"timestamp"`
	Findings  []Finding `json:"findings"`
	Clean     bool      `json:"clean"`
}

// JSONLEntry is one line in the retrospector findings log.
type JSONLEntry struct {
	Type      string    `json:"type"` // "doc_inconsistency" or "doc_audit_clean"
	Timestamp string    `json:"timestamp"`
	Repo      string    `json:"repo"`
	Findings  []Finding `json:"findings,omitempty"`
	Summary   string    `json:"summary"`
}

// RotationMode identifies one of the four deep analysis modes.
type RotationMode string

const (
	ModeDocConsistency   RotationMode = "doc_consistency"
	ModeConflictAnalysis RotationMode = "conflict_analysis"
	ModeCIAnalysis       RotationMode = "ci_analysis"
	ModeProcessWaste     RotationMode = "process_waste"
)

// RotationState tracks which deep analysis mode ran last.
type RotationState struct {
	LastMode    RotationMode `json:"last_mode"`
	LastRun     time.Time    `json:"last_run"`
	CycleCount  int          `json:"cycle_count"`
	ModeHistory []ModeRun    `json:"mode_history"`
}

// ModeRun records a single execution of a deep analysis mode.
type ModeRun struct {
	Mode      RotationMode `json:"mode"`
	Timestamp time.Time    `json:"timestamp"`
	Clean     bool         `json:"clean"`
}

// AllModes returns the ordered list of rotation modes.
func AllModes() []RotationMode {
	return []RotationMode{
		ModeDocConsistency,
		ModeConflictAnalysis,
		ModeCIAnalysis,
		ModeProcessWaste,
	}
}

// DocSet holds parsed data from all four planning document sources.
type DocSet struct {
	StoryFiles      map[string]StoryEntry // keyed by story ID
	EpicList        map[string]EpicEntry  // keyed by epic ID
	EpicsAndStories map[string]StoryEntry // keyed by story ID
	Roadmap         map[string]StoryEntry // keyed by story ID
	RoadmapEpics    map[string]EpicEntry  // keyed by epic ID
	EpicListEpics   map[string]EpicEntry  // keyed by epic ID (alias for EpicList)
}
