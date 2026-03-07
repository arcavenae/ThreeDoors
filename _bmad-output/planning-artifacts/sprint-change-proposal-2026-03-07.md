# Sprint Change Proposal: Daily Planning Mode

**Date:** 2026-03-07
**Proposed By:** BMAD Course Correction Workflow
**Change Scope:** Major — New Epic Proposal
**Status:** Approved for Planning

---

## Section 1: Issue Summary

### Problem Statement

ThreeDoors currently operates as a **reactive** tool — users open it, pick a door, close it. It lacks the **proactive rituals and rhythms** that make productivity tools sticky. Without a daily planning engagement hook, the tool misses the strongest retention mechanism in productivity software: morning planning rituals.

Research shows daily planning rituals are the primary driver behind Sunsama's 95% retention rate. ThreeDoors has all the infrastructure for this (session tracking, mood capture, task pool management, calendar awareness) but no guided planning flow.

### Context

- Identified during UX & Workflow Improvements Research (`docs/research/ux-workflow-improvements-research.md`)
- Ranked **P0 priority** with **High impact** and **Medium effort**
- Part of "Phase 1: Daily Engagement Loop" in the research recommendations
- Research identifies this as addressing abandonment reason #3 (maintenance burden) by creating a daily rhythm

### Evidence

1. **Sunsama retention data**: 95% retention attributed to daily planning ritual
2. **Abandonment research**: 73% task manager churn within 30 days; daily rituals directly counter this
3. **Existing infrastructure**: Session metrics, mood capture, calendar awareness, values/goals display all exist but aren't connected into a planning flow
4. **User behavior gap**: No guided "what am I committing to today?" flow exists — users only get reactive "what should I do next?"

---

## Section 2: Impact Analysis

### Epic Impact

**No existing epics are affected.** This is a net-new epic proposal that builds on completed infrastructure:

| Dependency | Epic | Status | What It Provides |
|-----------|------|--------|-----------------|
| Session tracking | Epic 1 (Story 1.5) | Complete | Session metrics, daily engagement data |
| Mood capture | Epic 3 | Complete | Energy-adjacent UX, mood state tracking |
| Values/goals display | Epic 3 | Complete | Guided multi-step TUI flow patterns |
| Calendar awareness | Epic 12 | Complete | Available time blocks for planning |
| Task categorization | Epic 4 | Complete | Task type, effort level, context tags |
| Onboarding flow | Epic 10 | Complete | Multi-step wizard UX patterns |

**No existing epics need modification, removal, or resequencing.**

### Story Impact

- No current or future stories require changes
- New stories will be created within the new epic
- No dependency conflicts with in-progress work (Epics 23, 24, 25, 26)

### Artifact Updates Required

| Artifact | Change Needed |
|----------|--------------|
| **PRD** | Add new functional requirements (FR97-FR103) for daily planning mode |
| **Architecture** | Add PlanningView component to TUI layer, "today's focus" concept to core domain |
| **Epic List** | Add Epic 27: Daily Planning Mode |
| **ROADMAP.md** | Add Epic 27 to active epics |

### Technical Impact

- New TUI view (`PlanningView`) following existing Bubbletea MVU patterns
- New "today's focus" transient state on tasks (tag-based or session-scoped)
- Energy level prompt (extends existing mood capture patterns)
- Planning session timer (reuses `tea.Tick` pattern from existing codebase)
- No infrastructure, deployment, or breaking changes

---

## Section 3: Recommended Approach

### Selected Path: Direct Adjustment (Add New Epic)

**Rationale:**
- This is a clean addition — no existing work needs modification
- All prerequisite infrastructure is complete
- The feature is self-contained within a new epic
- Medium effort with high retention impact
- No timeline risk to in-progress epics (23, 24, 25, 26)

**Effort Estimate:** Medium (4-6 stories, 2-3 weeks at 2-4 hrs/week)
**Risk Level:** Low — builds on proven patterns and existing infrastructure
**Timeline Impact:** None on current work; adds to roadmap as parallel/sequential P1 epic

### Alternatives Considered

- **Rollback**: N/A (new feature, nothing to revert)
- **MVP Review**: Not needed (post-MVP feature that doesn't affect current scope)
- **Defer**: Not recommended — research shows this is the highest-impact retention feature available

---

## Section 4: Detailed Change Proposals

### 4.1 PRD Changes

**Add new requirements section: Daily Planning Mode**

```
NEW:
FR97: The system shall provide a daily planning mode accessible via `threedoors plan` CLI command
or `:plan` TUI command that guides users through a structured morning planning ritual

FR98: The planning mode shall present yesterday's incomplete tasks with options to continue,
defer, or drop each task

FR99: The planning mode shall allow users to select 3-5 tasks as "today's focus" from the
full task pool, with focused tasks receiving priority in door selection

FR100: The planning mode shall prompt for current energy level (high/medium/low) and use
this to filter focus task suggestions to match available energy

FR101: The planning session shall be time-boxed (5-10 minutes) with a visible progress
indicator showing completion percentage

FR102: Today's focus tasks shall receive elevated priority in door selection scoring,
appearing more frequently as doors until completed or the day ends

FR103: The system shall track planning session completion as a daily engagement metric
in session logs

Rationale: Daily planning rituals are the strongest retention mechanism in productivity
tools. This addresses the gap between ThreeDoors' reactive usage pattern and the proactive
daily engagement that drives long-term retention.
```

### 4.2 Architecture Changes

**Add to TUI Layer (`internal/tui`):**
- `PlanningView` — new Bubbletea model implementing the planning flow
- Follows existing multi-step flow patterns (onboarding wizard, values setup)
- Sub-views: ReviewView (yesterday's tasks), SelectView (focus picker), EnergyView (energy prompt)

**Add to Core Domain (`internal/core` or `internal/tasks`):**
- "Today's focus" concept — either transient session state or tag-based (`+focus`)
- Planning session metrics — extends existing JSONL session logging
- Focus-aware door selection scoring — boost for focus-tagged tasks

**No changes to:**
- Adapter layer (planning is UI + core domain only)
- Sync engine (focus state is local/session-scoped)
- Intelligence layer (calendar awareness already exists, can be consumed)

### 4.3 Epic Addition

**Add Epic 27: Daily Planning Mode** to the epic list and ROADMAP.md with 4-6 stories covering:
1. Planning data model and focus state
2. Review incomplete tasks flow
3. Focus selection flow
4. Energy level matching
5. Planning session metrics
6. CLI `plan` subcommand

---

## Section 5: Implementation Handoff

### Change Scope Classification: Moderate

**Requires:**
- Product Owner/SM: Backlog addition and prioritization
- Architect: Architecture doc updates (minimal — follows existing patterns)
- Development team: Story implementation

### Handoff Plan

| Role | Responsibility |
|------|---------------|
| **PM (BMAD)** | Update PRD with FR97-FR103, update epic list |
| **Architect (BMAD)** | Define architecture for PlanningView and focus state |
| **SM (BMAD)** | Create epic and stories, add to ROADMAP.md |
| **Dev workers** | Implement stories in dependency order |

### Success Criteria

1. Daily planning flow accessible via `:plan` command in TUI
2. Users can review yesterday's incomplete tasks (continue/defer/drop)
3. Users can select 3-5 focus tasks for the day
4. Energy level prompt filters focus suggestions
5. Planning session is time-boxed with progress indicator
6. Focus tasks appear with elevated priority in door selection
7. Planning completion tracked in session metrics

---

## Approval

**Approved for planning pipeline execution.** Proceeding to:
1. Party Mode discussion for agent consensus
2. PRD update with new requirements
3. Architecture definition
4. Epic and story creation
