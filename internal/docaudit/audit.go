package docaudit

import (
	"fmt"
	"sort"
	"time"
)

// Auditor performs cross-reference checks across the planning doc chain.
type Auditor struct {
	docs DocSet
}

// NewAuditor creates an Auditor from a fully-parsed DocSet.
func NewAuditor(docs DocSet) *Auditor {
	return &Auditor{docs: docs}
}

// Run performs the full consistency audit and returns an AuditResult.
func (a *Auditor) Run() AuditResult {
	var findings []Finding
	findings = append(findings, a.checkStoryStatusMismatches()...)
	findings = append(findings, a.checkOrphanedStories()...)
	findings = append(findings, a.checkPhantomStories()...)
	findings = append(findings, a.checkEpicStatusMismatches()...)

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Type != findings[j].Type {
			return findings[i].Type < findings[j].Type
		}
		if findings[i].EpicID != findings[j].EpicID {
			return findings[i].EpicID < findings[j].EpicID
		}
		return findings[i].StoryID < findings[j].StoryID
	})

	return AuditResult{
		Timestamp: time.Now().UTC(),
		Findings:  findings,
		Clean:     len(findings) == 0,
	}
}

// checkStoryStatusMismatches compares story file statuses against ROADMAP table entries.
func (a *Auditor) checkStoryStatusMismatches() []Finding {
	var findings []Finding

	// Story files are authoritative for individual story status.
	for id, storyFile := range a.docs.StoryFiles {
		// Compare against ROADMAP.
		if roadmap, ok := a.docs.Roadmap[id]; ok {
			if !statusesMatch(storyFile.Status, roadmap.Status) {
				findings = append(findings, Finding{
					Type:        FindingStatusMismatch,
					StoryID:     id,
					Description: fmt.Sprintf("Story %s status differs between story file and ROADMAP", id),
					Expected:    storyFile.Status,
					Actual:      roadmap.Status,
					Authority:   "story_file",
					Fix:         fmt.Sprintf("Update ROADMAP.md story %s status from %q to %q", id, roadmap.Status, storyFile.Status),
				})
			}
		}
	}

	return findings
}

// checkOrphanedStories finds story files that have no corresponding entry in
// the planning docs (epics-and-stories.md or ROADMAP.md).
func (a *Auditor) checkOrphanedStories() []Finding {
	var findings []Finding

	for id, sf := range a.docs.StoryFiles {
		_, inEAS := a.docs.EpicsAndStories[id]
		_, inRoadmap := a.docs.Roadmap[id]

		if !inEAS && !inRoadmap {
			findings = append(findings, Finding{
				Type:        FindingOrphanedStory,
				StoryID:     id,
				Description: fmt.Sprintf("Story file %s exists but is not referenced in epics-and-stories.md or ROADMAP.md", id),
				Expected:    "Story referenced in planning docs",
				Actual:      fmt.Sprintf("Story file exists with status %q but no planning doc reference", sf.Status),
				Authority:   "epics_and_stories",
				Fix:         fmt.Sprintf("Add story %s to epics-and-stories.md and ROADMAP.md, or remove the orphaned story file", id),
			})
		}
	}

	return findings
}

// checkPhantomStories finds entries in planning docs that have no corresponding
// story file.
func (a *Auditor) checkPhantomStories() []Finding {
	var findings []Finding

	// Check ROADMAP entries against story files.
	for id := range a.docs.Roadmap {
		if _, ok := a.docs.StoryFiles[id]; !ok {
			findings = append(findings, Finding{
				Type:        FindingPhantomStory,
				StoryID:     id,
				Description: fmt.Sprintf("ROADMAP.md references story %s but no story file exists", id),
				Expected:    fmt.Sprintf("docs/stories/%s.story.md should exist", id),
				Actual:      "No story file found",
				Authority:   "roadmap",
				Fix:         fmt.Sprintf("Create story file docs/stories/%s.story.md or remove the ROADMAP entry", id),
			})
		}
	}

	// Check epics-and-stories entries against story files.
	for id := range a.docs.EpicsAndStories {
		if _, ok := a.docs.StoryFiles[id]; !ok {
			// Only report if not already reported via ROADMAP check.
			if _, inRoadmap := a.docs.Roadmap[id]; !inRoadmap {
				findings = append(findings, Finding{
					Type:        FindingPhantomStory,
					StoryID:     id,
					Description: fmt.Sprintf("epics-and-stories.md references story %s but no story file exists", id),
					Expected:    fmt.Sprintf("docs/stories/%s.story.md should exist", id),
					Actual:      "No story file found",
					Authority:   "epics_and_stories",
					Fix:         fmt.Sprintf("Create story file docs/stories/%s.story.md or remove the planning doc entry", id),
				})
			}
		}
	}

	return findings
}

// checkEpicStatusMismatches compares epic statuses between epic-list.md and
// epics-and-stories.md, and between epic-list.md and ROADMAP.md.
func (a *Auditor) checkEpicStatusMismatches() []Finding {
	var findings []Finding

	// epic-list.md is authoritative for epic status alongside epics-and-stories.md.
	for id, epicList := range a.docs.EpicList {
		// Compare against ROADMAP epic entries.
		if roadmap, ok := a.docs.RoadmapEpics[id]; ok {
			if !epicStatusesMatch(epicList.Status, roadmap.Status) {
				findings = append(findings, Finding{
					Type:        FindingEpicMismatch,
					EpicID:      id,
					Description: fmt.Sprintf("Epic %s status differs between epic-list.md and ROADMAP.md", id),
					Expected:    epicList.Status,
					Actual:      roadmap.Status,
					Authority:   "epic_list",
					Fix:         fmt.Sprintf("Update ROADMAP.md epic %s status to match epic-list.md: %q", id, epicList.Status),
				})
			}
		}
	}

	return findings
}

// statusesMatch compares two normalized story statuses, treating certain
// statuses as equivalent.
func statusesMatch(authoritative, other string) bool {
	a := normalizeStatus(authoritative)
	b := normalizeStatus(other)
	return a == b
}

// epicStatusesMatch compares two normalized epic statuses.
func epicStatusesMatch(a, b string) bool {
	return normalizeEpicStatus(a) == normalizeEpicStatus(b)
}
