# User Interface Design Goals

## Overall UX Vision

ThreeDoors presents as a conversational partner rather than a demanding taskmaster. The central interface metaphor is literal: **three doors, three tasks, three different on-ramps to action**. At each session start, the user is presented with three carefully selected tasks that are very different from each other—different types of activities, different effort levels, different contexts—but all represent good starting points based on priorities. This design serves dual purposes: it gets the user in the habit of doing *something* (reducing inertia), and it teaches the tool about the user's current state by observing which types of tasks they gravitate toward or avoid.

The interface should feel like opening a dialogue, not confronting a backlog. Users are greeted with options that respect their current capacity—whether focused, overwhelmed, or stuck—and celebrate any choice as progress.

## Key Interaction Paradigms

**The Three Doors (Primary Interaction):**
The main interface presents three tasks simultaneously as entry points. These tasks should be:
- **Intentionally diverse** - Different types of activities (e.g., creative vs. administrative vs. physical, or high-focus vs. low-friction vs. context-switching)
- **Small at first** - Especially in early usage, doors should present approachable tasks to build momentum
- **All viable next steps** - Each represents a legitimate priority, not filler options
- **Learning opportunities** - User's choice (or avoidance) informs the system about current mood, energy, and capacity

Over time, the system learns: "On Tuesday mornings, user picks Door 1 (focused work). On Friday afternoons, user picks Door 3 (quick wins). User never picks administrative tasks before 10am."

**Door Selection Feedback (Interaction Feel):**
- **Strong visual response** - Every keypress must produce a visible, satisfying response. The selected door should draw the eye through contrast, scale, or emphasis — not just a minor color shift. Unselected doors should visually recede (dim/fade) to create a focus funnel effect.
- **Reversibility** - Pressing the same selection key again toggles the door off, returning to neutral state. Users who know they can undo a selection explore more freely.
- **Consistent escape routes** - 'q' (quit) works from every non-text-input view. Users should never feel trapped in a screen.

**Door Refresh & Feedback (MVP Core):**
- **Refresh/New Doors** - Simple keystroke ('s' or 'down arrow') to generate a new set of three doors if current options don't appeal. No judgment, no friction—just new options.
- **Task Management Feedback** - Options to indicate task status or modify tasks:
  - 'c': Mark selected task as complete.
  - 'b': Mark selected task as blocked.
  - 'i': Mark selected task as in progress.
  - 'e': Expand selected task (into more tasks).
  - 'f': Fork selected task (clone/split).
  - 'p': Procrastinate/avoid selected task.
  - "Blocked" - Task cannot proceed (captures blocker)
  - "Not now" - Task is valid but doesn't fit current mood/context (teaches system about state)
  - "Needs breakdown" - Task is too big/unclear (MVP: flag for later attention; Post-MVP: may trigger breakdown assistance)
  - "Other comment" - Freeform note about the task (refactoring, context, etc.)

These interactions serve dual purposes: give users control (preventing feeling trapped) and provide rich learning signal to the system about task suitability, blockers, and user state.

**Choose-Your-Own-Adventure Navigation:**
Beyond the three doors, other decision points present 3-5 contextual options rather than requiring command memorization. Options adapt based on state and history.

**Progressive Disclosure:**
Start simple, reveal complexity only when needed. Quick add mode for speed, expanded capture for context when desired. Don't force decisions upfront.

**Persistent Context:**
Values/goals remain visible (but unobtrusive) throughout the session—likely as a subtle header or footer—reminding users of the "why" while working on the "what."

**Encouraging Tone:**
All messaging embodies "progress over perfection." Copy celebrates any action ("You picked a door and started. That's what matters.").

## Core Screens and Views

From a product perspective, these are the critical views necessary to deliver MVP value:

1. **Three Doors Dashboard (Primary Interface)** - Session entry point presenting three diverse tasks as "doors" to choose, with minimal surrounding context. Core question: "Which door feels right today?" Includes refresh option and per-door feedback mechanism.

2. **Task List View** - Full task display when user wants to see beyond the three doors, with filtering and status

3. **Quick Add Flow** - Minimal-friction task capture (possibly single input field)

4. **Extended Capture Flow** - Optional deeper capture including "why" context and task metadata (effort, type, context)

5. **Values/Goals Setup** - Initial and ongoing management of user-defined values that guide prioritization

6. **Progress View** - Visualization showing "better than yesterday" metrics and door choice patterns over time (e.g., "You've opened 5 doors this week, up from 3 last week" and "You tend to pick Door 1 in mornings, Door 3 in afternoons")

7. **Health Check View** - Diagnostic display showing Apple Notes connectivity and sync status

8. **Improvement Prompt** - End-of-session single question asking for one improvement

## Accessibility

**None** - MVP focuses on terminal interface for single user (developer). Accessibility requirements deferred to future phases when/if user base expands beyond CLI-comfortable users.

## Branding

**Terminal Aesthetic with Warmth:**
Leverage Charm Bracelet/Bubbletea's capabilities for styled terminal UI—think clean, readable typography with subtle use of color for status indication (green for progress, yellow for prompts, red sparingly for errors).

**Three Doors Visual Metaphor:**
The main interface will render three visual "doors" arranged horizontally. Doors use **portrait-oriented proportions** (taller than wide) to visually read as actual doors rather than cards or panels. Each door includes door-like signifiers: a lintel/header with door number, panel dividers creating upper/lower sections, an asymmetric handle on the right side at ~60% height, and a threshold/floor line at the bottom. Doors dynamically adjust their width based on terminal width and their height to 60% of terminal height (minimum 12 lines), with content vertically centered within each door. This makes the interface feel immersive and reinforces the door metaphor at any terminal size.
```
┌──── [1] ────┐  ┌──── [2] ────┐  ┌──── [3] ────┐
│             │  │             │  │             │
│  [Task A]   │  │  [Task B]   │  │  [Task C]   │
│  Quick win  │  │  Deep work  │  │  Creative   │
│  ~5min      │  │  ~30min     │  │  ~15min     │
│             │  │             │  │             │
├─────────────┤  ├─────────────┤  ├─────────────┤
│             │  │             │  │             │
│          ●  │  │          ●  │  │          ●  │
│             │  │             │  │             │
│             │  │             │  │             │
│             │  │             │  │             │
└─────────────┘  └─────────────┘  └─────────────┘
  ▔▔▔▔▔▔▔▔▔▔▔      ▔▔▔▔▔▔▔▔▔▔▔      ▔▔▔▔▔▔▔▔▔▔▔
  (Portrait orientation, min 12 lines, panel dividers, handles, thresholds)

Press A/Left, W/Up, or D/Right to select  |  S/Down to re-roll  |  Q to quit
Press C (complete), B (blocked), I (in progress), E (expand), F (fork), P (procrastinate) for task actions
```

When terminal height is insufficient for portrait proportions (< 16 rows), doors gracefully degrade to compact card layout while preserving full task readability.

**"Progress Over Perfection" Visual Language:**
Use asymmetry, incomplete progress bars, and "good enough" indicators. The three doors might be slightly different sizes or styles, reinforcing that perfection isn't required—just pick one and start.

## Target Device and Platforms

**Primary: macOS Terminal Emulators (iTerm2, Terminal.app, Alacritty)**
- CLI/TUI optimized for 80x24 minimum, responsive to larger terminal sizes
- Assumes modern terminal with 256-color support minimum
- Keyboard-driven navigation (keys 'a'/'left', 'w'/'up', 'd'/'right' for door selection; 's'/'down' for re-roll; 'q' to quit; 'c', 'b', 'i', 'e', 'f', 'p' for task actions)

**Secondary: Remote Terminal Access**
- Should function over SSH connections (for future Geodesic/remote environment access)
- ASCII fallback for constrained environments

**Mobile Access (Indirect):**
- No dedicated mobile UI in MVP
- Mobile interaction happens through Apple Notes app directly (view/edit tasks on iPhone)
- Sync bidirectionally when user returns to terminal interface

---
