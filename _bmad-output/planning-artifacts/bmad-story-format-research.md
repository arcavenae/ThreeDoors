# BMAD v6 Story File Format Research

**Date:** 2026-03-11
**Purpose:** Determine canonical story file status format in BMAD Method v6

## Findings

### Upstream BMAD v6 Template (Canonical)

**Source:** `github.com/bmadcode/BMAD-METHOD` at `src/bmm/workflows/4-implementation/create-story/template.md`

The upstream v6 story template uses **plain markdown text** for status — no heading, no frontmatter:

```markdown
# Story {{epic_num}}.{{story_num}}: {{story_title}}

Status: ready-for-dev

## Story

As a {{role}},
...
```

Key observations:
- `Status:` is a **plain line** directly under the `# Story` title heading
- No `##` heading prefix
- No YAML/HTML frontmatter
- No bold (`**Status:**`) formatting
- Valid status values: `backlog`, `ready-for-dev`, `in-progress`, `done`
- The create-story workflow (`workflow.md`) confirms status is set inline, not in frontmatter

### Local BMAD Core Template (`.bmad-core/templates/story-tmpl.yaml`)

Our installed `story-template-v2` defines Status as a section with `type: choice` and values `[Draft, Approved, InProgress, Review, Done]`. The YAML template doesn't prescribe the rendered markdown format — it defines the data model. The rendering depends on whatever agent interprets the template.

### Our Local Story File Formats (Empirical)

Our 270+ story files use **four different formats**, reflecting evolution over time:

| Format | Count | Era | Example |
|--------|-------|-----|---------|
| `**Status:** Done` | 135 | Epics 16-39 (mid-project) | `**Status:** Done (PR #305)` |
| `## Status: Done` | 116 | Epics 36+ (recent) | `## Status: Done (PR #540)` |
| `Status: Done` | 17 | Epics 1-13 (early) | `Status: Done` |
| HTML comment frontmatter | 5 | Epics 1-3 (earliest) | `<!-- status: "Done" -->` |

The `## Status:` heading format is the **most recently adopted** convention in this project and is used by all stories from approximately Epic 36 onward. The `**Status:**` bold format was the longest-running convention (Epics 16-39).

### BMAD v6 Upstream vs Our Project

| Aspect | Upstream v6 | Our Project (recent) |
|--------|-------------|---------------------|
| Format | `Status: value` (plain line) | `## Status: value` (h2 heading) |
| Location | Line 3, after title | Line 3, after title |
| Values | `backlog`, `ready-for-dev`, `in-progress`, `done` | `Draft`, `Approved`, `InProgress`, `Done (PR #NNN)` |
| Frontmatter | None | None |

Both approaches place status near the top of the file, immediately after the title. Neither uses YAML/HTML frontmatter. The difference is purely whether status is a plain text line or a markdown heading.

### BMAD v6 Style Guide / Conventions

No explicit style guide or conventions document was found in the upstream repo dictating story file format. The template file itself is the authoritative reference. The `story-tmpl.yaml` in our local `.bmad-core/` defines the data model but not the rendering format.

## Recommendation

**The upstream BMAD v6 canonical format is `Status: value` as a plain text line (no heading).**

However, our project has standardized on `## Status: value` for 116+ recent stories, and this format has practical advantages:
1. **Visibility** — renders as a visible heading in GitHub/IDE markdown preview
2. **Grepability** — `^## Status:` is a more precise grep pattern than `^Status:` which could match other contexts
3. **Consistency** — our project's CLAUDE.md, worker instructions, and tooling all reference updating `## Status:`

**Pragmatic recommendation:** Continue using `## Status: value` (our established convention) rather than migrating to match upstream exactly. The semantic meaning is identical; the difference is cosmetic. Migrating 116+ files for a formatting preference would create churn with no functional benefit.

If strict upstream alignment is desired in the future, a bulk migration could be done in a single PR. But the BMAD framework itself is format-agnostic — what matters is that status is machine-readable and consistently formatted within a project.

## Sources

- Upstream template: `https://github.com/bmadcode/BMAD-METHOD/blob/main/src/bmm/workflows/4-implementation/create-story/template.md`
- Upstream workflow: `https://github.com/bmadcode/BMAD-METHOD/blob/main/src/bmm/workflows/4-implementation/create-story/workflow.md`
- Local template: `.bmad-core/templates/story-tmpl.yaml` (story-template-v2)
- Local stories: `docs/stories/*.story.md` (270+ files sampled)
- Local checklists: `.bmad-core/checklists/story-draft-checklist.md`, `story-dod-checklist.md`
