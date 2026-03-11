package retrospector

import (
	"fmt"
	"sort"
	"time"
)

const recurringPatternThreshold = 3

// RecurringCIPattern represents a CI failure category that recurs across
// multiple PRs, with its traced fix layer and proposed action.
type RecurringCIPattern struct {
	Category    FailureCategory
	Occurrences int
	FixLayer    SpecChainLayer
	FixProposal string
	AffectedPRs []int
}

// CIAnalysisResult holds the output of a CI failure deep analysis run.
type CIAnalysisResult struct {
	TotalPRs          int
	TotalFailures     int
	CategoryBreakdown map[FailureCategory]int
	RecurringPatterns []RecurringCIPattern
	HasUnclassified   bool
	UnclassifiedCount int
}

// AnalyzeCIFailures performs deep analysis on a batch of findings,
// classifying CI failures by taxonomy and detecting recurring patterns.
func AnalyzeCIFailures(findings []Finding) CIAnalysisResult {
	result := CIAnalysisResult{
		TotalPRs:          len(findings),
		CategoryBreakdown: make(map[FailureCategory]int),
	}

	// Track which PRs contribute to each category
	categoryPRs := make(map[FailureCategory][]int)

	for _, f := range findings {
		if f.CIFirstPass || len(f.CIFailures) == 0 {
			continue
		}
		result.TotalFailures++

		categories := ClassifyCIFailures(f.CIFailures)
		// Deduplicate categories per PR
		seen := make(map[FailureCategory]bool)
		for _, cat := range categories {
			result.CategoryBreakdown[cat]++
			if !seen[cat] {
				seen[cat] = true
				categoryPRs[cat] = append(categoryPRs[cat], f.PR)
			}
		}
	}

	// Detect recurring patterns (3+ PRs with same category)
	for cat, prs := range categoryPRs {
		if len(prs) >= recurringPatternThreshold {
			result.RecurringPatterns = append(result.RecurringPatterns, RecurringCIPattern{
				Category:    cat,
				Occurrences: len(prs),
				FixLayer:    SpecChainLayerFor(cat),
				FixProposal: FixProposalFor(cat),
				AffectedPRs: prs,
			})
		}
	}

	// Sort recurring patterns by occurrence count descending
	sort.Slice(result.RecurringPatterns, func(i, j int) bool {
		return result.RecurringPatterns[i].Occurrences > result.RecurringPatterns[j].Occurrences
	})

	// Flag unclassified failures
	if count, ok := result.CategoryBreakdown[CategoryUnclassified]; ok && count > 0 {
		result.HasUnclassified = true
		result.UnclassifiedCount = count
	}

	return result
}

// FileCIAnalysisRecommendations converts CI analysis results into BOARD.md
// recommendations, filing one per recurring pattern and one for unclassified
// failures. Respects the kill switch.
func FileCIAnalysisRecommendations(bw *BoardWriter, ks *KillSwitch, analysis CIAnalysisResult) ([]Recommendation, error) {
	if ks.IsReadOnly() {
		return nil, nil
	}

	var recs []Recommendation

	// File recommendations for recurring patterns
	for _, rp := range analysis.RecurringPatterns {
		if len(recs) >= maxRecommendationsPerBatch {
			break
		}

		id, err := bw.NextID()
		if err != nil {
			return recs, fmt.Errorf("get next ID: %w", err)
		}

		text := fmt.Sprintf("Recurring %s CI failures across %d PRs — fix at %s layer: %s",
			rp.Category, rp.Occurrences, rp.FixLayer, rp.FixProposal)

		rec := Recommendation{
			ID:         id,
			Text:       text,
			Date:       time.Now().UTC(),
			Source:     "retrospector",
			Confidence: ScoreConfidence(rp.Occurrences),
			Evidence:   prNumberStrings(rp.AffectedPRs),
			Link:       formatEvidenceLink(rp.AffectedPRs),
			Awaiting:   "Supervisor review",
		}

		if err := bw.AppendRecommendation(rec); err != nil {
			return recs, fmt.Errorf("append recommendation %s: %w", id, err)
		}
		recs = append(recs, rec)
	}

	// File recommendation for unclassified failures if present
	if analysis.HasUnclassified && analysis.UnclassifiedCount >= recurringPatternThreshold && len(recs) < maxRecommendationsPerBatch {
		id, err := bw.NextID()
		if err != nil {
			return recs, fmt.Errorf("get next ID for unclassified: %w", err)
		}

		text := fmt.Sprintf("%d unclassified CI failures detected — requires human review to extend failure taxonomy",
			analysis.UnclassifiedCount)

		rec := Recommendation{
			ID:         id,
			Text:       text,
			Date:       time.Now().UTC(),
			Source:     "retrospector",
			Confidence: ConfidenceLow,
			Evidence:   nil,
			Link:       "—",
			Awaiting:   "Supervisor review",
		}

		if err := bw.AppendRecommendation(rec); err != nil {
			return recs, fmt.Errorf("append unclassified recommendation %s: %w", id, err)
		}
		recs = append(recs, rec)
	}

	return recs, nil
}
