package retrospector

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestShouldCreatePR(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		confidence Confidence
		recDate    time.Time
		now        time.Time
		want       bool
	}{
		{
			name:       "high confidence and old enough",
			confidence: ConfidenceHigh,
			recDate:    baseTime.Add(-49 * time.Hour),
			now:        baseTime,
			want:       true,
		},
		{
			name:       "high confidence but too recent",
			confidence: ConfidenceHigh,
			recDate:    baseTime.Add(-24 * time.Hour),
			now:        baseTime,
			want:       false,
		},
		{
			name:       "medium confidence rejected",
			confidence: ConfidenceMedium,
			recDate:    baseTime.Add(-72 * time.Hour),
			now:        baseTime,
			want:       false,
		},
		{
			name:       "low confidence rejected",
			confidence: ConfidenceLow,
			recDate:    baseTime.Add(-72 * time.Hour),
			now:        baseTime,
			want:       false,
		},
		{
			name:       "exactly at threshold",
			confidence: ConfidenceHigh,
			recDate:    baseTime.Add(-48 * time.Hour),
			now:        baseTime,
			want:       false,
		},
		{
			name:       "just past threshold",
			confidence: ConfidenceHigh,
			recDate:    baseTime.Add(-48*time.Hour - time.Minute),
			now:        baseTime,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := Recommendation{
				ID:         "P-001",
				Confidence: tt.confidence,
				Date:       tt.recDate,
			}
			got := ShouldCreatePR(rec, tt.now)
			if got != tt.want {
				t.Errorf("ShouldCreatePR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSelfModification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		files   []string
		wantErr bool
	}{
		{
			name:    "no agent files",
			files:   []string{"CLAUDE.md", "internal/foo/bar.go"},
			wantErr: false,
		},
		{
			name:    "other agent file",
			files:   []string{"agents/merge-queue.md"},
			wantErr: false,
		},
		{
			name:    "self definition direct",
			files:   []string{"agents/retrospector.md"},
			wantErr: true,
		},
		{
			name:    "self definition with prefix",
			files:   []string{"some/path/agents/retrospector.md"},
			wantErr: true,
		},
		{
			name:    "self definition among others",
			files:   []string{"CLAUDE.md", "agents/retrospector.md", "internal/foo.go"},
			wantErr: true,
		},
		{
			name:    "empty file list",
			files:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CheckSelfModification(tt.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSelfModification() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != ErrSelfModification {
				t.Errorf("expected ErrSelfModification, got %v", err)
			}
		})
	}
}

func TestDetectAgentRestarts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		files []string
		want  []string
	}{
		{
			name:  "no agent files",
			files: []string{"CLAUDE.md", "internal/foo.go"},
			want:  nil,
		},
		{
			name:  "single agent file",
			files: []string{"agents/merge-queue.md"},
			want:  []string{"merge-queue"},
		},
		{
			name:  "multiple agent files",
			files: []string{"agents/merge-queue.md", "agents/pr-shepherd.md", "CLAUDE.md"},
			want:  []string{"merge-queue", "pr-shepherd"},
		},
		{
			name:  "non-md files in agents dir ignored",
			files: []string{"agents/README.txt"},
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DetectAgentRestarts(tt.files)
			if len(got) != len(tt.want) {
				t.Fatalf("DetectAgentRestarts() = %v, want %v", got, tt.want)
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("DetectAgentRestarts()[%d] = %q, want %q", i, g, tt.want[i])
				}
			}
		})
	}
}

func TestBuildProposal(t *testing.T) {
	t.Parallel()

	t.Run("successful proposal", func(t *testing.T) {
		t.Parallel()
		rec := Recommendation{
			ID:         "P-007",
			Text:       "Strengthen lint rules",
			Confidence: ConfidenceHigh,
			Date:       time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC),
			Source:     "retrospector",
			Evidence:   []string{"#100", "#101", "#102"},
		}
		files := []string{"CLAUDE.md"}

		proposal, err := BuildProposal(rec, files)
		if err != nil {
			t.Fatalf("BuildProposal() unexpected error: %v", err)
		}

		if proposal.Branch != "slaes/P-007" {
			t.Errorf("Branch = %q, want %q", proposal.Branch, "slaes/P-007")
		}
		if proposal.RecommendationID != "P-007" {
			t.Errorf("RecommendationID = %q, want %q", proposal.RecommendationID, "P-007")
		}
		if !strings.Contains(proposal.Title, "P-007") {
			t.Errorf("Title %q should contain recommendation ID", proposal.Title)
		}
		if !strings.Contains(proposal.Body, "P-007") {
			t.Errorf("Body should contain recommendation ID")
		}
		if !strings.Contains(proposal.Body, "High") {
			t.Errorf("Body should contain confidence score")
		}
		if len(proposal.AgentRestarts) != 0 {
			t.Errorf("AgentRestarts should be empty, got %v", proposal.AgentRestarts)
		}
	})

	t.Run("proposal with agent restarts", func(t *testing.T) {
		t.Parallel()
		rec := Recommendation{
			ID:         "P-008",
			Text:       "Add isolation guardrail to merge-queue",
			Confidence: ConfidenceHigh,
			Source:     "retrospector",
		}
		files := []string{"agents/merge-queue.md"}

		proposal, err := BuildProposal(rec, files)
		if err != nil {
			t.Fatalf("BuildProposal() unexpected error: %v", err)
		}

		if len(proposal.AgentRestarts) != 1 || proposal.AgentRestarts[0] != "merge-queue" {
			t.Errorf("AgentRestarts = %v, want [merge-queue]", proposal.AgentRestarts)
		}
		if !strings.Contains(proposal.Body, "Agent Restarts Required") {
			t.Errorf("Body should note agent restarts")
		}
		if !strings.Contains(proposal.Body, "merge-queue") {
			t.Errorf("Body should list merge-queue for restart")
		}
	})

	t.Run("blocked by self-modification", func(t *testing.T) {
		t.Parallel()
		rec := Recommendation{
			ID:         "P-009",
			Text:       "Improve retrospector definition",
			Confidence: ConfidenceHigh,
			Source:     "retrospector",
		}
		files := []string{"agents/retrospector.md"}

		_, err := BuildProposal(rec, files)
		if err != ErrSelfModification {
			t.Errorf("BuildProposal() should return ErrSelfModification, got %v", err)
		}
	})
}

func TestCreatePR(t *testing.T) {
	t.Parallel()

	t.Run("successful PR creation", func(t *testing.T) {
		t.Parallel()
		runner := newMockRunner()
		// Set up responses for all commands in the workflow
		runner.responses["git checkout -b slaes/P-007"] = nil
		runner.responses["git add CLAUDE.md"] = nil
		runner.responses["git commit -S -m slaes: P-007\n\nRecommendation: Strengthen lint rules\nConfidence: High"] = nil
		runner.responses["git push -u origin slaes/P-007"] = nil
		runner.responses["gh pr create --title slaes: P-007 — Strengthen lint rules --body test body"] = []byte("https://github.com/arcaven/ThreeDoors/pull/500\n")

		pc := NewPRCreatorWithRunner(runner, "arcaven/ThreeDoors")
		proposal := PRProposal{
			RecommendationID: "P-007",
			Confidence:       ConfidenceHigh,
			Rationale:        "Strengthen lint rules",
			Branch:           "slaes/P-007",
			Title:            "slaes: P-007 — Strengthen lint rules",
			Body:             "test body",
			FilesChanged:     []string{"CLAUDE.md"},
		}

		url, err := pc.CreatePR(proposal)
		if err != nil {
			t.Fatalf("CreatePR() unexpected error: %v", err)
		}
		if url != "https://github.com/arcaven/ThreeDoors/pull/500" {
			t.Errorf("URL = %q, want PR URL", url)
		}

		// Verify command sequence
		if len(runner.calls) < 4 {
			t.Fatalf("expected at least 4 commands, got %d", len(runner.calls))
		}
		// Branch creation
		if !strings.HasPrefix(runner.calls[0], "git checkout") {
			t.Errorf("first command should be git checkout, got %s", runner.calls[0])
		}
		// Staging
		if !strings.HasPrefix(runner.calls[1], "git add") {
			t.Errorf("second command should be git add, got %s", runner.calls[1])
		}
		// Commit (must include -S for signing)
		if !strings.Contains(runner.calls[2], "git commit -S") {
			t.Errorf("third command should be git commit -S, got %s", runner.calls[2])
		}
		// Push
		if !strings.HasPrefix(runner.calls[3], "git push") {
			t.Errorf("fourth command should be git push, got %s", runner.calls[3])
		}
	})

	t.Run("branch creation failure", func(t *testing.T) {
		t.Parallel()
		runner := newMockRunner()
		runner.errors["git checkout -b slaes/P-007"] = fmt.Errorf("branch exists")

		pc := NewPRCreatorWithRunner(runner, "arcaven/ThreeDoors")
		_, err := pc.CreatePR(PRProposal{
			Branch:       "slaes/P-007",
			FilesChanged: []string{"CLAUDE.md"},
		})
		if err == nil {
			t.Fatal("expected error for branch creation failure")
		}
		if !strings.Contains(err.Error(), "create branch") {
			t.Errorf("error should mention branch creation: %v", err)
		}
	})
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a longer string that needs truncation", 20, "this is a longer ..."},
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormatPRBody(t *testing.T) {
	t.Parallel()

	rec := Recommendation{
		ID:         "P-010",
		Text:       "Test recommendation",
		Confidence: ConfidenceHigh,
		Source:     "retrospector",
		Date:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		Evidence:   []string{"#400", "#401"},
	}

	t.Run("without agent restarts", func(t *testing.T) {
		t.Parallel()
		body := formatPRBody(rec, nil)
		if !strings.Contains(body, "P-010") {
			t.Error("body should contain recommendation ID")
		}
		if !strings.Contains(body, "High") {
			t.Error("body should contain confidence")
		}
		if !strings.Contains(body, "#400") {
			t.Error("body should contain evidence")
		}
		if strings.Contains(body, "Agent Restarts") {
			t.Error("body should not contain restart section when no restarts")
		}
		if !strings.Contains(body, "must NOT be auto-merged") {
			t.Error("body should contain auto-merge warning")
		}
	})

	t.Run("with agent restarts", func(t *testing.T) {
		t.Parallel()
		body := formatPRBody(rec, []string{"merge-queue", "pr-shepherd"})
		if !strings.Contains(body, "Agent Restarts Required") {
			t.Error("body should contain restart section")
		}
		if !strings.Contains(body, "`merge-queue`") {
			t.Error("body should list merge-queue")
		}
		if !strings.Contains(body, "`pr-shepherd`") {
			t.Error("body should list pr-shepherd")
		}
	})
}
