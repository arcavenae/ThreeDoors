package core

import (
	"strings"
)

// typeTokens maps inline tag tokens to TaskType values.
var typeTokens = map[string]TaskType{
	"#creative":       TypeCreative,
	"#administrative": TypeAdministrative,
	"#technical":      TypeTechnical,
	"#physical":       TypePhysical,
}

// effortTokens maps inline tag tokens to TaskEffort values.
var effortTokens = map[string]TaskEffort{
	"@quick-win": EffortQuickWin,
	"@medium":    EffortMedium,
	"@deep-work": EffortDeepWork,
}

// locationTokens maps inline tag tokens to TaskLocation values.
var locationTokens = map[string]TaskLocation{
	"+home":     LocationHome,
	"+work":     LocationWork,
	"+errands":  LocationErrands,
	"+anywhere": LocationAnywhere,
}

// ParseInlineTags extracts categorization tags from task text.
// Tags: #type, @effort, +location (case-insensitive).
// Matched tokens are stripped from text. Unrecognized tokens are left in text.
// Duplicate tags of the same category: last one wins.
func ParseInlineTags(text string) (cleanText string, taskType TaskType, effort TaskEffort, loc TaskLocation) {
	words := strings.Fields(text)
	var remaining []string

	for _, word := range words {
		lower := strings.ToLower(word)

		if tt, ok := typeTokens[lower]; ok {
			taskType = tt
			continue
		}
		if ef, ok := effortTokens[lower]; ok {
			effort = ef
			continue
		}
		if lo, ok := locationTokens[lower]; ok {
			loc = lo
			continue
		}
		remaining = append(remaining, word)
	}

	cleanText = strings.Join(remaining, " ")
	return
}
