package connection

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectionResult describes a detected tool and how it was found.
type DetectionResult struct {
	ProviderName string            // registry name: "github", "todoist", etc.
	Reason       string            // human-readable: "gh CLI", "API token found", "vault found at ~/vault"
	PreFill      map[string]string // settings to pre-fill (e.g., path, token env var name)
}

// Detector checks whether a particular tool is installed and available.
type Detector interface {
	// Detect checks if the tool is present and returns a result if found.
	// Returns nil if the tool is not detected.
	Detect() *DetectionResult
}

// DetectAll runs all provided detectors and returns results for detected tools.
func DetectAll(detectors []Detector) []DetectionResult {
	var results []DetectionResult
	for _, d := range detectors {
		if r := d.Detect(); r != nil {
			results = append(results, *r)
		}
	}
	return results
}

// DefaultDetectors returns the built-in set of tool detectors.
func DefaultDetectors() []Detector {
	return []Detector{
		&GHDetector{},
		&TodoistDetector{},
		&ObsidianDetector{},
		&JiraDetector{},
	}
}

// GHDetector checks for the GitHub CLI (gh) being installed and authenticated.
type GHDetector struct {
	// lookPathFn allows injecting a custom exec.LookPath for testing.
	lookPathFn func(string) (string, error)
	// runCmdFn allows injecting a custom command runner for testing.
	runCmdFn func(string, ...string) ([]byte, error)
}

// Detect checks if gh CLI is installed and authenticated.
func (d *GHDetector) Detect() *DetectionResult {
	lookPath := d.lookPathFn
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	runCmd := d.runCmdFn
	if runCmd == nil {
		runCmd = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).Output()
		}
	}

	// Check if gh is installed
	if _, err := lookPath("gh"); err != nil {
		return nil
	}

	// Check if gh is authenticated
	out, err := runCmd("gh", "auth", "status")
	if err != nil {
		return nil
	}

	reason := "gh CLI"
	if strings.Contains(string(out), "Logged in") {
		reason = "gh CLI authenticated"
	}

	return &DetectionResult{
		ProviderName: "github",
		Reason:       reason,
		PreFill:      map[string]string{},
	}
}

// TodoistDetector checks for Todoist API token availability.
type TodoistDetector struct {
	// lookupEnvFn allows injecting a custom env lookup for testing.
	lookupEnvFn func(string) (string, bool)
}

// Detect checks for TODOIST_API_TOKEN or THREEDOORS_TODOIST_TOKEN env vars.
func (d *TodoistDetector) Detect() *DetectionResult {
	lookupEnv := d.lookupEnvFn
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	envVars := []string{
		"TODOIST_API_TOKEN",
		"THREEDOORS_TODOIST_TOKEN",
	}

	for _, key := range envVars {
		if val, ok := lookupEnv(key); ok && val != "" {
			return &DetectionResult{
				ProviderName: "todoist",
				Reason:       "API token found",
				PreFill:      map[string]string{"token_env": key},
			}
		}
	}

	return nil
}

// ObsidianDetector checks for .obsidian/ directories in common locations.
type ObsidianDetector struct {
	// homeDirFn allows injecting a custom home directory for testing.
	homeDirFn func() (string, error)
	// statFn allows injecting a custom os.Stat for testing.
	statFn func(string) (os.FileInfo, error)
	// globFn allows injecting a custom filepath.Glob for testing.
	globFn func(string) ([]string, error)
}

// Detect looks for .obsidian/ directories in common vault locations.
func (d *ObsidianDetector) Detect() *DetectionResult {
	homeDir := d.homeDirFn
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	stat := d.statFn
	if stat == nil {
		stat = os.Stat
	}
	glob := d.globFn
	if glob == nil {
		glob = filepath.Glob
	}

	home, err := homeDir()
	if err != nil {
		return nil
	}

	// Check common vault locations
	searchPaths := []string{
		filepath.Join(home, "Documents"),
		filepath.Join(home, "Desktop"),
		home,
	}

	for _, base := range searchPaths {
		pattern := filepath.Join(base, "*", ".obsidian")
		matches, err := glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			// The vault is the parent of .obsidian
			vaultPath := filepath.Dir(match)
			return &DetectionResult{
				ProviderName: "obsidian",
				Reason:       fmt.Sprintf("vault found at %s", vaultPath),
				PreFill:      map[string]string{"path": vaultPath},
			}
		}
	}

	return nil
}

// JiraDetector checks for Jira configuration files.
type JiraDetector struct {
	// homeDirFn allows injecting a custom home directory for testing.
	homeDirFn func() (string, error)
	// statFn allows injecting a custom os.Stat for testing.
	statFn func(string) (os.FileInfo, error)
	// lookupEnvFn allows injecting a custom env lookup for testing.
	lookupEnvFn func(string) (string, bool)
}

// Detect checks for Jira config files or environment variables.
func (d *JiraDetector) Detect() *DetectionResult {
	homeDir := d.homeDirFn
	if homeDir == nil {
		homeDir = os.UserHomeDir
	}
	stat := d.statFn
	if stat == nil {
		stat = os.Stat
	}
	lookupEnv := d.lookupEnvFn
	if lookupEnv == nil {
		lookupEnv = os.LookupEnv
	}

	// Check env vars first
	envVars := []string{
		"JIRA_API_TOKEN",
		"THREEDOORS_JIRA_TOKEN",
	}
	for _, key := range envVars {
		if val, ok := lookupEnv(key); ok && val != "" {
			return &DetectionResult{
				ProviderName: "jira",
				Reason:       "API token found",
				PreFill:      map[string]string{"token_env": key},
			}
		}
	}

	// Check for config files
	home, err := homeDir()
	if err != nil {
		return nil
	}

	configPaths := []string{
		filepath.Join(home, ".jira.d"),
		filepath.Join(home, ".config", "jira"),
		filepath.Join(home, ".atlassian"),
	}

	for _, p := range configPaths {
		if info, err := stat(p); err == nil && info.IsDir() {
			return &DetectionResult{
				ProviderName: "jira",
				Reason:       fmt.Sprintf("config found at %s", p),
				PreFill:      map[string]string{},
			}
		}
	}

	return nil
}
