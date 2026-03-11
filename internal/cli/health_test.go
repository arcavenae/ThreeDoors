package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestHealthCommandRegistered(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	for _, cmd := range root.Commands() {
		if cmd.Name() == "doctor" {
			for _, alias := range cmd.Aliases {
				if alias == "health" {
					return // health is registered as alias of doctor
				}
			}
			t.Error("doctor command exists but missing 'health' alias")
			return
		}
	}
	t.Error("doctor command not registered on root")
}

func TestHealthCheckerResultsJSON(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{}
	hc := core.NewHealthChecker(provider)
	result := hc.RunAll()

	checks := make([]healthCheckJSON, 0, len(result.Items))
	for _, item := range result.Items {
		checks = append(checks, healthCheckJSON{
			Name:    item.Name,
			Status:  string(item.Status),
			Message: item.Message,
		})
	}
	data := healthResultJSON{
		Overall:    string(result.Overall),
		DurationMs: result.Duration.Milliseconds(),
		Checks:     checks,
	}

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, true)
	err := formatter.WriteJSON("health", data, nil)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
		return
	}

	var env JSONEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if env.Command != "health" {
		t.Errorf("command = %q, want %q", env.Command, "health")
	}

	envData, ok := env.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("data is not a map: %T", env.Data)
	}

	if _, ok := envData["overall"]; !ok {
		t.Error("missing 'overall' field in data")
	}
	if _, ok := envData["duration_ms"]; !ok {
		t.Error("missing 'duration_ms' field in data")
	}
	checksRaw, ok := envData["checks"]
	if !ok {
		t.Fatal("missing 'checks' field in data")
	}
	checksArr, ok := checksRaw.([]interface{})
	if !ok {
		t.Fatalf("checks is not an array: %T", checksRaw)
	}
	if len(checksArr) == 0 {
		t.Error("expected at least one health check")
	}
}

func TestHealthCheckerTableOutput(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{}
	hc := core.NewHealthChecker(provider)
	result := hc.RunAll()

	var buf bytes.Buffer
	formatter := NewOutputFormatter(&buf, false)

	tw := formatter.TableWriter()
	for _, item := range result.Items {
		_, _ = tw.Write([]byte(item.Name + "\t" + string(item.Status) + "\t" + item.Message + "\n"))
	}
	_ = tw.Flush()

	output := buf.String()
	if !strings.Contains(output, "Database") {
		t.Errorf("table output missing 'Database', got:\n%s", output)
	}
}

func TestExitCodeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
		want int
	}{
		{"success", ExitSuccess, 0},
		{"general error", ExitGeneralError, 1},
		{"not found", ExitNotFound, 2},
		{"validation", ExitValidation, 3},
		{"provider error", ExitProviderError, 4},
		{"ambiguous input", ExitAmbiguousInput, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.code != tt.want {
				t.Errorf("exit code %s = %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}
