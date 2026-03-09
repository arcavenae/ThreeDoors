# Research Report: Making Doors More Door-Like

**Date:** 2026-03-09
**Method:** 5-round BMAD Party Mode with 7 agents
**Agents:** Innovation Strategist, Design Thinking Coach, Creative Problem Solver, Analyst, UX Designer, Dev, Architect
**Status:** Research complete — no implementation code produced

---

## Executive Summary

ThreeDoors' door metaphor works *cognitively* (users understand "pick a door") but not *viscerally*. Current doors render as rectangular boxes with borders — they read as cards/panels, not doors. Through 5 rounds of multi-agent deliberation, we identified 9 concrete proposals, scored them on feasibility, visual impact, SOUL.md alignment, and width cost, and converged on 5 adopted recommendations that collectively transform boxes into doors.

**The key insight:** A door is not just a shape — it's defined by **asymmetry** (hinge side vs opening side), **hardware** (a handle you can grab), **behavior** (it opens, it's ajar, it closes), and **context** (it exists in a wall, on a floor). Current rendering captures almost none of these.

---

## Current State Assessment

### Door Rendering Architecture
- **Framework:** Bubbletea + Lipgloss
- **Themes:** Classic, Modern, Sci-Fi, Shoji (default: Modern)
- **Anatomy System:** `DoorAnatomy` struct calculates LintelRow, ContentStart, PanelDivider, HandleRow (60% height), ThresholdRow
- **Animation:** Spring physics (harmonica) for selection emphasis, 60 FPS, Lab-space color interpolation
- **Shadows:** Half-block characters (▐, █, ▄) in `themes/shadow.go`
- **Minimum Width:** 15 characters per door
- **Height:** 60% of terminal height, min 10 rows

### "Doorness" Score by Theme (7 dimensions)

| Dimension | Classic | Modern | Sci-Fi | Shoji | Description |
|-----------|---------|--------|--------|-------|-------------|
| Frame (doorway) | ✅ | ⚠️ | ✅ | ✅ | The architectural opening |
| Panel (door leaf) | ⚠️ | ❌ | ❌ | ⚠️ | The thing that moves |
| Hardware (handle) | ✅ | ⚠️ | ❌ | ✅ | Affordance for interaction |
| Hinge | ❌ | ❌ | ❌ | ❌ | Asymmetric mounting point |
| Threshold | ✅ | ❌ | ⚠️ | ❌ | Floor line / boundary |
| What's Behind | ❌ | ❌ | ❌ | ❌ | Mystery / unknown |
| Gap/Depth | ⚠️ | ⚠️ | ✅ | ❌ | Suggests 3D depth |
| **Score** | **3.5/7** | **1/7** | **2.5/7** | **2/7** | |

**Key gaps:** No theme has hinges (0/4), no theme shows what's behind (0/4), handle placement is inline with content rather than at the edge.

### Current Door Rendering (Classic Theme, Approximate)

```
╭──────────────────────────╮
│ [todo]                   │
│                          │
│ Buy groceries            │
│                          │
│ 🛒 quick @home           │
│                          │
├──────────────────────────┤
│                          │
│            ●             │  ← handle centered (wrong!)
│                          │
│                          │
│▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔│
╰──────────────────────────╯▐
                            ▄  ← shadow
```

---

## Round 1: Current State Analysis

### Key Findings
1. Current doors score ~3.5/7 on "doorness" dimensions at best (Classic theme)
2. Biggest gaps: hinge (0/4 themes), what's behind (0/4), panel distinction (0.5/4)
3. Handle placement is anatomically correct (60% height) but rendered wrong (centered in content flow, not side-mounted)
4. The theme system architecture supports richer door rendering without interface changes
5. Width constraint (min 15 chars) requires graceful degradation
6. The vertical proportion (0.6 height ratio) already helps with door-like aspect ratio

### What Makes Something "Feel Like a Door" (7 Dimensions)
1. **Frame** — the doorway itself (the architectural opening)
2. **Panel** — the door leaf (the thing that moves)
3. **Hardware** — handle/knob (the affordance for interaction)
4. **Hinge** — asymmetric mounting point (implies opening direction)
5. **Threshold** — the floor line (boundary marker)
6. **What's behind** — the mystery (the whole point of a door)
7. **Gap/crack** — light under the door, or around edges (suggests depth)

---

## Round 2: Terminal UI Visual Possibilities

### Unicode Character Catalog for Door Rendering

**Box Drawing (U+2500–U+257F):**
- Light/heavy/double lines: `┌─┐│└─┘` vs `┏━┓┃┗━┛` vs `╔═╗║╚═╝`
- Three distinct visual weights — perfect for frame-within-frame

**Block Elements (U+2580–U+259F):**
- `█▓▒░` — full, dark, medium, light shade — depth gradient potential
- `▀▄▌▐` — half blocks — subcharacter precision for bevels/edges

**Geometric Shapes:**
- `●○◉◎` — doorknobs at different states
- `◐◑` — half-filled circles for handle turning animation

### Key Technique: Nested Frame

```
╔═══════════════════╗   ← Heavy frame (doorframe)
║ ┌───────────────┐ ║   ← Light frame (door panel)
║ │               │ ║
║ │  Task text    │ ║
║ │               │ ● ← Handle at right edge
║ │               │ ║
║ └───────────────┘ ║
╚═══════════════════╝
```

### Animation Capabilities
1. **Spring physics** — already implemented. Can drive door "swing," handle "turn," panel "recede"
2. **Frame-by-frame character swap** — knob turning: `●` → `◐` → `○` → `◑` → `●`
3. **Color interpolation** — Lab-space blending for light spill, wood grain tint
4. **Braille patterns (U+2800–U+28FF)** — subpixel detail but compatibility/accessibility concerns. **Recommendation: avoid.**

### Key Findings
- Unicode gives us 3 weight levels for frame-within-frame
- Block elements can create depth, texture, wall context
- Spring physics can drive opening/closing animations
- Width budget: wall texture costs ~4 chars; nested frame ~4 more; leaves ~7 for content at min width
- Rendering pipeline has no constraints on character complexity — same cost per cell

---

## Round 3: Door Interaction Metaphors — How Doors BEHAVE

### The 5-Stage Door Interaction Journey
1. **Approach** — See three doors, closed. *Curiosity*
2. **Choose** — Touch the handle. *Commitment*
3. **Turn** — Handle rotates, mechanism engages. *Anticipation*
4. **Open** — Door swings, light spills out. *Revelation*
5. **Step through** — Cross threshold, committed. *Action*

**Current app covers:** Stages 1, 2, 5. **Completely missing:** Stages 3 and 4.

### Interaction Metaphor Proposals

**A. Crack of Light** — On selection, thin line of "light" at door edge:
```
Before:           After selection:
┌─────────┐       ┌─────────╮
│         │       │         ╎░
│  Task   │  →    │  Task   ╎░
│         │       │         ╎░
└─────────┘       └─────────╯
```

**B. Handle Turn** — On selection, handle character animates:
```
Frame 1: ──●    (at rest)
Frame 2: ──◐    (turning)
Frame 3: ──○    (turned)
Frame 4: ──◑    (spring back if deselected)
```

**C. Door Swing** — On Enter, perspective animation:
```
Frame 1:        Frame 3:        Frame 5:
┌────────┐      ╲        ┐      Detail
│        │       ╲  Task │      view
│  Task  │        ╲     │      fills
│        │         ╲   │       the
└────────┘          ╲─┘       space
```

**D. The Peek** — Side panel showing task details when door is "cracked":
```
Selected (peeking):
┌──────┐╎
│      │╎ "Buy groceries"
│ Task │╎ "Est: 15min"
│      │╎ "Priority: low"
└──────┘╎
```

### Key Findings
- Door interaction is a 5-stage journey; currently missing stages 3 and 4
- Crack of Light is highest value-to-cost ratio (1 char width, immediate feedback)
- Handle Turn is trivial to implement, high satisfaction
- Door Swing is moderate complexity with text reflow risk
- Peek requires significant layout rework — better as separate feature
- Space/Enter toggle (Story 36.4) maps naturally to reversing open animation
- State machine: DoorClosed → DoorAjar → DoorOpening → DoorOpen

---

## Round 4: Convergence — All Proposals Scored

### Complete Proposal List

| Proposal | Feasibility (1-5) | Visual Impact (1-5) | SOUL Alignment (1-5) | Width Cost | **TOTAL** |
|----------|-------------------|--------------------|-----------------------|------------|-----------|
| A. Nested Frame | 5 | 4 | 4 | -4 chars | **13** |
| B. Side Handle | 5 | 4 | 5 | 0 chars | **14** |
| C. Crack of Light | 5 | 5 | 5 | -1 char | **15** ⭐ |
| D. Handle Turn | 5 | 3 | 5 | 0 chars | **13** |
| E. Wall Context | 3 | 5 | 3 | -4 chars | **11** |
| F. Threshold Line | 5 | 3 | 4 | 0 chars | **12** |
| G. Hinge Marks | 4 | 3 | 4 | 0 chars | **11** |
| H. Door Swing | 2 | 5 | 3 | N/A | **10** |
| I. Light Spill | 3 | 4 | 4 | -2 chars | **11** |

### Mockups for Each Proposal

**A. Nested Frame:**
```
╔══════════════════╗
║ ┌──────────────┐ ║
║ │  Buy milk    │ ║
║ │         [w]● │ ║
║ └──────────────┘ ║
╚══════════════════╝
```

**B. Side Handle:**
```
┌──────────────┐
│  Buy milk    │
│              ●─  ← handle at right edge
└──────────────┘
```

**C. Crack of Light:**
```
Unselected:           Selected:
┌──────────┐          ┌──────────╮
│  Task    │    →     │  Task    ╎░
└──────────┘          └──────────╯
```

**E. Wall Context:**
```
▒╔═══════╗▒▒╔═══════╗▒▒╔═══════╗▒
▒║ Door  ║▒▒║ Door  ║▒▒║ Door  ║▒
▒╚═══════╝▒▒╚═══════╝▒▒╚═══════╝▒
▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒
```

**F. Threshold Line:**
```
 ┌─────┐   ┌─────┐   ┌─────┐
 │     │   │     │   │     │
 └─────┘   └─────┘   └─────┘
 ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

**G. Hinge Marks:**
```
╓──────────────┐
║  Buy milk    │
╟              ●─
╙──────────────┘
```

---

## Round 5: Final Deliberation — Adopted Recommendations

### ⭐ Recommendation 1: Side-Mounted Door Handle (Proposal B)

**Why:** The single most impactful static change. Moves the doorknob from inline content to the right edge of the panel at HandleRow. Makes every door immediately read as "door" rather than "card." Zero animation complexity, zero width cost.

```
Current:              Proposed:
┌──────────────┐      ┌──────────────┐
│              │      │              │
│  Buy milk    │      │  Buy milk    │
│     ●        │  →   │              ●─
│              │      │              │
└──────────────┘      └──────────────┘
```

### ⭐ Recommendation 2: Crack of Light on Selection (Proposal C)

**Why:** Highest-scoring proposal. Delivers immediate, visible feedback when selecting a door. The door becomes *slightly ajar* — the metaphor IS the interaction. Reversible on deselect.

```
Unselected:           Selected:
┌──────────┐          ┌──────────╮
│          │          │          ╎░
│  Task    │    →     │  Task    ╎░
│          │          │          ╎░
└──────────┘          └──────────╯
```

### ⭐ Recommendation 3: Handle Turn Micro-Animation (Proposal D)

**Why:** Delightful micro-interaction. 4-frame handle rotation synced with spring physics. ~20 LOC.

```
Frame 0: ●    (at rest)
Frame 1: ◐    (turning)
Frame 2: ○    (turned — door ajar)
Deselect: ◑ → ●  (spring back)
```

### ⭐ Recommendation 4: Hinge Marks on Left Edge (Proposal G)

**Why:** Creates asymmetry that reads "door." Left edge heavier (hinge side), right lighter (opening side). Zero width cost.

```
╓──────────────┐
║              │
║  Buy milk    │
╟              ●─
║              │
╙──────────────┘
```

### ⭐ Recommendation 5: Continuous Threshold/Floor Line (Proposal F)

**Why:** Grounds all doors in shared physical space. Very low cost (~15 LOC).

```
 ┌─────┐   ┌─────┐   ┌─────┐
 │     │   │     │   │     │
 │     ●─  │     ●─  │     ●─
 │     │   │     │   │     │
 └─────┘   └─────┘   └─────┘
 ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

---

## Rejected Alternatives

| Rejected Proposal | Reason for Rejection |
|---|---|
| **A. Nested Frame** | High width cost (-4 chars) at minimum width. At 15-char doors, only 7 chars for content. Benefits achievable through Hinge Marks + Side Handle at lower cost. *Could revisit for wide terminals only.* |
| **E. Wall Context** | 4 chars per side too expensive. Threshold Line achieves grounding at zero width cost. Conflicts with some themes (Shoji lattice, Modern minimalism). |
| **H. Door Swing** | Highest complexity (~200+ LOC), text reflow risk, adds animation latency = friction. SOUL.md prioritizes reducing friction. *Could be a future spike.* |
| **I. Light Spill** | Per-cell background colors require theme-specific tuning. Crack of Light achieves similar effect more simply. *Could layer on top of C in future.* |

---

## Before & After Mockup

### Before (current):
```
╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮
│ [todo]           │  │ [in-progress]    │  │ [todo]           │
│                  │  │                  │  │                  │
│ Buy groceries    │  │ Write report     │  │ Fix leaky faucet │
│                  │  │                  │  │                  │
│ 🛒 quick @home   │  │ 📋 medium @desk   │  │ 🔧 quick @home   │
│        ●         │  │        ●         │  │        ●         │
│                  │  │                  │  │                  │
╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯
```

### After (all 5 recommendations, no selection):
```
╓──────────────────╮  ╓──────────────────╮  ╓──────────────────╮
║ [todo]           │  ║ [in-progress]    │  ║ [todo]           │
║                  │  ║                  │  ║                  │
║ Buy groceries    │  ║ Write report     │  ║ Fix leaky faucet │
║                  │  ║                  │  ║                  │
║ 🛒 quick @home   │  ║ 📋 medium @desk   │  ║ 🔧 quick @home   │
╟                  ●  ╟                  ●  ╟                  ●
║                  │  ║                  │  ║                  │
╙──────────────────╯  ╙──────────────────╯  ╙──────────────────╯
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

### After (middle door selected — Crack of Light + Handle Turn):
```
╓──────────────────╮  ╓──────────────────╮  ╓──────────────────╮
║ [todo]           │  ║ [in-progress]    ╎░ ║ [todo]           │
║                  │  ║                  ╎░ ║                  │
║ Buy groceries    │  ║ Write report     ╎░ ║ Fix leaky faucet │
║                  │  ║                  ╎░ ║                  │
║ 🛒 quick @home   │  ║ 📋 medium @desk   ╎░ ║ 🔧 quick @home   │
╟                  ●  ╟                  ○░ ╟                  ●
║                  │  ║                  ╎░ ║                  │
╙──────────────────╯  ╙──────────────────╯  ╙──────────────────╯
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

---

## Technical Feasibility Notes

### Files Affected
- `internal/tui/themes/anatomy.go` — Add HingeCol, handle edge placement
- `internal/tui/themes/classic.go` — Hinge marks, side handle
- `internal/tui/themes/modern.go` — Hinge marks, side handle
- `internal/tui/themes/scifi.go` — Hinge marks, side handle
- `internal/tui/themes/shoji.go` — Hinge marks, side handle
- `internal/tui/animation.go` — Handle turn frames, crack-of-light state
- `internal/tui/doors_view.go` — Threshold line, crack-of-light rendering
- `internal/tui/styles.go` — New style constants for crack/threshold

### Constraints
- **Minimum door width:** 15 chars. Crack of Light costs 1 char. All other proposals cost 0.
- **Width fallback:** When doors fall below theme MinWidth, Classic is used. Door-like features should be present in Classic's compact mode.
- **Animation frame rate:** Already 60 FPS via spring physics. Handle turn and crack-of-light use same tick mechanism.
- **Theme independence:** Each theme renders differently. Hinge marks and handle placement need per-theme character choices.
- **Race detector:** All TUI changes must pass `go test -race ./internal/tui/...`

### Architecture Fit
- Theme `Render()` signature already provides all needed state
- `DoorAnatomy` struct already calculates handle position — extend with `HingeCol`
- `DoorAnimation` already tracks per-door emphasis — extend with `openProgress` and handle frame state
- Shadow system is precedent for post-render character decoration

---

## Suggested Epic/Story Breakdown

### Epic: "Door-Like Doors" (needs PM number allocation)

**Story 1: Side-Mounted Handle + Hinge Marks (B+G)**
- Move doorknob from inline content to right edge at HandleRow
- Add double-weight characters on left border for hinge visual
- Update all 4 themes with per-theme hinge/handle characters
- Update anatomy.go with HingeCol calculation
- Tests: visual snapshot tests for each theme
- ~50 LOC per theme, Low risk

**Story 2: Continuous Threshold Line (F)**
- Render floor line after `JoinHorizontal()` in `doors_view.go`
- Line spans full width, uses `▔` or `─` character
- Respect theme colors for threshold
- Tests: verify threshold renders at various widths
- ~15 LOC, Very Low risk

**Story 3: Crack of Light Effect (C)**
- On selection, replace right border chars with crack chars (╎) and append shade (░)
- Synchronize with spring animation emphasis
- Reverse on deselect
- Update animation state machine with crack state
- Tests: animation state transitions, visual output
- ~50 LOC, Low risk
- **Depends on:** Story 1

**Story 4: Handle Turn Micro-Animation (D)**
- 4-frame handle character sequence synced with spring emphasis
- ● (0.0) → ◐ (0.3) → ○ (0.6) → ○ (1.0, settled)
- Reverse on deselect: ○ → ◑ → ●
- Tests: frame selection at various emphasis values
- ~20 LOC, Low risk
- **Depends on:** Story 1

### Dependency Graph
```
Story 1 (Handle + Hinge) ──┬──→ Story 3 (Crack of Light)
                           └──→ Story 4 (Handle Turn)
Story 2 (Threshold) ──────────→ (independent)
```

Stories 1 & 2 can parallelize. Stories 3 & 4 can parallelize after Story 1.

---

## Noted Opportunities (Out of Scope)

1. **Door-Opening Transition Animation** — Perspective shrink when pressing Enter, transitioning to detail view. Completes full 5-stage door interaction journey. ~200+ LOC, high complexity. Separate spike/epic.

2. **Light Spill Enhancement** — Warm color gradient layered on Crack of Light. Per-cell background colors, theme-specific tuning. Polish story after Crack of Light ships.

3. **Nested Frame for Wide Terminals** — When door width > 30, render frame-within-frame. Width-adaptive. Lower priority.

4. **Wall Context for Wide Terminals** — Shade characters flanking doors when terminal allows. "Doors in a wall" effect.

5. **Per-Theme Door Sound Effects** — If terminal bell/system sounds become feasible. Extremely speculative.

---

## Decision Record

**Decision:** Adopt 5 proposals (B, C, D, F, G) for making doors more door-like. Reject 4 proposals (A, E, H, I).

**Rationale:** Selected proposals maximize door-feel with minimum friction, zero-to-minimal width cost, and low implementation risk (~200-250 LOC total). Rejected proposals either cost too much width (A, E), add too much complexity/friction (H), or are achievable more simply through adopted alternatives (I via C).

**SOUL.md Alignment:** All 5 adopted proposals directly serve:
- "The UI should feel like physical objects — doors that open, selections that click into place"
- "Keypresses should produce visible, satisfying responses"
- "Subtle is not the same as invisible"

**Participants:** Victor (Innovation Strategist), Maya (Design Thinking Coach), Dr. Quinn (Creative Problem Solver), Mary (Analyst), Sally (UX Designer), Amelia (Dev), Winston (Architect)
