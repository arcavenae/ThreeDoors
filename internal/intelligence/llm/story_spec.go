package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StorySpec represents a BMAD-style story specification output from task decomposition.
type StorySpec struct {
	Epic      string   `json:"epic"`
	StoryID   string   `json:"story_id"`
	Title     string   `json:"title"`
	UserStory string   `json:"user_story"`
	ACs       []string `json:"acceptance_criteria"`
	Tasks     []string `json:"tasks"`
	DevNotes  string   `json:"dev_notes"`
}

// DecompositionResult holds the full output of a task decomposition operation.
type DecompositionResult struct {
	SourceTask  string      `json:"source_task"`
	Stories     []StorySpec `json:"stories"`
	Backend     string      `json:"backend"`
	GeneratedAt time.Time   `json:"generated_at"`
}

// Validate checks that a StorySpec has all required fields populated.
func (s *StorySpec) Validate() error {
	if s.StoryID == "" {
		return fmt.Errorf("validate story spec: story_id is required")
	}
	if s.Title == "" {
		return fmt.Errorf("validate story spec %s: title is required", s.StoryID)
	}
	if s.UserStory == "" {
		return fmt.Errorf("validate story spec %s: user_story is required", s.StoryID)
	}
	if len(s.ACs) == 0 {
		return fmt.Errorf("validate story spec %s: at least one acceptance criterion is required", s.StoryID)
	}
	if len(s.Tasks) == 0 {
		return fmt.Errorf("validate story spec %s: at least one task is required", s.StoryID)
	}
	return nil
}

// ToMarkdown renders the story spec as a BMAD-compatible Markdown document.
func (s *StorySpec) ToMarkdown() string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Story %s: %s\n\n", s.StoryID, s.Title)
	fmt.Fprintf(&b, "## Story\n\n%s\n\n", s.UserStory)
	fmt.Fprintf(&b, "## Status\n\nReady for Dev\n\n")

	fmt.Fprintf(&b, "## Acceptance Criteria\n\n")
	for _, ac := range s.ACs {
		fmt.Fprintf(&b, "- %s\n", ac)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "## Tasks\n\n")
	for i, task := range s.Tasks {
		fmt.Fprintf(&b, "%d. %s\n", i+1, task)
	}
	b.WriteString("\n")

	if s.DevNotes != "" {
		fmt.Fprintf(&b, "## Dev Notes\n\n%s\n", s.DevNotes)
	}

	return b.String()
}

// ParseStorySpecs extracts StorySpec objects from JSON output.
// Accepts either a JSON array of specs or a JSON object with a "stories" key.
func ParseStorySpecs(raw string) ([]StorySpec, error) {
	trimmed := strings.TrimSpace(raw)

	// Try parsing as a DecompositionResult first.
	var result DecompositionResult
	if err := json.Unmarshal([]byte(trimmed), &result); err == nil && len(result.Stories) > 0 {
		return result.Stories, nil
	}

	// Try parsing as a direct array of StorySpec.
	var specs []StorySpec
	if err := json.Unmarshal([]byte(trimmed), &specs); err == nil && len(specs) > 0 {
		return specs, nil
	}

	// Try parsing as a single StorySpec.
	var single StorySpec
	if err := json.Unmarshal([]byte(trimmed), &single); err == nil && single.StoryID != "" {
		return []StorySpec{single}, nil
	}

	return nil, fmt.Errorf("parse story specs: could not parse LLM output as story specs")
}
