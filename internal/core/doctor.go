package core

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// CheckStatus represents the outcome of a single doctor check.
type CheckStatus int

const (
	// CheckOK means the check passed.
	CheckOK CheckStatus = iota
	// CheckInfo means informational notice, no action needed.
	CheckInfo
	// CheckSkip means the check was skipped (not applicable).
	CheckSkip
	// CheckWarn means the check found a non-critical issue.
	CheckWarn
	// CheckFail means the check found a critical issue.
	CheckFail
)

// String returns the display label for a CheckStatus.
func (s CheckStatus) String() string {
	switch s {
	case CheckOK:
		return "OK"
	case CheckInfo:
		return "INFO"
	case CheckSkip:
		return "SKIP"
	case CheckWarn:
		return "WARN"
	case CheckFail:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}

// Icon returns the flutter-style icon for a CheckStatus.
func (s CheckStatus) Icon() string {
	switch s {
	case CheckOK:
		return "[✓]"
	case CheckInfo:
		return "[i]"
	case CheckSkip:
		return "[ ]"
	case CheckWarn:
		return "[!]"
	case CheckFail:
		return "[✗]"
	default:
		return "[?]"
	}
}

// CheckResult represents the outcome of a single check within a category.
type CheckResult struct {
	Name       string
	Status     CheckStatus
	Message    string
	Suggestion string
}

// CategoryResult represents the outcome of an entire check category.
type CategoryResult struct {
	Name   string
	Status CheckStatus
	Checks []CheckResult
}

// CategoryCheckFunc is a function that runs all checks for a category.
type CategoryCheckFunc func() CategoryResult

// TerminalInfo holds terminal capability information for doctor checks.
// Fields are set by the caller; zero values are treated as "unknown".
type TerminalInfo struct {
	Width        int
	Height       int
	ColorProfile string // "Ascii", "ANSI256", "TrueColor", etc.
}

// DoctorChecker performs category-based system diagnostics.
type DoctorChecker struct {
	configDir    string
	terminal     TerminalInfo
	categories   []registeredCategory
	versionCheck *VersionChecker
}

type registeredCategory struct {
	name    string
	checkFn CategoryCheckFunc
}

// DoctorResult holds the complete output of a doctor run.
type DoctorResult struct {
	Categories []CategoryResult
	Duration   time.Duration
}

// IssueCount returns the total number of warnings and errors across all categories.
func (r *DoctorResult) IssueCount() (warnings, errors int) {
	for _, cat := range r.Categories {
		for _, check := range cat.Checks {
			switch check.Status {
			case CheckWarn:
				warnings++
			case CheckFail:
				errors++
			}
		}
	}
	return warnings, errors
}

// CategoryIssueCount returns the number of categories that have at least one issue.
func (r *DoctorResult) CategoryIssueCount() int {
	count := 0
	for _, cat := range r.Categories {
		for _, check := range cat.Checks {
			if check.Status == CheckWarn || check.Status == CheckFail {
				count++
				break
			}
		}
	}
	return count
}

// OverallStatus returns the worst status across all categories.
func (r *DoctorResult) OverallStatus() CheckStatus {
	worst := CheckOK
	for _, cat := range r.Categories {
		if cat.Status > worst {
			worst = cat.Status
		}
	}
	return worst
}

// NewDoctorChecker creates a DoctorChecker that looks for config in configDir.
func NewDoctorChecker(configDir string) *DoctorChecker {
	dc := &DoctorChecker{configDir: configDir}
	dc.RegisterCategory("Environment", dc.checkEnvironment)
	dc.RegisterCategory("Task Data", dc.checkTaskData)
	dc.RegisterCategory("Sync", dc.checkSync)
	dc.RegisterCategory("Database", dc.checkDatabase)
	dc.RegisterCategory("Version", dc.checkVersion)
	return dc
}

// SetTerminalInfo sets the terminal capability information for environment checks.
func (dc *DoctorChecker) SetTerminalInfo(info TerminalInfo) {
	dc.terminal = info
}

// SetVersionInfo configures the version checker with the current version and channel.
// If not called, the Version category reports "dev build" by default.
// The httpClient and releasesURL parameters are optional — pass nil/"" for defaults.
func (dc *DoctorChecker) SetVersionInfo(version, channel string, httpClient *http.Client, releasesURL string) {
	vc := NewVersionChecker(version, channel, dc.configDir)
	if httpClient != nil {
		vc.HTTPClient = httpClient
	}
	if releasesURL != "" {
		vc.ReleasesURL = releasesURL
	}
	dc.versionCheck = vc
}

// checkVersion runs the Version category checks.
func (dc *DoctorChecker) checkVersion() CategoryResult {
	if dc.versionCheck == nil {
		// No version info configured — treat as dev build
		return CategoryResult{
			Checks: []CheckResult{{
				Name:    "Current version",
				Status:  CheckInfo,
				Message: "Running dev build",
			}},
		}
	}
	return CategoryResult{Checks: dc.versionCheck.Check()}
}

// RegisterCategory adds a check category that runs in registration order.
func (dc *DoctorChecker) RegisterCategory(name string, fn CategoryCheckFunc) {
	dc.categories = append(dc.categories, registeredCategory{name: name, checkFn: fn})
}

// Run executes all registered categories in order and returns the result.
func (dc *DoctorChecker) Run() DoctorResult {
	start := time.Now().UTC()
	var categories []CategoryResult
	for _, cat := range dc.categories {
		result := cat.checkFn()
		result.Name = cat.name
		result.Status = worstCheckStatus(result.Checks)
		categories = append(categories, result)
	}
	return DoctorResult{
		Categories: categories,
		Duration:   time.Since(start),
	}
}

// checkEnvironment runs the Environment category checks.
func (dc *DoctorChecker) checkEnvironment() CategoryResult {
	var checks []CheckResult

	// Check 1: Config directory exists and has correct permissions
	checks = append(checks, dc.checkConfigDir())

	// Check 2: Config file valid YAML and schema version
	checks = append(checks, dc.checkConfigFile())

	// Check 3: Terminal width
	checks = append(checks, dc.checkTerminalWidth())

	// Check 4: Color profile
	checks = append(checks, dc.checkColorProfile())

	// Check 5: Go runtime version
	checks = append(checks, checkGoVersion())

	return CategoryResult{Checks: checks}
}

// checkConfigDir verifies the config directory exists and is accessible.
func (dc *DoctorChecker) checkConfigDir() CheckResult {
	result := CheckResult{Name: "Config directory"}

	info, err := os.Stat(dc.configDir)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckFail
			result.Message = "Config directory not found"
			result.Suggestion = "Run threedoors to create it during onboarding"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot stat config directory: %v", err)
		return result
	}

	if !info.IsDir() {
		result.Status = CheckFail
		result.Message = "Config path exists but is not a directory"
		return result
	}

	// Check read+write permission by attempting to read the directory
	f, err := os.Open(dc.configDir)
	if err != nil {
		result.Status = CheckFail
		result.Message = "Config directory is not readable"
		result.Suggestion = fmt.Sprintf("Run: chmod 700 %s", dc.configDir)
		return result
	}
	_ = f.Close()

	// Check write permission with a temp file
	tmpPath := fmt.Sprintf("%s/.doctor-check.tmp", dc.configDir)
	tf, err := os.Create(tmpPath)
	if err != nil {
		result.Status = CheckWarn
		result.Message = "Config directory is not writable"
		result.Suggestion = fmt.Sprintf("Run: chmod 700 %s", dc.configDir)
		return result
	}
	_ = tf.Close()
	_ = os.Remove(tmpPath)

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Config directory exists (%s)", dc.configDir)
	return result
}

// checkConfigFile verifies the config file is valid YAML with a schema version.
func (dc *DoctorChecker) checkConfigFile() CheckResult {
	result := CheckResult{Name: "Config file"}

	configPath := fmt.Sprintf("%s/config.yaml", dc.configDir)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckWarn
			result.Message = "Config file not found (using defaults)"
			result.Suggestion = "Run: threedoors config init"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read config file: %v", err)
		return result
	}

	// Validate YAML parses
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Config file is not valid YAML: %v", err)
		result.Suggestion = "Check config.yaml for syntax errors"
		return result
	}

	// Check schema_version
	sv, ok := parsed["schema_version"]
	if !ok {
		result.Status = CheckWarn
		result.Message = "Config file missing schema_version field"
		result.Suggestion = fmt.Sprintf("Add schema_version: %d to config.yaml", CurrentSchemaVersion)
		return result
	}

	// yaml.v3 parses integers as int
	var version int
	switch v := sv.(type) {
	case int:
		version = v
	case float64:
		version = int(v)
	default:
		result.Status = CheckWarn
		result.Message = "Config file has non-numeric schema_version"
		return result
	}

	if version > CurrentSchemaVersion {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Unsupported config schema version %d (max supported: %d)", version, CurrentSchemaVersion)
		result.Suggestion = "Update ThreeDoors to the latest version"
		return result
	}

	if version < CurrentSchemaVersion {
		result.Status = CheckInfo
		result.Message = fmt.Sprintf("Config schema version %d (current: %d) — will be auto-migrated", version, CurrentSchemaVersion)
		return result
	}

	// Check required fields
	if _, hasProvider := parsed["provider"]; !hasProvider {
		if _, hasProviders := parsed["providers"]; !hasProviders {
			result.Status = CheckWarn
			result.Message = "Config file missing required field: provider"
			result.Suggestion = "Add provider: textfile to config.yaml"
			return result
		}
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Config file valid (schema v%d)", version)
	return result
}

// checkTerminalWidth reports on terminal width, warning if too narrow.
func (dc *DoctorChecker) checkTerminalWidth() CheckResult {
	result := CheckResult{Name: "Terminal size"}

	if dc.terminal.Width == 0 && dc.terminal.Height == 0 {
		result.Status = CheckInfo
		result.Message = "Terminal size not available"
		return result
	}

	if dc.terminal.Width < 40 {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("Narrow terminal (%d columns) may affect theme display", dc.terminal.Width)
		result.Suggestion = "Widen your terminal to at least 40 columns"
		return result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Terminal size %d×%d", dc.terminal.Width, dc.terminal.Height)
	return result
}

// checkColorProfile reports the terminal color profile.
func (dc *DoctorChecker) checkColorProfile() CheckResult {
	result := CheckResult{Name: "Color support"}

	if dc.terminal.ColorProfile == "" {
		result.Status = CheckInfo
		result.Message = "Color profile not available"
		return result
	}

	if dc.terminal.ColorProfile == "Ascii" {
		result.Status = CheckInfo
		result.Message = "Terminal supports ASCII only — colors will be degraded"
		return result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Color profile: %s", dc.terminal.ColorProfile)
	return result
}

// checkGoVersion reports the Go runtime version as INFO.
func checkGoVersion() CheckResult {
	return CheckResult{
		Name:    "Go runtime",
		Status:  CheckInfo,
		Message: fmt.Sprintf("Built with %s", runtime.Version()),
	}
}

// worstCheckStatus returns the worst (highest severity) status in a list of checks.
func worstCheckStatus(checks []CheckResult) CheckStatus {
	worst := CheckOK
	for _, c := range checks {
		if c.Status > worst {
			worst = c.Status
		}
	}
	return worst
}

// SortCheckStatuses sorts statuses by severity (most severe first).
func SortCheckStatuses(statuses []CheckStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i] > statuses[j]
	})
}
