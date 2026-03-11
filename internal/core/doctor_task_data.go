package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// checkTaskData runs all Task Data integrity checks.
func (dc *DoctorChecker) checkTaskData() CategoryResult {
	var checks []CheckResult

	tasksPath := filepath.Join(dc.configDir, "tasks.yaml")

	// Check 1: File existence and YAML validity
	tasks, fileCheck := dc.checkTaskFile(tasksPath)
	checks = append(checks, fileCheck)
	if fileCheck.Status == CheckFail {
		// Cannot continue checks if file is unreadable
		checks = append(checks, dc.checkLegacyFiles())
		return CategoryResult{Checks: checks}
	}

	// Check 2: Task summary by status
	checks = append(checks, dc.checkTaskSummary(tasks))

	// Check 3: Per-task validation
	checks = append(checks, dc.checkTaskValidation(tasks))

	// Check 4: Cross-task consistency
	checks = append(checks, dc.checkUniqueIDs(tasks))
	checks = append(checks, dc.checkDependencyRefs(tasks))
	checks = append(checks, dc.checkBlockerConsistency(tasks))
	checks = append(checks, dc.checkTimestampConsistency(tasks))

	// Check 5: Legacy detection
	checks = append(checks, dc.checkLegacyFiles())
	checks = append(checks, dc.checkLegacyFields(tasks))

	return CategoryResult{Checks: checks}
}

// checkTaskFile verifies tasks.yaml exists and parses as valid YAML.
func (dc *DoctorChecker) checkTaskFile(path string) ([]*Task, CheckResult) {
	result := CheckResult{Name: "Task file"}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckFail
			result.Message = "tasks.yaml not found"
			result.Suggestion = "Run: threedoors add \"your first task\""
			return nil, result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read tasks.yaml: %v", err)
		return nil, result
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		result.Status = CheckOK
		result.Message = "tasks.yaml exists (empty)"
		return nil, result
	}

	var tasks []*Task
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("tasks.yaml is not valid YAML: %v", err)
		result.Suggestion = "Check tasks.yaml for syntax errors"
		return nil, result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("tasks.yaml valid (%d tasks)", len(tasks))
	return tasks, result
}

// checkTaskSummary produces an INFO check with task count by status.
func (dc *DoctorChecker) checkTaskSummary(tasks []*Task) CheckResult {
	result := CheckResult{
		Name:   "Task summary",
		Status: CheckInfo,
	}

	if len(tasks) == 0 {
		result.Message = "0 tasks loaded"
		return result
	}

	counts := make(map[TaskStatus]int)
	for _, t := range tasks {
		counts[t.Status]++
	}

	var parts []string
	// Show counts in a meaningful order
	for _, status := range []TaskStatus{
		StatusTodo, StatusInProgress, StatusBlocked,
		StatusInReview, StatusComplete, StatusDeferred, StatusArchived,
	} {
		if c := counts[status]; c > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c, status))
		}
	}

	result.Message = fmt.Sprintf("%d tasks loaded (%s)", len(tasks), strings.Join(parts, ", "))
	return result
}

// checkTaskValidation runs Validate() on each task and reports failures.
func (dc *DoctorChecker) checkTaskValidation(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Task validation"}

	var failures []string
	for _, t := range tasks {
		if err := t.Validate(); err != nil {
			id := t.ID
			if id == "" {
				id = "(empty ID)"
			}
			failures = append(failures, fmt.Sprintf("%s: %v", id, err))
		}
	}

	if len(failures) == 0 {
		result.Status = CheckOK
		result.Message = fmt.Sprintf("All %d tasks pass validation", len(tasks))
		return result
	}

	result.Status = CheckWarn
	result.Message = fmt.Sprintf("%d task(s) failed validation: %s", len(failures), strings.Join(failures, "; "))
	return result
}

// checkUniqueIDs detects duplicate task IDs.
func (dc *DoctorChecker) checkUniqueIDs(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Unique task IDs"}

	idCounts := make(map[string]int)
	for _, t := range tasks {
		idCounts[t.ID]++
	}

	var dups []string
	for id, count := range idCounts {
		if count > 1 {
			dups = append(dups, fmt.Sprintf("Duplicate task ID: %s (appears %d times)", id, count))
		}
	}

	if len(dups) == 0 {
		result.Status = CheckOK
		result.Message = "All task IDs are unique"
		return result
	}

	result.Status = CheckFail
	result.Message = strings.Join(dups, "; ")
	return result
}

// checkDependencyRefs validates that all depends_on references point to existing tasks.
func (dc *DoctorChecker) checkDependencyRefs(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Dependency references"}

	idSet := make(map[string]bool)
	for _, t := range tasks {
		idSet[t.ID] = true
	}

	var dangling []string
	for _, t := range tasks {
		for _, depID := range t.DependsOn {
			if !idSet[depID] {
				dangling = append(dangling, fmt.Sprintf("Task %s depends on non-existent task %s", t.ID, depID))
			}
		}
	}

	if len(dangling) == 0 {
		result.Status = CheckOK
		result.Message = "All dependency references are valid"
		return result
	}

	result.Status = CheckWarn
	result.Message = strings.Join(dangling, "; ")
	result.Suggestion = "Remove the dependency or recreate the missing task"
	return result
}

// checkBlockerConsistency finds tasks with non-empty Blocker but status != blocked.
func (dc *DoctorChecker) checkBlockerConsistency(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Blocker consistency"}

	var issues []string
	for _, t := range tasks {
		if t.Blocker != "" && t.Status != StatusBlocked {
			issues = append(issues, fmt.Sprintf("Task %s has blocker set but status is %s", t.ID, t.Status))
		}
	}

	if len(issues) == 0 {
		result.Status = CheckOK
		result.Message = "All blocker fields are consistent with status"
		return result
	}

	result.Status = CheckWarn
	result.Message = strings.Join(issues, "; ")
	result.Suggestion = "Either clear the blocker field or set status to blocked"
	return result
}

// checkTimestampConsistency finds tasks with CompletedAt set but status not complete/archived.
func (dc *DoctorChecker) checkTimestampConsistency(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Timestamp consistency"}

	var issues []string
	for _, t := range tasks {
		if t.CompletedAt != nil && t.Status != StatusComplete && t.Status != StatusArchived {
			issues = append(issues, fmt.Sprintf("Task %s has completed_at set but status is %s", t.ID, t.Status))
		}
	}

	if len(issues) == 0 {
		result.Status = CheckOK
		result.Message = "All completion timestamps are consistent with status"
		return result
	}

	result.Status = CheckWarn
	result.Message = strings.Join(issues, "; ")
	result.Suggestion = "Either clear completed_at or set status to complete/archived"
	return result
}

// checkLegacyFiles detects legacy tasks.txt without tasks.yaml.
func (dc *DoctorChecker) checkLegacyFiles() CheckResult {
	result := CheckResult{Name: "Legacy files"}

	txtPath := filepath.Join(dc.configDir, "tasks.txt")
	yamlPath := filepath.Join(dc.configDir, "tasks.yaml")

	_, txtErr := os.Stat(txtPath)
	_, yamlErr := os.Stat(yamlPath)

	txtExists := txtErr == nil
	yamlExists := yamlErr == nil

	if txtExists && !yamlExists {
		result.Status = CheckWarn
		result.Message = "tasks.txt exists but tasks.yaml does not — legacy migration needed"
		result.Suggestion = "Run: threedoors migrate"
		return result
	}

	if txtExists && yamlExists {
		result.Status = CheckInfo
		result.Message = "Both tasks.txt and tasks.yaml exist — tasks.txt can be removed"
		return result
	}

	result.Status = CheckOK
	result.Message = "No legacy task files found"
	return result
}

// checkLegacyFields detects tasks using the deprecated source_provider field.
func (dc *DoctorChecker) checkLegacyFields(tasks []*Task) CheckResult {
	result := CheckResult{Name: "Legacy fields"}

	var legacy []string
	for _, t := range tasks {
		if t.SourceProvider != "" {
			legacy = append(legacy, t.ID)
		}
	}

	if len(legacy) == 0 {
		result.Status = CheckOK
		result.Message = "No legacy fields detected"
		return result
	}

	result.Status = CheckWarn
	result.Message = fmt.Sprintf("%d task(s) use deprecated source_provider field: %s", len(legacy), strings.Join(legacy, ", "))
	result.Suggestion = "Migrate to source_refs using: threedoors migrate"
	return result
}
