# Door Visual Redesign — Party Mode Research

**Date:** 2026-03-11
**Method:** 5-round BMAD Party Mode with 6 agents
**Agents:** Victor (Innovation Strategist), Sally (UX Designer), Winston (Architect), Amelia (Dev), Maya (Design Thinking Coach), Dr. Quinn (Creative Problem Solver)
**Status:** Research complete — no implementation code produced
**Trigger:** User report: "The current doors look bad — the shadow effect isn't working (hard to even perceive as a shadow), and the overall appearance needs significant improvement."

---

## Executive Summary

The current doors suffer from three independent failure modes: **no visual mass** (wireframe interiors), **no depth hierarchy** (shadow too subtle), and **no spatial context** (doors float in void). The shadow system — a single `▐` character in `#585858` on dark backgrounds — has a contrast ratio of ~2.7:1, below the threshold for human perception of an intentional depth cue.

Through 5 rounds of multi-agent deliberation, we identified a **three-layer approach** that transforms doors from wireframe rectangles into solid surfaces with 3D depth: (1) background-filled interiors for visual mass, (2) bevel lighting on borders for raised-surface perception, and (3) a gradient shadow overhaul for spatial depth.

**The key insight:** A shadow on a wireframe doesn't read as "shadow" — it reads as "artifact." You need to give doors *visual mass* (background fill) before shadows become coherent. All three layers must work together.

---

## Round 1: Diagnosis — What Specifically Makes the Current Doors Look Bad?

### Root Cause Analysis (Three Independent Failure Modes)

**Failure Mode 1: No Visual Mass**
- Doors are wireframes — box-drawing character borders with transparent (terminal background) interiors
- The brain reads them as "line drawings" or "cards," not solid objects
- No theme uses background colors for door interiors despite Lipgloss fully supporting `Style.Background()`
- The `Fill` field in `ThemeColors` exists but is never applied to interior rendering

**Failure Mode 2: No Depth Hierarchy**
- Shadow system (`shadow.go`) uses a single `▐` character in `#585858`
- On dark terminals (bg `#000000` to `#1a1a2e`), contrast ratio is ~2.7:1 — below perceptual threshold
- Shadow is 1 character wide (~4-5px at typical font) — too thin for shadow gestalt
- Bottom shadow is 1 row of `▄` — insufficient for depth perception
- Brain needs ~4:1 contrast and ~2-3 chars width to read "shadow cast by object"
- Selected door shadow (`#bcbcbc`, ~8:1 contrast) is visible but reads as "gray line" not "shadow"

**Failure Mode 3: No Spatial Context**
- Doors float independently in terminal void
- The threshold line (Story 48.2) helps but only addresses the floor — no wall/corridor context
- No visual connection between the three doors suggests "shared space"

### Current Shadow Architecture Issues

- `ApplyShadow()` is a string post-processor — splits rendered output by `\n` and appends characters
- Can only add characters to the right and bottom — cannot affect door interior, left-side depth, or background colors
- Operates on string representation after theme rendering — cannot add background colors to already-rendered content without re-parsing ANSI escape sequences
- Architecturally limited: depth cues should be part of theme rendering, not a bolt-on

### Per-Theme "Visual Mass" Assessment

| Theme | Has Background Fill? | Has Depth Cues? | Visual Mass Rating |
|-------|:---:|:---:|:---:|
| Classic | No | Shadow only (imperceptible) | 1/5 |
| Modern | No | Shadow only (imperceptible) | 1/5 |
| Sci-Fi | Partial (░ shade rails give thickness) | Shadow + shade rails | 2.5/5 |
| Shoji | No | Shadow only | 1.5/5 (lattice adds density) |
| Winter | No | Shadow + frost dots | 1.5/5 |
| Spring | No | Shadow only | 1/5 |
| Summer | No | Shadow only | 1/5 |
| Autumn | No | Shadow only | 1/5 |

### SOUL.md Alignment Gap

> "The UI should feel like physical objects — doors that open, selections that click into place."

> "The difference between 'I clicked a flat screen' and 'I pressed a physical button' is the difference between adequate and delightful."

**Current state: firmly in "flat screen" territory.**

---

## Round 2: Exploration — What Techniques Exist?

### Reference TUI Applications

**Charm ecosystem:**
- **Glow** — Background-colored panels, rounded borders, subtle padding, color hierarchy. Cards have *substance*.
- **Soft Serve** — Panels with background fills, color-coded sections, clear visual hierarchy.
- **Huh** — Background-filled input fields create sense of interactive surfaces.

**Outside Charm:**
- **Lazygit** — Dense but clear. Background highlighting for selected items.
- **btop++** — Rich color gradients, dense information, background fills everywhere.

**Pattern:** The best TUI apps use **background colors** as their primary visual mass tool.

### Unicode Character Catalog for Depth

**Block Elements (U+2580–U+259F):**
- `█▓▒░` — full, dark, medium, light shade — gradient shadow potential
- `▀▄▌▐` — half blocks — subcharacter precision for bevels/edges

**Box Drawing (U+2500–U+257F):**
- Light: `┌─┐│└─┘` — standard weight
- Heavy: `┏━┓┃┗━┛` — heavier weight for emphasis
- Double: `╔═╗║╚═╝` — double-line for hinge asymmetry
- Mixed: `┑┙┡┩` — top-heavy/bottom-heavy corners for bevel effect

### Technique Catalog

| Technique | Method | Terminal Compat | Perf Impact | Width Cost |
|-----------|--------|:---:|:---:|:---:|
| **Background fill** | `lipgloss.Style.Background()` | Universal | Zero | 0 |
| **Multi-shade shadow** | 2-3 cols of ▓▒░ or bg-colored spaces | Universal | Negligible | +1-2 chars |
| **Edge highlight (bevel)** | Different fg color for top/left vs bottom/right borders | Universal | Zero | 0 |
| **Interior texture** | Dim fg shade chars over bg color | TrueColor recommended | Zero | 0 |
| **Color gradient rows** | Per-row background color stepping | TrueColor (ANSI256 fallback) | Zero | 0 |
| **Panel zone shading** | Different bg color above/below divider | Universal | Zero | 0 |

### The Bevel Technique (Key Discovery)

The classic "raised button" effect from GUI design works beautifully in TUI:
- Top and left edges: lighter color (light source is top-left, convention)
- Bottom and right edges: darker color (in shadow)
- Combined with background fill: instant 3D "raised surface" perception
- This is THE missing depth cue — shadow alone is only half the equation

---

## Round 3: Proposals

### Proposal A2: Background Fill Integration (Score: 25/30 — HIGHEST)

Add background color to all door interior rows. Transform wireframes into solid surfaces.

**Change:** Each theme's content/blank row rendering adds a background:
```go
bgStyle := lipgloss.NewStyle().Background(theme.Colors.Fill)
blankLine := style.Render(hingeV) + bgStyle.Render(strings.Repeat(" ", inner)) + style.Render(openV)
```

**Impact:** Massive. Doors become colored surfaces instead of transparent wireframes.
**Cost:** Zero width. ~10 LOC per theme.

### Proposal S1: Raised Surface — Bevel Lighting (Score: 24/30)

Classic raised-button effect. Top/left edges lighter, bottom/right darker.

**Unselected door:**
```
┌━━━━━━━━━━━━━━━━━━┑
┃                  │        Top/left border: bright (#888)
┃  [todo]          │        Right/bottom border: dim (#444)
┃  Buy groceries   │        Interior bg: #0d0d1a
┠                  ●        Content fg: #d0d0d0
┃                  │
╘━━━━━━━━━━━━━━━━━━┙▒░
 ░░░░░░░░░░░░░░░░░░░░
```

**Selected door (enhanced bevel):**
```
┏━━━━━━━━━━━━━━━━━┱
┃                  ╎░       All borders: bright white (#eee)
┃  [todo]          ╎░       Interior bg: #141428 (brighter)
┃  Buy groceries   ╎░       Content fg: #ffffff (bold)
┠                  ○░       Crack of light + handle turn
┃                  ╎░       Shadow: wider (3-col)
┗━━━━━━━━━━━━━━━━━┹▓▒░
 ░░░░░░░░░░░░░░░░░░░░░
```

### Proposal V1: Solid Surface + Gradient Shadow (Score: 24/30)

Background-filled doors with 2-char gradient shadow instead of 1-char flat shadow.

```
╓──────────────────┐░░
║                  │░░   Shadow: 2 cols (▓░ or bg-colored spaces)
║   Buy groceries  │░░   Shadow colors: #404040 → #262626
║                  │░░   Interior bg: #0d0d1a
╟                  ●░░
║                  │░░
╙──────────────────┘░░
 ░░░░░░░░░░░░░░░░░░░░░
```

### Proposal S2: Panel Zone Shading (Score: 22/30)

Differentiated upper/lower panels — upper panel slightly lighter, lower darker.

```
╓──────────────────┐
║▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒│    Upper panel: bg #12122a
║▒ [todo]          │
║▒ Buy groceries   │
╟──────────────────┤
║░░░░░░░░░░░░░░░░░│    Lower panel: bg #0a0a1e (darker)
║░░░░░░░░░░░░░░░░●│
╙──────────────────┘
```

### Proposal V2: Full Corridor (Score: 19/30 — Rejected)

Doors embedded in "wall" context with shade characters flanking.

```
▓▓╓──────────────┐▓▓▓╓──────────────┐▓▓▓╓──────────────┐▓▓
▓▓║  Buy milk    │▓▓▓║  Write code  │▓▓▓║  Fix faucet  │▓▓
▓▓╙──────────────┘▓▓▓╙──────────────┘▓▓▓╙──────────────┘▓▓
```

**Rejected:** Width cost of 6+ chars too expensive at min 15-char doors.

### Proposal W1: Theme-Level Depth System Refactor (Score: 23/30)

Refactor `ApplyShadow()` out of post-processing into theme `Render()` functions.

**ThemeColors extension:**
```go
type ThemeColors struct {
    Frame     lipgloss.TerminalColor  // existing
    Fill      lipgloss.TerminalColor  // REPURPOSE: door interior bg (upper panel)
    FillLower lipgloss.TerminalColor  // NEW: lower panel bg (darker)
    Accent    lipgloss.TerminalColor  // existing
    Selected  lipgloss.TerminalColor  // existing
    Highlight lipgloss.TerminalColor  // NEW: top/left edge (light)
    Shadow    lipgloss.TerminalColor  // NEW: bottom/right edge (dark)
    ShadowMid lipgloss.TerminalColor  // NEW: middle shadow gradient
    ShadowFar lipgloss.TerminalColor  // NEW: outer shadow gradient
}
```

---

## Round 4: Evaluation Matrix

| Proposal | Visual Impact (1-5) | Feasibility (1-5) | Terminal Compat (1-5) | Perf (1-5) | SOUL Alignment (1-5) | Width Cost | **TOTAL /25** |
|----------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **A2** Background Fill | 5 | 5 | 5 | 5 | 5 | 0 chars | **25** ⭐⭐ |
| **S1** Raised Surface (Bevel) | 5 | 4 | 5 | 5 | 5 | 0 chars | **24** ⭐ |
| **V1** Gradient Shadow | 5 | 4 | 5 | 5 | 5 | +1-2 chars | **24** ⭐ |
| **W1** Theme-Level Depth Refactor | 4 | 4 | 5 | 5 | 5 | 0 chars | **23** |
| **S2** Panel Zones | 3 | 5 | 5 | 5 | 4 | 0 chars | **22** |
| **A1** Quick Win Shadow Fix | 3 | 5 | 5 | 5 | 3 | +1 char | **21** |
| **V2** Full Corridor | 5 | 2 | 4 | 4 | 4 | +6 chars | **19** |
| **W2** Adaptive Depth | 2 | 3 | 5 | 5 | 3 | 0 chars | **18** |

### Key Evaluation Insights

1. **A2 scores highest** — single most impactful change at zero width cost and maximum feasibility
2. **A2 is the foundation** — without bg fill, shadow fixes just make a better shadow on a still-flat wireframe
3. **S1 + V1 tie as comprehensive solutions** — bevel + gradient shadow together create full 3D effect
4. **W1 is architectural prerequisite** for doing shadow gradient cleanly
5. **V2 is beautiful but impractical** at narrow widths

---

## Round 5: Converged Recommendations

### The Three-Layer Approach

All three layers must work together. Implementing only one produces partial improvement; all three together create the "physical objects" feel that SOUL.md demands.

#### Layer 1: Background Fill (Immediate, Standalone)

**What:** Add background color to all door interior rows across all 8 themes.

**How:** Modify each theme's `Render()` function to apply `lipgloss.NewStyle().Background(theme.Colors.Fill)` to interior space characters. The `Fill` field already exists in `ThemeColors` but is currently unused for rendering.

**Impact:** Doors transform from wireframes to solid surfaces. Massive visual upgrade.

**Cost:** Zero width. ~10 LOC per theme (~80 LOC total).

**Mockup (bg fill only, no other changes):**
```
╓──────────────────┐    ░ = bg-colored space (#0d0d2a for Classic)
║░░░░░░░░░░░░░░░░░░│    Content text rendered over bg color
║░ [todo]          │    Background creates "solid surface" feel
║░                 │
║░ Buy groceries   │
║░ 🛒 quick @home  │
╟──────────────────┤
║░░░░░░░░░░░░░░░░░░│
║░░░░░░░░░░░░░░░░░●│
║░░░░░░░░░░░░░░░░░░│
╙──────────────────┘
```

#### Layer 2: Bevel Lighting (Depth Perception)

**What:** Top/left border characters rendered in a lighter shade than bottom/right borders. Classic "raised button" effect.

**How:** Two border color variables per theme: `Highlight` (for top border + left/hinge border) and `ShadowEdge` (for bottom border + right border). Applied to border character rendering in each theme's row loop.

**Impact:** Doors gain 3D "raised surface" effect. Combined with bg fill, doors look like physical objects with an implicit light source (top-left).

**Cost:** Zero width. ~15 LOC per theme (~120 LOC total).

**Border color assignment:**
```
╓━━━━━━━━━━━━━━━━━━┑     ← top: Highlight color
║                  │     ← left (hinge): Highlight | right: ShadowEdge
║  Content         │
╟━━━━━━━━━━━━━━━━━━┤     ← divider: ShadowEdge color
║                  │
║                  ●
╘━━━━━━━━━━━━━━━━━━┙     ← bottom: ShadowEdge color
```

**Implementation detail — mixed corners for bevel:**
- Top-left corner: `┌` rendered in Highlight (both edges it touches are lit)
- Top-right corner: `┑` rendered in Highlight-to-Shadow transition
- Bottom-left corner: `╘` rendered in Highlight-to-Shadow transition
- Bottom-right corner: `┙` rendered in ShadowEdge (both edges in shadow)

#### Layer 3: Shadow Overhaul (Spatial Depth)

**What:** Replace 1-char imperceptible shadow with 2-3 char gradient shadow. Refactor shadow rendering from post-process into theme pipeline.

**How:** Each theme renders its own shadow columns in its row loop. Shadow uses gradient:
- Near column: `▓` or space with bg `#404040`
- Far column: `░` or space with bg `#262626`
- Bottom shadow: 2 rows, offset 1 char right

**Width-adaptive:**
- Width < minWidth + 2: no shadow (current behavior preserved)
- Width >= minWidth + 2: 1-col shadow (minimal)
- Width >= minWidth + 4: 2-col gradient shadow
- Width >= minWidth + 6: 3-col gradient shadow (wide terminals)

**Impact:** Proper drop-shadow effect. Doors visually "lift off" the terminal background.

**Cost:** +1-2 chars width (reduced from current +1 to adaptive). ~150 LOC total.

**Shadow color specifications:**
```
Door  ▓  ░
      ^  ^
      Near shadow: darker than door bg, lighter than terminal bg
      Far shadow: nearly terminal bg, creates fade-out
```

### Combined Final Mockup — All Three Layers

**Three doors, none selected (Classic theme, width ~30 each):**
```
╓━━━━━━━━━━━━━━━━━━┑     ╓━━━━━━━━━━━━━━━━━━┑     ╓━━━━━━━━━━━━━━━━━━┑
║                  │▓░   ║                  │▓░   ║                  │▓░
║  [todo]          │▓░   ║  [in-progress]   │▓░   ║  [todo]          │▓░
║                  │▓░   ║                  │▓░   ║                  │▓░
║  Buy groceries   │▓░   ║  Write report    │▓░   ║  Fix faucet      │▓░
║                  │▓░   ║                  │▓░   ║                  │▓░
║  🛒 quick @home  │▓░   ║  📋 medium @desk  │▓░   ║  🔧 quick @home  │▓░
╟━━━━━━━━━━━━━━━━━━┤▓░   ╟━━━━━━━━━━━━━━━━━━┤▓░   ╟━━━━━━━━━━━━━━━━━━┤▓░
║                  │▓░   ║                  │▓░   ║                  │▓░
║                  ●▓░   ║                  ●▓░   ║                  ●▓░
║                  │▓░   ║                  │▓░   ║                  │▓░
╘━━━━━━━━━━━━━━━━━━┙▓░   ╘━━━━━━━━━━━━━━━━━━┙▓░   ╘━━━━━━━━━━━━━━━━━━┙▓░
 ░░░░░░░░░░░░░░░░░░░░░    ░░░░░░░░░░░░░░░░░░░░░    ░░░░░░░░░░░░░░░░░░░░░
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

Key visual changes from current:
- Interior spaces have background color (#0d0d2a) — doors are SOLID
- Top border + left (hinge) border: lighter color (Highlight)
- Bottom border + right border: darker color (ShadowEdge)
- Shadow: 2-column gradient (▓░) with proper contrast
- Bottom shadow: 1 row, offset, gradient fade

**Middle door selected (with Epic 48 crack-of-light + handle turn):**
```
╓━━━━━━━━━━━━━━━━━━┑     ┏━━━━━━━━━━━━━━━━━━┱     ╓━━━━━━━━━━━━━━━━━━┑
║                  │▓░   ┃                  ╎░▓░  ║                  │▓░
║  [todo]          │▓░   ┃  [in-progress]   ╎░▓░  ║  [todo]          │▓░
║                  │▓░   ┃                  ╎░▓░  ║                  │▓░
║  Buy groceries   │▓░   ┃  Write report    ╎░▓░  ║  Fix faucet      │▓░
║                  │▓░   ┃                  ╎░▓░  ║                  │▓░
║  🛒 quick @home  │▓░   ┃  📋 medium @desk  ╎░▓░  ║  🔧 quick @home  │▓░
╟━━━━━━━━━━━━━━━━━━┤▓░   ┠━━━━━━━━━━━━━━━━━━┤░▓░  ╟━━━━━━━━━━━━━━━━━━┤▓░
║                  │▓░   ┃                  ╎░▓░  ║                  │▓░
║                  ●▓░   ┃                  ○░▓░  ║                  ●▓░
║                  │▓░   ┃                  ╎░▓░  ║                  │▓░
╘━━━━━━━━━━━━━━━━━━┙▓░   ┗━━━━━━━━━━━━━━━━━━┹▓▒░  ╘━━━━━━━━━━━━━━━━━━┙▓░
 ░░░░░░░░░░░░░░░░░░░░░    ░░░░░░░░░░░░░░░░░░░░░░   ░░░░░░░░░░░░░░░░░░░░░
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```

Selected door differences:
- Brighter interior bg (#141428 vs #0d0d2a)
- All borders in bright white (selected override)
- Crack-of-light (╎░) on right edge
- Handle turned (○ vs ●)
- Wider shadow (3-col ▓▒░)
- Unselected doors dimmed (Faint)

### Width-Adaptive Mockups

**Narrow terminal (width ~60, door width ~18):**
```
╓━━━━━━━━━━━━━━┑   ╓━━━━━━━━━━━━━━┑   ╓━━━━━━━━━━━━━━┑
║  [todo]      │▓  ║  [in-prog]   │▓  ║  [todo]      │▓
║  Buy milk    │▓  ║  Write code  │▓  ║  Fix faucet  │▓
╟━━━━━━━━━━━━━━┤▓  ╟━━━━━━━━━━━━━━┤▓  ╟━━━━━━━━━━━━━━┤▓
║            ● │▓  ║            ● │▓  ║            ● │▓
╘━━━━━━━━━━━━━━┙▓  ╘━━━━━━━━━━━━━━┙▓  ╘━━━━━━━━━━━━━━┙▓
 ░░░░░░░░░░░░░░░░   ░░░░░░░░░░░░░░░░   ░░░░░░░░░░░░░░░░
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```
Shadow: 1 col at narrow width. Bg fill + bevel still present.

**Very narrow terminal (width ~48, door width ~15 — minimum):**
```
╓━━━━━━━━━━━━┑ ╓━━━━━━━━━━━━┑ ╓━━━━━━━━━━━━┑
║  [todo]    │ ║  [in-prog] │ ║  [todo]    │
║  Buy milk  │ ║  Write rpt │ ║  Fix stuff │
╟━━━━━━━━━━━━┤ ╟━━━━━━━━━━━━┤ ╟━━━━━━━━━━━━┤
║          ● │ ║          ● │ ║          ● │
╘━━━━━━━━━━━━┙ ╘━━━━━━━━━━━━┙ ╘━━━━━━━━━━━━┙
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```
No shadow at minimum width. Bg fill + bevel still present.

**Wide terminal (width ~120, door width ~38):**
```
╓━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┑       ╓━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┑       ╓━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┑
║                                  │▓▒░    ║                                  │▓▒░    ║                                  │▓▒░
║  [todo]                          │▓▒░    ║  [in-progress]                   │▓▒░    ║  [todo]                          │▓▒░
║                                  │▓▒░    ║                                  │▓▒░    ║                                  │▓▒░
║  Buy groceries for the week      │▓▒░    ║  Write quarterly report          │▓▒░    ║  Fix the leaky kitchen faucet    │▓▒░
║                                  │▓▒░    ║                                  │▓▒░    ║                                  │▓▒░
╟━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┤▓▒░    ╟━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┤▓▒░    ╟━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┤▓▒░
║                                  │▓▒░    ║                                  │▓▒░    ║                                  │▓▒░
║                                  ●▓▒░    ║                                  ●▓▒░    ║                                  ●▓▒░
║                                  │▓▒░    ║                                  │▓▒░    ║                                  │▓▒░
╘━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┙▓▒░    ╘━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┙▓▒░    ╘━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┙▓▒░
 ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░
▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔
```
At wide widths: 3-column shadow gradient (▓▒░) for maximum depth.

---

## Per-Theme Color Palettes

### Classic Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#0d0d2a` | `17` | `0` |
| Fill (lower) | `#080820` | `17` | `0` |
| Highlight | `#7070ff` | `63` | `5` |
| ShadowEdge | `#3a3a8f` | `61` | `4` |
| Shadow Near | `#2a2a50` | `237` | `8` |
| Shadow Far | `#15152a` | `235` | `0` |

### Modern Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#0d0d0d` | `233` | `0` |
| Fill (lower) | `#080808` | `232` | `0` |
| Highlight | `#666666` | `241` | `7` |
| ShadowEdge | `#2a2a2a` | `235` | `8` |
| Shadow Near | `#222222` | `235` | `8` |
| Shadow Far | `#111111` | `233` | `0` |

### Sci-Fi Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#0a1a2e` | `17` | `0` |
| Fill (lower) | `#061425` | `17` | `0` |
| Highlight | `#00d7ff` | `45` | `14` |
| ShadowEdge | `#005f7f` | `24` | `4` |
| Shadow Near | `#003f5f` | `23` | `4` |
| Shadow Far | `#001a2f` | `17` | `0` |

### Shoji Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#1a1508` | `234` | `0` |
| Fill (lower) | `#141005` | `233` | `0` |
| Highlight | `#e8c888` | `186` | `11` |
| ShadowEdge | `#8f7540` | `137` | `3` |
| Shadow Near | `#6a5530` | `94` | `3` |
| Shadow Far | `#3a2a18` | `236` | `0` |

### Winter Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#0a0f1a` | `233` | `0` |
| Fill (lower) | `#060a14` | `232` | `0` |
| Highlight | `#a0d2e8` | `152` | `14` |
| ShadowEdge | `#4a6a80` | `66` | `4` |
| Shadow Near | `#354f60` | `59` | `8` |
| Shadow Far | `#1a2a38` | `236` | `0` |

### Spring Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#0a1a0d` | `233` | `0` |
| Fill (lower) | `#061408` | `232` | `0` |
| Highlight | `#80e090` | `114` | `10` |
| ShadowEdge | `#306838` | `65` | `2` |
| Shadow Near | `#204a28` | `22` | `2` |
| Shadow Far | `#102a14` | `233` | `0` |

### Summer Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#1a1508` | `234` | `0` |
| Fill (lower) | `#141005` | `233` | `0` |
| Highlight | `#ffd060` | `221` | `11` |
| ShadowEdge | `#8f7020` | `136` | `3` |
| Shadow Near | `#6a5518` | `94` | `3` |
| Shadow Far | `#3a2a08` | `236` | `0` |

### Autumn Theme
| Color Role | TrueColor | ANSI256 | ANSI |
|------------|:---------:|:-------:|:----:|
| Fill (upper) | `#1a0f08` | `233` | `0` |
| Fill (lower) | `#140a05` | `232` | `0` |
| Highlight | `#e09040` | `172` | `3` |
| ShadowEdge | `#8f5020` | `130` | `1` |
| Shadow Near | `#6a3818` | `94` | `1` |
| Shadow Far | `#3a1a08` | `236` | `0` |

---

## Implementation Sequence (Recommended Story Breakdown)

### Story A: ThemeColors Extension + Background Fill
- Extend `ThemeColors` with `FillLower`, `Highlight`, `ShadowEdge`, `ShadowNear`, `ShadowFar`
- Add background color to interior rows in all 8 themes
- Upper panel uses `Fill`, lower panel uses `FillLower`
- ~80 LOC, zero width cost
- **No dependencies** — can start immediately

### Story B: Bevel Lighting
- Top border + left/hinge border rendered in `Highlight` color
- Bottom border + right border rendered in `ShadowEdge` color
- Update all 8 themes' border rendering
- Mixed corner characters where highlight meets shadow
- ~120 LOC, zero width cost
- **No dependencies** — can parallelize with Story A

### Story C: Shadow Overhaul
- Refactor `ApplyShadow()` into per-theme shadow rendering within `Render()`
- Width-adaptive shadow: 0/1/2/3 columns based on available width
- Gradient colors from `ShadowNear` → `ShadowFar`
- Bottom shadow row with offset and fade
- ~150 LOC, +1-2 chars width (adaptive)
- **Depends on Story A** (needs new ThemeColors fields)

### Story D: Panel Zone Shading
- Upper panel (above PanelDivider) uses `Fill` background
- Lower panel (below PanelDivider) uses `FillLower` background (darker)
- Creates visual two-panel door effect
- ~60 LOC, zero width cost
- **Depends on Story A** (needs `FillLower` field)

### Story E: Width-Adaptive Shadow Tuning
- Tune shadow column count based on door width thresholds
- Fine-tune gradient colors per theme for optimal contrast
- Add golden file tests for shadow at various widths
- ~40 LOC
- **Depends on Story C**

### Dependency Graph
```
Story A (ThemeColors + Bg Fill) ──┬──→ Story C (Shadow Overhaul) ──→ Story E (Tuning)
                                  └──→ Story D (Panel Zones)
Story B (Bevel) ─────────────────────→ (independent)
```

Stories A & B can parallelize. Stories C & D can parallelize after A.

---

## Rejected Alternatives

| Rejected | Reason for Rejection |
|---|---|
| **V2 Full Corridor** | Width cost of 6+ chars too expensive at min 15-char doors. Beautiful but impractical for the default experience. Could revisit as a "wide terminal only" mode in a future epic. |
| **W2 Adaptive Depth (Terminal Detection)** | Over-engineering for the current state. `CompleteColor` already handles ANSI fallbacks. Background fill works universally on ANSI256+. Not needed now. |
| **Interior Texture (explicit shade chars as wood grain)** | Competes with content readability. Background color achieves "mass" without visual noise. Text contrast must be preserved — subtle bg color is better than visible texture characters. |
| **Braille Patterns (U+2800–U+28FF)** | Subpixel detail possible but compatibility/accessibility concerns. Not suitable for primary rendering. |

---

## Architectural Notes

### Files Affected
- `internal/tui/themes/theme.go` — Extend `ThemeColors` struct
- `internal/tui/themes/classic.go` — Bg fill, bevel, shadow
- `internal/tui/themes/modern.go` — Bg fill, bevel, shadow
- `internal/tui/themes/scifi.go` — Bg fill, bevel, shadow
- `internal/tui/themes/shoji.go` — Bg fill, bevel, shadow
- `internal/tui/themes/winter.go` — Bg fill, bevel, shadow
- `internal/tui/themes/spring.go` — Bg fill, bevel, shadow
- `internal/tui/themes/summer.go` — Bg fill, bevel, shadow
- `internal/tui/themes/autumn.go` — Bg fill, bevel, shadow
- `internal/tui/themes/shadow.go` — Refactor or deprecate
- `internal/tui/styles.go` — New color constants if needed
- `internal/tui/doors_view.go` — Shadow rendering integration

### Performance Impact
Zero. Background colors use the same ANSI escape sequence structure as foreground colors. Gradient shadow adds a few extra `strings.Repeat` calls per door per frame — negligible at 60 FPS.

### Terminal Compatibility
- Background colors: universal (ANSI 16+)
- TrueColor backgrounds: iTerm2, kitty, WezTerm, Alacritty, Windows Terminal, GNOME Terminal
- Falls back to ANSI256 via `CompleteColor` — already the project's standard approach
- Block elements (▓▒░): universal in Unicode-capable terminals

### Integration with Epic 48 (In-Progress)
- Story 48.1 (Handle + Hinge): **Done** — side-mounted handles and hinge asymmetry already implemented
- Story 48.2 (Threshold): **Done** — continuous floor line already in place
- Story 48.3 (Crack of Light): **Not Started** — will benefit from background fill (crack effect more visible against solid surface vs wireframe)
- Story 48.4 (Handle Turn): **Not Started** — independent of visual mass improvements

The visual redesign stories can be implemented as a new epic or added to Epic 48 as extension stories.

---

## Decision Record

**Decision:** Adopt three-layer approach (background fill + bevel lighting + shadow gradient) for door visual redesign.

**Rationale:** All three failure modes (no mass, no depth, imperceptible shadow) must be addressed together for coherent visual improvement. Background fill is the foundation — without it, shadow fixes just make a better shadow on a still-flat wireframe. Bevel lighting adds the missing "raised surface" depth cue. Gradient shadow provides spatial depth with proper contrast ratios.

**SOUL.md Alignment:**
- "The UI should feel like physical objects — doors that open, selections that click into place" → Background fill + bevel = solid, physical-feeling surfaces
- "The difference between 'I clicked a flat screen' and 'I pressed a physical button'" → Bevel lighting creates the "button" depth perception
- "Subtle is not the same as invisible" → Current shadow IS invisible (~2.7:1 contrast). New shadow targets 5:1+ contrast.

**Participants:** Victor (Innovation Strategist), Sally (UX Designer), Winston (Architect), Amelia (Dev), Maya (Design Thinking Coach), Dr. Quinn (Creative Problem Solver)
