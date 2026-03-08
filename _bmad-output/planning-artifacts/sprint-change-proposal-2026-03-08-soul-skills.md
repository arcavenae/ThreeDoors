# Sprint Change Proposal: SOUL.md + Custom Multiclaude Skills

**Date:** 2026-03-08
**Triggered by:** Research spike (docs/research/ai-tooling-findings.md)
**Scope Classification:** Moderate — New epic, PRD updates, no code changes
**Proposed Epic:** Epic 34: SOUL.md + Custom Multiclaude Skills

---

## Section 1: Issue Summary

### Problem Statement

ThreeDoors has extensive AI-assisted development infrastructure (BMAD framework, 50+ slash commands, MCP server, self-driving dev pipeline) but lacks two critical pieces:

1. **No SOUL.md** — The project's philosophy ("progress over perfection", "three doors not three hundred", "work with human nature") is scattered across story files and the PRD. AI agents make isolated decisions without consistent personality/behavioral guidelines. Each agent independently interprets the product vision.

2. **No project-specific development skills** — Despite having 50+ BMAD commands, there are zero ThreeDoors-specific slash commands. Common workflows (pre-PR validation, adapter compliance checks, story creation, design pattern enforcement) are repeated manually in every story file, averaging ~40-50 duplicated lines per story.

### Evidence

- **11 story files** each contain near-identical Pre-PR Submission Checklist blocks (~10 lines each)
- **~500 lines** of duplicated content across existing stories (coding standards, atomic write pattern, error wrapping, MVU reminders)
- **50+ BMAD commands** in `.claude/commands/` but **0 ThreeDoors-specific commands**
- Historical CI failures (PRs #9, #10, #16, #23, #24) from missed formatting/linting steps that a `/pre-pr` skill would have caught
- CLAUDE.md already exists with coding standards but has no personality/philosophy layer

### Discovery Context

Research agent (zealous-rabbit) conducted a comprehensive analysis on 2026-03-02, producing 7 findings with priority recommendations. The CLAUDE.md recommendation (Finding 1) has already been implemented. Findings 2-4 (SOUL.md, custom skills, DRY reduction) remain unaddressed.

---

## Section 2: Impact Analysis

### Epic Impact

**No existing epics are affected.** This is a purely additive proposal that creates a new Epic 34.

| Aspect | Impact |
|--------|--------|
| Current epics (23, 24, 25, 26, 30-33) | No changes needed |
| Epic dependencies | No dependencies on or from other epics |
| Epic ordering | No resequencing needed |
| Priority | P1 — Improves all future development velocity |

### Story Impact

- **Existing stories:** Will be retroactively updated to remove duplicated content and align with current project standards (Story 34.5)
- **Future stories:** All future stories benefit by removing ~40-50 lines of boilerplate (Pre-PR Checklist, coding standards reminders, pattern documentation)
- **Story file template:** Will be updated via `/new-story` skill to reference CLAUDE.md instead of embedding standards
- **Living documentation policy:** Completed stories are updated when code improvements diverge from spec descriptions — specs must always reflect reality

### Artifact Conflicts

| Artifact | Conflict? | Changes Needed |
|----------|-----------|----------------|
| PRD | Minor addition | Add FR138-FR142 (SOUL.md, skills requirements) |
| Architecture | Minor addition | Add section on SOUL.md integration and skills |
| UI/UX | None | No user-facing changes |
| CLAUDE.md | None | Already exists, SOUL.md complements it |
| CI/CD | None | No pipeline changes |
| ROADMAP.md | Addition | Add Epic 34 to active epics |

### Technical Impact

- **Code changes:** None — SOUL.md and skills are documentation/configuration files
- **Infrastructure:** None
- **Deployment:** None
- **Testing:** Skills can be manually tested; no automated test infrastructure needed

---

## Section 3: Recommended Approach

### Selected Path: Direct Adjustment (Option 1)

**Rationale:** This is a purely additive, low-risk change that creates configuration and documentation files. No code changes, no architectural shifts, no rollbacks needed.

| Factor | Assessment |
|--------|-----------|
| Effort | Low — primarily writing/configuration |
| Risk | Low — no code changes, no breaking changes |
| Timeline impact | None — can run in parallel with other work |
| Long-term value | High — improves every future story's quality and reduces boilerplate |

### Alternatives Considered

- **Rollback (Option 2):** Not applicable — no existing work to revert
- **MVP Review (Option 3):** Not applicable — this is post-MVP enhancement that doesn't affect any MVP scope

---

## Section 4: Detailed Change Proposals

### 4.1: Create SOUL.md

**File:** `SOUL.md` (project root)

**Content:** Project philosophy document capturing:
- What ThreeDoors is (personal achievement partner)
- Core philosophy: Progress Over Perfection, Work With Human Nature, Three Doors Not Three Hundred, Local-First Privacy-Always, Meet Users Where They Are, Solo Dev Reality
- Design principles for AI agents (friction reduction, simplicity, data respect, pattern following)
- What ThreeDoors is NOT (not project management, not habit tracker, not second brain)
- The feeling we're going for ("a friend saying pick one and go")

**Rationale:** Based on Finding 2 from research. Captures the "why" behind decisions so agents make aligned choices when stories don't specify every detail. Referenced from CLAUDE.md.

### 4.2: Create /pre-pr Skill

**File:** `.claude/commands/pre-pr.md`

**Content:** Automated 7-step pre-PR validation:
1. Branch freshness check (git fetch + rebase status)
2. Formatting check (gofumpt -l .)
3. Linting (golangci-lint run ./...)
4. Tests (go test ./... -count=1)
5. Dead code check (go vet ./...)
6. Scope review (git diff --stat)
7. Commit cleanliness (check for fixup/wip commits)

**Rationale:** P0 priority from research. Automates the most-violated checklist. Would have prevented CI failures in PRs #9, #10, #16, #23, #24.

### 4.3: Create /validate-adapter Skill

**File:** `.claude/commands/validate-adapter.md`

**Content:** TaskProvider compliance checker:
- Read TaskProvider interface
- Find all implementations
- Verify method coverage, error wrapping, factory registration, test files, atomic writes
- Report compliance table

**Rationale:** P2 priority. Prevents adapter compliance issues as new integrations are added (Todoist, GitHub Issues, Linear).

### 4.4: Create /check-patterns Skill

**File:** `.claude/commands/check-patterns.md`

**Content:** Design pattern violation scanner:
- Direct status mutation without IsValidTransition()
- Direct file writes bypassing atomic pattern
- fmt.Println in TUI code
- Panics in user code
- Provider instantiation outside factory
- Missing error wrapping

**Rationale:** P2 priority. Catches anti-patterns before PR creation.

### 4.5: Create /new-story Skill

**File:** `.claude/commands/new-story.md`

**Content:** Story template generator:
- Prompts for story ID and title
- Uses latest story format as reference
- Creates docs/stories/{id}.story.md with standard structure
- References CLAUDE.md instead of embedding checklist
- Only includes story-specific dev notes

**Rationale:** P3 priority. Ensures consistent story format, removes boilerplate from creation.

### 4.6: PRD Updates

**Section:** Requirements — add new functional requirements

```
FR138: The system shall maintain a SOUL.md document at the project root defining
the project's philosophy, design principles, and behavioral guidelines for AI
agents — ensuring consistent decision-making aligned with ThreeDoors values
(progress over perfection, work with human nature, local-first)

FR139: The project shall provide a /pre-pr skill that automates the 7-step
pre-PR validation checklist (branch freshness, formatting, linting, tests, dead
code, scope review, commit cleanliness) — reducing CI failures and enforcing
NFR-CQ1 through NFR-CQ5

FR140: The project shall provide a /validate-adapter skill that checks
TaskProvider implementations for interface compliance, error wrapping patterns,
factory registration, test coverage, and atomic write usage

FR141: The project shall provide a /check-patterns skill that scans the codebase
for design pattern violations (direct status mutation, direct file writes,
fmt.Println in TUI, panics, factory bypass, missing error wrapping)

FR142: The project shall provide a /new-story skill that generates story files
from a standard template, referencing CLAUDE.md for coding standards and
pre-PR checklists instead of embedding them
```

---

## Section 5: Implementation Handoff

### Scope Classification: Minor

This is a documentation and configuration change. No code modifications. Can be implemented directly by a development worker agent.

### Handoff Plan

| Role | Responsibility |
|------|---------------|
| PM agent | Update PRD with FR138-FR142 |
| Architect agent | Create architecture section for SOUL.md + skills integration |
| SM agent | Create Epic 34 breakdown with stories |
| Dev worker | Implement SOUL.md, create skill files, update ROADMAP.md |

### Success Criteria

1. SOUL.md exists at project root and is referenced from CLAUDE.md
2. Four custom skills (/pre-pr, /validate-adapter, /check-patterns, /new-story) exist in `.claude/commands/`
3. /pre-pr skill successfully runs all 7 validation steps
4. PRD updated with FR138-FR142
5. ROADMAP.md updated with Epic 34
6. Future stories reference CLAUDE.md Pre-PR Checklist instead of embedding it

### Epic 34 Story Breakdown (Proposed)

| Story | Title | Priority | Effort |
|-------|-------|----------|--------|
| 34.1 | Create SOUL.md Project Philosophy Document | P1 | Small |
| 34.2 | Create /pre-pr Pre-PR Validation Skill | P0 | Small |
| 34.3 | Create /validate-adapter and /check-patterns Skills | P2 | Small |
| 34.4 | Create /new-story Story Template Skill | P3 | Small |
| 34.5 | DRY Existing Story Files & Retroactive Spec Alignment | P1 | Medium |

---

## Change Navigation Checklist Results

### Section 1: Understand the Trigger and Context
- [x] 1.1: Triggering source identified (research spike, not a story)
- [x] 1.2: Core problem defined (missing SOUL.md, missing project-specific skills, ~500 lines duplication)
- [x] 1.3: Evidence gathered (11 stories, 50+ BMAD commands, 0 project commands, historical CI failures)

### Section 2: Epic Impact Assessment
- [x] 2.1: No current epics affected (additive proposal)
- [x] 2.2: New epic proposed (Epic 34)
- [x] 2.3: No future epics impacted
- [x] 2.4: No epics invalidated
- [x] 2.5: No resequencing needed

### Section 3: Artifact Conflict Analysis
- [x] 3.1: PRD needs minor additions (FR138-FR142)
- [x] 3.2: Architecture needs minor addition (SOUL.md + skills section)
- [N/A] 3.3: No UI/UX impact
- [x] 3.4: ROADMAP.md update needed

### Section 4: Path Forward Evaluation
- [x] 4.1: Direct Adjustment — Viable (Low effort, Low risk)
- [N/A] 4.2: Rollback — Not applicable
- [N/A] 4.3: MVP Review — Not applicable
- [x] 4.4: Selected approach: Direct Adjustment

### Section 5: Sprint Change Proposal
- [x] 5.1: Issue summary created
- [x] 5.2: Epic and artifact impact documented
- [x] 5.3: Recommended path with rationale
- [x] 5.4: PRD MVP impact (none) and action plan
- [x] 5.5: Agent handoff plan established

### Section 6: Final Review
- [x] 6.1: All sections addressed
- [x] 6.2: Proposal is consistent and actionable
- [x] 6.3: Approved (YOLO mode — automated pipeline)
- [x] 6.4: Sprint status update pending (Epic 34 to be added)
- [x] 6.5: Handoff responsibilities defined

---

**Correct Course workflow complete, arcaven!**

Next steps: PM agent updates PRD → Architect creates architecture → SM creates stories → Dev implements.
