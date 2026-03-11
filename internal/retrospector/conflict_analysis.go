package retrospector

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// HotFile represents a file that appears in 3+ concurrent PRs,
// indicating a high merge conflict risk.
type HotFile struct {
	Path     string
	PRs      []int
	Count    int
	EpicRefs []string
}

// EpicCollision represents a pair of epics whose PRs routinely conflict.
type EpicCollision struct {
	EpicA         string
	EpicB         string
	SharedFiles   []string
	ConflictCount int
}

// RebaseChurnEntry represents a PR with excessive rebasing (3+),
// along with the traced root cause.
type RebaseChurnEntry struct {
	PR            int
	RebaseCount   int
	RootCause     RebaseRootCause
	ConcurrentPRs []int
	ConflictFiles []string
	BranchAge     time.Duration
}

// RebaseRootCause categorizes why a PR needed excessive rebasing.
type RebaseRootCause string

const (
	RootCauseConcurrentPRs RebaseRootCause = "concurrent_prs_same_files"
	RootCauseLongLived     RebaseRootCause = "long_lived_branch"
	RootCauseDependency    RebaseRootCause = "dependency_chain"
	RootCauseUnknown       RebaseRootCause = "unknown"
)

// DispatchSafetyScore rates how safe it is to dispatch parallel workers
// for a given set of file paths. Score ranges from 0.0 (dangerous) to 1.0 (safe).
type DispatchSafetyScore struct {
	Score    float64
	HotFiles []HotFile
	Rating   string
}

const (
	hotFileThreshold     = 3
	longLivedBranchHours = 48
)

// ConflictAnalyzer performs merge conflict rate analysis on accumulated findings.
type ConflictAnalyzer struct {
	findings []Finding
}

// NewConflictAnalyzer creates an analyzer from a set of findings.
func NewConflictAnalyzer(findings []Finding) *ConflictAnalyzer {
	return &ConflictAnalyzer{findings: findings}
}

// DetectHotFiles identifies files appearing in 3+ concurrent PRs.
// Two PRs are considered concurrent if their time windows overlap
// (one was created before the other was merged).
func (ca *ConflictAnalyzer) DetectHotFiles() []HotFile {
	// Build a map of file → list of PRs that changed it
	filePRs := map[string][]int{}
	fileEpics := map[string]map[string]bool{}

	for _, f := range ca.findings {
		for _, path := range f.FileList {
			filePRs[path] = append(filePRs[path], f.PR)
			if f.EpicRef != "" {
				if fileEpics[path] == nil {
					fileEpics[path] = map[string]bool{}
				}
				fileEpics[path][f.EpicRef] = true
			}
		}
	}

	// Filter to files that appear in concurrent PRs
	var hotFiles []HotFile
	for path, prs := range filePRs {
		concurrentPRs := ca.findConcurrentPRsForFile(path, prs)
		if len(concurrentPRs) >= hotFileThreshold {
			var epics []string
			for e := range fileEpics[path] {
				epics = append(epics, e)
			}
			sort.Strings(epics)

			hotFiles = append(hotFiles, HotFile{
				Path:     path,
				PRs:      concurrentPRs,
				Count:    len(concurrentPRs),
				EpicRefs: epics,
			})
		}
	}

	// Sort by count descending
	sort.Slice(hotFiles, func(i, j int) bool {
		return hotFiles[i].Count > hotFiles[j].Count
	})

	return hotFiles
}

// findConcurrentPRsForFile checks which PRs touching a file had overlapping time windows.
func (ca *ConflictAnalyzer) findConcurrentPRsForFile(path string, prNumbers []int) []int {
	// Build a lookup of PR number → finding
	prMap := map[int]Finding{}
	for _, f := range ca.findings {
		prMap[f.PR] = f
	}

	// A PR is "concurrent" if it overlaps with at least one other PR touching this file.
	// Overlap: PR A was created before PR B was merged AND PR B was created before PR A was merged.
	concurrent := map[int]bool{}
	for i, prA := range prNumbers {
		for j, prB := range prNumbers {
			if i >= j {
				continue
			}
			fA, okA := prMap[prA]
			fB, okB := prMap[prB]
			if !okA || !okB {
				continue
			}

			if ca.prsOverlap(fA, fB) {
				concurrent[prA] = true
				concurrent[prB] = true
			}
		}
	}

	var result []int
	for pr := range concurrent {
		result = append(result, pr)
	}
	sort.Ints(result)
	return result
}

// prsOverlap checks if two PRs had overlapping life spans.
// Falls back to timestamp proximity (within 24h) if created_at/merged_at are unavailable.
func (ca *ConflictAnalyzer) prsOverlap(a, b Finding) bool {
	// If we have full lifecycle data, use precise overlap check
	if !a.CreatedAt.IsZero() && !a.MergedAt.IsZero() &&
		!b.CreatedAt.IsZero() && !b.MergedAt.IsZero() {
		return a.CreatedAt.Before(b.MergedAt) && b.CreatedAt.Before(a.MergedAt)
	}

	// Fallback: consider PRs "concurrent" if they were merged within 24h of each other
	diff := a.Timestamp.Sub(b.Timestamp)
	if diff < 0 {
		diff = -diff
	}
	return diff < 24*time.Hour
}

// DetectEpicCollisions identifies pairs of epics whose PRs routinely conflict
// by sharing files.
func (ca *ConflictAnalyzer) DetectEpicCollisions() []EpicCollision {
	// Map: epicPair → shared files
	type epicPair struct {
		a, b string
	}
	pairFiles := map[epicPair]map[string]bool{}
	pairCount := map[epicPair]int{}

	// Build a map of file → epics
	fileEpics := map[string]map[string]bool{}
	for _, f := range ca.findings {
		if f.EpicRef == "" {
			continue
		}
		for _, path := range f.FileList {
			if fileEpics[path] == nil {
				fileEpics[path] = map[string]bool{}
			}
			fileEpics[path][f.EpicRef] = true
		}
	}

	// For each file touched by 2+ epics, record the collision
	for path, epics := range fileEpics {
		epicList := make([]string, 0, len(epics))
		for e := range epics {
			epicList = append(epicList, e)
		}
		sort.Strings(epicList)

		for i := 0; i < len(epicList); i++ {
			for j := i + 1; j < len(epicList); j++ {
				pair := epicPair{epicList[i], epicList[j]}
				if pairFiles[pair] == nil {
					pairFiles[pair] = map[string]bool{}
				}
				pairFiles[pair][path] = true
				pairCount[pair]++
			}
		}
	}

	var collisions []EpicCollision
	for pair, files := range pairFiles {
		var fileList []string
		for f := range files {
			fileList = append(fileList, f)
		}
		sort.Strings(fileList)

		collisions = append(collisions, EpicCollision{
			EpicA:         pair.a,
			EpicB:         pair.b,
			SharedFiles:   fileList,
			ConflictCount: pairCount[pair],
		})
	}

	// Sort by conflict count descending
	sort.Slice(collisions, func(i, j int) bool {
		return collisions[i].ConflictCount > collisions[j].ConflictCount
	})

	return collisions
}

// CalculateDispatchSafety scores how safe it is to dispatch parallel workers
// based on hot file analysis. Score: 1.0 = safe, 0.0 = dangerous.
func (ca *ConflictAnalyzer) CalculateDispatchSafety() DispatchSafetyScore {
	hotFiles := ca.DetectHotFiles()
	if len(hotFiles) == 0 {
		return DispatchSafetyScore{
			Score:  1.0,
			Rating: "safe",
		}
	}

	// Deduct points based on hot file severity
	score := 1.0
	for _, hf := range hotFiles {
		penalty := float64(hf.Count-hotFileThreshold+1) * 0.1
		if penalty > 0.3 {
			penalty = 0.3
		}
		score -= penalty
	}
	if score < 0 {
		score = 0
	}

	rating := "safe"
	switch {
	case score < 0.3:
		rating = "dangerous"
	case score < 0.6:
		rating = "caution"
	case score < 0.8:
		rating = "moderate"
	}

	return DispatchSafetyScore{
		Score:    score,
		HotFiles: hotFiles,
		Rating:   rating,
	}
}

// AnalyzeRebaseChurn identifies PRs with 3+ rebases and traces root causes.
func (ca *ConflictAnalyzer) AnalyzeRebaseChurn() []RebaseChurnEntry {
	var entries []RebaseChurnEntry

	for _, f := range ca.findings {
		if f.RebaseCount < 3 {
			continue
		}

		entry := RebaseChurnEntry{
			PR:            f.PR,
			RebaseCount:   f.RebaseCount,
			ConflictFiles: f.ConflictFiles,
		}

		// Trace root cause
		entry.RootCause, entry.ConcurrentPRs = ca.traceRebaseRootCause(f)

		// Calculate branch age if lifecycle data available
		if !f.CreatedAt.IsZero() && !f.MergedAt.IsZero() {
			entry.BranchAge = f.MergedAt.Sub(f.CreatedAt)
		}

		entries = append(entries, entry)
	}

	// Sort by rebase count descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RebaseCount > entries[j].RebaseCount
	})

	return entries
}

// traceRebaseRootCause determines why a PR needed excessive rebasing.
func (ca *ConflictAnalyzer) traceRebaseRootCause(target Finding) (RebaseRootCause, []int) {
	// Check for concurrent PRs touching the same files
	var concurrentPRs []int
	for _, other := range ca.findings {
		if other.PR == target.PR {
			continue
		}
		if !ca.prsOverlap(target, other) {
			continue
		}
		// Check for shared files
		if ca.sharesFiles(target, other) {
			concurrentPRs = append(concurrentPRs, other.PR)
		}
	}

	if len(concurrentPRs) > 0 {
		sort.Ints(concurrentPRs)
		return RootCauseConcurrentPRs, concurrentPRs
	}

	// Check for long-lived branch
	if !target.CreatedAt.IsZero() && !target.MergedAt.IsZero() {
		age := target.MergedAt.Sub(target.CreatedAt)
		if age > longLivedBranchHours*time.Hour {
			return RootCauseLongLived, nil
		}
	}

	return RootCauseUnknown, nil
}

// sharesFiles checks if two findings have overlapping file lists.
func (ca *ConflictAnalyzer) sharesFiles(a, b Finding) bool {
	if len(a.FileList) == 0 || len(b.FileList) == 0 {
		return false
	}
	bSet := map[string]bool{}
	for _, f := range b.FileList {
		bSet[f] = true
	}
	for _, f := range a.FileList {
		if bSet[f] {
			return true
		}
	}
	return false
}

// GenerateConflictReport produces a structured report of all conflict analysis.
type ConflictReport struct {
	HotFiles        []HotFile
	EpicCollisions  []EpicCollision
	DispatchSafety  DispatchSafetyScore
	RebaseChurn     []RebaseChurnEntry
	Recommendations []string
}

// Analyze runs the full conflict analysis and produces a report with recommendations.
func (ca *ConflictAnalyzer) Analyze() ConflictReport {
	report := ConflictReport{
		HotFiles:       ca.DetectHotFiles(),
		EpicCollisions: ca.DetectEpicCollisions(),
		DispatchSafety: ca.CalculateDispatchSafety(),
		RebaseChurn:    ca.AnalyzeRebaseChurn(),
	}

	// Generate recommendations
	for _, hf := range report.HotFiles {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Hot file %s touched by %d concurrent PRs (%s); sequence PRs modifying this file",
				hf.Path, hf.Count, formatPRList(hf.PRs)))
	}

	for _, ec := range report.EpicCollisions {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Epics %s and %s share %d files; consider sequencing these epics or splitting shared code",
				ec.EpicA, ec.EpicB, len(ec.SharedFiles)))
	}

	for _, rc := range report.RebaseChurn {
		var mitigation string
		switch rc.RootCause {
		case RootCauseConcurrentPRs:
			mitigation = fmt.Sprintf("sequence with PRs %s", formatPRList(rc.ConcurrentPRs))
		case RootCauseLongLived:
			mitigation = "break into smaller PRs to reduce branch lifetime"
		default:
			mitigation = "investigate concurrent work patterns"
		}
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("PR #%d required %d rebases (%s); %s",
				rc.PR, rc.RebaseCount, rc.RootCause, mitigation))
	}

	return report
}

func formatPRList(prs []int) string {
	parts := make([]string, len(prs))
	for i, pr := range prs {
		parts[i] = fmt.Sprintf("#%d", pr)
	}
	return strings.Join(parts, ", ")
}
