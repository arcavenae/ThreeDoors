package cli

import (
	"testing"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestCliContext_Fields(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	provider := &fakeProvider{}

	ctx := &cliContext{
		provider: provider,
		pool:     pool,
	}

	if ctx.pool == nil {
		t.Error("pool should not be nil")
	}
	if ctx.provider == nil {
		t.Error("provider should not be nil")
	}
	if ctx.provider.Name() != "fake" {
		t.Errorf("provider.Name() = %q, want %q", ctx.provider.Name(), "fake")
	}
}

func TestCliContext_PoolOperations(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	task := core.NewTask("Test task")
	task.ID = "ctx-test-id"
	pool.AddTask(task)

	provider := &fakeProvider{}
	ctx := &cliContext{provider: provider, pool: pool}

	found := ctx.pool.FindByPrefix("ctx-test")
	if len(found) != 1 {
		t.Fatalf("FindByPrefix returned %d results, want 1", len(found))
	}
	if found[0].Text != "Test task" {
		t.Errorf("task text = %q, want %q", found[0].Text, "Test task")
	}
}

func TestCliContext_ProviderSaveTask(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	provider := &fakeProvider{}
	ctx := &cliContext{provider: provider, pool: pool}

	task := core.NewTask("Save me")
	if err := ctx.provider.SaveTask(task); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}
	if len(provider.saved) != 1 {
		t.Errorf("saved %d tasks, want 1", len(provider.saved))
	}
}
