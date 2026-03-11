package retrospector

// ScoreConfidence determines the confidence level based on the number
// of supporting data points. The rules are:
//   - High: 5+ data points across multiple PRs
//   - Medium: 2-4 data points
//   - Low: 1 data point or extrapolation from limited evidence
func ScoreConfidence(dataPoints int) Confidence {
	switch {
	case dataPoints >= 5:
		return ConfidenceHigh
	case dataPoints >= 2:
		return ConfidenceMedium
	default:
		return ConfidenceLow
	}
}

// CountEvidenceForPattern counts findings that match a given predicate,
// returning the count and the matching findings for evidence linking.
func CountEvidenceForPattern(findings []Finding, match func(Finding) bool) (int, []Finding) {
	var matched []Finding
	for _, f := range findings {
		if match(f) {
			matched = append(matched, f)
		}
	}
	return len(matched), matched
}
