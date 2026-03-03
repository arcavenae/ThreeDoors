package adapters_test

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/adapters"
	"github.com/arcaven/ThreeDoors/internal/tasks"
)

// TestObsidianAdapterContract runs the full contract test suite
// against the ObsidianAdapter to validate compliance.
func TestObsidianAdapterContract(t *testing.T) {
	factory := func(t *testing.T) tasks.TaskProvider {
		t.Helper()
		dir := t.TempDir()
		return tasks.NewObsidianAdapter(dir, "", "")
	}

	adapters.RunContractTests(t, factory)
}
