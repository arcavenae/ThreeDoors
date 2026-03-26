# Issue Tracker

<!-- authority-tiers:
  owner: [arcaven]
  contributors: []
-->
<!-- last-patrol: 2026-03-08T00:00:00Z -->

## Open Issues

| Issue # + Title | Status | Linked Story | Linked PR(s) | Reporter | Date Reported | Last Envoy Update | Current State |
|-----------------|--------|--------------|---------------|----------|---------------|-------------------|---------------|
| [#219 Door selection lacks tactile feedback and intuitive interaction patterns](https://github.com/arcavenae/ThreeDoors/issues/219) | triaged | `docs/stories/36.1.story.md` | #221 (triage) | arcaven | 2026-03-08 | 2026-03-08 | Triage complete. UX design stories created (Epic 36). Implementation not started. |
| [#244 Move signing secrets to a protected GitHub environment](https://github.com/arcavenae/ThreeDoors/issues/244) | open | — | — | arcaven | 2026-03-08 | 2026-03-08 | Security audit finding. No triage started. |
| [#245 Replace softprops/action-gh-release with gh CLI](https://github.com/arcavenae/ThreeDoors/issues/245) | open | — | — | arcaven | 2026-03-08 | 2026-03-08 | Security audit finding. No triage started. |
| [#246 Pass secrets via env vars instead of ${{ }} interpolation in run blocks](https://github.com/arcavenae/ThreeDoors/issues/246) | open | — | — | arcaven | 2026-03-08 | 2026-03-08 | Security audit finding. No triage started. |
| [#248 Pin golangci-lint version in CI](https://github.com/arcavenae/ThreeDoors/issues/248) | open | — | — | arcaven | 2026-03-08 | 2026-03-08 | Security audit finding. No triage started. |
| [#252 Reconcile planning docs: epic-list.md, epics-and-stories.md, ROADMAP.md are out of sync](https://github.com/arcavenae/ThreeDoors/issues/252) | open | — | — | arcaven | 2026-03-08 | 2026-03-08 | Planning doc drift detected by analyst audit. No triage started. |

## Recently Resolved

| Issue # + Title | Status | Linked Story | Linked PR(s) | Reporter | Date Reported | Date Resolved | Resolution |
|-----------------|--------|--------------|---------------|----------|---------------|---------------|------------|
| [#218 Panic: nil pointer dereference when textfile provider not registered](https://github.com/arcavenae/ThreeDoors/issues/218) | resolved | `docs/stories/23.11.story.md` | #220 (triage), #225 (fix) | arcaven | 2026-03-08 | 2026-03-08 | Fixed via nil provider guard in CLI and MCP server. |

## SOUL.md Alignment Reference

### Three-Category Classification

Every issue is classified into one of three alignment categories:

1. **Clearly Aligned** — The request fits SOUL.md values, ROADMAP.md scope, and existing patterns. Proceed with normal triage.

2. **Clearly Misaligned** — The request contradicts core project values documented in SOUL.md. The envoy can recognize these and respond with a polite decline referencing the specific principle, without supervisor escalation.

3. **Gray Area** — The request is interesting but alignment is uncertain. ALWAYS escalate to supervisor. Never reject gray-area requests unilaterally.

### Common Misalignment Patterns

| Request Pattern | SOUL.md Principle | Conflict | Response Approach |
|----------------|-------------------|----------|-------------------|
| "Show more than 3 tasks" | Three Doors, Not Three Hundred | The constraint IS the feature. Showing 3 tasks is the core design choice that reduces decision friction. | Explain that the limit is intentional — it works *with* human psychology by eliminating choice paralysis. |
| "Add cloud sync/accounts" | Local-First, Privacy-Always | Data stays on the user's machine. No accounts, no cloud sync unless user explicitly configures it. | Explain data sovereignty philosophy and that integrations use local APIs or user-provided tokens. |
| "Team features/sharing" | Personal tool for one person at a time | ThreeDoors is not a project management tool — no team features. | Suggest purpose-built tools (Jira, Linear) and note ThreeDoors has adapters to integrate with them. |
| "Gamification/streaks" | Not a habit tracker | ThreeDoors focuses on action over motivation. No streaks, no gamification, no guilt. | Explain focus on helping users *do* things, not tracking whether they did them. |
| "Knowledge graph/tagging" | Not a second brain | ThreeDoors is not trying to organize knowledge — no linking, no tagging taxonomy. | Suggest Obsidian (ThreeDoors has an Obsidian adapter) for knowledge management. |
| "Analytics dashboard" | Progress Over Perfection | The goal is action, not optimization. Imperfect action beats perfect planning. | Explain that ThreeDoors measures success by tasks started, not productivity metrics. |
| "Web/mobile version" | Solo Dev Reality | Built by one person. Every feature must justify its complexity. Cross-platform adds significant burden. | Explain resource constraints. Note MCP integration (Epic 24) as alternative for remote access via LLM agents. |

### Polite Decline Template

When declining a misaligned request, follow this structure:

1. **Thank genuinely** — "Thanks for suggesting this! I can see how [feature] would be useful."
2. **Acknowledge the need** — Recognize the real problem behind the request. The reporter isn't wrong to want it.
3. **Cite the specific principle** — Reference the SOUL.md value it conflicts with. Never say "we just don't want to" — give the philosophical or architectural reason.
4. **Suggest alternatives** — If possible, point to a different tool, an existing adapter, or a way ThreeDoors addresses their underlying need differently.
5. **Invite discussion** — "If you think there's a way to achieve what you're after within that philosophy, we'd love to hear more!"

**Example:**

> Thanks for suggesting this! I can see how showing more tasks would feel more productive. ThreeDoors intentionally limits the view to three tasks — our [SOUL.md](../SOUL.md) says "Three Doors, Not Three Hundred" because the constraint itself is the feature. We've found that limiting choices actually helps people take action by eliminating choice paralysis. That said, if you think there's a way to achieve what you're after within that philosophy, we'd love to hear more!
