package cli

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
)

func TestVersionHumanOutput(t *testing.T) {
	t.Parallel()

	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	Version, Commit, BuildDate = "1.2.3", "abc1234", "2025-01-15T10:00:00Z"
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
	})

	var buf bytes.Buffer
	err := writeVersion(&buf, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	checks := []string{"ThreeDoors 1.2.3", "abc1234", "2025-01-15T10:00:00Z", runtime.Version()}
	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q, got:\n%s", want, output)
		}
	}
}

func TestVersionJSONOutput(t *testing.T) {
	t.Parallel()

	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	Version, Commit, BuildDate = "2.0.0", "def5678", "2025-06-01T00:00:00Z"
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
	})

	var buf bytes.Buffer
	err := writeVersion(&buf, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if env.Command != "version" {
		t.Errorf("command = %q, want %q", env.Command, "version")
	}
	if env.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", env.SchemaVersion)
	}

	data, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %T", env.Data)
	}
	if data["version"] != "2.0.0" {
		t.Errorf("version = %v, want %q", data["version"], "2.0.0")
	}
	if data["commit"] != "def5678" {
		t.Errorf("commit = %v, want %q", data["commit"], "def5678")
	}
	if data["build_date"] != "2025-06-01T00:00:00Z" {
		t.Errorf("build_date = %v, want %q", data["build_date"], "2025-06-01T00:00:00Z")
	}
	if data["go_version"] != runtime.Version() {
		t.Errorf("go_version = %v, want %q", data["go_version"], runtime.Version())
	}
}

func TestVersionCommandRegistered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version command not registered on root")
	}
}
