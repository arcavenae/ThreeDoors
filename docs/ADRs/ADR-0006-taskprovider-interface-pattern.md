# ADR-0006: TaskProvider Interface Pattern

- **Status:** Accepted
- **Date:** 2025-11-10 (Epic 2), formalized in Epic 7
- **Decision Makers:** Architecture reviews
- **Related PRs:** #20 (Story 2.1), #68 (Story 7.1), #72 (Story 7.3), #89 (Story 9.2)
- **Related ADRs:** ADR-0005 (Layered Architecture), ADR-0007 (Compile-Time Registration)

## Context

ThreeDoors initially supported only YAML file-based task storage. Epic 2 (Apple Notes) introduced the need for a second task source, requiring an abstraction for storage backends.

## Decision

Define a `TaskProvider` interface as the core abstraction for all task storage backends:

```go
type TaskProvider interface {
    Name() string
    Load() ([]*Task, error)
    Save(task *Task) error
    Delete(taskID string) error
    Watch() (<-chan ChangeEvent, error)
}
```

All adapters (text file, Apple Notes, Obsidian, Jira, Apple Reminders, GitHub Issues) implement this interface.

## Design Principles

- **Accept interfaces, return concrete types** — factory functions return `*ConcreteProvider`, not `TaskProvider`
- **Full interface implementation** — even read-only adapters implement all methods, returning `ErrReadOnly` for unsupported operations (Decision H5)
- **Contract tests validate compliance** — shared test suite verifies every adapter meets the interface contract (Story 9.2)
- **No partial interfaces** — rejected splitting into `ReadProvider`/`WriteProvider` to keep the abstraction simple

## Implemented Providers

| Provider | Epic | Read | Write | Watch | PR |
|----------|------|------|-------|-------|-----|
| TextFileProvider | 1 | Yes | Yes | Yes | #8 |
| AppleNotesProvider | 2 | Yes | Yes | Yes | #17, #21 |
| ObsidianProvider | 8 | Yes | Yes | Yes | #73, #75 |
| JiraProvider | 19 | Yes | Yes | Yes | #138, #150 |
| RemindersProvider | 20 | Yes | Yes | No | #137, #148, #155 |
| GitHubProvider | 26 | Yes | Yes | Yes | #201, #202, #204 |

## Consequences

### Positive
- Adding a new task source requires only implementing one interface
- Contract tests catch adapter bugs automatically
- Core domain code is provider-agnostic
- Multi-source aggregation (Epic 13) treats all providers uniformly

### Negative
- Read-only adapters carry boilerplate `ErrReadOnly` methods
- `Watch()` semantics vary by provider (filesystem events vs. polling)
- Interface evolution requires updating all implementations
