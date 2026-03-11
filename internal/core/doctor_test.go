package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckStatus_Ordering(t *testing.T) {
	t.Parallel()

	// Verify severity ordering: FAIL > WARN > SKIP > INFO > OK
	tests := []struct {
		name   string
		a, b   CheckStatus
		aWorse bool
	}{
		{"FAIL > WARN", CheckFail, CheckWarn, true},
		{"WARN > SKIP", CheckWarn, CheckSkip, true},
		{"SKIP > INFO", CheckSkip, CheckInfo, true},
		{"INFO > OK", CheckInfo, CheckOK, true},
		{"FAIL > OK", CheckFail, CheckOK, true},
		{"OK < FAIL", CheckOK, CheckFail, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.a > tt.b
			if got != tt.aWorse {
				t.Errorf("%s > %s = %v, want %v", tt.a, tt.b, got, tt.aWorse)
			}
		})
	}
}

func TestCheckStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status CheckStatus
		want   string
	}{
		{CheckOK, "OK"},
		{CheckInfo, "INFO"},
		{CheckSkip, "SKIP"},
		{CheckWarn, "WARN"},
		{CheckFail, "FAIL"},
		{CheckStatus(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.String(); got != tt.want {
				t.Errorf("CheckStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestCheckStatus_Icon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status CheckStatus
		want   string
	}{
		{CheckOK, "[✓]"},
		{CheckInfo, "[i]"},
		{CheckSkip, "[ ]"},
		{CheckWarn, "[!]"},
		{CheckFail, "[✗]"},
		{CheckStatus(99), "[?]"},
	}
	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			t.Parallel()
			if got := tt.status.Icon(); got != tt.want {
				t.Errorf("CheckStatus(%d).Icon() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestWorstCheckStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		checks []CheckResult
		want   CheckStatus
	}{
		{
			name:   "empty list returns OK",
			checks: nil,
			want:   CheckOK,
		},
		{
			name: "all OK",
			checks: []CheckResult{
				{Status: CheckOK},
				{Status: CheckOK},
			},
			want: CheckOK,
		},
		{
			name: "one WARN among OK",
			checks: []CheckResult{
				{Status: CheckOK},
				{Status: CheckWarn},
				{Status: CheckOK},
			},
			want: CheckWarn,
		},
		{
			name: "FAIL trumps WARN",
			checks: []CheckResult{
				{Status: CheckWarn},
				{Status: CheckFail},
				{Status: CheckOK},
			},
			want: CheckFail,
		},
		{
			name: "INFO does not override WARN",
			checks: []CheckResult{
				{Status: CheckInfo},
				{Status: CheckWarn},
			},
			want: CheckWarn,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := worstCheckStatus(tt.checks)
			if got != tt.want {
				t.Errorf("worstCheckStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoctorChecker_ConfigDirMissing(t *testing.T) {
	dc := NewDoctorChecker("/nonexistent/path/that/does/not/exist")
	result := dc.Run()

	if len(result.Categories) < 1 {
		t.Fatalf("expected at least 1 category, got %d", len(result.Categories))
	}

	env := result.Categories[0]
	if env.Name != "Environment" {
		t.Errorf("category name = %q, want %q", env.Name, "Environment")
	}
	if env.Status != CheckFail {
		t.Errorf("category status = %v, want %v", env.Status, CheckFail)
	}

	// Config dir check should fail
	if len(env.Checks) < 1 {
		t.Fatal("expected at least 1 check")
	}
	if env.Checks[0].Status != CheckFail {
		t.Errorf("config dir check status = %v, want %v", env.Checks[0].Status, CheckFail)
	}
	if env.Checks[0].Message != "Config directory not found" {
		t.Errorf("config dir message = %q, want %q", env.Checks[0].Message, "Config directory not found")
	}
	if env.Checks[0].Suggestion != "Run threedoors to create it during onboarding" {
		t.Errorf("config dir suggestion = %q, want %q", env.Checks[0].Suggestion, "Run threedoors to create it during onboarding")
	}
}

func TestDoctorChecker_ConfigDirExistsNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]

	// Config dir should pass
	if env.Checks[0].Status != CheckOK {
		t.Errorf("config dir check = %v, want %v (message: %s)", env.Checks[0].Status, CheckOK, env.Checks[0].Message)
	}

	// Config file should warn (missing)
	if env.Checks[1].Status != CheckWarn {
		t.Errorf("config file check = %v, want %v (message: %s)", env.Checks[1].Status, CheckWarn, env.Checks[1].Message)
	}
}

func TestDoctorChecker_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := fmt.Sprintf("schema_version: %d\nprovider: textfile\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	for _, check := range env.Checks {
		if check.Status == CheckFail {
			t.Errorf("check %q failed: %s", check.Name, check.Message)
		}
	}
}

func TestDoctorChecker_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("{{{{not yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	configFileCheck := env.Checks[1]
	if configFileCheck.Status != CheckFail {
		t.Errorf("invalid YAML check = %v, want %v", configFileCheck.Status, CheckFail)
	}
}

func TestDoctorChecker_OldSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte("schema_version: 1\nprovider: textfile\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	configFileCheck := env.Checks[1]
	if configFileCheck.Status != CheckInfo {
		t.Errorf("old schema version check = %v, want %v (message: %s)", configFileCheck.Status, CheckInfo, configFileCheck.Message)
	}
}

func TestDoctorChecker_ReadOnlyDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("test requires non-root user")
	}
	tmpDir := t.TempDir()
	roDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(roDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roDir, 0o755) })

	dc := NewDoctorChecker(roDir)
	result := dc.Run()

	env := result.Categories[0]
	if env.Checks[0].Status != CheckWarn {
		t.Errorf("readonly dir check = %v, want %v (message: %s)", env.Checks[0].Status, CheckWarn, env.Checks[0].Message)
	}
}

func TestDoctorResult_IssueCount(t *testing.T) {
	t.Parallel()

	result := DoctorResult{
		Categories: []CategoryResult{
			{
				Checks: []CheckResult{
					{Status: CheckOK},
					{Status: CheckWarn},
					{Status: CheckFail},
				},
			},
			{
				Checks: []CheckResult{
					{Status: CheckOK},
					{Status: CheckWarn},
				},
			},
		},
	}

	warnings, errors := result.IssueCount()
	if warnings != 2 {
		t.Errorf("warnings = %d, want 2", warnings)
	}
	if errors != 1 {
		t.Errorf("errors = %d, want 1", errors)
	}
}

func TestDoctorResult_CategoryIssueCount(t *testing.T) {
	t.Parallel()

	result := DoctorResult{
		Categories: []CategoryResult{
			{Checks: []CheckResult{{Status: CheckOK}}},
			{Checks: []CheckResult{{Status: CheckWarn}}},
			{Checks: []CheckResult{{Status: CheckFail}}},
		},
	}

	if got := result.CategoryIssueCount(); got != 2 {
		t.Errorf("CategoryIssueCount() = %d, want 2", got)
	}
}

func TestDoctorResult_OverallStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		categories []CategoryResult
		want       CheckStatus
	}{
		{
			name:       "no categories",
			categories: nil,
			want:       CheckOK,
		},
		{
			name: "all OK",
			categories: []CategoryResult{
				{Status: CheckOK},
				{Status: CheckOK},
			},
			want: CheckOK,
		},
		{
			name: "one WARN",
			categories: []CategoryResult{
				{Status: CheckOK},
				{Status: CheckWarn},
			},
			want: CheckWarn,
		},
		{
			name: "FAIL trumps all",
			categories: []CategoryResult{
				{Status: CheckWarn},
				{Status: CheckFail},
			},
			want: CheckFail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := DoctorResult{Categories: tt.categories}
			if got := r.OverallStatus(); got != tt.want {
				t.Errorf("OverallStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoctorChecker_RegisterCategory(t *testing.T) {
	dc := NewDoctorChecker(t.TempDir())

	// Environment is registered by default
	called := false
	dc.RegisterCategory("Custom", func() CategoryResult {
		called = true
		return CategoryResult{
			Checks: []CheckResult{{Status: CheckOK, Message: "custom check"}},
		}
	})

	result := dc.Run()

	if !called {
		t.Error("custom category was not called")
	}
	// Find the Custom category (it should be the last one registered)
	lastCat := result.Categories[len(result.Categories)-1]
	if lastCat.Name != "Custom" {
		t.Errorf("last category name = %q, want %q", lastCat.Name, "Custom")
	}
}

func TestDoctorChecker_CategoriesRunInOrder(t *testing.T) {
	dc := &DoctorChecker{configDir: t.TempDir()}
	var order []string

	dc.RegisterCategory("First", func() CategoryResult {
		order = append(order, "First")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
	})
	dc.RegisterCategory("Second", func() CategoryResult {
		order = append(order, "Second")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
	})
	dc.RegisterCategory("Third", func() CategoryResult {
		order = append(order, "Third")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
	})

	dc.Run()

	if len(order) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(order))
	}
	for i, want := range []string{"First", "Second", "Third"} {
		if order[i] != want {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want)
		}
	}
}

func TestDoctorResult_Duration(t *testing.T) {
	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	if result.Duration <= 0 {
		t.Errorf("Duration = %v, want > 0", result.Duration)
	}
}

func TestDoctorChecker_UnsupportedSchemaVersion(t *testing.T) {
	tmpDir := t.TempDir()
	content := fmt.Sprintf("schema_version: %d\nprovider: textfile\n", CurrentSchemaVersion+1)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	configFileCheck := env.Checks[1]
	if configFileCheck.Status != CheckFail {
		t.Errorf("unsupported schema check = %v, want %v (message: %s)", configFileCheck.Status, CheckFail, configFileCheck.Message)
	}
	wantMsg := fmt.Sprintf("Unsupported config schema version %d (max supported: %d)", CurrentSchemaVersion+1, CurrentSchemaVersion)
	if configFileCheck.Message != wantMsg {
		t.Errorf("message = %q, want %q", configFileCheck.Message, wantMsg)
	}
	if configFileCheck.Suggestion != "Update ThreeDoors to the latest version" {
		t.Errorf("suggestion = %q, want %q", configFileCheck.Suggestion, "Update ThreeDoors to the latest version")
	}
}

func TestDoctorChecker_MissingProviderField(t *testing.T) {
	tmpDir := t.TempDir()
	content := fmt.Sprintf("schema_version: %d\ntheme: classic\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	configFileCheck := env.Checks[1]
	if configFileCheck.Status != CheckWarn {
		t.Errorf("missing provider check = %v, want %v (message: %s)", configFileCheck.Status, CheckWarn, configFileCheck.Message)
	}
	if configFileCheck.Message != "Config file missing required field: provider" {
		t.Errorf("message = %q, want %q", configFileCheck.Message, "Config file missing required field: provider")
	}
}

func TestDoctorChecker_ProvidersListSatisfiesRequiredField(t *testing.T) {
	tmpDir := t.TempDir()
	content := fmt.Sprintf("schema_version: %d\nproviders:\n  - name: textfile\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	env := result.Categories[0]
	configFileCheck := env.Checks[1]
	if configFileCheck.Status != CheckOK {
		t.Errorf("providers list check = %v, want %v (message: %s)", configFileCheck.Status, CheckOK, configFileCheck.Message)
	}
}

func TestDoctorChecker_TerminalWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		info       TerminalInfo
		wantStatus CheckStatus
		wantMsg    string
	}{
		{
			name:       "unknown terminal",
			info:       TerminalInfo{},
			wantStatus: CheckInfo,
			wantMsg:    "Terminal size not available",
		},
		{
			name:       "narrow terminal",
			info:       TerminalInfo{Width: 30, Height: 24},
			wantStatus: CheckWarn,
			wantMsg:    "Narrow terminal (30 columns) may affect theme display",
		},
		{
			name:       "exactly 40 columns",
			info:       TerminalInfo{Width: 40, Height: 24},
			wantStatus: CheckOK,
			wantMsg:    "Terminal size 40×24",
		},
		{
			name:       "wide terminal",
			info:       TerminalInfo{Width: 120, Height: 40},
			wantStatus: CheckOK,
			wantMsg:    "Terminal size 120×40",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := NewDoctorChecker(t.TempDir())
			dc.SetTerminalInfo(tt.info)
			result := dc.Run()

			env := result.Categories[0]
			// Terminal size is check index 2 (after config dir and config file)
			var termCheck CheckResult
			for _, c := range env.Checks {
				if c.Name == "Terminal size" {
					termCheck = c
					break
				}
			}
			if termCheck.Name == "" {
				t.Fatal("Terminal size check not found")
			}
			if termCheck.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", termCheck.Status, tt.wantStatus)
			}
			if termCheck.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", termCheck.Message, tt.wantMsg)
			}
		})
	}
}

func TestDoctorChecker_ColorProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		profile    string
		wantStatus CheckStatus
		wantMsg    string
	}{
		{
			name:       "unknown profile",
			profile:    "",
			wantStatus: CheckInfo,
			wantMsg:    "Color profile not available",
		},
		{
			name:       "ascii only",
			profile:    "Ascii",
			wantStatus: CheckInfo,
			wantMsg:    "Terminal supports ASCII only — colors will be degraded",
		},
		{
			name:       "ansi256",
			profile:    "ANSI256",
			wantStatus: CheckOK,
			wantMsg:    "Color profile: ANSI256",
		},
		{
			name:       "truecolor",
			profile:    "TrueColor",
			wantStatus: CheckOK,
			wantMsg:    "Color profile: TrueColor",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dc := NewDoctorChecker(t.TempDir())
			dc.SetTerminalInfo(TerminalInfo{ColorProfile: tt.profile})
			result := dc.Run()

			env := result.Categories[0]
			var colorCheck CheckResult
			for _, c := range env.Checks {
				if c.Name == "Color support" {
					colorCheck = c
					break
				}
			}
			if colorCheck.Name == "" {
				t.Fatal("Color support check not found")
			}
			if colorCheck.Status != tt.wantStatus {
				t.Errorf("status = %v, want %v", colorCheck.Status, tt.wantStatus)
			}
			if colorCheck.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", colorCheck.Message, tt.wantMsg)
			}
		})
	}
}

func TestDoctorChecker_GoVersion(t *testing.T) {
	t.Parallel()

	dc := NewDoctorChecker(t.TempDir())
	result := dc.Run()

	env := result.Categories[0]
	var goCheck CheckResult
	for _, c := range env.Checks {
		if c.Name == "Go runtime" {
			goCheck = c
			break
		}
	}
	if goCheck.Name == "" {
		t.Fatal("Go runtime check not found")
	}
	if goCheck.Status != CheckInfo {
		t.Errorf("status = %v, want %v", goCheck.Status, CheckInfo)
	}
	if goCheck.Message == "" {
		t.Error("Go runtime message is empty")
	}
}

func TestDoctorChecker_EnvironmentCheckCount(t *testing.T) {
	tmpDir := t.TempDir()
	content := fmt.Sprintf("schema_version: %d\nprovider: textfile\n", CurrentSchemaVersion)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := NewDoctorChecker(tmpDir)
	dc.SetTerminalInfo(TerminalInfo{Width: 80, Height: 24, ColorProfile: "TrueColor"})
	result := dc.Run()

	env := result.Categories[0]
	// Expected checks: Config directory, Config file, Terminal size, Color support, Go runtime
	if len(env.Checks) != 5 {
		t.Errorf("expected 5 environment checks, got %d", len(env.Checks))
		for i, c := range env.Checks {
			t.Logf("  check[%d]: %s (%s) — %s", i, c.Name, c.Status, c.Message)
		}
	}
}

func TestRunWithOptions_CategoryFilter(t *testing.T) {
	t.Parallel()

	dc := &DoctorChecker{configDir: t.TempDir()}
	var called []string
	dc.RegisterCategory("Environment", func() CategoryResult {
		called = append(called, "Environment")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK, Message: "ok"}}}
	})
	dc.RegisterCategory("Task Data", func() CategoryResult {
		called = append(called, "Task Data")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK, Message: "ok"}}}
	})
	dc.RegisterCategory("Version", func() CategoryResult {
		called = append(called, "Version")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK, Message: "ok"}}}
	})

	result := dc.RunWithOptions(DoctorOptions{Categories: []string{"env"}})

	// Only Environment should have been called
	if len(called) != 1 || called[0] != "Environment" {
		t.Errorf("called = %v, want [Environment]", called)
	}

	// All categories should appear in result
	if len(result.Categories) != 3 {
		t.Fatalf("got %d categories, want 3", len(result.Categories))
	}

	// Environment should be OK, others should be SKIP
	if result.Categories[0].Status != CheckOK {
		t.Errorf("Environment status = %v, want OK", result.Categories[0].Status)
	}
	if result.Categories[1].Status != CheckSkip {
		t.Errorf("Task Data status = %v, want SKIP", result.Categories[1].Status)
	}
	if result.Categories[2].Status != CheckSkip {
		t.Errorf("Version status = %v, want SKIP", result.Categories[2].Status)
	}
}

func TestRunWithOptions_MultipleCategoryFilter(t *testing.T) {
	t.Parallel()

	dc := &DoctorChecker{configDir: t.TempDir()}
	var called []string
	dc.RegisterCategory("Environment", func() CategoryResult {
		called = append(called, "Environment")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
	})
	dc.RegisterCategory("Sync", func() CategoryResult {
		called = append(called, "Sync")
		return CategoryResult{Checks: []CheckResult{{Status: CheckWarn, Message: "stale"}}}
	})
	dc.RegisterCategory("Version", func() CategoryResult {
		called = append(called, "Version")
		return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
	})

	result := dc.RunWithOptions(DoctorOptions{Categories: []string{"env", "sync"}})

	if len(called) != 2 {
		t.Errorf("called = %v, want [Environment Sync]", called)
	}
	if result.Categories[2].Status != CheckSkip {
		t.Errorf("Version should be skipped, got %v", result.Categories[2].Status)
	}
}

func TestRunWithOptions_EmptyFilter_RunsAll(t *testing.T) {
	t.Parallel()

	dc := &DoctorChecker{configDir: t.TempDir()}
	callCount := 0
	for _, name := range []string{"A", "B", "C"} {
		name := name
		dc.RegisterCategory(name, func() CategoryResult {
			callCount++
			return CategoryResult{Checks: []CheckResult{{Status: CheckOK}}}
		})
	}

	dc.RunWithOptions(DoctorOptions{})
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestHasFixableIssues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result DoctorResult
		want   bool
	}{
		{
			name: "no issues",
			result: DoctorResult{Categories: []CategoryResult{
				{Status: CheckOK, Checks: []CheckResult{{Status: CheckOK}}},
			}},
			want: false,
		},
		{
			name: "warning with suggestion",
			result: DoctorResult{Categories: []CategoryResult{
				{Status: CheckWarn, Checks: []CheckResult{
					{Status: CheckWarn, Message: "stale", Suggestion: "Run sync"},
				}},
			}},
			want: true,
		},
		{
			name: "warning without suggestion",
			result: DoctorResult{Categories: []CategoryResult{
				{Status: CheckWarn, Checks: []CheckResult{
					{Status: CheckWarn, Message: "stale"},
				}},
			}},
			want: false,
		},
		{
			name: "fail with suggestion",
			result: DoctorResult{Categories: []CategoryResult{
				{Status: CheckFail, Checks: []CheckResult{
					{Status: CheckFail, Message: "missing", Suggestion: "Create it"},
				}},
			}},
			want: true,
		},
		{
			name: "skipped category excluded",
			result: DoctorResult{Categories: []CategoryResult{
				{Status: CheckSkip, Checks: []CheckResult{
					{Status: CheckWarn, Suggestion: "fix it"},
				}},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.result.HasFixableIssues()
			if got != tt.want {
				t.Errorf("HasFixableIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssueCount_SkipsSkippedCategories(t *testing.T) {
	t.Parallel()

	result := DoctorResult{Categories: []CategoryResult{
		{Status: CheckWarn, Checks: []CheckResult{
			{Status: CheckWarn},
		}},
		{Status: CheckSkip, Checks: []CheckResult{
			{Status: CheckFail}, // should be excluded
		}},
	}}

	warnings, errors := result.IssueCount()
	if warnings != 1 {
		t.Errorf("warnings = %d, want 1", warnings)
	}
	if errors != 0 {
		t.Errorf("errors = %d, want 0", errors)
	}
}

func TestValidCategoryKeys(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"env": "Environment", "tasks": "Task Data", "providers": "Providers",
		"sessions": "Session Data", "sync": "Sync", "db": "Database", "version": "Version",
	}
	for key, want := range expected {
		got, ok := ValidCategoryKeys[key]
		if !ok {
			t.Errorf("ValidCategoryKeys missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("ValidCategoryKeys[%q] = %q, want %q", key, got, want)
		}
	}
}
