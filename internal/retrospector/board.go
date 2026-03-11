package retrospector

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	pendingHeader = "## Pending Recommendations"
	tableHeader   = "| ID | Recommendation | Date | Source | Link | Awaiting |"
	tableSep      = "|----|----------------|------|--------|------|----------|"
)

var pIDRegex = regexp.MustCompile(`^\|\s*P-(\d+)\s*\|`)

// BoardWriter reads and appends recommendations to BOARD.md.
type BoardWriter struct {
	path string
}

// NewBoardWriter creates a BoardWriter for the given BOARD.md path.
func NewBoardWriter(path string) *BoardWriter {
	return &BoardWriter{path: path}
}

// NextID scans BOARD.md for the highest existing P-NNN ID and returns
// the next available ID string (e.g., "P-007").
func (w *BoardWriter) NextID() (string, error) {
	content, err := os.ReadFile(w.path)
	if err != nil {
		return "", fmt.Errorf("read board %s: %w", w.path, err)
	}
	return nextIDFromContent(string(content))
}

func nextIDFromContent(content string) (string, error) {
	maxID := 0
	for _, line := range strings.Split(content, "\n") {
		matches := pIDRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			n, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if n > maxID {
				maxID = n
			}
		}
	}
	return fmt.Sprintf("P-%03d", maxID+1), nil
}

// AppendRecommendation adds a new recommendation row to the Pending
// Recommendations table in BOARD.md.
func (w *BoardWriter) AppendRecommendation(rec Recommendation) error {
	content, err := os.ReadFile(w.path)
	if err != nil {
		return fmt.Errorf("read board %s: %w", w.path, err)
	}

	newContent, err := appendRecommendationToContent(string(content), rec)
	if err != nil {
		return err
	}

	return os.WriteFile(w.path, []byte(newContent), 0o600)
}

func appendRecommendationToContent(content string, rec Recommendation) (string, error) {
	lines := strings.Split(content, "\n")

	// Find the Pending Recommendations section
	pendingIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == pendingHeader {
			pendingIdx = i
			break
		}
	}
	if pendingIdx < 0 {
		return "", fmt.Errorf("section %q not found in BOARD.md", pendingHeader)
	}

	// Find the last table row in the Pending Recommendations section.
	// We insert after the last P-NNN row, or after the table separator
	// if the table is empty.
	insertIdx := -1
	for i := pendingIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		// Stop at the next section header
		if strings.HasPrefix(trimmed, "## ") {
			break
		}
		// Track table header/sep and data rows
		if trimmed == tableHeader || trimmed == tableSep {
			insertIdx = i
			continue
		}
		if pIDRegex.MatchString(trimmed) {
			insertIdx = i
		}
	}

	if insertIdx < 0 {
		return "", fmt.Errorf("could not locate table in %q section", pendingHeader)
	}

	row := formatRecommendationRow(rec)

	// Insert after insertIdx
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:insertIdx+1]...)
	result = append(result, row)
	result = append(result, lines[insertIdx+1:]...)

	return strings.Join(result, "\n"), nil
}

func formatRecommendationRow(rec Recommendation) string {
	source := rec.Source
	if rec.Confidence != "" {
		source = fmt.Sprintf("%s (%s)", rec.Source, rec.Confidence)
	}
	date := rec.Date.Format(time.DateOnly)
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s |",
		rec.ID, rec.Text, date, source, rec.Link, rec.Awaiting)
}
