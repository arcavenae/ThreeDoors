package retrospector

import (
	"strings"
	"testing"
	"time"
)

func TestReadFindingsFrom(t *testing.T) {
	t.Parallel()

	input := `{"pr":100,"story":"43.1","ac_match":"full","ci_first_pass":true,"conflicts":0,"rebase_count":1,"timestamp":"2026-03-10T14:30:00Z","repo":"ThreeDoors"}
{"pr":101,"story":"43.2","ac_match":"partial","ci_first_pass":false,"ci_failures":["lint"],"conflicts":2,"rebase_count":3,"timestamp":"2026-03-10T15:45:00Z","repo":"ThreeDoors"}
`
	findings, err := readFindingsFrom(strings.NewReader(input))
	if err != nil {
		t.Fatalf("readFindingsFrom: %v", err)
	}

	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2", len(findings))
	}

	f0 := findings[0]
	if f0.PR != 100 {
		t.Errorf("f0.PR = %d, want 100", f0.PR)
	}
	if f0.Story != "43.1" {
		t.Errorf("f0.Story = %q, want 43.1", f0.Story)
	}
	if f0.ACMatch != "full" {
		t.Errorf("f0.ACMatch = %q, want full", f0.ACMatch)
	}
	if !f0.CIFirstPass {
		t.Error("f0.CIFirstPass should be true")
	}
	wantTime := time.Date(2026, 3, 10, 14, 30, 0, 0, time.UTC)
	if !f0.Timestamp.Equal(wantTime) {
		t.Errorf("f0.Timestamp = %v, want %v", f0.Timestamp, wantTime)
	}

	f1 := findings[1]
	if f1.CIFirstPass {
		t.Error("f1.CIFirstPass should be false")
	}
	if len(f1.CIFailures) != 1 || f1.CIFailures[0] != "lint" {
		t.Errorf("f1.CIFailures = %v, want [lint]", f1.CIFailures)
	}
	if f1.Conflicts != 2 {
		t.Errorf("f1.Conflicts = %d, want 2", f1.Conflicts)
	}
}

func TestReadFindingsFromEmpty(t *testing.T) {
	t.Parallel()

	findings, err := readFindingsFrom(strings.NewReader(""))
	if err != nil {
		t.Fatalf("readFindingsFrom empty: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0", len(findings))
	}
}

func TestReadFindingsFromBlankLines(t *testing.T) {
	t.Parallel()

	input := `
{"pr":100,"ac_match":"full","ci_first_pass":true,"conflicts":0,"rebase_count":0,"timestamp":"2026-03-10T14:30:00Z","repo":"ThreeDoors"}

`
	findings, err := readFindingsFrom(strings.NewReader(input))
	if err != nil {
		t.Fatalf("readFindingsFrom: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1", len(findings))
	}
}

func TestReadFindingsFromInvalidJSON(t *testing.T) {
	t.Parallel()

	input := `{"pr":100}
not valid json
`
	_, err := readFindingsFrom(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("error should reference line 2: %v", err)
	}
}

func TestReadFindingsNonexistentFile(t *testing.T) {
	t.Parallel()

	findings, err := ReadFindings("/tmp/does-not-exist-retrospector-test.jsonl")
	if err != nil {
		t.Fatalf("ReadFindings nonexistent: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0 for nonexistent file", len(findings))
	}
}
