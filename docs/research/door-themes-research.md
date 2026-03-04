# Door Theme System Research

**Date:** 2026-03-03
**Status:** Research / Exploration
**Author:** Engineering

---

## 1. Current State

### How Doors Are Rendered Today

The three doors are rendered in `internal/tui/doors_view.go` using Lipgloss bordered boxes. Each door is a rectangle created by applying a `lipgloss.Style` with a border, padding, and a computed width.

**Key styles from `internal/tui/styles.go`:**

```go
doorStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(colorAccent).
    Padding(1, 2)

selectedDoorStyle = lipgloss.NewStyle().
    Border(lipgloss.ThickBorder()).
    BorderForeground(colorDoorBright).
    Padding(1, 2)
```

**Per-door colors** are applied when terminal width >= 60 characters:
- Door 0 (left): cyan (`86`)
- Door 1 (center): magenta (`212`)
- Door 2 (right): yellow (`220`)

**Content inside each door** is assembled as:
1. Status indicator (e.g., `[todo]`) in the status color
2. Task text
3. Source provider badge (if applicable)
4. Category badges (type icon, effort, location)
5. Avoidance indicator (if bypassed 5+ times)

The three rendered door strings are joined horizontally with `lipgloss.JoinHorizontal(lipgloss.Top, ...)`.

**Current visual result** is three rounded-corner boxes side by side, differentiated only by border color. They are functional but visually uniform -- there is no "door" identity or personality.

---

## 2. Theme Concept

### What a Theme System Could Look Like

The idea is to make each door visually distinct -- not just by color, but by shape, ornamentation, and character. Instead of three identical rounded boxes, each door would have its own ASCII/ANSI art frame that evokes a different style of door.

### Architectural Approach

A theme in this context is a function (or struct) that knows how to render a door frame around arbitrary content. The simplest viable design:

```go
// DoorTheme defines the visual frame for a door.
type DoorTheme struct {
    Name   string
    Render func(content string, width int, selected bool) string
}
```

A theme registry would hold available themes:

```go
var DefaultThemes = []DoorTheme{
    ClassicWoodenDoor,
    CastleDoor,
    SciFiDoor,
    // ...
}
```

At session start, three themes are randomly selected (or the user picks a theme set). Each door in the trio gets a different theme, making them visually distinct.

This keeps things simple -- no interface hierarchies, no abstract factories. Just a slice of render functions.

### Integration Point

In `DoorsView.View()`, the current code:

```go
style := doorStyle.Width(doorWidth)
renderedDoors = append(renderedDoors, style.Render(content))
```

Would become:

```go
theme := dv.themes[i]
renderedDoors = append(renderedDoors, theme.Render(content, doorWidth, i == dv.selectedDoorIndex))
```

---

## 3. ANSI Mockups

All mockups are designed for a roughly 28-character-wide, 12-15-character-tall frame. Task text is shown inside or beneath the door.

### 3.1 Classic Wooden Door

A simple rectangular door with a panel and round doorknob.

```
    ┌──────────────────────┐
    │▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒│
    │▒┌──────────────────┐▒│
    │▒│                  │▒│
    │▒│  Fix login bug   │▒│
    │▒│                  │▒│
    │▒└──────────────────┘▒│
    │▒                    ▒│
    │▒        ┌──┐        ▒│
    │▒        │◉ │        ▒│
    │▒        └──┘        ▒│
    │▒                    ▒│
    │▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒│
    └──────────────────────┘
```

**Lipgloss approach:** Custom border using box-drawing characters. The `▒` fill simulates wood grain. Content is placed in the upper panel area. The doorknob (`◉`) is decorative.

### 3.2 Castle / Medieval Door

An arched top with iron studs and heavy stone frame.

```
          ╭──────────╮
        ╭─╯            ╰─╮
      ╭─╯                ╰─╮
    ┌─╯                    ╰─┐
    │ ◆                    ◆ │
    │                        │
    │   Update API docs      │
    │                        │
    │ ◆        ⚒         ◆ │
    │                        │
    │ ◆                    ◆ │
    │                        │
    │ ◆                    ◆ │
    ╞════════════════════════╡
    └────────────────────────┘
```

**Notes:** The arch is drawn with curved box-drawing characters (`╭╮╰╯`). Iron studs are `◆` diamonds placed symmetrically. The heavy bottom rail uses double-line characters (`═`). The forged iron hinge symbol `⚒` adds character.

### 3.3 Modern / Minimalist Door

Clean lines, generous whitespace, understated elegance.

```
    ─────────────────────────
    │                       │
    │                       │
    │                       │
    │   Write unit tests    │
    │                       │
    │                       │
    │                   ●   │
    │                       │
    │                       │
    │                       │
    ─────────────────────────
```

**Notes:** No rounded corners, no ornamentation. Single thin lines. The doorknob is a minimal filled circle (`●`) placed asymmetrically to the right side. This theme relies on negative space. Border color alone provides personality.

### 3.4 Sci-Fi / Spaceship Door

Mechanical panels, rivets, and a sliding-door aesthetic.

```
    ╔══╤════════════════╤══╗
    ║░░│                │░░║
    ║░░│                │░░║
    ║░░│  Deploy to     │░░║
    ║░░│  staging       │░░║
    ║░░│                │░░║
    ╠══╪════════════════╪══╣
    ║▓▓│ ◈  ◈      ◈  ◈│▓▓║
    ║░░│                │░░║
    ║░░│   [ACCESS]     │░░║
    ║░░│                │░░║
    ╚══╧════════════════╧══╝
```

**Notes:** Double-line outer frame (`╔╗╚╝═║`) for a heavy industrial feel. Side rails use shade characters (`░▓`). Interior divided into upper content panel and lower control panel with rivets (`◈`). The `[ACCESS]` label is a decorative touch suggesting an airlock or console. Mid-bar uses `╠╣╪` cross characters.

### 3.5 Saloon / Western Doors

Swinging half-doors with slatted panels, showing the task "over the top."

```
           Review PR #42

    ╱│╲                ╱│╲
    ┌┤├────────────────┤├┐
    │╞╡                ╞╡│
    │├┤  ═══════════   ├┤│
    │╞╡  ═══════════   ╞╡│
    │├┤  ═══════════   ├┤│
    │╞╡  ═══════════   ╞╡│
    │├┤                ├┤│
    │╞╡     ◎         ╞╡│
    └┤├────────────────┤├┘
```

**Notes:** The swinging-door hinge brackets (`╱│╲`) sit at the top. Horizontal slats are drawn with `═` lines. The push-plate is `◎`. Task text is displayed above the doors (since saloon doors are short). Side hinges use alternating `├┤` and `╞╡` for a barrel-hinge look.

### 3.6 Vault / Safe Door

A heavy circular handle on a thick steel door.

```
    ╔════════════════════════╗
    ║ ┌────────────────────┐ ║
    ║ │                    │ ║
    ║ │  Backup database   │ ║
    ║ │                    │ ║
    ║ │      ╭────╮        │ ║
    ║ │      │ ╳╳ │        │ ║
    ║ │      ╰────╯        │ ║
    ║ │                    │ ║
    ║ └────────────────────┘ ║
    ║ ▪ ▪ ▪ ▪ ▪ ▪ ▪ ▪ ▪ ▪  ║
    ╚════════════════════════╝
```

**Notes:** Double-line outer frame for the thick vault wall. Inner recessed panel uses single lines. The circular handle is drawn with curved box characters and `╳` crosshatch for the locking mechanism. Bottom row of rivets (`▪`) reinforces the heavy steel look. The double-border gap between inner and outer frames creates a sense of depth/thickness.

### 3.7 Japanese Shoji Door

A sliding screen with a lattice grid pattern.

```
    ┬──┬──┬──┬──┬──┬──┬──┬
    │  │  │  │  │  │  │  │
    ├──┼──┼──┼──┼──┼──┼──┤
    │  │  │  │  │  │  │  │
    ├──┼──┼──┼──┼──┼──┼──┤
    │  │ Clean up   │  │  │
    │  │ backlog    │  │  │
    ├──┼──┼──┼──┼──┼──┼──┤
    │  │  │  │  │  │  │  │
    ├──┼──┼──┼──┼──┼──┼──┤
    │  │  │  │  │  │  │  │
    ┴──┴──┴──┴──┴──┴──┴──┴
```

**Notes:** The grid pattern uses `┼` cross junctions and `─│` lines to simulate the wooden lattice of a shoji screen. Task text overlays the central cells, as if written on the paper pane. Top and bottom use `┬` and `┴` for the frame rail. No doorknob -- shoji doors slide. The regularity of the grid is the visual signature.

### 3.8 Garden Gate

A wrought-iron garden gate with decorative scrollwork.

```
         ◠◡◠◡◠◡◠◡◠◡◠
    ┌────╫──────────╫────┐
    │    ║          ║    │
    │  ╔═╩══════════╩═╗  │
    │  ║              ║  │
    │  ║ Prune old    ║  │
    │  ║ branches     ║  │
    │  ║              ║  │
    │  ╚══════════════╝  │
    │    │  │    │  │    │
    │    │  │    │  │    │
    │    │  │    │  │    │
    └────┴──┴────┴──┴────┘
```

**Notes:** Decorative scrollwork at the top using `◠◡` wave characters. Double-line inner frame for the ornate ironwork feel. Vertical bars at the bottom simulate gate pickets. The content sits in the upper panel section, as if on a sign hanging on the gate.

---

## 4. Implementation Notes

### 4.1 Theme as a Render Function

Each theme is a function that takes content and dimensions and returns a fully rendered string. No Lipgloss `Border()` is used -- the theme draws its own frame character by character.

```go
type DoorTheme struct {
    Name     string
    Render   func(content string, width, height int, selected bool) string
    // Colors holds the palette for this theme
    Colors   ThemeColors
}

type ThemeColors struct {
    Frame    lipgloss.Color
    Fill     lipgloss.Color
    Accent   lipgloss.Color
    Selected lipgloss.Color
}
```

### 4.2 Interaction with Lipgloss

Lipgloss is still used for:
- **Color application:** `lipgloss.NewStyle().Foreground(color).Render(char)` for coloring individual frame characters
- **Text styling:** Bold, italic, faint for content inside the door
- **Horizontal joining:** `lipgloss.JoinHorizontal()` to lay out the three doors side by side
- **Width measurement:** `lipgloss.Width()` to measure rendered string widths accounting for ANSI escape codes

Lipgloss is NOT used for:
- Border drawing (themes handle this manually)
- Padding (themes control internal spacing)

### 4.3 Content Wrapping

Task text needs to be word-wrapped to fit inside the theme's content area. The content area width varies by theme (some have thicker frames than others). Use `lipgloss.NewStyle().Width(innerWidth).Render(text)` for word wrapping, or a manual word-wrap function.

### 4.4 Selection Mode

When a door is selected (highlighted), themes should have a "selected" variant. Options:
- Brighten or change the frame color
- Add a glow effect (extra layer of characters around the frame)
- Bold the frame characters
- Add a pointer/cursor indicator (`>>` or `▶`) beside the door

The simplest approach: each theme's `Render` function receives a `selected bool` and adjusts colors accordingly.

### 4.5 Theme Assignment Strategy

Three options, from simplest to most complex:

1. **Random per session:** At session start, randomly pick 3 themes from the pool. Each door gets a different theme. Themes persist for the session.

2. **User-selectable theme set:** A config option like `theme: medieval` that picks a cohesive set of three door variants. This requires designing variant sets rather than mix-and-match.

3. **Category-driven:** Map task types to door themes (creative tasks get the garden gate, technical tasks get the sci-fi door, etc.). Visually reinforces task categories.

**Recommendation:** Start with option 1 (random per session). It is the simplest, requires no config, and provides visual variety. Option 3 is the most interesting but couples visual design to domain logic.

### 4.6 Width Adaptation

Themes need to handle varying terminal widths. Each theme should define:
- A minimum width (below which it falls back to a simple bordered box)
- A target width (where it looks best)
- How to scale: some themes (shoji grid) scale by adding/removing grid cells; others (vault door) scale by adjusting internal padding

### 4.7 File Organization

```
internal/tui/
    themes/
        theme.go          // DoorTheme type, ThemeColors, registry
        classic.go        // Classic wooden door
        castle.go         // Medieval door
        modern.go         // Minimalist door
        scifi.go          // Spaceship door
        saloon.go         // Western swinging doors
        vault.go          // Safe door
        shoji.go          // Japanese screen
        garden.go         // Garden gate
```

Each file is small (one render function, one color set). The `theme.go` file defines the type and the registry slice.

---

## 5. Feasibility Assessment

### What Works Well in Terminals

**Box-drawing characters** (`─│┌┐└┘├┤┬┴┼` and double-line variants `═║╔╗╚╝`) render consistently across modern terminal emulators (iTerm2, Alacritty, kitty, Windows Terminal, GNOME Terminal). These are the backbone of all the mockups above and are highly reliable.

**Block/shade elements** (`░▒▓█`) work well for fill patterns. They render at consistent widths and are good for simulating textures (wood grain, metal panels).

**Geometric shapes** (`●○◉◎◆◇▪▫▸▶`) are well-supported and useful for doorknobs, rivets, studs, and decorative elements.

**Curved box-drawing** (`╭╮╰╯`) is supported in most modern terminals. These enable arched tops (castle door) and rounded handles (vault door).

**Color via ANSI 256-color palette** is universally supported. Lipgloss handles this cleanly. Coloring frame characters adds significant visual impact with minimal code complexity.

### What Is Risky

**Unicode decorative characters** (`◠◡` waves, `⚒` tools, `╳` crosses) have inconsistent widths across fonts and terminals. Some monospace fonts render these as double-width, which breaks alignment. These should be used sparingly and tested.

**Emoji** in frame elements should be avoided entirely. Emoji widths are notoriously inconsistent (some terminals render as 1 cell, others as 2). The existing emoji in task type icons (`typeIcon()`) already carries this risk for content, but it should not spread to frame rendering.

**Complex multi-line art** that depends on exact character alignment is fragile. If any single character in the frame renders at an unexpected width, the entire door collapses visually. Themes should be tested across at least: iTerm2, Terminal.app, Alacritty, and one Linux terminal.

**True color (24-bit)** works in most modern terminals but not all. Sticking to the 256-color palette (which Lipgloss `Color("123")` uses) is safer.

### Theme-by-Theme Feasibility

| Theme | Feasibility | Risk | Notes |
|-------|------------|------|-------|
| Classic Wooden | High | Low | Uses only basic box-drawing + `▒` fill |
| Castle / Medieval | Medium | Medium | Arch requires `╭╮╰╯` curves; `◆` studs may vary |
| Modern / Minimalist | High | Low | Simplest theme; just lines and `●` |
| Sci-Fi / Spaceship | High | Low | Double-line box-drawing is universal; `◈` is the only risk |
| Saloon / Western | Medium | Medium | Complex hinge brackets; horizontal slats need precise alignment |
| Vault / Safe | High | Low | Standard box-drawing; `╳` inside handle is minor risk |
| Japanese Shoji | High | Low | Pure grid of `┼─│` characters; very reliable |
| Garden Gate | Medium | Medium | `◠◡` wave characters are highest risk; pickets are fine |

### Recommended Starting Set

For a first implementation, start with the three highest-feasibility themes that also provide the most visual contrast:

1. **Modern / Minimalist** -- clean, reliable baseline
2. **Sci-Fi / Spaceship** -- heavy double-line frame, strong visual identity
3. **Japanese Shoji** -- grid pattern is unique and uses only basic box-drawing

This trio covers three very different visual aesthetics while using only well-supported Unicode characters.

### Performance Considerations

Rendering custom frames is string manipulation -- no I/O, no computation. Even the most complex theme adds microseconds to render time. This is not a concern for a TUI that redraws on keypress events.

### Testing Strategy

- **Golden file tests:** Render each theme at known widths and compare output against golden files (the project already has `golden_test.go` infrastructure).
- **Width boundary tests:** Verify themes degrade gracefully at minimum widths.
- **Terminal screenshot tests:** Manual verification across terminal emulators for the initial release. Automated screenshot testing (e.g., with `vhs` from Charm) could follow.

---

## 6. Open Questions

1. Should themes persist across sessions (saved in user config) or re-randomize each time?
2. Should the door number label (Door 1, 2, 3) be part of the theme frame or overlaid separately?
3. Do we want a "theme preview" command that shows all available themes?
4. Should themes have seasonal variants (spooky doors in October, festive doors in December)?
5. How do themes interact with the existing per-door color system? Replace it, or layer on top?
