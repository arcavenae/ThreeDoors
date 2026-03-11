# GitHub Pages User Guide — Research & Plan

**Date:** 2026-03-10
**Type:** Research artifact (no implementation)
**Story:** TBD (needs story creation before implementation)

---

## 1. Static Site Generator Recommendation

### Evaluated Options

| SSG | Language | Strengths | Weaknesses | Go Ecosystem Fit |
|-----|----------|-----------|------------|------------------|
| **MkDocs Material** | Python | Built-in search, versioning (mike), admonitions, tabs, dark mode, zero JS knowledge needed | Python dependency | ★★★★★ — GoReleaser uses it |
| **Hugo** | Go | Fastest builds, Go-native, huge theme ecosystem | Docs-specific themes less polished, more config overhead | ★★★★ — Go-native but general-purpose |
| **mdBook** | Rust | Dead simple, book-style navigation, used by Rust ecosystem | Limited theming, no built-in versioning, Rust-centric | ★★★ — Simple but too minimal |
| **Docusaurus** | React/JS | Versioning built-in, MDX components, Meta-backed | Heavy JS toolchain, React knowledge needed, overkill for CLI docs | ★★ — Wrong ecosystem |
| **Jekyll** | Ruby | GitHub Pages native support (no CI needed) | Slow builds, dated, not designed for technical docs | ★★ — Legacy choice |

### Recommendation: **MkDocs Material**

**Rationale:**
1. **GoReleaser precedent** — The most successful Go CLI docs site (goreleaser.com) uses MkDocs Material. Following this proven pattern reduces risk.
2. **Zero JS/React overhead** — Pure Markdown authoring. No build toolchain beyond `pip install mkdocs-material`.
3. **Built-in search** — Client-side lunr.js search with suggestions, highlighting, and sharing. No external service needed.
4. **Versioning via mike** — Native integration for `stable` vs `alpha` docs without complex config.
5. **Admonitions, tabs, code annotations** — Perfect for CLI docs (command examples, config snippets, warnings).
6. **Dark mode** — Automatic toggle, respects system preference. Matches TUI aesthetic.
7. **Minimal maintenance** — Single `mkdocs.yml` config file. Content is just `.md` files.

**Rejected alternatives:**
- **Hugo**: Go-native is appealing but docs-specific themes (Docsy, etc.) are more complex to configure than MkDocs Material, and the Go ecosystem's best docs site (GoReleaser) chose MkDocs Material over Hugo despite being a Go project. That's a strong signal.
- **Docusaurus**: Powerful but brings an entire React/Node toolchain. Overkill for a CLI project with no interactive components needed.
- **mdBook**: Too minimal — no built-in versioning, limited theming, and primarily serves the Rust ecosystem.
- **Jekyll**: GitHub Pages supports it natively (no CI needed), but it's slow, dated, and not designed for technical documentation.

---

## 2. Content Audit — What Exists vs. What's Needed

### Existing Documentation Inventory

| Source | Content | User-Facing? | Reusable for Guide? |
|--------|---------|:------------:|:-------------------:|
| `README.md` (28.5KB) | Installation, features, keybindings, CLI ref, MCP tools, config | ✅ | ✅ High — primary source |
| `docs/user-guide.md` (35KB) | Comprehensive feature walkthrough, all workflows | ✅ | ✅ High — richest source |
| `SOUL.md` | Philosophy, design values, what ThreeDoors is NOT | ✅ | ✅ Philosophy section |
| `CHANGELOG.md` (18.8KB) | Release history, feature summaries | ✅ | ✅ Release notes section |
| `docs/adapter-developer-guide.md` | TaskProvider interface, contract tests | Partial | ✅ Developer/extension guide |
| `docs/architecture/` (19 files) | Full architecture documentation | ❌ Dev only | ⚠️ Selected sections only |
| `docs/ADRs/` (33 ADRs) | Architectural decisions | ❌ Dev only | ❌ Not user-facing |
| `docs/prd/` (17 files) | Product requirements | ❌ Internal | ❌ Not user-facing |
| `docs/stories/` (262 files) | Story specs and acceptance criteria | ❌ Internal | ❌ Not user-facing |

### Content Gaps (not currently documented anywhere)

1. **Visual screenshots/GIFs** — No screenshots of the TUI in action (doors view, search, themes)
2. **Task file format reference** — YAML schema documented in architecture but not in user-facing docs
3. **Theme gallery** — Themes exist but no visual preview/comparison
4. **Provider setup walkthroughs** — Config examples exist but no step-by-step setup guides per provider
5. **Troubleshooting runbook** — Brief section in user-guide; needs expansion
6. **FAQ** — No dedicated FAQ
7. **Migration guide** — No docs for users coming from other task managers
8. **MCP tool detailed reference** — Tools listed in README but no parameter schemas or examples per tool
9. **Session analytics interpretation** — How to read JSONL session logs, what metrics mean
10. **Keyboard shortcut cheat sheet** — Exists scattered; needs a printable/downloadable reference

---

## 3. Proposed Guide Structure

```
docs-site/
├── mkdocs.yml                    # Site configuration
├── docs/
│   ├── index.md                  # Landing page (hero, tagline, quick links)
│   ├── philosophy.md             # From SOUL.md — why three doors?
│   │
│   ├── getting-started/
│   │   ├── installation.md       # Homebrew, binary, go install, source
│   │   ├── quickstart.md         # First launch → first completed task (5 min)
│   │   └── concepts.md           # Doors, tasks, providers, sessions
│   │
│   ├── guide/
│   │   ├── task-management.md    # Creating, completing, snoozing, deferring
│   │   ├── search-and-commands.md # Search (/), command palette (:)
│   │   ├── doors-interaction.md  # Door selection, refresh, avoidance detection
│   │   ├── themes.md             # Theme system, gallery with previews
│   │   ├── keybindings.md        # Full reference table + customization
│   │   └── sessions.md           # Session tracking, mood, analytics
│   │
│   ├── providers/
│   │   ├── overview.md           # Multi-source concept, provider pattern
│   │   ├── local-files.md        # YAML task file format, examples
│   │   ├── apple-notes.md        # Setup, sync behavior
│   │   ├── apple-reminders.md    # Setup, sync behavior
│   │   ├── jira.md               # Configuration, field mapping
│   │   ├── github-issues.md      # Token setup, filtering
│   │   ├── todoist.md            # API key, project mapping
│   │   └── obsidian.md           # Vault path, task format
│   │
│   ├── cli/
│   │   ├── commands.md           # Full CLI reference (threedoors ...)
│   │   └── mcp-server.md         # MCP tools reference with schemas
│   │
│   ├── configuration/
│   │   ├── config-file.md        # ~/.config/threedoors/config.yaml schema
│   │   ├── environment.md        # Environment variables
│   │   └── data-directory.md     # Data/log file locations
│   │
│   ├── advanced/
│   │   ├── integrations.md       # LLM agents, automation, scripting
│   │   ├── task-dependencies.md  # Dependency graph, blocking tasks
│   │   └── extending.md          # Writing custom providers (from adapter guide)
│   │
│   ├── troubleshooting.md        # Common issues, diagnostics, FAQ
│   ├── changelog.md              # From CHANGELOG.md
│   └── contributing.md           # How to contribute
```

### Navigation Design (mkdocs.yml)

```yaml
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
    - Integrations: advanced/integrations.md
    - Task Dependencies: advanced/task-dependencies.md
    - Custom Providers: advanced/extending.md
  - Troubleshooting: troubleshooting.md
  - Changelog: changelog.md
  - Contributing: contributing.md
```

---

## 4. GitHub Actions Workflow

### Recommended: Separate Docs Workflow

A separate workflow (not modifying the existing `ci.yml`) that deploys on merge to main when docs content changes.

```yaml
# .github/workflows/docs.yml
name: docs

on:
  push:
    branches: [main]
    paths:
      - 'docs-site/**'
      - 'mkdocs.yml'
      - '.github/workflows/docs.yml'

permissions:
  contents: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0  # Required for mike versioning

      - uses: actions/setup-python@v5
        with:
          python-version: 3.x

      - run: echo "cache_id=$(date --utc '+%V')" >> $GITHUB_ENV

      - uses: actions/cache@v5
        with:
          key: mkdocs-material-${{ env.cache_id }}
          path: ~/.cache
          restore-keys: |
            mkdocs-material-

      - run: pip install mkdocs-material mike

      # For initial setup (no versioning):
      - run: mkdocs gh-deploy --force

      # For versioned docs (later):
      # - run: mike deploy --push --update-aliases $VERSION latest
```

### Integration with Existing CI

The existing `ci.yml` already has a `docs-pass` job that detects docs-only PRs and skips code CI. The new `docs.yml` workflow complements this — it only runs on push to main (not on PRs), so docs deploy only after merge.

### Path Filtering Decision

Two options for where docs source lives:

| Option | Pros | Cons |
|--------|------|------|
| **`docs-site/`** (new directory) | Clean separation from dev docs, no confusion | New directory, need to keep in sync |
| **`docs/`** (existing) | Reuse existing content, single source | Mixes user-facing with dev-internal docs, harder to filter |

**Recommendation:** New `docs-site/` directory. The existing `docs/` contains 262 story files, 33 ADRs, 17 PRD files — none user-facing. Mixing these in would create confusion and bloat the nav. User guide content should be curated, not auto-included.

---

## 5. Custom Domain

### Setup Steps

1. **Register domain** (e.g., `docs.threedoors.dev` or `threedoors.dev`)
2. **DNS configuration:**
   - For apex domain: A records pointing to GitHub Pages IPs (185.199.108-111.153)
   - For subdomain: CNAME record pointing to `arcaven.github.io`
3. **Repository configuration:**
   - Settings → Pages → Custom domain → enter domain
   - This creates a `CNAME` file in the `gh-pages` branch
4. **HTTPS:** GitHub auto-provisions Let's Encrypt certificate (takes ~15 min after DNS propagates)
5. **MkDocs config:**
   ```yaml
   # mkdocs.yml
   site_url: https://docs.threedoors.dev/
   ```

### Domain Options

| Domain | Cost | Notes |
|--------|------|-------|
| `arcaven.github.io/ThreeDoors` | Free | Default, no setup needed |
| `threedoors.dev` | ~$12/yr | Professional, matches project name |
| `docs.threedoors.dev` | Same | Subdomain approach, separates docs from potential landing page |

**Recommendation:** Start with `arcaven.github.io/ThreeDoors` (free, zero config). Add custom domain later if the project grows. The MkDocs `site_url` can be updated without rebuilding content.

---

## 6. Search

### MkDocs Material Built-in Search

MkDocs Material includes a built-in search plugin powered by **lunr.js** — no external service needed.

**Features included out of the box:**
- Client-side full-text search (no server required)
- Search suggestions (autocomplete on typing)
- Search highlighting (highlights matches on result pages)
- Search sharing (deep-linkable search queries)
- Keyboard shortcut (`/` or `s` to focus search — matches ThreeDoors' own TUI shortcut!)
- Multi-language stemming support

**Configuration:**

```yaml
# mkdocs.yml
plugins:
  - search:
      lang: en
      separator: '[\s\-\.]+'

theme:
  features:
    - search.suggest
    - search.highlight
    - search.share
```

**Performance:** The search index is shipped as a static JSON file. For a docs site of this size (~25 pages), the index will be <50KB gzipped. No performance concerns.

**No external services needed.** Algolia DocSearch is an alternative for very large sites, but lunr.js is more than sufficient here and avoids external dependencies — consistent with ThreeDoors' local-first philosophy.

---

## 7. Versioned Documentation

### Strategy: Stable vs. Alpha

ThreeDoors already ships two channels via Homebrew:
- `threedoors` — stable releases
- `threedoors-a` — alpha channel (every push to main)

The docs should mirror this:

| Version | Content | Default? |
|---------|---------|----------|
| `stable` | Docs for latest tagged release | ✅ Yes |
| `alpha` | Docs for latest main (may include unreleased features) | No |

### Implementation with mike

```yaml
# mkdocs.yml
extra:
  version:
    provider: mike
    default: stable
```

**Deployment workflow (versioned):**

```yaml
# For stable releases (triggered by release.yml):
- run: |
    pip install mkdocs-material mike
    mike deploy --push --update-aliases $TAG stable
    mike set-default --push stable

# For alpha (triggered on push to main):
- run: |
    pip install mkdocs-material mike
    mike deploy --push alpha
```

**User experience:** A version selector dropdown appears in the header. Users on `stable` see release-matching docs. Users on `alpha` see the latest.

### Phased Rollout

1. **Phase 1:** No versioning. Single docs site deployed on push to main. (`mkdocs gh-deploy --force`)
2. **Phase 2:** Add mike versioning when the first stable release is tagged. Deploy `stable` + `alpha`.
3. **Phase 3:** (If needed) Per-version docs (v1.0, v1.1, etc.) using mike's full version management.

**Recommendation:** Start with Phase 1. Versioning adds complexity and the project is pre-1.0. Add mike when stable releases begin.

---

## 8. Exemplary Go Documentation Sites

### GoReleaser (goreleaser.com) — MkDocs Material

- **SSG:** MkDocs Material
- **Structure:** Getting Started → Docs (customization, builds, packaging, publishing) → Pro → Blog
- **Search:** Built-in MkDocs Material search with suggestions and highlighting
- **Versioning:** Version selector in header
- **Design:** Clean dark theme, code tabs for different languages, admonitions for tips/warnings
- **Takeaway:** Best-in-class Go CLI docs. Direct model for ThreeDoors.

### Charm.sh → charm.land

- **Design:** Highly custom, brand-focused landing page
- **Approach:** Individual tool docs (Bubbletea, Lipgloss, etc.) live in GitHub READMEs and pkg.go.dev
- **Takeaway:** Not a good model for structured user documentation. More of a portfolio/brand site.

### Cobra (cobra.dev)

- **Structure:** Getting Started → How-To → Reference → Blog
- **Design:** Custom theme with night/winter modes
- **Takeaway:** Good structure (tutorial → how-to → reference pattern) but less polished than GoReleaser.

### Key Patterns from Exemplary Sites

1. **"Getting Started" is always first** — Installation → Quick Start → Core Concepts
2. **Separate reference from tutorials** — CLI commands are reference; "managing tasks" is a tutorial
3. **Code examples are king** — Every concept demonstrated with copy-pasteable examples
4. **Dark mode default** — Terminal tool users expect dark themes
5. **Minimal landing page** — Hero + tagline + "Get Started" button. Don't bury the lede.

---

## 9. Minimal `mkdocs.yml` Starter

```yaml
site_name: ThreeDoors
site_url: https://arcaven.github.io/ThreeDoors/
site_description: Task management with radical simplicity — three doors, one choice
repo_url: https://github.com/arcaven/ThreeDoors
repo_name: arcaven/ThreeDoors

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
  - pymdownx.keys           # Render keyboard shortcuts like ++ctrl+k++
  - attr_list
  - md_in_html
  - toc:
      permalink: true

extra:
  social:
    - icon: fontawesome/brands/github
      link: https://github.com/arcaven/ThreeDoors
```

---

## 10. Implementation Estimate & Phasing

### Phase 1: MVP Docs Site
- Set up `docs-site/` directory with `mkdocs.yml`
- Port content from README.md and docs/user-guide.md into structured pages
- Landing page with hero
- Getting Started (installation, quickstart, concepts)
- Core guide sections (task management, search, themes, keybindings)
- CLI reference
- GitHub Actions workflow for auto-deploy
- **Scope:** ~15 pages, largely ported from existing content

### Phase 2: Deep Content
- Provider setup walkthroughs (one page per provider)
- MCP server detailed reference with tool schemas
- Configuration reference (full YAML schema)
- Screenshots and terminal recordings (asciinema/VHS)
- Troubleshooting expansion
- **Scope:** ~10 additional pages + visual assets

### Phase 3: Polish
- mike versioning (stable vs alpha)
- Custom domain
- Theme gallery with visual previews
- FAQ based on real user questions
- Contributing guide
- **Scope:** Configuration + content refinement

### Content Sourcing Strategy

Most content already exists and can be ported:

| Target Page | Source | Effort |
|-------------|--------|--------|
| Installation | README.md §Installation | Low — copy + reformat |
| Quick Start | README.md §Quick Start | Low — copy + reformat |
| Core Concepts | docs/user-guide.md §Core Concepts | Low — copy + reformat |
| Task Management | docs/user-guide.md §Task Management | Medium — restructure |
| Keybindings | README.md §Key Bindings | Low — table reformatting |
| CLI Reference | README.md §CLI Reference | Low — copy + expand |
| MCP Server | README.md §MCP + docs/user-guide.md | Medium — add schemas |
| Configuration | README.md §Configuration | Medium — expand examples |
| Themes | docs/user-guide.md §Themes | Medium — add gallery |
| Providers | docs/user-guide.md §Task Sources | High — split + expand per provider |
| Troubleshooting | docs/user-guide.md §Troubleshooting | Medium — expand |
| Philosophy | SOUL.md | Low — light editing |

---

## 11. Open Questions for Decision

1. **Docs source location:** `docs-site/` (recommended) vs. reuse `docs/`?
2. **Custom domain:** Start with `arcaven.github.io/ThreeDoors` or register `threedoors.dev` upfront?
3. **Screenshots:** Use static PNGs or terminal recordings (VHS/asciinema)?
4. **Versioning timeline:** Add mike now or wait for stable release?
5. **Scope of initial launch:** MVP (Phase 1 only) or Phase 1+2 together?

---

## References

- [Material for MkDocs — Publishing](https://squidfunk.github.io/mkdocs-material/publishing-your-site/)
- [Material for MkDocs — Alternatives Comparison](https://squidfunk.github.io/mkdocs-material/alternatives/)
- [Material for MkDocs — Versioning with mike](https://squidfunk.github.io/mkdocs-material/setup/setting-up-versioning/)
- [Material for MkDocs — Search Plugin](https://squidfunk.github.io/mkdocs-material/plugins/search/)
- [GoReleaser Docs](https://goreleaser.com) — exemplary MkDocs Material site for Go CLI
- [mike — MkDocs version manager](https://github.com/jimporter/mike)
- [GitHub Pages custom domain docs](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site)
