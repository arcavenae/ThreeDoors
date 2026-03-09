//go:build integration

package todoist

import (
	"os"
	"testing"
)

// Integration tests require a real Todoist account and API token.
// Run with: go test -tags=integration -run TestIntegration ./internal/adapters/todoist/...
//
// Required environment variable:
//   TODOIST_API_TOKEN - a valid Todoist API token

func TestIntegrationLoadTasks(t *testing.T) {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		t.Skip("TODOIST_API_TOKEN not set")
	}

	client := NewClient(AuthConfig{APIToken: token})
	cfg := &TodoistConfig{APIToken: token}
	p := NewTodoistProvider(client, cfg)

	tasks, err := p.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	t.Logf("Loaded %d tasks from Todoist", len(tasks))

	for _, task := range tasks {
		if task.ID == "" {
			t.Error("task has empty ID")
		}
		if task.Text == "" {
			t.Error("task has empty Text")
		}
		if task.Status == "" {
			t.Error("task has empty Status")
		}
	}
}

func TestIntegrationHealthCheck(t *testing.T) {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		t.Skip("TODOIST_API_TOKEN not set")
	}

	client := NewClient(AuthConfig{APIToken: token})
	cfg := &TodoistConfig{APIToken: token}
	p := NewTodoistProvider(client, cfg)

	result := p.HealthCheck()
	if result.Overall != "OK" {
		t.Errorf("HealthCheck Overall = %q, want OK", result.Overall)
	}
	t.Logf("HealthCheck duration: %v", result.Duration)
}

func TestIntegrationName(t *testing.T) {
	token := os.Getenv("TODOIST_API_TOKEN")
	if token == "" {
		t.Skip("TODOIST_API_TOKEN not set")
	}

	client := NewClient(AuthConfig{APIToken: token})
	cfg := &TodoistConfig{APIToken: token}
	p := NewTodoistProvider(client, cfg)

	if name := p.Name(); name != "todoist" {
		t.Errorf("Name() = %q, want %q", name, "todoist")
	}
}
