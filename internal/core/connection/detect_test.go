package connection

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// --- GH Detector Tests ---

func TestGHDetector_Detected(t *testing.T) {
	t.Parallel()

	d := &GHDetector{
		lookPathFn: func(name string) (string, error) {
			if name == "gh" {
				return "/usr/local/bin/gh", nil
			}
			return "", fmt.Errorf("not found")
		},
		runCmdFn: func(name string, args ...string) ([]byte, error) {
			return []byte("Logged in to github.com"), nil
		},
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.ProviderName != "github" {
		t.Errorf("ProviderName = %q, want %q", result.ProviderName, "github")
	}
	if result.Reason != "gh CLI authenticated" {
		t.Errorf("Reason = %q, want %q", result.Reason, "gh CLI authenticated")
	}
}

func TestGHDetector_InstalledButNotAuthenticated(t *testing.T) {
	t.Parallel()

	d := &GHDetector{
		lookPathFn: func(name string) (string, error) {
			return "/usr/local/bin/gh", nil
		},
		runCmdFn: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("not authenticated")
		},
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when gh not authenticated, got %+v", result)
	}
}

func TestGHDetector_NotInstalled(t *testing.T) {
	t.Parallel()

	d := &GHDetector{
		lookPathFn: func(name string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when gh not installed, got %+v", result)
	}
}

func TestGHDetector_AuthOutputWithoutLoggedIn(t *testing.T) {
	t.Parallel()

	d := &GHDetector{
		lookPathFn: func(name string) (string, error) {
			return "/usr/local/bin/gh", nil
		},
		runCmdFn: func(name string, args ...string) ([]byte, error) {
			return []byte("some other output"), nil
		},
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.Reason != "gh CLI" {
		t.Errorf("Reason = %q, want %q", result.Reason, "gh CLI")
	}
}

// --- Todoist Detector Tests ---

func TestTodoistDetector_TokenFound(t *testing.T) {
	t.Parallel()

	d := &TodoistDetector{
		lookupEnvFn: func(key string) (string, bool) {
			if key == "TODOIST_API_TOKEN" {
				return "test-token-123", true
			}
			return "", false
		},
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.ProviderName != "todoist" {
		t.Errorf("ProviderName = %q, want %q", result.ProviderName, "todoist")
	}
	if result.Reason != "API token found" {
		t.Errorf("Reason = %q, want %q", result.Reason, "API token found")
	}
	if result.PreFill["token_env"] != "TODOIST_API_TOKEN" {
		t.Errorf("PreFill[token_env] = %q, want %q", result.PreFill["token_env"], "TODOIST_API_TOKEN")
	}
}

func TestTodoistDetector_ThreeDoorsTokenFound(t *testing.T) {
	t.Parallel()

	d := &TodoistDetector{
		lookupEnvFn: func(key string) (string, bool) {
			if key == "THREEDOORS_TODOIST_TOKEN" {
				return "td-token", true
			}
			return "", false
		},
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.PreFill["token_env"] != "THREEDOORS_TODOIST_TOKEN" {
		t.Errorf("PreFill[token_env] = %q, want %q", result.PreFill["token_env"], "THREEDOORS_TODOIST_TOKEN")
	}
}

func TestTodoistDetector_NoToken(t *testing.T) {
	t.Parallel()

	d := &TodoistDetector{
		lookupEnvFn: func(key string) (string, bool) {
			return "", false
		},
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when no token found, got %+v", result)
	}
}

func TestTodoistDetector_EmptyToken(t *testing.T) {
	t.Parallel()

	d := &TodoistDetector{
		lookupEnvFn: func(key string) (string, bool) {
			if key == "TODOIST_API_TOKEN" {
				return "", true
			}
			return "", false
		},
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when token is empty string, got %+v", result)
	}
}

func TestTodoistDetector_PriorityOrder(t *testing.T) {
	t.Parallel()

	d := &TodoistDetector{
		lookupEnvFn: func(key string) (string, bool) {
			switch key {
			case "TODOIST_API_TOKEN":
				return "primary-token", true
			case "THREEDOORS_TODOIST_TOKEN":
				return "secondary-token", true
			}
			return "", false
		},
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	// TODOIST_API_TOKEN should be preferred
	if result.PreFill["token_env"] != "TODOIST_API_TOKEN" {
		t.Errorf("should prefer TODOIST_API_TOKEN, got %q", result.PreFill["token_env"])
	}
}

// --- Obsidian Detector Tests ---

func TestObsidianDetector_VaultFound(t *testing.T) {
	t.Parallel()

	// Create a temp directory structure simulating an Obsidian vault
	tmpDir := t.TempDir()
	vaultDir := filepath.Join(tmpDir, "Documents", "MyVault")
	obsidianDir := filepath.Join(vaultDir, ".obsidian")
	if err := os.MkdirAll(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}

	d := &ObsidianDetector{
		homeDirFn: func() (string, error) { return tmpDir, nil },
		statFn:    os.Stat,
		globFn:    filepath.Glob,
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.ProviderName != "obsidian" {
		t.Errorf("ProviderName = %q, want %q", result.ProviderName, "obsidian")
	}
	if result.PreFill["path"] != vaultDir {
		t.Errorf("PreFill[path] = %q, want %q", result.PreFill["path"], vaultDir)
	}
	expectedReason := fmt.Sprintf("vault found at %s", vaultDir)
	if result.Reason != expectedReason {
		t.Errorf("Reason = %q, want %q", result.Reason, expectedReason)
	}
}

func TestObsidianDetector_NoVault(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create Documents but no vault
	if err := os.MkdirAll(filepath.Join(tmpDir, "Documents"), 0o755); err != nil {
		t.Fatal(err)
	}

	d := &ObsidianDetector{
		homeDirFn: func() (string, error) { return tmpDir, nil },
		statFn:    os.Stat,
		globFn:    filepath.Glob,
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when no vault found, got %+v", result)
	}
}

func TestObsidianDetector_HomeDirError(t *testing.T) {
	t.Parallel()

	d := &ObsidianDetector{
		homeDirFn: func() (string, error) { return "", fmt.Errorf("no home") },
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil on home dir error, got %+v", result)
	}
}

func TestObsidianDetector_VaultOnDesktop(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create empty Documents
	if err := os.MkdirAll(filepath.Join(tmpDir, "Documents"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Vault on Desktop
	vaultDir := filepath.Join(tmpDir, "Desktop", "Notes")
	obsidianDir := filepath.Join(vaultDir, ".obsidian")
	if err := os.MkdirAll(obsidianDir, 0o755); err != nil {
		t.Fatal(err)
	}

	d := &ObsidianDetector{
		homeDirFn: func() (string, error) { return tmpDir, nil },
		statFn:    os.Stat,
		globFn:    filepath.Glob,
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result for Desktop vault, got nil")
	}
	if result.PreFill["path"] != vaultDir {
		t.Errorf("PreFill[path] = %q, want %q", result.PreFill["path"], vaultDir)
	}
}

// --- Jira Detector Tests ---

func TestJiraDetector_EnvVarFound(t *testing.T) {
	t.Parallel()

	d := &JiraDetector{
		lookupEnvFn: func(key string) (string, bool) {
			if key == "JIRA_API_TOKEN" {
				return "jira-tok-123", true
			}
			return "", false
		},
		homeDirFn: func() (string, error) { return "/nonexistent", nil },
		statFn:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.ProviderName != "jira" {
		t.Errorf("ProviderName = %q, want %q", result.ProviderName, "jira")
	}
	if result.Reason != "API token found" {
		t.Errorf("Reason = %q, want %q", result.Reason, "API token found")
	}
	if result.PreFill["token_env"] != "JIRA_API_TOKEN" {
		t.Errorf("PreFill[token_env] = %q, want %q", result.PreFill["token_env"], "JIRA_API_TOKEN")
	}
}

func TestJiraDetector_ConfigDirFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	jiraDir := filepath.Join(tmpDir, ".jira.d")
	if err := os.MkdirAll(jiraDir, 0o755); err != nil {
		t.Fatal(err)
	}

	d := &JiraDetector{
		lookupEnvFn: func(key string) (string, bool) { return "", false },
		homeDirFn:   func() (string, error) { return tmpDir, nil },
		statFn:      os.Stat,
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
	if result.ProviderName != "jira" {
		t.Errorf("ProviderName = %q, want %q", result.ProviderName, "jira")
	}
	expectedReason := fmt.Sprintf("config found at %s", jiraDir)
	if result.Reason != expectedReason {
		t.Errorf("Reason = %q, want %q", result.Reason, expectedReason)
	}
}

func TestJiraDetector_AtlassianDirFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	atlDir := filepath.Join(tmpDir, ".atlassian")
	if err := os.MkdirAll(atlDir, 0o755); err != nil {
		t.Fatal(err)
	}

	d := &JiraDetector{
		lookupEnvFn: func(key string) (string, bool) { return "", false },
		homeDirFn:   func() (string, error) { return tmpDir, nil },
		statFn:      os.Stat,
	}

	result := d.Detect()
	if result == nil {
		t.Fatal("expected detection result, got nil")
	}
}

func TestJiraDetector_NothingDetected(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	d := &JiraDetector{
		lookupEnvFn: func(key string) (string, bool) { return "", false },
		homeDirFn:   func() (string, error) { return tmpDir, nil },
		statFn:      os.Stat,
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil when nothing detected, got %+v", result)
	}
}

func TestJiraDetector_HomeDirError(t *testing.T) {
	t.Parallel()

	d := &JiraDetector{
		lookupEnvFn: func(key string) (string, bool) { return "", false },
		homeDirFn:   func() (string, error) { return "", fmt.Errorf("no home") },
		statFn:      os.Stat,
	}

	result := d.Detect()
	if result != nil {
		t.Errorf("expected nil on home dir error, got %+v", result)
	}
}

// --- DetectAll Tests ---

func TestDetectAll_NoDetectors(t *testing.T) {
	t.Parallel()

	results := DetectAll(nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDetectAll_AllDetected(t *testing.T) {
	t.Parallel()

	detectors := []Detector{
		&GHDetector{
			lookPathFn: func(string) (string, error) { return "/usr/bin/gh", nil },
			runCmdFn: func(string, ...string) ([]byte, error) {
				return []byte("Logged in to github.com"), nil
			},
		},
		&TodoistDetector{
			lookupEnvFn: func(key string) (string, bool) {
				if key == "TODOIST_API_TOKEN" {
					return "tok", true
				}
				return "", false
			},
		},
	}

	results := DetectAll(detectors)
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestDetectAll_NoneDetected(t *testing.T) {
	t.Parallel()

	detectors := []Detector{
		&GHDetector{
			lookPathFn: func(string) (string, error) { return "", fmt.Errorf("not found") },
		},
		&TodoistDetector{
			lookupEnvFn: func(string) (string, bool) { return "", false },
		},
	}

	results := DetectAll(detectors)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestDetectAll_PartialDetection(t *testing.T) {
	t.Parallel()

	detectors := []Detector{
		&GHDetector{
			lookPathFn: func(string) (string, error) { return "", fmt.Errorf("not found") },
		},
		&TodoistDetector{
			lookupEnvFn: func(key string) (string, bool) {
				if key == "TODOIST_API_TOKEN" {
					return "tok", true
				}
				return "", false
			},
		},
	}

	results := DetectAll(detectors)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ProviderName != "todoist" {
		t.Errorf("expected todoist, got %q", results[0].ProviderName)
	}
}

// --- DefaultDetectors Tests ---

func TestDefaultDetectors_ReturnsAllDetectors(t *testing.T) {
	t.Parallel()

	detectors := DefaultDetectors()
	if len(detectors) != 4 {
		t.Errorf("expected 4 default detectors, got %d", len(detectors))
	}
}

// --- DetectionResult Tests ---

func TestDetectionResult_Fields(t *testing.T) {
	t.Parallel()

	r := DetectionResult{
		ProviderName: "github",
		Reason:       "gh CLI authenticated",
		PreFill:      map[string]string{"token_env": "GH_TOKEN"},
	}

	if r.ProviderName != "github" {
		t.Errorf("ProviderName = %q, want %q", r.ProviderName, "github")
	}
	if r.Reason != "gh CLI authenticated" {
		t.Errorf("Reason = %q, want %q", r.Reason, "gh CLI authenticated")
	}
	if r.PreFill["token_env"] != "GH_TOKEN" {
		t.Errorf("PreFill[token_env] = %q, want %q", r.PreFill["token_env"], "GH_TOKEN")
	}
}
