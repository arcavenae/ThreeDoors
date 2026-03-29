# Dark Factory Research Consolidation & Planning Roadmap

**Date:** 2026-03-29
**Type:** Consolidation Analysis
**Inputs:** 9 research artifacts (R-003 through R-011), BOARD.md entries Q-005 through Q-017, P-008 through P-012

---

## 1. Research Summary Matrix

| # | Research | Key Findings | Decisions Made | Open Questions |
|---|----------|-------------|----------------|----------------|
| R-003 | Dark Factory Core | L0-L4 autonomy spectrum; separate-repo architecture; gallery model (3-5 variants x 3-5 generations); dispose-and-rebuild with immutable specs; two-tier AI judges; Green/Yellow/Red governance | 6-phase roadmap starting with provenance tagging | 5 (OQ-1 to OQ-5: repo visibility, trigger authority, disposal preservation, budget caps, provenance mandate) |
| R-004 | Quota Throttling | Max plan has NO programmatic quota API; JSONL file heuristics only; dynamic heartbeat throttling; CLAUDE_CONFIG_DIR for multi-plan; time-of-day awareness | 4-phase: passive monitoring → adaptive heartbeats → multi-plan → full budget | 4 (QT-Q1 to QT-Q4: TOS multi-account, window reset detection, API key mode, actual token budget) |
| R-005 | DFCP Permissions | GitHub App for prod tokens, PAT for PoC; CODEOWNERS + `require_code_owner_review`; 4-tier gate taxonomy; `.dfcp.yaml` config file; CI-based scope/design checks | 7 decisions (D-DFCP-1 to D-DFCP-7) | 5 (OQ-DFCP-1 to OQ-DFCP-5: story CODEOWNERS, scope-check blocking, org rulesets, timing, agents/ gating) |
| R-006 | aclaude Personas | Layered Overlay (Option B) recommended; per-agent opt-in with immersion levels; protocol-critical agents (merge-queue, pr-shepherd, project-watchdog) NEVER get personas; manual prototype first | 4 options evaluated, Option B adopted | 5 (theme selection, comedic themes, aclaude dependency, immersion floor, persona in PR artifacts) |
| R-007 | Operator UX | Root cause: tmux paste-buffer injection; workspace window is explicitly injection-free; CronCreate heartbeats are redundant (daemon wake loop already exists); double-injection per heartbeat round-trip | 3 short-term, 4 medium-term, 3 long-term fixes | 4 (OQ-1 to OQ-4: per-agent wake interval, append-system-prompt-file, human operator agent type, wake interval too aggressive) |
| R-008 | Perplexity/Research Supervisor | Official MCP server with 4 tools; project-level config; persistent research-supervisor as soft gatekeeper; complementary with Gemini (Perplexity for web/citations, Gemini for deep research); 5-phase rollout | 7 decisions; MCP as integration path | 6 (OQ-1 to OQ-6: daily budget, cache commit policy, deep research routing, Pro subscription, MCP scope, heartbeat cron) |
| R-009 | Slack Bot Control Plane | Bot-per-machine with shared Slack app; Python (Slack Bolt) recommended; Socket Mode (no public URL); RBAC via YAML + channel gates; JSONL audit trail; "ourbot" NOT found on GitHub | Language: Python Bolt; architecture: bot-per-machine | 7 (OQ-SB-1 to OQ-SB-7: ourbot location, repo location, Socket Mode, workspace, Switchboard, tracking scope, NL commands) |
| R-010 | Chainlink/dollspace-gay Lessons | "Enforce via tooling, not via prompts" — core philosophy; hook-enforced git safety; session handoff; intervention tracking; VDD adversarial review; auto-learning memory; ACP agent protocol; blast radius limiting | 3 short-term, 4 medium-term, 3 long-term + 4 "don't adopt" | 5 (OQ-CL-1 to OQ-CL-5: hooks vs prompts, handoff storage, intervention tracking scope, adversarial review, blast radius enforcement) |
| R-011 | Orchestrator Repo Pattern | multiclaude has ZERO submodule handling (#1 blocker); BMAD as submodule + symlink bridge; repos.yaml manifest; 5-layer enforcement; profile-based enhancement packaging; variant agent definitions for 5/6 agents | 6 decisions (ORC-D-001 to ORC-D-006) | 7 (OQ-ORC-1 to OQ-ORC-7: symlink scanning, submodule opt-in, factory creation, story location, inheritance, rebase conflicts, repo-router) |

---

## 2. Natural Groupings

### Group A: Operational Foundation (Infrastructure & Agent Reliability)

**Why together:** These fix the platform you're standing on. Proceeding with dark factory or orchestrator work on an unreliable base multiplies problems.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-007 | Operator UX fixes (S-1: human→workspace, S-2: drop CronCreate heartbeats) | Zero-code, immediate value; reduces supervisor noise |
| R-010 (S-1) | Hook-enforced git safety for workers | Replaces INC-002 prompt guardrail with mechanical enforcement |
| R-010 (S-2) | Session handoff protocol for persistent agents | Fixes critical gap: agents lose all context on restart |
| R-004 (Phase 1) | Passive quota monitoring (`multiclaude quota status`) | Visibility before control; enables informed scaling decisions |
| R-007 (M-1) | Daemon-native heartbeats (replaces CronCreate) | Survives restarts, eliminates supervisor double-injection |

**Shared prerequisite:** None — this group can start immediately.

### Group B: Golden Repo Hardening & DFCP Foundation

**Why together:** CODEOWNERS, gate enforcement, and provenance tagging are all prerequisites for dark factory. They must be designed as one system.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-005 (Phase A) | CODEOWNERS + ruleset update for ThreeDoors | Immediately actionable; hardens golden repo before factory work |
| R-005 (Phase B) | CI scope-check workflow (validates story reference in PRs) | Transforms behavioral gate into technical gate |
| R-003 (Phase 0) | Provenance tagging (L0-L4 in stories, commits, PRs) | Zero-risk foundation; prerequisite for everything dark-factory |
| R-005 | `.dfcp.yaml` configuration file spec | Machine-readable profile declaration |
| R-010 (S-3) | Typed comments on story files (`[decision]`, `[observation]`, `[blocker]`) | Enriches audit trail before factory variant evaluation |

**Shared prerequisite:** Group A's operator UX fixes (working from workspace window).

### Group C: Dark Factory PoC & Gallery Infrastructure

**Why together:** The core dark factory concept — separate repos, gallery variants, dispose-and-rebuild — forms a single coherent initiative.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-003 (Phase 1) | Single dark factory PoC (manual, one variant) | Validates core hypothesis before building infra |
| R-003 (Phase 2) | Gallery coordinator (`multiclaude dark-factory create`, 3 variants, Tier 1 judges) | One-command multi-variant |
| R-003 (Phase 3) | Feedback loop (spec versioning, extraction, divergence, dispose-rebuild) | Full iterative cycle |
| R-003 (Phase 4-5) | AI judges panel, full autonomy, scheduled runs | Maturity features |
| R-005 (Phase B-C) | Dark factory token setup + repo template | GitHub App/PAT, CI template, relaxed CLAUDE.md |

**Shared prerequisite:** Group B (DFCP/provenance foundation).

### Group D: Orchestrator & Multi-Repo Support

**Why together:** The orchestrator pattern (aae-orc) requires multiclaude changes that are distinct from dark factory but share the DFCP foundation.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-011 (blocker) | multiclaude daemon patch: `git submodule update --init` post-worktree | #1 blocker; without this, orchestrator repos don't work at all |
| R-011 | repos.yaml manifest + 5-layer enforcement | Defines how multi-repo DFCP works |
| R-011 | Profile-based enhancement packaging (single-project vs orchestrator) | Install/upgrade system for multiclaude-enhancements |
| R-011 | Variant agent definitions (merge-queue, pr-shepherd, etc.) | Agents need orchestrator awareness |
| R-011 | BMAD as submodule + symlink/copy bridge | Command discovery in orchestrator repos |

**Shared prerequisite:** Group B (DFCP foundation) + the daemon submodule patch can start in parallel.

### Group E: Research & Intelligence Capabilities

**Why together:** Perplexity and the research supervisor form a coherent "agent intelligence" layer, independent of dark factory.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-008 (Phase 0) | **MCP config inheritance research (R-014)** | **PREREQUISITE:** Understand global→project→local cascade, per-worktree settings.local.json, worker dispatch flags before deploying any MCP server |
| R-008 (Phase 1) | Perplexity MCP server installation **with per-session toggle** (D-188) | ~30 min setup; **MCP DISABLED by default**; explicit opt-in per session required. User HAS API key. |
| R-008 (Phase 2) | Research-supervisor persistent agent definition | Dedup, routing, caching |
| R-008 (Phase 3-4) | Budget tracking, knowledge base, cache management | Sustainability features |
| R-010 (M-4) | Auto-learning from tool execution | Complements research — agents learn from doing + searching |

**Shared prerequisite:** Perplexity API key (user HAS key). **NEW prerequisite:** R-014 MCP config inheritance research must complete before Phase 1 — need to understand how to implement per-session toggle safely. **MCP Audit (2026-03-29):** Zero MCP servers currently active anywhere.

### Group F: External Integration (Slack Bot)

**Why together:** Fully independent initiative. Its own repo, own deployment, own lifecycle.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-009 (Phase 0-1) | Slack bot PoC → core control plane | `/mc-status`, `/mc-workers`, `/mc-dispatch` |
| R-009 (Phase 2) | Multi-machine bot-per-machine coordination | Multiple machines reporting to one Slack workspace |
| R-009 (Phase 3-4) | Rich integration, PR lifecycle threads, NL commands | Maturity features |

**Shared prerequisite:** Slack workspace setup, Slack App creation. Can proceed in parallel with everything else.

### Group G: Experience & Polish (aclaude Personas)

**Why together:** Nice-to-have that enhances agent personality without affecting functionality.

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-006 (Phase 1) | Manual persona prototype (append to worker.md, test) | ~1 hour validation |
| R-006 (Phase 2) | multiclaude persona config support | Code change to spawn pipeline |
| R-006 (Phase 3) | Theme sharing between aclaude and multiclaude | Future integration |

**Shared prerequisite:** None, but low priority. Should follow Group A.

### Group H: Marvel — End-State Platform (Long-Term Target)

**Why together:** These are the "2-6 month" items that feed into Marvel, the end-state platform. All Groups A-G are prototype work; Marvel is the real product.

**Marvel Vision (Human Operator Directive, 2026-03-29):**
Marvel is the top-level platform. Features: job management, message queue, traffic routing, secrets management, workspace/sandbox management, configurable REPL, crew management, cron/event triggers, git-native services (DAG), loadable content packs (multiclaude, BMAD, gastown, pennyfarthing, dollspace-gay).

**Evolution pipeline:** ThreeDoors (prototype) → multiclaude-enhancements (export) → aae-orc (mining/planning) → Marvel (real product).

| Research | What Goes Here | Rationale |
|----------|---------------|-----------|
| R-010 (L-1) | Replace tmux paste-buffer with proper agent protocol (ACP/JSON-RPC) | Fixes root cause of operator UX issues; Marvel needs proper IPC |
| R-010 (L-2) | Typed agent specialization (Explorer/Haiku, Implementer/Opus, etc.) | Cost optimization + quality improvement; Marvel crew management |
| R-010 (L-3) | VDD-style verification layer (adversarial pre-PR review) | Quality gate before PR creation; Marvel quality pipeline |
| R-010 (M-3) | Multi-model adversarial PR review (magpie-style) | Free via CLI provider subscriptions |
| R-007 (L-1) | Claude Code MCP server for message delivery | Eliminates tmux injection entirely; Marvel message queue |
| R-007 (L-2) | Operator dashboard (web UI) | Rich UX beyond tmux; Marvel control plane |
| R-004 (Phase 3-4) | Multi-plan support + full budget system | Scaling infrastructure; Marvel job management |
| **NEW** | Marvel platform architecture design | Job management, traffic routing, secrets, DAG services, content packs |
| **NEW** | Content pack system for loadable extensions | multiclaude, BMAD, gastown, pennyfarthing, dollspace-gay as packs |

**Shared prerequisite:** Groups A-C must be stable first. All prototype work feeds Marvel.

---

## 3. Dependency Map

```
                    ┌─────────────────────────┐
                    │   GROUP A               │
                    │   Operational Foundation │
                    │   (Start immediately)    │
                    │                         │
                    │   • Operator UX fixes   │
                    │   • Hook enforcement    │
                    │   • Session handoff     │
                    │   • Quota monitoring    │
                    │   • Daemon heartbeats   │
                    └──────────┬──────────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
              ▼                ▼                ▼
   ┌──────────────────┐  ┌─────────┐  ┌────────────────┐
   │   GROUP B        │  │GROUP E  │  │  GROUP F       │
   │   Golden Repo    │  │Research │  │  Slack Bot     │
   │   Hardening      │  │Perplexity│  │  (independent) │
   │                  │  │(Phase 1 │  │                │
   │   • CODEOWNERS   │  │immediate│  │  Can start     │
   │   • CI gates     │  │)        │  │  anytime       │
   │   • Provenance   │  └────┬────┘  └────────────────┘
   │   • .dfcp.yaml   │       │
   │   • Typed comments│      │
   └────────┬─────────┘       │
            │                 │
     ┌──────┼─────────────────┘
     │      │
     ▼      ▼
   ┌──────────────────┐   ┌─────────────────────┐
   │   GROUP C        │   │   GROUP D            │
   │   Dark Factory   │   │   Orchestrator       │
   │   PoC & Gallery  │   │   Multi-Repo         │
   │                  │   │                      │
   │   • Single PoC   │   │   • Submodule patch  │◄── Can start in parallel
   │   • Gallery coord│   │   • repos.yaml       │    (daemon patch is
   │   • Feedback loop│   │   • Profiles          │     independent)
   │   • Judges panel │   │   • Agent variants   │
   └────────┬─────────┘   └──────────┬──────────┘
            │                        │
            └────────────┬───────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   GROUP G            │
              │   Personas           │
              │   (nice-to-have)     │
              │                      │
              │   • Manual prototype │
              │   • Config support   │
              └──────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │   GROUP H            │
              │   MARVEL PLATFORM    │
              │   (end-state target) │
              │                      │
              │   • Agent protocol   │
              │   • Typed agents     │
              │   • VDD verification │
              │   • Multi-model review│
              │   • Operator dashboard│
              │   • Multi-plan budget│
              │   • Content packs    │
              │   • Job management   │
              │   • Message queue    │
              └──────────────────────┘

  PARALLEL TRACKS (can start anytime):
  ┌──────────┐  ┌──────────┐
  │ GROUP E  │  │ GROUP F  │
  │ Perplexity│  │ Slack Bot│
  │ Phase 1  │  │          │
  └──────────┘  └──────────┘
```

---

## 4. Consolidated Question List

Questions are numbered Q-C-NNN (C for consolidated), prioritized by how much work they unblock.

### Tier 1: CRITICAL — Blocks the most downstream work

These questions were raised in multiple research pieces (convergence = highest priority).

**ALL TIER 1 QUESTIONS DECIDED BY HUMAN OPERATOR ON 2026-03-29.**

| # | Question | Raised In | Options | Decision (2026-03-29) | Unblocks |
|---|----------|-----------|---------|----------------------|----------|
| Q-C-001 | **Should golden repo CODEOWNERS be applied immediately or wait?** | R-005 (OQ-DFCP-4) | Immediately / After Phase 0 / After Phase 1 | **✅ DECIDED: YES, apply now** | Groups B, C, D |
| Q-C-002 | **Should `agents/*.md` changes require human review via CODEOWNERS?** | R-005 (OQ-DFCP-5), R-006 (persona safety) | Yes / No | **✅ DECIDED: YES, require human review** | Groups B, G |
| Q-C-003 | **Dark factory repo visibility: public or private?** | R-003 (OQ-1), R-005 | Public / Private | **✅ DECIDED: PRIVATE but configurable** | Group C |
| Q-C-004 | **Maximum budget per dark factory run?** | R-003 (OQ-4) | Fixed / Configurable / Unlimited | **✅ DECIDED: CONFIGURABLE (no fixed default)** — owner sets per project needs | Group C |
| Q-C-005 | **Should hook-enforced git safety replace prompt-level INC-002?** | R-010 (OQ-CL-1), R-011 (enforcement layer 5) | Hooks / Prompt-only / Both | **✅ DECIDED: HOOKS (replace prompt-level)** | Groups A, B, C, D |
| Q-C-006 | **Should the multiclaude daemon patch for submodule init be always-on or opt-in?** | R-011 (OQ-ORC-2) | Always-on / Opt-in / Configurable | **✅ DECIDED: YES with skip flag** (always-on, opt-out) | Group D |

### Tier 2: HIGH — Blocks one group or major feature

| # | Question | Raised In | Options | Recommendation | Unblocks |
|---|----------|-----------|---------|----------------|----------|
| Q-C-007 | **Provenance tagging: mandatory or opt-in?** | R-003 (OQ-5) | Mandatory for all / Mandatory for AI + opt-in for human / Opt-in | **Mandatory for AI, opt-in for human** | Group B Phase 0 |
| Q-C-008 | **Perplexity daily budget cap?** | R-008 (OQ-1) | $5 / $10 / Configurable | **$5/day to start, configurable** | Group E |
| Q-C-009 | **Where does the Slack bot repo live?** | R-009 (OQ-SB-2, OQ-SB-6) | ThreeDoors epic / Own repo / multiclaude module | **Own repo** — spans all multiclaude repos, not ThreeDoors-specific | Group F |
| Q-C-010 | **Session handoff: per-agent files or daemon feature?** | R-010 (OQ-CL-2), R-007 | Per-agent files / Shared store / Daemon feature | **Daemon feature** — daemon already manages agent lifecycle | Group A |
| Q-C-011 | **Should CronCreate heartbeats be dropped entirely?** (daemon wake loop already nudges every 2 min) | R-007 (S-2) | Drop / Keep / Replace with daemon-native | **Drop immediately, replace with daemon-native in M-1** | Group A |
| Q-C-012 | **CI scope-check workflow: block merge or just warn?** | R-005 (OQ-DFCP-2) | Block / Warn / Off | **Warn initially**, upgrade to block after validation | Group B |
| Q-C-013 | **Should workers call `perplexity_research` (expensive deep research) directly?** | R-008 (OQ-3) | Direct / Via research-supervisor only | **Direct for search/ask; research-supervisor for deep research** — expensive queries benefit from dedup | Group E |
| Q-C-014 | **Does Claude Code's file scanner follow symlinks in `.claude/commands/`?** | R-011 (OQ-ORC-1) | Needs empirical test | **Test immediately** — determines BMAD-as-submodule bridge strategy | Group D |

### Tier 3: MEDIUM — Important but doesn't block critical path

| # | Question | Raised In | Options | Recommendation | Unblocks |
|---|----------|-----------|---------|----------------|----------|
| Q-C-015 | **Daemon wake loop interval: keep 2 min or make configurable?** | R-007 (OQ-1, OQ-4) | Keep 2 min / Increase to 5 min / Configurable per-agent | **Configurable** — persistent agents need different intervals than watchdogs | Group A medium-term |
| Q-C-016 | **Should dark factory output be preserved after disposal?** | R-003 (OQ-3) | Full / Metadata only / Nothing | **Metadata only** (spec refinements, feedback, test summaries) | Group C |
| Q-C-017 | **Persona theme selection: global or per-agent?** | R-006 (Q1) | Global / Per-agent / Both | **Global with per-agent override** — simpler default, flexibility where needed | Group G |
| Q-C-018 | **Should immersion be hard-capped at `none` for merge-queue/pr-shepherd/project-watchdog?** | R-006 (Q4) | Hard cap / Trust user | **Hard cap** — protocol corruption risk is not worth the flexibility | Group G |
| Q-C-019 | **Persona in PR descriptions and commit messages?** | R-006 (Q5) | Yes / No | **No** — persona affects conversational tone only, not artifacts | Group G |
| Q-C-020 | **Perplexity MCP: project-level or user-level config?** | R-008 (OQ-6) | Project / User / Both | **Project-level** — scoped budget tracking; user can add globally separately | Group E |
| Q-C-021 | **Who can trigger a dark factory run?** | R-003 (OQ-2) | Anyone / Owner / Configurable | **Configurable with owner default** | Group C |
| Q-C-022 | **Where do orchestrator-level stories live?** | R-011 (OQ-ORC-4) | Orchestrator `docs/stories/` / Distributed across sub-repos | **Orchestrator repo** — single view for project-watchdog | Group D |
| Q-C-023 | **Multiple Max subscriptions on same machine — TOS implications?** | R-004 (QT-Q1) | Compliant / Gray area / Risk | **Gray area** — CLAUDE_CONFIG_DIR is officially endorsed but multi-account unclear. Proceed cautiously. | Group H |
| Q-C-024 | **Slack bot: Socket Mode or webhooks?** | R-009 (OQ-SB-3) | Socket Mode / Webhooks | **Socket Mode** — no public URL needed, works behind NAT | Group F |
| Q-C-025 | **Where is "ourbot"?** | R-009 (OQ-SB-1) | GitLab? Other org? Fresh start? | **User to confirm** — searched all known GitHub orgs, not found | Group F |
| Q-C-026 | **Should blast radius limiting be hooks or alerts?** | R-010 (OQ-CL-5) | Hard enforcement / Soft alerts / Both | **Alerts first, then hooks** — need to calibrate thresholds before enforcing | Group H |
| Q-C-027 | **Multi-model adversarial review or improve arch-watchdog?** | R-010 (OQ-CL-4) | Multi-model / Better watchdog / Both | **Both** — arch-watchdog for passive, multi-model for active PR review | Group H |

---

## 5. NEW STRATEGIC DIRECTION: Platform Extraction (Added 2026-03-29)

**Human operator directive:** Extract multiclaude customizations OUT of ThreeDoors into a standalone reusable platform.

### Vision
- **Goal:** Standalone platform that works with ANY project, not coupled to ThreeDoors
- **Likely path:** New repo, used as submodule in aae-orc for dark factory
- **Evolution:** Starts as multiclaude replacement/fork, evolves into something new
- **Impact:** Reframes Groups C and D — orchestrator and dark factory work should target the extracted platform, not ThreeDoors-specific customizations

### multiclaude Licensing (R-013)

**Source:** `dlorenc/multiclaude` on GitHub (public repo, Go binary)
- **Module:** `github.com/dlorenc/multiclaude`
- **README states:** MIT license
- **LICENSE file:** ❌ MISSING — no LICENSE file exists in the repo
- **npm "multiclaude":** Different project entirely (by dexhorthy-humanlayer, scaffolding tool) — NOT the same software

**Legal assessment:**
- MIT intent is clear from README but technically incomplete without a LICENSE file
- Under copyright law, absence of a license file defaults to "all rights reserved"
- **Risk: LOW** — MIT is maximally permissive: fork, modify, redistribute, commercial use all allowed
- The author's public declaration of MIT in README strongly indicates intent

**Recommended actions:**
1. Request dlorenc add a formal LICENSE file to the repo (ideal path — formalizes what's already stated)
2. OR proceed with fork, documenting the MIT claim from README as basis
3. **Not a hard blocker** for development, but should be resolved before public redistribution of derivative works

### Phase 0 Blocker Assessment

Licensing is **not a hard blocker** for starting platform extraction work:
- Development and internal use can proceed under the stated MIT intent
- The blocker only activates for **public redistribution** of derivative works
- Filing a GitHub issue or PR to add LICENSE file would resolve this quickly

### Impact on Planning Sequence

The platform extraction strategy means:
- **Group D (Orchestrator)** work should be designed for the new platform, not as ThreeDoors-specific changes
- **Group C (Dark Factory)** infrastructure should target the extracted platform
- **Groups A & B** remain unchanged — they stabilize the current environment regardless
- A new **Phase 0.5** may be needed between Phase 1 and Phase 2 for platform repo setup

---

## 6. Recommended Planning Sequence

### Phase 1: Stabilize & Harden (Plan Together First)

**Groups A + B (partial) — estimated 1-2 weeks of planning**

**What:** Operational foundation + golden repo hardening. These are the "fix the floor before building the house" items.

**Why first:**
- Every other initiative runs on this infrastructure
- Hook enforcement (R-010 S-1) makes all subsequent work safer
- Operator UX fixes (R-007 S-1, S-2) reduce friction for everything
- CODEOWNERS (R-005 Phase A) is prerequisite for dark factory
- Provenance tagging (R-003 Phase 0) is prerequisite for dark factory
- Quota monitoring (R-004 Phase 1) prevents hitting walls during dark factory

**Human decisions needed before planning:**
- Q-C-001: Apply CODEOWNERS now? (recommend: yes)
- Q-C-002: Gate `agents/*.md`? (recommend: yes)
- Q-C-005: Hooks vs prompts for git safety? (recommend: hooks)
- Q-C-007: Provenance mandatory? (recommend: mandatory for AI)
- Q-C-010: Session handoff mechanism? (recommend: daemon feature)
- Q-C-011: Drop CronCreate heartbeats? (recommend: yes)
- Q-C-012: CI scope-check blocking? (recommend: warn initially)

**Immediate actions (zero planning needed):**
- Human works in workspace window, not supervisor (R-007 S-1) — behavioral change only
- Drop CronCreate heartbeats from startup checklist (R-007 S-2) — MEMORY.md edit only
- Install Perplexity MCP server (R-008 Phase 1) — 30 min setup, parallel track

### Phase 2: Dark Factory PoC (Plan After Phase 1 Decisions)

**Groups C (Phase 1-2) + B (remaining) — estimated 1 week of planning**

**What:** Single dark factory PoC, gallery coordinator, DFCP token setup, repo template.

**Why second:**
- Validates core hypothesis before building gallery infrastructure
- Requires CODEOWNERS and provenance from Phase 1
- PoC data informs decisions about gallery size, budget, timing

**Human decisions needed before planning:**
- Q-C-003: Private factory repos? (recommend: yes)
- Q-C-004: Budget per run? (recommend: $50 configurable)
- Q-C-016: What survives disposal? (recommend: metadata only)
- Q-C-021: Who can trigger factory runs? (recommend: configurable)

### Phase 3: Scale & Integrate (Plan After PoC Results)

**Groups D + E + remaining C — estimated 2 weeks of planning**

**What:** Orchestrator support, research supervisor, gallery feedback loop, AI judges.

**Why third:**
- Orchestrator pattern (R-011) needs validated dark factory concept
- Research supervisor (R-008) is more valuable once dark factory exists (research-before-coding)
- Gallery feedback loop (R-003 Phase 3) depends on PoC learnings

**Human decisions needed before planning:**
- Q-C-006: Submodule init always-on? (recommend: yes with skip flag)
- Q-C-008: Perplexity daily budget? (recommend: $5/day)
- Q-C-013: Deep research routing? (recommend: via supervisor)
- Q-C-014: Symlink test result (empirical)
- Q-C-022: Orchestrator story location? (recommend: orchestrator repo)

### Parallel Tracks (Can Start Anytime)

**Group F (Slack Bot) — independent lifecycle**

Plan and execute whenever bandwidth allows. Not on the critical path for any other group.

**Human decisions needed:**
- Q-C-009: Own repo (recommend: yes)
- Q-C-024: Socket Mode (recommend: yes)
- Q-C-025: Where is "ourbot"? (user to confirm)

**Group G (Personas) — nice-to-have**

Low priority. Manual prototype (1 hour) can happen anytime as validation. Full implementation after Phase 1.

**Human decisions needed:**
- Q-C-017: Global vs per-agent theme (recommend: global with override)
- Q-C-018: Hard cap for critical agents (recommend: yes)
- Q-C-019: No persona in artifacts (recommend: correct)

### Phase 4: Marvel Platform (Long-term, Plan After Phase 3)

**Group H — requires Phase 1-3 stability. Marvel is the end-state platform.**

**Marvel Vision:** All Groups A-G are prototype work. Marvel is the real product with: job management, message queue, traffic routing, secrets management, workspace/sandbox management, configurable REPL, crew management, cron/event triggers, git-native services (DAG), loadable content packs (multiclaude, BMAD, gastown, pennyfarthing, dollspace-gay).

**Evolution pipeline:** ThreeDoors (prototype) → multiclaude-enhancements (export) → aae-orc (mining/planning) → Marvel (real product).

Architecture evolution items (agent protocol, typed agents, VDD verification, multi-model review, operator dashboard, multi-plan budget) are now framed as Marvel infrastructure rather than standalone improvements.

**Human decisions needed:**
- Q-C-023: Multi-account TOS (research first, decide later)
- Q-C-026: Blast radius enforcement (alerts then hooks)
- Q-C-027: Adversarial review strategy (both)
- Marvel architecture scope and initial repo setup

---

## 7. Cross-Cutting Themes

Several themes emerged across multiple research artifacts:

### Theme 1: "Enforce via Tooling, Not Prompts" (R-010, R-005, R-011)
Three independent research pieces converge on this: hook-based enforcement is mechanically reliable while prompt-level "DO NOT" instructions are suggestions that agents can violate. This should be the governing principle for all governance design going forward.

### Theme 2: Disposability as Safety (R-003, R-005)
The dark factory's core insight — code is the most disposable artifact — has broader implications. Any system where failure means "wasted compute, not lost data" is inherently safe to experiment with.

### Theme 3: Complementary Tools, Not Replacements (R-008, R-010)
Perplexity complements Gemini (web search vs deep research). multiclaude's orchestration complements chainlink's enforcement. The pattern is "best tool for each job" not "one tool to rule them all."

### Theme 4: Convention Over Enforcement (R-008, R-003)
The Perplexity research supervisor is a "soft gatekeeper" — all agents CAN access Perplexity directly, but the convention routes complex research through the supervisor. This matches the existing merge-queue pattern and avoids bottlenecks.

### Theme 5: The Workspace Window Is Underutilized (R-007)
The most impactful finding may be the simplest: the workspace window exists specifically for human interaction and is explicitly exempted from all automated injections. This is a zero-code fix that addresses the #1 operator pain point.

---

## Appendix: Research-to-Group Mapping

| Research | Group A | Group B | Group C | Group D | Group E | Group F | Group G | Group H |
|----------|---------|---------|---------|---------|---------|---------|---------|---------|
| R-003 Dark Factory | | Phase 0 | Phases 1-5 | | | | | |
| R-004 Quota | Phase 1 | | | | | | | Phases 3-4 |
| R-005 DFCP | | Phase A, CI | Phases B-C | | | | | |
| R-006 Personas | | | | | | | All | |
| R-007 Operator UX | S-1, S-2, M-1 | | | | | | | L-1, L-2 |
| R-008 Perplexity | | | | | All | | | |
| R-009 Slack Bot | | | | | | All | | |
| R-010 Chainlink | S-1, S-2 | S-3 | | | M-4 | | | L-1, L-2, L-3, M-3 |
| R-011 Orchestrator | | | | All | | | | |
