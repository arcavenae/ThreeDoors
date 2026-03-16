# SpecKit Evaluation: Potential Replacement for BMAD Planning Toolchain

**Date:** 2026-03-15
**Type:** Research Report
**Requested by:** Supervisor
**Produced by:** Worker (cool-otter)

---

## Executive Summary

GitHub Spec Kit (github.com/github/spec-kit) is a specification-driven development (SDD) toolkit that enforces a linear Specify → Plan → Tasks → Implement workflow. After thorough evaluation, **SpecKit is not a suitable replacement for our BMAD-based toolchain**. It solves a fundamentally different problem (greenfield feature scaffolding) than what our system handles (ongoing multi-agent project governance with 377 stories, 40+ epics, and 764+ merged PRs). However, SpecKit's "constitution" concept and cross-agent template portability offer ideas worth borrowing.

**Recommendation:** Do not adopt SpecKit. Selectively borrow its constitution pattern. Consider Intent (Augment Code) for a future evaluation if bidirectional spec-code sync becomes a pain point.

---

## 1. What Is SpecKit?

### Overview

SpecKit is GitHub's open-source (MIT) toolkit for Spec-Driven Development. It provides a CLI (`specify-cli`, installed via `uv`) and a set of slash commands that guide AI coding agents through a structured workflow. Currently at ~v0.1.4 with 28K+ GitHub stars.

**Repository:** https://github.com/github/spec-kit

### Core Workflow (Strictly Linear)

| Step | Command | Output |
|------|---------|--------|
| 1. Constitution | `/speckit.constitution` | `constitution.md` — project governance principles |
| 2. Specify | `/speckit.specify` | `spec.md` — feature requirements (what/why, not how) |
| 3. Clarify | `/speckit.clarify` | Resolves ambiguities in spec |
| 4. Plan | `/speckit.plan` | `plan.md` — architecture, tech stack, components |
| 5. Analyze | `/speckit.analyze` | Cross-artifact consistency check |
| 6. Tasks | `/speckit.tasks` | `tasks.md` — 10-20 atomic micro-tasks |
| 7. Implement | `/speckit.implement` | Generated code, task by task |

### File Structure

```
.specify/            # Config, templates, scripts
.specify/memory/     # constitution.md
specs/               # Feature-sliced specs (001-feature-name/)
  001-feature/
    spec.md
    plan.md
    tasks.md
.github/prompts/     # AI agent templates
```

### Key Characteristics

- **Vendor-neutral:** Templates for 8+ AI agents (Claude, Copilot, Cursor, Gemini, etc.)
- **Feature-sliced:** Each feature gets its own numbered directory with isolated spec/plan/tasks
- **Git-integrated:** Feature branches tied to spec directories
- **Static specs:** Specifications don't auto-update when code changes
- **No multi-agent orchestration:** Single-agent workflow only

---

## 2. Comparison with Our Current BMAD System

### What We Have Today

| Capability | Our System | Details |
|---|---|---|
| Planning docs | PRD shards (14 files), epic-list.md, epics-and-stories.md, ROADMAP.md | Interconnected doc chain with defined authority hierarchy |
| Story management | 377 story files (`docs/stories/X.Y.story.md`) | Individual story files with ACs, status tracking, epic grouping |
| Agent orchestration | 12+ BMAD role-based agents (PM, Architect, SM, QA, Dev, etc.) | Multi-agent party mode for design decisions |
| Planning artifacts | 186 artifacts in `_bmad-output/planning-artifacts/` | Research reports, party mode transcripts, architecture spikes |
| Decision tracking | `docs/decisions/BOARD.md` | Adopted + rejected alternatives with rationale |
| Validation | `/bmad-bmm-validate-prd`, `/bmad-bmm-check-implementation-readiness` | PRD validation, implementation readiness checks |
| Reconciliation | `/reconcile-docs` | Cross-doc consistency verification |
| Status tracking | `/bmad-bmm-sprint-status`, story file status fields | Sprint status, story-level Done/In Progress/Not Started |
| Multi-agent runtime | multiclaude (supervisor, merge-queue, pr-shepherd, envoy, etc.) | Persistent agents with inter-agent messaging |

### Feature Comparison Matrix

| Capability | SpecKit | Our BMAD System | Winner |
|---|---|---|---|
| **Greenfield feature scaffolding** | Excellent — designed for this | Good — `/plan-work` pipeline | SpecKit |
| **Ongoing project governance** | None — no concept of sprints, epics, or status | Deep — epic/story hierarchy, status sync, sprint tracking | BMAD |
| **Multi-agent orchestration** | None — single-agent only | 12+ specialized agents with party mode | BMAD |
| **Cross-doc consistency** | `/speckit.analyze` (within one feature) | `/reconcile-docs` across all planning docs + story files | BMAD |
| **Decision recording** | None | BOARD.md with adopted + rejected options | BMAD |
| **Brownfield support** | Retrofit mode with `--no-git`, but limited | Native — story-driven with full epic backlog | BMAD |
| **AI agent portability** | Excellent — 8+ agents supported | IDE-agnostic via BMAD-METHOD, but tightly coupled to Claude | SpecKit |
| **Story/task granularity** | 10-20 micro-tasks per feature | Full story files with ACs, estimates, dependencies | BMAD |
| **Spec-code sync** | Manual (static specs) | Manual (story files updated post-implementation) | Tie |
| **Learning curve** | Low — 7 commands | High — 21+ agents, multiple slash commands, doc hierarchy | SpecKit |
| **Community/ecosystem** | 28K stars, active GitHub maintenance | Custom/internal, BMAD-METHOD open source | SpecKit |
| **Operational overhead** | Low — files in `.specify/` and `specs/` | High — 14 PRD shards, 377 stories, 186 artifacts, ROADMAP sync | SpecKit |
| **Course correction** | No formal mechanism — update specs manually | `/course-correct` with sprint change proposals | BMAD |
| **Retrospectives** | None | `/bmad-bmm-retrospective` with lesson extraction | BMAD |
| **Quality gates** | `/speckit.checklist` (pre-implementation only) | QA agent, test design, NFR assessment, pre-PR validation | BMAD |

### Where SpecKit Wins

1. **Simplicity** — 7 commands vs our 50+ slash commands and agent ecosystem
2. **Portability** — Works with any AI agent, not just Claude
3. **Constitution concept** — A single governance doc that constrains all AI behavior (similar to our CLAUDE.md but more formalized for spec-level work)
4. **Low barrier to entry** — New contributor can understand the workflow in minutes

### Where BMAD Wins (Decisively)

1. **Scale** — We have 40+ epics, 377 stories, 764+ PRs. SpecKit has no concept of managing this.
2. **Governance** — Multi-agent party mode, decision recording, doc authority chains — none exist in SpecKit
3. **Ongoing management** — Sprint status, reconciliation, course corrections — SpecKit is greenfield-only in practice
4. **Quality assurance** — QA agent, test architecture, pre-PR validation — SpecKit has basic checklists only
5. **Traceability** — Story → epic → PRD requirement → PR chain is fully tracked

---

## 3. Could SpecKit Replace or Simplify Any of Our Workflows?

### Could Replace: Nothing Critical

SpecKit's scope (single-feature specification → implementation) doesn't overlap with our pain points. Our challenges are:
- Keeping 3 planning docs in sync (epic-list, epics-and-stories, ROADMAP)
- Managing 377 story files and their status transitions
- Coordinating 6+ concurrent agents
- Decision tracking across party mode sessions

SpecKit addresses none of these.

### Could Supplement: Constitution Pattern

SpecKit's `constitution.md` concept — a single file that establishes inviolable project principles — is interesting. We approximate this with `SOUL.md` + `CLAUDE.md`, but SpecKit's constitution is specifically designed to constrain AI spec-writing and planning decisions, not just coding behavior.

**Borrowable idea:** Create a `CONSTITUTION.md` that merges SOUL.md principles with BMAD governance rules into a format optimized for AI planning agents.

### Could Supplement: Feature-Scoped Spec Isolation

SpecKit's `specs/001-feature/` directory structure could be useful for large new features that need their own spec/plan/tasks lifecycle before being decomposed into our story format. Currently our `/plan-work` pipeline goes straight from research to stories.

**Borrowable idea:** For large epics (5+ stories), create a `docs/specs/epic-NN/` directory with spec.md and plan.md before story decomposition.

---

## 4. Migration Assessment

### What Migration Would Look Like

Migration from BMAD to SpecKit is **not feasible** for ThreeDoors:

| Factor | Assessment |
|---|---|
| **Effort** | Massive — would require restructuring 377 stories, 14 PRD shards, 186 artifacts into SpecKit's flat feature-sliced format |
| **Data loss** | Severe — no SpecKit equivalent for party mode transcripts, decision board, sprint change proposals, retrospective findings |
| **Agent ecosystem** | Complete loss — SpecKit has no multi-agent concept; all 6 persistent agents + BMAD role agents would be abandoned |
| **Governance** | Complete loss — no doc authority chain, no reconciliation, no project-watchdog, no merge-queue scope checks |
| **Status tracking** | Degraded — SpecKit tracks task completion within one feature, not across 40+ epics |
| **Risk** | Extremely high — would halt all parallel development during migration |

### What We'd Lose

- Party mode multi-agent design sessions
- Decision board with rejected alternatives
- Sprint change proposal process
- Doc reconciliation and staleness detection
- Persistent agent ecosystem (merge-queue, pr-shepherd, envoy, project-watchdog)
- Epic/story hierarchy and cross-cutting traceability
- 186 accumulated planning artifacts with institutional knowledge

### What We'd Gain

- Simpler file structure for new features
- AI agent portability (currently locked to Claude)
- Lower cognitive overhead for new contributors
- Community-supported tooling updates

**The gains do not justify the losses.** ThreeDoors is not greenfield — it's a mature project with deep institutional knowledge encoded in its planning system.

---

## 5. Other Tools Worth Considering

Based on research into the SDD landscape:

### Intent (Augment Code) — $60/mo

**Most interesting alternative.** Intent offers "living specs" with bidirectional spec-code synchronization — specs auto-update when code changes. Also has multi-agent coordination with isolated worktrees (similar to our multiclaude model).

| Strength | Relevance to Us |
|---|---|
| Living specs (bidirectional sync) | Would solve our spec-staleness problem |
| Multi-agent with coordinator | Similar to our supervisor pattern |
| Isolated worktrees per agent | We already have this via multiclaude |

**Limitation:** Proprietary, limited benchmarks, variable pricing. Worth a future spike if spec drift becomes a blocker.

### OpenSpec — Free (Open Source)

Lightweight alternative (~250 lines output vs SpecKit's ~800). Has "delta markers" that track spec changes explicitly and "proposal gates" for brownfield modifications. Better suited for ongoing projects than SpecKit.

| Strength | Relevance to Us |
|---|---|
| Brownfield-first design | Matches our situation |
| Delta markers for change tracking | Could help with story status tracking |
| Proposal gates | Similar to our sprint change proposals |

**Limitation:** Static specs during implementation, manual reconciliation required. Not enough advantage over our current system.

### Kiro (Amazon) — Free Tier Available

Uses structured EARS notation for acceptance criteria. AWS ecosystem integration. Single-agent (Claude only).

**Not relevant:** Too narrowly focused on AWS projects. No multi-agent support. No advantage over BMAD.

### Cursor with .cursorrules

Pseudo-specs only. No structured workflow. Not a real contender.

---

## 6. Recommendation

### Primary: Do Not Adopt SpecKit

SpecKit is an excellent tool for a different problem. It excels at scaffolding greenfield features with a single AI agent. ThreeDoors is a mature project with 40+ completed epics, 377 stories, and an orchestrated multi-agent system. Adopting SpecKit would be like replacing a factory's production management system with a single workbench.

### Secondary: Borrow Two Ideas

1. **Constitution pattern** — Consider creating a `docs/CONSTITUTION.md` that consolidates SOUL.md + CLAUDE.md governance principles into a format specifically optimized for AI planning decisions. Lower priority — current SOUL.md + CLAUDE.md combo works.

2. **Feature-scoped spec isolation** — For large new epics (5+ stories), consider a `docs/specs/epic-NN/` directory with dedicated spec.md and plan.md before story decomposition. This adds a structured "think before you write stories" step to `/plan-work`.

### Tertiary: Watch Intent (Augment Code)

If spec-code drift becomes a significant pain point, evaluate Intent's bidirectional living specs. Its multi-agent + worktree model aligns with our multiclaude architecture. No action needed now.

### What NOT to Do

- Do not attempt to migrate existing stories/epics into any SDD tool format
- Do not adopt SpecKit "just for new features" — the cognitive overhead of two parallel systems outweighs the benefit
- Do not abandon BMAD's multi-agent party mode — it's our most valuable planning differentiator

---

## Sources

- [GitHub Spec Kit Repository](https://github.com/github/spec-kit)
- [SpecKit Official Site](https://speckit.org/)
- [LogRocket: Exploring Spec-Driven Development with GitHub Spec Kit](https://blog.logrocket.com/github-spec-kit/)
- [Scott Logic: Putting Spec Kit Through Its Paces](https://blog.scottlogic.com/2025/11/26/putting-spec-kit-through-its-paces-radical-idea-or-reinvented-waterfall.html)
- [Augment Code: 6 Best Spec-Driven Development Tools](https://www.augmentcode.com/tools/best-spec-driven-development-tools)
- [Deep Dive Guide: GitHub Spec Kit](https://redreamality.com/garden/notes/github-spec-kit-guide/)
- [Microsoft Developer Blog: Diving Into Spec-Driven Development](https://developer.microsoft.com/blog/spec-driven-development-spec-kit)
- [IntuitionLabs: GitHub Spec Kit Guide](https://intuitionlabs.ai/articles/spec-driven-development-spec-kit)
- [GSD vs Spec Kit vs OpenSpec vs Taskmaster AI (Medium)](https://medium.com/@richardhightower/agentic-coding-gsd-vs-spec-kit-vs-openspec-vs-taskmaster-ai-where-sdd-tools-diverge-0414dcb97e46)
