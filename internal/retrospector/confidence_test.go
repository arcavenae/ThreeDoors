package retrospector

import (
	"testing"
)

func TestScoreConfidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dataPoints int
		want       Confidence
	}{
		{"zero points", 0, ConfidenceLow},
		{"one point", 1, ConfidenceLow},
		{"two points", 2, ConfidenceMedium},
		{"three points", 3, ConfidenceMedium},
		{"four points", 4, ConfidenceMedium},
		{"five points", 5, ConfidenceHigh},
		{"ten points", 10, ConfidenceHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ScoreConfidence(tt.dataPoints)
			if got != tt.want {
				t.Errorf("ScoreConfidence(%d) = %q, want %q", tt.dataPoints, got, tt.want)
			}
		})
	}
}

func TestCountEvidenceForPattern(t *testing.T) {
	t.Parallel()

	findings := []Finding{
		{PR: 100, CIFirstPass: false},
		{PR: 101, CIFirstPass: true},
		{PR: 102, CIFirstPass: false},
		{PR: 103, CIFirstPass: true},
		{PR: 104, CIFirstPass: false},
	}

	count, matched := CountEvidenceForPattern(findings, func(f Finding) bool {
		return !f.CIFirstPass
	})

	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if len(matched) != 3 {
		t.Errorf("matched len = %d, want 3", len(matched))
	}

	// Verify correct PRs matched
	wantPRs := map[int]bool{100: true, 102: true, 104: true}
	for _, f := range matched {
		if !wantPRs[f.PR] {
			t.Errorf("unexpected PR %d in matched", f.PR)
		}
	}
}

func TestCountEvidenceForPatternEmpty(t *testing.T) {
	t.Parallel()

	count, matched := CountEvidenceForPattern(nil, func(f Finding) bool {
		return true
	})
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	if len(matched) != 0 {
		t.Errorf("matched len = %d, want 0", len(matched))
	}
}
