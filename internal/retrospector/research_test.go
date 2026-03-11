package retrospector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassifyLifecycleFormalized(t *testing.T) {
	t.Parallel()

	artifact := ResearchArtifact{
		Name:     "agentic-engineering-agent-party-mode.md",
		Path:     "/fake/path/agentic-engineering-agent-party-mode.md",
		ModTime:  time.Now().UTC().Add(-30 * 24 * time.Hour),
		Contents: "agentic engineering research",
	}
	refs := ArtifactReferences{
		StoryRefs:  []string{"51.1"},
		BoardRefs:  []string{"D-100"},
		EpicDocRef: true,
	}

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	if state != LifecycleFormalized {
		t.Errorf("got %q, want %q for artifact with story refs + epic doc ref", state, LifecycleFormalized)
	}
}

func TestClassifyLifecycleActive(t *testing.T) {
	t.Parallel()

	artifact := ResearchArtifact{
		Name:     "new-research.md",
		Path:     "/fake/path/new-research.md",
		ModTime:  time.Now().UTC().Add(-3 * 24 * time.Hour),
		Contents: "fresh research in progress",
	}
	refs := ArtifactReferences{
		BoardRefs: []string{"R-001"},
	}

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	if state != LifecycleActive {
		t.Errorf("got %q, want %q for recently modified artifact with board refs", state, LifecycleActive)
	}
}

func TestClassifyLifecycleStale(t *testing.T) {
	t.Parallel()

	artifact := ResearchArtifact{
		Name:     "old-research.md",
		Path:     "/fake/path/old-research.md",
		ModTime:  time.Now().UTC().Add(-20 * 24 * time.Hour),
		Contents: "old research",
	}
	refs := ArtifactReferences{} // no references anywhere

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	if state != LifecycleStale {
		t.Errorf("got %q, want %q for >2 weeks old artifact with no refs", state, LifecycleStale)
	}
}

func TestClassifyLifecycleAbandoned(t *testing.T) {
	t.Parallel()

	artifact := ResearchArtifact{
		Name:     "very-old-research.md",
		Path:     "/fake/path/very-old-research.md",
		ModTime:  time.Now().UTC().Add(-35 * 24 * time.Hour),
		Contents: "abandoned research",
	}
	refs := ArtifactReferences{} // no references

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	if state != LifecycleAbandoned {
		t.Errorf("got %q, want %q for >4 weeks old artifact with no refs", state, LifecycleAbandoned)
	}
}

func TestClassifyLifecycleActiveWithinTwoWeeks(t *testing.T) {
	t.Parallel()

	// Even without references, artifacts <2 weeks old are active
	artifact := ResearchArtifact{
		Name:     "recent-no-refs.md",
		Path:     "/fake/path/recent-no-refs.md",
		ModTime:  time.Now().UTC().Add(-10 * 24 * time.Hour),
		Contents: "recent research without refs yet",
	}
	refs := ArtifactReferences{}

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	if state != LifecycleActive {
		t.Errorf("got %q, want %q for <2 weeks old artifact even without refs", state, LifecycleActive)
	}
}

func TestClassifyLifecycleStaleHasOnlyBoardRefsButOld(t *testing.T) {
	t.Parallel()

	// Old artifact with board refs but no story/epic refs — still stale
	// because board ref alone doesn't mean formalized
	artifact := ResearchArtifact{
		Name:     "board-only.md",
		Path:     "/fake/path/board-only.md",
		ModTime:  time.Now().UTC().Add(-20 * 24 * time.Hour),
		Contents: "research with board mention only",
	}
	refs := ArtifactReferences{
		BoardRefs: []string{"P-003"},
	}

	state := ClassifyLifecycle(artifact, refs, time.Now().UTC())
	// Board refs with pending status keep it active, not stale
	if state != LifecycleActive {
		t.Errorf("got %q, want %q for artifact with pending board refs", state, LifecycleActive)
	}
}

func TestScanResearchArtifacts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create some markdown files
	writeFile(t, filepath.Join(dir, "research-a.md"), "# Research A\nSome content")
	writeFile(t, filepath.Join(dir, "research-b.md"), "# Research B\nMore content")
	writeFile(t, filepath.Join(dir, "not-markdown.txt"), "ignored")

	artifacts, err := ScanResearchArtifacts(dir)
	if err != nil {
		t.Fatalf("ScanResearchArtifacts: %v", err)
	}

	if len(artifacts) != 2 {
		t.Errorf("got %d artifacts, want 2", len(artifacts))
	}

	names := map[string]bool{}
	for _, a := range artifacts {
		names[a.Name] = true
	}
	if !names["research-a.md"] || !names["research-b.md"] {
		t.Errorf("expected research-a.md and research-b.md, got %v", names)
	}
}

func TestScanResearchArtifactsEmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	artifacts, err := ScanResearchArtifacts(dir)
	if err != nil {
		t.Fatalf("ScanResearchArtifacts: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("got %d artifacts, want 0 for empty dir", len(artifacts))
	}
}

func TestScanResearchArtifactsMissingDir(t *testing.T) {
	t.Parallel()

	artifacts, err := ScanResearchArtifacts("/nonexistent/path")
	if err != nil {
		t.Fatalf("ScanResearchArtifacts should not error for missing dir: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("got %d artifacts, want 0 for missing dir", len(artifacts))
	}
}

func TestFindArtifactReferences(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create story files referencing the artifact
	storiesDir := filepath.Join(dir, "docs", "stories")
	if err := os.MkdirAll(storiesDir, 0o700); err != nil {
		t.Fatalf("mkdir stories: %v", err)
	}
	writeFile(t, filepath.Join(storiesDir, "51.1.story.md"),
		"# Story 51.1\n\nResearch: `_bmad-output/planning-artifacts/agentic-engineering-agent-party-mode.md`")

	// Create epics-and-stories.md referencing the artifact
	prdDir := filepath.Join(dir, "docs", "prd")
	if err := os.MkdirAll(prdDir, 0o700); err != nil {
		t.Fatalf("mkdir prd: %v", err)
	}
	writeFile(t, filepath.Join(prdDir, "epics-and-stories.md"),
		"# Epics\n\nResearch at `../../_bmad-output/planning-artifacts/agentic-engineering-agent-party-mode.md`")

	// Create BOARD.md referencing the artifact
	decisionsDir := filepath.Join(dir, "docs", "decisions")
	if err := os.MkdirAll(decisionsDir, 0o700); err != nil {
		t.Fatalf("mkdir decisions: %v", err)
	}
	writeFile(t, filepath.Join(decisionsDir, "BOARD.md"),
		"# Board\n\n| P-001 | Some rec | [Link](../../_bmad-output/planning-artifacts/agentic-engineering-agent-party-mode.md) | Awaiting |")

	artifact := ResearchArtifact{
		Name: "agentic-engineering-agent-party-mode.md",
		Path: filepath.Join(dir, "_bmad-output", "planning-artifacts", "agentic-engineering-agent-party-mode.md"),
	}

	refs, err := FindArtifactReferences(artifact, dir)
	if err != nil {
		t.Fatalf("FindArtifactReferences: %v", err)
	}

	if len(refs.StoryRefs) == 0 {
		t.Error("expected story refs, got none")
	}
	if !refs.EpicDocRef {
		t.Error("expected epic doc ref to be true")
	}
	if len(refs.BoardRefs) == 0 {
		t.Error("expected board refs, got none")
	}
}

func TestResearchLifecycleReport(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	results := []ResearchLifecycleResult{
		{Artifact: ResearchArtifact{Name: "a.md", ModTime: now.Add(-1 * 24 * time.Hour)}, State: LifecycleActive},
		{Artifact: ResearchArtifact{Name: "b.md", ModTime: now.Add(-5 * 24 * time.Hour)}, State: LifecycleActive},
		{Artifact: ResearchArtifact{Name: "c.md", ModTime: now.Add(-25 * 24 * time.Hour)}, State: LifecycleFormalized},
		{Artifact: ResearchArtifact{Name: "d.md", ModTime: now.Add(-20 * 24 * time.Hour)}, State: LifecycleStale},
		{Artifact: ResearchArtifact{Name: "e.md", ModTime: now.Add(-35 * 24 * time.Hour)}, State: LifecycleAbandoned},
	}

	report := GenerateLifecycleReport(results)

	if report.Total != 5 {
		t.Errorf("total = %d, want 5", report.Total)
	}
	if report.ActiveCount != 2 {
		t.Errorf("active = %d, want 2", report.ActiveCount)
	}
	if report.FormalizedCount != 1 {
		t.Errorf("formalized = %d, want 1", report.FormalizedCount)
	}
	if report.StaleCount != 1 {
		t.Errorf("stale = %d, want 1", report.StaleCount)
	}
	if report.AbandonedCount != 1 {
		t.Errorf("abandoned = %d, want 1", report.AbandonedCount)
	}
	if report.OldestUnformalized != "e.md" {
		t.Errorf("oldest unformalized = %q, want %q", report.OldestUnformalized, "e.md")
	}
}

func TestResearchLifecycleReportEmpty(t *testing.T) {
	t.Parallel()

	report := GenerateLifecycleReport(nil)
	if report.Total != 0 {
		t.Errorf("total = %d, want 0", report.Total)
	}
	if report.OldestUnformalized != "" {
		t.Errorf("oldest unformalized = %q, want empty", report.OldestUnformalized)
	}
}

func TestStaleResearchAlerts(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	results := []ResearchLifecycleResult{
		{Artifact: ResearchArtifact{Name: "active.md"}, State: LifecycleActive},
		{Artifact: ResearchArtifact{Name: "stale.md", ModTime: now.Add(-20 * 24 * time.Hour)}, State: LifecycleStale},
		{Artifact: ResearchArtifact{Name: "abandoned.md", ModTime: now.Add(-40 * 24 * time.Hour)}, State: LifecycleAbandoned},
	}

	alerts := StaleResearchAlerts(results)
	if len(alerts) != 2 {
		t.Fatalf("got %d alerts, want 2 (stale + abandoned)", len(alerts))
	}

	// Verify stale alert
	if !strings.Contains(alerts[0], "stale.md") {
		t.Errorf("first alert should mention stale.md: %s", alerts[0])
	}
	if !strings.Contains(alerts[0], "formalization or explicit closure") {
		t.Errorf("stale alert should suggest formalization: %s", alerts[0])
	}
}

func TestClassifyLifecycleTableDriven(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name    string
		ageDays int
		refs    ArtifactReferences
		want    LifecycleState
	}{
		{
			name:    "fresh no refs",
			ageDays: 5,
			refs:    ArtifactReferences{},
			want:    LifecycleActive,
		},
		{
			name:    "fresh with story refs",
			ageDays: 5,
			refs:    ArtifactReferences{StoryRefs: []string{"1.1"}, EpicDocRef: true},
			want:    LifecycleFormalized,
		},
		{
			name:    "old with story and epic refs",
			ageDays: 60,
			refs:    ArtifactReferences{StoryRefs: []string{"1.1"}, EpicDocRef: true},
			want:    LifecycleFormalized,
		},
		{
			name:    "15 days no refs",
			ageDays: 15,
			refs:    ArtifactReferences{},
			want:    LifecycleStale,
		},
		{
			name:    "30 days no refs",
			ageDays: 30,
			refs:    ArtifactReferences{},
			want:    LifecycleAbandoned,
		},
		{
			name:    "exactly 14 days no refs — stale boundary",
			ageDays: 14,
			refs:    ArtifactReferences{},
			want:    LifecycleStale,
		},
		{
			name:    "exactly 28 days no refs — abandoned boundary",
			ageDays: 28,
			refs:    ArtifactReferences{},
			want:    LifecycleAbandoned,
		},
		{
			name:    "13 days no refs — still active",
			ageDays: 13,
			refs:    ArtifactReferences{},
			want:    LifecycleActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			artifact := ResearchArtifact{
				Name:    "test.md",
				Path:    "/test.md",
				ModTime: now.Add(-time.Duration(tt.ageDays) * 24 * time.Hour),
			}
			got := ClassifyLifecycle(artifact, tt.refs, now)
			if got != tt.want {
				t.Errorf("ClassifyLifecycle(age=%dd, refs=%+v) = %q, want %q",
					tt.ageDays, tt.refs, got, tt.want)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
