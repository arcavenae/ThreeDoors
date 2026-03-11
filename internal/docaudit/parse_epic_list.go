package docaudit

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// epicListHeaderRe matches epic headers like:
//
//	**Epic 1: Three Doors Technical Demo** COMPLETE
//	**Epic 16: iPhone Mobile App (SwiftUI)** ICEBOX
var epicListHeaderRe = regexp.MustCompile(`\*\*Epic\s+(\d+(?:\.\d+)?)\s*:\s*(.+?)\*\*\s*(.*)`)

// ParseEpicList reads docs/prd/epic-list.md and extracts epic statuses.
func ParseEpicList(path string) (map[string]EpicEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open epic-list %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	result := make(map[string]EpicEntry)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		m := epicListHeaderRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		epicID := m[1]
		title := strings.TrimSpace(m[2])
		statusRaw := strings.TrimSpace(m[3])

		result[epicID] = EpicEntry{
			ID:     epicID,
			Title:  title,
			Status: normalizeEpicStatus(statusRaw),
			Source: "epic_list",
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan epic-list %s: %w", path, err)
	}
	return result, nil
}

// normalizeEpicStatus normalizes epic status strings.
func normalizeEpicStatus(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)

	switch {
	case strings.Contains(lower, "complete"):
		return "Complete"
	case strings.Contains(lower, "icebox"):
		return "Icebox"
	case strings.Contains(lower, "in progress"):
		return "In Progress"
	case strings.Contains(lower, "not started"):
		return "Not Started"
	case s == "":
		return "Not Started"
	default:
		return s
	}
}
