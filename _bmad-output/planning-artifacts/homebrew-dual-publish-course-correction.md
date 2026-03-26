# Party Mode: Homebrew Dual Publish Course Correction

**Date:** 2026-03-09
**Participants:** PM (John), Architect (Winston), Dev (Amelia), Test Architect (Murat)
**Topic:** Course correction for Homebrew dual distribution — signing parity, toggle mechanism, verification, retention, and formula generation

---

## Context

ThreeDoors has two Homebrew distribution channels:
- **Stable:** GoReleaser on `v*` tags → pushes `Formula/threedoors.rb` to `arcavenae/homebrew-tap`. Runs on `ubuntu-latest`. **NOT signed.**
- **Alpha:** Every push to main → creates GitHub pre-release. PR #273 (open) adds `threedoors-a` alpha formula publishing. Alpha binaries ARE signed when `vars.SIGNING_ENABLED == 'true'` (runs on `macos-latest`).

**Ironic gap:** Alpha is MORE secure than stable (signed vs unsigned). This course correction addresses signing parity and five related infrastructure questions.

---

## Question 1: Signing for Stable Releases

### Adopted: Post-GoReleaser signing job on macOS runner (Option B)

GoReleaser builds on `ubuntu-latest` as today, uploads unsigned binaries. A subsequent `sign-stable` job runs on `macos-latest`, downloads the darwin binaries from the GoReleaser release, signs and notarizes them, and re-uploads to the same GitHub release. This mirrors the alpha pipeline's proven pattern (build on ubuntu, sign on macOS).

**Key details:**
- GoReleaser creates tar.gz archives, not bare binaries. The signing job must extract binaries, sign, re-archive, update checksums, and re-upload.
- Same Apple Developer ID certificate signs both channels. The `--identifier` flag differentiates (`com.arcavenae.threedoors` vs `com.arcavenae.threedoors-a`).
- Gated by `vars.SIGNING_ENABLED == 'true'` (same toggle as alpha signing).
- **Fail-open:** If signing fails, unsigned binaries remain (current baseline, not a regression).

### Rejected Options

| Option | Why Rejected |
|--------|-------------|
| (a) Move GoReleaser to macOS runner | Doesn't add signing by itself; GoReleaser doesn't do codesign; loses native Linux builds |
| (c) Replace GoReleaser with custom pipeline | Throws away GoReleaser's changelog, archive packaging, formula push, checksums; massive scope increase for marginal benefit |
| Separate signing identities per channel | Apple certs are per-team, not per-binary; bundle IDs provide differentiation; unnecessary complexity |

---

## Question 2: Alpha Publishing Toggle

### Adopted: `vars.ALPHA_TAP_ENABLED`, default OFF

Matches the existing `vars.SIGNING_ENABLED` pattern. Controls whether the alpha formula is pushed to the tap after release creation. The alpha GitHub release itself always happens.

**Naming rationale:** `ALPHA_TAP_ENABLED` is more precise than `ALPHA_PUBLISHING_ENABLED` — it specifically controls the tap push, not the release creation.

**Default OFF rationale:**
1. Infrastructure must be validated before activation
2. Conscious activation by the repo owner
3. Consistent with `SIGNING_ENABLED` which also defaults to off

**Implementation:** One line addition to PR #273's formula push step:
```yaml
if: vars.ALPHA_TAP_ENABLED == 'true'
```

### Rejected Options

| Option | Why Rejected |
|--------|-------------|
| `ALPHA_PUBLISHING_ENABLED` naming | Less precise; the release creation should always happen, only the tap push is gated |
| Default to ON | First push would attempt formula push before infrastructure is validated |

---

## Question 3: Two Formulas, Same Binary

### Adopted: Two separate formulas, no shared template

Reasons:
1. **GoReleaser generates `threedoors.rb` automatically** from its `brews:` config — we don't control the template
2. **Structurally different:** Stable uses tar.gz archives with simple `install` block; alpha downloads bare binaries with `on_arm`/`on_intel`/`on_linux` blocks
3. **SOUL.md principle:** "A little copying is better than a little dependency" — two 30-line Ruby files is fine

### Rejected Options

| Option | Why Rejected |
|--------|-------------|
| Shared formula template | GoReleaser controls stable formula; different structures make sharing impractical |

---

## Question 4: Alpha Release Verification

### Adopted: Lightweight tap CI monitoring (mirrors stable pattern)

Tiered approach:
1. **Syntax validation** — Already in PR #273 (`ruby -c threedoors-a.rb`). Catches structural errors before push.
2. **Tap CI monitoring** — Mirror `release-verify.yml` pattern, triggered by CI workflow completion. Look for alpha formula commit message pattern (`chore(formula): update threedoors-a`), wait for tap CI to pass. Shorter timeout than stable (alpha failures are less critical).
3. **Issue creation on failure** — Same pattern as stable: auto-create GitHub issue if tap CI fails.

### Rejected Options

| Option | Why Rejected |
|--------|-------------|
| Per-push `brew install` verification | Expensive (macOS runner + Homebrew install time); tap CI provides equivalent coverage |
| Weekly cron for deep verification | Over-engineering for now; tap CI monitoring is sufficient |

---

## Question 5: Alpha Release Retention

### Adopted: Keep last 30 alpha releases, delete older with `--cleanup-tag`

Rationale:
- 30 releases = roughly 2-6 days of active development history
- Formula always points to the latest release before cleanup runs
- `--cleanup-tag` also deletes associated git tags, preventing tag pollution
- Stable releases use `v*` tags, not `alpha-*` tags — cleanup only targets alpha

**Implementation:** Add cleanup step at the end of the `release` job:
```bash
gh release list --limit 200 \
  | grep 'alpha-' \
  | tail -n +31 \
  | awk '{print $1}' \
  | xargs -I{} gh release delete {} --yes --cleanup-tag
```

---

## Question 6: Formula Generation Approach

### Adopted: Template file over heredoc (bundled with existing work, not separate story)

Store `scripts/threedoors-a.rb.tmpl` with placeholders. Use `envsubst` to fill them. Benefits:
- Formula template is reviewable in PRs
- No heredoc indentation stripping hacks
- No sed portability issues (GNU vs BSD)
- Changes to formula structure tracked in version control

**NOT a separate story.** The current heredoc approach in PR #273 works. This is a quality improvement to bundle with the toggle or signing story when CI workflow is already being modified. "Progress over perfection."

### Rejected Options

| Option | Why Rejected |
|--------|-------------|
| Shell heredoc with sed (PR #273 current approach) | Works but fragile; sed portability concern; indentation hack |
| Separate story for template refactor | Not user-facing; bundle with existing CI work |

---

## Decisions Summary

| ID | Decision | Rationale |
|----|----------|-----------|
| CC-1 | Post-GoReleaser signing job on macOS runner for stable releases | Reuses proven alpha signing pipeline; doesn't replace GoReleaser; fail-open (unsigned is current baseline) |
| CC-2 | Single Apple Developer ID identity for both channels | Certificate is per-team; bundle ID differentiates; no need for separate certs |
| CC-3 | `vars.ALPHA_TAP_ENABLED` toggle, default OFF | Matches `SIGNING_ENABLED` pattern; conscious activation; prevents premature formula pushes |
| CC-4 | Two separate formulas (no shared template) | GoReleaser controls stable formula; different structures; copying > dependency |
| CC-5 | Alpha release verification via tap CI monitoring | Mirrors stable `release-verify.yml` pattern; lightweight; tap CI is primary gate |
| CC-6 | Keep last 30 alpha releases, delete older with `--cleanup-tag` | Prevents release pollution; formula always points to latest before cleanup; 2-6 days of history |
| CC-7 | Formula template file over heredoc (bundled, not separate story) | Cleaner, reviewable, no sed hacks; but not worth a separate story |

## Story Sequencing

| Story | Title | Dependencies |
|-------|-------|-------------|
| 38.1 | Alpha Homebrew Formula (`threedoors-a`) | None (PR #273, in review) |
| 38.2 | Alpha Publishing Toggle | 38.1 |
| 38.3 | Stable Release Signing & Notarization | None (independent) |
| 38.4 | Alpha Release Verification | 38.1, 38.2 (alpha must be live) |
| 38.5 | Alpha Release Retention Cleanup | None (independent) |
