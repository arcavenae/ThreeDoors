package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core/connection"
)

func TestBuildConnectSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		provider   string
		token      string
		flags      map[string]string
		wantErr    string
		wantKeys   map[string]string
		wantAbsent []string
	}{
		{
			name:     "todoist with token",
			provider: "todoist",
			token:    "tok_123",
			flags:    map[string]string{},
			wantKeys: map[string]string{"api_token": "tok_123"},
		},
		{
			name:     "todoist with project-ids and filter",
			provider: "todoist",
			token:    "tok_abc",
			flags:    map[string]string{"project-ids": "123,456", "filter": "today"},
			wantKeys: map[string]string{
				"api_token":   "tok_abc",
				"project_ids": "123,456",
				"filter":      "today",
			},
		},
		{
			name:       "github with repos",
			provider:   "github",
			token:      "",
			flags:      map[string]string{"repos": "owner/repo1,owner/repo2"},
			wantKeys:   map[string]string{"repos": "owner/repo1,owner/repo2"},
			wantAbsent: []string{"token"},
		},
		{
			name:     "github with repos and token",
			provider: "github",
			token:    "ghp_abc",
			flags:    map[string]string{"repos": "owner/repo"},
			wantKeys: map[string]string{
				"repos": "owner/repo",
				"token": "ghp_abc",
			},
		},
		{
			name:     "github missing repos",
			provider: "github",
			token:    "ghp_abc",
			flags:    map[string]string{},
			wantErr:  "missing required flags for github: --repos",
		},
		{
			name:     "jira with server and token",
			provider: "jira",
			token:    "jira_token",
			flags:    map[string]string{"server": "https://jira.example.com"},
			wantKeys: map[string]string{
				"url":       "https://jira.example.com",
				"api_token": "jira_token",
			},
		},
		{
			name:     "jira missing server",
			provider: "jira",
			token:    "jira_token",
			flags:    map[string]string{},
			wantErr:  "missing required flags for jira: --server",
		},
		{
			name:     "textfile with path",
			provider: "textfile",
			token:    "",
			flags:    map[string]string{"path": "~/tasks.yaml"},
			wantKeys: map[string]string{"path": "~/tasks.yaml"},
		},
		{
			name:     "textfile missing path",
			provider: "textfile",
			token:    "",
			flags:    map[string]string{},
			wantErr:  "missing required flags for textfile: --path",
		},
		{
			name:     "applenotes with note-title",
			provider: "applenotes",
			token:    "",
			flags:    map[string]string{"note-title": "My Tasks"},
			wantKeys: map[string]string{"note_title": "My Tasks"},
		},
		{
			name:     "applenotes missing note-title",
			provider: "applenotes",
			token:    "",
			flags:    map[string]string{},
			wantErr:  "missing required flags for applenotes: --note-title",
		},
		{
			name:     "obsidian with path",
			provider: "obsidian",
			token:    "",
			flags:    map[string]string{"path": "~/Documents/vault"},
			wantKeys: map[string]string{"vault_path": "~/Documents/vault"},
		},
		{
			name:     "obsidian with all options",
			provider: "obsidian",
			token:    "",
			flags:    map[string]string{"path": "~/vault", "tasks-folder": "tasks", "file-pattern": "*.md"},
			wantKeys: map[string]string{"vault_path": "~/vault", "tasks_folder": "tasks", "file_pattern": "*.md"},
		},
		{
			name:     "obsidian missing path",
			provider: "obsidian",
			token:    "",
			flags:    map[string]string{},
			wantErr:  "missing required flags for obsidian: --path",
		},
		{
			name:     "reminders with list",
			provider: "reminders",
			token:    "",
			flags:    map[string]string{"list": "Shopping"},
			wantKeys: map[string]string{"lists": "Shopping"},
		},
		{
			name:       "reminders no flags is valid",
			provider:   "reminders",
			token:      "",
			flags:      map[string]string{},
			wantKeys:   map[string]string{},
			wantAbsent: []string{"lists"},
		},
		{
			name:     "linear with token and team-ids",
			provider: "linear",
			token:    "lin_api_123",
			flags:    map[string]string{"team-ids": "TEAM-1,TEAM-2"},
			wantKeys: map[string]string{"api_key": "lin_api_123", "team_ids": "TEAM-1,TEAM-2"},
		},
		{
			name:       "linear no flags is valid",
			provider:   "linear",
			token:      "",
			flags:      map[string]string{},
			wantKeys:   map[string]string{},
			wantAbsent: []string{"api_key"},
		},
		{
			name:     "clickup with required flags",
			provider: "clickup",
			token:    "pk_tok",
			flags:    map[string]string{"team-id": "team123", "space-ids": "space1"},
			wantKeys: map[string]string{"api_token": "pk_tok", "team_id": "team123", "space_ids": "space1"},
		},
		{
			name:     "clickup with list-ids",
			provider: "clickup",
			token:    "pk_tok",
			flags:    map[string]string{"team-id": "team123", "list-ids": "list1,list2"},
			wantKeys: map[string]string{"api_token": "pk_tok", "team_id": "team123", "list_ids": "list1,list2"},
		},
		{
			name:     "clickup missing team-id",
			provider: "clickup",
			token:    "pk_tok",
			flags:    map[string]string{"space-ids": "space1"},
			wantErr:  "missing required flags for clickup: --team-id",
		},
		{
			name:     "unknown provider passes flags through",
			provider: "unknown",
			token:    "tok",
			flags:    map[string]string{"repos": "r", "server": "s"},
			wantKeys: map[string]string{"repos": "r", "server": "s", "token": "tok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings, err := buildConnectSettings(tt.provider, tt.token, tt.flags)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for k, want := range tt.wantKeys {
				got, ok := settings[k]
				if !ok {
					t.Errorf("missing settings key %q", k)
					continue
				}
				if got != want {
					t.Errorf("settings[%q] = %q, want %q", k, got, want)
				}
			}

			for _, k := range tt.wantAbsent {
				if _, ok := settings[k]; ok {
					t.Errorf("settings key %q should be absent", k)
				}
			}
		})
	}
}

func TestRunConnectTo(t *testing.T) {
	t.Parallel()

	t.Run("successful connection text output", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
			TaskCount:    10,
		}}
		svc := newTestService(t, manager, checker)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"todoist", "Personal", "tok_123",
			map[string]string{},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		out := buf.String()
		for _, want := range []string{
			"Connection created:",
			"Name:     Personal",
			"Provider: todoist",
			"ID:",
			"Connection test:",
			"✓ DNS resolution",
			"✓ Authentication",
			"✓ Rate limit",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("missing %q in output:\n%s", want, out)
			}
		}
	})

	t.Run("successful connection json output", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
		}}
		svc := newTestService(t, manager, checker)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"todoist", "Work", "tok_456",
			map[string]string{},
			svc, manager, &buf, true, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Command != "connect" {
			t.Errorf("command = %q, want %q", env.Command, "connect")
		}
		if env.Error != nil {
			t.Fatalf("unexpected error: %+v", env.Error)
		}

		data, ok := env.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("data type = %T, want map", env.Data)
		}
		if data["name"] != "Work" {
			t.Errorf("name = %v, want Work", data["name"])
		}
		if data["provider"] != "todoist" {
			t.Errorf("provider = %v, want todoist", data["provider"])
		}
		if data["id"] == nil || data["id"] == "" {
			t.Error("id should be non-empty")
		}

		testData, ok := data["test"].(map[string]interface{})
		if !ok {
			t.Fatalf("test type = %T, want map", data["test"])
		}
		if testData["healthy"] != true {
			t.Errorf("healthy = %v, want true", testData["healthy"])
		}
	})

	t.Run("missing required flags", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"github", "OSS", "",
			map[string]string{},
			svc, manager, &buf, false, configPath,
		)
		if err == nil {
			t.Fatal("expected error for missing --repos")
		}
		if !strings.Contains(err.Error(), "--repos") {
			t.Errorf("error = %v, want containing --repos", err)
		}
	})

	t.Run("missing required flags json", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"jira", "Work", "tok",
			map[string]string{},
			svc, manager, &buf, true, configPath,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var env JSONEnvelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json decode: %v", err)
		}
		if env.Error == nil {
			t.Fatal("expected JSON error envelope")
		}
		if env.Error.Code != ExitValidation {
			t.Errorf("error code = %d, want %d", env.Error.Code, ExitValidation)
		}
		if !strings.Contains(env.Error.Message, "--server") {
			t.Errorf("error message = %q, want containing --server", env.Error.Message)
		}
	})

	t.Run("github with repos", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"github", "OSS", "",
			map[string]string{"repos": "owner/repo1,owner/repo2"},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Name:     OSS") {
			t.Errorf("missing connection name in output:\n%s", out)
		}
		if !strings.Contains(out, "Provider: github") {
			t.Errorf("missing provider in output:\n%s", out)
		}

		// Verify connection was added to manager.
		conn, err := manager.GetByLabel("OSS")
		if err != nil {
			t.Fatalf("connection not in manager: %v", err)
		}
		if conn.Settings["repos"] != "owner/repo1,owner/repo2" {
			t.Errorf("repos setting = %q, want %q", conn.Settings["repos"], "owner/repo1,owner/repo2")
		}
	})

	t.Run("jira with server and token", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"jira", "Work", "jira_tok",
			map[string]string{"server": "https://jira.company.com"},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		conn, err := manager.GetByLabel("Work")
		if err != nil {
			t.Fatalf("connection not in manager: %v", err)
		}
		if conn.Settings["url"] != "https://jira.company.com" {
			t.Errorf("url setting = %q, want %q", conn.Settings["url"], "https://jira.company.com")
		}
		if conn.Settings["api_token"] != "jira_tok" {
			t.Errorf("api_token setting = %q, want %q", conn.Settings["api_token"], "jira_tok")
		}
	})

	t.Run("textfile with path", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"textfile", "Notes", "",
			map[string]string{"path": "~/tasks.yaml"},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		conn, err := manager.GetByLabel("Notes")
		if err != nil {
			t.Fatalf("connection not in manager: %v", err)
		}
		if conn.Settings["path"] != "~/tasks.yaml" {
			t.Errorf("path setting = %q, want %q", conn.Settings["path"], "~/tasks.yaml")
		}
	})

	t.Run("health check failure still creates connection", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		checker := &stubHealthChecker{result: connection.HealthCheckResult{
			APIReachable: false,
			TokenValid:   false,
			RateLimitOK:  true,
		}}
		svc := newTestService(t, manager, checker)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"todoist", "Broken", "tok",
			map[string]string{},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Connection created:") {
			t.Error("connection should still be created despite health check failure")
		}
		if !strings.Contains(out, "✗ DNS resolution") {
			t.Errorf("should show failed DNS check:\n%s", out)
		}

		// Connection should exist in manager.
		if _, err := manager.GetByLabel("Broken"); err != nil {
			t.Errorf("connection should exist in manager: %v", err)
		}
	})

	t.Run("no health checker skips test", func(t *testing.T) {
		t.Parallel()

		manager := connection.NewConnectionManager(nil)
		svc := newTestService(t, manager, nil)
		configPath := t.TempDir() + "/config.yaml"

		var buf bytes.Buffer
		err := runConnectTo(
			"todoist", "NoChecker", "tok",
			map[string]string{},
			svc, manager, &buf, false, configPath,
		)
		if err != nil {
			t.Fatalf("runConnectTo: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Connection created:") {
			t.Error("connection should be created")
		}
		if !strings.Contains(out, "test: skipped") {
			t.Errorf("should show test skipped:\n%s", out)
		}
	})
}

func TestFormatConnectTest(t *testing.T) {
	t.Parallel()

	t.Run("all checks pass", func(t *testing.T) {
		t.Parallel()
		result := connection.HealthCheckResult{
			APIReachable: true,
			TokenValid:   true,
			RateLimitOK:  true,
		}
		ct := formatConnectTest(result)
		if !ct.Healthy {
			t.Error("expected healthy=true")
		}
		if len(ct.Checks) != 5 {
			t.Errorf("len(checks) = %d, want 5", len(ct.Checks))
		}
		for _, c := range ct.Checks {
			if !c.Passed {
				t.Errorf("check %q should pass", c.Name)
			}
		}
	})

	t.Run("api unreachable", func(t *testing.T) {
		t.Parallel()
		result := connection.HealthCheckResult{
			APIReachable: false,
			TokenValid:   true,
			RateLimitOK:  true,
		}
		ct := formatConnectTest(result)
		if ct.Healthy {
			t.Error("expected healthy=false")
		}
	})
}

func TestKnownProviderSpecsParity(t *testing.T) {
	t.Parallel()

	// All 9 registered providers must have entries in knownProviderSpecs.
	registeredProviders := []string{
		"applenotes", "clickup", "github", "jira", "linear",
		"obsidian", "reminders", "textfile", "todoist",
	}

	for _, name := range registeredProviders {
		if _, ok := knownProviderSpecs[name]; !ok {
			t.Errorf("registered provider %q has no entry in knownProviderSpecs", name)
		}
	}

	// knownProviderSpecs should not have entries for unregistered providers.
	for name := range knownProviderSpecs {
		found := false
		for _, rp := range registeredProviders {
			if name == rp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("knownProviderSpecs has entry for unregistered provider %q", name)
		}
	}
}

func TestValidArgsMatchKnownProviderSpecs(t *testing.T) {
	t.Parallel()

	cmd := newConnectCmd()

	validArgsSet := make(map[string]bool, len(cmd.ValidArgs))
	for _, arg := range cmd.ValidArgs {
		validArgsSet[arg] = true
	}

	specsSet := make(map[string]bool, len(knownProviderSpecs))
	for name := range knownProviderSpecs {
		specsSet[name] = true
	}

	// Every ValidArg must be in knownProviderSpecs.
	for _, arg := range cmd.ValidArgs {
		if !specsSet[arg] {
			t.Errorf("ValidArgs contains %q which is not in knownProviderSpecs", arg)
		}
	}

	// Every knownProviderSpec must be in ValidArgs.
	for name := range knownProviderSpecs {
		if !validArgsSet[name] {
			t.Errorf("knownProviderSpecs contains %q which is not in ValidArgs", name)
		}
	}
}

func TestConnectCommand(t *testing.T) {
	t.Parallel()

	t.Run("no label non-tty shows terminal requirement", func(t *testing.T) {
		t.Parallel()
		cmd := newConnectCmd()
		cmd.SetArgs([]string{"todoist"})

		// In test context, stdin is not a TTY, so the wizard should fail
		// with a terminal requirement message.
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error when no --label provided in non-TTY")
		}
		if !strings.Contains(err.Error(), "interactive wizard requires a terminal") {
			t.Errorf("error = %v, want terminal requirement message", err)
		}
	})

	t.Run("no args shows usage error", func(t *testing.T) {
		t.Parallel()
		cmd := newConnectCmd()
		cmd.SetArgs([]string{})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error when no provider given")
		}
	})
}
