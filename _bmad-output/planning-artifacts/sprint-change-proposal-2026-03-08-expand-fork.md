# Sprint Change Proposal: Expand/Fork Key Implementations

**Date:** 2026-03-08
**Proposed by:** BMAD Pipeline (fancy-bear worker)
**Change Scope:** Minor — Direct implementation by development team

---

## Section 1: Issue Summary

**Problem Statement:** The ThreeDoors TUI detail view has two key actions — `[E]xpand` (manual sub-task creation) and `[F]ork` (variant creation) — that have basic stub implementations but lack the full feature depth specified in Design Decision H9. Expand currently provides only a single-line text input that creates an unlinked new task. Fork creates a simple copy with `NewTask(text)` but doesn't implement the "variant creation" semantics (stripping assignee, resetting status, adding context about the original task). Neither feature tracks parent-child relationships, shows subtask lists, or provides the user experience described in the PRD.

**Context:** Decision H9 in `docs/design-decisions-needed.md` has been resolved:
- **Expand = Option A (manual sub-task creation):** Opens a form to add sub-tasks manually, with parent-child relationship tracking
- **Fork = Option B (variant creation):** Creates a copy with some fields stripped (no assignee, reset status) for exploring a new direction
- **LLM decomposition stays separate** under the `[G]enerate` key (Epic 14, already complete)

**Evidence:**
- `internal/tui/detail_view.go` lines 147-152: Expand opens a basic text input, emits `ExpandTaskMsg` but creates an unlinked task
- `internal/tui/detail_view.go` lines 150-152: Fork calls `core.NewTask(dv.task.Text)` — a flat copy with no variant semantics
- The `core.Task` struct has no `ParentID` or `SubtaskIDs` fields for parent-child tracking
- No subtask rendering in the detail view or doors view
- The `ExpandTaskMsg` and `TaskAddedMsg` are handled in the main model but don't establish relationships
- PRD (`docs/prd/user-interface-design-goals.md`) lists both `[E]xpand` and `[F]ork` as core task actions

---

## Section 2: Impact Analysis

### Epic Impact
- **No existing epics affected.** This is a purely additive new epic (proposed: Epic 31).
- **Dependencies:** All prerequisite infrastructure is complete:
  - Epic 3 (Enhanced Interaction) — detail view foundation
  - Epic 13 (Multi-Source Aggregation) — cross-reference/linking patterns exist via enrichment DB
  - The `enrichment.CrossReference` system provides a model for parent-child relationships
- **No blocking relationships** with remaining active work (Epics 23, 24, 25, 26, 30).

### Story Impact
- No current or future stories require changes.
- New stories to be created within the new epic.
- Decision M8 (Link Relationship Types) adds `parent-of` and `child-of` relationship types which directly support this feature.

### Artifact Conflicts

| Artifact | Impact | Change Needed |
|----------|--------|---------------|
| PRD (`docs/prd/user-interface-design-goals.md`) | Lists E/F as core actions but no detailed spec | Add Expand/Fork feature specifications |
| PRD (`docs/prd/requirements.md`) | No specific requirements for Expand/Fork behavior | Add FR requirements for subtask/variant features |
| Architecture (`docs/architecture/data-models.md`) | Task struct lacks ParentID/SubtaskIDs | Add parent-child fields to Task model |
| Architecture (`docs/architecture/components.md`) | DetailView component spec doesn't mention subtask UI | Update DetailView component description |
| ROADMAP.md | Expand/Fork not listed | Add Epic 31 to Active Epics |
| Epics document | No Expand/Fork epic exists | Add Epic 31 with stories |
| `docs/design-decisions-needed.md` | H9 decided but not yet implemented | Mark H9 as "Implementation Planned (Epic 31)" |

### Technical Impact
- **Task model extension:** Add `ParentID *string` and computed subtask list to `core.Task`
- **Enrichment DB:** Use existing `CrossReference` with `parent-of`/`child-of` relationship types OR add native parent-child to Task struct
- **Detail view:** Enhanced expand flow with multi-subtask creation, subtask list rendering, fork with variant semantics
- **Doors view:** Optionally show subtask count badge on parent tasks
- **YAML persistence:** Add `parent_id` field to task YAML schema (backward-compatible — optional field)
- **Minimal blast radius:** Changes are additive to existing code paths

---

## Section 3: Recommended Approach

**Selected Path:** Direct Adjustment — Add new Epic 31 (Expand/Fork Key Implementations) with 4-5 stories.

**Rationale:**
- Both features are already stubbed in the codebase, reducing implementation risk
- The enrichment DB's cross-reference system provides infrastructure for relationship tracking
- Decision M8 already adds the `parent-of`/`child-of` relationship types needed
- The detail view's existing mode-based input system (`DetailModeExpandInput`) provides the UX pattern
- Estimated effort: 3-5 days total across all stories
- No architectural changes required — extends existing patterns

**Effort estimate:** Low-Medium
**Risk level:** Low
**Timeline impact:** None — parallel to other active work

---

## Section 4: Detailed Change Proposals

### 4.1 Task Model Changes

**File:** `internal/core/task.go`

```
OLD:
(no ParentID field)

NEW:
- ParentID  *string  `yaml:"parent_id,omitempty"`

Rationale: Track parent-child relationships natively in the task model.
Allow tasks to reference their parent, enabling subtask queries.
```

### 4.2 Expand Enhancement

**File:** `internal/tui/detail_view.go`

```
OLD:
- Single text input creates unlinked task via ExpandTaskMsg
- No subtask list rendering in detail view

NEW:
- Expand creates subtask with ParentID set to current task's ID
- Detail view shows subtask list when task has children
- Subtask count visible in status line
- User can add multiple subtasks in sequence

Rationale: Fulfill H9 Decision A — manual sub-task creation with parent-child tracking
```

### 4.3 Fork Enhancement

**File:** `internal/tui/detail_view.go`

```
OLD (line 150-152):
case "f", "F":
    forked := core.NewTask(dv.task.Text)
    return func() tea.Msg { return TaskAddedMsg{Task: forked} }

NEW:
case "f", "F":
    forked := core.ForkTask(dv.task)  // New factory method
    return func() tea.Msg { return TaskForkedMsg{Original: dv.task, Variant: forked} }

Rationale: ForkTask creates a variant by copying text+context but resetting
status to todo, clearing blocker, preserving effort, and adding a note
"Forked from: [original task text]". TaskForkedMsg allows the main model
to establish a cross-reference between original and variant.
```

### 4.4 PRD Updates

**File:** `docs/prd/requirements.md`

```
NEW requirements to add:
- FR-EXPAND-1: User can press [E] in detail view to create a subtask linked to the current task
- FR-EXPAND-2: Subtasks appear in the parent task's detail view with their status
- FR-EXPAND-3: Subtasks are eligible for door selection independently
- FR-FORK-1: User can press [F] in detail view to create a variant of the current task
- FR-FORK-2: Variants reset status to todo and clear blockers
- FR-FORK-3: Variants are cross-referenced to the original task via enrichment DB
```

### 4.5 Architecture Updates

**File:** `docs/architecture/data-models.md`

```
NEW: Add ParentID field to Task model
NEW: Document subtask query patterns (get children by ParentID)
NEW: Document ForkTask factory method semantics
```

---

## Section 5: Implementation Handoff

**Change Scope Classification:** Minor — Direct implementation by development team

**Handoff Plan:**
- **Development team:** Implement Epic 31 stories via standard story-driven workflow
- **No PO/SM coordination needed** — purely additive feature completion
- **No architect involvement needed** — extends existing patterns

**Implementation Order:**
1. Story 31.1: Task model ParentID extension + subtask query methods
2. Story 31.2: Enhanced Expand — subtask creation with parent-child linking
3. Story 31.3: Subtask list rendering in detail view + doors view badge
4. Story 31.4: Enhanced Fork — variant creation with ForkTask factory
5. Story 31.5: Cross-reference integration for fork variants (optional — may fold into 31.4)

**Success Criteria:**
- `[E]xpand` creates subtasks linked to parent via ParentID
- Subtasks visible in parent's detail view
- `[F]ork` creates variant with reset status, cross-referenced to original
- All new features have table-driven tests
- Backward-compatible YAML schema (existing tasks load without ParentID)
- `make lint && make test` pass with zero warnings

---

## Approval

**Status:** Approved (automated pipeline — YOLO mode)
**Next Steps:** Proceed to Party Mode for multi-agent discussion, then PRD updates, architecture, and epic/story creation.
