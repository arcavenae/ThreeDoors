package retrospector

import "strings"

// FailureCategory represents a classified CI failure type.
type FailureCategory string

const (
	CategoryRace         FailureCategory = "race"
	CategoryLint         FailureCategory = "lint"
	CategoryFlakiness    FailureCategory = "flakiness"
	CategoryBuild        FailureCategory = "build"
	CategoryCoverage     FailureCategory = "coverage"
	CategoryUnclassified FailureCategory = "unclassified"
)

// SpecChainLayer represents the layer in the spec chain where a fix should
// be applied.
type SpecChainLayer string

const (
	LayerStorySpec       SpecChainLayer = "story-spec"
	LayerCLAUDEMD        SpecChainLayer = "CLAUDE.md"
	LayerCodingStandards SpecChainLayer = "coding-standards"
	LayerArchitecture    SpecChainLayer = "architecture"
	LayerHumanReview     SpecChainLayer = "human-review"
)

// classificationRule pairs a detection signal with its failure category.
type classificationRule struct {
	signals  []string
	category FailureCategory
}

// classificationRules are evaluated in order; first match wins.
var classificationRules = []classificationRule{
	{
		signals:  []string{"race", "data race"},
		category: CategoryRace,
	},
	{
		signals:  []string{"lint", "golangci-lint", "gofumpt", "go vet"},
		category: CategoryLint,
	},
	{
		signals:  []string{"flaky", "timed out", "timeout", "intermittent"},
		category: CategoryFlakiness,
	},
	{
		signals:  []string{"build failed", "compilation", "cannot find package", "import cycle", "compile"},
		category: CategoryBuild,
	},
	{
		signals:  []string{"coverage decreased", "coverage below", "coverage drop", "coverage regression"},
		category: CategoryCoverage,
	},
}

// ClassifyCIFailure classifies a single CI failure string into a
// FailureCategory based on detection signals.
func ClassifyCIFailure(failure string) FailureCategory {
	lower := strings.ToLower(failure)
	for _, rule := range classificationRules {
		for _, signal := range rule.signals {
			if strings.Contains(lower, signal) {
				return rule.category
			}
		}
	}
	return CategoryUnclassified
}

// ClassifyCIFailures classifies a slice of CI failure strings, returning
// the corresponding categories in the same order.
func ClassifyCIFailures(failures []string) []FailureCategory {
	categories := make([]FailureCategory, len(failures))
	for i, f := range failures {
		categories[i] = ClassifyCIFailure(f)
	}
	return categories
}

// SpecChainLayerFor maps a failure category to the spec-chain layer where
// the fix should be applied.
func SpecChainLayerFor(category FailureCategory) SpecChainLayer {
	switch category {
	case CategoryRace:
		return LayerStorySpec
	case CategoryLint:
		return LayerCLAUDEMD
	case CategoryFlakiness:
		return LayerCodingStandards
	case CategoryBuild:
		return LayerArchitecture
	case CategoryCoverage:
		return LayerStorySpec
	default:
		return LayerHumanReview
	}
}

// FixProposalFor returns a human-readable fix proposal for the given
// failure category, targeting the appropriate spec-chain layer.
func FixProposalFor(category FailureCategory) string {
	switch category {
	case CategoryRace:
		return "Add to story spec: 'MUST pass go test -race locally before PR submission'"
	case CategoryLint:
		return "Strengthen CLAUDE.md pre-commit lint rules: require make lint to pass before committing"
	case CategoryFlakiness:
		return "Add coding standards entry: identify and quarantine flaky test patterns; require deterministic test design"
	case CategoryBuild:
		return "Review architecture for build dependency issues; ensure clean builds in CI environment"
	case CategoryCoverage:
		return "Add to story spec: 'Test coverage must not decrease from baseline'"
	case CategoryUnclassified:
		return "Unclassified CI failure type requires human review to extend taxonomy"
	default:
		return "Unknown failure category requires human review"
	}
}

// TallyCategories counts the occurrences of each failure category.
func TallyCategories(categories []FailureCategory) map[FailureCategory]int {
	tally := make(map[FailureCategory]int)
	for _, c := range categories {
		tally[c]++
	}
	return tally
}
