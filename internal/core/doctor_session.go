package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RegisterSessionDataChecks adds the Session Data category to the doctor.
// Called from NewDoctorChecker to keep registration centralized.
func (dc *DoctorChecker) RegisterSessionDataChecks() {
	dc.RegisterCategory("Session Data", dc.checkSessionData)
}

// checkSessionData runs all session and analytics cache checks.
func (dc *DoctorChecker) checkSessionData() CategoryResult {
	var checks []CheckResult

	checks = append(checks, dc.checkSessionsFile())
	checks = append(checks, dc.checkPatternsFile())

	return CategoryResult{Checks: checks}
}

// sessionLineResult holds per-line validation results for sessions.jsonl.
type sessionLineResult struct {
	corruptLines   []int
	incompletLines []int
	totalLines     int
	validSessions  int
	firstSessionAt *time.Time
	lastSessionAt  *time.Time
}

// checkSessionsFile validates sessions.jsonl line-by-line.
func (dc *DoctorChecker) checkSessionsFile() CheckResult {
	result := CheckResult{Name: "Session history"}

	sessionsPath := filepath.Join(dc.configDir, "sessions.jsonl")

	f, err := os.Open(sessionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckInfo
			result.Message = "No sessions recorded yet"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read sessions file: %v", err)
		return result
	}
	defer func() { _ = f.Close() }()

	lr := validateSessionLines(f)

	if lr.totalLines == 0 {
		result.Status = CheckInfo
		result.Message = "No sessions recorded yet"
		return result
	}

	// Build status message parts
	var parts []string
	parts = append(parts, fmt.Sprintf("%d sessions", lr.validSessions))

	if lr.firstSessionAt != nil && lr.lastSessionAt != nil {
		parts = append(parts, fmt.Sprintf("first: %s, last: %s",
			lr.firstSessionAt.Format("2006-01-02"),
			lr.lastSessionAt.Format("2006-01-02")))
	}

	// Check for corrupt lines
	if len(lr.corruptLines) > 0 {
		result.Status = CheckWarn
		lineNums := formatLineNumbers(lr.corruptLines)
		result.Message = fmt.Sprintf("%s; %d corrupt lines (lines %s)",
			strings.Join(parts, ", "), len(lr.corruptLines), lineNums)
		result.Suggestion = "Run threedoors doctor --fix to remove corrupt lines"
		return result
	}

	// Check for incomplete entries (valid JSON but missing required fields)
	if len(lr.incompletLines) > 0 {
		result.Status = CheckWarn
		lineNums := formatLineNumbers(lr.incompletLines)
		result.Message = fmt.Sprintf("%s; %d incomplete entries (lines %s)",
			strings.Join(parts, ", "), len(lr.incompletLines), lineNums)
		result.Suggestion = "Incomplete entries are missing session_id or timestamps"
		return result
	}

	result.Status = CheckOK
	result.Message = strings.Join(parts, ", ")
	return result
}

// validateSessionLines reads a sessions.jsonl file and validates each line.
func validateSessionLines(f *os.File) sessionLineResult {
	var lr sessionLineResult
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, MaxJSONLLineSize), MaxJSONLLineSize)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lr.totalLines++

		var sm SessionMetrics
		if err := json.Unmarshal([]byte(line), &sm); err != nil {
			lr.corruptLines = append(lr.corruptLines, lineNum)
			continue
		}

		// Check required fields
		if sm.SessionID == "" || sm.StartTime.IsZero() {
			lr.incompletLines = append(lr.incompletLines, lineNum)
			continue
		}

		lr.validSessions++

		// Track first/last session times
		t := sm.StartTime
		if lr.firstSessionAt == nil || t.Before(*lr.firstSessionAt) {
			copied := t
			lr.firstSessionAt = &copied
		}
		if lr.lastSessionAt == nil || t.After(*lr.lastSessionAt) {
			copied := t
			lr.lastSessionAt = &copied
		}
	}

	return lr
}

// checkPatternsFile validates patterns.json.
func (dc *DoctorChecker) checkPatternsFile() CheckResult {
	result := CheckResult{Name: "Pattern cache"}

	patternsPath := filepath.Join(dc.configDir, "patterns.json")

	data, err := os.ReadFile(patternsPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckOK
			result.Message = "Pattern cache not yet generated"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read pattern cache: %v", err)
		return result
	}

	if len(data) == 0 {
		result.Status = CheckWarn
		result.Message = "Pattern cache is empty"
		result.Suggestion = "Run threedoors doctor --fix to delete (will regenerate)"
		return result
	}

	var report PatternReport
	if err := json.Unmarshal(data, &report); err != nil {
		result.Status = CheckWarn
		result.Message = "Pattern cache corrupt"
		result.Suggestion = "Run threedoors doctor --fix to delete (will regenerate)"
		return result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Pattern cache valid (%d sessions analyzed at %s)",
		report.SessionCount, report.GeneratedAt.Format("2006-01-02"))
	return result
}

// formatLineNumbers formats a slice of line numbers as a comma-separated string.
// For more than 5 lines, shows the first 5 followed by "and N more".
func formatLineNumbers(lines []int) string {
	if len(lines) <= 5 {
		strs := make([]string, len(lines))
		for i, n := range lines {
			strs[i] = fmt.Sprintf("%d", n)
		}
		return strings.Join(strs, ", ")
	}
	strs := make([]string, 5)
	for i := 0; i < 5; i++ {
		strs[i] = fmt.Sprintf("%d", lines[i])
	}
	return fmt.Sprintf("%s and %d more", strings.Join(strs, ", "), len(lines)-5)
}
