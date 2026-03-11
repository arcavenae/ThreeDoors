package docaudit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// storyFileNameRe matches story file names like "51.5.story.md" or "3.5.1.story.md".
var storyFileNameRe = regexp.MustCompile(`^(\d+(?:\.\d+)+)\.story\.md$`)

// storyStatusRe matches status lines in various formats:
//   - ## Status: Done (PR #123)
//   - **Status:** Not Started
//   - Status: In Progress
//   - - **Status:** In Review (PR #456)
var storyStatusRe = regexp.MustCompile(`(?:^#+\s*|^\*\*|^-\s*\*\*|^)Status:?\*?\*?\s*(.+)`)

// storyTitleRe matches the story title from the first heading.
var storyTitleRe = regexp.MustCompile(`^#\s+Story\s+[\d.]+:\s*(.+)`)

// ParseStoryFiles reads all story files from the given directory and extracts
// their IDs and statuses.
func ParseStoryFiles(dir string) (map[string]StoryEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read story dir %s: %w", dir, err)
	}

	result := make(map[string]StoryEntry)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := storyFileNameRe.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		storyID := matches[1]
		se, err := parseOneStoryFile(filepath.Join(dir, entry.Name()), storyID)
		if err != nil {
			return nil, err
		}
		result[storyID] = se
	}
	return result, nil
}

func parseOneStoryFile(path, storyID string) (StoryEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return StoryEntry{}, fmt.Errorf("open story file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	se := StoryEntry{
		ID:     storyID,
		Source: "story_file",
		Status: "Unknown",
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if se.Title == "" {
			if m := storyTitleRe.FindStringSubmatch(line); m != nil {
				se.Title = strings.TrimSpace(m[1])
			}
		}

		if m := storyStatusRe.FindStringSubmatch(line); m != nil {
			se.Status = normalizeStatus(strings.TrimSpace(m[1]))
			break // status found, no need to continue
		}
	}
	if err := scanner.Err(); err != nil {
		return StoryEntry{}, fmt.Errorf("scan story file %s: %w", path, err)
	}
	return se, nil
}

// normalizeStatus cleans up status strings for consistent comparison.
func normalizeStatus(s string) string {
	s = strings.TrimSpace(s)
	// Remove trailing markdown artifacts
	s = strings.TrimRight(s, "*")
	s = strings.TrimSpace(s)

	lower := strings.ToLower(s)

	switch {
	case strings.HasPrefix(lower, "done"):
		return "Done"
	case strings.HasPrefix(lower, "in review"):
		return "In Review"
	case strings.HasPrefix(lower, "in progress"):
		return "In Progress"
	case strings.HasPrefix(lower, "not started"):
		return "Not Started"
	case lower == "draft":
		return "Draft"
	case lower == "icebox":
		return "Icebox"
	default:
		return s
	}
}
