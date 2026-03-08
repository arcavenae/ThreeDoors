# Party Mode Artifact: Open Source License Selection

**Date:** 2026-03-08
**Participants:** Winston (Architect), John (PM), Mary (Analyst), Amelia (Dev), Quinn (QA), Bob (SM)
**Topic:** Choosing an open source license for ThreeDoors
**Trigger:** PR #262 defaulted to MIT without discussion

---

## Adopted Approach: MIT License

**Rationale:**

1. **Charm ecosystem alignment** — Bubbletea, Lipgloss, Bubbles, and all Charm libraries are MIT. Using the same license eliminates cognitive friction for contributors familiar with the ecosystem.

2. **Maximum adapter/plugin ecosystem freedom** — ThreeDoors' `TaskProvider` interface is the primary extension point. MIT allows third parties to write proprietary adapters (e.g., corporate Jira integrations, internal tools) without license concerns. This aligns with SOUL.md's "meet users where they are" principle.

3. **Zero maintenance overhead** — Single LICENSE file, no NOTICE file, no source headers needed. Ideal for a single-maintainer project directing AI agents.

4. **Homebrew fully compatible** — Accepted in both custom taps and homebrew-core without restrictions.

5. **Dependency compatible** — All dependencies are MIT, BSD-3-Clause, Apache-2.0, or MPL-2.0. No copyleft constraints. MIT is compatible with all of them.

6. **Contributor friendly** — Simplest widely-recognized license. No CLA needed. Lowest barrier to contribution.

7. **Competitive moat is not the code** — ThreeDoors' value is in its philosophy (three doors, not three hundred), the AI-agent development methodology, and community — not the Go source code. A large company forking the code gains nothing without the design philosophy.

---

## Rejected Alternatives

### Apache 2.0
- **Why considered:** Explicit patent grant and patent retaliation clause provide stronger legal protection for contributors and users.
- **Why rejected:** Added complexity (NOTICE file, longer license text) provides marginal benefit for a personal productivity tool with a single maintainer. The patent threat model doesn't apply — ThreeDoors doesn't implement novel algorithms or patentable methods. Minor GPLv2 incompatibility is an unnecessary complication. Breaks ecosystem alignment with Charm (all MIT).

### GPL v3
- **Why considered:** Forces all derivative works (including forks) to remain open source. Maximum protection against proprietary forks.
- **Why rejected:** Kills the adapter ecosystem. Third parties cannot write proprietary `TaskProvider` implementations. This directly contradicts SOUL.md's "meet users where they are" — users with corporate Jira/Linear setups can't share internal adapters without open-sourcing them. Data shows GPL CLI tools (e.g., todo.txt CLI) have fewer contributors than MIT equivalents. Also creates friction with corporate contributors whose legal teams avoid GPL.

### AGPL v3
- **Why considered:** Strongest copyleft — even network use triggers source disclosure.
- **Why rejected:** ThreeDoors is a local-first CLI tool, not a SaaS product. AGPL's network clause provides zero benefit and maximum contributor friction. It would scare away potential contributors and corporate adopters for no practical gain. Massively misaligned with the project's nature.

### BSD 2-Clause / 3-Clause
- **Why considered:** Permissive, simple, well-understood.
- **Why rejected:** Functionally equivalent to MIT but less recognizable. Breaks ecosystem alignment with Charm. The 3-Clause variant's "no endorsement" clause adds unnecessary text for no practical benefit in this context. MIT is the de facto standard for Go TUI projects.

### BSL (Business Source License)
- **Why considered:** Prevents commercial use for a defined period, then converts to open source. Used by HashiCorp, MariaDB, Sentry.
- **Why rejected:** NOT DFSG-compatible. Would block homebrew-core submission entirely. Not recognized as "open source" by OSI. HashiCorp's BSL switch caused massive community backlash and fork (OpenTofu). For a project trying to build a community, this is toxic. Also: ThreeDoors has no commercial threat model that justifies this complexity.

### SSPL (Server Side Public License)
- **Why considered:** MongoDB's license targeting cloud providers who offer the software as a service.
- **Why rejected:** Same problems as BSL, plus more extreme. Not DFSG-compatible. Not OSI-approved. Designed for infrastructure SaaS — completely wrong fit for a local-first CLI tool. Would permanently block Homebrew acceptance and community trust.

### MPL 2.0 (Mozilla Public License)
- **Why considered:** File-level copyleft — modifications to existing files must be shared, but new files (including adapters) can be proprietary.
- **Why rejected:** Adds unnecessary complexity. The file-level copyleft distinction is confusing and creates legal ambiguity for contributors. Less well-understood than MIT in the Go ecosystem. The protection it offers (modifications to core files must be shared) provides marginal value — most forks would modify core files anyway, making the protection theoretical.

---

## Key Decision Factors

| Factor | Winner | Notes |
|--------|--------|-------|
| Homebrew compatibility | Tie (MIT/Apache/BSD/GPL) | All DFSG-compatible licenses accepted |
| Ecosystem alignment | MIT | Charm ecosystem is 100% MIT |
| Plugin/adapter freedom | MIT/Apache/BSD | GPL/AGPL block proprietary adapters |
| Simplicity | MIT | One file, no headers, no NOTICE |
| Patent protection | Apache 2.0 | But unnecessary for this project |
| Anti-fork protection | GPL/AGPL | But kills ecosystem and contributions |
| Contributor friendliness | MIT | Most recognized, least friction |
| Future license change | MIT | Easiest to relicense (with contributor consent) |

## Consensus

**Unanimous: MIT License.** All agents agreed that MIT best serves ThreeDoors' goals of community growth, adapter ecosystem development, Homebrew distribution, and alignment with the Charm TUI ecosystem. The threat model (large company forks) doesn't justify the ecosystem costs of copyleft, and the project's competitive moat (philosophy + methodology) is not protectable by any license.
