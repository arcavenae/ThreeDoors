# Typed Comments on Story Files

Typed comments bring structure to the implementation notes that accumulate in story files during development. Instead of unstructured prose, each comment carries a prefix that makes it machine-parseable and human-scannable.

## Comment Types

| Prefix | Purpose | When to Use |
|--------|---------|-------------|
| `[decision]` | A choice between alternatives with rationale | When you pick approach A over B and want to record why |
| `[observation]` | A notable finding during implementation | When you discover something unexpected or worth noting |
| `[blocker]` | Work is blocked on something external | When you cannot proceed without external resolution |
| `[risk]` | A risk or concern identified during implementation | When you spot a potential problem that may surface later |
| `[deviation]` | Implementation differs from the original AC/design | When you intentionally diverge from the story spec |

## Format

Each typed comment uses markdown blockquote syntax with a structured layout:

```markdown
> [decision] 2026-03-29 — Chose X over Y because Z.
> Alternative considered: Y — rejected because W.
```

**Required elements:**
- Prefix in brackets: `[decision]`, `[observation]`, etc.
- ISO date: `YYYY-MM-DD`
- Em dash separator: ` — `
- One-line summary after the dash

**Optional elements:**
- Follow-up lines (indented with `>`) for additional detail
- `Alternative considered:` lines for decisions
- `Impact:` lines for risks and deviations
- `Blocked on:` lines for blockers with owner/ticket reference

## Examples

### Decision

```markdown
> [decision] 2026-03-15 — Used atomic file writes (write-tmp-sync-rename) instead of direct writes.
> Alternative considered: direct write with flock — rejected because rename is atomic on POSIX and simpler.
```

### Observation

```markdown
> [observation] 2026-03-15 — The existing TaskProvider interface already supports the filtering we need.
> No new interface methods required — can use the existing List() with client-side filtering.
```

### Blocker

```markdown
> [blocker] 2026-03-15 — Cannot test CI workflow without GitHub Actions runner access.
> Blocked on: @skippy to enable Actions on the repo. Workaround: test locally with `act`.
```

### Risk

```markdown
> [risk] 2026-03-15 — The YAML parser silently drops unknown fields, which could mask config errors.
> Impact: Users may think a config option is active when it's being ignored.
```

### Deviation

```markdown
> [deviation] 2026-03-15 — AC3 specified a dropdown menu but implemented as keyboard shortcut instead.
> Reason: TUI has no mouse support; keyboard shortcut is more consistent with existing UX patterns.
> Impact: None — functionality is equivalent, just different interaction model.
```

## Placement in Story Files

Typed comments go in the `## Implementation Notes` section of the story file:

```markdown
## Implementation Notes

> [decision] 2026-03-15 — Chose approach A over B because of X.

> [observation] 2026-03-16 — Found that the existing code already handles edge case Y.

> [blocker] 2026-03-16 — Blocked on dependency Z being merged first.
```

If the story file doesn't have an `## Implementation Notes` section, add one before the `## Provenance` section (or at the end if no Provenance section exists).

## Extraction and Tooling

Typed comments are designed to be grep-extractable:

```bash
# All decisions across all stories
grep -r '> \[decision\]' docs/stories/

# All blockers (active and resolved)
grep -r '> \[blocker\]' docs/stories/

# All deviations from a specific epic
grep -r '> \[deviation\]' docs/stories/42.*.story.md

# All typed comments from a specific story
grep '> \[' docs/stories/74.5.story.md
```

**Future tooling opportunities:**
- **Decision extraction:** Feed `[decision]` comments into `docs/decisions/BOARD.md` automatically
- **Retrospector analysis:** Pattern detection across stories (e.g., recurring blockers)
- **Dark factory evaluation:** Compare decisions and outcomes across variant implementations
- **Session handoff:** Quickly scan for `[blocker]` and `[decision]` when picking up someone else's work
- **Risk register:** Aggregate `[risk]` comments into a project-level risk view

## Guidelines

1. **Keep the taxonomy small** — only the 5 types above. If something doesn't fit, use `[observation]` as the catch-all.
2. **One comment per blockquote** — separate multiple comments with blank lines.
3. **Date is mandatory** — it enables timeline reconstruction and staleness detection.
4. **Summary must fit on one line** — use follow-up lines for detail, not the summary.
5. **Don't over-comment** — typed comments are for notable events, not a running diary. A typical story might have 2-5 typed comments.
