package retrospector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNextIDFromContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "empty file",
			content: "",
			want:    "P-001",
		},
		{
			name: "existing P-006",
			content: `## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-001 | Rec 1 | 2026-03-04 | source | link | review |
| P-006 | Rec 6 | 2026-03-09 | source | link | review |
`,
			want: "P-007",
		},
		{
			name: "non-sequential IDs",
			content: `| P-003 | something | 2026-03-04 | source | link | review |
| P-010 | something | 2026-03-05 | source | link | review |
| P-001 | something | 2026-03-06 | source | link | review |
`,
			want: "P-011",
		},
		{
			name: "IDs in Decided section too",
			content: `## Pending Recommendations

| P-005 | rec | 2026-03-09 | source | link | review |

## Decided

| P-002 | old | 2026-03-01 | source | link |
`,
			want: "P-006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := nextIDFromContent(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("nextIDFromContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppendRecommendationToContent(t *testing.T) {
	t.Parallel()

	boardContent := `# Knowledge Decisions Board

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-001 | Migrate to Justfile | 2026-03-04 | Research spike | link | Owner sign-off |

## Decided

| ID | Decision | Date | Rationale | Link |
`

	rec := Recommendation{
		ID:         "P-002",
		Text:       "Add CI caching",
		Date:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Source:     "retrospector",
		Confidence: ConfidenceHigh,
		Link:       "Evidence: PR #100",
		Awaiting:   "Supervisor review",
	}

	got, err := appendRecommendationToContent(boardContent, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "| P-002 | Add CI caching | 2026-03-10 | retrospector (High) | Evidence: PR #100 | Supervisor review |") {
		t.Errorf("recommendation row not found in output:\n%s", got)
	}

	// Verify P-001 is still present
	if !strings.Contains(got, "P-001") {
		t.Error("existing P-001 row was removed")
	}

	// Verify Decided section is still present
	if !strings.Contains(got, "## Decided") {
		t.Error("Decided section was removed")
	}
}

func TestAppendRecommendationToEmptyTable(t *testing.T) {
	t.Parallel()

	boardContent := `## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|

## Decided
`

	rec := Recommendation{
		ID:         "P-001",
		Text:       "First rec",
		Date:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Source:     "retrospector",
		Confidence: ConfidenceLow,
		Link:       "—",
		Awaiting:   "Supervisor review",
	}

	got, err := appendRecommendationToContent(boardContent, rec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "P-001") {
		t.Errorf("recommendation not found in output:\n%s", got)
	}
}

func TestAppendRecommendationMissingSection(t *testing.T) {
	t.Parallel()

	boardContent := `# Board

## Decided

| ID | Decision |
`

	rec := Recommendation{ID: "P-001"}
	_, err := appendRecommendationToContent(boardContent, rec)
	if err == nil {
		t.Fatal("expected error for missing section")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBoardWriterIntegration(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	boardPath := filepath.Join(dir, "BOARD.md")

	content := `# Board

## Pending Recommendations

| ID | Recommendation | Date | Source | Link | Awaiting |
|----|----------------|------|--------|------|----------|
| P-001 | Existing rec | 2026-03-04 | source | link | review |

## Decided
`
	if err := os.WriteFile(boardPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write board: %v", err)
	}

	bw := NewBoardWriter(boardPath)

	// Check next ID
	id, err := bw.NextID()
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id != "P-002" {
		t.Errorf("NextID = %q, want P-002", id)
	}

	// Append a recommendation
	rec := Recommendation{
		ID:         id,
		Text:       "New recommendation",
		Date:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Source:     "retrospector",
		Confidence: ConfidenceMedium,
		Link:       "Evidence: PR #200",
		Awaiting:   "Supervisor review",
	}
	if err := bw.AppendRecommendation(rec); err != nil {
		t.Fatalf("AppendRecommendation: %v", err)
	}

	// Verify file contents
	data, err := os.ReadFile(boardPath)
	if err != nil {
		t.Fatalf("read board: %v", err)
	}
	result := string(data)

	if !strings.Contains(result, "P-002") {
		t.Error("P-002 not found in board")
	}
	if !strings.Contains(result, "New recommendation") {
		t.Error("recommendation text not found in board")
	}

	// Verify next ID incremented
	id2, err := bw.NextID()
	if err != nil {
		t.Fatalf("NextID after append: %v", err)
	}
	if id2 != "P-003" {
		t.Errorf("NextID after append = %q, want P-003", id2)
	}
}

func TestFormatRecommendationRow(t *testing.T) {
	t.Parallel()

	rec := Recommendation{
		ID:         "P-007",
		Text:       "Consider caching",
		Date:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Source:     "retrospector",
		Confidence: ConfidenceHigh,
		Link:       "Evidence: PR #100 and PR #200",
		Awaiting:   "Supervisor review",
	}

	got := formatRecommendationRow(rec)
	want := "| P-007 | Consider caching | 2026-03-10 | retrospector (High) | Evidence: PR #100 and PR #200 | Supervisor review |"
	if got != want {
		t.Errorf("formatRecommendationRow:\ngot:  %s\nwant: %s", got, want)
	}
}
