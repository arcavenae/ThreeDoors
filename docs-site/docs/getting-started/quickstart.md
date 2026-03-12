# Quick Start

Get from zero to your first completed task in under 5 minutes.

## First Launch

Run `threedoors` in your terminal. On first launch, the onboarding wizard walks you through five steps:

1. **Welcome** — Introduction to the Three Doors concept
2. **Key bindings** — Interactive tutorial where you press keys to learn the controls
3. **Values** — Set 1–5 personal values or goals (displayed as a reminder while you work)
4. **Import** — Optionally import tasks from CSV or Apple Notes
5. **Done** — You're ready to go

After onboarding, ThreeDoors creates its data directory at `~/.threedoors/` with a default text file provider.

## Your First Task

### Add a task

Open the command palette with ++colon++ and type:

```
:add Buy groceries
```

Press ++enter++ to confirm. The task is added to your pool.

### Pick a door

Three doors appear. Each hides a task. Select one:

| Key | Door |
|-----|------|
| ++a++ or ++left++ | Left door |
| ++w++ or ++up++ | Center door |
| ++d++ or ++right++ | Right door |

Press ++space++ or ++enter++ to open the selected door and see the task details.

### Complete the task

In the detail view, press ++c++ to mark the task complete. You'll see a celebration message with your daily completion count.

### Refresh doors

Don't like your options? Press ++s++ or ++down++ to get three new doors from your task pool.

## Typical Workflow

1. Launch `threedoors`
2. Look at your three doors
3. Pick one (++a++, ++w++, or ++d++, then ++enter++)
4. Take action: complete it (++c++), mark in-progress (++i++), or return it to the pool (++p++)
5. Repeat

!!! tip
    Press ++question++ at any time to open the keybinding overlay showing all available keys for the current view.

## Next Steps

- Learn the [Core Concepts](concepts.md) behind ThreeDoors
- Explore [Task Management](../guide/task-management.md) for statuses, categorization, and more
- Connect your existing tasks via [Task Sources](../providers/overview.md)
