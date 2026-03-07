# Sprint Change Proposal — Task Dependencies & Blocked-Task Filtering

**Date:** 2026-03-07
**Change Type:** New Epic Proposal
**Scope Classification:** Moderate
**Triggered By:** UX & Workflow Improvements Research (docs/research/ux-workflow-improvements-research.md, Proposal #4)

---

## Section 1: Issue Summary

### Problem Statement

Tasks that depend on other tasks can currently appear in the Three Doors selection even when their prerequisites aren't done. The `blocked` status exists but requires manual flagging — the system doesn't understand task relationships structurally. This violates the core design principle #5: **"Ready means ready — the 3 doors should only show tasks the user can act on right now."**

### Discovery Context

Identified through systematic UX & Workflow Improvements Research as a P1 (High Impact, Medium Effort) improvement. Taskwarrior's dependency coefficient (8.0 in its urgency formula) demonstrates that dependency-aware filtering is foundational for user trust in task presentation systems.

### Supporting Evidence

1. **Research finding:** "Tasks whose dependencies aren't complete are automatically filtered from door selection" — ranked P1 in the priority matrix
2. **Existing infrastructure gap:** The `cross_references` table in the enrichment DB tracks generic "related" relationships, not dependency chains with directionality
3. **Manual workaround burden:** Users must manually set `blocked` status and remember to unblock when prerequisites complete — error-prone and creates maintenance friction
4. **Taskwarrior precedent:** Dependency coefficient of 8.0 (out of ~30 total urgency points) — the second-highest weighted factor after due dates, demonstrating industry validation of this feature's importance

---

## Section 2: Impact Analysis

### Epic Impact

- **No existing epics affected.** This is a purely additive change introducing a new Epic 29.
- **Epic 23 (CLI):** Future extension opportunity — `threedoors task deps add/rm/list` commands could be added in a later story
- **Epic 24 (MCP):** Future extension opportunity — `add_dependency` / `remove_dependency` MCP tools
- **Epic 28 (Snooze/Defer):** Complementary — both improve door pool quality through different mechanisms (time-based vs dependency-based filtering)

### Story Impact

- No current or future stories require modification
- New stories will be created under Epic 29

### Artifact Conflicts

| Artifact | Impact | Details |
|----------|--------|---------|
| PRD (requirements.md) | Addition | New requirements FR110-FR115 for dependency model, filtering, TUI indicators, auto-unblock |
| PRD (product-scope.md) | Addition | New phase section for Task Dependencies |
| Architecture (data-models.md) | Addition | `DependsOn []string` field on Task, DependencyGraph component |
| Architecture (components.md) | Addition | DependencyResolver component, modified GetAvailableForDoors, TUI blocked-by indicator |
| Epics & Stories | Addition | New Epic 29 with 4 stories |
| ROADMAP.md | Addition | New Epic 29 entry |

### Technical Impact

- **Task model extension:** Add `DependsOn []string` field (list of task IDs)
- **Door selection filter:** Modify `GetAvailableForDoors()` to exclude tasks with incomplete dependencies
- **Dependency resolution:** New `DependencyResolver` component for graph traversal and cycle detection
- **TUI changes:** "Blocked by: [task]" indicator on doors, dependency management in detail view
- **Auto-unblock:** When a dependency completes, check if any dependents are now unblocked
- **Provider impact:** TextFileProvider YAML serialization (straightforward), other providers map dependencies through enrichment DB

---

## Section 3: Recommended Approach

### Selected Path: Direct Adjustment (Add New Epic)

**Rationale:**
- This is a clean addition with no conflicts against in-progress work
- Medium implementation effort — extends existing patterns (Task model fields, GetAvailableForDoors filter, TUI views)
- Follows the established pattern set by Epic 28 (Snooze/Defer) — another "door pool quality" feature
- No rollback or MVP scope change needed

**Effort Estimate:** Medium (4 stories, ~2-3 worker sessions)
**Risk Assessment:** Low
- No breaking changes to existing functionality
- Dependency graph adds complexity but cycle detection is well-understood
- Cross-provider dependencies require enrichment DB storage (existing infrastructure)

**Timeline Impact:** None — slots into backlog alongside existing active epics

**Trade-offs Considered:**
- **Alternative: Extend existing cross_references** — Rejected. Cross-references are undirected "related" links. Dependencies require directionality (A depends on B, not B depends on A) and completion-state awareness. Overloading cross_references would create semantic confusion.
- **Alternative: Reuse blocked status** — Rejected. `blocked` is a manual status with a free-text reason. Dependencies are structural relationships between specific tasks. They serve different purposes and should coexist.

---

## Section 4: Detailed Change Proposals

### PRD Changes (requirements.md)

**Addition — New section "Phase 6+: Task Dependencies & Blocked-Task Filtering":**

```markdown
## Phase 6+ - Task Dependencies & Blocked-Task Filtering (Accepted)

*The following requirements add a native dependency graph for tasks, ensuring the Three Doors only present genuinely actionable tasks by automatically filtering those whose prerequisites are incomplete.*

**Task Dependencies:**

**FR110:** The system shall support a `depends_on` field on tasks containing a list of task IDs that must be completed before the task becomes actionable — stored as `depends_on: [task-id-1, task-id-2]` in YAML and persisted through the enrichment DB for cross-provider dependencies

**FR111:** Tasks whose dependencies include any task not in `complete` status shall be automatically excluded from door selection by `GetAvailableForDoors()` — the filter checks dependency completion state on every door refresh, requiring no manual status management

**FR112:** The system shall display a "Blocked by: [task text]" indicator on tasks in the doors view and detail view when they have incomplete dependencies, showing the first incomplete dependency's text (truncated to 40 characters) with a count of additional blockers if more than one exists

**FR113:** When a task is marked complete and other tasks depend on it, the system shall check all dependents — if a dependent's dependencies are now all complete, emit a `dependency_unblocked` notification event and refresh the doors view to potentially include the newly unblocked task

**FR114:** The system shall provide dependency management in the task detail view: `+` key to add a dependency (opens task search/picker), `-` key to remove a selected dependency, with the dependency list displayed in the detail view below the notes section

**FR115:** The system shall detect and reject circular dependencies when adding a new dependency — if adding dependency A->B would create a cycle (B already depends on A directly or transitively), the operation fails with a user-visible error message "Cannot add dependency: would create a circular chain"
```

### PRD Changes (product-scope.md)

**Addition — New phase section:**

```markdown
## Phase 3.5+: Task Dependencies & Blocked-Task Filtering

**In Scope:**
- `depends_on` field on Task model (list of task IDs)
- Automatic filtering of dependency-blocked tasks from door selection
- "Blocked by: [task]" indicator in doors view and detail view
- Auto-unblock notification when dependencies complete
- Dependency management in detail view (+/- keys)
- Circular dependency detection and rejection
- Session metrics logging for dependency events

**Out of Scope for this Phase:**
- Cross-provider dependency resolution (deferred — requires enrichment DB extension)
- CLI `threedoors task deps` commands (deferred to Epic 23 extension)
- MCP `add_dependency` / `remove_dependency` tools (deferred to Epic 24 extension)
- Visual dependency graph rendering (deferred — text indicators sufficient for v1)
```

### Architecture Changes

**New ADRs needed (to be created in architecture document):**

- **ADR-29.1:** DependsOn field as `[]string` (task IDs) on Task struct
- **ADR-29.2:** Dependency-aware door filtering in GetAvailableForDoors
- **ADR-29.3:** DependencyResolver component with topological sort and cycle detection
- **ADR-29.4:** Auto-unblock flow on task completion
- **ADR-29.5:** TUI blocked-by indicator design

### Epic & Story Changes

**New Epic 29: Task Dependencies & Blocked-Task Filtering** with 4 stories (detailed in epics document)

---

## Section 5: Implementation Handoff

### Change Scope: Moderate

Requires backlog addition (new epic) and PRD/architecture updates, but no existing work needs modification.

### Handoff Plan

| Role | Responsibility |
|------|---------------|
| **PM Agent** | Update PRD with FR110-FR115, update product-scope.md |
| **Architect Agent** | Create architecture decision document for Epic 29 |
| **SM Agent** | Add Epic 29 to sprint planning, update ROADMAP.md |
| **Dev Workers** | Implement stories 29.1-29.4 after architecture is approved |

### Success Criteria

1. PRD updated with FR110-FR115 requirements
2. Architecture document created with ADRs 29.1-29.5
3. Epic 29 with 4 stories created and validated
4. ROADMAP.md updated with Epic 29 entry
5. All stories pass implementation readiness check

### Next Steps

1. Run Party Mode to get multi-agent consensus on the proposal
2. Update PRD with accepted requirements
3. Create architecture decision document
4. Create epics and stories breakdown
5. Add to ROADMAP.md and sprint planning

---

**Proposal Status:** Ready for review
**Recommended Action:** Approve and proceed with BMAD pipeline (PRD edit -> Architecture -> Epics & Stories)
