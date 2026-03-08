# Error Handling Strategy

## General Approach

**Error Model:** Idiomatic Go - errors as values

**Error Propagation:**
- Functions return `(result, error)` tuple
- Callers check `if err != nil` explicitly
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Sentinel errors for expected conditions

**Example:**
```go
func LoadTasks() (*TaskPool, error) {
    data, err := os.ReadFile(tasksPath)
    if err != nil {
        if os.IsNotExist(err) {
            return createDefaultTasks()
        }
        return nil, fmt.Errorf("failed to read tasks file: %w", err)
    }
    // ... continue processing
}
```

## Logging Standards

**Library:** Standard library `log` package

**Format:** Plain text to stderr

**Levels:**
- **ERROR:** Critical failures - `ERROR: Failed to save tasks: ...`
- **WARN:** Recoverable issues - `WARNING: Skipping invalid task: ...`
- **INFO:** Normal operations - `INFO: Loaded 12 tasks`

**Required Context:**
- Operation being performed
- File paths involved
- Error details from wrapped errors

## Error Handling Patterns

### File I/O Errors

**User Experience:**
- Missing file → Create samples, show welcome message
- Corrupted file → Backup, create new, show warning
- Permission denied → Show error with fix instructions

### YAML Parsing Errors

**Strategy:** Skip invalid tasks, continue with valid ones

```go
for i, task := range yamlData.Tasks {
    if err := task.Validate(); err != nil {
        log.Printf("WARNING: Skipping invalid task at index %d: %v\n", i, err)
        continue
    }
    pool.AddTask(task)
}
```

### Status Transition Errors

**Strategy:** Validate before attempting, show clear error to user

```go
if err := statusManager.ValidateTransition(from, to); err != nil {
    m.errorMessage = err.Error()
    return m, nil // Don't transition view
}
```

### Atomic Write Failures

**Strategy:** Retry once, rollback on failure, preserve original

```go
// Write to temp file
if err := os.WriteFile(tempPath, data, 0644); err != nil {
    return fmt.Errorf("failed to write temp file: %w", err)
}

// Atomic rename
if err := os.Rename(tempPath, targetPath); err != nil {
    os.Remove(tempPath) // Cleanup
    return fmt.Errorf("failed to commit changes: %w", err)
}
```

### Provider Initialization Errors

**Strategy:** Check for nil provider before use, return actionable error

Factory functions like `NewProviderFromConfig()` can return `nil` when both
provider resolution and fallback fail. All callers must nil-check before
dereferencing. Follow the pattern established in `bootstrap.go`:

```go
provider := core.NewProviderFromConfig(cfg)
if provider == nil {
    return fmt.Errorf("no task provider available; check provider configuration")
}
```

**Known call sites requiring nil defense:**
- `internal/cli/doors.go` — `loadTaskPool()` (fixed in Story 23.11)
- `cmd/threedoors-mcp/main.go` — MCP server init (fixed in Story 23.11)
- `internal/cli/bootstrap.go` — already defended

**Future consideration:** Refactor `NewProviderFromConfig()` to return
`(TaskProvider, error)` to eliminate the nil-return pattern at the source.

## Error Message Guidelines

**User-Facing Errors (in TUI):**
- ✅ "Could not save tasks. Check disk space and try again."
- ❌ "os.WriteFile: no space left on device"

**Developer Errors (in logs):**
- ✅ `ERROR: Failed to save tasks to ~/.threedoors/tasks.yaml: no space left on device`
- ❌ Generic error messages without context

---
