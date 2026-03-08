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

## Step 5: Update Planning Docs

1. Add stories to the appropriate epic in `docs/prd/epics-and-stories.md`
2. Update `ROADMAP.md` if new work affects priorities
3. Add a decision entry to `docs/decisions/BOARD.md`

## Step 6: Report

Output:
- Link to the sprint change proposal
- List of created stories with file paths
- Summary of planning doc updates
- Any decisions that need human approval before implementation begins
