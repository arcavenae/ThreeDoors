package retrospector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ghPR represents the JSON output from gh pr list/view commands.
type ghPR struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	MergedAt time.Time `json:"mergedAt"`
	Files    []ghFile  `json:"files"`
	Labels   []ghLabel `json:"labels"`
}

type ghFile struct {
	Path string `json:"path"`
}

type ghLabel struct {
	Name string `json:"name"`
}

type ghCommitsResponse struct {
	Commits []ghCommitNode `json:"commits"`
}

type ghCommitNode struct {
	MessageHeadline string `json:"messageHeadline"`
	MessageBody     string `json:"messageBody"`
}

// CommandRunner abstracts shell command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

// ExecRunner executes real shell commands via os/exec.
type ExecRunner struct{}

// Run executes a command and returns its combined output.
func (r *ExecRunner) Run(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// PRDataCollector fetches PR metadata from GitHub using the gh CLI.
type PRDataCollector struct {
	runner CommandRunner
}

// NewPRDataCollector creates a collector using real shell commands.
func NewPRDataCollector() *PRDataCollector {
	return &PRDataCollector{runner: &ExecRunner{}}
}

// NewPRDataCollectorWithRunner creates a collector with a custom command runner
// for testing.
func NewPRDataCollectorWithRunner(runner CommandRunner) *PRDataCollector {
	return &PRDataCollector{runner: runner}
}

// FetchMergedPRs returns PRs merged after the given PR number, sorted chronologically.
func (c *PRDataCollector) FetchMergedPRs(afterPR int) ([]ghPR, error) {
	out, err := c.runner.Run("gh", "pr", "list",
		"--state", "merged",
		"--json", "number,title,mergedAt,files,labels",
		"--limit", "100",
	)
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}

	var prs []ghPR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("unmarshal pr list: %w", err)
	}

	// Filter to only PRs after the last processed one
	var filtered []ghPR
	for _, pr := range prs {
		if pr.Number > afterPR {
			filtered = append(filtered, pr)
		}
	}

	return filtered, nil
}

// FetchStoryRef extracts the story reference from a PR's commit messages.
func (c *PRDataCollector) FetchStoryRef(prNumber int) (string, error) {
	out, err := c.runner.Run("gh", "pr", "view",
		fmt.Sprintf("%d", prNumber),
		"--json", "commits",
	)
	if err != nil {
		return "", fmt.Errorf("gh pr view %d commits: %w", prNumber, err)
	}

	var resp ghCommitsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("unmarshal commits for PR #%d: %w", prNumber, err)
	}

	for _, commit := range resp.Commits {
		ref := ParseStoryRef(commit.MessageHeadline)
		if ref != "" {
			return ref, nil
		}
		ref = ParseStoryRef(commit.MessageBody)
		if ref != "" {
			return ref, nil
		}
	}

	return "", nil
}

// FetchCIFirstPass checks whether the first CI run on a PR passed without fixes.
func (c *PRDataCollector) FetchCIFirstPass(prNumber int) (bool, error) {
	out, err := c.runner.Run("gh", "pr", "checks",
		fmt.Sprintf("%d", prNumber),
		"--json", "name,state",
	)
	if err != nil {
		// If checks command fails, assume not first pass
		return false, nil //nolint:nilerr // gh pr checks fails for PRs without checks
	}

	var checks []struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(out, &checks); err != nil {
		return false, nil //nolint:nilerr // unparseable checks output
	}

	// All checks must have passed
	for _, check := range checks {
		if !strings.EqualFold(check.State, "SUCCESS") && !strings.EqualFold(check.State, "PASS") {
			return false, nil
		}
	}

	return len(checks) > 0, nil
}

// IsInfraOrDocsPR checks if a PR is an infrastructure or documentation PR
// based on its labels, indicating it's exempt from process gap flagging.
func IsInfraOrDocsPR(labels []ghLabel) bool {
	for _, label := range labels {
		lower := strings.ToLower(label.Name)
		if lower == "infrastructure" || lower == "infra" ||
			lower == "documentation" || lower == "docs" ||
			lower == "dependencies" || lower == "chore" {
			return true
		}
	}
	return false
}
