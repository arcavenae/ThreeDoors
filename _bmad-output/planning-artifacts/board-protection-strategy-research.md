# BOARD.md Protection Strategy Research

**Date:** 2026-03-29
**Context:** Story 74.1 added CODEOWNERS protection for `docs/decisions/BOARD.md`. This research evaluates whether that creates a governance bottleneck and recommends an alternative.
**Companion to:** [DFCP Permissions Research](dfcp-permissions-research.md) (R-005)

---

## Problem Statement

BOARD.md is updated by most agents across all three development phases (research, plan, implement). On 2026-03-29 alone, 12 PRs touched BOARD.md — ALL would have required human approval under CODEOWNERS protection. This effectively blocks the dark factory's autonomous operation behind a single human reviewer.

## Current State Analysis

### BOARD.md by the Numbers

| Section | Entry Count | Risk Profile | Typical Authors | Update Frequency |
|---------|-------------|-------------|-----------------|------------------|
| **Decided (D-xxx)** | 202 | HIGH — architectural commitments | Party mode, research spikes, owner | Multiple per day during active work |
| **Open Questions (Q-xxx)** | 20 | ZERO — just questions | Any agent, research workers | Sporadic |
| **Active Research (R-xxx)** | 14 | LOW — tracking, not commitments | Research workers | Sporadic |
| **Pending Recommendations (P-xxx)** | 16 | LOW — proposals awaiting review | Party mode, research workers | After research completes |

**Key insight:** Only Decided entries (D-xxx) represent binding architectural commitments. The other 3 sections are operational tracking — they record status, not policy.

### Who Modifies What

| Agent Type | Sections Modified | Sensitivity |
|-----------|------------------|-------------|
| Research workers | R-xxx (add/update), Q-xxx (add), D-xxx (add from research) | R/Q: low. D: high. |
| Party mode | P-xxx (add), D-xxx (add from consensus) | P: low. D: high. |
| Implementation workers | D-xxx (rarely, mark as implemented) | D: medium |
| project-watchdog | Status updates across all sections | Low |
| Owner/Supervisor | All sections | Intentional |

## Options Evaluated

### Option A: Split into Two Files

**BOARD.md** (governance, CODEOWNERS-protected) — Decided entries only
**BOARD-OPS.md** (operational, ungated) — Open Questions, Active Research, Pending Recommendations

| Pros | Cons |
|------|------|
| Clean separation of concern | Breaks all existing links to BOARD.md sections |
| Decided entries get proper protection | Two files to maintain, cross-reference |
| Agents freely update operational tracking | Agents must know which file to use |
| Simple CODEOWNERS rule | Migration effort for 50+ existing entries |

**Verdict:** Good separation but high migration cost and link breakage.

### Option B: Split into Three+ Files (ADR-style)

**DECISIONS.md** — Protected, D-xxx only
**RESEARCH-LOG.md** — Ungated, R-xxx entries
**QUESTIONS.md** — Ungated, Q-xxx entries
**RECOMMENDATIONS.md** — Ungated, P-xxx entries

| Pros | Cons |
|------|------|
| Maximum granularity | Four files instead of one |
| Each file has clear ownership | Loses the "single dashboard" value |
| Mirrors ADR directory pattern | Over-engineered for current scale |

**Verdict:** Over-engineered. The value of BOARD.md is its single-dashboard nature.

### Option C: Remove BOARD.md from CODEOWNERS (Status Quo Ante)

Keep BOARD.md as a single ungated file. Rely on behavioral enforcement (CLAUDE.md instructions, agent definitions) to prevent unauthorized decision modifications.

| Pros | Cons |
|------|------|
| Zero migration effort | No technical enforcement for decisions |
| Agents remain productive | Behavioral gates are leaky (per D-DFCP-5) |
| Single dashboard preserved | "Enforce via tooling, not prompts" violated (R-010 lesson) |

**Verdict:** Too permissive. Decisions are the one section that warrants protection.

### Option D: CI-Based Section Protection (Recommended)

Keep BOARD.md as a single file. **Remove it from CODEOWNERS.** Instead, add a CI check that:

1. Detects when the `## Decided` section is modified (specifically, when existing D-xxx rows are changed or deleted)
2. **Adding** new D-xxx entries is ALLOWED (agents need to record new decisions)
3. **Modifying or deleting** existing D-xxx entries triggers a warning label (`status.needs-human`) but does NOT block merge
4. Changes to Q-xxx, R-xxx, P-xxx sections are always ungated

This is a "soft gate" — it flags sensitive changes for human attention without blocking the pipeline.

| Pros | Cons |
|------|------|
| Single file preserved | Requires CI workflow implementation |
| Adding decisions is ungated (productive) | Soft gate can be ignored (but audit trail exists) |
| Modifying decisions is flagged (safe) | More complex than CODEOWNERS |
| No migration needed | CI check needs maintenance |
| Aligns with "enforce via tooling" (R-010) | |
| Preserves dashboard value | |

**Verdict:** Best balance of safety and productivity.

### Option E: Two-File Split with CI Augmentation (Runner-up)

**BOARD.md** — Single dashboard (ungated), but with CI soft-gate on D-xxx modifications
**DECISIONS-PROTECTED.md** — Subset: only "strategic" decisions (D-080+, dark factory era) in CODEOWNERS

| Pros | Cons |
|------|------|
| Protects the most sensitive recent decisions | Two files partially overlap |
| Legacy decisions (D-001 to D-079) ungated (stable, rarely modified) | Confusing which file to update |

**Verdict:** Adds complexity without proportional benefit over Option D.

## Recommendation: Option D — CI-Based Section Protection

### How It Works

1. **Remove `docs/decisions/BOARD.md` from `.github/CODEOWNERS`**
2. **Add a CI check** (GitHub Action) that runs on PRs modifying `docs/decisions/BOARD.md`:
   - Parse the diff for changes to lines matching `| D-\d+ |`
   - If only NEW D-xxx lines are added → PASS (green check)
   - If existing D-xxx lines are modified or deleted → WARN (add `status.needs-human` label)
   - Changes to Q-xxx, R-xxx, P-xxx sections → always PASS
3. **merge-queue respects `status.needs-human`** — already part of its protocol (Story 74.1)

### Why This Is the Right Balance

| Concern | How Addressed |
|---------|--------------|
| Dangerous decision modifications | Flagged by CI, labeled for human review |
| Agent productivity | Adding new entries is ungated |
| Operational tracking | Q/R/P sections fully ungated |
| Audit trail | Git blame + CI check history |
| Dashboard value | Single file preserved |
| "Enforce via tooling" principle | CI check is mechanical, not behavioral |

### Implementation Sketch

```yaml
# .github/workflows/board-protection.yml
name: BOARD.md Decision Protection
on:
  pull_request:
    paths: ['docs/decisions/BOARD.md']

jobs:
  check-decisions:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Check for decision modifications
        run: |
          # Get diff of BOARD.md in the Decided section
          DIFF=$(git diff origin/main...HEAD -- docs/decisions/BOARD.md)

          # Check for MODIFIED (not added) D-xxx lines
          # Lines starting with - (removed) that contain D-xxx pattern
          MODIFIED=$(echo "$DIFF" | grep -E '^-\| D-[0-9]+ \|' || true)

          if [ -n "$MODIFIED" ]; then
            echo "::warning::Existing decision entries were modified or removed. Flagging for human review."
            echo "NEEDS_HUMAN=true" >> $GITHUB_ENV
          fi
      - name: Label PR if needed
        if: env.NEEDS_HUMAN == 'true'
        run: |
          gh pr edit ${{ github.event.pull_request.number }} \
            --add-label "status.needs-human"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### What Remains Protected via CODEOWNERS

All other governance files from Story 74.1 stay in CODEOWNERS:
- SOUL.md, CLAUDE.md, .claude/ — agent behavior (HIGH risk)
- ROADMAP.md, epic-list.md, epics-and-stories.md — scope control (HIGH risk)
- .github/ — infrastructure (HIGH risk)
- agents/ — agent definitions (HIGH risk)

BOARD.md is the only file moved to CI-based protection because it's the only governance file that agents must update frequently as part of normal operations.

## Rejected Alternatives Summary

| Alternative | Why Rejected |
|------------|-------------|
| Keep BOARD.md in CODEOWNERS | Blocks ~90% of PRs; defeats dark factory purpose |
| Split into multiple files | Breaks single-dashboard value; high migration cost |
| Remove all protection | "Enforce via tooling, not prompts" violated; decisions are architectural commitments |
| Approval-by-label system | GitHub doesn't support conditional CODEOWNERS based on labels |
| Section-level CODEOWNERS | GitHub CODEOWNERS operates at file level only, not section level |
| ADR-directory-only approach | ADRs already exist for formal decisions; BOARD.md serves a different purpose (dashboard/tracking) |

## Open Questions

| # | Question | Recommendation |
|---|----------|----------------|
| OQ-BPS-1 | Should the CI check also flag modifications to P-xxx entries (pending recommendations)? | No — recommendations are proposals, not commitments. They're expected to be updated as they progress. |
| OQ-BPS-2 | Should the CI check hard-block (required status check) or soft-warn (label only)? | Start with soft-warn (label). Upgrade to hard-block after validation period if warranted. |
| OQ-BPS-3 | Should duplicate D-xxx numbers (e.g., D-080 appears twice) be flagged by CI? | Yes — this is a data quality issue. Current BOARD.md has several duplicates (D-059, D-074, D-080-D-086). |

## Decisions

| Decision | Adopted | Rejected | Rationale |
|----------|---------|----------|-----------|
| BPS-D-1: BOARD.md protection mechanism | CI-based section protection (Option D) | CODEOWNERS (bottleneck), file split (breaks dashboard), no protection (too permissive) | Only D-xxx modifications need gating; CI can distinguish adds from edits; preserves single-file dashboard |
| BPS-D-2: Gate behavior | Soft-warn with label, not hard-block | Hard-block (too restrictive for agents), no warning (invisible) | Soft-warn flags for human attention without blocking pipeline; merge-queue already respects `status.needs-human` |
| BPS-D-3: Scope of protection | Only existing D-xxx modifications/deletions; new D-xxx additions ungated | All D-xxx changes gated (blocks new decisions), all sections gated (same as CODEOWNERS) | Agents must be able to record new decisions freely; only retroactive changes to commitments warrant scrutiny |

---

## Sources

- Current BOARD.md: 460 lines, 252 entries across 4 sections
- DFCP Research (R-005): Gate taxonomy (Authority, Scope, Design, Quality)
- Story 74.1: CODEOWNERS implementation
- R-010 (chainlink lessons): "Enforce via tooling, not prompts"
- GitHub CODEOWNERS docs: File-level granularity only
- GitHub Actions docs: PR labeling and diff analysis
