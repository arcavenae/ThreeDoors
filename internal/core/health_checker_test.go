package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupHealthTestDir sets up an isolated temp directory for health check tests.
// Always call this at the start of tests that touch the file system.
func setupHealthTestDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	SetHomeDir(tmpDir)
	t.Cleanup(func() { SetHomeDir("") })
	return tmpDir
}

// --- CheckTaskFile Tests ---

func TestCheckTaskFile_FileExistsAndWritable(t *testing.T) {
	setupHealthTestDir(t)
	testTasks := []*Task{newTestTask("a", "Task A", StatusTodo, baseTime)}
	if err := saveTestTasks(testTasks); err != nil {
		t.Fatal(err)
	}

	hc := NewHealthChecker(&MockProvider{Tasks: testTasks})
	item := hc.CheckTaskFile()

	if item.Status != HealthOK {
		t.Errorf("CheckTaskFile() status = %v, want %v", item.Status, HealthOK)
	}
	if item.Suggestion != "" {
		t.Errorf("CheckTaskFile() suggestion = %q, want empty", item.Suggestion)
	}
}

func TestCheckTaskFile_FileMissing(t *testing.T) {
	setupHealthTestDir(t)
	// Don't create any file — dir also doesn't exist

	hc := NewHealthChecker(&MockProvider{})
	item := hc.CheckTaskFile()

	if item.Status != HealthFail {
		t.Errorf("CheckTaskFile() status = %v, want %v", item.Status, HealthFail)
	}
	if !strings.Contains(item.Suggestion, "directory permissions") {
		t.Errorf("CheckTaskFile() suggestion = %q, want containing 'directory permissions'", item.Suggestion)
	}
}

func TestCheckTaskFile_DirNotWritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("test requires non-root user")
	}
	setupHealthTestDir(t)
	testTasks := []*Task{newTestTask("a", "Task A", StatusTodo, baseTime)}
	if err := saveTestTasks(testTasks); err != nil {
		t.Fatal(err)
	}
	// Make the directory read-only so temp file creation fails
	configPath, err := GetConfigDirPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(configPath, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(configPath, 0o755) })

	hc := NewHealthChecker(&MockProvider{Tasks: testTasks})
	item := hc.CheckTaskFile()

	if item.Status != HealthFail {
		t.Errorf("CheckTaskFile() status = %v, want %v", item.Status, HealthFail)
	}
	if !strings.Contains(item.Suggestion, "file permissions") {
		t.Errorf("CheckTaskFile() suggestion = %q, want containing 'file permissions'", item.Suggestion)
	}
}

// --- CheckDatabaseReadWrite Tests ---

func TestCheckDatabaseReadWrite(t *testing.T) {
	tests := []struct {
		name           string
		provider       *MockProvider
		wantStatus     HealthStatus
		wantMsgContain string
		wantSugContain string
	}{
		{
			name: "provider loads OK with 5 tasks",
			provider: &MockProvider{Tasks: []*Task{
				newTestTask("1", "T1", StatusTodo, baseTime),
				newTestTask("2", "T2", StatusTodo, baseTime),
				newTestTask("3", "T3", StatusTodo, baseTime),
				newTestTask("4", "T4", StatusTodo, baseTime),
				newTestTask("5", "T5", StatusTodo, baseTime),
			}},
			wantStatus:     HealthOK,
			wantMsgContain: "5 tasks loaded successfully",
		},
		{
			name:           "provider load error",
			provider:       &MockProvider{LoadErr: fmt.Errorf("disk err")},
			wantStatus:     HealthFail,
			wantSugContain: "corrupt",
		},
		{
			name:           "provider loads 0 tasks",
			provider:       &MockProvider{Tasks: []*Task{}},
			wantStatus:     HealthOK,
			wantMsgContain: "0 tasks loaded successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := NewHealthChecker(tt.provider)
			item := hc.CheckDatabaseReadWrite()

			if item.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", item.Status, tt.wantStatus)
			}
			if tt.wantMsgContain != "" && !strings.Contains(item.Message, tt.wantMsgContain) {
				t.Errorf("message = %q, want containing %q", item.Message, tt.wantMsgContain)
			}
			if tt.wantSugContain != "" && !strings.Contains(item.Suggestion, tt.wantSugContain) {
				t.Errorf("suggestion = %q, want containing %q", item.Suggestion, tt.wantSugContain)
			}
		})
	}
}

// --- CheckSyncStatus Tests ---

func TestCheckSyncStatus_RecentSync(t *testing.T) {
	setupHealthTestDir(t)
	state := SyncState{
		LastSyncTime:  time.Now().Add(-1 * time.Hour),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}
	if err := SaveSyncState(state); err != nil {
		t.Fatal(err)
	}

	hc := NewHealthChecker(&MockProvider{})
	item := hc.CheckSyncStatus()

	if item.Status != HealthOK {
		t.Errorf("status = %v, want %v", item.Status, HealthOK)
	}
	if !strings.Contains(item.Message, "Last sync:") {
		t.Errorf("message = %q, want containing 'Last sync:'", item.Message)
	}
}

func TestCheckSyncStatus_OldSync(t *testing.T) {
	setupHealthTestDir(t)
	state := SyncState{
		LastSyncTime:  time.Now().Add(-48 * time.Hour),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}
	if err := SaveSyncState(state); err != nil {
		t.Fatal(err)
	}

	hc := NewHealthChecker(&MockProvider{})
	item := hc.CheckSyncStatus()

	if item.Status != HealthWarn {
		t.Errorf("status = %v, want %v", item.Status, HealthWarn)
	}
	if !strings.Contains(item.Suggestion, "Press S") {
		t.Errorf("suggestion = %q, want containing 'Press S'", item.Suggestion)
	}
}

func TestCheckSyncStatus_NoSyncStateFile(t *testing.T) {
	setupHealthTestDir(t)
	// Don't create any sync state file

	hc := NewHealthChecker(&MockProvider{})
	item := hc.CheckSyncStatus()

	if item.Status != HealthWarn {
		t.Errorf("status = %v, want %v", item.Status, HealthWarn)
	}
	if !strings.Contains(item.Suggestion, "No sync history") {
		t.Errorf("suggestion = %q, want containing 'No sync history'", item.Suggestion)
	}
}

func TestCheckSyncStatus_CorruptFile(t *testing.T) {
	tmpDir := setupHealthTestDir(t)
	configPath := filepath.Join(tmpDir, ".threedoors")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configPath, "sync_state.yaml"), []byte("{{{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	hc := NewHealthChecker(&MockProvider{})
	item := hc.CheckSyncStatus()

	if item.Status != HealthWarn {
		t.Errorf("status = %v, want %v", item.Status, HealthWarn)
	}
	if !strings.Contains(item.Suggestion, "corrupt") {
		t.Errorf("suggestion = %q, want containing 'corrupt'", item.Suggestion)
	}
}

// --- CheckAppleNotesAccess Tests ---

func TestCheckAppleNotesAccess_TextFileProvider(t *testing.T) {
	setupHealthTestDir(t)
	hc := NewHealthChecker(&inMemoryProvider{})
	item := hc.CheckAppleNotesAccess()

	if item.Status != HealthWarn {
		t.Errorf("status = %v, want %v", item.Status, HealthWarn)
	}
	if !strings.Contains(item.Message, "Apple Notes not configured - using text file backend") {
		t.Errorf("message = %q, want containing exact AC 4 string", item.Message)
	}
}

func TestCheckAppleNotesAccess_MockProviderSuccess(t *testing.T) {
	hc := NewHealthChecker(&MockProvider{Tasks: []*Task{
		newTestTask("a", "Task A", StatusTodo, baseTime),
	}})
	item := hc.CheckAppleNotesAccess()

	if item.Status != HealthOK {
		t.Errorf("status = %v, want %v", item.Status, HealthOK)
	}
}

func TestCheckAppleNotesAccess_MockProviderFailure(t *testing.T) {
	hc := NewHealthChecker(&MockProvider{LoadErr: fmt.Errorf("access denied")})
	item := hc.CheckAppleNotesAccess()

	if item.Status != HealthFail {
		t.Errorf("status = %v, want %v", item.Status, HealthFail)
	}
	if !strings.Contains(item.Suggestion, "Full Disk Access") {
		t.Errorf("suggestion = %q, want containing 'Full Disk Access'", item.Suggestion)
	}
}

// --- Overall Status Determination Tests ---

func TestHealthCheckResult_OverallStatus(t *testing.T) {
	tests := []struct {
		name        string
		statuses    []HealthStatus
		wantOverall HealthStatus
	}{
		{
			name:        "all OK",
			statuses:    []HealthStatus{HealthOK, HealthOK, HealthOK, HealthOK},
			wantOverall: HealthOK,
		},
		{
			name:        "one WARN rest OK",
			statuses:    []HealthStatus{HealthOK, HealthWarn, HealthOK, HealthOK},
			wantOverall: HealthWarn,
		},
		{
			name:        "one FAIL rest OK",
			statuses:    []HealthStatus{HealthOK, HealthOK, HealthFail, HealthOK},
			wantOverall: HealthFail,
		},
		{
			name:        "mixed WARN and FAIL",
			statuses:    []HealthStatus{HealthWarn, HealthWarn, HealthFail, HealthOK},
			wantOverall: HealthFail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]HealthCheckItem, len(tt.statuses))
			for i, s := range tt.statuses {
				items[i] = HealthCheckItem{
					Name:   fmt.Sprintf("Check %d", i),
					Status: s,
				}
			}
			overall := computeOverallStatus(items)
			if overall != tt.wantOverall {
				t.Errorf("overall = %v, want %v", overall, tt.wantOverall)
			}
		})
	}
}

// --- RunAll Integration Test ---

func TestRunAll_Integration(t *testing.T) {
	setupHealthTestDir(t)

	// 1. Create tasks file so CheckTaskFile passes
	testTasks := []*Task{newTestTask("a", "Task A", StatusTodo, time.Now().UTC())}
	if err := saveTestTasks(testTasks); err != nil {
		t.Fatal(err)
	}
	// 2. Create sync state so CheckSyncStatus passes
	if err := SaveSyncState(SyncState{
		LastSyncTime:  time.Now().UTC(),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}); err != nil {
		t.Fatal(err)
	}
	// 3. MockProvider with tasks so CheckDatabaseReadWrite passes
	// MockProvider is not TextFileProvider → CheckAppleNotesAccess returns OK
	provider := &MockProvider{Tasks: testTasks}
	hc := NewHealthChecker(provider)
	result := hc.RunAll()

	if len(result.Items) != 4 {
		t.Errorf("RunAll() returned %d items, want 4", len(result.Items))
	}
	if result.Overall != HealthOK {
		t.Errorf("RunAll() overall = %v, want %v", result.Overall, HealthOK)
		for _, item := range result.Items {
			t.Logf("  %s: %v — %s (suggestion: %s)", item.Name, item.Status, item.Message, item.Suggestion)
		}
	}
	if result.Duration <= 0 {
		t.Errorf("RunAll() duration = %v, want > 0", result.Duration)
	}
}

// --- Performance Test ---

func TestRunAll_Performance(t *testing.T) {
	setupHealthTestDir(t)

	testTasks := make([]*Task, 50)
	for i := range testTasks {
		testTasks[i] = newTestTask(fmt.Sprintf("id-%d", i), fmt.Sprintf("Task %d", i), StatusTodo, time.Now().UTC())
	}
	if err := saveTestTasks(testTasks); err != nil {
		t.Fatal(err)
	}
	if err := SaveSyncState(SyncState{
		LastSyncTime:  time.Now().UTC(),
		TaskSnapshots: make(map[string]TaskSnapshot),
	}); err != nil {
		t.Fatal(err)
	}

	provider := &MockProvider{Tasks: testTasks}
	hc := NewHealthChecker(provider)
	result := hc.RunAll()

	if result.Duration > 3*time.Second {
		t.Errorf("RunAll took %v, expected < 3s", result.Duration)
	}
}

// --- Nil Provider Edge Case ---

func TestCheckDatabaseReadWrite_NilProvider(t *testing.T) {
	hc := NewHealthChecker(nil)
	item := hc.CheckDatabaseReadWrite()

	if item.Status != HealthFail {
		t.Errorf("status = %v, want %v", item.Status, HealthFail)
	}
}

func TestCheckAppleNotesAccess_NilProvider(t *testing.T) {
	hc := NewHealthChecker(nil)
	item := hc.CheckAppleNotesAccess()

	if item.Status != HealthFail {
		t.Errorf("status = %v, want %v", item.Status, HealthFail)
	}
}
