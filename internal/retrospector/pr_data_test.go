package retrospector

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// mockRunner records commands and returns preset responses.
type mockRunner struct {
	responses map[string][]byte
	errors    map[string]error
	calls     []string
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	key := name
	for _, a := range args {
		key += " " + a
	}
	m.calls = append(m.calls, key)

	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}
	return nil, fmt.Errorf("no mock response for: %s", key)
}

func TestPRDataCollector_FetchMergedPRs(t *testing.T) {
	t.Parallel()

	prs := []ghPR{
		{Number: 100, Title: "feat: old PR", MergedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Number: 200, Title: "feat: new PR", MergedAt: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)},
		{Number: 300, Title: "feat: newest PR", MergedAt: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)},
	}
	data, _ := json.Marshal(prs)

	runner := newMockRunner()
	runner.responses["gh pr list --state merged --json number,title,mergedAt,files,labels --limit 100"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	result, err := collector.FetchMergedPRs(150)
	if err != nil {
		t.Fatalf("FetchMergedPRs() error = %v", err)
	}

	// Should only include PRs after #150
	if len(result) != 2 {
		t.Fatalf("FetchMergedPRs() returned %d PRs, want 2", len(result))
	}
	if result[0].Number != 200 {
		t.Errorf("First PR = %d, want 200", result[0].Number)
	}
	if result[1].Number != 300 {
		t.Errorf("Second PR = %d, want 300", result[1].Number)
	}
}

func TestPRDataCollector_FetchMergedPRs_NoneAfter(t *testing.T) {
	t.Parallel()

	prs := []ghPR{
		{Number: 100, Title: "feat: old PR"},
	}
	data, _ := json.Marshal(prs)

	runner := newMockRunner()
	runner.responses["gh pr list --state merged --json number,title,mergedAt,files,labels --limit 100"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	result, err := collector.FetchMergedPRs(200)
	if err != nil {
		t.Fatalf("FetchMergedPRs() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("FetchMergedPRs() returned %d PRs, want 0", len(result))
	}
}

func TestPRDataCollector_FetchStoryRef(t *testing.T) {
	t.Parallel()

	commits := ghCommitsResponse{
		Commits: []ghCommitNode{
			{MessageHeadline: "feat: add widget (Story 51.3)"},
		},
	}
	data, _ := json.Marshal(commits)

	runner := newMockRunner()
	runner.responses["gh pr view 100 --json commits"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	ref, err := collector.FetchStoryRef(100)
	if err != nil {
		t.Fatalf("FetchStoryRef() error = %v", err)
	}
	if ref != "51.3" {
		t.Errorf("FetchStoryRef() = %q, want %q", ref, "51.3")
	}
}

func TestPRDataCollector_FetchStoryRef_NoRef(t *testing.T) {
	t.Parallel()

	commits := ghCommitsResponse{
		Commits: []ghCommitNode{
			{MessageHeadline: "chore: update deps"},
		},
	}
	data, _ := json.Marshal(commits)

	runner := newMockRunner()
	runner.responses["gh pr view 200 --json commits"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	ref, err := collector.FetchStoryRef(200)
	if err != nil {
		t.Fatalf("FetchStoryRef() error = %v", err)
	}
	if ref != "" {
		t.Errorf("FetchStoryRef() = %q, want empty", ref)
	}
}

func TestPRDataCollector_FetchStoryRef_InBody(t *testing.T) {
	t.Parallel()

	commits := ghCommitsResponse{
		Commits: []ghCommitNode{
			{
				MessageHeadline: "feat: something",
				MessageBody:     "Implements Story 10.5 requirements",
			},
		},
	}
	data, _ := json.Marshal(commits)

	runner := newMockRunner()
	runner.responses["gh pr view 300 --json commits"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	ref, err := collector.FetchStoryRef(300)
	if err != nil {
		t.Fatalf("FetchStoryRef() error = %v", err)
	}
	if ref != "10.5" {
		t.Errorf("FetchStoryRef() = %q, want %q", ref, "10.5")
	}
}

func TestPRDataCollector_FetchCIFirstPass_AllPass(t *testing.T) {
	t.Parallel()

	checks := []struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}{
		{Name: "build", State: "SUCCESS"},
		{Name: "lint", State: "SUCCESS"},
	}
	data, _ := json.Marshal(checks)

	runner := newMockRunner()
	runner.responses["gh pr checks 100 --json name,state"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	passed, err := collector.FetchCIFirstPass(100)
	if err != nil {
		t.Fatalf("FetchCIFirstPass() error = %v", err)
	}
	if !passed {
		t.Error("FetchCIFirstPass() = false, want true")
	}
}

func TestPRDataCollector_FetchCIFirstPass_OneFailed(t *testing.T) {
	t.Parallel()

	checks := []struct {
		Name  string `json:"name"`
		State string `json:"state"`
	}{
		{Name: "build", State: "SUCCESS"},
		{Name: "lint", State: "FAILURE"},
	}
	data, _ := json.Marshal(checks)

	runner := newMockRunner()
	runner.responses["gh pr checks 200 --json name,state"] = data

	collector := NewPRDataCollectorWithRunner(runner)
	passed, err := collector.FetchCIFirstPass(200)
	if err != nil {
		t.Fatalf("FetchCIFirstPass() error = %v", err)
	}
	if passed {
		t.Error("FetchCIFirstPass() = true, want false")
	}
}

func TestPRDataCollector_FetchCIFirstPass_NoChecks(t *testing.T) {
	t.Parallel()

	runner := newMockRunner()
	runner.errors["gh pr checks 300 --json name,state"] = fmt.Errorf("no checks")

	collector := NewPRDataCollectorWithRunner(runner)
	passed, err := collector.FetchCIFirstPass(300)
	if err != nil {
		t.Fatalf("FetchCIFirstPass() error = %v", err)
	}
	if passed {
		t.Error("FetchCIFirstPass() = true, want false when no checks")
	}
}
