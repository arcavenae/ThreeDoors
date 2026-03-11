package docaudit

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// roadmapEpicRe matches ROADMAP.md epic headers like:
//
//	### Epic 51: SLAES — Self-Learning Agentic Engineering System (P1) — 0/10 stories done
var roadmapEpicRe = regexp.MustCompile(`^###\s+Epic\s+(\d+(?:\.\d+)?)\s*:\s*(.+)`)

// roadmapProgressRe extracts progress from epic headers like "— 6/6 stories done" or "— 0/10 stories done".
var roadmapProgressRe = regexp.MustCompile(`(\d+)/(\d+)\s+stories?\s+done`)

// roadmapTableRowRe matches story table rows like:
//
//	| 51.5 | Doc Consistency Audit (Periodic Cross-Check) | Not Started | P1 | 51.1 |
var roadmapTableRowRe = regexp.MustCompile(`^\|\s*([\d.]+)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|`)

// ParseRoadmap reads ROADMAP.md and extracts story statuses from table rows
// and epic progress information.
func ParseRoadmap(path string) (map[string]StoryEntry, map[string]EpicEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open roadmap %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	stories := make(map[string]StoryEntry)
	epics := make(map[string]EpicEntry)

	var currentEpicID string
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 512*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for epic headers.
		if m := roadmapEpicRe.FindStringSubmatch(line); m != nil {
			currentEpicID = m[1]
			title := strings.TrimSpace(m[2])
			entry := EpicEntry{
				ID:     currentEpicID,
				Title:  stripPriority(title),
				Status: "Not Started",
				Source: "roadmap",
			}
			// Extract progress from the same line.
			if pm := roadmapProgressRe.FindStringSubmatch(line); pm != nil {
				done := atoi(pm[1])
				total := atoi(pm[2])
				entry.StoriesDone = done
				entry.StoryCount = total
				if done > 0 && done == total {
					entry.Status = "Complete"
				} else if done > 0 {
					entry.Status = "In Progress"
				}
			}
			// Check for COMPLETE in the line itself.
			if strings.Contains(strings.ToUpper(title), "COMPLETE") {
				entry.Status = "Complete"
			}
			epics[currentEpicID] = entry
			continue
		}

		// Check for story table rows (within current epic).
		if currentEpicID != "" {
			if m := roadmapTableRowRe.FindStringSubmatch(line); m != nil {
				storyID := m[1]
				// Skip table header rows.
				if storyID == "Story" || strings.Contains(storyID, "-") {
					continue
				}
				title := strings.TrimSpace(m[2])
				status := strings.TrimSpace(m[3])
				stories[storyID] = StoryEntry{
					ID:     storyID,
					Title:  title,
					Status: normalizeStatus(status),
					Source: "roadmap",
				}
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("scan roadmap %s: %w", path, err)
	}
	return stories, epics, nil
}

// atoi converts a string to int, returning 0 on failure.
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
