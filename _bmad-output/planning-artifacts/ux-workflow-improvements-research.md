# UX & Workflow Improvements Research

## Purpose

Research into how to make ThreeDoors more useful and sticky as a daily productivity tool. Analyzes the current feature set, reviews productivity UX research, and ranks potential improvements by impact vs effort.

## Current State Assessment

ThreeDoors already has a strong feature set beyond its core "3 doors" mechanic:

| Category | Features |
|----------|----------|
| **Core** | 3-door selection with diversity scoring, 7-status task lifecycle, inline tag parsing (#type @effort +location) |
| **Intelligence** | Avoidance detection (5+ bypasses), calendar-aware door selection (effort vs available time), pattern analysis from session history |
| **Engagement** | Mood tracking (7 presets + custom), session reflection prompts, values/goals footer, next-steps flow after completion |
| **Power User** | Search + command palette (`:` prefix), task cross-references, multi-provider backends (textfile/obsidian/applenotes), WAL sync |
| **Insights** | 7-day sparkline trends, mood-completion correlations, door position bias detection, streak tracking |
| **Onboarding** | 5-step wizard: welcome, keybinding tutorial, values setup, task import, reference |

### What's Working

The core "3 doors" constraint directly addresses backlog horror and decision fatigue — the #1 reason task managers get abandoned (73% churn within 30 days). The diversity scoring, avoidance detection, and calendar-aware selection add intelligence without complexity.

### What's Missing

The app currently operates as a **reactive** tool (open it, pick a door, close it). It lacks the **proactive** rituals and rhythms that make tools like Sunsama sticky — daily planning, focus sessions, quick capture from outside the TUI.

---

## Why Task Managers Get Abandoned

Research identifies these failure modes:

1. **Setup tax** — teams average 8 hours per person setting up productivity systems before abandoning them
2. **Feature overwhelm** — 67% of users feel stressed by apps with too many features; the tool becomes meta-work
3. **Maintenance burden** — one skipped day means the system falls out of sync; re-syncing costs more than the benefit
4. **Mental model mismatch** — imposing GTD/Kanban/etc. on users whose brains work differently
5. **Backlog horror** — adding tasks faster than completing them; seeing 200 overdue items triggers anxiety, not motivation

ThreeDoors already mitigates #2 and #5 by design. The improvements below target #3 (maintenance burden) and add proactive engagement loops.

---

## Improvement Proposals — Ranked by Impact vs Effort

### Tier 1: High Impact, Low-Medium Effort

#### 1. Quick Capture CLI (non-interactive add)

**Problem:** Adding tasks requires opening the TUI, pressing `:add`, typing, confirming. This friction means tasks get lost — the #1 failure mode for task managers.

**Proposal:** `threedoors add "review PR from alex" #technical @quick-win` works from any terminal, no TUI session needed. Support stdin pipe: `echo "investigate build failure" | threedoors add`.

**What exists:** The TUI has `:add` and `:add-ctx` commands, plus inline tag parsing. The core `AddTask()` logic exists.

**Effort:** Low — wire existing `AddTask` logic to a CLI subcommand in `cmd/threedoors/`. Parse flags for `--why` context.

**Impact:** High — removes the biggest friction point. Every task captured is a task that won't be forgotten.

---

#### 2. Daily Planning Mode

**Problem:** Users open ThreeDoors reactively ("what should I do?") rather than proactively ("what am I committing to today?"). Without a planning ritual, the tool lacks the daily engagement hook that makes Sunsama's 95% retention rate possible.

**Proposal:** A `threedoors plan` command or `:plan` TUI command that walks through:
1. **Review** — show yesterday's incomplete tasks, ask: continue / defer / drop each
2. **Select** — from the full pool, pick 3-5 tasks as "today's focus" (these get priority in door selection)
3. **Energy check** — "How's your energy?" (high/medium/low) — filters today's focus to match

The planning session should be time-bounded (5-10 minutes) with a progress indicator.

**What exists:** Values/goals onboarding flow demonstrates guided multi-step TUI flows. Mood capture shows energy-adjacent UX. Session metrics track daily engagement.

**Effort:** Medium — new view + "today's focus" concept in task model (could be a tag or transient state).

**Impact:** High — daily planning rituals are the strongest retention mechanism in productivity tools. Creates a morning habit around ThreeDoors.

---

#### 3. Snooze/Defer as First-Class Action

**Problem:** When a task appears in the doors but isn't actionable today, the user's options are limited: give feedback ("not now"), re-roll, or just ignore it. The task will keep reappearing. This creates "I can't do this now but don't want to lose it" anxiety.

**Proposal:** `S` key on a door → quick defer with options: "Tomorrow", "Next week", "Pick date", "Someday". Deferred tasks disappear from door selection until their return date. A `:deferred` command shows what's snoozed.

**What exists:** The `deferred` status exists in the task model but isn't well-surfaced in the UX. Door feedback has "not now" but it doesn't actually remove the task from rotation.

**Effort:** Low — add a defer date field to tasks, filter deferred tasks in `GetAvailableForDoors()`, add a simple date-picker TUI.

**Impact:** High — Akiflow's "snooze to tomorrow" is one of its most-used features. Reduces noise in the door pool and gives users honest control over timing.

---

#### 4. Task Dependencies (Blocked-Task Filtering)

**Problem:** A task that depends on another task shouldn't appear in the doors if its prerequisite isn't done. Currently, "blocked" status requires manual flagging — the system doesn't understand task relationships structurally.

**Proposal:** Add `depends_on` field to tasks. Tasks whose dependencies aren't complete are automatically filtered from door selection. Show a "blocked by: [task]" indicator. When a dependency completes, its dependents become "unblocked" (could trigger a notification/flash).

**What exists:** Cross-references exist in the enrichment DB (`cross_references` table) but track generic "related" relationships, not dependency chains. The `blocked` status exists but is manually set.

**Effort:** Medium — extend task model, modify `GetAvailableForDoors()` filter, add dependency resolution logic.

**Impact:** High — Taskwarrior's dependency system (coefficient 8.0 in urgency formula) demonstrates that filtering blocked tasks is foundational for trust. Users need to know the 3 doors are all *actionable*.

---

### Tier 2: High Impact, Medium-High Effort

#### 5. Focus Timer (Pomodoro/Flow Sessions)

**Problem:** Selecting a task is only half the battle. Users also struggle with sustained focus on the selected task. A 2025 meta-analysis found Pomodoro-structured work "consistently improved focus, reduced mental fatigue, and enhanced sustained task performance."

**Proposal:** After selecting a door, option to "Start focus session" → countdown timer (default 25 min, configurable 25/50/90). During the session: task name + timer displayed, keyboard locked except for pause/stop. On timer completion: "Completed? / Continue? / Switch?" prompt. Session time tracked in metrics.

**What exists:** Session metrics already track duration. The TUI infrastructure supports modal views.

**Effort:** Medium — new timer view, tick-based countdown via `tea.Tick`, session tracking integration.

**Impact:** High — transforms ThreeDoors from "task picker" to "focus companion." The timer creates a natural rhythm and makes sessions measurable. Developer note: 25 minutes is often too short for coding — the 50/90 minute options are essential.

---

#### 6. Energy-Level Matching

**Problem:** A high-energy deep-work task at 2pm post-lunch is a recipe for procrastination. The current effort tags (quick-win/medium/deep-work) exist but aren't connected to the user's current state.

**Proposal:** Prompt for energy level at session start (or infer from time of day as a default). Filter door selection to prefer tasks matching current energy. Show energy indicator on doors: `[high energy]` tag. Allow override ("show me everything").

**What exists:** Effort tags on tasks, mood tracking (energy-adjacent), calendar-aware time context scoring.

**Effort:** Medium — energy prompt view, extend door selection scoring to include energy match, time-of-day inference heuristic.

**Impact:** Medium-High — research consistently shows demanding tasks in peak energy windows produce better outcomes. Simple heuristic, large behavioral impact.

---

#### 7. Momentum Scoring

**Problem:** Returning to a task you worked on yesterday has lower cognitive startup cost than picking up something cold. The current diversity scoring doesn't account for task recency-of-interaction.

**Proposal:** Factor "last touched" timestamp into door selection scoring. Tasks with recent notes, status changes, or detail views get a small scoring boost. This favors continuation over context-switching.

**What exists:** Session metrics record detail views and status changes per task. The enrichment DB tracks task metadata with timestamps.

**Effort:** Low — add a "last_interacted" timestamp to tasks, include a small momentum bonus in `DiversityScore()` or `SelectDoors()`.

**Impact:** Medium — reduces context-switching overhead. Especially valuable for multi-day tasks where maintaining flow matters.

---

### Tier 3: Medium Impact, Variable Effort

#### 8. Completion Velocity (Replace Streak Anxiety)

**Problem:** Streak tracking (current: consecutive completion days) creates anxiety when life interrupts. A broken streak can trigger tool abandonment rather than re-engagement.

**Proposal:** Replace streak emphasis with "completion velocity": tasks completed this week vs last week, shown as a direction arrow and ratio. `This week: 12 tasks (+33% vs last week)`. No penalty for missed days — just honest throughput measurement.

**What exists:** Insights dashboard has sparkline trends and streak tracking. Session metrics have all the data needed.

**Effort:** Low — modify insights view to emphasize velocity over streaks.

**Impact:** Medium — psychologically healthier metric that encourages without punishing. Aligned with research showing velocity tracking > streak tracking for long-term retention.

---

#### 9. Batch Triage Mode

**Problem:** After importing tasks or returning from a break, there may be 20+ tasks that need classification (effort, type, location). Going through them one-by-one in the detail view is tedious.

**Proposal:** `:triage` command shows tasks one at a time in a rapid-fire format: task text → press `q/m/d` for quick-win/medium/deep-work → press `c/a/t/p` for creative/admin/technical/physical → next task. Skip with `Enter`. Progress bar shows completion.

**What exists:** Inline tag parsing handles classification at add-time. This is for retroactive/bulk classification.

**Effort:** Medium — new triage view with rapid-fire UX.

**Impact:** Medium — reduces maintenance burden (#3 abandonment reason). Makes it easy to keep the task pool well-categorized.

---

#### 10. Smart Scheduling Heuristics

**Problem:** The current door selection uses diversity + time-context scoring but doesn't consider urgency, deadline proximity, or dependency bottlenecks.

**Proposal:** Add a lightweight urgency score inspired by Taskwarrior:
- **Due date proximity** — nonlinear curve (due in 2 days >> due in 7 days >>> due in 30 days)
- **Blocking score** — tasks that block other tasks get a bonus
- **Explicit priority** — a user-set `+next` tag that dominates scoring (like Taskwarrior's 15.0 coefficient)
- **Age creep** — old tasks slowly rise in score (0.01/day) to prevent permanent stagnation

**What exists:** Diversity scoring, time-context scoring, effort tags.

**Effort:** Medium-High — add due dates to task model, implement urgency formula, integrate with door selection.

**Impact:** Medium — makes the 3-door selection feel genuinely intelligent rather than random-with-diversity.

---

#### 11. Time Estimates with Time-Fit Filtering

**Problem:** A 4-hour deep-work task appearing when you have 20 minutes before a meeting wastes a door slot.

**Proposal:** Add rough time estimates to tasks (15m/30m/1h/2h/4h or small/medium/large). When calendar data shows limited time, filter doors to tasks that fit the available window.

**What exists:** Calendar integration with `TimeContext` already provides available time data. Effort tags provide rough size proxies.

**Effort:** Low-Medium — could extend existing effort tags or add explicit duration field. The time-fit filtering logic partially exists in `TimeContextScore()`.

**Impact:** Medium — makes calendar-aware selection more precise. Especially useful for users with meeting-heavy schedules.

---

### Tier 4: Lower Impact or Speculative

#### 12. Habit/Recurring Task Support

**Problem:** Some tasks recur daily/weekly (exercise, review PRs, clean inbox). Currently these must be manually re-added.

**Proposal:** Recurring task templates: `threedoors add "review PRs" --recur daily`. Completed recurring tasks auto-create the next instance.

**Effort:** Medium — recurrence engine, template storage, auto-creation logic.

**Impact:** Medium — important for users who mix one-off and recurring tasks. Risk: recurring tasks can dominate the pool if not balanced.

---

#### 13. Session Bookmarks / "Continue Where I Left Off"

**Problem:** When returning to ThreeDoors after a break, the user has no memory of what they were working on.

**Proposal:** On startup, if a previous session had an active task (viewed in detail but not completed), offer to resume: "Last session you were working on [task]. Continue?"

**Effort:** Low — read last session from `sessions.jsonl`, check if the task is still active.

**Impact:** Low-Medium — reduces re-orientation time. Small quality-of-life improvement.

---

#### 14. Global Hotkey / Background Daemon

**Problem:** Quick capture requires switching to a terminal and typing a command. A system-level hotkey would make capture instant.

**Proposal:** A background daemon (`threedoors daemon`) that listens for a configurable system hotkey, pops up a minimal capture overlay.

**Effort:** High — OS-level hotkey registration (platform-specific), daemon process management, IPC.

**Impact:** Medium — reduces capture friction further, but the CLI `add` subcommand covers 80% of the use case.

---

## Summary Matrix

| # | Improvement | Impact | Effort | Priority |
|---|------------|--------|--------|----------|
| 1 | Quick Capture CLI | High | Low | **P0** |
| 2 | Daily Planning Mode | High | Medium | **P0** |
| 3 | Snooze/Defer | High | Low | **P0** |
| 4 | Task Dependencies | High | Medium | **P1** |
| 5 | Focus Timer | High | Medium | **P1** |
| 6 | Energy-Level Matching | Medium-High | Medium | **P1** |
| 7 | Momentum Scoring | Medium | Low | **P2** |
| 8 | Completion Velocity | Medium | Low | **P2** |
| 9 | Batch Triage Mode | Medium | Medium | **P2** |
| 10 | Smart Scheduling | Medium | Medium-High | **P2** |
| 11 | Time Estimates | Medium | Low-Medium | **P2** |
| 12 | Recurring Tasks | Medium | Medium | **P3** |
| 13 | Session Bookmarks | Low-Medium | Low | **P3** |
| 14 | Global Hotkey | Medium | High | **P3** |

## Recommended Implementation Order

**Phase 1 — Daily Engagement Loop (P0)**
1. Quick Capture CLI — lowest effort, highest friction removal
2. Snooze/Defer — makes the door pool trustworthy
3. Daily Planning Mode — creates the morning habit

**Phase 2 — Focus & Intelligence (P1)**
4. Focus Timer — transforms ThreeDoors into a focus companion
5. Task Dependencies — makes doors consistently actionable
6. Energy-Level Matching — personalizes door selection

**Phase 3 — Polish & Depth (P2)**
7. Momentum Scoring + Completion Velocity — better engagement metrics
8. Smart Scheduling — intelligent door selection
9. Batch Triage — reduces maintenance burden

**Phase 4 — Power Features (P3)**
10. Recurring Tasks, Session Bookmarks, Global Hotkey

## Key Design Principles

Based on the research, any additions should follow these principles:

1. **Capture is sacred** — zero-friction task capture matters more than any other feature
2. **Show don't ask** — suggest a default, let the user override (reduces decisions)
3. **Rituals over features** — a daily planning flow creates more retention than 10 new buttons
4. **Velocity over streaks** — measure throughput without punishing breaks
5. **Ready means ready** — the 3 doors should only show tasks the user can act on right now
6. **Time-box everything** — planning sessions, focus sessions, triage — all should have a timer to prevent the tool from becoming the work

## References

- Sunsama daily planning ritual methodology
- Taskwarrior urgency scoring formula and dependency system
- Akiflow command bar and snooze patterns
- Things 3 opinionated simplicity design
- Pomodoro Technique meta-analysis (2025) on focus and sustained task performance
- Nielsen Norman Group: minimize cognitive load for usability
- Duolingo/Plotline research on streaks and gamification retention (35% churn reduction with streaks + milestones, but streak anxiety as counter-risk)
- Motion/Reclaim AI automatic scheduling approaches
- GTD "trusted system" and inbox capture principles
