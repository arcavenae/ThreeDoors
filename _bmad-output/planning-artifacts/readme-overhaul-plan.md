# README.md Overhaul Plan

> **Status: Planned — see Epic XX (README Overhaul), Stories XX.1–XX.5 (PR #502)**
> Stories created but epic number not yet assigned. Stories XX.1–XX.5 exist in `docs/stories/`.

> Research artifact — no code changes.
> Worker: nice-penguin | Date: 2026-03-10

---

## 1. Executive Summary

The current README.md is **already comprehensive** (747 lines) with good content coverage. The overhaul focuses on:
- **Visual polish** — badges, emojis, hero section
- **Scannability** — foldable sections, TOC, navigation
- **Completeness** — feature list audit against 35+ completed epics
- **Best practices** — inspired by bubbletea, lazygit, fzf, glow READMEs

---

## 2. Badges

### Current Badges (3)
```markdown
[![Go Version](https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Built with Bubbletea](https://img.shields.io/badge/Built%20with-Bubbletea-purple)](https://github.com/charmbracelet/bubbletea)
```

### Proposed Badge Set (10)

Wrap in `<p align="center">` for centered display (like glow/lazygit):

```markdown
<p align="center">
  <!-- Build & Quality -->
  <a href="https://github.com/arcavenae/ThreeDoors/actions/workflows/ci.yml"><img src="https://github.com/arcavenae/ThreeDoors/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/arcavenae/ThreeDoors"><img src="https://goreportcard.com/badge/github.com/arcavenae/ThreeDoors" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/arcavenae/ThreeDoors"><img src="https://pkg.go.dev/badge/github.com/arcavenae/ThreeDoors.svg" alt="Go Reference"></a>

  <!-- Version & Distribution -->
  <a href="https://github.com/arcavenae/ThreeDoors/releases/latest"><img src="https://img.shields.io/github/v/release/arcaven/ThreeDoors?style=flat&label=release&color=green" alt="Latest Release"></a>
  <a href="https://github.com/arcavenae/ThreeDoors/releases"><img src="https://img.shields.io/github/downloads/arcaven/ThreeDoors/total?style=flat&color=blue" alt="Downloads"></a>
  <a href="https://formulae.brew.sh/formula/threedoors"><img src="https://img.shields.io/badge/homebrew-threedoors-FBB040?logo=homebrew&logoColor=white" alt="Homebrew"></a>

  <!-- Meta -->
  <a href="https://img.shields.io/badge/Go-1.25.4+-00ADD8?logo=go&logoColor=white"><img src="https://img.shields.io/badge/Go-1.25.4+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go Version"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/Built%20with-Bubbletea-purple?logo=data:image/svg+xml;base64,..." alt="Built with Bubbletea"></a>
  <a href="https://github.com/arcavenae/ThreeDoors/stargazers"><img src="https://img.shields.io/github/stars/arcaven/ThreeDoors?style=flat&color=yellow" alt="Stars"></a>
</p>
```

### Badge Notes
- **Go Report Card**: `goreportcard.com/report/github.com/arcavenae/ThreeDoors` — auto-generated, no setup needed
- **GitHub Actions CI**: Uses the workflow file name `ci.yml` — confirmed exists
- **pkg.go.dev**: Auto-indexes public Go modules
- **Homebrew badge**: Static shield since it's a custom tap (`arcavenae/tap/threedoors`), not formulae.brew.sh — adjust URL to point to tap repo
- **Downloads**: `github.com/arcavenae/ThreeDoors` total downloads from releases
- **Stars**: Social proof
- **Test coverage**: Omitted for now — requires codecov/coveralls integration (could be a follow-up story). Add when available:
  ```
  [![Coverage](https://codecov.io/gh/arcavenae/ThreeDoors/branch/main/graph/badge.svg)](https://codecov.io/gh/arcavenae/ThreeDoors)
  ```

---

## 3. Table of Contents

### Recommendation: Manual TOC with Anchor Links

GitHub doesn't support auto-generated TOC natively. Options:
1. **Manual TOC** ✅ — Full control, works everywhere
2. **`gh-md-toc`** — External tool, requires running before commit
3. **GitHub's built-in TOC button** — Exists (hamburger icon on rendered README) but not visible inline

### Proposed TOC

```markdown
## 📑 Table of Contents

- [🚪 What is ThreeDoors?](#-what-is-threedoors)
- [📸 Screenshots](#-screenshots)
- [📦 Installation](#-installation)
- [🚀 Quick Start](#-quick-start)
- [✨ Features](#-features)
- [⌨️ Key Bindings](#️-key-bindings)
- [💻 CLI Reference](#-cli-reference)
- [🤖 MCP Server](#-mcp-server)
- [🎨 Themes](#-themes)
- [📁 Data & Privacy](#-data--privacy)
- [🔧 Configuration](#-configuration)
- [🛠️ Development](#️-development)
- [🧭 Philosophy](#-philosophy)
- [🤝 Contributing](#-contributing)
- [📚 Documentation](#-documentation)
- [📜 License](#-license)
```

Place immediately after the hero section (badges + tagline), before "What is ThreeDoors?"

---

## 4. Foldable Content (`<details><summary>`)

### Sections to Make Foldable

These sections are **reference material** that clutters scanning. Wrap in `<details>`:

#### 4.1 Installation Options (keep Quick Start visible, fold details)
```markdown
<details>
<summary>📦 <strong>Installation Options</strong> (Homebrew, binary, go install, source)</summary>

### Option 1: Homebrew (recommended)
...
### Option 2: Download Pre-built Binary
...
### Option 3: Install with `go install`
...
### Option 4: Build from Source
...

</details>
```

#### 4.2 Configuration Reference
```markdown
<details>
<summary>🔧 <strong>Configuration Reference</strong> — config.yaml provider setup</summary>

Full YAML config example with all providers...

</details>
```

#### 4.3 CLI Command Reference (full details)
```markdown
<details>
<summary>💻 <strong>Full CLI Reference</strong> — all commands and flags</summary>

### `task` — Task Management
...
### `doors` — Three Doors in the Terminal
...
(etc.)

</details>
```

#### 4.4 Key Bindings Tables
```markdown
<details>
<summary>⌨️ <strong>Key Bindings</strong> — all views</summary>

### Three Doors View
| Key | Action |
...

### Task Detail View
...

### Command Palette
...

</details>
```

#### 4.5 MCP Server Details
```markdown
<details>
<summary>🤖 <strong>MCP Server</strong> — setup, tools, Claude Desktop config</summary>

### Running the MCP Server
...
### Available MCP Tools
...
### Claude Desktop Configuration
...

</details>
```

#### 4.6 Development Setup
```markdown
<details>
<summary>🛠️ <strong>Development</strong> — tech stack, project structure, make targets</summary>

...

</details>
```

#### 4.7 Data Directory Structure
```markdown
<details>
<summary>📁 <strong>Data Directory</strong> — ~/.threedoors/ contents</summary>

...

</details>
```

### What Stays Expanded
- Hero/badges/tagline
- TOC
- "What is ThreeDoors?" (elevator pitch)
- Screenshots/GIFs
- Quick Start (short version)
- Features list (the selling point)
- Philosophy
- Contributing (short version)
- License
- Links

**Result:** The README becomes a **landing page** with scannable highlights, and users expand sections they care about.

---

## 5. Navigation

### What GitHub Supports
- **Anchor links** in TOC → ✅ use extensively
- **"Back to top" links** → ✅ add `[↑ Back to top](#threedoors-)` after each major section
- **Wiki sidebar** → Not currently using a wiki; could add but low priority
- **Pinned header links** → Not supported by GitHub markdown

### Proposed Navigation Pattern
```markdown
<!-- After each major section -->
<div align="right"><a href="#threedoors-">↑ Back to top</a></div>
```

### Cross-Link Strategy
- Link from Features → relevant foldable section (e.g., "See [CLI Reference](#-full-cli-reference)")
- Link from Quick Start → Installation details
- Link from Features → MCP section, Themes section, etc.

---

## 6. Features List — Comprehensive Audit

### Audit Methodology
Cross-referenced all 35 completed epics from ROADMAP.md against current README features section.

### Current Coverage Assessment
The current features section is **already quite good**. Missing items from completed epics:

| Epic | Feature | Currently Listed? |
|------|---------|-------------------|
| 1 | Three Doors core | ✅ |
| 2 | Apple Notes integration | ✅ |
| 3 | Enhanced interaction (search, commands) | ✅ |
| 3.5 | Platform readiness / tech debt | N/A (internal) |
| 4 | Learning & intelligent selection | ✅ |
| 5 | macOS distribution | ✅ |
| 6 | Enrichment database | ✅ |
| 7 | Plugin/adapter SDK | ✅ (implicit in multi-provider) |
| 8 | Obsidian integration | ✅ |
| 9 | Testing strategy | N/A (internal) |
| 10 | First-run onboarding | ✅ |
| 11 | Sync observability / offline-first | ✅ |
| 12 | Calendar awareness | ✅ |
| 13 | Multi-source aggregation | ✅ |
| 14 | LLM task decomposition | ✅ |
| 15 | Psychology research | N/A (internal) |
| 17 | Door themes | ✅ |
| 18 | Docker E2E testing | N/A (internal) |
| 19 | Jira integration | ✅ |
| 20 | Apple Reminders | ✅ |
| 21 | Sync protocol hardening | ✅ (WAL, offline queue) |
| 22 | Self-driving dev pipeline | N/A (internal) |
| 23 | CLI interface | ✅ |
| 24 | MCP server | ✅ |
| 25 | Todoist integration | ✅ |
| 26 | GitHub Issues integration | ✅ |
| 27 | Daily Planning Mode | ❌ **MISSING** |
| 28 | Snooze/Defer | ✅ |
| 29 | Task Dependencies | ✅ |
| 32 | Undo completion | ✅ |
| 33 | Seasonal themes (3/4 done) | ❌ **MISSING** |
| 34 | SOUL.md + dev skills | N/A (internal) |
| 35 | Door visual appearance | ❌ **MISSING** (proportional doors) |
| 36 | Door selection feedback | ❌ **MISSING** (tactile selection) |
| 37 | Persistent BMAD agents | N/A (internal) |
| 38 | Dual Homebrew distribution | ✅ (alpha channel) |
| 39 | Keybinding display system | ✅ (? overlay, context bar) |
| 40 | Beautiful Stats | ❌ **MISSING** (sparklines, heatmaps, milestone celebrations) |
| 41 | Charm Ecosystem | ❌ **MISSING** (huh forms, etc.) |

### Missing Features to Add

1. **📅 Daily Planning Mode** — Morning planning ritual: review yesterday's progress, set today's priorities, plan your three focus doors
2. **🌸 Seasonal Themes** — Auto-switching seasonal variants (spring, summer, autumn, winter) based on current date
3. **🚪 Door-Like Visuals** — Proportional door rendering with visual weight and realistic proportions
4. **✨ Selection Feedback** — Tactile selection animations with visual confirmation
5. **📊 Beautiful Statistics** — Sparkline charts, activity heatmaps, streak flames, progress bars in the TUI
6. **🧩 Charm Ecosystem** — Built with charmbracelet/huh for wizard forms, modern TUI components

### Proposed Features Organization (by category)

```markdown
## ✨ Features

### 🚪 Core Experience
- Three Doors Display — view three randomly selected tasks
- Refresh mechanism — re-roll when nothing appeals
- Door-like visuals — proportional rendering with realistic door aesthetics
- Selection feedback — tactile animations and visual confirmation
- Seven-state task workflow (todo → in-progress → in-review → complete, blocked, deferred, archived)
- Quick add with inline tagging (#creative @work)
- Context capture (:add --why)
- Cross-reference linking between tasks
- Task dependencies with auto-unblock on completion
- Undo completion — reverse accidental completions
- Snooze/defer — hide tasks until a specific date

### 🔍 Search & Commands
- Quick search (/) with fuzzy filtering
- Vi-style command palette (:)
- 15+ built-in commands

### 📊 Analytics & Insights
- Beautiful statistics — sparklines, heatmaps, streak flames, progress bars
- Session metrics — selections, bypasses, timing
- Daily completion tracking with streaks
- Insights dashboard (:dashboard)
- Mood correlation analysis
- Avoidance detection (10+ bypasses → intervention)
- Pattern analysis — position bias, type preferences
- Daily planning mode — morning ritual for setting priorities

### 😊 Wellbeing
- Mood logging — 7 presets (focused, energized, tired, stressed, neutral, calm, distracted)
- Mood-productivity correlation
- Burnout risk assessment (via MCP)
- Values & goals display
- Door feedback — rate doors for improved selection

### 🎨 Themes & Visuals
- 4 door themes — classic, modern, scifi, shoji
- Seasonal variants — spring, summer, autumn, winter (auto-switch)
- Live theme picker (:theme)
- Context-sensitive keybinding bar
- Full keybinding overlay (?)

### 🔌 Integrations
- 📄 Text File (YAML) — default local storage
- 🍎 Apple Notes — bidirectional sync
- 📋 Apple Reminders — macOS Reminders sync
- 🔵 Jira — REST API with JQL filtering
- 🐙 GitHub Issues — repository issue import
- 💎 Obsidian — vault task reading with daily notes
- ✅ Todoist — REST API with project/filter support
- 🔌 Multi-provider aggregation with deduplication
- 🩺 Health check (:health / CLI)

### 💾 Reliability
- Write-ahead log (WAL) — crash-safe persistence
- Offline queue — local changes replay on reconnect
- Sync status indicator per provider
- Conflict resolution — duplicate detection and merge UI
- Atomic file writes

### 📅 Calendar & Planning
- macOS Calendar.app reader (AppleScript)
- ICS file and CalDAV cache support
- Free block detection between events
- Daily planning mode — morning review and prioritization

### 🤖 AI & Automation
- LLM task decomposition — Claude or Ollama
- Git integration for generated stories
- Suggestions view (:suggestions) for LLM proposals
- MCP server — 15 tools for LLM agent integration
- Dual transport — stdio and SSE

### 💻 Interfaces
- Interactive TUI — Bubbletea-powered terminal UI
- Headless CLI — scripting, automation, --json output
- MCP server — LLM agent integration
- Shell completions — bash, zsh, fish
- First-run onboarding wizard

### 📦 Distribution
- Homebrew — stable + alpha channels (side-by-side)
- macOS code-signed and Apple-notarized
- Cross-platform binaries (macOS ARM/Intel, Linux x86_64)
- Automatic GitHub Releases on merge to main

### 🔒 Privacy
- All data local — ~/.threedoors/
- No telemetry, no accounts, no tracking
- Offline-first — works without network
- API tokens stay local
```

---

## 7. Emojis Strategy

### Philosophy
Tasteful, consistent, functional — not decorative spam. Each emoji serves as a **visual anchor** for scanning.

### Where to Use Emojis

| Location | Pattern | Example |
|----------|---------|---------|
| Section headers | One emoji prefix | `## 📦 Installation` |
| Feature category headers | One emoji prefix | `### 🚪 Core Experience` |
| Feature bullet points | One emoji prefix per item | `- 🔍 **Quick Search** — ...` |
| Badge labels | Logo icons via shields.io | `?logo=go&logoColor=white` |
| Status indicators | ✅ ❌ 🚧 ⚠️ | In feature comparison tables |
| TOC entries | Match section emoji | `- [📦 Installation](#-installation)` |
| Tagline/hero | Sparingly, brand-relevant | `🚪🚪🚪` (keep existing) |

### Emoji Palette (consistent set)

```
🚪 — doors/core concept     📦 — installation/distribution
✨ — features               🔍 — search
📊 — analytics/stats        😊 — mood/wellbeing
🎨 — themes/visuals         🔌 — integrations
💾 — reliability/data       📅 — calendar/planning
🤖 — AI/automation          💻 — CLI/interfaces
🔒 — privacy/security       🛠️ — development
🧭 — philosophy             🤝 — contributing
📚 — documentation          📜 — license
📑 — table of contents      🚀 — quick start
📸 — screenshots            ⌨️ — keybindings
```

### What NOT to Do
- Don't use multiple emojis per line
- Don't use emojis in code blocks
- Don't use emojis in badge alt-text (accessibility)
- Don't use random/unrelated emojis — keep the palette consistent

---

## 8. Screenshots / GIFs

### Placeholder Strategy

Since we don't have screenshots yet, create placeholder sections with clear instructions:

```markdown
## 📸 Screenshots

<p align="center">
  <!-- TODO: Add hero GIF showing three doors selection flow -->
  <em>🚧 Hero screenshot coming soon — three doors selection in action</em>
</p>

<details>
<summary>📷 More screenshots</summary>

| View | Screenshot |
|------|-----------|
| Three Doors | <!-- TODO: doors view screenshot --> |
| Task Detail | <!-- TODO: detail view screenshot --> |
| Dashboard | <!-- TODO: insights dashboard --> |
| Theme Picker | <!-- TODO: theme picker with 4 themes --> |
| Seasonal Themes | <!-- TODO: seasonal variants --> |
| Search | <!-- TODO: quick search in action --> |
| Onboarding | <!-- TODO: first-run wizard --> |

</details>
```

### Recommended Screenshots to Capture (follow-up story)
1. **Hero GIF** (most impactful) — 3-5 second animation showing door selection flow
2. **Four themes side-by-side** — classic, modern, scifi, shoji
3. **Seasonal variants** — spring, summer, autumn, winter
4. **Dashboard/stats view** — sparklines, heatmaps
5. **Search in action** — fuzzy filtering
6. **Onboarding wizard** — first-run experience
7. **CLI output** — `threedoors doors` and `threedoors stats`

### Technical Notes
- Use `vhs` (charmbracelet/vhs) to record terminal GIFs — it's the Charm ecosystem tool for this
- Compress GIFs (like lazygit does: `*-compressed.gif`)
- Max width: 800px for GitHub rendering
- Store in `docs/images/` or `assets/` directory

---

## 9. Inspiration from Best-in-Class READMEs

### Bubbletea
- ✅ Hero image/logo at top
- ✅ Animated GIF demo
- ✅ Tutorial walkthrough
- ✅ "In the Wild" showcase section
- Take: Hero visual + demo GIF pattern

### Lazygit
- ✅ Centered badge cluster
- ✅ Feature → GIF proof pairs
- ✅ Conversational elevator pitch
- ✅ Comprehensive install section
- ✅ Contributor avatar grid
- Take: Badge cluster, feature→proof pattern, personality in pitch

### fzf
- ✅ `<details>` foldable sections
- ✅ GitHub admonitions (`> [!TIP]`, `> [!NOTE]`)
- ✅ Extensive screenshots in tables
- ✅ Auto-TOC
- Take: Foldable sections, admonitions for tips/warnings

### Glow
- ✅ `<p align="center">` for badges and hero
- ✅ Clean, minimal structure
- ✅ Personality in descriptions
- Take: Centered layout, clean structure

---

## 10. Proposed README Structure

```
┌─────────────────────────────────────────────┐
│  # ThreeDoors 🚪🚪🚪                       │
│                                             │
│  <p align="center">                         │
│    [badge] [badge] [badge] ... (10 badges)  │
│  </p>                                       │
│                                             │
│  <p align="center">                         │
│    <em>tagline / one-liner</em>             │
│  </p>                                       │
│                                             │
│  <p align="center">                         │
│    [Hero GIF placeholder]                   │
│  </p>                                       │
├─────────────────────────────────────────────┤
│  ## 📑 Table of Contents                    │
│  (manual, with emoji-prefixed anchors)      │
├─────────────────────────────────────────────┤
│  ## 🚪 What is ThreeDoors?                  │
│  (elevator pitch — problem + solution)      │
│  Keep existing content, lightly edit        │
├─────────────────────────────────────────────┤
│  ## 📸 Screenshots                          │
│  (hero screenshots + foldable gallery)      │
├─────────────────────────────────────────────┤
│  ## 🚀 Quick Start                          │
│  (condensed: brew install + 5 commands)     │
│                                             │
│  > [!TIP] See full installation options...  │
├─────────────────────────────────────────────┤
│  ## ✨ Features                             │
│  (comprehensive, organized by category)     │
│  (11 categories, ~60 features)              │
│  EXPANDED — this is the selling point       │
├─────────────────────────────────────────────┤
│  <details> 📦 Installation Options          │
│    (4 options, full details)                │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  <details> ⌨️ Key Bindings                  │
│    (3 view tables + command palette)        │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  <details> 💻 CLI Reference                 │
│    (full command reference)                 │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  <details> 🤖 MCP Server                    │
│    (setup, 15 tools, Claude Desktop config) │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  <details> 🔧 Configuration                 │
│    (full config.yaml reference)             │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  <details> 📁 Data Directory                │
│    (~/.threedoors/ structure)               │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  ## 🔒 Data & Privacy                       │
│  (short — 5 bullet points, EXPANDED)        │
├─────────────────────────────────────────────┤
│  ## 🧭 Philosophy                           │
│  (6 principles, EXPANDED)                   │
├─────────────────────────────────────────────┤
│  <details> 🛠️ Development                   │
│    (tech stack, project structure, make)     │
│  </details>                                 │
├─────────────────────────────────────────────┤
│  ## 🤝 Contributing                         │
│  (short version, EXPANDED)                  │
├─────────────────────────────────────────────┤
│  ## 📚 Documentation                        │
│  (links to docs, EXPANDED)                  │
├─────────────────────────────────────────────┤
│  ## 📜 License                              │
│  (one line + acknowledgments)               │
├─────────────────────────────────────────────┤
│  ## 🔗 Links                                │
│  (repo, issues, releases)                   │
├─────────────────────────────────────────────┤
│  <p align="center">                         │
│    <strong>"Progress over perfection.       │
│    Three doors. One choice.                 │
│    Move forward." 🚪✨</strong>             │
│  </p>                                       │
│                                             │
│  <div align="right">                        │
│    <a href="#threedoors-">↑ Back to top</a> │
│  </div>                                     │
└─────────────────────────────────────────────┘
```

### Key Structural Changes from Current
1. **Badges**: 3 → 10, centered
2. **TOC**: Added (currently none)
3. **Screenshots**: New section with placeholders
4. **Features**: Reorganized into 11 categories (from 8), ~15 missing features added
5. **6 sections folded**: Installation, Key Bindings, CLI Reference, MCP, Configuration, Development
6. **Quick Start**: Condensed (currently duplicates Installation)
7. **Configuration**: Extracted from "User Guide" into own foldable section
8. **Back-to-top links**: Added after each major section
9. **Admonitions**: Use `> [!TIP]` and `> [!NOTE]` where appropriate

### Estimated Impact
- **Visible length**: ~300 lines (from 747) — rest is foldable
- **Total length**: ~850 lines (added content for missing features, TOC, screenshots)
- **Scan time**: Dramatically reduced — TOC + foldables mean users find what they need fast

---

## 11. Follow-Up Stories (Opportunities)

These are NOT in scope for the README overhaul but worth noting:

1. **Screenshot capture story** — Use `charmbracelet/vhs` to create terminal GIFs for the README
2. **Codecov integration story** — Add test coverage badge (requires CI changes)
3. **GitHub wiki** — Move detailed reference docs (CLI, MCP, config) to wiki, link from README
4. **Contribution guidelines** — `CONTRIBUTING.md` file with detailed dev setup (de-clutter README)
5. **Changelog** — `CHANGELOG.md` or auto-generated from git tags

---

## 12. Implementation Notes

- The overhaul is content reorganization + additions — no code changes
- All existing content is preserved, just restructured
- Foldable sections use standard GitHub-supported `<details>` HTML
- Emojis are all standard Unicode — no custom images needed
- Badge URLs need verification after implementation (some may 404 if repo is private)
- The TOC anchor links must match the actual rendered heading IDs (GitHub auto-generates these from heading text, including emoji)
