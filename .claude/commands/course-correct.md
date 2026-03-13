---
name: 'course-correct'
description: 'Formalize a course correction: describe a problem, create a sprint change proposal, optionally run party mode validation, and generate stories. Usage: /course-correct <problem description>'
---

# Course Correction Workflow

**Goal:** Turn a detected problem into a structured correction with stories ready for implementation.

**Input:** The user provides a problem description as $ARGUMENTS.

## Step 1: Problem Analysis

1. Read ROADMAP.md, SOUL.md, and recent CHANGELOG.md entries for context.
2. Analyze the problem described in $ARGUMENTS:
   - What is the observable symptom?
   - What is the root cause (or best hypothesis)?
   - What is the blast radius (how many epics/stories/agents are affected)?
   - Is this a repeat of a previous course correction? Check `_bmad-output/planning-artifacts/` for related artifacts.

## Step 2: Draft Sprint Change Proposal

Create a sprint change proposal at `_bmad-output/planning-artifacts/sprint-change-proposal-{date}-{slug}.md` with:

```markdown
# Sprint Change Proposal: {title}
**Date:** {today}
**Triggered by:** {problem description}
**Severity:** {Low | Medium | High | Critical}

## Problem Statement
{What's wrong and why it matters}

## Impact Analysis
{What's affected, what breaks if we don't fix it}

## Proposed Approach
{The fix, with rationale}

## Rejected Alternatives
{What we considered and why we didn't choose it}

## Stories Required
{List of stories needed, with brief descriptions}

## Effort Estimate
{T-shirt size for total effort}
```

## Step 3: Party Mode Validation (Optional)

Ask the user: "Run party mode to validate this proposal? (y/n)"

If yes:
1. Run `/bmad-party-mode` with the proposal as context
2. Incorporate feedback into the proposal
3. Update the artifact with party mode consensus

If no: proceed directly to Step 4.

## Step 4: Create Stories

For each story identified in the proposal:
1. Create a story file in `docs/stories/` using the project's story template
2. Set status to `Not Started`
3. Reference the sprint change proposal in the story's context section

## Step 5: Update PRD Content Docs

<critical>
Historically, course corrections updated only the index docs (epic-list, epics-and-stories, ROADMAP) and skipped the content docs (requirements, product-scope, epic-details). This resulted in 47+ epics with no PRD coverage. ALL content docs MUST be updated.
</critical>

1. **`docs/prd/requirements.md`**: Add new functional requirements (FR) and non-functional requirements (NFR) for every new capability introduced by this correction. Each requirement needs an ID, description, and acceptance criteria.
2. **`docs/prd/product-scope.md`**: Add the new feature/scope to the appropriate phase section.
3. **`docs/prd/epic-details.md`**: Add a detailed breakdown for any new epic including: title, description, goals, key features, technical considerations, and story list.

## Step 6: Update PRD Index Docs

1. Add stories to the appropriate epic in `docs/prd/epics-and-stories.md`
2. Update `docs/prd/epic-list.md` with any new epic entries
3. Update `ROADMAP.md` if new work affects priorities
4. Add a decision entry to `docs/decisions/BOARD.md`

## Step 7: Validation Gate

Before reporting, verify:
- [ ] `requirements.md` has FR/NFR entries for every new capability
- [ ] `product-scope.md` has the feature in the correct phase section
- [ ] `epic-details.md` has a detailed breakdown for any new epic
- [ ] `epics-and-stories.md` has epic and story outlines
- [ ] `epic-list.md` has epic entry (if new epic)

If any are missing, go back to Step 5 and complete them.

## Step 8: Report

Output:
- Link to the sprint change proposal
- List of created stories with file paths
- Summary of PRD content doc updates (requirements, product-scope, epic-details)
- Summary of PRD index doc updates (epics-and-stories, epic-list, ROADMAP)
- Any decisions that need human approval before implementation begins
