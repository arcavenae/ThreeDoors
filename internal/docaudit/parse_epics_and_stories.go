package docaudit

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// epicHeaderRe matches epic section headers like:
//
//	## Epic 51: SLAES — Self-Learning Agentic Engineering System (P1)
var epicHeaderRe = regexp.MustCompile(`^##\s+Epic\s+(\d+(?:\.\d+)?)\s*:\s*(.+)`)

// epicStatusLineRe matches the status line within an epic section:
//
//	**Status:** Not Started (0/10 stories)
//	**Status:** COMPLETE
var epicStatusLineRe = regexp.MustCompile(`\*\*Status:\*\*\s*(.+)`)

// storySubheadingRe matches story sub-headings within epics-and-stories.md:
//
//	#### Story 51.5: Doc Consistency Audit (Periodic Cross-Check)
var storySubheadingRe = regexp.MustCompile(`^####?\s+Story\s+([\d.]+)\s*:\s*(.+)`)

// ParseEpicsAndStories reads docs/prd/epics-and-stories.md and extracts both
// epic-level and story-level information.
func ParseEpicsAndStories(path string) (map[string]StoryEntry, map[string]EpicEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open epics-and-stories %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	stories := make(map[string]StoryEntry)
	epics := make(map[string]EpicEntry)

	var currentEpicID string
	scanner := bufio.NewScanner(f)
	// Increase the scanner buffer for large files.
	buf := make([]byte, 0, 256*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for epic headers.
		if m := epicHeaderRe.FindStringSubmatch(line); m != nil {
			currentEpicID = m[1]
			title := strings.TrimSpace(m[2])
			// Strip trailing priority marker like "(P1)".
			title = stripPriority(title)
			epics[currentEpicID] = EpicEntry{
				ID:     currentEpicID,
				Title:  title,
				Status: "Not Started",
				Source: "epics_and_stories",
			}
			continue
		}

		// Check for epic status line (within current epic context).
		if currentEpicID != "" {
			if m := epicStatusLineRe.FindStringSubmatch(line); m != nil {
				entry := epics[currentEpicID]
				entry.Status = normalizeEpicStatus(m[1])
				epics[currentEpicID] = entry
				continue
			}
		}

		// Check for story sub-headings.
		if m := storySubheadingRe.FindStringSubmatch(line); m != nil {
			storyID := m[1]
			title := strings.TrimSpace(m[2])
			stories[storyID] = StoryEntry{
				ID:     storyID,
				Title:  title,
				Status: inferStoryStatusFromTitle(title),
				Source: "epics_and_stories",
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scan epics-and-stories %s: %w", path, err)
	}
	return stories, epics, nil
}

// stripPriority removes trailing " (P0)", " (P1)", " (P2)" etc. from titles.
func stripPriority(s string) string {
	re := regexp.MustCompile(`\s*\(P\d\)\s*$`)
	return re.ReplaceAllString(s, "")
}

// inferStoryStatusFromTitle checks if a title ends with a status emoji.
func inferStoryStatusFromTitle(title string) string {
	if strings.HasSuffix(title, "✅") {
		return "Done"
	}
	// The epics-and-stories.md format doesn't have per-story status lines;
	// status is tracked at the epic level and in story files.
	return "Referenced"
}
