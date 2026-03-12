# Core Concepts

## The Three Doors Philosophy

Traditional task lists create choice paralysis. When you have 50+ tasks, picking one becomes its own task. ThreeDoors solves this by presenting exactly three options — pick one, take action, move on.

!!! quote "Progress over perfection"
    Imperfect action beats perfect planning. The whole point of showing three doors is to get someone moving — not to optimize their todo list.

### Behavioral Science Foundation

The three-door constraint is grounded in research:

- **Choice overload** (Iyengar & Lepper) — Too many options reduces satisfaction and action
- **Cognitive capacity** (Cowan) — Working memory holds 3–5 chunks; three is the sweet spot
- **Decision fatigue** (Baumeister) — Fewer decisions preserves energy for actual work
- **Hick's Law** — Response time scales with the number of options; three minimizes decision latency

### Progress Over Perfection

ThreeDoors doesn't try to find the "optimal" task. It presents three reasonable options and trusts you to pick. Any forward motion beats standing still.

## How Door Selection Works

When you open ThreeDoors, three tasks are randomly selected from your pool. The selection algorithm:

- Excludes blocked, deferred, and archived tasks
- Avoids showing recently-displayed tasks
- Prefers diversity across task types and effort levels
- If you've logged a mood, biases toward tasks that correlate with productivity for that mood (based on your history)

Don't like your options? Press ++s++ or ++down++ to refresh and get three new doors.

## Key Principles

### Local-First, Privacy-Always

- Data stays on your machine — no telemetry, no analytics, no phone-home
- No accounts, no sign-up, no cloud sync (unless you explicitly configure it)
- Integrations use local APIs or user-provided tokens — never intermediary services

### Meet Users Where They Are

People already have tasks in Apple Notes, Jira, text files, and more. ThreeDoors integrates with existing tools through its [provider system](../providers/overview.md) — it doesn't ask you to migrate.

### Three Doors, Not Three Hundred

Show 3 tasks. Not 5. Not "all of them with filters." The constraint IS the feature. When in doubt, show less.
