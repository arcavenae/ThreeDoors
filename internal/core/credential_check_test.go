package core

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/intelligence/llm"
)

func TestWarnCredentialExposure_WorldReadableWithTokens(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name: "jira",
				Settings: map[string]string{
					"url":       "https://test.atlassian.net",
					"api_token": "secret-token-123",
				},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if !warned {
		t.Error("expected warning for world-readable config with tokens")
	}
	output := buf.String()
	if !strings.Contains(output, "WARNING") {
		t.Errorf("output should contain WARNING, got: %s", output)
	}
	if !strings.Contains(output, "chmod 600") {
		t.Errorf("output should contain remediation instructions, got: %s", output)
	}
	if !strings.Contains(output, configPath) {
		t.Errorf("output should contain config path, got: %s", output)
	}
}

func TestWarnCredentialExposure_SecurePermissions(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name: "jira",
				Settings: map[string]string{
					"api_token": "secret-token-123",
				},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if warned {
		t.Error("should not warn when file has secure permissions")
	}
	if buf.Len() != 0 {
		t.Errorf("should not produce output, got: %s", buf.String())
	}
}

func TestWarnCredentialExposure_WorldReadableNoTokens(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name: "textfile",
				Settings: map[string]string{
					"task_file": "~/tasks.yaml",
				},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if warned {
		t.Error("should not warn when config has no credentials")
	}
}

func TestWarnCredentialExposure_EmptyTokenNotCounted(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name: "jira",
				Settings: map[string]string{
					"api_token": "",
				},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if warned {
		t.Error("should not warn when token fields are empty")
	}
}

func TestWarnCredentialExposure_NonexistentFile(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{}
	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, "/nonexistent/config.yaml", cfg)

	if warned {
		t.Error("should not warn for nonexistent file")
	}
}

func TestWarnCredentialExposure_NilConfig(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, "/some/path", nil)

	if warned {
		t.Error("should not warn for nil config")
	}
}

func TestWarnCredentialExposure_MultipleProviders(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name:     "textfile",
				Settings: map[string]string{"task_file": "~/tasks.yaml"},
			},
			{
				Name: "todoist",
				Settings: map[string]string{
					"api_token": "todoist-token",
				},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if !warned {
		t.Error("should warn when any provider has credentials and file is permissive")
	}
}

func TestWarnCredentialExposure_GroupReadable(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("unix permissions not supported on Windows")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o640); err != nil {
		t.Fatal(err)
	}

	cfg := &ProviderConfig{
		Providers: []ProviderEntry{
			{
				Name:     "github",
				Settings: map[string]string{"token": "gh-token"},
			},
		},
	}

	var buf bytes.Buffer
	warned := WarnCredentialExposure(&buf, configPath, cfg)

	if !warned {
		t.Error("should warn when file is group-readable and has credentials")
	}
}

func TestConfigHasCredentials_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		key    string
		value  string
		expect bool
	}{
		{"lowercase api_token", "api_token", "val", true},
		{"uppercase API_TOKEN", "API_TOKEN", "val", true},
		{"mixed Api_Token", "Api_Token", "val", true},
		{"token key", "token", "val", true},
		{"password key", "password", "val", true},
		{"secret key", "secret", "val", true},
		{"api_key", "api_key", "val", true},
		{"non-credential key", "url", "https://example.com", false},
		{"empty value", "api_token", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &ProviderConfig{
				Providers: []ProviderEntry{
					{
						Name:     "test",
						Settings: map[string]string{tt.key: tt.value},
					},
				},
			}
			got := configHasCredentials(cfg)
			if got != tt.expect {
				t.Errorf("configHasCredentials() with key=%q value=%q = %v, want %v",
					tt.key, tt.value, got, tt.expect)
			}
		})
	}
}

func TestYamlDashPreventsTokenSerialization(t *testing.T) {
	t.Parallel()

	cfg := &ProviderConfig{
		SchemaVersion: CurrentSchemaVersion,
		Provider:      "textfile",
		LLM: llm.Config{
			Claude: llm.ClaudeConfig{APIKey: "should-not-appear"},
		},
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := SaveProviderConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "should-not-appear") {
		t.Error("yaml:\"-\" field should not be serialized, but API key appeared in output")
	}
}
