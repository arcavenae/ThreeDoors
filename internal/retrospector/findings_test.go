package retrospector

import (
	"strings"
	"testing"
	"time"
)

func TestReadFindingsFrom(t *testing.T) {
	t.Parallel()

	input := `{"pr":100,"story_ref":"43.1","ac_match":"full","ci_first_pass":true,"conflicts":0,"rebase_count":1,"timestamp":"2026-03-10T14:30:00Z","title":"feat: add widget"}
{"pr":101,"story_ref":"43.2","ac_match":"partial","ci_first_pass":false,"conflicts":2,"rebase_count":3,"timestamp":"2026-03-10T15:45:00Z","files_changed":5}
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
	if f0.StoryRef != "43.1" {
		t.Errorf("f0.StoryRef = %q, want 43.1", f0.StoryRef)
	}
	if f0.ACMatch != ACMatchFull {
		t.Errorf("f0.ACMatch = %q, want %q", f0.ACMatch, ACMatchFull)
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
	if f1.Conflicts != 2 {
		t.Errorf("f1.Conflicts = %d, want 2", f1.Conflicts)
	}
	if f1.FilesChanged != 5 {
		t.Errorf("f1.FilesChanged = %d, want 5", f1.FilesChanged)
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
{"pr":100,"ac_match":"full","ci_first_pass":true,"conflicts":0,"rebase_count":0,"timestamp":"2026-03-10T14:30:00Z"}

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
