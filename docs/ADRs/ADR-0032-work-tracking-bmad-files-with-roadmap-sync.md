# ADR-0032: BMAD Epic/Story Files as Primary Tracker with ROADMAP Sync

- **Status:** Accepted
- **Date:** 2026-03-08
- **Decision Makers:** Project founder
- **Related PRs:** #265
- **Related ADRs:** ADR-0025 (Story-Driven Development)

## Context

The project currently has two work tracking systems:
- **BMAD:** Epic files (`docs/prd/epics-and-stories.md`, `docs/prd/epic-list.md`) and story files (`docs/stories/*.story.md`) — used by BMAD agents for story creation, implementation, and review
- **Multiclaude:** `ROADMAP.md` — used by merge-queue for scope checks and worker prioritization

Both have problems. ROADMAP.md drifts from reality (Issue #252). BMAD files are richer but not what multiclaude agents read. The two systems overlap and diverge.

Additionally, future projects may span multiple repos, and different projects may use different tracking tools (GitHub Issues, GitLab, Jira). The tracking location should ideally be configurable per project.

## Decision

1. **BMAD epic/story files are the source of truth** for work tracking
2. **ROADMAP.md syncs from BMAD files**, not the other way around — it's a derived view for multiclaude's scope-checking needs
3. **Multiclaude agents should be updated to read BMAD files** (epic-list.md, story files) rather than relying solely on ROADMAP.md
4. **`/reconcile-docs`** handles the sync between BMAD files and ROADMAP.md
5. **Per-project tracking location is a future capability** — when projects span repos or use external tools (Jira, GitLab), the system should support specifying where tracking lives and how to query it. This is deferred until a concrete need arises.

## Rationale

- BMAD files are richer (acceptance criteria, tasks, dependencies, dev notes) — they serve both planning and implementation
- Story files are already the source of truth per CLAUDE.md rules — formalizing this eliminates ambiguity
- ROADMAP.md serves a real purpose (scope checking) but should be a projection, not a competing tracker
- Deferring per-project tracking configuration avoids building infrastructure for a problem that doesn't exist yet
- The `/reconcile-docs` command (PR #265) already handles the sync gap

## Rejected Alternatives

- **GitHub Issues as canonical tracker:** GH Issues are repo-scoped, which doesn't work well for multi-repo projects. Also loses the rich story file format that agents rely on.
- **Jira as canonical tracker:** Adds tool complexity and cost. The project uses Jira elsewhere but doesn't need it here yet.
- **Hybrid GH Issues + story files:** Two places to look creates drift risk. Adds migration effort for uncertain benefit.
- **Status quo without formalization:** The ambiguity about which system is primary causes drift. Making BMAD files explicitly primary resolves this.

## Open Questions (Deferred)

- How to specify per-project tracking location (config file? env var? agent parameter?)
- How to query external trackers (Jira API? GitLab API?) and distinguish work items from non-tracking items
- How to handle projects that span repos where story files live in one repo but implementation happens in another
