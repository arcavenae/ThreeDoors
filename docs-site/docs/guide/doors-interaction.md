# Doors Interaction

## Selecting a Door

Three doors are presented, each hiding a task from your pool. Select one:

| Key | Door |
|-----|------|
| ++a++ or ++left++ | Left door |
| ++w++ or ++up++ | Center door |
| ++d++ or ++right++ | Right door |

Press ++space++ or ++enter++ to open the selected door and view the task details.

Press ++esc++ in the detail view to return to the doors.

## Refreshing Doors

Press ++s++ or ++down++ to get three new doors from your task pool. The previous set is discarded and a fresh selection is drawn.

!!! info
    The selection algorithm avoids showing recently-displayed tasks, prefers diversity across task types and effort levels, and adapts to your logged mood. See [Core Concepts](../getting-started/concepts.md#how-door-selection-works) for details.

## Door Feedback

Press ++n++ on a door (without opening it) to give quick feedback:

1. **Blocked** — can't do this right now
2. **Not now** — not the right time
3. **Needs breakdown** — too big, needs splitting
4. **Other** — free-text comment

Feedback is recorded in session metrics for pattern analysis.

## Mood Logging

Press ++m++ from the doors view to log your current mood. Moods influence door selection — if you historically complete more technical tasks when focused, ThreeDoors will surface technical tasks when you log "Focused."

See [Task Management — Mood Tracking](task-management.md#mood-tracking) for the full mood list.

## Avoidance Detection

If a task has been shown 10+ times and you've never selected it, ThreeDoors gently asks what's going on:

- **Reconsider** — set it to in-progress and tackle it now
- **Break down** — open it in detail view to rethink it
- **Defer** — explicitly postpone it
- **Archive** — remove it from the pool entirely

This happens at most once per task per session.

## Intelligent Features

### Mood-Aware Selection

Once you have sufficient session data, logging a mood influences which tasks appear:

1. Looks up your mood in the pattern report
2. Finds your preferred task type and effort level for that mood
3. Scores candidate door sets for diversity + mood alignment
4. Selects the highest-scoring set
5. Enforces a diversity floor (won't show three identical types)

### Pattern Recognition

After 3+ sessions, ThreeDoors analyzes your behavior:

- **Door position bias** — do you always pick the left door?
- **Task type preferences** — which categories do you complete most?
- **Time-of-day patterns** — when are you most productive?
- **Mood correlations** — which moods lead to more completions?

View these with `:dashboard` or `:insights`.
