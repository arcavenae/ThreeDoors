---
name: 'implement-story'
description: 'Full story implementation workflow: prepare story, enrich with party mode, TDD cycle (tea → dev), simplify, review, and iterate until all ACs pass. Usage: /implement-story <story-identifier>'
---

# Full Story Implementation Workflow

This is the complete, end-to-end workflow for implementing a story from preparation through PR creation. It orchestrates multiple BMAD agents and Claude Code skills in the correct sequence.

**Input:** The user provides a story identifier (e.g., "4.5", "3.1") as $ARGUMENTS. If no argument is provided, ask for the story identifier before proceeding.

<critical>
- Execute ALL phases in order. Do NOT skip phases.
- Do NOT stop between phases unless a HALT condition is triggered.
- Each phase builds on the previous — context must carry forward.
- When invoking BMAD agents/skills, follow their instructions completely before moving to the next phase.
- The story identifier "$ARGUMENTS" should be used to locate the story. Convert dot notation (e.g., "4.5") to the file pattern used in the project (e.g., "4-5-*.md") when searching for story files in {project-root}/_bmad-output/implementation-artifacts/ or {project-root}/docs/stories/.
</critical>

---

## Phase 1: Story Preparation

**Goal:** Ensure the story file exists and is ready for development.

1. Search for the story file matching the identifier "$ARGUMENTS" in:
   - `{project-root}/_bmad-output/implementation-artifacts/` (primary, pattern: `{epic}-{story}-*.md`)
   - `{project-root}/docs/stories/` (fallback, pattern: `{epic}.{story}.story.md`)
2. If the story file exists and has status `ready-for-dev` or `in-progress`, note its path and proceed to Phase 2.
3. If the story file does NOT exist or needs preparation:
   - Launch `/sm` (Scrum Master agent) to create/prepare the story.
   - Provide the story identifier so SM knows which story to create.
   - Wait for SM to complete story creation.
   - Verify the story file now exists with proper structure (ACs, Tasks, Dev Notes).

---

## Phase 2: Story Enrichment — Dev Readiness

**Goal:** Ensure the story has everything the dev agent needs to succeed.

1. Launch `/party-mode` with the prompt:
   > "Review story $ARGUMENTS — what is missing from the story for dev to succeed? Consider: acceptance criteria completeness, task breakdown granularity, technical specifications, architecture guidance, edge cases, error handling requirements, and integration points."
2. Accept ALL recommendations from the party mode discussion.
3. Apply the accepted recommendations to the story file.
4. Save the updated story file.

---

## Phase 3: Story Enrichment — Test Readiness

**Goal:** Ensure the story has everything needed for test creation.

1. Launch `/party-mode` with the prompt:
   > "Review story $ARGUMENTS — what is missing from the story for tea (Test Engineering Architect) to create comprehensive tests so that dev can succeed with TDD? Consider: testable acceptance criteria, test data requirements, mock/stub specifications, boundary conditions, integration test scenarios, and expected behaviors for error cases."
2. Accept ALL recommendations from the party mode discussion.
3. Apply the accepted recommendations to the story file.
4. Save the updated story file.

---

## Phase 4: Red Phase — Create Failing Tests

**Goal:** Write comprehensive tests that define the expected behavior (all tests should FAIL since code doesn't exist yet).

1. Launch `/tea` (Test Engineering Architect agent) to create tests for the story.
2. Provide the story file path so TEA has full context.
3. TEA should create:
   - Unit tests for each acceptance criterion
   - Integration tests for component interactions
   - Edge case and error handling tests
4. Run the test suite to confirm tests FAIL (red phase validation).
5. If tests pass unexpectedly, investigate — the tests may not be testing new functionality correctly.

---

## Phase 5: Green Phase — Implement Code

**Goal:** Write the minimum code to make all tests pass.

1. Launch `/dev` (Developer agent) to implement the story.
2. Provide the story file path so DEV has full context.
3. DEV should follow the story's Tasks/Subtasks in order.
4. For each task:
   - Implement the minimum code to make related tests pass.
   - Run tests after each task to verify progress.
5. Continue until ALL tests pass (green phase validation).
6. Run the full test suite to verify no regressions.

---

## Phase 6: Refactor — Simplify and Polish

**Goal:** Review code for reuse opportunities, quality, and efficiency.

1. Run `/simplify` to review all changed code.
2. Address any issues found:
   - Code duplication → extract shared utilities
   - Quality issues → fix code smells
   - Efficiency problems → optimize hot paths
3. Re-run the full test suite after any refactoring to ensure tests still pass.
4. If tests fail after refactoring, fix immediately before proceeding.

---

## Phase 7: Acceptance Review

**Goal:** Verify ALL acceptance criteria and tasks are fully met.

1. Run `/bmad-bmm-code-review` (adversarial code review) against the implementation.
2. The review should verify:
   - Every acceptance criterion in the story is satisfied
   - Every task and subtask is complete
   - Tests adequately cover the acceptance criteria
   - No gaps between what the story requires and what was implemented
3. Any gap found is treated as a bug.

### If gaps are found:

1. Update the story file with corrections — add new tasks for each gap found.
2. Return to **Phase 4** (TEA → DEV → Simplify → Review cycle).
3. Do NOT create a PR until ALL acceptance criteria are met with zero gaps.

### If ALL acceptance criteria are met:

Proceed to Phase 8.

---

## Phase 8: Create Pull Request

**Goal:** Package the completed work into a PR.

1. Stage all changed files (code, tests, story file updates).
2. Create a descriptive commit with the story identifier and title.
3. Push the branch and create a PR with:
   - Title referencing the story (e.g., "feat: Story $ARGUMENTS — {story title}")
   - Body summarizing:
     - What was implemented
     - Tests added
     - Files changed
     - Acceptance criteria satisfied
4. Report the PR URL to the user.

---

## Iteration Protocol

If at any point during Phases 4-7 the cycle needs to repeat:

```
Phase 4 (TEA: write/update failing tests)
  → Phase 5 (DEV: make tests pass)
    → Phase 6 (Simplify: refactor)
      → Phase 7 (Review: verify ACs)
        → If gaps: back to Phase 4
        → If clean: Phase 8 (PR)
```

Maximum iterations: 3. If after 3 full cycles gaps remain, HALT and report the remaining issues to the user for guidance.
