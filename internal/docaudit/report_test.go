package docaudit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteJSONL_Clean(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	now := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	result := AuditResult{
		Timestamp: now,
		Findings:  nil,
		Clean:     true,
	}

	if err := WriteJSONL(path, result); err != nil {
		t.Fatalf("WriteJSONL() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var entry JSONLEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if entry.Type != "doc_audit_clean" {
		t.Errorf("type = %q, want %q", entry.Type, "doc_audit_clean")
	}
	if entry.Repo != "ThreeDoors" {
		t.Errorf("repo = %q, want %q", entry.Repo, "ThreeDoors")
	}
}

func TestWriteJSONL_WithFindings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	now := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	result := AuditResult{
		Timestamp: now,
		Findings: []Finding{
			{
				Type:        FindingStatusMismatch,
				StoryID:     "1.1",
				Description: "test finding",
				Expected:    "Done",
				Actual:      "Not Started",
				Authority:   "story_file",
				Fix:         "update ROADMAP",
			},
		},
		Clean: false,
	}

	if err := WriteJSONL(path, result); err != nil {
		t.Fatalf("WriteJSONL() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var entry JSONLEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if entry.Type != "doc_inconsistency" {
		t.Errorf("type = %q, want %q", entry.Type, "doc_inconsistency")
	}
	if len(entry.Findings) != 1 {
		t.Errorf("findings count = %d, want 1", len(entry.Findings))
	}
}

func TestWriteJSONL_Appends(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "findings.jsonl")

	now := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	clean := AuditResult{Timestamp: now, Clean: true}

	if err := WriteJSONL(path, clean); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if err := WriteJSONL(path, clean); err != nil {
		t.Fatalf("second write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("lines = %d, want 2", len(lines))
	}
}

func TestFormatHumanSummary_Clean(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	result := AuditResult{Timestamp: now, Clean: true}

	var buf bytes.Buffer
	if err := FormatHumanSummary(&buf, result); err != nil {
		t.Fatalf("FormatHumanSummary() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "CLEAN") {
		t.Errorf("expected CLEAN in output, got: %s", out)
	}
}

func TestFormatHumanSummary_WithFindings(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	result := AuditResult{
		Timestamp: now,
		Findings: []Finding{
			{Type: FindingStatusMismatch, StoryID: "1.1", Description: "mismatch", Expected: "Done", Actual: "Not Started", Fix: "fix it"},
			{Type: FindingOrphanedStory, StoryID: "99.1", Description: "orphan", Expected: "ref", Actual: "none", Fix: "add ref"},
		},
	}

	var buf bytes.Buffer
	if err := FormatHumanSummary(&buf, result); err != nil {
		t.Fatalf("FormatHumanSummary() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "2 inconsistencies") {
		t.Errorf("expected '2 inconsistencies' in output, got: %s", out)
	}
	if !strings.Contains(out, "Status Mismatches") {
		t.Errorf("expected 'Status Mismatches' section in output")
	}
	if !strings.Contains(out, "Orphaned Stories") {
		t.Errorf("expected 'Orphaned Stories' section in output")
	}
}

func TestFormatSupervisorMessage_Clean(t *testing.T) {
	t.Parallel()
	result := AuditResult{Clean: true}
	msg := FormatSupervisorMessage(result)
	if msg != "" {
		t.Errorf("expected empty message for clean audit, got: %q", msg)
	}
}

func TestFormatSupervisorMessage_WithFindings(t *testing.T) {
	t.Parallel()
	result := AuditResult{
		Findings: []Finding{
			{Type: FindingStatusMismatch},
			{Type: FindingStatusMismatch},
			{Type: FindingOrphanedStory},
		},
	}
	msg := FormatSupervisorMessage(result)
	if !strings.Contains(msg, "3 inconsistencies") {
		t.Errorf("expected '3 inconsistencies' in message, got: %q", msg)
	}
	if !strings.Contains(msg, "2 status mismatches") {
		t.Errorf("expected '2 status mismatches' in message, got: %q", msg)
	}
}
