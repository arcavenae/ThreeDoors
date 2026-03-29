package dispatch

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const multiclaude = "multiclaude"

// taskSuffix is appended to every worker task description.
const taskSuffix = `
IMPORTANT: Sign all commits (git commit -S). Do NOT add Co-Authored-By for AI.
PRs target arcavenae/ThreeDoors via
gh pr create --repo arcavenae/ThreeDoors --head <branch>.`

// CLIDispatcher implements Dispatcher by wrapping the multiclaude CLI.
type CLIDispatcher struct {
	runner CommandRunner
}

// NewCLIDispatcher creates a CLIDispatcher with the given CommandRunner.
func NewCLIDispatcher(runner CommandRunner) *CLIDispatcher {
	return &CLIDispatcher{runner: runner}
}

// CheckAvailable validates that the multiclaude binary is on PATH.
func (d *CLIDispatcher) CheckAvailable(_ context.Context) error {
	_, err := exec.LookPath(multiclaude)
	if err != nil {
		return fmt.Errorf("dispatch check-available: %w", err)
	}
	return nil
}

// CreateWorker creates a new multiclaude worker with the given task description.
func (d *CLIDispatcher) CreateWorker(ctx context.Context, task string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	out, err := d.runner.Run(ctx, multiclaude, "worker", "create", task)
	if err != nil {
		return "", fmt.Errorf("dispatch create-worker: %w", err)
	}

	name := parseWorkerName(string(out))
	if name == "" {
		return "", fmt.Errorf("dispatch create-worker: could not parse worker name from output: %s", string(out))
	}

	return name, nil
}

// ListWorkers returns all active multiclaude workers.
func (d *CLIDispatcher) ListWorkers(ctx context.Context) ([]WorkerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	out, err := d.runner.Run(ctx, multiclaude, "worker", "list")
	if err != nil {
		return nil, fmt.Errorf("dispatch list-workers: %w", err)
	}

	return parseWorkerList(string(out)), nil
}

// GetHistory returns recent worker history entries.
func (d *CLIDispatcher) GetHistory(ctx context.Context, limit int) ([]HistoryEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	out, err := d.runner.Run(ctx, multiclaude, "repo", "history", "-n", strconv.Itoa(limit))
	if err != nil {
		return nil, fmt.Errorf("dispatch get-history: %w", err)
	}

	return parseHistory(string(out)), nil
}

// RemoveWorker removes the named multiclaude worker.
func (d *CLIDispatcher) RemoveWorker(ctx context.Context, name string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	_, err := d.runner.Run(ctx, multiclaude, "worker", "rm", name)
	if err != nil {
		return fmt.Errorf("dispatch remove-worker: %w", err)
	}

	return nil
}

// BuildTaskDescription constructs a rich worker prompt from a QueueItem.
func BuildTaskDescription(item QueueItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n", item.TaskText)

	if item.Context != "" {
		fmt.Fprintf(&b, "\nContext: %s\n", item.Context)
	}

	if len(item.AcceptanceCriteria) > 0 {
		fmt.Fprintf(&b, "\nAcceptance Criteria:\n")
		for _, ac := range item.AcceptanceCriteria {
			fmt.Fprintf(&b, "- %s\n", ac)
		}
	}

	if item.Scope != "" {
		fmt.Fprintf(&b, "\nScope: %s\n", item.Scope)
	}

	fmt.Fprintf(&b, "%s", taskSuffix)
	return b.String()
}

// parseWorkerName extracts the worker name from multiclaude worker create output.
// Expected output contains a line like "Created worker: <name>" or just the name.
func parseWorkerName(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Created worker:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Created worker:"))
		}
		if strings.HasPrefix(line, "Worker created:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Worker created:"))
		}
	}
	// Fallback: if output is a single non-empty line, treat it as the name.
	trimmed := strings.TrimSpace(output)
	if !strings.Contains(trimmed, "\n") && trimmed != "" {
		return trimmed
	}
	return ""
}

// parseWorkerList parses multiclaude worker list output into WorkerInfo slices.
// Expected format: tabular or line-based output with name, status, branch, task fields.
func parseWorkerList(output string) []WorkerInfo {
	var workers []WorkerInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isHeaderLine(line) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		w := WorkerInfo{
			Name:   fields[0],
			Status: fields[1],
		}
		if len(fields) >= 3 {
			w.Branch = fields[2]
		}
		if len(fields) >= 4 {
			w.Task = strings.Join(fields[3:], " ")
		}
		workers = append(workers, w)
	}

	return workers
}

// parseHistory parses multiclaude repo history output into HistoryEntry slices.
func parseHistory(output string) []HistoryEntry {
	var entries []HistoryEntry
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isHeaderLine(line) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		e := HistoryEntry{
			WorkerName: fields[0],
			Status:     fields[1],
		}

		for i := 2; i < len(fields); i++ {
			f := fields[i]
			if strings.HasPrefix(f, "#") {
				if n, err := strconv.Atoi(strings.TrimPrefix(f, "#")); err == nil {
					e.PRNumber = n
				}
			} else if strings.HasPrefix(f, "https://") || strings.HasPrefix(f, "http://") {
				e.PRURL = f
			} else if t, err := time.Parse(time.RFC3339, f); err == nil {
				t = t.UTC()
				e.CompletedAt = &t
			}
		}

		// Remaining fields that aren't parsed as PR/URL/time form the summary.
		var summaryParts []string
		for i := 2; i < len(fields); i++ {
			f := fields[i]
			if strings.HasPrefix(f, "#") || strings.HasPrefix(f, "https://") || strings.HasPrefix(f, "http://") {
				continue
			}
			if _, err := time.Parse(time.RFC3339, f); err == nil {
				continue
			}
			summaryParts = append(summaryParts, f)
		}
		if len(summaryParts) > 0 {
			e.Summary = strings.Join(summaryParts, " ")
		}

		entries = append(entries, e)
	}

	return entries
}

// isHeaderLine detects table header or separator lines.
func isHeaderLine(line string) bool {
	lower := strings.ToLower(line)
	if strings.HasPrefix(lower, "name") || strings.HasPrefix(lower, "worker") {
		return true
	}
	// Separator lines like "----" or "===="
	trimmed := strings.TrimSpace(line)
	return len(trimmed) > 2 && (strings.Count(trimmed, "-") > len(trimmed)/2 || strings.Count(trimmed, "=") > len(trimmed)/2)
}
