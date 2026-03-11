package saga

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindingsLog_AppendAndReadAll(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	fl := NewFindingsLog(path, 200)

	finding := SagaFinding{
		Type:            FindingSagaDetected,
		Branch:          "fix/ci-lint",
		PR:              500,
		SagaType:        SagaTypeOverlap,
		WorkerCount:     2,
		WorkerNames:     []string{"w1", "w2"},
		FailureRelation: FailureRelated,
		Recommendation:  []Recommendation{RecommendTargetedFix},
		Timestamp:       time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		Repo:            "ThreeDoors",
	}

	if err := fl.Append(finding); err != nil {
		t.Fatalf("append: %v", err)
	}

	entries, err := fl.ReadAll()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.Type != FindingSagaDetected {
		t.Errorf("type: got %q, want %q", got.Type, FindingSagaDetected)
	}
	if got.Branch != "fix/ci-lint" {
		t.Errorf("branch: got %q, want %q", got.Branch, "fix/ci-lint")
	}
	if got.PR != 500 {
		t.Errorf("pr: got %d, want 500", got.PR)
	}
	if got.WorkerCount != 2 {
		t.Errorf("worker_count: got %d, want 2", got.WorkerCount)
	}
	if got.Repo != "ThreeDoors" {
		t.Errorf("repo: got %q, want %q", got.Repo, "ThreeDoors")
	}
}

func TestFindingsLog_MultipleAppends(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	fl := NewFindingsLog(path, 200)

	for i := range 3 {
		f := SagaFinding{
			Type:      FindingSagaDetected,
			Branch:    "fix/ci",
			PR:        500 + i,
			SagaType:  SagaTypeOverlap,
			Timestamp: time.Date(2026, 3, 10, 12, i, 0, 0, time.UTC),
			Repo:      "ThreeDoors",
		}
		if err := fl.Append(f); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	entries, err := fl.ReadAll()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestFindingsLog_RollingRetention(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	fl := NewFindingsLog(path, 3) // Only keep 3 entries.

	for i := range 5 {
		f := SagaFinding{
			Type:      FindingSagaDetected,
			Branch:    "fix/ci",
			PR:        500 + i,
			SagaType:  SagaTypeOverlap,
			Timestamp: time.Date(2026, 3, 10, 12, i, 0, 0, time.UTC),
			Repo:      "ThreeDoors",
		}
		if err := fl.Append(f); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	entries, err := fl.ReadAll()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (rolling), got %d", len(entries))
	}

	// Oldest entries should have been removed — first remaining PR should be 502.
	if entries[0].PR != 502 {
		t.Errorf("oldest retained PR: got %d, want 502", entries[0].PR)
	}
}

func TestFindingsLog_ReadAll_NoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	fl := NewFindingsLog(path, 200)
	_, err := fl.ReadAll()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFindingsLog_ReadAll_SkipsNonSagaEntries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	// Write a mix of saga and non-saga entries.
	content := `{"type":"saga_detected","branch":"fix/ci","pr":500,"saga_type":"overlap","worker_count":2,"worker_names":["w1","w2"],"failure_relation":"related","recommendations":["targeted_fix"],"timestamp":"2026-03-10T12:00:00Z","repo":"ThreeDoors"}
{"pr":501,"story":"43.2","ac_match":"full","ci_first_pass":true,"conflicts":0,"rebase_count":1,"timestamp":"2026-03-10T13:00:00Z","repo":"ThreeDoors"}
{"type":"saga_detected","branch":"fix/test","pr":502,"saga_type":"escalation_trap","worker_count":3,"worker_names":["w1","w2","w3"],"failure_relation":"related","recommendations":["root_cause_analysis"],"timestamp":"2026-03-10T14:00:00Z","repo":"ThreeDoors"}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fl := NewFindingsLog(path, 200)
	entries, err := fl.ReadAll()
	if err != nil {
		t.Fatalf("read all: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 saga entries (skipping per-PR entry), got %d", len(entries))
	}
}

func TestFindingsLog_AtomicWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	fl := NewFindingsLog(path, 200)

	f := SagaFinding{
		Type:      FindingSagaDetected,
		Branch:    "fix/ci",
		PR:        500,
		SagaType:  SagaTypeOverlap,
		Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		Repo:      "ThreeDoors",
	}
	if err := fl.Append(f); err != nil {
		t.Fatalf("append: %v", err)
	}

	// Verify no .tmp file remains.
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not remain after successful write")
	}
}

func TestFindingsLog_FilePermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	fl := NewFindingsLog(path, 200)

	f := SagaFinding{
		Type:      FindingSagaDetected,
		Branch:    "fix/ci",
		PR:        500,
		Timestamp: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		Repo:      "ThreeDoors",
	}
	if err := fl.Append(f); err != nil {
		t.Fatalf("append: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions: got %o, want 0600", perm)
	}
}

func TestFindingFromAlert(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	alert := SagaAlert{
		Type:   SagaTypeOverlap,
		Branch: "fix/ci",
		Workers: []WorkerRecord{
			{Name: "w1", Timestamp: now},
			{Name: "w2", Timestamp: now.Add(1 * time.Hour)},
		},
		FailureRelation: FailureRelated,
		Recommendations: []Recommendation{RecommendTargetedFix},
		Timestamp:       now,
	}

	finding := FindingFromAlert(alert, 500, "ThreeDoors")

	if finding.Type != FindingSagaDetected {
		t.Errorf("type: got %q, want %q", finding.Type, FindingSagaDetected)
	}
	if finding.Branch != "fix/ci" {
		t.Errorf("branch: got %q, want %q", finding.Branch, "fix/ci")
	}
	if finding.PR != 500 {
		t.Errorf("pr: got %d, want 500", finding.PR)
	}
	if finding.WorkerCount != 2 {
		t.Errorf("worker_count: got %d, want 2", finding.WorkerCount)
	}
	if len(finding.WorkerNames) != 2 || finding.WorkerNames[0] != "w1" {
		t.Errorf("worker_names: got %v, want [w1, w2]", finding.WorkerNames)
	}
	if finding.Repo != "ThreeDoors" {
		t.Errorf("repo: got %q, want %q", finding.Repo, "ThreeDoors")
	}
}
