# Open Source License Selection Research — ThreeDoors

**Date:** 2026-03-08
**Author:** bold-squirrel (worker agent)
**Status:** Complete
**Party Mode Artifact:** `_bmad-output/planning-artifacts/license-selection-party-mode.md`

---

## Executive Summary

ThreeDoors needs an open source license for Homebrew distribution (PR #262 defaulted to MIT without discussion). After evaluating 8 license options against Homebrew requirements, dependency compatibility, plugin ecosystem implications, contributor friendliness, and protection against commercial forks, **MIT is the recommended license**.

Key reasons: alignment with the Charm/Bubbletea ecosystem (all MIT), maximum freedom for third-party adapter development, zero maintenance overhead, and full Homebrew compatibility. The competitive moat for ThreeDoors is its philosophy and AI-agent methodology, not the source code — making copyleft protection unnecessary.

---

## 1. Homebrew License Requirements

### Custom Tap
- **No license restrictions.** Any license (or no license) works for a custom tap.
- Custom taps are self-hosted GitHub repositories — Homebrew imposes no review.

### homebrew-core
- **Must be DFSG-compatible** (Debian Free Software Guidelines).
- Accepted licenses include: MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, GPL-2.0, GPL-3.0, LGPL, MPL-2.0, ISC, Unlicense, CC0.
- **Not accepted:** BSL (Business Source License), SSPL, proprietary, "source available" licenses.
- The formula must declare `license "MIT"` (or equivalent) matching the repository's LICENSE file.
- `brew audit --strict --new --online` validates the license declaration.

### Licenses That Make Acceptance Easier
- **MIT, BSD, Apache-2.0:** Zero friction. Most Go projects in homebrew-core use these.
- **GPL:** Accepted but requires careful `license` declaration (GPL-2.0-only vs GPL-2.0-or-later).
- **AGPL:** Technically accepted but unusual for CLI tools; may draw scrutiny during review.
- **BSL/SSPL:** Rejected outright — not DFSG-compatible.

---

## 2. License Evaluation Matrix

| License | Homebrew | Anti-Fork | Adapters | Deps OK | Community | Relicense |
|---------|----------|-----------|----------|---------|-----------|-----------|
| MIT | Yes | No | Proprietary OK | Yes | Excellent | Easy |
| Apache 2.0 | Yes | No | Proprietary OK | Yes | Good | Moderate |
| GPL v3 | Yes | Yes | Must be GPL | Yes | Mixed | Hard |
| AGPL v3 | Yes* | Yes | Must be AGPL | Yes | Poor | Very Hard |
| BSD 2/3-Clause | Yes | No | Proprietary OK | Yes | Good | Easy |
| BSL | **No** | Yes (temp) | Restricted | Yes | Poor | N/A |
| SSPL | **No** | Yes | Restricted | Yes | Very Poor | N/A |
| MPL 2.0 | Yes | Partial | New files OK | Yes | Moderate | Moderate |

*AGPL is technically DFSG-compatible but unusual for CLI tools.

---

## 3. Detailed License Evaluations

### MIT License

**Overview:** Permissive. Do anything with the code, just keep the copyright notice.

**Pros:**
- Simplest widely-recognized license — one file, ~20 lines
- 100% aligned with Charm ecosystem (Bubbletea, Lipgloss, Bubbles all MIT)
- Zero maintenance overhead — no NOTICE file, no source headers
- Maximum adapter ecosystem freedom — third parties can write proprietary `TaskProvider` implementations
- Most contributor-friendly — lowest legal barrier
- Easy to relicense later (with contributor consent, or if single maintainer)

**Cons:**
- No patent protection — contributors don't grant explicit patent licenses
- No protection against proprietary forks
- No requirement to share modifications

**Homebrew:** Fully compatible with homebrew-core and custom taps.

**Best for:** Projects prioritizing adoption, ecosystem growth, and simplicity.

---

### Apache License 2.0

**Overview:** Permissive with explicit patent grant and patent retaliation clause.

**Pros:**
- Explicit patent grant protects users from contributor patent claims
- Patent retaliation clause — if someone sues over patents, their license terminates
- Well-understood by corporate legal teams
- Compatible with GPL v3 (but NOT GPL v2)

**Cons:**
- Requires NOTICE file for attribution
- Convention (not requirement) to add license headers to source files
- Longer, more complex license text (~175 lines vs MIT's ~20)
- Not aligned with Charm ecosystem (all MIT)
- Minor GPL v2 incompatibility

**Homebrew:** Fully compatible.

**Best for:** Projects with corporate contributors, patent-sensitive domains (e.g., cloud infrastructure, ML).

---

### GPL v3

**Overview:** Strong copyleft. All derivative works must be distributed under GPL v3.

**Pros:**
- Maximum protection against proprietary forks — any fork must remain GPL
- Ensures all modifications are shared with the community
- Strong community signal for "free software" values

**Cons:**
- **Kills the adapter ecosystem** — all `TaskProvider` implementations must be GPL
- Corporate contributors' legal teams often prohibit GPL contributions
- Users cannot integrate ThreeDoors into proprietary workflows without licensing concerns
- Contradicts SOUL.md's "meet users where they are" — corporate adapter authors blocked
- Harder to relicense later (need consent from all contributors)
- Data shows GPL CLI tools have fewer contributors than MIT equivalents

**Homebrew:** Compatible, but formula declaration must specify `GPL-3.0-only` or `GPL-3.0-or-later`.

**Best for:** Projects that prioritize software freedom ideology over adoption.

---

### AGPL v3

**Overview:** GPL v3 + network use triggers source disclosure.

**Pros:**
- Strongest copyleft — even hosting as a service requires source disclosure
- Prevents cloud providers from offering ThreeDoors-as-a-service without contributing back

**Cons:**
- ThreeDoors is a **local-first CLI tool** — the network clause is irrelevant
- Maximum contributor friction — corporate legal teams almost universally prohibit AGPL
- All cons of GPL v3, amplified
- Community perception: "hostile to business"

**Homebrew:** Technically DFSG-compatible but unusual; may draw scrutiny from Homebrew reviewers.

**Best for:** SaaS products concerned about cloud providers (e.g., MongoDB pre-SSPL).

---

### BSD 2-Clause / 3-Clause

**Overview:** Permissive. Very similar to MIT.

**Pros:**
- Simple and well-understood
- BSD 2-Clause is arguably the simplest license (shorter than MIT)
- BSD 3-Clause adds "no endorsement" clause

**Cons:**
- Functionally equivalent to MIT but less recognizable in the Go ecosystem
- Not aligned with Charm ecosystem (all MIT)
- The 3-Clause "no endorsement" provision adds text for negligible practical benefit
- Go community strongly favors MIT — BSD is more common in C/Unix heritage projects

**Homebrew:** Fully compatible.

**Best for:** Projects with BSD/Unix heritage or specific need for the no-endorsement clause.

---

### BSL (Business Source License)

**Overview:** Source-available with commercial restrictions. Converts to open source after a defined period (typically 3-4 years).

**Pros:**
- Prevents commercial competition during the restriction period
- Source code is visible (transparency)
- Eventually becomes fully open source

**Cons:**
- **NOT DFSG-compatible — blocks homebrew-core submission**
- **NOT recognized as "open source" by OSI**
- HashiCorp's BSL switch (2023) caused massive community backlash and OpenTofu fork
- Creates "is this really open source?" debates that damage community trust
- Complex legal text with nuanced "additional use grant" definitions
- No community expectation of contributions

**Homebrew:** NOT accepted in homebrew-core. Could work in a custom tap.

**Best for:** VC-backed infrastructure companies with real commercial competition (e.g., databases, observability).

---

### SSPL (Server Side Public License)

**Overview:** MongoDB's license. Extremely strong copyleft targeting cloud service providers.

**Pros:**
- Maximum protection against cloud providers offering the software as a service

**Cons:**
- **NOT DFSG-compatible — blocks homebrew-core**
- **NOT OSI-approved**
- Designed for infrastructure SaaS — completely wrong fit for a local-first CLI tool
- Even more contentious than BSL
- Linux distributions (Fedora, Debian) have rejected SSPL software

**Homebrew:** NOT accepted.

**Best for:** Database companies worried about AWS/Azure/GCP offering managed versions.

---

### MPL 2.0 (Mozilla Public License)

**Overview:** File-level copyleft. Modifications to existing files must be shared; new files can be any license.

**Pros:**
- "Weak copyleft" — modifications to core code must be shared, but new adapter files can be proprietary
- Compatible with GPL v2 and v3
- Used by HashiCorp (pre-BSL), Mozilla, Terraform providers

**Cons:**
- File-level distinction is confusing for contributors
- Creates legal ambiguity about what constitutes "modification" vs "new file"
- Less well-understood than MIT in the Go ecosystem
- Added complexity for marginal protection benefit
- Most meaningful forks modify core files anyway, making the protection theoretical

**Homebrew:** Fully compatible (DFSG-compliant).

**Best for:** Projects wanting to protect core while allowing proprietary extensions — but simpler alternatives (MIT + trademark) often achieve this better.

---

## 4. Dependency License Audit

All ThreeDoors dependencies were audited for license compatibility:

### Direct Dependencies (go.mod `require`)

| Package | License | Copyleft? |
|---------|---------|-----------|
| charmbracelet/bubbles | MIT | No |
| charmbracelet/bubbletea | MIT | No |
| charmbracelet/lipgloss | MIT | No |
| charmbracelet/x/* | MIT | No |
| fsnotify/fsnotify | BSD-3-Clause | No |
| google/go-github | BSD-3-Clause | No |
| google/uuid | BSD-3-Clause | No |
| mattn/go-isatty | MIT | No |
| muesli/termenv | MIT | No |
| spf13/cobra | Apache-2.0 | No |
| golang.org/x/* | BSD-3-Clause | No |
| gopkg.in/yaml.v3 | MIT | No |
| modernc.org/sqlite | BSD-3-Clause | No |

### Indirect Dependencies (notable)

| Package | License | Copyleft? |
|---------|---------|-----------|
| hashicorp/golang-lru/v2 | MPL-2.0 | File-level |
| spf13/pflag | BSD-3-Clause | No |
| inconshreveable/mousetrap | Apache-2.0 | No |
| google/pprof | Apache-2.0 | No |
| modernc.org/libc | BSD-3-Clause | No |

### Constraint Analysis

- **No GPL dependencies** — ThreeDoors is free to use any license.
- **One MPL-2.0 dependency** (hashicorp/golang-lru/v2) — file-level copyleft. Compatible with MIT, Apache-2.0, BSD, and GPL. Does NOT constrain the project's license choice. Any modifications to golang-lru source files must be shared, but this doesn't affect ThreeDoors' own code.
- **Apache-2.0 dependencies** (cobra, mousetrap, pprof) — compatible with MIT. The Apache patent grant flows to users of these libraries regardless of ThreeDoors' license.

**Conclusion:** No dependency constrains ThreeDoors' license choice. All options remain available.

---

## 5. Project-Specific Considerations

### Nature of the Project
- **Personal productivity tool** — not infrastructure, not a database, not a SaaS platform
- **Local-first, privacy-always** (SOUL.md) — no server component, no network service
- **Single maintainer** directing AI agents — relicensing is straightforward
- **Plugin architecture** — `TaskProvider` interface is the primary extension point

### The "Commercial Fork" Threat
- ThreeDoors' value proposition is its **philosophy** (three doors, not three hundred) and **AI-agent development methodology** — neither of which is protectable by a license
- A large company would build their own task management tool, not fork a niche TUI
- The Go code is the least defensible part of the project
- Even if someone forks: they can't fork the community, the SOUL.md philosophy, or the user trust
- **Conclusion:** The commercial fork threat doesn't justify copyleft's ecosystem costs

### Plugin/Adapter Ecosystem
- ThreeDoors already integrates with Apple Notes, Obsidian, Jira, GitHub Issues, Apple Reminders
- Future adapters (Todoist, Linear) are on the roadmap
- Corporate users may want to write internal adapters for proprietary systems
- **MIT allows this; GPL blocks it** — this is the decisive factor for adapter ecosystem growth

### Future License Changes
- With a single maintainer, relicensing from MIT to Apache 2.0 or even GPL is straightforward (no contributor consent needed for sole-authored code)
- If external contributors join, their MIT contributions can be relicensed only with consent
- A CLA (Contributor License Agreement) could be added later if relicensing flexibility is important
- **Starting with MIT keeps all options open**

---

## 6. Recommendation

### Primary: MIT License

MIT is the recommended license for ThreeDoors based on:

1. **Ecosystem alignment** — all Charm libraries are MIT
2. **Adapter ecosystem freedom** — proprietary adapters are welcome
3. **Simplicity** — one file, zero maintenance
4. **Homebrew compatibility** — full homebrew-core eligibility
5. **Contributor friendliness** — lowest barrier to entry
6. **Future flexibility** — easy to relicense if needed

### Implementation

1. Add `LICENSE` file with MIT text, copyright holder: arcaven / ThreeDoors contributors
2. Update Homebrew formula to declare `license "MIT"`
3. No source file headers needed
4. No NOTICE file needed
5. No CLA needed (consider adding one if contributor base grows)

### If the Owner Disagrees

If there's a strong desire to prevent commercial forks while keeping the adapter ecosystem open, **MPL 2.0** is the next-best option. It would require modifications to ThreeDoors core files to be shared while allowing proprietary adapters in separate files. However, the added legal complexity is likely not worth it for this project.

---

## References

- [OSI License List](https://opensource.org/licenses/)
- [DFSG Guidelines](https://www.debian.org/social_contract#guidelines)
- [Homebrew Acceptable Formulae — License Requirements](https://docs.brew.sh/Acceptable-Formulae)
- [Choose a License (GitHub)](https://choosealicense.com/)
- [Charm License (MIT)](https://github.com/charmbracelet/bubbletea/blob/main/LICENSE)
- [HashiCorp BSL FAQ](https://www.hashicorp.com/en/bsl)
- [SPDX License List](https://spdx.org/licenses/)
- Party Mode Artifact: `_bmad-output/planning-artifacts/license-selection-party-mode.md`
