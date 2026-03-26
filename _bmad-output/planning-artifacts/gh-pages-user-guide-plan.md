# GitHub Pages User Guide — Planning Artifact

> **Status: Planned — Epic 61 (4 stories created)**
> Research merged (PR #481, PR #500). Epic and stories created in docs/stories/61.*.story.md.

**Created:** 2026-03-10
**Planned:** 2026-03-11
**Type:** /plan-work research → implementation stories
**Epic:** 61 (pre-allocated by project-watchdog)

---

## Executive Summary

ThreeDoors has a comprehensive 1,173-line user guide (`docs/user-guide.md`) plus a 745-line README, but no web-hosted documentation. Publishing via GitHub Pages would make the guide discoverable via search engines, accessible without cloning the repo, and provide a professional landing page for new users.

## Current State

### What Exists
- **`docs/user-guide.md`** — 1,173 lines covering installation, core concepts, all features, integrations (Jira, GitHub Issues, Apple Notes, Apple Reminders, Todoist, Obsidian), CLI reference, MCP server, configuration, troubleshooting
- **`README.md`** — 745 lines with installation, quick start, feature overview, badges
- **`SOUL.md`** — 101 lines of project philosophy
- **`docs/architecture/`** — 25 files of technical architecture (developer-facing, not user-facing)
- **`docs/adapter-developer-guide.md`** — TaskProvider interface docs for extension developers
- **`CHANGELOG.md`** — 18.8KB of release history
- **No GitHub Pages infrastructure** — no `_config.yml`, `mkdocs.yml`, `gh-pages` branch, or deployment workflow

### What Doesn't Exist
- Static site generator configuration
- GitHub Actions workflow for Pages deployment
- Multi-page navigation structure
- Search functionality for docs
- Custom domain configuration
- Versioned documentation
- Screenshots/GIFs of the TUI in action

---

## Approach Decision

### Adopted: MkDocs + Material for MkDocs

**Rationale:**
1. **Markdown-native** — existing `.md` files work with zero conversion
2. **Material theme** — best-in-class docs theme with search, dark mode, navigation, mobile-responsive out of the box
3. **GoReleaser precedent** — the most successful Go CLI docs site uses MkDocs Material; following a proven pattern reduces risk
4. **Built-in search** — lunr.js client-side search, no backend needed
5. **Low maintenance** — content stays as `.md` files; site rebuilds on push to main
6. **Admonitions, tabs, code copy** — ideal for CLI docs with command examples and config snippets
7. **Dark mode** — automatic toggle, matches TUI aesthetic

### Rejected Alternatives

| Option | Why Rejected |
|--------|-------------|
| **Hugo** | Go-native is appealing but docs-specific themes (Docsy) are more complex; GoReleaser chose MkDocs Material over Hugo despite being a Go project — strong signal |
| **Jekyll** | GitHub Pages supports it natively (no CI needed) but slow builds, dated, and not designed for technical documentation |
| **Docusaurus** | Powerful but brings entire React/Node toolchain; overkill for CLI tool docs with no interactive components |
| **mdBook** | Too minimal — no built-in versioning, limited theming, primarily serves Rust ecosystem |
| **Plain HTML** | No search, no navigation, manual maintenance burden |
| **GitHub Wiki** | Not version-controlled with the repo; limited formatting; can't PR changes |

---

## Content Architecture

### Source Location Decision

**Adopted: `docs-site/` directory** (new, separate from `docs/`)

The existing `docs/` contains 262 story files, 33 ADRs, 17 PRD files — none user-facing. Mixing user-facing and internal docs would create confusion and bloat the navigation. User guide content should be curated, not auto-included.

### Proposed Site Structure

```
docs-site/
├── mkdocs.yml
├── requirements-docs.txt
└── docs/
    ├── index.md                    # Landing page (hero, tagline, quick links)
    ├── philosophy.md               # From SOUL.md — why three doors?
    ├── getting-started/
    │   ├── installation.md         # All install methods
    │   ├── quickstart.md           # First launch → first completed task (5 min)
    │   └── concepts.md             # Doors, tasks, providers, sessions
    ├── guide/
    │   ├── task-management.md      # Creating, completing, snoozing, deferring
    │   ├── search-and-commands.md  # Search (/), command palette (:)
    │   ├── doors-interaction.md    # Door selection, refresh, avoidance detection
    │   ├── themes.md               # Theme system, gallery with previews
    │   ├── keybindings.md          # Full reference table + customization
    │   └── sessions.md            # Session tracking, mood, analytics
    ├── providers/
    │   ├── overview.md             # Multi-source concept, provider pattern
    │   ├── local-files.md          # YAML task file format, examples
    │   ├── apple-notes.md          # Setup, sync behavior
    │   ├── apple-reminders.md      # Setup, sync behavior
    │   ├── jira.md                 # Configuration, field mapping
    │   ├── github-issues.md        # Token setup, filtering
    │   ├── todoist.md              # API key, project mapping
    │   └── obsidian.md             # Vault path, task format
    ├── cli/
    │   ├── commands.md             # Full CLI reference
    │   └── mcp-server.md           # MCP tools reference with schemas
    ├── configuration/
    │   ├── config-file.md          # ~/.config/threedoors/config.yaml schema
    │   ├── environment.md          # Environment variables
    │   └── data-directory.md       # Data/log file locations
    ├── advanced/
    │   ├── task-dependencies.md    # Dependency graph, blocking tasks
    │   └── extending.md            # Writing custom providers (from adapter guide)
    ├── troubleshooting.md          # Common issues, diagnostics, FAQ
    └── changelog.md                # From CHANGELOG.md
```

---

## Implementation Stories

### Story N.1: MkDocs Infrastructure & GitHub Pages Deployment

**Goal:** Set up MkDocs with Material theme and GitHub Actions workflow to deploy to GitHub Pages on push to main.

**Acceptance Criteria:**
- [ ] `docs-site/mkdocs.yml` configuration file with Material theme, site metadata, navigation stub
- [ ] `docs-site/requirements-docs.txt` pinning MkDocs + Material versions
- [ ] `docs-site/docs/index.md` landing page with project branding, feature highlights, install quick-reference, and "Get Started" link
- [ ] `docs-site/docs/philosophy.md` extracted from SOUL.md
- [ ] GitHub Actions workflow `.github/workflows/docs.yml` that builds and deploys on push to `main` (path-filtered to `docs-site/**` and `mkdocs.yml` changes only)
- [ ] Site deploys successfully to `https://arcavenae.github.io/ThreeDoors/`
- [ ] Dark/light mode toggle works
- [ ] Client-side search works (even with minimal content)
- [ ] `make docs` and `make docs-serve` targets added to Makefile for local preview
- [ ] Navigation tabs configured (even if most sections are stubs/empty)
- [ ] Existing `docs/user-guide.md` is NOT modified or moved (preserved as-is)

**Tasks:**
1. Create `docs-site/` directory structure
2. Create `docs-site/mkdocs.yml` with Material theme configuration (dark-first, deep purple + amber palette, navigation tabs, search, code copy)
3. Create `docs-site/requirements-docs.txt` with pinned dependencies
4. Create `docs-site/docs/index.md` landing page (extract hero content from README)
5. Create `docs-site/docs/philosophy.md` from SOUL.md
6. Create `.github/workflows/docs.yml` GitHub Actions workflow with path filtering
7. Add `make docs` and `make docs-serve` targets to Makefile
8. Test local build: `cd docs-site && pip install -r requirements-docs.txt && mkdocs serve`

**Effort:** Small-Medium (mostly configuration)

**Notes:**
- The `.github/workflows/docs.yml` file will require manual merge by the project owner due to OAuth workflow scope limitation (merge-queue's token lacks `workflow` scope).

---

### Story N.2: Content Split — Getting Started & Core Guide

**Goal:** Port content from the monolithic user guide into the getting-started and core guide sections of the docs site.

**Depends on:** Story N.1 merged

**Acceptance Criteria:**
- [ ] `docs-site/docs/getting-started/installation.md` — all installation methods (Homebrew stable + alpha, binary, go install, source build) with prerequisites
- [ ] `docs-site/docs/getting-started/quickstart.md` — first launch, onboarding wizard walkthrough, "your first completed task in 5 minutes" narrative
- [ ] `docs-site/docs/getting-started/concepts.md` — Three Doors philosophy, selection algorithm, behavioral science foundation, "progress over perfection"
- [ ] `docs-site/docs/guide/task-management.md` — task statuses, transitions, quick add, categorization, undo completion
- [ ] `docs-site/docs/guide/search-and-commands.md` — search (`/`), command palette (`:`), available commands
- [ ] `docs-site/docs/guide/doors-interaction.md` — door selection, refresh, feedback options (blocked/not now/needs breakdown), mood logging
- [ ] `docs-site/docs/guide/keybindings.md` — complete key binding tables for all views (doors, detail, search, help), keybinding overlay
- [ ] `docs-site/docs/guide/sessions.md` — session metrics tracking, mood correlation, pattern insights
- [ ] All content accurately reflects current implementation
- [ ] Navigation in `mkdocs.yml` updated with all new pages
- [ ] No broken internal links (verified with `mkdocs build --strict`)

**Content Sources:**
- `docs/user-guide.md` §Getting Started, §Core Concepts, §Basic Usage, §Task Management, §Search and Commands, §Snooze and Defer, §Undo Completion, §Intelligent Features, §Session Metrics
- `README.md` §Key Bindings
- `SOUL.md` for philosophy context

**Tasks:**
1. Extract getting-started sections from user-guide.md into 3 pages
2. Extract core guide sections from user-guide.md into 5 pages
3. Add MkDocs-specific enhancements: admonitions for tips/warnings, code copy buttons, keyboard key rendering (++key++)
4. Update `mkdocs.yml` navigation
5. Run `mkdocs build --strict` to verify no broken links or warnings

**Effort:** Medium (content extraction, restructuring, and verification)

---

### Story N.3: Content Split — Integrations / Task Sources

**Goal:** Create dedicated per-integration pages with setup instructions, configuration, and troubleshooting.

**Depends on:** Story N.1 merged

**Acceptance Criteria:**
- [ ] `docs-site/docs/providers/overview.md` — multi-source architecture concept, connection manager, how providers work, mixing sources
- [ ] `docs-site/docs/providers/local-files.md` — YAML task file format, file location, examples
- [ ] `docs-site/docs/providers/apple-notes.md` — prerequisites, setup, sync behavior, limitations, troubleshooting
- [ ] `docs-site/docs/providers/apple-reminders.md` — prerequisites, setup, sync behavior, limitations
- [ ] `docs-site/docs/providers/jira.md` — OAuth setup, field mapping, JQL filtering, config examples
- [ ] `docs-site/docs/providers/github-issues.md` — token setup, repo filtering, label mapping
- [ ] `docs-site/docs/providers/todoist.md` — API key setup, project mapping, sync behavior
- [ ] `docs-site/docs/providers/obsidian.md` — vault path configuration, task format, frontmatter mapping
- [ ] Each page follows consistent structure: Overview → Prerequisites → Setup → Configuration → Usage → Troubleshooting
- [ ] Navigation updated in `mkdocs.yml`
- [ ] No broken links (`mkdocs build --strict`)

**Content Sources:**
- `docs/user-guide.md` §Task Sources through §Obsidian Integration
- `docs/adapter-developer-guide.md` for architecture context
- `README.md` §Task Sources

**Tasks:**
1. Extract integration sections from user-guide.md into 8 individual pages
2. Apply consistent page structure to each (prerequisites, setup, config, usage, troubleshooting)
3. Create providers overview page explaining multi-source architecture
4. Add admonitions for platform-specific notes (e.g., Apple integrations macOS-only)
5. Update mkdocs.yml navigation
6. Verify with `mkdocs build --strict`

**Effort:** Medium (8 pages, mostly extraction + structural consistency)

---

### Story N.4: Content Split — CLI, MCP, Configuration & Advanced

**Goal:** Complete the content split with CLI reference, MCP server docs, configuration reference, and remaining feature pages.

**Depends on:** Story N.1 merged

**Acceptance Criteria:**
- [ ] `docs-site/docs/cli/commands.md` — full CLI command reference with all subcommands, flags, and examples
- [ ] `docs-site/docs/cli/mcp-server.md` — MCP server setup, available tools list, usage with LLM agents, example prompts
- [ ] `docs-site/docs/configuration/config-file.md` — complete `config.yaml` schema reference with all fields, defaults, and examples
- [ ] `docs-site/docs/configuration/environment.md` — environment variables that affect behavior
- [ ] `docs-site/docs/configuration/data-directory.md` — data directory layout, log files, session files
- [ ] `docs-site/docs/guide/themes.md` — available themes with descriptions (screenshots out of scope; note as future enhancement)
- [ ] `docs-site/docs/advanced/task-dependencies.md` — dependency types, linking, blocking behavior
- [ ] `docs-site/docs/advanced/extending.md` — writing custom TaskProvider implementations (extracted from adapter-developer-guide.md)
- [ ] `docs-site/docs/troubleshooting.md` — common issues, diagnostics commands, FAQ
- [ ] `docs-site/docs/changelog.md` — extracted from CHANGELOG.md
- [ ] All navigation finalized in `mkdocs.yml`
- [ ] Full site builds without warnings (`mkdocs build --strict`)
- [ ] All internal cross-links work

**Content Sources:**
- `docs/user-guide.md` §CLI Reference, §MCP Server, §Configuration, §Themes, §Task Dependencies, §Offline and Sync, §Troubleshooting
- `README.md` §CLI Reference, §MCP Server
- `docs/adapter-developer-guide.md` for extending section
- `CHANGELOG.md`

**Tasks:**
1. Extract CLI reference with full command documentation
2. Extract MCP server section with tool reference
3. Create configuration reference pages
4. Extract themes, dependencies, and troubleshooting pages
5. Port adapter developer guide into extending page
6. Copy and format changelog
7. Finalize all navigation in mkdocs.yml
8. Run `mkdocs build --strict` for final verification

**Effort:** Medium

---

## Story Dependency Graph

```
N.1 (Infrastructure) ──→ N.2 (Getting Started + Core Guide)
                     ├──→ N.3 (Integrations / Task Sources)
                     └──→ N.4 (CLI, Config, Advanced)
```

N.1 is the prerequisite. N.2, N.3, and N.4 can be parallelized after N.1 merges (they touch different files within `docs-site/`).

---

## Starter mkdocs.yml Configuration

```yaml
site_name: ThreeDoors
site_url: https://arcavenae.github.io/ThreeDoors/
site_description: Task management with radical simplicity — three doors, one choice
repo_url: https://github.com/arcavenae/ThreeDoors
repo_name: arcavenae/ThreeDoors

theme:
  name: material
  palette:
    - media: "(prefers-color-scheme: dark)"
      scheme: slate
      primary: deep purple
      accent: amber
      toggle:
        icon: material/brightness-7
        name: Switch to light mode
    - media: "(prefers-color-scheme: light)"
      scheme: default
      primary: deep purple
      accent: amber
      toggle:
        icon: material/brightness-4
        name: Switch to dark mode
  features:
    - navigation.tabs
    - navigation.sections
    - navigation.expand
    - navigation.top
    - search.suggest
    - search.highlight
    - search.share
    - content.code.copy
    - content.tabs.link
  icon:
    repo: fontawesome/brands/github

plugins:
  - search:
      lang: en

markdown_extensions:
  - admonition
  - pymdownx.details
  - pymdownx.superfences
  - pymdownx.tabbed:
      alternate_style: true
  - pymdownx.highlight:
      anchor_linenums: true
  - pymdownx.inlinehilite
  - pymdownx.keys
  - attr_list
  - md_in_html
  - toc:
      permalink: true

nav:
  - Home: index.md
  - Philosophy: philosophy.md
  - Getting Started:
    - Installation: getting-started/installation.md
    - Quick Start: getting-started/quickstart.md
    - Core Concepts: getting-started/concepts.md
  - User Guide:
    - Task Management: guide/task-management.md
    - Search & Commands: guide/search-and-commands.md
    - Doors Interaction: guide/doors-interaction.md
    - Themes: guide/themes.md
    - Keybindings: guide/keybindings.md
    - Sessions & Analytics: guide/sessions.md
  - Task Sources:
    - Overview: providers/overview.md
    - Local Files: providers/local-files.md
    - Apple Notes: providers/apple-notes.md
    - Apple Reminders: providers/apple-reminders.md
    - Jira: providers/jira.md
    - GitHub Issues: providers/github-issues.md
    - Todoist: providers/todoist.md
    - Obsidian: providers/obsidian.md
  - CLI & MCP:
    - CLI Commands: cli/commands.md
    - MCP Server: cli/mcp-server.md
  - Configuration:
    - Config File: configuration/config-file.md
    - Environment Variables: configuration/environment.md
    - Data Directory: configuration/data-directory.md
  - Advanced:
    - Task Dependencies: advanced/task-dependencies.md
    - Custom Providers: advanced/extending.md
  - Troubleshooting: troubleshooting.md
  - Changelog: changelog.md

extra:
  social:
    - icon: fontawesome/brands/github
      link: https://github.com/arcavenae/ThreeDoors
```

---

## GitHub Actions Workflow Sketch

```yaml
# .github/workflows/docs.yml
name: docs

on:
  push:
    branches: [main]
    paths:
      - 'docs-site/**'
      - '.github/workflows/docs.yml'

permissions:
  contents: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-python@v5
        with:
          python-version: 3.x

      - run: echo "cache_id=$(date --utc '+%V')" >> $GITHUB_ENV

      - uses: actions/cache@v4
        with:
          key: mkdocs-material-${{ env.cache_id }}
          path: ~/.cache
          restore-keys: |
            mkdocs-material-

      - run: pip install -r docs-site/requirements-docs.txt

      - run: cd docs-site && mkdocs gh-deploy --force
```

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| `docs/` dual purpose confusion | Developers unsure which docs are user-facing | Using separate `docs-site/` directory eliminates this |
| Workflow scope limitation | merge-queue can't merge PR adding `.github/workflows/docs.yml` | Flag Story N.1 PR for manual merge by project owner |
| Content drift after split | Monolithic `docs/user-guide.md` diverges from site pages | Delete `docs/user-guide.md` after N.4 completes and all content is verified on site. Or mark deprecated with redirect notice. |
| MkDocs Python dependency | Developers need Python for local docs preview | Pin in `requirements-docs.txt`; `make docs-serve` handles setup; CI deploys automatically |
| Content accuracy after port | Ported content may describe features that changed | Each story AC requires verification against current codebase |

---

## Opportunities (Out of Scope)

These are noted for future consideration but NOT part of this epic:

- Custom domain (e.g., `docs.threedoors.dev`)
- Versioned documentation via mike (stable vs alpha channels)
- Screenshots/GIFs of TUI using VHS or asciinema
- Theme gallery with visual previews
- API documentation generation from Go source
- Contributing guide for developers
- Blog/announcement section
- Per-version docs for tagged releases
- FAQ based on real user questions/issues
- Search analytics to understand what users look for

---

## Open Questions for Decision

1. **Custom domain:** Start with `arcavenae.github.io/ThreeDoors` (free, zero config) or register a domain upfront?
2. **Versioning timeline:** Add mike versioning now or defer until stable releases begin?
3. **Delete monolithic guide?** After N.4, remove `docs/user-guide.md` or keep as single-page alternative?
4. **Screenshots scope:** Include a follow-up story for VHS terminal recordings, or defer indefinitely?

---

## References

- [Material for MkDocs — Publishing](https://squidfunk.github.io/mkdocs-material/publishing-your-site/)
- [Material for MkDocs — Alternatives](https://squidfunk.github.io/mkdocs-material/alternatives/)
- [GoReleaser Docs](https://goreleaser.com) — exemplary MkDocs Material site for Go CLI
- [GitHub Pages custom domain docs](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site)
