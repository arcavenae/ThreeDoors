# Party Mode Artifact: CI/Security Hardening Triage

**Date:** 2026-03-08
**Issues:** #244, #245, #246, #248
**Participants:** Winston (Architect), John (PM), Bob (SM), Murat (Test Architect)

## Adopted Approach

**Single story (0.31) under Epic 0, Priority P1.**

### Rationale

- All four issues target the same CI workflow file (`ci.yml`)
- No code-level dependencies between them — but all serve the same goal: CI security hardening
- Combined diff is small (~50 lines of YAML changes)
- Four separate stories would create unnecessary ceremony overhead (4x CI runs, reviews, merges)
- Low effort, high value — security hygiene that should ship promptly

### Implementation Order

1. **#248 — Pin golangci-lint version** — trivial 2-line change, instant reproducibility win
2. **#246 — Pass secrets via env vars** — mechanical refactor, pattern already partially applied in workflow
3. **#245 — Replace softprops/action-gh-release with gh CLI** — slightly more involved; must match exact behavior (asset uploads, release notes, pre-release flag)
4. **#244 — Move signing secrets to protected environment** — requires manual GitHub Settings changes as prerequisite; workflow update references `environment: release`

### Key Considerations

- **#244 prerequisite:** Repo owner must create the "release" environment in GitHub Settings and move secrets before the workflow change will work. Story should note this as a manual step.
- **Verification plan:** No traditional unit tests apply. Verification = CI passes + ideally a test release is attempted.
- **#245 risk:** The `gh release create` replacement must exactly match `softprops/action-gh-release` behavior for asset uploads, notes format, and pre-release flag.

## Rejected Options

### Option: Four Separate Stories (one per issue)

**Rejected because:** Creates unnecessary overhead. Four rounds of CI, reviews, merges for what amounts to small, related YAML changes in the same file. The issues share a common goal and don't conflict — combining them is cleaner.

### Option: Defer #244 (protected environment) to a separate story

**Considered because:** #244 requires manual GitHub Settings work that can't be automated via PR, unlike the other three which are pure YAML changes.

**Rejected because:** The YAML change for #244 is small and fits naturally with the other hardening changes. The manual prerequisite can be documented in the story's acceptance criteria. Splitting it out would create an unnecessary second story.

### Option: Priority P2 (nice to have)

**Rejected because:** These are established security best practices with low implementation effort. The signing secrets exposure (#244) and supply chain risk (#245) represent real — if low-probability — threats. P1 is appropriate given the effort/value ratio.
