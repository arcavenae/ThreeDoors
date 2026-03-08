# Party Mode Session: Expand/Fork Key Implementations

**Date:** 2026-03-08
**Topic:** Epic 31 — Expand/Fork Key Implementations Design Decisions
**Participants:** Winston (Architect), Sally (UX Designer), Amelia (Dev), John (PM)

---

## Design Decisions Resolved

### 1. ParentID Location: Native to Task Struct

**Decision:** Add `ParentID *string` as a native field on `core.Task` (YAML: `parent_id,omitempty`).

**Rationale (Winston):** Parent-child is a core domain relationship, not optional metadata. The enrichment DB's CrossReference system is designed for optional metadata (link discovery, dedup). If ParentID lives in enrichment, you can't query subtasks without the enrichment DB, breaking clean separation. A native field works across all providers and lets TaskPool answer "give me children of X" without external dependencies.

**CrossReference usage:** The enrichment DB's `parent-of`/`child-of` types (Decision M8) remain for cross-provider relationships. Fork relationships use the enrichment DB since they're optional metadata.

### 2. Fork Semantics: Variant Creation

**Decision:** `core.ForkTask(original *Task) *Task` factory method.

**Preserves:** Text, Context, Effort, Tags
**Resets:** Status → todo, Blocker → "", Notes → empty, Timestamps → now (UTC)
**Adds:** Note "Forked from: [truncated original text]"
**Does NOT copy:** ParentID (fork is independent)
**Cross-reference:** Main model creates enrichment DB cross-reference (forked-from) in `TaskForkedMsg` handler. Core package stays free of enrichment DB awareness.

### 3. Property Inheritance: None

**Decision:** Subtasks do NOT inherit effort, tags, or context from parent.

**Rationale (John):** Each subtask is its own unit of work. Auto-inheriting effort=3 from a parent to 5 subtasks quintuples the total effort estimate — misleading. Parent's context explains why the breakdown exists, but subtasks need their own sizing.

### 4. Subtask Completion → Parent: No Auto-Completion

**Decision:** Never auto-complete a parent when all subtasks complete. Show completion ratio "3/5 subtasks complete" in detail view.

**Additional:** Parent tasks with children are excluded from door rotation via `GetAvailableForDoors()` filter. This communicates "you broke this down, now work the pieces."

### 5. Sequential Expand Mode (Bonus)

**Decision (Sally):** After pressing Enter on one subtask, stay in `DetailModeExpandInput` for the next one. Show running count: "Subtask N added. Next subtask (Esc to finish):". Only Esc exits expand mode.

---

## Implementation Notes (Amelia)

- `ParentID *string` on `core.Task`, `yaml:"parent_id,omitempty"`
- `TaskPool.GetSubtasks(parentID string) []*Task` method
- `ExpandTaskMsg` handler sets `newTask.ParentID = &parentTask.ID`
- `ForkTask` returns `*Task`, main model handles cross-reference
- `handleExpandInput` stays in expand mode on Enter, only exits on Esc
- Subtask rendering in detail view: indented list with status icons

## Detail View Subtask Rendering (Sally)

```
Write architecture document
  ├─ [TODO] Draft high-level overview
  ├─ [DONE] Data models section
  └─ [TODO] Components section

Subtasks: 1/3 complete
```
