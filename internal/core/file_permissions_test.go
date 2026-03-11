package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncLogCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sl := NewSyncLog(tmpDir)

	entry := SyncLogEntry{
		Provider:  "test",
		Operation: "sync",
		Summary:   "test sync",
	}

	if err := sl.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "sync.log"))
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("sync.log permissions = %o, want 0600", perm)
	}
}

func TestImprovementWriterCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	if err := WriteImprovement(tmpDir, "sess-1", "test improvement"); err != nil {
		t.Fatalf("WriteImprovement: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "improvements.txt"))
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("improvements.txt permissions = %o, want 0600", perm)
	}
}

func TestPlanningMetricsCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	sessionsPath := filepath.Join(tmpDir, "sessions.jsonl")

	event := PlanningSessionEvent{
		TasksReviewed: 3,
	}

	if err := LogPlanningSession(sessionsPath, event); err != nil {
		t.Fatalf("LogPlanningSession: %v", err)
	}

	info, err := os.Stat(sessionsPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("sessions.jsonl permissions = %o, want 0600", perm)
	}
}

func TestMetricsWriterCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	mw := NewMetricsWriter(tmpDir)

	metrics := &SessionMetrics{
		TasksCompleted: 3,
	}

	if err := mw.AppendSession(metrics); err != nil {
		t.Fatalf("AppendSession: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, "sessions.jsonl"))
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("sessions.jsonl permissions = %o, want 0600", perm)
	}
}

func TestSaveProviderConfigCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := defaultProviderConfig()
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveProviderConfig: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("config.yaml permissions = %o, want 0600", perm)
	}
}

func TestSaveValuesConfigCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &ValuesConfig{Values: []string{"focus"}}
	if err := SaveValuesConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveValuesConfig: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("config.yaml permissions = %o, want 0600", perm)
	}
}

func TestPatternAnalyzerSavePatternsCreatesFileWith0600(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	pa := NewPatternAnalyzer()
	path := filepath.Join(tmpDir, "patterns.json")

	report := &PatternReport{}
	if err := pa.SavePatterns(report, path); err != nil {
		t.Fatalf("SavePatterns: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("patterns.json permissions = %o, want 0600", perm)
	}
}

func TestDedupStoreCreatesDirectoryWith0700(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "dedup")
	storePath := filepath.Join(subDir, "store.yaml")

	_, err := NewDedupStore(storePath)
	if err != nil {
		t.Fatalf("NewDedupStore: %v", err)
		return
	}

	info, err := os.Stat(subDir)
	if err != nil {
		t.Fatalf("stat: %v", err)
		return
	}
	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("dedup dir permissions = %o, want 0700", perm)
	}
}
