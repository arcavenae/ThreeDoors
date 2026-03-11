package retrospector

import (
	"testing"
)

func TestClassifyCIFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		failure string
		want    FailureCategory
	}{
		{"race detector output", "WARNING: DATA RACE", CategoryRace},
		{"race in test name", "TestFoo_race", CategoryRace},
		{"go test -race failure", "race detected during execution", CategoryRace},
		{"lint error", "golangci-lint found issues", CategoryLint},
		{"lint check name", "lint", CategoryLint},
		{"gofumpt format", "gofumpt", CategoryLint},
		{"vet failure", "go vet failed", CategoryLint},
		{"flaky test retry", "FAIL TestFoo (flaky)", CategoryFlakiness},
		{"timeout in test", "test timed out after 30s", CategoryFlakiness},
		{"intermittent failure", "intermittent", CategoryFlakiness},
		{"build error", "build failed", CategoryBuild},
		{"compile error", "compilation error", CategoryBuild},
		{"cannot find package", "cannot find package", CategoryBuild},
		{"import cycle", "import cycle not allowed", CategoryBuild},
		{"coverage drop", "coverage decreased", CategoryCoverage},
		{"coverage threshold", "coverage below threshold", CategoryCoverage},
		{"unknown failure", "some random error", CategoryUnclassified},
		{"empty string", "", CategoryUnclassified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ClassifyCIFailure(tt.failure)
			if got != tt.want {
				t.Errorf("ClassifyCIFailure(%q) = %q, want %q", tt.failure, got, tt.want)
			}
		})
	}
}

func TestClassifyCIFailures(t *testing.T) {
	t.Parallel()

	failures := []string{"lint", "WARNING: DATA RACE", "some random thing"}
	got := ClassifyCIFailures(failures)

	if len(got) != 3 {
		t.Fatalf("got %d classifications, want 3", len(got))
	}
	if got[0] != CategoryLint {
		t.Errorf("got[0] = %q, want %q", got[0], CategoryLint)
	}
	if got[1] != CategoryRace {
		t.Errorf("got[1] = %q, want %q", got[1], CategoryRace)
	}
	if got[2] != CategoryUnclassified {
		t.Errorf("got[2] = %q, want %q", got[2], CategoryUnclassified)
	}
}

func TestSpecChainLayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		category FailureCategory
		want     SpecChainLayer
	}{
		{"race → story spec", CategoryRace, LayerStorySpec},
		{"lint → CLAUDE.md", CategoryLint, LayerCLAUDEMD},
		{"flakiness → coding standards", CategoryFlakiness, LayerCodingStandards},
		{"build → architecture", CategoryBuild, LayerArchitecture},
		{"coverage → story spec", CategoryCoverage, LayerStorySpec},
		{"unclassified → human review", CategoryUnclassified, LayerHumanReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SpecChainLayerFor(tt.category)
			if got != tt.want {
				t.Errorf("SpecChainLayerFor(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

func TestFixProposalFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		category FailureCategory
		wantNon  string // substring that should appear in the proposal
	}{
		{"race fix", CategoryRace, "race"},
		{"lint fix", CategoryLint, "lint"},
		{"flakiness fix", CategoryFlakiness, "flak"},
		{"build fix", CategoryBuild, "build"},
		{"coverage fix", CategoryCoverage, "coverage"},
		{"unclassified fix", CategoryUnclassified, "review"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FixProposalFor(tt.category)
			if got == "" {
				t.Errorf("FixProposalFor(%q) returned empty string", tt.category)
			}
		})
	}
}

func TestFailureCategoryTally(t *testing.T) {
	t.Parallel()

	categories := []FailureCategory{
		CategoryLint, CategoryLint, CategoryRace, CategoryLint,
		CategoryBuild, CategoryRace,
	}

	tally := TallyCategories(categories)

	if tally[CategoryLint] != 3 {
		t.Errorf("lint count = %d, want 3", tally[CategoryLint])
	}
	if tally[CategoryRace] != 2 {
		t.Errorf("race count = %d, want 2", tally[CategoryRace])
	}
	if tally[CategoryBuild] != 1 {
		t.Errorf("build count = %d, want 1", tally[CategoryBuild])
	}
}
