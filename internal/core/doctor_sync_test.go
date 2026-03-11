package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestCheckSyncState_NoFile(t *testing.T) {
	t.Parallel()
	dc := &DoctorChecker{configDir: t.TempDir()}
	result := dc.checkSyncState()

	if result.Status != CheckInfo {
		t.Errorf("status = %v, want %v", result.Status, CheckInfo)
	}
	if result.Message != "No sync history" {
		t.Errorf("message = %q, want %q", result.Message, "No sync history")
	}
}

func TestCheckSyncState_InvalidYAML(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, syncStateFile), []byte("{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkSyncState()

	if result.Status != CheckFail {
		t.Errorf("status = %v, want %v", result.Status, CheckFail)
	}
}

func TestCheckSyncState_ZeroTime(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	state := SyncState{TaskSnapshots: map[string]TaskSnapshot{}}
	data, err := yaml.Marshal(&state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, syncStateFile), data, 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkSyncState()

	if result.Status != CheckInfo {
		t.Errorf("status = %v, want %v", result.Status, CheckInfo)
	}
	if result.Message != "No sync history" {
		t.Errorf("message = %q, want %q", result.Message, "No sync history")
	}
}

func TestCheckSyncState_Fresh(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	state := SyncState{
		LastSyncTime:  time.Now().UTC().Add(-1 * time.Hour),
		TaskSnapshots: map[string]TaskSnapshot{},
	}
	data, err := yaml.Marshal(&state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, syncStateFile), data, 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkSyncState()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckOK, result.Message)
	}
}

func TestCheckSyncState_Stale(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	state := SyncState{
		LastSyncTime:  time.Now().UTC().Add(-3 * 24 * time.Hour),
		TaskSnapshots: map[string]TaskSnapshot{},
	}
	data, err := yaml.Marshal(&state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, syncStateFile), data, 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkSyncState()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != "Last sync: 3 days ago" {
		t.Errorf("message = %q, want %q", result.Message, "Last sync: 3 days ago")
	}
	if result.Suggestion != "Press S in doors view to trigger sync" {
		t.Errorf("suggestion = %q, want %q", result.Suggestion, "Press S in doors view to trigger sync")
	}
}

func TestCheckWALQueue_NoFile(t *testing.T) {
	t.Parallel()
	dc := &DoctorChecker{configDir: t.TempDir()}
	result := dc.checkWALQueue()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v", result.Status, CheckOK)
	}
	if result.Message != "No pending WAL entries" {
		t.Errorf("message = %q, want %q", result.Message, "No pending WAL entries")
	}
}

func TestCheckWALQueue_EmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, walFile), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkWALQueue()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v", result.Status, CheckOK)
	}
}

func TestCheckWALQueue_HealthyEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	entries := []WALEntry{
		{Sequence: 1, Operation: WALOpSave, TaskID: "t1", Timestamp: time.Now().UTC(), Retries: 0},
		{Sequence: 2, Operation: WALOpDelete, TaskID: "t2", Timestamp: time.Now().UTC(), Retries: 3},
	}
	writeWALFixture(t, tmpDir, entries)

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkWALQueue()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckOK, result.Message)
	}
	if result.Message != "2 pending WAL entries" {
		t.Errorf("message = %q, want %q", result.Message, "2 pending WAL entries")
	}
}

func TestCheckWALQueue_StuckEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	entries := []WALEntry{
		{Sequence: 1, Operation: WALOpSave, TaskID: "t1", Timestamp: time.Now().UTC(), Retries: 0},
		{Sequence: 2, Operation: WALOpSave, TaskID: "t2", Timestamp: time.Now().UTC(), Retries: 10},
		{Sequence: 3, Operation: WALOpSave, TaskID: "t3", Timestamp: time.Now().UTC(), Retries: 15},
	}
	writeWALFixture(t, tmpDir, entries)

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkWALQueue()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != fmt.Sprintf("2 stuck operations (retries >= %d)", walStuckRetries) {
		t.Errorf("message = %q", result.Message)
	}
}

func TestCheckWALQueue_ExcessiveBacklog(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Write 10001 valid entries
	f, err := os.Create(filepath.Join(tmpDir, walFile))
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(f)
	for i := 0; i < 10001; i++ {
		entry := WALEntry{
			Sequence:  int64(i + 1),
			Operation: WALOpSave,
			TaskID:    fmt.Sprintf("t%d", i),
			Timestamp: time.Now().UTC(),
		}
		if err := encoder.Encode(entry); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkWALQueue()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != "Excessive backlog: 10001 entries" {
		t.Errorf("message = %q, want %q", result.Message, "Excessive backlog: 10001 entries")
	}
}

func TestCheckWALQueue_CorruptEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	content := `{"seq":1,"op":"save","task_id":"t1","timestamp":"2025-01-01T00:00:00Z","retries":0}
not valid json
{"seq":2,"op":"save","task_id":"t2","timestamp":"2025-01-01T00:00:00Z","retries":0}
`
	if err := os.WriteFile(filepath.Join(tmpDir, walFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkWALQueue()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != "1 corrupt entries in WAL queue" {
		t.Errorf("message = %q, want %q", result.Message, "1 corrupt entries in WAL queue")
	}
}

func TestCheckOrphanedTmpFiles_None(t *testing.T) {
	t.Parallel()
	dc := &DoctorChecker{configDir: t.TempDir()}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v", result.Status, CheckOK)
	}
	if result.Message != "No orphaned temp files" {
		t.Errorf("message = %q, want %q", result.Message, "No orphaned temp files")
	}
}

func TestCheckOrphanedTmpFiles_RecentTmp(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	// Create a fresh .tmp file (not orphaned)
	if err := os.WriteFile(filepath.Join(tmpDir, "sync_state.yaml.tmp"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckOK, result.Message)
	}
}

func TestCheckOrphanedTmpFiles_OldTmp(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a .tmp file and backdate it
	tmpPath := filepath.Join(tmpDir, "sync_state.yaml.tmp")
	if err := os.WriteFile(tmpPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	if err := os.Chtimes(tmpPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != "1 orphaned temp files found" {
		t.Errorf("message = %q, want %q", result.Message, "1 orphaned temp files found")
	}
	if result.Suggestion != "Run threedoors doctor --fix" {
		t.Errorf("suggestion = %q, want %q", result.Suggestion, "Run threedoors doctor --fix")
	}
}

func TestCheckOrphanedTmpFiles_MultipleOld(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	oldTime := time.Now().UTC().Add(-3 * time.Hour)
	for _, name := range []string{"a.tmp", "b.tmp", "c.tmp"} {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}
	if result.Message != "3 orphaned temp files found" {
		t.Errorf("message = %q, want %q", result.Message, "3 orphaned temp files found")
	}
}

func TestCheckSync_Integration(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkSync()

	if len(result.Checks) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(result.Checks))
	}

	names := []string{"Sync state", "WAL queue", "Temp files"}
	for i, want := range names {
		if result.Checks[i].Name != want {
			t.Errorf("check[%d].Name = %q, want %q", i, result.Checks[i].Name, want)
		}
	}
}

func TestCheckSync_RegisteredInDoctor(t *testing.T) {
	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	found := false
	for _, cat := range result.Categories {
		if cat.Name == "Sync" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sync category not found in doctor run")
	}
}

func TestFormatDuration_SyncUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"30 minutes", 30 * time.Minute, "30 minutes"},
		{"5 hours", 5 * time.Hour, "5 hours"},
		{"3 days", 72 * time.Hour, "3 days"},
		{"7 days", 168 * time.Hour, "7 days"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatDuration(tt.d); got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// writeWALFixture writes WAL entries as JSONL to the standard WAL file in the given dir.
func writeWALFixture(t *testing.T, dir string, entries []WALEntry) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, walFile))
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(f)
	for _, entry := range entries {
		if err := encoder.Encode(entry); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
