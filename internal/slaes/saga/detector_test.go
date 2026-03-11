package saga

import (
	"strings"
	"testing"
	"time"
)

func TestDetector_Analyze_Overlap(t *testing.T) {
	t.Parallel()
	d := NewDetector(3)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	overlap := BranchOverlap{
		Branch: "fix/ci-lint",
		Workers: []WorkerRecord{
			{Name: "w1", Branch: "fix/ci-lint", Timestamp: now},
			{Name: "w2", Branch: "fix/ci-lint", Timestamp: now.Add(1 * time.Hour)},
		},
	}

	alert := d.Analyze(overlap, nil)

	if alert.Type != SagaTypeOverlap {
		t.Errorf("type: got %q, want %q", alert.Type, SagaTypeOverlap)
	}
	if alert.Branch != "fix/ci-lint" {
		t.Errorf("branch: got %q, want %q", alert.Branch, "fix/ci-lint")
	}
	if len(alert.Workers) != 2 {
		t.Errorf("workers: got %d, want 2", len(alert.Workers))
	}
	if alert.FailureRelation != FailureUnknown {
		t.Errorf("failure relation: got %q, want %q", alert.FailureRelation, FailureUnknown)
	}
	if len(alert.Recommendations) == 0 {
		t.Error("expected at least one recommendation")
	}
	if alert.Recommendations[0] != RecommendTargetedFix {
		t.Errorf("recommendation[0]: got %q, want %q", alert.Recommendations[0], RecommendTargetedFix)
	}
}

func TestDetector_Analyze_EscalationTrap(t *testing.T) {
	t.Parallel()
	d := NewDetector(3)
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	overlap := BranchOverlap{
		Branch: "fix/ci",
		Workers: []WorkerRecord{
			{Name: "w1", Branch: "fix/ci", Timestamp: now},
			{Name: "w2", Branch: "fix/ci", Timestamp: now.Add(1 * time.Hour)},
			{Name: "w3", Branch: "fix/ci", Timestamp: now.Add(2 * time.Hour)},
		},
	}

	failures := []CIFailure{
		{WorkerName: "w1", Categories: []string{"lint"}, FilesFixed: nil, FilesBroken: []string{"internal/foo.go"}},
		{WorkerName: "w2", Categories: []string{"lint"}, FilesFixed: []string{"internal/foo.go"}, FilesBroken: []string{"internal/bar.go"}},
		{WorkerName: "w3", Categories: []string{"test"}, FilesFixed: []string{"internal/bar.go"}, FilesBroken: []string{"internal/baz.go"}},
	}

	alert := d.Analyze(overlap, failures)

	if alert.Type != SagaTypeEscalationTrap {
		t.Errorf("type: got %q, want %q", alert.Type, SagaTypeEscalationTrap)
	}
	if !containsRecommendation(alert.Recommendations, RecommendRootCause) {
		t.Error("expected RecommendRootCause in recommendations")
	}
	if !containsRecommendation(alert.Recommendations, RecommendRevert) {
		t.Error("expected RecommendRevert in recommendations")
	}
	if !strings.Contains(alert.Summary, "Escalation trap") {
		t.Errorf("summary should mention escalation trap: %q", alert.Summary)
	}
}

func TestDetector_Analyze_OverlapWithEscalateThreshold(t *testing.T) {
	t.Parallel()
	d := NewDetector(3) // threshold = 3
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	overlap := BranchOverlap{
		Branch: "fix/ci",
		Workers: []WorkerRecord{
			{Name: "w1", Branch: "fix/ci", Timestamp: now},
			{Name: "w2", Branch: "fix/ci", Timestamp: now.Add(30 * time.Minute)},
			{Name: "w3", Branch: "fix/ci", Timestamp: now.Add(1 * time.Hour)},
		},
	}

	// No escalation trap pattern — just independent failures.
	failures := []CIFailure{
		{WorkerName: "w1", Categories: []string{"lint"}},
		{WorkerName: "w2", Categories: []string{"test"}},
		{WorkerName: "w3", Categories: []string{"build"}},
	}

	alert := d.Analyze(overlap, failures)

	if alert.Type != SagaTypeOverlap {
		t.Errorf("type: got %q, want %q", alert.Type, SagaTypeOverlap)
	}
	if !containsRecommendation(alert.Recommendations, RecommendEscalate) {
		t.Error("expected RecommendEscalate when worker count >= threshold")
	}
}

func TestClassifyFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		failures []CIFailure
		want     FailureRelation
	}{
		{
			name:     "no failures",
			failures: nil,
			want:     FailureUnknown,
		},
		{
			name:     "single failure",
			failures: []CIFailure{{WorkerName: "w1", Categories: []string{"lint"}}},
			want:     FailureUnknown,
		},
		{
			name: "related by category",
			failures: []CIFailure{
				{WorkerName: "w1", Categories: []string{"lint"}},
				{WorkerName: "w2", Categories: []string{"lint", "test"}},
			},
			want: FailureRelated,
		},
		{
			name: "related by file",
			failures: []CIFailure{
				{WorkerName: "w1", FilesFixed: []string{"internal/foo.go"}},
				{WorkerName: "w2", FilesBroken: []string{"internal/foo.go"}},
			},
			want: FailureRelated,
		},
		{
			name: "independent",
			failures: []CIFailure{
				{WorkerName: "w1", Categories: []string{"lint"}, FilesFixed: []string{"a.go"}},
				{WorkerName: "w2", Categories: []string{"test"}, FilesBroken: []string{"b.go"}},
			},
			want: FailureIndependent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyFailures(tt.failures)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsEscalationTrap(t *testing.T) {
	t.Parallel()
	d := NewDetector(3)

	tests := []struct {
		name     string
		failures []CIFailure
		want     bool
	}{
		{
			name:     "empty",
			failures: nil,
			want:     false,
		},
		{
			name:     "single failure",
			failures: []CIFailure{{WorkerName: "w1"}},
			want:     false,
		},
		{
			name: "fix-break chain",
			failures: []CIFailure{
				{WorkerName: "w1", FilesBroken: []string{"a.go"}},
				{WorkerName: "w2", FilesFixed: []string{"a.go"}, FilesBroken: []string{"b.go"}},
			},
			want: true,
		},
		{
			name: "no chain — files fixed but nothing broken",
			failures: []CIFailure{
				{WorkerName: "w1", FilesBroken: []string{"a.go"}},
				{WorkerName: "w2", FilesFixed: []string{"a.go"}},
			},
			want: false,
		},
		{
			name: "long chain",
			failures: []CIFailure{
				{WorkerName: "w1", FilesBroken: []string{"a.go"}},
				{WorkerName: "w2", FilesFixed: []string{"a.go"}, FilesBroken: []string{"b.go"}},
				{WorkerName: "w3", FilesFixed: []string{"b.go"}, FilesBroken: []string{"c.go"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := d.isEscalationTrap(tt.failures)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAlert(t *testing.T) {
	t.Parallel()
	alert := SagaAlert{
		Type:   SagaTypeOverlap,
		Branch: "fix/ci",
		Workers: []WorkerRecord{
			{Name: "w1", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
			{Name: "w2", Timestamp: time.Date(2026, 3, 10, 13, 0, 0, 0, time.UTC)},
		},
		FailureRelation: FailureRelated,
		Recommendations: []Recommendation{RecommendTargetedFix},
		Summary:         "test summary",
	}

	msg := FormatAlert(alert)

	if !strings.Contains(msg, "SAGA DETECTED") {
		t.Error("expected 'SAGA DETECTED' in message")
	}
	if !strings.Contains(msg, "fix/ci") {
		t.Error("expected branch name in message")
	}
	if !strings.Contains(msg, "w1") || !strings.Contains(msg, "w2") {
		t.Error("expected worker names in message")
	}
	if !strings.Contains(msg, "test summary") {
		t.Error("expected summary in message")
	}
}

func TestFormatAlert_WithFailureChain(t *testing.T) {
	t.Parallel()
	alert := SagaAlert{
		Type:   SagaTypeEscalationTrap,
		Branch: "fix/ci",
		Workers: []WorkerRecord{
			{Name: "w1", Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)},
		},
		FailureChain: []CIFailure{
			{WorkerName: "w1", Categories: []string{"lint"}, FilesFixed: []string{"a.go"}, FilesBroken: []string{"b.go"}},
		},
		FailureRelation: FailureRelated,
		Recommendations: []Recommendation{RecommendRootCause},
		Summary:         "escalation trap",
	}

	msg := FormatAlert(alert)
	if !strings.Contains(msg, "Failure Chain") {
		t.Error("expected 'Failure Chain' section in message")
	}
	if !strings.Contains(msg, "lint") {
		t.Error("expected failure categories in message")
	}
}

func containsRecommendation(recs []Recommendation, target Recommendation) bool {
	for _, r := range recs {
		if r == target {
			return true
		}
	}
	return false
}
