# ThreeDoors — Soul Document

## What Is ThreeDoors?

A personal achievement partner disguised as a todo app. It doesn't manage tasks —
it helps humans *do* things by working with their psychology, not against it.

## Core Philosophy

### Progress Over Perfection

Imperfect action beats perfect planning. The whole point of showing three doors
is to get someone moving — not to optimize their todo list. Every design decision
should reduce friction to starting, not add friction for "correctness."

### Work With Human Nature

Choice paralysis is real. Overwhelm is real. The "I'll do it later" trap is real.
ThreeDoors exists because traditional todo apps ignore these realities. Features
should acknowledge that humans are not rational task-processing machines.

### Three Doors, Not Three Hundred

Show 3 tasks. Not 5. Not "all of them with filters." The constraint IS the feature.
When in doubt, show less. Resist the urge to add "just one more option."

### Local-First, Privacy-Always

- Data stays on the user's machine. Period.
- No telemetry, no analytics, no phone-home.
- No accounts, no sign-up, no cloud sync (unless the user explicitly configures it).
- Integrations (Apple Notes, Jira, GitHub, etc.) use local APIs or user-provided tokens — never intermediary services.

### Meet Users Where They Are

People already have tasks in Apple Notes, Jira, Linear, text files. ThreeDoors
integrates with existing tools — it doesn't ask users to migrate. The adapter
pattern exists precisely for this: plug into what people already use.

### Solo Dev Reality

This is built by one person in limited hours per week. Every feature must justify
its complexity. Prefer the simple solution that works today over the elegant
solution that takes three sprints. If a feature requires more than one story
to be useful, reconsider the decomposition.

### The Director's Role

The human maintainer's primary job is decision-making, quality review, and
direction-setting — not implementation. With agents handling 96% of PRs,
the bottleneck is never "can we build it fast enough?" — it's "are we building
the right thing?" Agent infrastructure should minimize the decisions that
require human attention, not maximize the work agents can do.

This project was dormant for 4 months (Nov 2025 — Mar 2026). What unlocked
velocity wasn't better tools — it was adopting BMAD as a methodology. The
framework provides the structure that lets agents work autonomously while
keeping the human in control of direction.

## Design Principles for AI Agents

When implementing a story and you face a decision not covered by the spec:

1. **Would this reduce friction?** If yes, do it. If it adds a step for the user, don't.
2. **Is this the simplest thing that works?** Ship that. Refactor later if needed.
3. **Does this respect the user's data?** No silent writes, no data loss, atomic operations.
4. **Does this follow existing patterns?** Check how similar things are done in the codebase.
   Don't invent new patterns when existing ones work.
5. **Would the user notice this?** If it's an internal refactor with no visible change,
   keep the scope minimal. If it's user-facing, make it feel effortless.

## What ThreeDoors Is NOT

- Not a project management tool (no Gantt charts, no sprints, no team features)
- Not a habit tracker (no streaks, no gamification, no guilt)
- Not a second brain (no knowledge graph, no linking, no tagging taxonomy)
- Not trying to be everything to everyone — it's a personal tool for one person at a time

## Every Interaction Should Feel Deliberate

Keypresses should produce visible, satisfying responses. The UI should feel like
physical objects — doors that open, selections that click into place. Subtle is
not the same as invisible. When the user acts, the UI should answer unmistakably.

The difference between "I clicked a flat screen" and "I pressed a physical button"
is the difference between adequate and delightful. We want the button feel.

## The Feeling We're Going For

Opening ThreeDoors should feel like a friend saying: "Hey, here are three things
you could do right now. Pick one. Any one. Let's go."

Not: "You have 47 overdue tasks. Here's a productivity report."
