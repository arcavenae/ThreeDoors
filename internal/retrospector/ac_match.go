package retrospector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// storyRefPattern matches "Story X.Y" in commit messages and PR titles.
var storyRefPattern = regexp.MustCompile(`(?i)story\s+(\d+\.\d+)`)

// ParseStoryRef extracts a story reference (e.g. "51.3") from text.
// Returns empty string if no reference found.
func ParseStoryRef(text string) string {
	matches := storyRefPattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// CalculateACMatch determines how well a PR's changed files align with a story's tasks.
// storiesDir is the path to docs/stories/, storyRef is e.g. "51.3",
// and prFiles is the list of files changed in the PR.
func CalculateACMatch(storiesDir string, storyRef string, prFiles []string) (ACMatch, error) {
	if storyRef == "" {
		return ACMatchNoStory, nil
	}

	storyPath := filepath.Join(storiesDir, storyRef+".story.md")
	taskPaths, err := extractTaskPaths(storyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Story file doesn't exist — can't verify, treat as partial
			return ACMatchPartial, nil
		}
		return ACMatchNone, fmt.Errorf("extract task paths from %s: %w", storyPath, err)
	}

	if len(taskPaths) == 0 {
		// Story has no identifiable file-referencing tasks — can't verify
		return ACMatchFull, nil
	}

	matched := 0
	for _, taskPath := range taskPaths {
		for _, prFile := range prFiles {
			if pathOverlaps(taskPath, prFile) {
				matched++
				break
			}
		}
	}

	switch {
	case matched == len(taskPaths):
		return ACMatchFull, nil
	case matched > 0:
		return ACMatchPartial, nil
	default:
		return ACMatchNone, nil
	}
}

// extractTaskPaths reads a story file and extracts directory/file paths
// mentioned in task descriptions. These are used to compare against PR file changes.
func extractTaskPaths(storyPath string) ([]string, error) {
	f, err := os.Open(storyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only

	var paths []string
	scanner := bufio.NewScanner(f)

	// Pattern that captures full file paths like internal/widget/handler.go
	// or directory paths like internal/widget
	pathPattern := regexp.MustCompile(`(?:internal|cmd|pkg)/[\w/]+(?:\.go)?`)

	inTasks := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## Tasks") || strings.HasPrefix(line, "### Task") {
			inTasks = true
			continue
		}
		if inTasks && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			break
		}

		if !inTasks {
			continue
		}

		for _, match := range pathPattern.FindAllString(line, -1) {
			paths = appendUnique(paths, match)
		}
	}

	return paths, scanner.Err()
}

// pathOverlaps checks whether a task path reference overlaps with a PR file path.
// A task path "internal/retrospector" overlaps with "internal/retrospector/log.go".
// A task path "log.go" overlaps with "internal/retrospector/log.go".
func pathOverlaps(taskPath, prFile string) bool {
	if strings.Contains(prFile, taskPath) {
		return true
	}
	// Check if the task path is just a filename that matches the PR file's basename
	if !strings.Contains(taskPath, "/") {
		return filepath.Base(prFile) == taskPath
	}
	return false
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
