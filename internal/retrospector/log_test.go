package retrospector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testFinding(pr int) Finding {
	return Finding{
		PR:           pr,
		StoryRef:     "51.3",
		ACMatch:      ACMatchFull,
		CIFirstPass:  true,
		Conflicts:    0,
		RebaseCount:  1,
		Timestamp:    time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		Title:        "feat: test PR",
		FilesChanged: 3,
	}
}

func TestFindingsLog_Append_CreatesFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	f := testFinding(100)
	if err := log.Append(f); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	if _, err := os.Stat(log.Path()); os.IsNotExist(err) {
		t.Fatal("Expected findings log to be created")
	}
}

func TestFindingsLog_Append_ValidJSONL(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	f := testFinding(100)
	if err := log.Append(f); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	data, err := os.ReadFile(log.Path())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var parsed Finding
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if parsed.PR != 100 {
		t.Errorf("PR = %d, want 100", parsed.PR)
	}
	if parsed.ACMatch != ACMatchFull {
		t.Errorf("ACMatch = %q, want %q", parsed.ACMatch, ACMatchFull)
	}
	if !parsed.CIFirstPass {
		t.Error("CIFirstPass = false, want true")
	}
}

func TestFindingsLog_Append_MultipleEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	for i := 1; i <= 5; i++ {
		if err := log.Append(testFinding(i)); err != nil {
			t.Fatalf("Append(%d) error = %v", i, err)
		}
	}

	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(findings) != 5 {
		t.Errorf("ReadAll() returned %d findings, want 5", len(findings))
	}
}

func TestFindingsLog_ReadAll_EmptyFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("ReadAll() returned %d findings, want 0", len(findings))
	}
}

func TestFindingsLog_Archive_TriggersAt1001(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	// Write 1001 entries directly to avoid archival on each append
	file, err := os.OpenFile(log.Path(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	baseTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		f := testFinding(i)
		f.Timestamp = baseTime.Add(time.Duration(i) * time.Minute)
		data, _ := json.Marshal(f)
		_, _ = file.Write(append(data, '\n'))
	}
	_ = file.Close()

	// The 1001st append should trigger archival
	last := testFinding(1001)
	last.Timestamp = baseTime.Add(1001 * time.Minute)
	if err := log.Append(last); err != nil {
		t.Fatalf("Append(1001) error = %v", err)
	}

	// Active log should now have RetainEntries (500) entries
	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(findings) != RetainEntries {
		t.Errorf("Active log has %d entries after archive, want %d", len(findings), RetainEntries)
	}

	// Archive file should exist
	archiveDir := filepath.Join(tmpDir, "archive")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("ReadDir(archive) error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("Expected archive file to be created")
	}

	// Archive should contain the older entries
	archivePath := filepath.Join(archiveDir, entries[0].Name())
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("Open archive: %v", err)
	}
	defer archiveFile.Close() //nolint:errcheck

	archiveCount := 0
	scanner := bufio.NewScanner(archiveFile)
	for scanner.Scan() {
		archiveCount++
	}
	// 1001 total - 500 retained = 501 archived
	if archiveCount != 501 {
		t.Errorf("Archive has %d entries, want 501", archiveCount)
	}
}

func TestFindingsLog_Archive_RetainsNewestEntries(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	// Write entries with ascending PR numbers and timestamps
	file, err := os.OpenFile(log.Path(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	baseTime := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= 1000; i++ {
		f := testFinding(i)
		f.Timestamp = baseTime.Add(time.Duration(i) * time.Minute)
		data, _ := json.Marshal(f)
		_, _ = file.Write(append(data, '\n'))
	}
	_ = file.Close()

	// Trigger archival
	last := testFinding(1001)
	last.Timestamp = baseTime.Add(1001 * time.Minute)
	if err := log.Append(last); err != nil {
		t.Fatalf("Append(1001) error = %v", err)
	}

	// Verify retained entries are the newest (highest PR numbers)
	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// The first retained entry should be PR 502 (1001 - 500 + 1)
	if findings[0].PR != 502 {
		t.Errorf("First retained PR = %d, want 502", findings[0].PR)
	}
	// The last retained entry should be PR 1001
	if findings[len(findings)-1].PR != 1001 {
		t.Errorf("Last retained PR = %d, want 1001", findings[len(findings)-1].PR)
	}
}

func TestFindingsLog_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	if err := log.Append(testFinding(1)); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	info, err := os.Stat(log.Path())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("File permissions = %o, want 600", perm)
	}
}

func TestFindingsLog_PreservesAllFields(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	f := Finding{
		PR:           42,
		StoryRef:     "10.5",
		ACMatch:      ACMatchPartial,
		CIFirstPass:  false,
		Conflicts:    3,
		RebaseCount:  2,
		Timestamp:    time.Date(2026, 3, 10, 15, 30, 0, 0, time.UTC),
		Title:        "fix: broken widget (Story 10.5)",
		FilesChanged: 7,
		ProcessGap:   true,
	}

	if err := log.Append(f); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("ReadAll() returned %d findings, want 1", len(findings))
	}

	got := findings[0]
	if got.PR != 42 {
		t.Errorf("PR = %d, want 42", got.PR)
	}
	if got.StoryRef != "10.5" {
		t.Errorf("StoryRef = %q, want %q", got.StoryRef, "10.5")
	}
	if got.ACMatch != ACMatchPartial {
		t.Errorf("ACMatch = %q, want %q", got.ACMatch, ACMatchPartial)
	}
	if got.CIFirstPass {
		t.Error("CIFirstPass = true, want false")
	}
	if got.Conflicts != 3 {
		t.Errorf("Conflicts = %d, want 3", got.Conflicts)
	}
	if got.RebaseCount != 2 {
		t.Errorf("RebaseCount = %d, want 2", got.RebaseCount)
	}
	if got.FilesChanged != 7 {
		t.Errorf("FilesChanged = %d, want 7", got.FilesChanged)
	}
	if !got.ProcessGap {
		t.Error("ProcessGap = false, want true")
	}
}

func TestFindingsLog_NoStoryEntry(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	log := NewFindingsLog(tmpDir)

	f := Finding{
		PR:          200,
		ACMatch:     ACMatchNoStory,
		CIFirstPass: true,
		Timestamp:   time.Now().UTC(),
		ProcessGap:  true,
	}

	if err := log.Append(f); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	findings, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if findings[0].ACMatch != ACMatchNoStory {
		t.Errorf("ACMatch = %q, want %q", findings[0].ACMatch, ACMatchNoStory)
	}
	if !findings[0].ProcessGap {
		t.Error("ProcessGap = false, want true for no-story PR")
	}
}
