# BMAD Sidecar Installation Architecture — Research Report

**Date:** 2026-03-30
**Researcher:** jolly-eagle (worker agent)
**Status:** Complete

---

## 1. BMAD's Directory Structure — What Exists Today

A full BMAD installation in this repo (ThreeDoors) creates **four distinct filesystem layers**:

### Layer 1: `.bmad-core/` (75 files) — IDE-Facing Shared Assets
The "classic" BMAD install. Contains:
- `core-config.yaml` — per-project configuration (story locations, PRD paths, architecture paths)
- `agents/` (10 files) — agent persona definitions (PM, architect, QA, dev, etc.)
- `tasks/` (23 files) — executable task definitions
- `templates/` (13 files) — document templates (PRD, story, architecture, etc.)
- `checklists/` (6 files) — validation checklists
- `data/` (6 files) — reference data (elicitation methods, brainstorming techniques)
- `workflows/` (6 files) — workflow definitions (greenfield/brownfield variants)
- `utils/` (2 files) — utility docs
- `install-manifest.yaml` — tracks file hashes for upgrade detection

**Version:** 4.44.3 (installed 2025-11-11). Referenced by `BMad/tasks/*` and `BMad/agents/*` commands via `.bmad-core/{type}/{name}` paths.

### Layer 2: `_bmad/` (501 files) — Module System (v6)
The newer modular BMAD install. Contains:
- `core/` — core module: agents, tasks, workflows (party-mode, brainstorming, advanced-elicitation)
- `bmm/` — Business Method Module: full workflow pipeline (analysis, planning, solutioning, implementation)
- `cis/` — Creative & Innovation Sidecar: storytelling, design thinking, innovation strategy, problem solving
- `tea/` — Test Engineering & Architecture: testarch workflows (trace, framework, CI, ATDD, etc.)
- `_config/` (25 files) — agent customization configs, IDE config, workflow/task manifests
- `_memory/` (4 files) — agent memory (per-agent preferences, documentation standards)

**Version:** 6.0.4 (installed 2026-03-02). Referenced by `bmad-*` commands via `{project-root}/_bmad/` paths.

### Layer 3: `.claude/commands/` (98 BMAD-related files) — Claude Code Skill Stubs
Two sub-patterns:
- **Top-level stubs** (65 files): `bmad-*.md` — each is a thin wrapper that loads a workflow/task from `_bmad/`
- **BMad/ subdirectory** (33 files): `BMad/agents/*.md` and `BMad/tasks/*.md` — self-contained agent personas and task runners that reference `.bmad-core/`

### Layer 4: `_bmad-output/` (243 files) — Per-Project Mutable Output
All generated artifacts: planning-artifacts, implementation-artifacts, research-reports. This is the ONLY directory that is truly per-project mutable state.

### Summary: Shared vs Per-Project

| Component | Shared (read-only) | Per-Project (mutable) |
|-----------|:------------------:|:---------------------:|
| `.bmad-core/agents/`, `tasks/`, `templates/`, `checklists/`, `data/`, `workflows/`, `utils/` | Yes | No |
| `.bmad-core/core-config.yaml` | No | Yes |
| `.bmad-core/install-manifest.yaml` | No | Yes |
| `_bmad/core/`, `_bmad/bmm/`, `_bmad/cis/`, `_bmad/tea/` (agents, workflows, tasks) | Yes | No |
| `_bmad/_config/` (customization YAMLs) | No | Yes |
| `_bmad/_memory/` | No | Yes |
| `_bmad/core/config.yaml` | No | Yes |
| `.claude/commands/bmad-*.md` | Yes | No |
| `.claude/commands/BMad/` | Yes | No |
| `_bmad-output/` | No | Yes |

---

## 2. Claude Code's Discovery Mechanisms

### Skill/Command Discovery (precedence order)
1. **Managed policy** (enterprise, highest priority)
2. **User-level:** `~/.claude/skills/<name>/` (personal, applies to all projects)
3. **Project-level:** `.claude/skills/<name>/` (checked in, applies to all team members)
4. **Plugin-level:** bundled with installed plugins (lowest priority)
5. **Legacy commands:** `.claude/commands/<name>.md` — still works, same precedence rules

**Key insight:** User-level skills at `~/.claude/skills/` are discovered globally for ALL projects. This is the primary vector for a sidecar approach.

### Rules Discovery
- `~/.claude/rules/*.md` — user-level, loaded for all projects
- `.claude/rules/*.md` — project-level
- **Symlinks are explicitly supported** in `.claude/rules/`
- Recursive subdirectory discovery

### CLAUDE.md Loading
- `./CLAUDE.md` and `./.claude/CLAUDE.md` (project)
- `~/.claude/CLAUDE.md` (user)
- **`@import` syntax** can pull in files from anywhere: `@~/.bmad/some-file.md`
- `--add-dir` with `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1` loads CLAUDE.md from external dirs
- Skills from `--add-dir` directories are auto-discovered

### Settings Scope
- User (`~/.claude/settings.json`) — global, **cannot reference project paths**
- Project (`.claude/settings.json`) — team-shared
- Local (`.claude/settings.local.json`) — personal overrides
- **Hooks** can use `$CLAUDE_PROJECT_DIR` for project-relative paths at runtime

### MCP Servers
- Configured at user (`~/.claude/mcp.json`) or project (`.claude/mcp.json`) level
- **Cannot expose skills/commands** — only tools and prompts
- Could theoretically serve BMAD reference data as tools, but not workflows

---

## 3. Evaluation of Techniques

### Technique A: User-Level Skills + Symlinks (RECOMMENDED)

**Approach:** Install BMAD's shared read-only assets once at `~/.bmad/` (or `~/.claude/skills/bmad/`). Use symlinks to make per-project customization work.

```
~/.bmad/                          # Single BMAD install (shared)
  core/                           # .bmad-core equivalent
  modules/                        # _bmad modules (bmm, cis, tea, core)
  commands/                       # .claude/commands/bmad-*.md stubs

~/.claude/skills/bmad-*/          # Symlinks or copies pointing to ~/.bmad/commands/
  OR
~/.claude/commands/bmad-*.md      # Command stubs at user level

Per-project:
  .bmad-core/core-config.yaml     # Project-specific config (required)
  _bmad/_config/                  # Project-specific customizations
  _bmad/_memory/                  # Project-specific agent memory
  _bmad-output/                   # Project-specific output
```

**How it works:**
1. BMAD command stubs in `~/.claude/commands/` are discovered globally by Claude Code
2. Command stubs reference `{project-root}/_bmad/...` — these paths resolve at runtime to the current project
3. The `_bmad/` directory in each project is mostly symlinks to `~/.bmad/modules/`, with per-project overrides for config/memory
4. `.bmad-core/` in each project is a symlink to `~/.bmad/core/`, except `core-config.yaml` which is a real file

**Pros:**
- Single upgrade point: update `~/.bmad/`, all projects see new version
- Claude Code's native discovery works — no hacks needed
- `{project-root}` resolution handles multi-project seamlessly
- Per-project data stays isolated naturally
- No special shell wrappers or filesystem tricks needed
- Works with git worktrees (symlinks resolve correctly)

**Cons:**
- Symlink management is manual (needs a setup script per project)
- Git sees symlinks as symlinks — team members need the same `~/.bmad/` layout or the links break
- `.bmad-core/core-config.yaml` must remain a real file per-project
- Two version layers (v4 `.bmad-core` + v6 `_bmad`) complicate the shared layout

**Complexity:** Low-Medium
**Fragility:** Low (symlinks are robust on macOS/Linux)

---

### Technique B: Directory Symlinks (Simpler Variant)

**Approach:** Instead of individual file symlinks, symlink entire directories.

```
~/.bmad/                          # Canonical BMAD install
  bmad-core/                      # All .bmad-core read-only files
  modules/                        # All _bmad module files (core, bmm, cis, tea)

Per-project:
  .bmad-core -> ~/.bmad/bmad-core   # Directory symlink (or selective)
  _bmad/core -> ~/.bmad/modules/core
  _bmad/bmm -> ~/.bmad/modules/bmm
  _bmad/cis -> ~/.bmad/modules/cis
  _bmad/tea -> ~/.bmad/modules/tea
  _bmad/_config/                    # Real dir, per-project
  _bmad/_memory/                    # Real dir, per-project
  _bmad/core/config.yaml            # PROBLEM: can't override single file in symlinked dir
```

**Pros:**
- Very simple to set up
- Fewer symlinks to manage
- Clear visual distinction (symlinked dirs vs real dirs)

**Cons:**
- **Critical issue:** Can't override single files within a symlinked directory (e.g., `core-config.yaml` inside `.bmad-core/`). Would need to either:
  - Not symlink `.bmad-core/` and instead symlink each subdirectory
  - Use a wrapper that copies + patches the config
- Git treats directory symlinks as files — `git status` shows them as symlinks, not directories
- Some tools (IDE file watchers, linters) may not follow symlinks into directories

**Complexity:** Low
**Fragility:** Medium (the override problem is real)

---

### Technique C: `--add-dir` with Shell Wrapper

**Approach:** A shell wrapper (e.g., `bmad-claude`) that launches Claude Code with `--add-dir ~/.bmad`.

```bash
#!/bin/bash
# ~/bin/bmad-claude
CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1 \
  claude --add-dir ~/.bmad "$@"
```

```
~/.bmad/
  .claude/commands/bmad-*.md      # Commands discovered via --add-dir
  .claude/skills/bmad-*/          # Skills discovered via --add-dir
  CLAUDE.md                       # Optional: BMAD-specific instructions
  modules/                        # BMAD module files
```

**Pros:**
- Clean separation — BMAD lives entirely outside the project
- No symlinks in the project at all
- Skills from `--add-dir` are auto-discovered
- With the env var, CLAUDE.md from `~/.bmad/` is also loaded

**Cons:**
- **Must always use the wrapper** — forgetting `--add-dir` means no BMAD
- `{project-root}` in command stubs resolves to the PROJECT, not `~/.bmad/` — but BMAD workflows reference `{project-root}/_bmad/` which won't exist without symlinks
- Doesn't solve the `.bmad-core` runtime reference problem (agent tasks do `.bmad-core/type/name`)
- Multiclaude agents wouldn't use the wrapper — need additional configuration
- IDE integrations (VS Code, JetBrains) would need custom launch configs

**Complexity:** Medium
**Fragility:** High (easy to forget the wrapper; breaks IDE integrations)

---

### Technique D: Git Submodule

**Approach:** BMAD is a git submodule pinned to a release tag.

```
Per-project:
  _bmad-shared/ -> git submodule (https://github.com/bmad-method/bmad)
  .bmad-core -> _bmad-shared/bmad-core (symlink)
  _bmad/core -> _bmad-shared/modules/core (symlinks per module)
  _bmad/_config/                  # Per-project
  _bmad/_memory/                  # Per-project
```

**Pros:**
- Version pinning per project (can have different BMAD versions per repo)
- Standard git workflow for upgrades (`git submodule update`)
- Team members get BMAD automatically on clone

**Cons:**
- **Not actually shared** — each repo gets its own copy of BMAD
- Submodule management is notoriously painful (detached HEAD, forgotten updates)
- Still needs symlinks for `.bmad-core/` and `_bmad/` resolution
- Doesn't reduce disk usage or provide single-upgrade-point benefit
- Adds git complexity for every contributor

**Complexity:** High
**Fragility:** High

---

### Technique E: Overlay Filesystem (bindfs/overlayfs)

**Approach:** Mount `~/.bmad/` as an overlay on the project directory so BMAD files appear to be in-project.

```bash
# Mount shared BMAD as lower layer, project as upper layer
# Files from ~/.bmad/ appear alongside project files
bindfs --map=readonly ~/.bmad/bmad-core .bmad-core
```

**Pros:**
- Completely transparent to Claude Code — files appear native
- Per-project files (configs, memory) override shared files naturally with overlayfs
- No symlinks in git

**Cons:**
- **macOS support is poor** — overlayfs is Linux-only; bindfs requires macFUSE (third-party kernel extension)
- Security implications of FUSE/kernel extensions
- Must mount before every session, unmount after
- Fragile across system updates (macFUSE breaks on macOS upgrades regularly)
- Invisible to git — can't tell what's real vs mounted
- Breaks with git worktrees (each worktree needs its own mount)

**Complexity:** Very High
**Fragility:** Very High (especially on macOS)

---

### Technique F: BMAD as MCP Server

**Approach:** Build an MCP server that exposes BMAD workflows, templates, and agent personas as tools.

**Pros:**
- Ultimate clean separation — no files in project at all
- Could dynamically adapt to project context
- Single install via `~/.claude/mcp.json`

**Cons:**
- **MCP cannot expose skills/commands** — only tools
- Would require rewriting all BMAD workflows as MCP tool implementations
- Massive engineering effort for marginal benefit
- Loses the "read the markdown file" simplicity that makes BMAD work
- BMAD's design is fundamentally file-based — fighting the architecture

**Complexity:** Extreme
**Fragility:** Medium (once built, but building it is huge)

---

### Technique G: User-Level Commands Only (Minimal Sidecar)

**Approach:** Only share the `.claude/commands/` stubs at user level. Keep ALL BMAD runtime files per-project.

```
~/.claude/commands/bmad-*.md      # Thin stubs, shared globally
~/.claude/commands/BMad/          # Agent/task commands, shared globally

Per-project (unchanged):
  .bmad-core/                     # Full copy
  _bmad/                          # Full copy
  _bmad-output/                   # Per-project output
```

**Pros:**
- Simplest possible approach
- No symlinks at all
- BMAD installer still works per-project for `_bmad/` and `.bmad-core/`
- Stubs are identical across projects — perfect for sharing
- Upgrade stubs once at `~/.claude/commands/`, upgrade runtime per-project via installer

**Cons:**
- Doesn't achieve single-install for runtime files
- Still need to run BMAD installer per-project for `_bmad/` and `.bmad-core/`
- Only saves maintaining the command stubs (98 files)

**Complexity:** Very Low
**Fragility:** Very Low

---

## 4. Recommended Architecture

### Primary Recommendation: Hybrid Approach (Technique A + G)

**Phase 1 (immediate): User-Level Commands (Technique G)**

Move all `.claude/commands/bmad-*.md` and `.claude/commands/BMad/` to `~/.claude/commands/`. This gives global BMAD skill availability with zero symlinks or complexity.

```bash
# One-time setup
mkdir -p ~/.claude/commands/BMad/agents ~/.claude/commands/BMad/tasks
cp .claude/commands/bmad-*.md ~/.claude/commands/
cp -r .claude/commands/BMad/* ~/.claude/commands/BMad/
```

**Result:** All BMAD slash commands available in every project. Runtime files (`.bmad-core/`, `_bmad/`) still installed per-project.

**Phase 2 (optional): Shared Runtime via Symlinks (Technique A)**

For power users managing many repos:

```bash
# One-time setup: create canonical BMAD install
mkdir -p ~/.bmad
cp -r .bmad-core ~/.bmad/bmad-core
cp -r _bmad/core ~/.bmad/modules-core
cp -r _bmad/bmm ~/.bmad/modules-bmm
cp -r _bmad/cis ~/.bmad/modules-cis
cp -r _bmad/tea ~/.bmad/modules-tea

# Per-project setup script (~/.bmad/setup-project.sh):
#!/bin/bash
PROJECT_DIR="${1:-.}"

# Symlink read-only module directories
for mod in core bmm cis tea; do
  ln -sfn ~/.bmad/modules-$mod "$PROJECT_DIR/_bmad/$mod"
done

# Symlink .bmad-core subdirectories (not the config)
for dir in agents tasks templates checklists data workflows utils; do
  ln -sfn ~/.bmad/bmad-core/$dir "$PROJECT_DIR/.bmad-core/$dir"
done

# Ensure per-project directories exist
mkdir -p "$PROJECT_DIR/_bmad/_config" "$PROJECT_DIR/_bmad/_memory" "$PROJECT_DIR/_bmad-output"

echo "BMAD sidecar linked for $PROJECT_DIR"
```

**Result:** Runtime files shared via symlinks, per-project config/memory/output isolated.

### Setup for Multiclaude Compatibility

Multiclaude creates worktrees in `/Users/skippy/.multiclaude/wts/ThreeDoors/<worker>/`. Since worktrees share the `.git` directory with the main checkout, symlinks committed to git will resolve correctly in worktrees IF they use absolute paths. Relative symlinks may break.

**Recommendation:** Use absolute symlinks pointing to `~/.bmad/` (which expands to `/Users/skippy/.bmad/`). These work in all worktrees.

---

## 5. Upgrade Workflow

### Phase 1 (Commands Only)
```bash
# When BMAD releases a new version:
# 1. Install in any project to get the new command stubs
# 2. Copy the updated stubs to ~/.claude/commands/
cp .claude/commands/bmad-*.md ~/.claude/commands/
cp -r .claude/commands/BMad/* ~/.claude/commands/BMad/
# Done — all projects now have new commands
```

### Phase 2 (Full Sidecar)
```bash
# When BMAD releases a new version:
# 1. Install in a temp directory
mkdir /tmp/bmad-upgrade && cd /tmp/bmad-upgrade
# (run BMAD installer here)

# 2. Update canonical install
cp -r .bmad-core/* ~/.bmad/bmad-core/
cp -r _bmad/core/* ~/.bmad/modules-core/
cp -r _bmad/bmm/* ~/.bmad/modules-bmm/
cp -r _bmad/cis/* ~/.bmad/modules-cis/
cp -r _bmad/tea/* ~/.bmad/modules-tea/

# 3. Update command stubs
cp .claude/commands/bmad-*.md ~/.claude/commands/
cp -r .claude/commands/BMad/* ~/.claude/commands/BMad/

# 4. All projects instantly see new version (symlinks resolve to updated files)
# 5. Per-project configs may need manual update if schema changed
```

---

## 6. Edge Cases and Failure Modes

### Symlink Issues
- **macOS Gatekeeper:** Some macOS security features may flag symlinked executables. BMAD files are markdown/YAML, not executables, so this shouldn't apply.
- **Git visibility:** `git status` shows symlinks as changes if they differ from committed state. Add symlinks to `.gitignore` if not committing them, or commit them for team sharing.
- **IDE indexing:** VS Code and JetBrains follow symlinks by default. No issues expected.

### Version Skew
- **Risk:** Shared runtime is v6.0.4 but a project needs v5 for compatibility.
- **Mitigation:** Don't symlink for that project — fall back to per-project install. The user-level commands (Phase 1) still work because stubs just point to `{project-root}/_bmad/` which resolves per-project.

### Multiclaude Worktrees
- **Risk:** Relative symlinks break in worktrees at different filesystem paths.
- **Mitigation:** Always use absolute symlinks (`/Users/skippy/.bmad/...`). Worktrees are at different paths but absolute symlinks resolve to the same canonical location.

### `{project-root}` Resolution
- BMAD commands use `{project-root}` which Claude Code resolves to the current working directory.
- When commands are at `~/.claude/commands/`, `{project-root}` still resolves to the PROJECT directory, not `~/.claude/`. This is correct behavior.
- **Risk:** If a project doesn't have `_bmad/` (no BMAD installed), the path resolves to nothing. Claude will report a file-not-found error.
- **Mitigation:** The setup script ensures the symlinks exist before work begins.

### `.bmad-core` Path References
- Some BMAD tasks reference `.bmad-core/type/name` (relative to project root).
- These continue to work as long as `.bmad-core/` exists in the project (whether as real files or symlinks).

### Concurrent Access
- Multiple Claude instances reading the same shared `~/.bmad/` files simultaneously is safe — they're read-only.
- Per-project `_bmad-output/` is write-isolated naturally.

### BMAD Installer Conflicts
- Running the BMAD installer on a project with symlinks will overwrite symlinks with real files.
- **Mitigation:** Re-run the setup script after any BMAD installer run. Or modify the installer to detect and preserve symlinks.

---

## 7. Decision Matrix

| Criterion | G (Commands Only) | A+G (Hybrid) | C (Wrapper) | D (Submodule) | E (Overlay) | F (MCP) |
|-----------|:-:|:-:|:-:|:-:|:-:|:-:|
| Single upgrade point | Partial | Yes | Yes | No | Yes | Yes |
| No per-project install | No | Mostly | No | No | Yes | Yes |
| Works with multiclaude | Yes | Yes | No | Yes | Maybe | Yes |
| Works with IDE | Yes | Yes | No | Yes | Maybe | N/A |
| Setup complexity | Very Low | Low | Medium | High | Very High | Extreme |
| Fragility | Very Low | Low | High | High | Very High | Medium |
| Team sharing | Easy | Needs script | Hard | Built-in | Impossible | Easy |

### Verdict

**Start with Phase 1 (Technique G)** — it's 15 minutes of work and immediately gives you global BMAD commands across all projects. This is the 80/20 solution.

**Move to Phase 2 (Technique A + G)** only if you're managing 3+ repos with BMAD and the per-project install overhead becomes painful. The symlink approach adds some maintenance burden but eliminates duplicate runtime files.

**Avoid** overlay filesystems, MCP rewrites, and `--add-dir` wrappers — they fight the tooling rather than working with it.
