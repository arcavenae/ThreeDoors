package adapters_test

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/adapters/obsidian"

	"github.com/arcavenae/ThreeDoors/internal/adapters"
	"github.com/arcavenae/ThreeDoors/internal/core"
)

// TestObsidianAdapterContract runs the full contract test suite
// against the ObsidianAdapter to validate compliance.
func TestObsidianAdapterContract(t *testing.T) {
	factory := func(t *testing.T) core.TaskProvider {
		t.Helper()
		dir := t.TempDir()
		return obsidian.NewObsidianAdapter(dir, "", "")
	}

	adapters.RunContractTests(t, factory)
}
