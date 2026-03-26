# Adapter Developer Guide

This guide covers how to build a new task provider adapter for ThreeDoors.

## Overview

ThreeDoors uses the **TaskProvider** interface to abstract task storage backends. Each adapter implements this interface and registers itself with the adapter registry. Users configure active providers via `~/.threedoors/config.yaml`.

## TaskProvider Interface

```go
// internal/tasks/provider.go
type TaskProvider interface {
    LoadTasks() ([]*Task, error)
    SaveTask(task *Task) error
    SaveTasks(tasks []*Task) error
    DeleteTask(taskID string) error
    MarkComplete(taskID string) error
}
```

### Method Contracts

| Method | Description | Error Behavior |
|--------|-------------|----------------|
| `LoadTasks()` | Returns all active tasks | Return error on I/O failure |
| `SaveTask(task)` | Upsert a single task (insert or update by ID) | Return error on write failure |
| `SaveTasks(tasks)` | Replace all tasks with the given slice | Return error on write failure |
| `DeleteTask(taskID)` | Remove a task by ID | May return nil for non-existent IDs |
| `MarkComplete(taskID)` | Mark a task as complete and remove from active set | Return error if task not found. Return `ErrReadOnly` if provider is read-only |

### Key Invariants

- `LoadTasks()` must return tasks saved by `SaveTasks()` with matching IDs and content.
- `SaveTask()` must create the task if the ID is new, or update if it exists.
- `DeleteTask()` must remove the task so subsequent `LoadTasks()` does not return it.
- All methods must be safe to call from any goroutine (thread-safe).

## Creating an Adapter

### Step 1: Implement the Interface

Create your adapter in a new file under `internal/tasks/`:

```go
// internal/tasks/my_adapter_provider.go
package tasks

type MyAdapterProvider struct {
    path string
}

func NewMyAdapterProvider(path string) *MyAdapterProvider {
    return &MyAdapterProvider{path: path}
}

func (p *MyAdapterProvider) LoadTasks() ([]*Task, error) {
    // Read tasks from your storage backend
}

func (p *MyAdapterProvider) SaveTask(task *Task) error {
    // Upsert a single task
}

func (p *MyAdapterProvider) SaveTasks(tasks []*Task) error {
    // Replace all tasks
}

func (p *MyAdapterProvider) DeleteTask(taskID string) error {
    // Remove a task by ID
}

func (p *MyAdapterProvider) MarkComplete(taskID string) error {
    // Mark task as complete
    // Return ErrReadOnly if your provider doesn't support writes
}
```

### Step 2: Register with the Adapter Registry

Add your adapter to `RegisterBuiltinAdapters()` in `internal/tasks/adapters.go`:

```go
func RegisterBuiltinAdapters(reg *Registry) {
    // ... existing adapters ...

    _ = reg.Register("myadapter", func(config *ProviderConfig) (TaskProvider, error) {
        // Extract settings from config
        path := "default/path"
        if len(config.Providers) > 0 {
            for _, p := range config.Providers {
                if p.Name == "myadapter" {
                    path = p.GetSetting("path", path)
                }
            }
        }
        return NewMyAdapterProvider(path), nil
    })
}
```

### Step 3: Define Config Schema

Users configure your adapter in `~/.threedoors/config.yaml`:

```yaml
providers:
  - name: myadapter
    settings:
      path: "/path/to/tasks"
      # Add any provider-specific settings here
```

The `ProviderEntry.Settings` field is a `map[string]string`. Use `GetSetting(key, fallback)` to read values with defaults:

```go
entry.GetSetting("path", "/default/path")
```

### Step 4: Run Contract Tests

Import and run the contract test suite to validate your implementation:

```go
// internal/tasks/my_adapter_provider_test.go
package tasks

import (
    "testing"

    "github.com/arcavenae/ThreeDoors/internal/adapters"
)

func TestMyAdapterContract(t *testing.T) {
    factory := func(t *testing.T) TaskProvider {
        t.Helper()
        dir := t.TempDir()
        return NewMyAdapterProvider(dir)
    }

    adapters.RunContractTests(t, factory)
}
```

The contract test suite validates:
- **CRUD operations**: Save/load round-trip, individual save, update, delete
- **Error handling**: Non-existent task deletion, non-existent task completion
- **Concurrent access**: Parallel reads, parallel writes without panics or corruption
- **Interface compliance**: All interface methods are tested

## Reference Implementation

The **TextFileProvider** (`internal/tasks/text_file_provider.go`) is the reference implementation. Key patterns to follow:

1. **Atomic writes**: Write to `.tmp` file, `Sync()`, then `Rename()` to final path
2. **Error wrapping**: Use `fmt.Errorf("operation: %w", err)` for error chain preservation
3. **UTC timestamps**: Always use `time.Now().UTC()`
4. **Constructor function**: Provide `NewMyProvider()` factory function

## Existing Adapters

| Name | File | Description |
|------|------|-------------|
| `textfile` | `text_file_provider.go` | YAML-based local file storage (default) |
| `applenotes` | `apple_notes_provider.go` | macOS Notes integration via AppleScript |

## Wrapper Providers

ThreeDoors provides wrapper providers for cross-cutting concerns:

- **FallbackProvider** (`fallback_provider.go`): Wraps a primary provider with automatic fallback to a secondary provider on failure or `ErrReadOnly`.
- **WALProvider** (`wal_provider.go`): Write-ahead log wrapper for offline-first sync. Queues failed writes and replays them later.

## Testing Requirements

Per project coding standards:
- Use stdlib `testing` package only (no testify)
- Table-driven tests for functions with >2 test cases
- Use `t.Helper()` in test helper functions
- Use `t.Cleanup()` instead of `defer` for resource cleanup
- Test files live alongside source: `foo.go` → `foo_test.go`
- Run `go test -race ./...` before pushing
