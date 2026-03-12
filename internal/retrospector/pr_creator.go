package retrospector

import (
	"fmt"
	"strings"
	"time"
)

const (
	// SLAESBranchPrefix is the branch name prefix for retrospector PRs.
	SLAESBranchPrefix = "slaes/"
	// SelfDefinitionPath is the path to the retrospector's own definition.
	SelfDefinitionPath = "agents/retrospector.md"
	// PendingHoursThreshold is how long a high-confidence recommendation must be
	// pending before the retrospector may create a PR for it.
	PendingHoursThreshold = 48
)

// ErrSelfModification is returned when a proposed PR would modify the
// retrospector's own agent definition (Watchmen safeguard #1).
var ErrSelfModification = fmt.Errorf("self-modification blocked: PR would modify %s", SelfDefinitionPath)

// PRProposal represents a proposed PR for a retrospector recommendation.
type PRProposal struct {
	RecommendationID string
	Confidence       Confidence
	Evidence         []string
	Rationale        string
	Branch           string
	Title            string
	Body             string
	FilesChanged     []string
	AgentRestarts    []string
}

// PRCreator creates PRs for high-confidence retrospector recommendations.
type PRCreator struct {
	runner CommandRunner
	repo   string
}

// NewPRCreator creates a PRCreator using real shell commands.
func NewPRCreator(repo string) *PRCreator {
	return &PRCreator{
		runner: &ExecRunner{},
		repo:   repo,
	}
}

// NewPRCreatorWithRunner creates a PRCreator with a custom command runner
// for testing.
func NewPRCreatorWithRunner(runner CommandRunner, repo string) *PRCreator {
	return &PRCreator{
		runner: runner,
		repo:   repo,
	}
}

// ShouldCreatePR determines whether a recommendation qualifies for PR creation.
// Requirements: High confidence + pending for >48 hours without human action.
func ShouldCreatePR(rec Recommendation, now time.Time) bool {
	if rec.Confidence != ConfidenceHigh {
		return false
	}
	pendingSince := rec.Date
	return now.Sub(pendingSince) > PendingHoursThreshold*time.Hour
}

// CheckSelfModification validates that no proposed file changes touch the
// retrospector's own definition. Returns ErrSelfModification if violated.
func CheckSelfModification(files []string) error {
	for _, f := range files {
		if f == SelfDefinitionPath || strings.HasSuffix(f, "/"+SelfDefinitionPath) {
			return ErrSelfModification
		}
	}
	return nil
}

// DetectAgentRestarts returns agent names that need restarting if the given
// files are modified. Agent definitions live in agents/*.md.
func DetectAgentRestarts(files []string) []string {
	var agents []string
	for _, f := range files {
		if !strings.HasPrefix(f, "agents/") || !strings.HasSuffix(f, ".md") {
			continue
		}
		name := strings.TrimPrefix(f, "agents/")
		name = strings.TrimSuffix(name, ".md")
		agents = append(agents, name)
	}
	return agents
}

// BuildProposal constructs a PRProposal from a recommendation and its proposed changes.
func BuildProposal(rec Recommendation, files []string) (PRProposal, error) {
	if err := CheckSelfModification(files); err != nil {
		return PRProposal{}, err
	}

	branch := SLAESBranchPrefix + rec.ID
	restarts := DetectAgentRestarts(files)

	body := formatPRBody(rec, restarts)

	return PRProposal{
		RecommendationID: rec.ID,
		Confidence:       rec.Confidence,
		Evidence:         rec.Evidence,
		Rationale:        rec.Text,
		Branch:           branch,
		Title:            fmt.Sprintf("slaes: %s — %s", rec.ID, truncate(rec.Text, 60)),
		Body:             body,
		FilesChanged:     files,
		AgentRestarts:    restarts,
	}, nil
}

// CreatePR executes the PR creation workflow: create branch, commit changes, push, open PR.
// Returns the PR URL on success.
func (pc *PRCreator) CreatePR(proposal PRProposal) (string, error) {
	// Create and checkout the branch
	if _, err := pc.runner.Run("git", "checkout", "-b", proposal.Branch); err != nil {
		return "", fmt.Errorf("create branch %s: %w", proposal.Branch, err)
	}

	// Stage changed files
	args := append([]string{"add"}, proposal.FilesChanged...)
	if _, err := pc.runner.Run("git", args...); err != nil {
		return "", fmt.Errorf("stage files: %w", err)
	}

	// Commit
	commitMsg := fmt.Sprintf("slaes: %s\n\nRecommendation: %s\nConfidence: %s",
		proposal.RecommendationID, proposal.Rationale, proposal.Confidence)
	if _, err := pc.runner.Run("git", "commit", "-S", "-m", commitMsg); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	// Push
	if _, err := pc.runner.Run("git", "push", "-u", "origin", proposal.Branch); err != nil {
		return "", fmt.Errorf("push branch %s: %w", proposal.Branch, err)
	}

	// Create PR
	out, err := pc.runner.Run("gh", "pr", "create",
		"--title", proposal.Title,
		"--body", proposal.Body,
	)
	if err != nil {
		return "", fmt.Errorf("create PR: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func formatPRBody(rec Recommendation, restarts []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## SLAES Recommendation\n\n")
	fmt.Fprintf(&b, "- **Recommendation ID:** %s\n", rec.ID)
	fmt.Fprintf(&b, "- **Confidence:** %s\n", rec.Confidence)
	fmt.Fprintf(&b, "- **Source:** %s\n", rec.Source)
	fmt.Fprintf(&b, "- **Date:** %s\n", rec.Date.Format(time.DateOnly))
	fmt.Fprintf(&b, "\n## Rationale\n\n%s\n", rec.Text)

	if len(rec.Evidence) > 0 {
		fmt.Fprintf(&b, "\n## Evidence\n\n")
		for _, e := range rec.Evidence {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}

	if len(restarts) > 0 {
		fmt.Fprintf(&b, "\n## Agent Restarts Required\n\n")
		fmt.Fprintf(&b, "After merge, the following agents need restart:\n\n")
		for _, a := range restarts {
			fmt.Fprintf(&b, "- `%s`\n", a)
		}
	}

	fmt.Fprintf(&b, "\n---\n\n*This PR was automatically created by the SLAES retrospector. It must NOT be auto-merged — merge-queue handles merge after human review.*\n")

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
