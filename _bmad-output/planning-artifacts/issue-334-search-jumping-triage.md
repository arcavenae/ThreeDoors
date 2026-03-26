# Issue #334 Triage: Search Results Jumping Due to Unstable Map Iteration

**Date:** 2026-03-09
**Issue:** [#334](https://github.com/arcavenae/ThreeDoors/issues/334)
**Participants:** Winston (Architect), Amelia (Dev), Sally (UX Designer)
**Rounds:** 1 (consensus reached immediately)

## Problem

`GetAllTasks()` in `internal/core/task_pool.go` iterates over `map[string]*Task`, which has randomized iteration order in Go. The search view's `filterTasks()` calls `GetAllTasks()` on every keystroke, causing the results list to reshuffle randomly even when the same set of tasks matches.

## Options Evaluated

### Option A: Sort results in filterTasks() after collecting them (ADOPTED)

Add `slices.SortFunc(matched, ...)` by `Task.Text` in `filterTasks()` after the match loop.

**Pros:**
- Minimal change (3-line diff)
- Zero blast radius — only affects the caller that needs ordering
- Idiomatic Go — consumer responsible for its own ordering needs
- No new dependencies (uses stdlib `slices` package, available since Go 1.21)

**Cons:**
- Other callers that iterate the map still get random order (but none currently need stable order)

### Option B: Change GetAllTasks() to return a sorted slice (REJECTED)

Sort inside `GetAllTasks()` so all callers get stable order.

**Rejected because:** Imposes sorting cost on callers that don't need it (e.g., `GetAvailableForDoors()` randomizes selection anyway). Violates "don't pay for what you don't use." Also changes the contract for all existing callers.

### Option C: Cache the task order and only re-sort when tasks change (REJECTED)

Maintain a sorted slice alongside the map, invalidated on Add/Update/Remove.

**Rejected because:** Premature optimization. Task count is small (dozens, not thousands). Adds invalidation complexity (new code paths in AddTask, UpdateTask, RemoveTask) for negligible performance gain. More moving parts = more bugs.

### Option D: Use a sync.Map or ordered data structure (REJECTED)

Replace `map[string]*Task` with an ordered data structure (btree, skip list, etc.).

**Rejected because:** `sync.Map` doesn't guarantee order either. An ordered map would require a new external dependency for a problem solved by a 3-line sort. Massively over-engineered for the actual need.

## Decision

**Option A: Sort results in `filterTasks()` using `slices.SortFunc` by `Task.Text` (alphabetical).**

Sort key rationale: Alphabetical by task text is the most user-predictable ordering. `CreatedAt` would also provide stability but is less intuitive — users scanning a filtered list expect lexicographic order when no explicit sort is specified.

## Implementation Notes

- File to modify: `internal/tui/search_view.go`, `filterTasks()` method (lines 84-97)
- Add `slices` import
- Add sort after the match loop, before return
- Test: verify same query produces same order across multiple calls
- This is a TUI change — race detector test required per CLAUDE.md
