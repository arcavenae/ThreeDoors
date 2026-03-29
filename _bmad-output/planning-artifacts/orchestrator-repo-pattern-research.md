# Orchestrator Repo Pattern & Multi-Repo multiclaude Enhancement Architecture

**Date:** 2026-03-29
**Researcher:** happy-fox (worker)
**Scope:** How to adapt multiclaude-enhancements for orchestrator repos (aae-orc pattern) that coordinate multiple sub-projects

---

## Executive Summary

Orchestrator repos like aae-orc represent a fundamentally different pattern from single-project repos like ThreeDoors. They coordinate work across multiple sub-repos, mix human-driven and AI-driven work, and need BMAD tooling available without tight coupling. This research covers six key topics: BMAD packaging, agent awareness, DFCP enforcement, worktree+submodule interactions, enhancement profiles, and agent definition adaptations.

**Critical finding:** multiclaude's worktree creation has zero submodule handling — workers in orchestrator repos would get empty submodule directories and fail to find BMAD skills. This is the #1 blocker for orchestrator repo support.

**Key recommendation:** A profile-based installation system for multiclaude-enhancements with "single-project" and "orchestrator" profiles, combined with a post-worktree hook for submodule initialization.

---

## Table of Contents

1. [BMAD as Optional Submodule](#1-bmad-as-optional-submodule)
2. [Agent Awareness of Orchestrator Pattern](#2-agent-awareness-of-orchestrator-pattern)
3. [Human vs AI Work Signals (DFCP in Practice)](#3-human-vs-ai-work-signals-dfcp-in-practice)
4. [Worktree + Submodule Interaction](#4-worktree--submodule-interaction)
5. [multiclaude-enhancements Packaging](#5-multiclaude-enhancements-packaging)
6. [Agent Definition Adaptations](#6-agent-definition-adaptations)
7. [Decisions Summary](#7-decisions-summary)
8. [Open Questions](#8-open-questions)

---

## 1. BMAD as Optional Submodule

### The Problem

ThreeDoors has BMAD installed directly: files physically copied into `.claude/commands/` at install time. This works for a single project but doesn't scale to orchestrator repos where:
- BMAD versioning matters across sub-repos
- Not every sub-repo needs BMAD (some are infrastructure, some are dark factory targets)
- The orchestrator itself may want BMAD for planning but sub-repos have their own governance

### Options Evaluated

#### Option A: Git Submodule Pointing to BMAD Repo ⭐ RECOMMENDED

```
aae-orc/
├── .claude/commands/          # Symlinked or generated from submodule
├── vendor/bmad/               # git submodule → bmad-method repo
│   ├── .claude/commands/      # BMAD commands (source)
│   ├── agents/
│   └── templates/
├── repos/
│   ├── threedoors/            # submodule → arcavenae/ThreeDoors
│   └── dark-factory-alpha/    # submodule → arcavenae/threedoors-df-alpha-001
└── CLAUDE.md
```

**Pros:**
- Version-pinned BMAD (reproducible)
- Centralized update: `git submodule update --remote vendor/bmad`
- Sub-repos don't need BMAD installed — orchestrator provides it
- Clean separation: BMAD is tooling, not project code

**Cons:**
- Claude Code does NOT auto-discover commands in submodule paths (confirmed via research)
- Requires a bridge mechanism (see "Command Discovery Bridge" below)
- `git worktree add` does NOT auto-init submodules (confirmed)

#### Option B: Install BMAD Directly in Orchestrator Repo

```
aae-orc/
├── .claude/commands/          # BMAD commands copied here (current ThreeDoors pattern)
├── _bmad/                     # BMAD config/templates
└── repos/
    └── ...submodules...
```

**Pros:** Simple, proven (ThreeDoors does this today), commands auto-discovered
**Cons:** Version coupling, manual update process, no clean separation

**Rejected** — couples BMAD version to the orchestrator repo, doesn't leverage the composability that submodules provide.

#### Option C: Dedicated BMAD Repo as Submodule

Same as Option A but the submodule points to a thin repo that only contains BMAD configuration for this specific orchestrator (customized templates, agent definitions, etc.), not the upstream BMAD method repo itself.

**Rejected** — adds a third repo to manage without clear benefit over Option A. The upstream BMAD repo already supports customization.

### Command Discovery Bridge (Required for Option A)

Claude Code only discovers commands in `<project-root>/.claude/commands/` and `~/.claude/commands/`. It does NOT follow symlinks into submodules (confirmed). Three bridge options:

#### Bridge 1: Post-Install Symlink Script ⭐ RECOMMENDED

```bash
#!/bin/bash
# scripts/link-bmad-commands.sh
# Run after: git submodule update --init
ln -sfn ../../vendor/bmad/.claude/commands/BMad .claude/commands/BMad
```

**Why:** Symlinks within `.claude/commands/` are likely followed by Claude Code's file scanner (it uses standard filesystem traversal), even though cross-directory symlinks aren't explicitly documented. The symlink target is still under the project root. Needs empirical verification (see OQ-ORC-1).

**Fallback if symlinks fail:** Physical copy with a `just update-bmad` task.

#### Bridge 2: Claude Code Plugin Wrapping

Package BMAD as a Claude Code plugin with `plugin.json` specifying `commands` paths. Plugins support extra command directories.

**Rejected for now** — plugins are still maturing; adds a dependency on Claude Code plugin infrastructure that may change.

#### Bridge 3: Physical Copy via Makefile/Justfile

```bash
# just update-bmad
rsync -a vendor/bmad/.claude/commands/BMad/ .claude/commands/BMad/
```

**Acceptable fallback** — proven pattern (ThreeDoors does this today), but loses the version-tracking benefits of symlinks.

### Decision: ORC-D-001

**Adopted:** Option A (submodule) + Bridge 1 (symlink script), with Bridge 3 as fallback.
**Rejected:** Option B (direct install — couples versions), Option C (extra repo — unnecessary indirection), Bridge 2 (plugin — immature infrastructure).

---

## 2. Agent Awareness of Orchestrator Pattern

### The CLAUDE.md Hierarchy Problem

Orchestrator repos have a CLAUDE.md at the root, but each sub-repo (submodule) also has its own CLAUDE.md. When an agent works in the orchestrator:

```
aae-orc/
├── CLAUDE.md                   # "I'm an orchestrator, here are my sub-repos"
├── repos/
│   ├── threedoors/
│   │   └── CLAUDE.md           # "I'm ThreeDoors, use Bubbletea, run just fmt..."
│   └── dark-factory/
│       └── CLAUDE.md           # "I'm a dark factory, relaxed governance..."
```

Claude Code reads the project-root CLAUDE.md automatically. When an agent `cd`s into a submodule directory, it does NOT automatically pick up that submodule's CLAUDE.md — it's still using the root's.

### Recommendation: Orchestrator CLAUDE.md as Navigation Layer

The orchestrator's CLAUDE.md should:

1. **Declare the orchestrator pattern** explicitly:
   ```markdown
   ## Project Type: Orchestrator
   This repo coordinates multiple sub-projects. Each sub-repo has its own
   CLAUDE.md with project-specific instructions. When working in a sub-repo,
   READ that sub-repo's CLAUDE.md first.
   ```

2. **Provide a manifest of sub-repos** with their roles:
   ```markdown
   ## Sub-Repo Manifest
   | Path | Repo | Type | Autonomy | Notes |
   |------|------|------|----------|-------|
   | repos/threedoors | arcavenae/ThreeDoors | Golden | L1-L2 | Human-governed |
   | repos/df-alpha | arcavenae/threedoors-df-alpha-001 | Dark Factory | L3-L4 | AI-autonomous |
   | infra/ci | — | Local | L2 | Shared CI templates |
   ```

3. **Include sub-repo CLAUDE.md loading instructions** as a standard practice:
   ```markdown
   ## Before Working in a Sub-Repo
   Always `Read` the sub-repo's CLAUDE.md before making changes there.
   The root CLAUDE.md provides orchestrator-level context; sub-repo CLAUDE.md
   provides project-specific rules that OVERRIDE general guidance.
   ```

### Decision: ORC-D-002

**Adopted:** Orchestrator CLAUDE.md as explicit navigation layer with manifest table and sub-repo loading protocol.
**Rejected:** Automatic CLAUDE.md inheritance (Claude Code doesn't support this), single monolithic CLAUDE.md covering all sub-repos (unmaintainable, violates separation of concerns).

---

## 3. Human vs AI Work Signals (DFCP in Practice)

### Building on Prior Research

The Dark Factory research (R-003) established the L0-L4 autonomy spectrum. The DFCP research (R-005) defined permission matrices and gate taxonomies. This section applies those patterns to the orchestrator context.

### repos.yaml Manifest ⭐ RECOMMENDED

```yaml
# aae-orc/repos.yaml
repos:
  - name: threedoors
    path: repos/threedoors
    github: arcavenae/ThreeDoors
    autonomy: L2          # AI-supervised, human approves gates
    governance: golden     # Full CODEOWNERS, branch protection, human review
    dfcp_profile: golden-repo
    codeowners:
      - path: "**"
        owners: ["@skippy"]

  - name: threedoors-df-alpha
    path: repos/df-alpha
    github: arcavenae/threedoors-df-alpha-001
    autonomy: L3          # AI-autonomous, human reviews output only
    governance: dark-factory
    dfcp_profile: dark-factory
    auto_merge: true
    disposal_policy: after-gallery-review

  - name: multiclaude-enhancements
    path: repos/mc-enhance
    github: arcavenae/multiclaude-enhancements
    autonomy: L1          # AI-assisted, human directs
    governance: golden
    dfcp_profile: golden-repo

  - name: infra
    path: infra/
    autonomy: L0          # Human-crafted only
    governance: locked     # No AI changes without explicit approval
    dfcp_profile: human-only
```

### Enforcement Layers

| Layer | Mechanism | What It Blocks | Where Defined |
|-------|-----------|---------------|---------------|
| 1. CLAUDE.md | Prompt instruction | Agent self-regulation | Orchestrator + sub-repo CLAUDE.md |
| 2. repos.yaml | Agent reads manifest | Agent knows boundaries | Orchestrator root |
| 3. CODEOWNERS | GitHub review gate | PRs to human-gated paths | Per-repo `.github/CODEOWNERS` |
| 4. Branch protection | GitHub enforcement | Direct push, unreviewed merge | Per-repo rulesets |
| 5. Claude Code hooks | Pre-tool enforcement | File writes to wrong paths | `.claude/settings.json` |

**Layer 5 is the key insight from the dollspace-gay research (R-010):** "Enforce via tooling, not via prompts." A PreToolUse hook can mechanically prevent agents from writing to `L0` (human-only) paths:

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Edit|Write",
      "hooks": [{
        "type": "command",
        "command": "python3 scripts/check-dfcp-boundary.py $FILE_PATH"
      }]
    }]
  }
}
```

The script reads `repos.yaml`, checks if the target file is in a human-only or golden path, and exits non-zero to block the write.

### Brownian Ratchet Principle

AI agents should be **free by default** in dark factory repos and **blocked by default** in human-gated repos. The ratchet works one way: promoting code from dark factory → golden repo requires explicit human action (the re-entry protocol from R-003). There is no mechanism for golden repo governance to leak into dark factory repos.

### Decision: ORC-D-003

**Adopted:** repos.yaml manifest + 5-layer enforcement (CLAUDE.md → manifest → CODEOWNERS → branch protection → hooks).
**Rejected:** Single CODEOWNERS for entire orchestrator (can't scope to sub-repos), CI-only enforcement (too late — catches at PR, not at write), label-based gating (labels are advisory, not enforceable at write time).

---

## 4. Worktree + Submodule Interaction

### The Blocker: multiclaude Has Zero Submodule Handling

**Confirmed via source code analysis** of `/Users/skippy/work/multiclaude/internal/worktree/worktree.go` and `/Users/skippy/work/multiclaude/internal/daemon/daemon.go`:

```go
// daemon.go ~line 1600 — this is ALL that happens:
func (m *Manager) CreateNewBranch(path, newBranch, startPoint string) error {
    _, err := m.runGit("worktree", "add", "-b", newBranch, path, startPoint)
    return err
}
```

The string "submodule" does not appear anywhere in multiclaude's Go source. No `git submodule update --init` runs after worktree creation.

### What Happens Today

1. multiclaude creates a worktree: `git worktree add -b work/agent-name /path/to/wt HEAD`
2. New worktree has `.gitmodules` file present
3. Submodule directories exist but are **empty**
4. Agent cannot find BMAD commands, sub-repo code, or anything in submodules
5. Agent fails or operates in degraded mode

### Git's Own Warning

From `man git-worktree` BUGS section:
> "Multiple checkout in general is still experimental, and the support for submodules is incomplete. It is NOT recommended to make multiple checkouts of a superproject."

### Additional Technical Details

- Each worktree gets **independent** submodule state (separate `.git/worktrees/<name>/modules/` directory)
- `git worktree move` is **blocked** on worktrees with initialized submodules
- `git worktree remove` requires `--force` when submodules are initialized
- Object stores are **duplicated** per worktree per submodule (disk space concern for large submodules)

### Patch Options

#### Option 1: Post-Worktree Hook in Daemon ⭐ RECOMMENDED

Add to `handleSpawnAgent` in daemon.go, after worktree creation:

```go
// After worktree creation, init submodules if .gitmodules exists
gitmodulesPath := filepath.Join(worktreePath, ".gitmodules")
if _, err := os.Stat(gitmodulesPath); err == nil {
    if _, err := m.runGit("-C", worktreePath, "submodule", "update", "--init", "--recursive"); err != nil {
        log.Printf("WARN: submodule init failed in %s: %v", worktreePath, err)
        // Non-fatal — repo may work without submodules
    }
}
```

**Pros:** Mechanical, always runs, no agent cooperation needed, works for all repos
**Cons:** Slows worktree creation for repos with large submodules, requires multiclaude code change

#### Option 2: Agent Self-Heal

Agent definitions include instructions to check and init submodules:

```markdown
## Worktree Setup
If you find empty directories that should contain code (submodules),
run: `git submodule update --init --recursive`
```

**Rejected** — prompt-level instruction (unreliable per R-010 findings), duplicated in every agent definition, doesn't work if agent doesn't realize submodules are missing.

#### Option 3: Shell Hook via Claude Code Settings

```json
{
  "hooks": {
    "PostStartup": [{
      "type": "command",
      "command": "git submodule update --init --recursive 2>/dev/null || true"
    }]
  }
}
```

**Acceptable supplement** — runs when Claude Code starts in a worktree, but PostStartup isn't well-documented and may not exist as a hook event. Needs verification.

### Decision: ORC-D-004

**Adopted:** Option 1 (daemon patch) as primary fix, with Option 2 as documentation supplement.
**Rejected:** Option 2 alone (prompt-level, unreliable), Option 3 (depends on undocumented hook event).

---

## 5. multiclaude-enhancements Packaging

### Current State

multiclaude-enhancements is a single package designed for repos like ThreeDoors:
- BMAD commands in `.claude/commands/`
- Agent definitions in `agents/`
- Story templates, planning docs
- Full BMAD governance pipeline

### Profile System ⭐ RECOMMENDED

Two profiles with shared core:

```
multiclaude-enhancements/
├── core/                      # Shared across all profiles
│   ├── agents/
│   │   ├── merge-queue.md
│   │   ├── pr-shepherd.md
│   │   ├── envoy.md
│   │   └── project-watchdog.md
│   ├── .claude/settings.json  # Hook-based enforcement (shared)
│   └── scripts/               # Shared utilities
│
├── profiles/
│   ├── single-project/        # ThreeDoors-style
│   │   ├── agents/
│   │   │   └── retrospector.md
│   │   ├── CLAUDE.md.template
│   │   ├── .claude/commands/  # BMAD commands (direct install)
│   │   └── docs/templates/    # Story templates, etc.
│   │
│   └── orchestrator/          # aae-orc-style
│       ├── agents/
│       │   ├── orchestrator-supervisor.md  # Multi-repo aware
│       │   └── repo-router.md              # Routes work to sub-repos
│       ├── CLAUDE.md.template               # Orchestrator-pattern CLAUDE.md
│       ├── repos.yaml.template              # DFCP manifest template
│       ├── scripts/
│       │   ├── link-bmad-commands.sh       # Submodule → .claude/commands bridge
│       │   └── check-dfcp-boundary.py      # Hook enforcement script
│       └── docs/templates/
│
├── install.sh                 # Interactive installer
└── upgrade.sh                 # Profile-aware upgrade
```

### Installation Flow

```bash
$ ./install.sh
multiclaude-enhancements installer

Detected: git repo with submodules
Recommended profile: orchestrator

Available profiles:
  1. single-project  — Single codebase, BMAD installed directly
  2. orchestrator    — Multi-repo coordinator, BMAD as submodule

Select profile [2]: 2

Installing orchestrator profile...
  ✓ Core agent definitions → agents/
  ✓ Orchestrator agents → agents/
  ✓ CLAUDE.md template → CLAUDE.md.template (review and rename)
  ✓ repos.yaml template → repos.yaml (customize for your repos)
  ✓ DFCP boundary check → scripts/check-dfcp-boundary.py
  ✓ BMAD command bridge → scripts/link-bmad-commands.sh

Run 'scripts/link-bmad-commands.sh' after adding BMAD submodule.
```

### Upgrade Strategy

```bash
$ ./upgrade.sh
Profile: orchestrator (detected from .multiclaude-profile)
Current version: 1.2.0
Available version: 1.3.0

Changes:
  + New agent: repo-router.md
  ~ Updated: merge-queue.md (orchestrator-aware merge logic)
  ~ Updated: check-dfcp-boundary.py (new L4 support)

Apply upgrade? [y/N]:
```

The `.multiclaude-profile` marker file records which profile was installed, enabling upgrade.sh to apply the right delta.

### Decision: ORC-D-005

**Adopted:** Profile-based packaging with shared core, interactive installer, and profile-aware upgrade.
**Rejected:** Monolithic package (orchestrator features bloat single-project installs), separate repos per profile (version divergence, double maintenance), feature flags in a single config (complexity grows combinatorially).

---

## 6. Agent Definition Adaptations

### Which Agents Need Orchestrator Variants?

| Agent | Needs Variant? | Why |
|-------|---------------|-----|
| **merge-queue** | Yes | Must understand multi-repo PR scope; some PRs only touch orchestrator, some touch sub-repos |
| **pr-shepherd** | Yes | Submodule update PRs are unique — updating the submodule pointer in the orchestrator after a sub-repo PR merges |
| **project-watchdog** | Yes | Must track status across sub-repos; story files may live in different repos |
| **envoy** | Yes | Issues can span repos; triage must route to correct sub-repo's issue tracker |
| **retrospector** | Minimal | Core analytics are per-repo; cross-repo metrics are a future enhancement |
| **supervisor** | Yes | Must coordinate workers across repos; dispatch to correct worktree/submodule |

### Key Adaptations

#### merge-queue (Orchestrator Variant)

- Read `repos.yaml` to determine which sub-repos a PR touches
- For sub-repo PRs: validate CI in the sub-repo, not just orchestrator CI
- For submodule pointer updates: verify the pointed-to commit exists and CI passed in the sub-repo
- Dark factory PRs: auto-merge if CI passes (0 required approvals per DFCP)
- Golden repo PRs: require CODEOWNERS approval per normal process

#### pr-shepherd (Orchestrator Variant)

- **New responsibility:** Create "submodule bump" PRs when sub-repo main advances
  ```bash
  cd repos/threedoors && git fetch origin main
  # If sub-repo has new commits:
  git add repos/threedoors
  git commit -m "chore: bump threedoors submodule to $(git -C repos/threedoors rev-parse --short HEAD)"
  ```
- Handle rebase of PRs that touch submodule pointers (tricky — rebase can create submodule conflicts)
- Coordinate submodule update ordering when multiple sub-repos update simultaneously

#### project-watchdog (Orchestrator Variant)

- Track story status across repos by reading story files from each sub-repo's `docs/stories/`
- Cross-reference sub-repo stories with orchestrator-level epics
- Detect when sub-repo story completion should trigger orchestrator-level milestone updates
- Maintain a unified status dashboard in the orchestrator's planning docs

#### envoy (Orchestrator Variant)

- Issues filed on the orchestrator repo may need routing to sub-repos
- Read `repos.yaml` to determine which repo an issue pertains to
- Cross-link issues between orchestrator and sub-repo issue trackers
- Acknowledge on the orchestrator issue, create the tracking issue in the sub-repo

#### supervisor (Orchestrator Variant)

- Worker dispatch includes which sub-repo the worker should operate in
- Workers get `repos.yaml` context so they know their boundaries
- Cross-repo dependency tracking: "Story X in repo A is blocked by Story Y in repo B"

### Decision: ORC-D-006

**Adopted:** Variant agent definitions for 5 of 6 agents (merge-queue, pr-shepherd, project-watchdog, envoy, supervisor). Retrospector gets minimal changes.
**Rejected:** Single agent definitions with conditional logic (too complex, prompt bloat), per-repo agent instances (multiplies agents unnecessarily — one merge-queue per sub-repo is overkill).

---

## 7. Decisions Summary

| ID | Decision | Adopted | Rejected | Rationale |
|----|----------|---------|----------|-----------|
| ORC-D-001 | BMAD in orchestrator repos | Submodule + symlink bridge | Direct install (couples versions), dedicated BMAD repo (unnecessary), Claude Code plugin (immature) | Version-pinned, centralized update, clean separation; symlink or copy bridges command discovery gap |
| ORC-D-002 | Agent awareness of orchestrator pattern | Orchestrator CLAUDE.md as navigation layer with manifest table | Automatic CLAUDE.md inheritance (unsupported), monolithic CLAUDE.md (unmaintainable) | Explicit > implicit; agents instructed to read sub-repo CLAUDE.md before working there |
| ORC-D-003 | Human/AI work signal enforcement | repos.yaml manifest + 5-layer enforcement | Single CODEOWNERS (can't scope sub-repos), CI-only (catches too late), labels (advisory only) | Defense in depth; hook enforcement is mechanical per R-010 findings |
| ORC-D-004 | Worktree + submodule handling | Daemon patch to run `submodule update --init` post-worktree | Agent self-heal (unreliable prompt), shell hook (undocumented event) | Mechanical fix at the right layer; agents shouldn't need to know about submodule internals |
| ORC-D-005 | Enhancement packaging | Profile system (single-project vs orchestrator) with shared core | Monolithic (bloat), separate repos (drift), feature flags (combinatorial complexity) | Clean separation, auto-detection, profile-aware upgrades |
| ORC-D-006 | Agent adaptations | Variant definitions for 5/6 agents | Conditional logic (prompt bloat), per-repo instances (multiplied complexity) | Each agent gets orchestrator-specific knowledge; shared core behavior preserved |

---

## 8. Open Questions

| ID | Question | Context |
|----|----------|---------|
| OQ-ORC-1 | Does Claude Code's file scanner follow symlinks within `.claude/commands/`? | Critical for Bridge 1 (symlink script). If not, fall back to Bridge 3 (physical copy). Needs empirical test. |
| OQ-ORC-2 | Should the daemon submodule patch be configurable (opt-in) or always-on? | Some repos may have submodules that are intentionally not initialized (e.g., optional dev tools). A `.multiclaude-config.yaml` flag could control this. |
| OQ-ORC-3 | How should dark factory repos be created from the orchestrator? | Currently manual (create repo, add submodule, configure). Should the orchestrator have a `just create-factory` command? Ties into gallery manifest from R-003. |
| OQ-ORC-4 | Where do orchestrator-level stories live? | In the orchestrator repo's `docs/stories/`? Or distributed across sub-repos? Affects project-watchdog's tracking scope. |
| OQ-ORC-5 | Should repos.yaml support inheritance (e.g., all dark factory repos share a profile)? | Reduces repetition but adds abstraction. YAGNI for now — explicit per-repo config is clearer. |
| OQ-ORC-6 | How to handle submodule update conflicts during worktree rebase? | When daemon refreshes a worktree (rebase onto main) and main has a submodule pointer change, the rebase can conflict on the submodule. multiclaude daemon needs a conflict resolution strategy. |
| OQ-ORC-7 | Should the orchestrator profile include a "repo-router" agent that auto-dispatches work? | Beyond variant agents — a new agent type that reads repos.yaml and routes incoming issues/work to the correct sub-repo's processes. |

---

## Appendix: Prior Research Referenced

| Research | Key Findings Used |
|----------|------------------|
| R-003 (Dark Factory) | L0-L4 autonomy spectrum, separate-repo architecture, re-entry protocol, factory manifest |
| R-005 (DFCP Permissions) | Permission matrices, CODEOWNERS template, token strategy, gate taxonomy, `.dfcp.yaml` |
| R-010 (dollspace-gay) | "Enforce via tooling, not via prompts" philosophy; hook-based enforcement; session handoff |
| R-007 (Operator UX) | Workspace window separation; agent communication architecture |
| R-009 (Slack Control Plane) | Multi-machine coordination; RBAC model applicable to multi-repo |
