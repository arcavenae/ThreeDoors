# Renovate + Dependabot Agent Interaction Analysis

**Date:** 2026-03-09
**Author:** wise-squirrel (research worker)
**Related:** Story 0.24, `dependency-management-party-mode.md`

---

## Executive Summary

Enabling Renovate and Dependabot will create **five distinct interaction points** with existing multiclaude agents. Two are **HIGH risk** (merge-queue scope rejection, arch-watchdog noise), two are **MEDIUM risk** (envoy over-triage, project-watchdog correlation failure), and one is **LOW risk** (pr-shepherd conflicts). Without mitigation, the worst case is merge-queue rejecting all dependency PRs as "out of scope" and envoy spinning up full BMAD triage pipelines for every CVE advisory.

---

## Agent-by-Agent Analysis

### A. MERGE-QUEUE vs RENOVATE/DEPENDABOT — HIGH RISK

**Current behavior:** merge-queue checks `gh pr list --label multiclaude` (line 10 of definition). It only processes PRs with the `multiclaude` label. Its checklist includes "Aligns with ROADMAP.md? (no out-of-scope features)" (line 20).

**Conflict #1 — Invisible PRs:** Renovate/Dependabot PRs will NOT have the `multiclaude` label. merge-queue literally filters by `--label multiclaude`, so it will **never see** dependency PRs at all. This means:
- Auto-merged PRs (patch/minor) bypass merge-queue entirely — **this is fine**, Renovate handles its own merge via GitHub's auto-merge.
- Major version PRs (non-auto-merge) sit without merge-queue processing — **this is a gap**. No agent picks them up. They require human review anyway per the party mode decision, but nobody surfaces them.

**Conflict #2 — Scope rejection if label added:** If someone adds `multiclaude` label to a dependency PR, merge-queue would scope-check it against ROADMAP.md. Dependency updates aren't roadmap items. merge-queue would flag them as "out-of-scope" and add `needs-human-input` label. This creates noise.

**Conflict #3 — Auto-merge + branch protection:** Renovate's auto-merge requires that branch protection rules allow it. If Renovate's GitHub App has merge permissions and CI passes, auto-merge works independently of merge-queue. No conflict here — they're parallel paths.

**Assessment:** The `--label multiclaude` filter is actually a natural firewall. merge-queue won't interact with dependency PRs at all. The gap is that major version PRs have no agent shepherding them.

**Mitigations needed:**
1. Story 0.24 AC19 ("merge-queue recognizes `dependencies` label") needs re-evaluation — the current architecture naturally separates these concerns via label filtering
2. For major version PRs: either (a) add a standing order for merge-queue to periodically check `--label breaking-change` PRs and notify supervisor, or (b) rely on humans checking GitHub's PR dashboard
3. Recommended: Option (a) — add a secondary scan to merge-queue for `--label dependencies --label breaking-change`

### B. PR-SHEPHERD vs RENOVATE — LOW RISK

**Current behavior:** pr-shepherd is designed for fork workflows. It queries `gh pr list --repo UPSTREAM/REPO --author @me`. Per MEMORY.md, the project switched from fork to direct push on 2026-03-07.

**Conflict analysis:** Even if pr-shepherd is still active:
- It filters by `--author @me` — Renovate PRs are authored by `renovate[bot]`, Dependabot by `dependabot[bot]`
- pr-shepherd would never see their PRs
- Its rebase behavior (only on conflict, not proactive per ADR-0030) wouldn't interfere even if it did see them

**Assessment:** No conflict. The author filter is a complete firewall.

**Mitigations needed:** None.

### C. ENVOY vs DEPENDABOT SECURITY ALERTS — MEDIUM-HIGH RISK

**Current behavior:** envoy watches `gh issue list --state open` and triages every new issue through the full BMAD pipeline: PM examination → party mode → PRD/arch updates → story creation → docs PR.

**Conflict #1 — Dependabot security alerts as issues:** GitHub's Dependabot security alerts do NOT appear as regular issues — they appear in the repository's Security tab (Security → Dependabot alerts). They are separate from the Issues tab. **envoy would NOT see them** via `gh issue list`.

**Conflict #2 — Dependabot security PRs as issues:** However, Dependabot CAN be configured to open security update PRs. If a user reports a vulnerability as a regular GitHub issue (e.g., "package X has CVE-YYYY-ZZZZ"), envoy WOULD triage it through the full pipeline. This is the intended behavior for user-reported issues.

**Conflict #3 — Renovate vulnerability PRs:** Renovate creates PRs (not issues) for vulnerabilities. envoy only watches issues, not PRs. No conflict.

**Conflict #4 — Future risk:** If the project enables GitHub's "Dependabot alerts" to create issues (a repository setting), every CVE would become an issue that envoy triages. With Go's dependency tree, this could mean 5-15 security issues per quarter, each triggering full party mode.

**Assessment:** Currently low risk because security alerts don't appear as regular issues. Becomes HIGH risk if the repository setting "Dependabot alerts → Create issues" is enabled. The full BMAD triage pipeline (PM → party mode → PRD → story → docs PR) for a "bump X from 1.2.3 to 1.2.4" fix would be massive overkill.

**Mitigations needed:**
1. Do NOT enable the "Create issues from Dependabot alerts" repository setting
2. Add a label-based fast-path to envoy's definition: issues labeled `dependencies` or `security` by bots get an abbreviated triage (acknowledge → verify fix PR exists → close) instead of full BMAD pipeline
3. Document this in envoy's agent definition as a "bot-generated issue handling" section

### D. ARCH-WATCHDOG vs DEPENDENCY UPDATES — HIGH RISK

**Current behavior:** arch-watchdog processes all merged PRs that change files in `internal/`, `cmd/`, or `go.mod`/`go.sum` (lines 46-47 of definition). For each, it reviews the diff for "new external dependencies added" and "dependency version compatible with Go module requirements."

**Conflict #1 — Every dependency PR triggers review:** Every merged Renovate PR changes `go.mod` and `go.sum`. arch-watchdog would process each one, checking for:
- "New external dependencies added?" — yes, for new transitive deps pulled in
- "Dependency version compatible with Go module requirements?" — yes, always
- "New packages or interfaces introduced?" — no, but it still runs the full analysis

**Conflict #2 — Volume:** With weekly grouped patch PRs + monthly action PRs + individual security PRs, arch-watchdog would process 4-8 dependency PRs per month. Each triggers a full architectural review of the diff. For a `go.sum` diff that's 200+ lines of hash changes, this is pure noise.

**Conflict #3 — False positives:** arch-watchdog might flag transitive dependency additions as "new external dependencies" needing justification, when they're just pulled in by a patch bump.

**Assessment:** HIGH noise risk. arch-watchdog will faithfully analyze every dependency PR diff and potentially message supervisor/project-watchdog about non-issues.

**Mitigations needed:**
1. Add a filter to arch-watchdog's definition: PRs labeled `dependencies` that ONLY change `go.mod`, `go.sum`, and `.github/` files should be SKIPPED (no architecture review needed)
2. The filter should check: `gh pr diff <number> --name-only` — if all changed files are in the skip list AND the PR has `dependencies` label, add to processed list without review
3. Exception: PRs that change `go.mod` AND `internal/` or `cmd/` files should still get full review (these are PRs that update deps AND modify code to accommodate the update)

### E. PROJECT-WATCHDOG vs DEPENDENCY PRs — MEDIUM RISK

**Current behavior:** project-watchdog processes merged PRs, identifies which story they relate to (from branch name, PR title, or commit messages), and updates story status + ROADMAP.md.

**Conflict #1 — Story correlation failure:** Renovate PRs have titles like `chore(deps): update module github.com/foo/bar to v1.2.3` and branches like `renovate/github.com-foo-bar-1.x`. project-watchdog would try to correlate these to a story and fail — no story exists for individual dependency updates.

**Conflict #2 — Noise messages:** When project-watchdog can't find a story, it would likely message supervisor asking about uncorrelated PRs. With 4-8 dependency PRs per month, that's 4-8 spurious messages.

**Conflict #3 — ROADMAP.md drift detection:** project-watchdog checks "does the merged work reveal PRD gaps?" Dependency updates don't reveal PRD gaps, but project-watchdog might flag them as "merged work not reflected in planning docs."

**Assessment:** MEDIUM risk. project-watchdog won't break anything, but it will generate noise trying to correlate dependency PRs to stories and potentially messaging supervisor about non-issues.

**Mitigations needed:**
1. Add a filter to project-watchdog's definition: PRs labeled `dependencies` should be SKIPPED (no story correlation needed)
2. The filter should be: if PR has `dependencies` label → add to processed list immediately, no further processing
3. Exception: PRs labeled `dependencies` AND `breaking-change` should still be flagged to supervisor (major version bumps may need a dedicated story)

---

## Risk Matrix

| Scenario | Likelihood | Severity | Impact Description | Mitigation | Agent Changes Needed |
|----------|-----------|----------|-------------------|------------|---------------------|
| merge-queue ignores dependency PRs entirely | Certain | Low | By design — auto-merge handles patch/minor. Gap for major PRs only | Add secondary scan for `--label breaking-change` PRs | Minor addition to merge-queue.md |
| merge-queue rejects dependency PR as out-of-scope | Low | Medium | Only if someone manually adds `multiclaude` label | Document: never add `multiclaude` label to bot PRs | Standing order, not agent change |
| pr-shepherd fights Renovate over rebases | None | N/A | `--author @me` filter prevents any interaction | None needed | None |
| envoy full-triages a Dependabot security issue | Low (currently) | High | Full BMAD pipeline for a patch bump = hours of wasted compute | Don't enable "create issues from alerts"; add bot-issue fast-path to envoy | Add "bot-generated issue" section to envoy.md |
| arch-watchdog reviews every dependency PR diff | Certain | Medium | 4-8 unnecessary architectural reviews/month, false positive messages | Add `dependencies` label filter to skip pure-dep PRs | Add filter logic to arch-watchdog.md |
| arch-watchdog flags transitive deps as undocumented | Likely | Low | Noise messages to supervisor about auto-added transitive deps | Same filter as above | Same as above |
| project-watchdog can't correlate dep PRs to stories | Certain | Medium | 4-8 spurious "uncorrelated PR" messages to supervisor/month | Add `dependencies` label filter to skip dep PRs | Add filter logic to project-watchdog.md |
| project-watchdog flags dep PRs as ROADMAP drift | Likely | Low | False "merged work not in planning docs" alerts | Same filter as above | Same as above |
| Renovate auto-merge + branch protection conflict | Unlikely | High | Auto-merge fails if Renovate app lacks permissions | Verify Renovate app permissions during setup | None (config, not agent) |
| 20+ dep PRs in one week overwhelm merge-queue | Very Low | Medium | Renovate grouping prevents this; only possible during security event | Renovate config already groups; no additional mitigation | None |

---

## Required Agent Definition Changes

### 1. arch-watchdog.md — Add dependency PR filter

**Location:** After the "Filtering for Code PRs" section (after line 53)

**Add:**
```markdown
### Dependency PR Filter

PRs labeled `dependencies` that ONLY change `go.mod`, `go.sum`, and/or `.github/` files
should be added to the processed list WITHOUT architecture review. These are automated
dependency updates from Renovate/Dependabot and do not introduce architectural changes.

To check:
```bash
# Get PR labels
gh pr view <number> --json labels --jq '.labels[].name'
# If "dependencies" label present, check files
gh pr diff <number> --name-only
# If ALL files match: go.mod, go.sum, .github/* → skip
# If ANY file is in internal/ or cmd/ → full review (code was modified for the update)
```
```

### 2. project-watchdog.md — Add dependency PR filter

**Location:** In the "On Merged PR Detected" section, after step 1 (correlation ID check)

**Add:**
```markdown
1.5. **Check for dependency PRs** — if the PR has a `dependencies` label (check with
`gh pr view <number> --json labels`), skip story correlation and planning doc updates.
Dependency updates from Renovate/Dependabot are infrastructure automation and don't
correspond to stories. Exception: if the PR also has `breaking-change` label, message
supervisor about the major version update.
```

### 3. envoy.md — Add bot-generated issue handling

**Location:** After the "Triage Pipeline" section

**Add:**
```markdown
## Bot-Generated Issue Handling

Issues opened by bots (`dependabot[bot]`, `renovate[bot]`, `github-actions[bot]`) should
follow an abbreviated triage path instead of the full BMAD pipeline:

1. **Identify as bot issue** — check issue author
2. **Acknowledge** — post a brief comment noting automated triage
3. **Check for existing fix PR** — search for open PRs with `dependencies` label
   addressing the same dependency
4. **If fix PR exists:** comment linking to it, close the issue
5. **If no fix PR:** message supervisor for guidance
6. **Do NOT run** PM examination or party mode for bot-generated dependency issues

This prevents the full BMAD pipeline from activating for routine dependency updates
that already have automated fix PRs.
```

### 4. merge-queue.md — Add breaking-change scan (optional)

**Location:** After the main loop description

**Add:**
```markdown
## Dependency Major Version PRs

Periodically check for major version dependency PRs that need human review:

```bash
gh pr list --label "breaking-change" --label "dependencies"
```

If found, notify supervisor:
```bash
multiclaude message send supervisor "Major dependency update PR #<number> needs human review: [title]"
```

Do NOT attempt to merge these — they require human review of breaking changes.
```

---

## Renovate/Dependabot Configuration Recommendations

The following config choices (already in Story 0.24) naturally reduce agent interaction noise:

1. **`enabledManagers: ["gomod"]`** — Renovate only touches Go deps, not GitHub Actions (Dependabot handles those separately)
2. **`labels: ["dependencies"]`** — All bot PRs get this label, enabling agent filtering
3. **`schedule: ["before 8am on weekday"]`** — Batch window avoids active development hours
4. **Grouping rules** — Weekly batch prevents PR flood (1-2 PRs/week vs 5-10/week ungrouped)
5. **Auto-merge for patch/minor** — Reduces PRs that need agent attention

### Additional config recommendation:

Add to `renovate.json`:
```json
{
  "prBodyNotes": ["This PR was created by Renovate automation. multiclaude agents: no story correlation needed."]
}
```

This provides a human-readable signal in PR descriptions that this is bot-generated.

---

## Standing Orders Needed

1. **Never add `multiclaude` label to bot-generated PRs** — this would cause merge-queue to scope-check them against ROADMAP.md and reject them
2. **Never enable "Create issues from Dependabot alerts"** repository setting — this would cause envoy to full-triage every CVE
3. **Major version dependency PRs** are the responsibility of the human maintainer, not the agent ecosystem — agents should notify but not attempt to process them

---

## Conclusion

The existing agent ecosystem is **surprisingly well-insulated** from dependency bot interference, primarily due to:
- merge-queue's `--label multiclaude` filter (natural firewall)
- pr-shepherd's `--author @me` filter (complete isolation)
- envoy watching issues, not PRs (Renovate/Dependabot primarily create PRs)

The main risks are from the **monitoring agents** (arch-watchdog and project-watchdog) that process ALL merged PRs regardless of source. These need label-based filters to skip `dependencies`-labeled PRs.

With the four agent definition changes above and three standing orders, the dependency bot ecosystem can coexist safely with the multiclaude agent ecosystem.
