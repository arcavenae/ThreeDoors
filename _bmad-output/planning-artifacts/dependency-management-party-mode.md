# Party Mode: Renovate + Dependabot Dependency Management

**Date:** 2026-03-08
**Participants:** PM, Architect, Dev, QA, SM
**Topic:** Automated dependency management for ThreeDoors

---

## Discussion Summary

### Q1: What should Renovate handle vs Dependabot?

**Adopted Approach:** Renovate handles Go module dependencies (go.mod/go.sum) with security scanning. Dependabot handles GitHub Actions version pinning (.github/workflows/*.yml).

**Rationale:** Renovate has superior grouping, scheduling, and auto-merge policies for application dependencies. Dependabot has native GitHub Actions ecosystem support and is zero-config for action version updates. This avoids overlap — each tool owns a distinct ecosystem.

**Rejected Options:**
- *Renovate for everything:* Renovate can manage GitHub Actions, but Dependabot does it natively with less config and better GitHub integration (security advisories appear directly in the Security tab). Rejected because the marginal benefit doesn't justify the extra Renovate config complexity.
- *Dependabot for everything:* Dependabot's grouping and auto-merge capabilities are weaker. It creates one PR per dependency update (no grouping in go ecosystem), which would flood the merge queue. Rejected for merge-queue noise.
- *Neither (manual updates):* Not viable for a project with 20+ Go dependencies. Security patches would be delayed.

### Q2: Auto-merge policy

**Adopted Approach:**
- **Patch updates:** Auto-merge if CI green (for both Go deps and GitHub Actions)
- **Minor updates:** Auto-merge if CI green, but only for non-breaking Go dependencies (charmbracelet, standard library wrappers)
- **Major updates:** Never auto-merge. Create PR, label `breaking-change`, require human review.
- **Security updates:** Auto-merge patches and minors if CI green, regardless of other policies. Major security updates get priority label but still require human review.

**Rationale:** ThreeDoors has strong test coverage (75%+ enforced) and a comprehensive CI pipeline (unit tests, race detector, benchmarks, Docker E2E). This gives high confidence that auto-merged patch/minor updates won't introduce regressions.

**Rejected Options:**
- *Auto-merge everything including majors:* Too risky. Major version bumps in charmbracelet/bubbletea could change TUI behavior in ways tests don't catch (visual regressions). Rejected.
- *Never auto-merge:* Creates review burden. Patch updates (e.g., v1.0.1 → v1.0.2) are almost always safe and shouldn't require human attention. Rejected for being too conservative.
- *Auto-merge only security:* Leaves routine patches piling up. Rejected.

### Q3: Interaction with merge-queue agent

**Adopted Approach:**
- Renovate/Dependabot PRs get auto-labeled `dependencies` by their respective bots
- Security PRs additionally labeled `security` by Renovate's vulnerability detection
- Merge-queue agent treats `dependencies` PRs as in-scope (infrastructure/Epic 0 work)
- No special merge priority — dependency PRs go through normal CI and merge-queue flow
- Renovate schedule: weekdays only, batch window 6-8 AM UTC to avoid merge-queue contention during active development

**Rationale:** Dependency updates are routine infrastructure. They should flow through the same quality gates as all other PRs. Scheduling during off-hours prevents merge-queue congestion.

**Rejected Options:**
- *Priority lane for security PRs:* Adds complexity to merge-queue. Security patches auto-merge anyway, so they'll flow through quickly. Rejected unless we encounter actual delays.
- *Separate CI workflow for dependency PRs:* Unnecessary duplication. The existing CI pipeline is sufficient. Rejected.
- *Skip merge-queue for auto-merged PRs:* Violates branch protection. Rejected.

### Q4: Grouping strategy

**Adopted Approach:**
- Group all non-security patch updates into a single weekly PR (`chore(deps): patch updates`)
- Group charmbracelet ecosystem updates together (bubbles, bubbletea, lipgloss, etc.)
- Security updates get individual PRs for clear audit trail
- GitHub Actions updates grouped monthly

**Rationale:** Grouping reduces PR noise. The charmbracelet ecosystem is tightly coupled and should be updated together. Security updates need individual PRs for compliance visibility.

**Rejected Options:**
- *No grouping (one PR per dep):* Too noisy. With 20+ deps, this would create 5-10 PRs per week. Rejected.
- *Group everything including security:* Obscures security patches in routine updates. Rejected for auditability.
- *Daily batches:* Too frequent. Weekly is sufficient for a project of this size. Rejected.

### Q5: Breaking changes from major version bumps

**Adopted Approach:**
- Major version PRs are never auto-merged
- PR description includes changelog link and breaking change summary (Renovate does this automatically)
- Label `breaking-change` applied automatically
- Developer reviews changelog, updates code if needed, then merges manually
- If multiple major updates accumulate, they can be tackled as a dedicated story

**Rationale:** Major bumps require human judgment. Renovate's release notes extraction gives reviewers the information they need.

**Rejected Options:**
- *Automated migration scripts:* Over-engineered for a project with this dependency count. Rejected.
- *Ignore major updates:* Falling behind on majors creates security risk and tech debt. Rejected.

### Q6: Envoy agent notification for security PRs

**Adopted Approach:** Not implemented in Phase 1. Renovate's `security` label on PRs is sufficient visibility. If security PRs start accumulating without review, add a GitHub Actions workflow to ping the relevant team.

**Rationale:** The envoy agent concept is not yet implemented in the multiclaude setup. Adding notification infrastructure before the agent exists is premature.

**Rejected Options:**
- *Build custom notification system:* YAGNI. Labels and PR descriptions provide enough visibility. Rejected.
- *Slack/email notifications:* No Slack/email integration exists. Rejected as out of scope.

### Q7: Renovate security features to enable

**Adopted Approach:**
- **Vulnerability alerts:** Enabled (Renovate's built-in vulnerability detection via OSV database)
- **Security updates priority:** Enabled (security PRs created immediately, not batched)
- **Release notes extraction:** Enabled (changelogs included in PR descriptions)
- **SLSA verification:** Deferred to Phase 2 or later (requires additional infrastructure)

**Rationale:** OSV-based vulnerability detection is zero-config and high-value. SLSA verification adds complexity and is better suited for a future hardening story.

**Rejected Options:**
- *Enable everything at once including SLSA:* Over-scoped for initial setup. Rejected.
- *Skip vulnerability alerts (rely on GitHub's native alerts):* Renovate's alerts are more actionable because they come with an update PR. Rejected.

### Q8: Impact on Homebrew formula maintenance

**Adopted Approach:** No direct impact. Homebrew formula updates are triggered by the release workflow, not by dependency changes. Dependency updates flow through normal CI → merge → release pipeline. The formula template in the repo is not affected by go.mod changes.

**Rationale:** The Homebrew formula references binary artifacts, not Go module dependencies. The only connection is that dependency updates might trigger a new release (if merged to main), which would then update the formula via the existing `update-homebrew` CI job.

**Rejected Options:**
- *Special handling for Homebrew-affecting deps:* No dependency directly affects the formula. Rejected as unnecessary.

---

## Consensus Decisions

1. **Renovate = Go deps, Dependabot = GitHub Actions** — clean separation, no overlap
2. **Auto-merge patches + non-breaking minors** — leveraging strong CI coverage
3. **Weekly grouped PRs** with security as individual PRs
4. **Standard merge-queue flow** — no special lanes or priority
5. **Weekday morning schedule** — avoid developer contention
6. **OSV vulnerability scanning** enabled from day one
7. **SLSA verification** deferred to future story
