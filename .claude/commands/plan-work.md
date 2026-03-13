# Plan Work — Research to Stories Pipeline

This is the complete planning workflow that takes research findings and formalizes them into actionable epics and stories via proper BMAD agents. It is the middle step in the three-step pipeline:

1. **Research** (ad-hoc directives produce artifacts)
2. **`/plan-work`** (THIS — formalizes research into PRD/architecture/epics/stories)
3. **`/implement-story`** (TDD implementation of individual stories)

**Input:** The user provides a research artifact path or description as $ARGUMENTS. If no argument is provided, ask what research to formalize.

<critical>
- Execute ALL phases in order. Do NOT skip phases.
- Do NOT stop between phases unless a HALT condition is triggered.
- Each phase builds on the previous — context must carry forward.
- ALWAYS invoke BMAD agents via their slash commands (`/bmad-bmm-agent-pm`, `/bmad-bmm-agent-sm`, etc.) to ensure all instructions are properly loaded. NEVER ad-hoc prompt agent behavior without loading the agent first.
- All outputs go into a SINGLE PR at the end — do not create intermediate PRs.
</critical>

---

## Phase 1: Load Research Context

**Goal:** Understand the research findings that need to be formalized.

1. Locate and read the research artifact specified by "$ARGUMENTS":
   - Check `_bmad-output/planning-artifacts/` for the artifact
   - Check `docs/decisions/BOARD.md` for related entries (Active Research, Pending Recommendations)
   - If the input is a description rather than a path, search for matching artifacts
2. Summarize the key findings, recommendations, and any decisions already made.
3. Check `ROADMAP.md` to understand current scope and where this work fits.
4. Check `docs/decisions/BOARD.md` for related prior decisions or open questions.

### HALT condition:
- If no research artifact can be found and the description is too vague to proceed, ask the user for clarification.

---

## Phase 2: PM Course Correction & PRD Analysis

**Goal:** Have the PM agent evaluate the research, decide on approach, AND assess impact on PRD content docs.

1. Launch `/bmad-bmm-agent-pm` (this properly loads all PM instructions).
2. Ask the PM to run a course correction (`/bmad-bmm-correct-course`) with the research findings as input.
3. The PM should produce:
   - Problem statement and impact analysis
   - Proposed approach with rationale
   - Rejected alternatives with reasons
   - Sprint change proposal saved to `_bmad-output/planning-artifacts/sprint-change-proposal-{date}-{slug}.md`
4. **PRD Impact Assessment (MANDATORY):** As part of the course correction, the PM MUST:
   - Read `docs/prd/requirements.md` and identify what new functional/non-functional requirements this work introduces
   - Read `docs/prd/product-scope.md` and identify which phase/section the new scope belongs to
   - Read `docs/prd/epic-details.md` and confirm no detail entry exists yet for the new epic
   - Include a "PRD Changes Required" section in the sprint change proposal listing specific additions needed for each content doc

<critical>
The PM's course correction MUST include PRD content doc analysis. If the sprint change proposal does not have a "PRD Changes Required" section, HALT and ask the PM to complete it. This is the root cause of PRD drift — see `_bmad-output/planning-artifacts/prd-process-bypass-investigation.md`.
</critical>

---

## Phase 3: Party Mode Validation

**Goal:** Multi-agent deliberation on the proposed approach AND the PRD changes.

1. Run `/bmad-party-mode` with the sprint change proposal as context.
2. Include at minimum: PM, Architect, UX Designer.
3. Include TEA if the proposal has significant testing implications.
4. Include relevant CIS agents (innovation strategist, design thinking coach) if the proposal involves UX or product innovation.
5. The party mode discussion should validate or refine:
   - Is the proposed approach correct?
   - Are there architectural concerns?
   - What UX considerations exist?
   - What is the right scope and priority?
   - **Are the proposed PRD content doc changes complete and correct?** (Review the "PRD Changes Required" section from the sprint change proposal)
   - **Are the new functional requirements properly scoped?** (Not too broad, not missing edge cases)
6. Save the party mode artifact to `_bmad-output/planning-artifacts/{topic}-party-mode.md`.
7. Apply accepted recommendations to the sprint change proposal.

---

## Phase 4: Architecture Review

**Goal:** Update architecture documentation if the work requires architectural changes.

1. If the party mode discussion identified architectural implications:
   - Launch `/bmad-bmm-agent-architect` (properly loads architect instructions)
   - Ask the architect to review and update `docs/architecture/` docs as needed
   - If new ADRs are warranted, create them in `docs/ADRs/`
2. If the work is purely feature/UX with no architectural impact, skip to Phase 5.

---

## Phase 5: PRD Content Doc Update

**Goal:** Update the PRD content docs with new requirements, scope, and epic details BEFORE creating index entries.

<critical>
This phase addresses the "Built But Not Wired" pattern where `/plan-work` historically updated index docs (epic-list, epics-and-stories, ROADMAP) but SKIPPED content docs (requirements, product-scope, epic-details). This resulted in 47+ epics with no PRD coverage. See `_bmad-output/planning-artifacts/prd-process-bypass-investigation.md`.

ALL THREE content docs MUST be updated. Do NOT skip to Phase 6.
</critical>

1. Launch `/bmad-bmm-agent-pm` (properly loads PM instructions).
2. Using the "PRD Changes Required" section from the sprint change proposal (Phase 2) and party mode feedback (Phase 3), the PM MUST update:
   - **`docs/prd/requirements.md`**: Add new functional requirements (FR) and non-functional requirements (NFR) for every new capability. Each requirement needs an ID, description, and acceptance criteria.
   - **`docs/prd/product-scope.md`**: Add the new feature/epic to the appropriate phase section. If the feature doesn't fit an existing phase, create a new phase section or extend the most relevant one.
   - **`docs/prd/epic-details.md`**: Add a detailed breakdown for the new epic including: epic title, description, goals, key features, technical considerations, and story list.
3. Verify each content doc was actually modified (not just opened and closed). Check `git diff` to confirm changes.

---

## Phase 6: PRD Index Update

**Goal:** Update the PRD index docs and roadmap with new epic/story entries.

1. The PM should:
   - Allocate an epic number (PM is the authority for epic numbers)
   - Update `docs/prd/epics-and-stories.md` with the new epic and story outlines
   - Update `docs/prd/epic-list.md` with the new epic entry
   - Update `ROADMAP.md` with the new epic and stories
2. Ensure all index docs are consistent with each other AND with the content docs updated in Phase 5.

---

## Phase 7: Story Creation

**Goal:** Create detailed, implementation-ready story files.

1. Launch `/bmad-bmm-agent-sm` (properly loads SM instructions).
2. For each story identified in the epic:
   - Use `/bmad-bmm-create-story` to create the story file
   - Ensure each story has: acceptance criteria, task breakdown, dev notes, dependencies
   - Story files go in `docs/stories/{epic}.{story}.story.md`
3. Verify all stories reference the sprint change proposal and party mode artifact.
4. Set the status of each newly created story to `Not Started`.

<critical>
Planning workers NEVER set story status to `Done` — that is reserved for `/implement-story` workers who have completed all acceptance criteria in code. A story file being *created* is not the same as a story being *implemented*.
</critical>

---

## Phase 8: Decision Recording

**Goal:** Record all decisions made during this planning process.

1. Update `docs/decisions/BOARD.md`:
   - Move any related items from "Active Research" or "Pending Recommendations" to "Decided" (if decisions were finalized)
   - Add new entries for decisions made during party mode
   - Link to the party mode artifact and sprint change proposal
2. If prior BOARD entries exist for this research, update them rather than creating duplicates.

---

## Phase 9: Create Pull Request

**Goal:** Package all planning artifacts into a single PR.

**IMPORTANT:** Do NOT set any story status to `Done`. Stories created by `/plan-work` must have status `Not Started` or `Draft`. Only `/implement-story` sets `Done` after all acceptance criteria are met in code.

### PRD Completeness Validation Gate (MANDATORY)

Before creating the PR, verify ALL of the following. If any check fails, go back to Phase 5 and complete the missing updates:

- [ ] `docs/prd/requirements.md` has new FR/NFR entries for every new capability introduced by this epic
- [ ] `docs/prd/product-scope.md` has the feature in the correct phase section
- [ ] `docs/prd/epic-details.md` has a detailed breakdown for the new epic
- [ ] `docs/prd/epics-and-stories.md` has the epic and story outlines
- [ ] `docs/prd/epic-list.md` has the epic entry
- [ ] `ROADMAP.md` has the epic and stories

<critical>
This validation gate exists because historically `/plan-work` updated only the index docs (epic-list, epics-and-stories, ROADMAP) and skipped the content docs (requirements, product-scope, epic-details), resulting in 47+ epics with no PRD coverage. Do NOT create the PR if content docs are missing updates.
</critical>

### PR Creation

1. Stage all changed/created files:
   - Sprint change proposal artifact
   - Party mode artifact
   - Architecture docs (if updated)
   - PRD content doc updates (requirements.md, product-scope.md, epic-details.md)
   - PRD index doc updates (epics-and-stories.md, epic-list.md)
   - ROADMAP.md updates
   - Story files
   - BOARD.md updates
   - Any new ADRs
2. Create a descriptive commit: `docs: plan {topic} — epic {N}, stories {N.1}-{N.X}`
3. Push the branch and create a PR with:
   - Title: `docs: {topic} planning — Epic {N} with {X} stories`
   - Body summarizing: research findings, approach adopted, rejected alternatives, stories created, architectural changes, **PRD docs updated**
4. Report the PR URL.

---

## Output Checklist

Before completing, verify ALL of these exist:

- [ ] Sprint change proposal in `_bmad-output/planning-artifacts/` (with "PRD Changes Required" section)
- [ ] Party mode artifact in `_bmad-output/planning-artifacts/`
- [ ] **PRD Content Docs Updated:**
  - [ ] `docs/prd/requirements.md` — new FR/NFR entries for every new capability
  - [ ] `docs/prd/product-scope.md` — feature added to correct phase section
  - [ ] `docs/prd/epic-details.md` — detailed breakdown for new epic
- [ ] **PRD Index Docs Updated:**
  - [ ] `docs/prd/epics-and-stories.md` — epic and story outlines
  - [ ] `docs/prd/epic-list.md` — epic entry
  - [ ] `ROADMAP.md` — epic and stories
- [ ] Story files created in `docs/stories/`
- [ ] BOARD.md updated with decisions
- [ ] Architecture docs updated (if applicable)
- [ ] Single PR with all artifacts

---

## HALT Conditions

Stop and ask the user if:
- The research conflicts with an existing architectural decision (requires human override)
- The PM cannot determine scope/priority (needs human input on roadmap placement)
- Party mode agents fundamentally disagree on approach (needs human tiebreaker)
- The work would require a new P0 epic (roadmap priority changes require human approval)
- Epic number collision detected (verify against BOARD.md Epic Number Registry)
