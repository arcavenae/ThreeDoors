# Party Mode Discussion: Door Theme Visual Redesign

**Date:** 2026-03-08
**Topic:** The door themes in ThreeDoors do not look like doors — how to achieve door-like appearance in terminal TUI
**Participants:** All BMM agents (PM, Architect, Analyst, UX Designer, QA, Dev, SM, Tech Writer) + All CIS agents (Storyteller, Design-Thinking-Coach, Brainstorming-Coach, Creative-Problem-Solver, Presentation-Master, Innovation-Strategist) + TEA (Test Architect)

---

## Round 1: Problem Framing

📋 **John (PM):** WHY don't the current themes look like doors? Let's be precise. A door has: a frame/jamb, a panel body, a handle/knob on one side, and optionally hinges on the other. Our current themes have frames (box-drawing characters) and the Modern theme has a knob (`●`), but they're all **portrait-oriented rectangles** — the same aspect ratio as a card or panel. Real doors are **tall and narrow** relative to their width. The fundamental issue is that we're rendering **landscape rectangles** (wider than tall) and calling them doors. The proportions are wrong.

🎨 **Sally (UX Designer):** Picture yourself standing in front of three doors. What do you see? Three tall, narrow rectangles side by side. You see the gap at the bottom. You see the frame around each one — thicker at the top (the header/lintel). You see one handle per door, positioned about 60% down on the latch side. Now look at our TUI: three squat, wide boxes. They look like cards in a Kanban board. The mental model is completely broken. Users don't *feel* like they're choosing a door — they feel like they're picking a sticky note.

📊 **Mary (Analyst):** This is a classic expectation gap. The product is literally called "Three Doors" — the metaphor is the brand. When the visual doesn't match the name, it creates cognitive dissonance. The Monty Hall problem reference reinforces that users expect *actual door shapes*. Market research shows: **metaphor fidelity drives engagement**. Users who feel the metaphor will interact more playfully with the product.

🧠 **Carson (Brainstorming Coach):** YES AND — this is a gift! The constraint of a terminal is actually beautiful because it forces us to be creative. Let me throw out some wild directions:
1. **Vertical doors** — What if each door is genuinely tall? Like 15+ rows high, narrow columns
2. **ASCII art door silhouette** — Literally draw a door outline with box characters, arch at top
3. **Doorknob placement as affordance** — The knob should be at a consistent vertical position that "reads" as a handle
4. **Negative space hallway** — Use background color to suggest a hallway/corridor with three doors in perspective
5. **Threshold/floor line** — A horizontal line at the bottom suggesting a floor/threshold
6. **Door number as room number** — Render the door number like a hotel room number plate above the door

---

## Round 2: Technical Feasibility & Design Approaches

🏗️ **Winston (Architect):** Let me ground this in what's actually achievable in a terminal. We have three constraints: (1) Character grid — no sub-character positioning, (2) Terminal width — typically 80-120 chars, split 3 ways gives ~25-35 chars per door, (3) Terminal height — typically 24-50 rows, but we need room for header/footer/status.

The most impactful change is **aspect ratio**. Currently our doors are about 8-10 rows tall and 25+ chars wide. A door-like proportion would be 15-20 rows tall and 18-22 chars wide. This is achievable — Bubbletea handles scrolling, and most modern terminals are at least 30 rows.

Key architectural elements for a "door-like" render:

```
    ┌─── [1] ───┐        ← Lintel / door number
    │           │
    │           │
    │  Content  │        ← Panel body (task text)
    │           │
    │        ●  │        ← Handle at ~60% height
    │           │
    │           │
    │           │
    ├───────────┤        ← Threshold
```

The **critical missing elements** are:
1. Taller aspect ratio (more empty rows)
2. A visible threshold/floor line at the bottom
3. Handle positioned at a realistic height (not bottom, not center)
4. Door number rendered as a "room number" at the top

🎨 **Caravaggio (Presentation Master):** OK here's what kills me about the current design — there's NO visual hierarchy that says "DOOR." Let me break down the visual language:

**What makes something read as a "door" in 2D:**
1. **Portrait rectangle** (taller than wide) — this is #1, non-negotiable
2. **Frame/trim** thicker than a simple border — double-line or bold outer, single-line inner
3. **Handle on ONE side** — asymmetry is key! Centered handles look like buttons, not doorknobs
4. **Panel divisions** — real doors have 2-6 raised panels. A horizontal divider or two creates this
5. **Threshold gap** — tiny gap or different character at the very bottom
6. **Header/lintel** — slightly wider top border or a number plate

**The 3-second test:** Can someone glance at the TUI for 3 seconds and think "those are doors"? Currently: no. With these changes: yes.

🔬 **Dr. Quinn (Creative Problem Solver):** AHA — the root cause is clear. We have a **semiotic failure**: the signifiers present (rectangle, border, content) map to "card/panel" not "door." We need to add the minimum signifiers that trigger "door" recognition. Using TRIZ's Principle of Segmentation:

**Minimum viable door signifiers (in priority order):**
1. Portrait aspect ratio (taller than wide)
2. Asymmetric handle placement (right side, 60% down)
3. Panel divisions (at least one horizontal divider creating upper/lower panels)
4. Threshold at bottom (different bottom border treatment)

Just these 4 changes would make any theme "read" as a door. Each theme can then add its own flavor (rivets for sci-fi, lattice for shoji, etc.) on top of this base "door grammar."

---

## Round 3: Theme-Specific Door Designs

🎨 **Sally (UX Designer):** Let me sketch what each theme could look like with proper door grammar:

### Classic Door
```
╭──── [1] ────╮
│             │
│   Task      │
│   text      │
│   here      │
│             │
├─────────────┤    ← Panel divider
│             │
│          ●  │    ← Doorknob
│             │
│             │
│             │
╰─────────────╯
  ▔▔▔▔▔▔▔▔▔▔▔     ← Threshold/floor shadow
```

### Modern Door
```
━━━━ 1 ━━━━━━━
┃             ┃
┃   Task      ┃
┃   text      ┃
┃             ┃
┃─────────────┃    ← Minimalist panel line
┃             ┃
┃          ○  ┃    ← Minimalist handle
┃             ┃
┃             ┃
┃             ┃
━━━━━━━━━━━━━━━
```

### Sci-Fi Door
```
╔═╤═══ 1 ═══╤═╗
║░│         │░║
║░│  Task   │░║
║░│  text   │░║
║░│         │░║
║░╞═════════╡░║    ← Bulkhead divider
║░│         │░║
║░│    ◈──┤ │░║    ← Access panel handle
║░│         │░║
║░│         │░║
║░│[ACCESS] │░║
╚═╧═════════╧═╝
  ▓▓▓▓▓▓▓▓▓▓▓     ← Floor grating
```

### Shoji Door
```
┬────┬─ 1 ─┬────┬
│    │      │    │
│    │ Task │    │
│    │ text │    │
├────┼──────┼────┤    ← Lattice cross-bar
│    │      │    │
│    │      │    │
│    │   ○  │    │    ← Sliding handle (recessed)
│    │      │    │
├────┼──────┼────┤    ← Lower lattice
│    │      │    │
┴────┴──────┴────┴
```

📖 **Sophia (Storyteller):** Each door tells a different story! The Classic is a warm, familiar home door — *come in, friend*. The Modern is a sleek office door — *efficiency awaits*. The Sci-Fi is an airlock — *adventure beyond*. The Shoji slides open with a whisper — *harmony within*. When users see these shapes, they don't just pick a task — they *open a door*. That's the magic of metaphor fidelity.

---

## Round 4: Implementation Strategy

💻 **Amelia (Dev):** Implementation approach:

1. **Base door template system** — Each theme's `Render()` function gets a `doorHeight` parameter in addition to width. The height should be calculated as `max(minDoorHeight, contentLines + paddingLines)` where `minDoorHeight` varies by theme (12-16 rows).

2. **Door anatomy constants** — Define structural positions relative to height:
   - `lintelRow = 0` (top frame with door number)
   - `contentStartRow = 2` (after lintel + padding)
   - `panelDividerRow = height * 0.45` (upper/lower panel divide)
   - `handleRow = height * 0.60` (knob/handle position)
   - `thresholdRow = height - 1` (bottom frame)

3. **Handle rendering** — Each theme defines its handle character and side:
   - Classic: `●` right side
   - Modern: `○` right side
   - Sci-Fi: `◈──┤` right side (access panel)
   - Shoji: `○` center-right (sliding handle)

4. **Aspect ratio enforcement** — `DoorsView` should calculate door height first (available terminal height minus chrome), then width. Minimum 12 rows for door appearance.

5. **Backward compatibility** — If terminal height < 16 rows, fall back to current compact rendering. Set a `compactMode` threshold.

🏃 **Bob (SM):** Clear story breakdown:
1. **Story A:** Base door anatomy model — define door height calculation, structural row positions, handle placement API
2. **Story B:** Update Classic theme — portrait orientation, panel divider, threshold
3. **Story C:** Update Modern theme — same anatomy with minimalist treatment
4. **Story D:** Update Sci-Fi theme — bulkhead divider, access panel handle, floor grating
5. **Story E:** Update Shoji theme — lattice with proper door proportions, sliding handle
6. **Story F:** Compact mode fallback — detect low-height terminals, graceful degradation
7. **Story G:** Update golden file tests — new expected renders for all themes at multiple sizes

🏗️ **Winston (Architect):** Architectural changes needed:
1. **Theme interface extension** — `Render()` signature needs height parameter: `Render(content string, width int, height int, selected bool) string`
2. **Door anatomy struct** — New `DoorAnatomy` type that calculates row positions from height
3. **DoorsView height management** — Currently only manages width distribution. Need vertical layout calculation.
4. **No new dependencies** — All changes use existing box-drawing and Lipgloss capabilities

---

## Round 5: Accessibility & Edge Cases

🧪 **Quinn (QA):** Testing concerns:
1. **Golden file updates** — Every theme test needs new baselines at door-like proportions
2. **Minimum terminal size** — What's the smallest terminal that still gets door-shaped rendering? Need tests at 80x24 (classic), 120x30 (comfortable), 60x20 (cramped)
3. **Content overflow** — What happens when task text is longer than the door's content area? Need word-wrap + truncation with ellipsis
4. **Selection state** — Selected door should still be clearly distinguishable with new proportions
5. **Screen reader testing** — Door anatomy shouldn't break screen reader content extraction

🧪 **Murat (TEA):** Accessibility matrix:

| Concern | Current | After Redesign | Mitigation |
|---------|---------|----------------|------------|
| Screen readers | Content in box | Content in taller box | No change — content extraction unaffected |
| Low vision | Color-differentiated | Same + structural | Better — shape recognition aids identification |
| Monochrome | All same shape | All same shape but taller | Better — panel dividers add structure without color |
| High contrast | Lipgloss colors | Same | No change |
| Narrow terminal | Falls back to Classic | Falls back to compact mode | Same approach, new threshold |
| Short terminal | Not addressed | Compact mode fallback | New — important addition |

🎨 **Maya (Design Thinking Coach):** Let's think about this from the user's lived experience. When they resize their terminal window, the doors should *gracefully transform* — from full door-like proportions down to compact cards. The user should feel like they're "zooming out" from the doors, not that the UI is breaking. The transition should feel intentional, not degraded.

⚡ **Victor (Innovation Strategist):** The bigger strategic move here: **door proportions enable future features**. Once doors are tall and narrow:
- **Opening animations** become possible — a door can "swing open" with progressive frame changes
- **Peek/preview** — partially opened door showing task detail
- **Door states** — closed (todo), ajar (active), open (selected), locked (blocked)
- **Depth/shadow** — half-block characters (`▀▄`) can create depth on one side
- These are all P2/P3 but the portrait proportion is the prerequisite for ALL of them

---

## Round 6: Consensus Recommendations

📋 **John (PM):** Let me synthesize the actionable recommendations:

### Consensus: Must-Have Changes (P0)
1. **Portrait aspect ratio** — Doors must be taller than wide. Minimum 12 rows height. This is the #1 fix.
2. **Panel divider** — At least one horizontal line creating upper/lower panels. This is the strongest "door" signifier after proportion.
3. **Asymmetric handle** — Knob/handle on the right side at ~60% height. Already partially done in Modern theme; apply to all themes.
4. **Threshold/floor line** — Different treatment at the bottom edge to suggest a floor/ground plane.
5. **Door number as room number** — Render door number in the lintel/header area, not as content.

### Consensus: Should-Have (P1)
6. **Compact mode fallback** — Graceful degradation for terminals under 16 rows high.
7. **Height-aware rendering** — `Render()` function takes height parameter; `DoorsView` calculates available height.
8. **Shadow/depth on one side** — Half-block characters or shade characters creating a 3D effect on the right/bottom edge of each door.

### Consensus: Nice-to-Have (P2 — defer to future epic)
9. **Door state animations** — Opening/closing visual transitions
10. **Perspective/vanishing point** — Hallway corridor effect with depth
11. **Door material textures** — Wood grain (`▒`), metal (`▓`), glass (` `)

### Accessibility Requirements
- All door signifiers must work in monochrome mode (structural, not color-dependent)
- Screen reader content extraction must remain unaffected
- Compact mode must preserve full task readability
- Minimum 4:3 content-to-chrome ratio (task text area vs decorative elements)

### Architecture Changes
- Extend `DoorTheme.Render()` signature to include height: `Render(content string, width int, height int, selected bool) string`
- Add `DoorAnatomy` helper to calculate structural row positions from height
- Add `MinHeight` field to `DoorTheme` struct alongside existing `MinWidth`
- Update `DoorsView` to calculate and distribute vertical space

---

## Summary

The fundamental fix is **proportion** — making doors portrait-oriented (taller than wide). Combined with panel dividers, asymmetric handles, and threshold lines, this transforms the current "card" appearance into recognizable doors while maintaining each theme's visual identity. The changes are achievable with existing box-drawing characters and Lipgloss — no new dependencies needed.

The door "grammar" (proportion + panels + handle + threshold) should be standardized across all themes, with each theme expressing these elements in its own visual language (rounded borders for Classic, heavy lines for Modern, double-line bulkheads for Sci-Fi, lattice grids for Shoji).

*Party mode discussion concluded. All agents in consensus on recommendations.*

## Decisions Summary

| Decision | Status | Rationale | Alternatives Rejected |
|----------|--------|-----------|----------------------|
| Portrait aspect ratio (taller than wide, min 12 rows) | Adopted | #1 fix for door recognition; real doors are tall and narrow | Landscape orientation (current — reads as cards, not doors) |
| Panel divider (horizontal line creating upper/lower panels) | Adopted | Strongest "door" signifier after proportion | Flat single-panel rendering (looks like a card) |
| Asymmetric handle placement (right side, ~60% height) | Adopted | Asymmetry signals "door" not "button"; matches real-world knob position | Centered handle (reads as button), no handle (loses affordance) |
| Threshold/floor line at bottom edge | Adopted | Suggests ground plane; completes door silhouette | Same border top and bottom (no spatial grounding) |
| Door number as room number in lintel | Adopted | Reinforces door metaphor; creates visual hierarchy | Number as content (loses metaphor), no number (loses identification) |
| Compact mode fallback for terminals < 16 rows | Adopted | Graceful degradation; don't break small terminals | Force minimum terminal size (user-hostile), no fallback (broken rendering) |
| Extend Render() with height parameter | Adopted | Enables proportional door rendering; DoorsView calculates available height | Fixed height (can't adapt to terminal), width-only (current — no vertical control) |
| Shadow/depth effect on one side | Adopted | Half-block characters create 3D effect cheaply | Flat rendering only (less immersive), full 3D (impossible in terminal) |
| Defer door animations to future epic | Rejected | P2 scope; portrait proportions are the prerequisite | — |
| Defer perspective/vanishing point to future epic | Rejected | P2 scope; too complex for initial door redesign | — |
| Defer material textures to future epic | Rejected | P2 scope; structural changes are higher priority | — |
