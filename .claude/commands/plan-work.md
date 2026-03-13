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

## Phase 2: PM Course Correction

**Goal:** Have the PM agent evaluate the research and decide on approach.

1. Launch `/bmad-bmm-agent-pm` (this properly loads all PM instructions).
2. Ask the PM to run a course correction (`/bmad-bmm-correct-course`) with the research findings as input.
3. The PM should produce:
   - Problem statement and impact analysis
   - Proposed approach with rationale
   - Rejected alternatives with reasons
   - Sprint change proposal saved to `_bmad-output/planning-artifacts/sprint-change-proposal-{date}-{slug}.md`

---

## Phase 3: Party Mode Validation

**Goal:** Multi-agent deliberation on the proposed approach.

1. Run `/bmad-party-mode` with the sprint change proposal as context.
2. Include at minimum: PM, Architect, UX Designer.
3. Include TEA if the proposal has significant testing implications.
4. Include relevant CIS agents (innovation strategist, design thinking coach) if the proposal involves UX or product innovation.
5. The party mode discussion should validate or refine:
   - Is the proposed approach correct?
   - Are there architectural concerns?
   - What UX considerations exist?
   - What is the right scope and priority?
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

## Phase 5: PRD Update

**Goal:** Update the Product Requirements Document with new epics/features.

1. Launch `/bmad-bmm-agent-pm` (properly loads PM instructions).
2. Ask the PM to:
   - Allocate an epic number (PM is the authority for epic numbers)
   - Update `docs/prd/epics-and-stories.md` with the new epic and story outlines
   - Update `docs/prd/epic-list.md` with the new epic entry
   - Update `ROADMAP.md` with the new epic and stories
3. Ensure all three planning docs are consistent.

---

## Phase 6: Story Creation

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

## Phase 7: Decision Recording

**Goal:** Record all decisions made during this planning process.

1. Update `docs/decisions/BOARD.md`:
   - Move any related items from "Active Research" or "Pending Recommendations" to "Decided" (if decisions were finalized)
   - Add new entries for decisions made during party mode
   - Link to the party mode artifact and sprint change proposal
2. If prior BOARD entries exist for this research, update them rather than creating duplicates.

---

## Phase 8: Create Pull Request

**Goal:** Package all planning artifacts into a single PR.

**IMPORTANT:** Do NOT set any story status to `Done`. Stories created by `/plan-work` must have status `Not Started` or `Draft`. Only `/implement-story` sets `Done` after all acceptance criteria are met in code.

1. Stage all changed/created files:
   - Sprint change proposal artifact
   - Party mode artifact
   - Architecture docs (if updated)
   - PRD updates (epics-and-stories.md, epic-list.md)
   - ROADMAP.md updates
   - Story files
   - BOARD.md updates
   - Any new ADRs
2. Create a descriptive commit: `docs: plan {topic} — epic {N}, stories {N.1}-{N.X}`
3. Push the branch and create a PR with:
   - Title: `docs: {topic} planning — Epic {N} with {X} stories`
   - Body summarizing: research findings, approach adopted, rejected alternatives, stories created, architectural changes
4. Report the PR URL.

---

## Output Checklist

Before completing, verify ALL of these exist:

- [ ] Sprint change proposal in `_bmad-output/planning-artifacts/`
- [ ] Party mode artifact in `_bmad-output/planning-artifacts/`
- [ ] Epic added to `docs/prd/epics-and-stories.md`
- [ ] Epic added to `docs/prd/epic-list.md`
- [ ] Epic added to `ROADMAP.md`
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
