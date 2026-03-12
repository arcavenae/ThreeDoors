# Task Dependencies

Tasks can declare dependencies on other tasks. A task with unmet dependencies is automatically filtered out of door selection — you won't see it until its prerequisites are complete.

---

## How It Works

1. Add a `depends_on` field to a task listing the IDs of prerequisite tasks
2. The **DependencyResolver** checks all tasks before door selection and filters out any with incomplete dependencies
3. When you complete a task, any tasks that depended solely on it are automatically unblocked and become eligible for door selection

---

## Dependency Types

### Direct Dependencies

A task can depend on one or more other tasks:

```yaml
- id: task-abc
  text: "Write integration tests"
  depends_on:
    - task-xyz  # Must complete "Design API" first
```

### Transitive Dependencies

If Task C depends on Task B, and Task B depends on Task A, then Task C won't appear in doors until both A and B are complete.

### Circular Dependencies

ThreeDoors detects circular dependencies and breaks the cycle by treating the involved tasks as having no dependencies. A warning is logged but the application continues normally.

---

## In the TUI

### Blocked-By Indicators

When viewing a task's detail, dependency relationships are visible:

- Tasks that **block** the current task are listed
- Tasks that are **waiting on** the current task are listed

### Linking Tasks

Press `l` in the detail view to link the current task to another task. Press `x` to browse existing links and navigate between related tasks.

### Dependency Graph

The MCP server exposes graph traversal tools for exploring dependency chains:

- `walk_graph` — BFS traversal from a starting task
- `find_paths` — Find paths between two tasks
- `get_critical_path` — Longest dependency chain
- `get_orphans` — Tasks with no relationships

---

## Session Metrics

Dependency events are logged in session metrics for pattern analysis:

- When a task is unblocked by completing its prerequisite
- When a blocked task is shown in the detail view
- When dependency links are created or removed
