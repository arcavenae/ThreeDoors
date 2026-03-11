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
		t.Fatal("expected at least 1 category")
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
