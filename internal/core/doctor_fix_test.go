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

// --- Orphaned temp file fix ---

func TestFixOrphanedTmpFiles_RemovesOldFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create old .tmp files
	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	for _, name := range []string{"a.tmp", "b.tmp"} {
		p := filepath.Join(tmpDir, name)
		if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}
	}

	dc := &DoctorChecker{configDir: tmpDir, fix: true}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}
	if result.Message != "FIXED: removed 2 stale .tmp files" {
		t.Errorf("message = %q, want %q", result.Message, "FIXED: removed 2 stale .tmp files")
	}

	// Verify files are actually gone
	for _, name := range []string{"a.tmp", "b.tmp"} {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); !os.IsNotExist(err) {
			t.Errorf("file %s still exists after fix", name)
		}
	}
}

func TestFixOrphanedTmpFiles_KeepsRecentFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create a recent .tmp file (should not be removed)
	if err := os.WriteFile(filepath.Join(tmpDir, "recent.tmp"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir, fix: true}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckOK {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckOK, result.Message)
	}
}

func TestFixOrphanedTmpFiles_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	p := filepath.Join(tmpDir, "orphan.tmp")
	if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir, fix: false}
	result := dc.checkOrphanedTmpFiles()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}

	// File should still exist
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Error("file was removed without --fix")
	}
}

// --- Corrupt patterns.json fix ---

func TestFixPatternsFile_DeletesCorruptFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	patternsPath := filepath.Join(dir, "patterns.json")
	if err := os.WriteFile(patternsPath, []byte("{not json}"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkPatternsFile()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}
	if result.Message != "FIXED: deleted patterns.json (will regenerate)" {
		t.Errorf("message = %q", result.Message)
	}

	// Verify file is gone
	if _, err := os.Stat(patternsPath); !os.IsNotExist(err) {
		t.Error("patterns.json still exists after fix")
	}
}

func TestFixPatternsFile_DeletesEmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	patternsPath := filepath.Join(dir, "patterns.json")
	if err := os.WriteFile(patternsPath, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkPatternsFile()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v", result.Status, CheckFixed)
	}

	if _, err := os.Stat(patternsPath); !os.IsNotExist(err) {
		t.Error("empty patterns.json still exists after fix")
	}
}

func TestFixPatternsFile_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	patternsPath := filepath.Join(dir, "patterns.json")
	if err := os.WriteFile(patternsPath, []byte("{bad}"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: false}
	result := dc.checkPatternsFile()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}

	// File should still exist
	if _, err := os.Stat(patternsPath); os.IsNotExist(err) {
		t.Error("patterns.json was removed without --fix")
	}
}

// --- Missing config.yaml fix ---

func TestFixConfigFile_GeneratesSampleConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	dc := &DoctorChecker{configDir: dir, fix: true, registry: NewRegistry()}
	result := dc.checkConfigFile()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}
	if result.Message != "FIXED: created sample config.yaml" {
		t.Errorf("message = %q", result.Message)
	}

	// Verify config.yaml was created
	configPath := filepath.Join(dir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config.yaml not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("config.yaml is empty")
	}
}

func TestFixConfigFile_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	dc := &DoctorChecker{configDir: dir, fix: false}
	result := dc.checkConfigFile()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}

	// config.yaml should not exist
	if _, err := os.Stat(filepath.Join(dir, "config.yaml")); !os.IsNotExist(err) {
		t.Error("config.yaml was created without --fix")
	}
}

// --- Directory permissions fix ---

func TestFixConfigDir_FixesPermissions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Set restrictive permissions (read-only)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Restore permissions so TempDir cleanup works
		_ = os.Chmod(dir, 0o700)
	})

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkConfigDir()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}
	if result.Message != "FIXED: set directory permissions to 700" {
		t.Errorf("message = %q", result.Message)
	}
}

func TestFixConfigDir_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o700)
	})

	dc := &DoctorChecker{configDir: dir, fix: false}
	result := dc.checkConfigDir()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckWarn, result.Message)
	}
}

// --- Legacy tasks.txt migration fix ---

func TestFixLegacyFiles_MigratesTasks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create tasks.txt with some tasks
	txtPath := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(txtPath, []byte("Buy groceries\nClean house\nRead a book\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkLegacyFiles()

	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}

	// Verify tasks.yaml was created
	yamlPath := filepath.Join(dir, "tasks.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("tasks.yaml not created: %v", err)
	}

	// Verify it contains tasks
	type tasksFileCheck struct {
		Tasks []*Task `yaml:"tasks"`
	}
	var tf tasksFileCheck
	if err := yaml.Unmarshal(data, &tf); err != nil {
		t.Fatalf("tasks.yaml is not valid YAML: %v", err)
	}
	if len(tf.Tasks) != 3 {
		t.Errorf("got %d tasks, want 3", len(tf.Tasks))
	}

	// Verify backup exists
	bakPath := txtPath + ".bak"
	if _, err := os.Stat(bakPath); os.IsNotExist(err) {
		t.Error("tasks.txt.bak was not created")
	}

	// Verify original tasks.txt is gone
	if _, err := os.Stat(txtPath); !os.IsNotExist(err) {
		t.Error("tasks.txt still exists (should have been renamed to .bak)")
	}
}

func TestFixLegacyFiles_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	txtPath := filepath.Join(dir, "tasks.txt")
	if err := os.WriteFile(txtPath, []byte("Some task\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: false}
	result := dc.checkLegacyFiles()

	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}

	// tasks.yaml should not exist
	yamlPath := filepath.Join(dir, "tasks.yaml")
	if _, err := os.Stat(yamlPath); !os.IsNotExist(err) {
		t.Error("tasks.yaml was created without --fix")
	}
}

// --- Version cache fix ---

func TestFixVersionCache_DeletesCorrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cachePath := filepath.Join(dir, versionCheckCacheFile)
	if err := os.WriteFile(cachePath, []byte("{not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkVersionCache()

	if result == nil {
		t.Fatal("expected non-nil result for corrupt cache")
	}
	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v (message: %s)", result.Status, CheckFixed, result.Message)
	}
	if result.Message != "FIXED: cleared corrupt version cache" {
		t.Errorf("message = %q", result.Message)
	}

	// Verify file is gone
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("corrupt version cache still exists")
	}
}

func TestFixVersionCache_DeletesStale(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cache := VersionCheckCache{
		CheckedAt:      time.Now().UTC().Add(-10 * 24 * time.Hour),
		LatestVersions: map[string]string{"stable": "1.0.0"},
	}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatal(err)
	}

	cachePath := filepath.Join(dir, versionCheckCacheFile)
	if err := os.WriteFile(cachePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkVersionCache()

	if result == nil {
		t.Fatal("expected non-nil result for stale cache")
	}
	if result.Status != CheckFixed {
		t.Errorf("status = %v, want %v", result.Status, CheckFixed)
	}
	if result.Message != "FIXED: cleared stale version cache" {
		t.Errorf("message = %q", result.Message)
	}
}

func TestFixVersionCache_NoFixWithoutFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cachePath := filepath.Join(dir, versionCheckCacheFile)
	if err := os.WriteFile(cachePath, []byte("{bad}"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: false}
	result := dc.checkVersionCache()

	if result == nil {
		t.Fatal("expected non-nil result for corrupt cache")
	}
	if result.Status != CheckWarn {
		t.Errorf("status = %v, want %v", result.Status, CheckWarn)
	}

	// File should still exist
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("version cache removed without --fix")
	}
}

func TestFixVersionCache_FreshCacheReturnsNil(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cache := VersionCheckCache{
		CheckedAt:      time.Now().UTC(),
		LatestVersions: map[string]string{"stable": "1.0.0"},
	}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, versionCheckCacheFile), data, 0o600); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkVersionCache()

	if result != nil {
		t.Errorf("expected nil result for fresh cache, got status=%v message=%q", result.Status, result.Message)
	}
}

// --- Report-only items NOT modified by --fix ---

func TestFix_DoesNotModifyCorruptTasksYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create corrupt tasks.yaml
	yamlPath := filepath.Join(dir, "tasks.yaml")
	corruptData := []byte("{{not yaml")
	if err := os.WriteFile(yamlPath, corruptData, 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkTaskData()

	// Find the task file check
	for _, check := range result.Checks {
		if check.Name == "Task file" {
			if check.Status != CheckFail {
				t.Errorf("corrupt tasks.yaml status = %v, want %v", check.Status, CheckFail)
			}
			break
		}
	}

	// Verify file was NOT modified
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(corruptData) {
		t.Error("corrupt tasks.yaml was modified by --fix (should be report-only)")
	}
}

func TestFix_DoesNotModifyDuplicateIDs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create tasks.yaml with duplicate IDs
	type tf struct {
		Tasks []*Task `yaml:"tasks"`
	}
	tasks := tf{Tasks: []*Task{
		{ID: "dup-1", Text: "Task A", Status: StatusTodo},
		{ID: "dup-1", Text: "Task B", Status: StatusTodo},
	}}
	data, err := yaml.Marshal(&tasks)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.yaml"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: dir, fix: true}
	result := dc.checkTaskData()

	for _, check := range result.Checks {
		if check.Name == "Unique task IDs" {
			if check.Status == CheckFixed {
				t.Error("duplicate IDs should NOT be auto-fixed")
			}
			break
		}
	}
}

// --- CheckFixed status type ---

func TestCheckFixed_String(t *testing.T) {
	t.Parallel()
	if got := CheckFixed.String(); got != "FIXED" {
		t.Errorf("CheckFixed.String() = %q, want %q", got, "FIXED")
	}
}

func TestCheckFixed_Icon(t *testing.T) {
	t.Parallel()
	if got := CheckFixed.Icon(); got != "[F]" {
		t.Errorf("CheckFixed.Icon() = %q, want %q", got, "[F]")
	}
}

// --- DoctorResult summary methods ---

func TestDoctorResult_FixedCount(t *testing.T) {
	t.Parallel()
	result := DoctorResult{
		Categories: []CategoryResult{
			{
				Checks: []CheckResult{
					{Status: CheckFixed},
					{Status: CheckOK},
					{Status: CheckFixed},
				},
			},
			{
				Checks: []CheckResult{
					{Status: CheckWarn},
					{Status: CheckFixed},
				},
			},
		},
	}

	if got := result.FixedCount(); got != 3 {
		t.Errorf("FixedCount() = %d, want 3", got)
	}
}

func TestDoctorResult_ManualCount(t *testing.T) {
	t.Parallel()
	result := DoctorResult{
		Categories: []CategoryResult{
			{
				Checks: []CheckResult{
					{Status: CheckFixed},
					{Status: CheckWarn},
					{Status: CheckFail},
				},
			},
		},
	}

	if got := result.ManualCount(); got != 2 {
		t.Errorf("ManualCount() = %d, want 2", got)
	}
}

// --- Integration: full doctor run with --fix ---

func TestDoctorRun_FixIntegration(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create config.yaml so Environment passes
	configContent := fmt.Sprintf("schema_version: %d\nprovider: textfile\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create orphaned .tmp file
	tmpPath := filepath.Join(dir, "stale.tmp")
	if err := os.WriteFile(tmpPath, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	if err := os.Chtimes(tmpPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create corrupt patterns.json
	if err := os.WriteFile(filepath.Join(dir, "patterns.json"), []byte("{bad}"), 0o600); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(dir)
	dc.SetFix(true)
	result := dc.Run()

	// Should have some fixed items
	if result.FixedCount() == 0 {
		t.Error("expected at least one fixed item in integration test")
	}

	// Orphaned .tmp should be gone
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("orphaned .tmp file still exists after fix run")
	}

	// patterns.json should be gone
	if _, err := os.Stat(filepath.Join(dir, "patterns.json")); !os.IsNotExist(err) {
		t.Error("corrupt patterns.json still exists after fix run")
	}
}
