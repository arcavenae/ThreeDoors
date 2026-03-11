package retrospector

import (
	"time"
)

// ACMatch represents how well a PR's changes align with its story's acceptance criteria.
type ACMatch string

const (
	// ACMatchFull indicates all story tasks were addressed by the PR.
	ACMatchFull ACMatch = "full"
	// ACMatchPartial indicates some story tasks were addressed.
	ACMatchPartial ACMatch = "partial"
	// ACMatchNone indicates no overlap between PR files and story tasks.
	ACMatchNone ACMatch = "none"
	// ACMatchNoStory indicates the PR had no story reference.
	ACMatchNoStory ACMatch = "no-story"
)

// Finding represents a single retrospector observation about a merged PR.
// Each finding is stored as one line in the JSONL findings log.
type Finding struct {
	PR           int       `json:"pr"`
	StoryRef     string    `json:"story_ref,omitempty"`
	ACMatch      ACMatch   `json:"ac_match"`
	CIFirstPass  bool      `json:"ci_first_pass"`
	CIFailures   []string  `json:"ci_failures,omitempty"`
	Conflicts    int       `json:"conflicts"`
	RebaseCount  int       `json:"rebase_count"`
	Timestamp    time.Time `json:"timestamp"`
	Title        string    `json:"title,omitempty"`
	FilesChanged int       `json:"files_changed"`
	ProcessGap   bool      `json:"process_gap,omitempty"`
}
