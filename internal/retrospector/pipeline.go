package retrospector

import (
	"fmt"
	"sort"
	"time"
)

const maxRecommendationsPerBatch = 3

// PatternType identifies the category of detected pattern.
type PatternType string

const (
	PatternCIFailure       PatternType = "ci_failure"
	PatternMergeConflict   PatternType = "merge_conflict"
	PatternACMismatch      PatternType = "ac_mismatch"
	PatternExcessiveRebase PatternType = "excessive_rebase"
)

// DetectedPattern represents a cross-PR pattern found during batch analysis.
type DetectedPattern struct {
	Type       PatternType
	Summary    string
	DataPoints int
	Confidence Confidence
	PRNumbers  []int
	Evidence   []Finding
}

// Pipeline aggregates per-PR findings and produces ranked recommendations.
type Pipeline struct {
	boardWriter *BoardWriter
	killSwitch  *KillSwitch
}

// NewPipeline creates a recommendation pipeline with the given board writer
// and kill switch.
func NewPipeline(bw *BoardWriter, ks *KillSwitch) *Pipeline {
	return &Pipeline{
		boardWriter: bw,
		killSwitch:  ks,
	}
}

// DetectPatterns analyzes a set of findings for cross-PR patterns.
func DetectPatterns(findings []Finding) []DetectedPattern {
	var patterns []DetectedPattern

	if p := detectCIFailurePattern(findings); p != nil {
		patterns = append(patterns, *p)
	}
	if p := detectMergeConflictPattern(findings); p != nil {
		patterns = append(patterns, *p)
	}
	if p := detectACMismatchPattern(findings); p != nil {
		patterns = append(patterns, *p)
	}
	if p := detectExcessiveRebasePattern(findings); p != nil {
		patterns = append(patterns, *p)
	}

	// Sort by confidence (High first) then by data points (descending)
	sort.Slice(patterns, func(i, j int) bool {
		ci := confidenceRank(patterns[i].Confidence)
		cj := confidenceRank(patterns[j].Confidence)
		if ci != cj {
			return ci > cj
		}
		return patterns[i].DataPoints > patterns[j].DataPoints
	})

	return patterns
}

// ProcessBatch runs batch analysis on findings, detects patterns,
// and files up to maxRecommendationsPerBatch recommendations to BOARD.md.
// Returns the recommendations that were filed.
func (p *Pipeline) ProcessBatch(findings []Finding) ([]Recommendation, error) {
	if p.killSwitch.IsReadOnly() {
		return nil, nil
	}

	patterns := DetectPatterns(findings)
	if len(patterns) == 0 {
		return nil, nil
	}

	// Rate-limit: max 3 per batch
	limit := maxRecommendationsPerBatch
	if len(patterns) < limit {
		limit = len(patterns)
	}

	var recs []Recommendation
	for i := 0; i < limit; i++ {
		pattern := patterns[i]

		id, err := p.boardWriter.NextID()
		if err != nil {
			return recs, fmt.Errorf("get next ID: %w", err)
		}

		rec := Recommendation{
			ID:         id,
			Text:       pattern.Summary,
			Date:       time.Now().UTC(),
			Source:     "retrospector",
			Confidence: pattern.Confidence,
			Evidence:   prNumberStrings(pattern.PRNumbers),
			Link:       formatEvidenceLink(pattern.PRNumbers),
			Awaiting:   "Supervisor review",
		}

		if err := p.boardWriter.AppendRecommendation(rec); err != nil {
			return recs, fmt.Errorf("append recommendation %s: %w", id, err)
		}
		recs = append(recs, rec)
	}

	return recs, nil
}

func detectCIFailurePattern(findings []Finding) *DetectedPattern {
	count, matched := CountEvidenceForPattern(findings, func(f Finding) bool {
		return !f.CIFirstPass
	})
	if count < 2 {
		return nil
	}

	// Count failure category frequencies
	failCounts := map[string]int{}
	for _, f := range matched {
		for _, cat := range f.CIFailures {
			failCounts[cat]++
		}
	}
	topCategory := ""
	topCount := 0
	for cat, c := range failCounts {
		if c > topCount {
			topCategory = cat
			topCount = c
		}
	}

	summary := fmt.Sprintf("CI failures detected in %d/%d PRs", count, len(findings))
	if topCategory != "" {
		summary += fmt.Sprintf("; most common: %s (%d occurrences)", topCategory, topCount)
	}

	return &DetectedPattern{
		Type:       PatternCIFailure,
		Summary:    summary,
		DataPoints: count,
		Confidence: ScoreConfidence(count),
		PRNumbers:  extractPRNumbers(matched),
		Evidence:   matched,
	}
}

func detectMergeConflictPattern(findings []Finding) *DetectedPattern {
	count, matched := CountEvidenceForPattern(findings, func(f Finding) bool {
		return f.Conflicts > 0
	})
	if count < 2 {
		return nil
	}

	totalConflicts := 0
	for _, f := range matched {
		totalConflicts += f.Conflicts
	}

	summary := fmt.Sprintf("Merge conflicts in %d/%d PRs (%d total conflicting files); consider PR sequencing or smaller PRs",
		count, len(findings), totalConflicts)

	return &DetectedPattern{
		Type:       PatternMergeConflict,
		Summary:    summary,
		DataPoints: count,
		Confidence: ScoreConfidence(count),
		PRNumbers:  extractPRNumbers(matched),
		Evidence:   matched,
	}
}

func detectACMismatchPattern(findings []Finding) *DetectedPattern {
	count, matched := CountEvidenceForPattern(findings, func(f Finding) bool {
		return f.ACMatch == ACMatchPartial || f.ACMatch == ACMatchNone
	})
	if count < 2 {
		return nil
	}

	summary := fmt.Sprintf("AC mismatch in %d/%d PRs; story task lists may need more specificity", count, len(findings))

	return &DetectedPattern{
		Type:       PatternACMismatch,
		Summary:    summary,
		DataPoints: count,
		Confidence: ScoreConfidence(count),
		PRNumbers:  extractPRNumbers(matched),
		Evidence:   matched,
	}
}

func detectExcessiveRebasePattern(findings []Finding) *DetectedPattern {
	count, matched := CountEvidenceForPattern(findings, func(f Finding) bool {
		return f.RebaseCount >= 3
	})
	if count < 2 {
		return nil
	}

	summary := fmt.Sprintf("Excessive rebasing (3+) in %d/%d PRs; consider reducing parallel PR count or improving sequencing",
		count, len(findings))

	return &DetectedPattern{
		Type:       PatternExcessiveRebase,
		Summary:    summary,
		DataPoints: count,
		Confidence: ScoreConfidence(count),
		PRNumbers:  extractPRNumbers(matched),
		Evidence:   matched,
	}
}

func confidenceRank(c Confidence) int {
	switch c {
	case ConfidenceHigh:
		return 3
	case ConfidenceMedium:
		return 2
	case ConfidenceLow:
		return 1
	default:
		return 0
	}
}

func extractPRNumbers(findings []Finding) []int {
	var nums []int
	for _, f := range findings {
		nums = append(nums, f.PR)
	}
	return nums
}

func prNumberStrings(prs []int) []string {
	var s []string
	for _, pr := range prs {
		s = append(s, fmt.Sprintf("#%d", pr))
	}
	return s
}

func formatEvidenceLink(prs []int) string {
	if len(prs) == 0 {
		return "—"
	}
	var parts []string
	for _, pr := range prs {
		parts = append(parts, fmt.Sprintf("PR #%d", pr))
	}
	return fmt.Sprintf("Evidence: %s", joinWithCommaAnd(parts))
}

func joinWithCommaAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return fmt.Sprintf("%s, and %s",
			joinSlice(items[:len(items)-1], ", "),
			items[len(items)-1])
	}
}

func joinSlice(items []string, sep string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}
