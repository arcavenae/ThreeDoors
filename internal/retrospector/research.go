package retrospector

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	staleDays     = 14
	abandonedDays = 28
)

// LifecycleState represents the current lifecycle state of a research artifact.
type LifecycleState string

const (
	// LifecycleActive is for artifacts that are recently modified or referenced in open work.
	LifecycleActive LifecycleState = "active"
	// LifecycleFormalized is for artifacts that have corresponding epics/stories.
	LifecycleFormalized LifecycleState = "formalized"
	// LifecycleStale is for artifacts >2 weeks old with no references.
	LifecycleStale LifecycleState = "stale"
	// LifecycleAbandoned is for artifacts >4 weeks old with no references.
	LifecycleAbandoned LifecycleState = "abandoned"
)

// ResearchArtifact represents a markdown file in the planning artifacts directory.
type ResearchArtifact struct {
	Name     string
	Path     string
	ModTime  time.Time
	Contents string
}

// ArtifactReferences tracks where a research artifact is referenced.
type ArtifactReferences struct {
	StoryRefs  []string // story IDs that reference this artifact
	BoardRefs  []string // BOARD.md entry IDs that reference this artifact
	EpicDocRef bool     // whether epics-and-stories.md references this artifact
}

// ResearchLifecycleResult pairs an artifact with its classified lifecycle state.
type ResearchLifecycleResult struct {
	Artifact ResearchArtifact
	Refs     ArtifactReferences
	State    LifecycleState
}

// LifecycleReport summarizes the state of all research artifacts.
type LifecycleReport struct {
	Total              int
	ActiveCount        int
	FormalizedCount    int
	StaleCount         int
	AbandonedCount     int
	OldestUnformalized string
	StaleArtifacts     []string
	AbandonedArtifacts []string
}

// ScanResearchArtifacts reads all .md files from the given directory.
// Returns an empty slice (not error) if the directory doesn't exist.
func ScanResearchArtifacts(dir string) ([]ResearchArtifact, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read artifacts dir %s: %w", dir, err)
	}

	var artifacts []ResearchArtifact
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		artifacts = append(artifacts, ResearchArtifact{
			Name:     entry.Name(),
			Path:     path,
			ModTime:  info.ModTime().UTC(),
			Contents: string(content),
		})
	}

	return artifacts, nil
}

// FindArtifactReferences searches for references to the given artifact across
// story files, epics-and-stories.md, and BOARD.md.
func FindArtifactReferences(artifact ResearchArtifact, projectRoot string) (ArtifactReferences, error) {
	var refs ArtifactReferences

	// Search story files
	storiesDir := filepath.Join(projectRoot, "docs", "stories")
	storyRefs, err := findRefsInDir(artifact.Name, storiesDir, "*.story.md")
	if err != nil {
		return refs, fmt.Errorf("search stories: %w", err)
	}
	refs.StoryRefs = storyRefs

	// Search epics-and-stories.md
	epicDocPath := filepath.Join(projectRoot, "docs", "prd", "epics-and-stories.md")
	found, err := fileContains(epicDocPath, artifact.Name)
	if err != nil {
		return refs, fmt.Errorf("search epic doc: %w", err)
	}
	refs.EpicDocRef = found

	// Search BOARD.md
	boardPath := filepath.Join(projectRoot, "docs", "decisions", "BOARD.md")
	boardRefs, err := findBoardRefsForArtifact(artifact.Name, boardPath)
	if err != nil {
		return refs, fmt.Errorf("search board: %w", err)
	}
	refs.BoardRefs = boardRefs

	return refs, nil
}

// ClassifyLifecycle determines the lifecycle state of a research artifact
// based on its references and age. The classification is:
//   - Formalized: has story refs AND epic doc ref (research became stories)
//   - Active: <2 weeks old, OR has pending board refs (still in progress)
//   - Abandoned: >4 weeks old with no references
//   - Stale: >2 weeks old with no references
func ClassifyLifecycle(artifact ResearchArtifact, refs ArtifactReferences, now time.Time) LifecycleState {
	// Formalized: has both story references and epic doc reference
	if len(refs.StoryRefs) > 0 && refs.EpicDocRef {
		return LifecycleFormalized
	}

	// Active if has pending board refs (recommendation in progress)
	if len(refs.BoardRefs) > 0 {
		return LifecycleActive
	}

	age := now.Sub(artifact.ModTime)

	// Within 2 weeks — active regardless of references
	if age < time.Duration(staleDays)*24*time.Hour {
		return LifecycleActive
	}

	// >4 weeks — abandoned
	if age >= time.Duration(abandonedDays)*24*time.Hour {
		return LifecycleAbandoned
	}

	// >2 weeks but <4 weeks — stale
	return LifecycleStale
}

// GenerateLifecycleReport aggregates lifecycle results into a summary report.
func GenerateLifecycleReport(results []ResearchLifecycleResult) LifecycleReport {
	var report LifecycleReport
	report.Total = len(results)

	var oldestUnformalizedTime time.Time
	for _, r := range results {
		switch r.State {
		case LifecycleActive:
			report.ActiveCount++
		case LifecycleFormalized:
			report.FormalizedCount++
		case LifecycleStale:
			report.StaleCount++
			report.StaleArtifacts = append(report.StaleArtifacts, r.Artifact.Name)
		case LifecycleAbandoned:
			report.AbandonedCount++
			report.AbandonedArtifacts = append(report.AbandonedArtifacts, r.Artifact.Name)
		}

		// Track oldest unformalized (anything not formalized)
		if r.State != LifecycleFormalized {
			if oldestUnformalizedTime.IsZero() || r.Artifact.ModTime.Before(oldestUnformalizedTime) {
				oldestUnformalizedTime = r.Artifact.ModTime
				report.OldestUnformalized = r.Artifact.Name
			}
		}
	}

	return report
}

// StaleResearchAlerts generates alert messages for stale and abandoned artifacts.
// Each alert suggests the appropriate action (formalization or closure).
func StaleResearchAlerts(results []ResearchLifecycleResult) []string {
	var alerts []string
	for _, r := range results {
		switch r.State {
		case LifecycleStale:
			alerts = append(alerts, fmt.Sprintf(
				"Stale research: %s — needs formalization or explicit closure (last modified %s)",
				r.Artifact.Name,
				r.Artifact.ModTime.Format(time.DateOnly),
			))
		case LifecycleAbandoned:
			alerts = append(alerts, fmt.Sprintf(
				"Abandoned research: %s — needs formalization or explicit closure (last modified %s, >4 weeks old)",
				r.Artifact.Name,
				r.Artifact.ModTime.Format(time.DateOnly),
			))
		}
	}
	return alerts
}

// findRefsInDir searches files matching a glob pattern in a directory for
// references to an artifact name. Returns a list of file base names that
// contain the reference.
func findRefsInDir(artifactName, dir, pattern string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("glob %s/%s: %w", dir, pattern, err)
	}

	var refs []string
	for _, path := range matches {
		found, err := fileContains(path, artifactName)
		if err != nil {
			return nil, fmt.Errorf("check %s: %w", path, err)
		}
		if found {
			// Extract story ID from filename like "51.1.story.md"
			base := filepath.Base(path)
			ref := strings.TrimSuffix(base, ".story.md")
			refs = append(refs, ref)
		}
	}

	return refs, nil
}

// fileContains checks whether a file contains the given substring.
// Returns false (not error) if the file doesn't exist.
func fileContains(path, needle string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close() //nolint:errcheck // read-only

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), needle) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

// findBoardRefsForArtifact searches BOARD.md for rows referencing the artifact.
// Returns the IDs (P-NNN, D-NNN, R-NNN, Q-NNN) of matching rows.
func findBoardRefsForArtifact(artifactName, boardPath string) ([]string, error) {
	f, err := os.Open(boardPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open board: %w", err)
	}
	defer f.Close() //nolint:errcheck // read-only

	var refs []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, artifactName) {
			continue
		}
		// Extract ID from table row: "| P-001 | ..." or "| D-123 | ..."
		if id := extractBoardID(line); id != "" {
			refs = append(refs, id)
		}
	}

	return refs, scanner.Err()
}

// extractBoardID extracts an ID (P-NNN, D-NNN, R-NNN, Q-NNN) from a
// markdown table row.
func extractBoardID(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return ""
	}

	parts := strings.SplitN(trimmed, "|", 3)
	if len(parts) < 3 {
		return ""
	}

	id := strings.TrimSpace(parts[1])
	if len(id) >= 3 && (strings.HasPrefix(id, "P-") ||
		strings.HasPrefix(id, "D-") ||
		strings.HasPrefix(id, "R-") ||
		strings.HasPrefix(id, "Q-")) {
		return id
	}

	return ""
}
