package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckTaskData_ValidFile(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-valid.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	// Should have checks for: file existence, task summary, per-task validation results
	if len(result.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}

	// File existence check should pass
	fileCheck := findCheck(t, result.Checks, "Task file")
	if fileCheck.Status != CheckOK {
		t.Errorf("task file check status = %v, want %v (message: %s)", fileCheck.Status, CheckOK, fileCheck.Message)
	}

	// Task summary should be INFO with status breakdown
	summaryCheck := findCheck(t, result.Checks, "Task summary")
	if summaryCheck.Status != CheckInfo {
		t.Errorf("task summary status = %v, want %v (message: %s)", summaryCheck.Status, CheckInfo, summaryCheck.Message)
	}
	// Should mention task count
	if summaryCheck.Message == "" {
		t.Error("task summary message should not be empty")
	}

	// No FAIL or WARN checks expected for valid data
	for _, check := range result.Checks {
		if check.Status == CheckFail {
			t.Errorf("unexpected FAIL check: %s — %s", check.Name, check.Message)
		}
	}
}

func TestCheckTaskData_FileNotFound(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	// No tasks.yaml file

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	fileCheck := findCheck(t, result.Checks, "Task file")
	if fileCheck.Status != CheckFail {
		t.Errorf("missing file check status = %v, want %v", fileCheck.Status, CheckFail)
	}
}

func TestCheckTaskData_InvalidYAML(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-invalid.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	fileCheck := findCheck(t, result.Checks, "Task file")
	if fileCheck.Status != CheckFail {
		t.Errorf("invalid YAML check status = %v, want %v (message: %s)", fileCheck.Status, CheckFail, fileCheck.Message)
	}
}

func TestCheckTaskData_DuplicateIDs(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-duplicate-ids.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	dupCheck := findCheck(t, result.Checks, "Unique task IDs")
	if dupCheck.Status != CheckFail {
		t.Errorf("duplicate ID check status = %v, want %v (message: %s)", dupCheck.Status, CheckFail, dupCheck.Message)
	}
	// Message should mention the duplicate ID
	if dupCheck.Message == "" {
		t.Error("duplicate ID message should not be empty")
	}
}

func TestCheckTaskData_DanglingDependencies(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-dangling-deps.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	depCheck := findCheck(t, result.Checks, "Dependency references")
	if depCheck.Status != CheckWarn {
		t.Errorf("dangling dep check status = %v, want %v (message: %s)", depCheck.Status, CheckWarn, depCheck.Message)
	}
	if depCheck.Suggestion == "" {
		t.Error("dangling dep check should have a suggestion")
	}
}

func TestCheckTaskData_BlockerStatusInconsistency(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-blocker-inconsistency.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	blockerCheck := findCheck(t, result.Checks, "Blocker consistency")
	if blockerCheck.Status != CheckWarn {
		t.Errorf("blocker consistency check status = %v, want %v (message: %s)", blockerCheck.Status, CheckWarn, blockerCheck.Message)
	}
}

func TestCheckTaskData_CompletedAtInconsistency(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-completed-inconsistency.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	tsCheck := findCheck(t, result.Checks, "Timestamp consistency")
	if tsCheck.Status != CheckWarn {
		t.Errorf("completed_at consistency check status = %v, want %v (message: %s)", tsCheck.Status, CheckWarn, tsCheck.Message)
	}
}

func TestCheckTaskData_LegacyTasksTxt(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	// Create tasks.txt but no tasks.yaml
	if err := os.WriteFile(filepath.Join(configDir, "tasks.txt"), []byte("some old tasks"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	legacyCheck := findCheck(t, result.Checks, "Legacy files")
	if legacyCheck.Status != CheckWarn {
		t.Errorf("legacy file check status = %v, want %v (message: %s)", legacyCheck.Status, CheckWarn, legacyCheck.Message)
	}
}

func TestCheckTaskData_LegacySourceProvider(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-legacy-source-provider.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	legacyCheck := findCheck(t, result.Checks, "Legacy fields")
	if legacyCheck.Status != CheckWarn {
		t.Errorf("legacy field check status = %v, want %v (message: %s)", legacyCheck.Status, CheckWarn, legacyCheck.Message)
	}
}

func TestCheckTaskData_PerTaskValidation(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	// Create a task with an invalid status
	taskYAML := `- id: "task-1"
  text: "Valid task"
  status: "todo"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
- id: "task-2"
  text: ""
  status: "todo"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(configDir, "tasks.yaml"), []byte(taskYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	validationCheck := findCheck(t, result.Checks, "Task validation")
	if validationCheck.Status != CheckWarn {
		t.Errorf("validation check status = %v, want %v (message: %s)", validationCheck.Status, CheckWarn, validationCheck.Message)
	}
}

func TestCheckTaskData_EmptyFile(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "tasks.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	fileCheck := findCheck(t, result.Checks, "Task file")
	if fileCheck.Status != CheckOK {
		t.Errorf("empty file check status = %v, want %v (message: %s)", fileCheck.Status, CheckOK, fileCheck.Message)
	}

	summaryCheck := findCheck(t, result.Checks, "Task summary")
	if summaryCheck.Status != CheckInfo {
		t.Errorf("empty summary status = %v, want %v", summaryCheck.Status, CheckInfo)
	}
}

func TestCheckTaskData_RegisteredAsCategory(t *testing.T) {
	configDir := t.TempDir()
	copyTestFixture(t, "testdata/tasks-valid.yaml", filepath.Join(configDir, "tasks.yaml"))

	dc := NewDoctorChecker(configDir)
	result := dc.Run()

	found := false
	for _, cat := range result.Categories {
		if cat.Name == "Task Data" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Task Data category not found in doctor results")
	}
}

func TestCheckTaskData_AllChecksPassForValidData(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	// Create a complete, valid task file with various statuses
	taskYAML := `- id: "task-1"
  text: "Todo task"
  status: "todo"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
- id: "task-2"
  text: "In progress task"
  status: "in-progress"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-02T00:00:00Z"
- id: "task-3"
  text: "Blocked task"
  status: "blocked"
  blocker: "Waiting on approval"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
- id: "task-4"
  text: "Completed task"
  status: "complete"
  completed_at: "2026-01-10T00:00:00Z"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-10T00:00:00Z"
- id: "task-5"
  text: "Depends on task-1"
  status: "todo"
  depends_on:
    - "task-1"
  created_at: "2026-01-01T00:00:00Z"
  updated_at: "2026-01-01T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(configDir, "tasks.yaml"), []byte(taskYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: configDir}
	result := dc.checkTaskData()

	for _, check := range result.Checks {
		if check.Status == CheckFail || check.Status == CheckWarn {
			t.Errorf("unexpected %v check %q: %s", check.Status, check.Name, check.Message)
		}
	}
}

// --- helpers ---

func copyTestFixture(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture to %s: %v", dst, err)
	}
}

func findCheck(t *testing.T, checks []CheckResult, name string) CheckResult {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("check %q not found in results (have: %v)", name, checkNames(checks))
	return CheckResult{}
}

func checkNames(checks []CheckResult) []string {
	names := make([]string, len(checks))
	for i, c := range checks {
		names[i] = c.Name
	}
	return names
}
