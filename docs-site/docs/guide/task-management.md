# Task Management

## Task Statuses

| Status | Description |
|--------|-------------|
| **Todo** | Default state for new tasks |
| **In-Progress** | Actively being worked on |
| **Blocked** | Cannot proceed; requires a blocker reason |
| **Complete** | Done (removed from active pool) |
| **Deferred** | Intentionally postponed |
| **Archived** | Removed from pool without completing |
| **In-Review** | Awaiting review (from in-progress) |

## Status Transitions

```
Todo ──→ In-Progress ──→ Complete
  │          │    │
  │          │    └──→ In-Review ──→ Complete
  │          │              │
  │          └──→ Blocked ──┘
  │
  ├──→ Blocked ──→ Todo / In-Progress / Complete
  ├──→ Complete (terminal)
  ├──→ Deferred ──→ Todo
  └──→ Archived (terminal)
```

## Adding Tasks

Three ways to add a task:

**Quick add** — open the command palette with ++colon++ and type:

```
:add Buy groceries
```

**Add with context** — captures the task and why it matters:

```
:add-ctx Refactor auth module
```

You'll be prompted: "Why is this important?" — your answer is stored as context.

**Inline alternative:**

```
:add --why Refactor auth module
```

!!! tip
    Task text is parsed for inline categorization automatically.

## Action Keys in Detail View

After opening a door, these keys act on the task:

| Key | Action | Description |
|-----|--------|-------------|
| ++c++ | Complete | Marks the task done. Removed from active pool, logged to `completed.txt`. Shows a celebration with your daily count. |
| ++i++ | In-Progress | Signals you're actively working on this task. |
| ++b++ | Blocked | Prompts for a reason (e.g., "waiting on API key"). ++enter++ to confirm, ++esc++ to cancel. |
| ++p++ | Procrastinate | Returns the task to the pool. No judgment — it'll come back around. |
| ++r++ | Rework | Returns the task to the pool, signaling it needs more thought. |

## Task Categorization

Press ++colon++ then type `tag` to categorize a task across three dimensions:

**Type:**

| Category | Icon | Examples |
|----------|------|----------|
| Creative | :art: | Design, writing, ideation |
| Administrative | :clipboard: | Email, scheduling, paperwork |
| Technical | :wrench: | Coding, debugging, system work |
| Physical | :muscle: | Exercise, errands, hands-on tasks |

**Effort:**

| Level | Meaning |
|-------|---------|
| Quick Win | Under 15 minutes |
| Medium | 15–60 minutes |
| Deep Work | Focused, extended effort |

**Location:** Home, Work, Errands, Anywhere

!!! info
    Categorization improves door selection over time — ThreeDoors learns which task types you prefer in different moods.

## Snooze and Defer

Press ++z++ in the task detail view to snooze a task:

- Set a return date — the task moves to `deferred` status with a `defer_until` timestamp
- When the defer date passes, the task auto-returns to `todo` status
- Snoozed tasks are excluded from door selection until they're due

View all snoozed tasks with `:deferred` in the command palette.

## Undo Completion

Accidentally completed a task? The `complete → todo` transition lets you reverse it:

- In the detail view of a completed task, undo the completion
- The task returns to `todo` status and re-enters the active pool
- Undo events are logged in session metrics for pattern analysis

## Mood Tracking

Press ++m++ at any time to log how you're feeling:

1. Focused
2. Tired
3. Stressed
4. Energized
5. Distracted
6. Calm
7. Other (type your own)

Mood entries are timestamped and stored in your session data. Over time, ThreeDoors correlates mood with task completion patterns and adjusts door selection accordingly.
