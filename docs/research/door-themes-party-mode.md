# Door Theme System — Party Mode Discussion

**Date:** 2026-03-03
**Participants:** PM (Pat), Architect (Alex), UX Designer (Jordan), Dev (Sam), QA (Taylor)
**Topic:** Door theme system implementation strategy
**Source Research:** `docs/research/door-themes-research.md` (PR #116)

---

## Discussion Summary

### 1. Which Themes to Include in v1

**Consensus: Three themes for v1, chosen for maximum visual contrast and minimum rendering risk.**

| Theme | Champion | Rationale |
|-------|----------|-----------|
| **Modern / Minimalist** | UX, PM | Clean baseline, zero rendering risk, works as default |
| **Sci-Fi / Spaceship** | Dev, Architect | Strong identity, double-line box-drawing is universal |
| **Japanese Shoji** | UX, QA | Unique grid pattern, pure `┼─│` characters, scales naturally |

**Also discussed:**
- **Classic Wooden** was a close fourth — `▒` fill is reliable, but visually similar to "just a box with texture." Deferred to v1.1.
- **Vault / Safe** also v1.1 candidate. Only concern: `╳` crosshatch in handle.
- **Castle, Saloon, Garden Gate** deferred — require cross-terminal testing before shipping.

**PM (Pat):** "Three themes gives enough variety to feel like a choice without overwhelming. The 'Classic' no-theme mode (current Lipgloss borders) should remain as a fourth option for users who prefer simplicity."

**Architect (Alex):** "Agreed. The Classic mode is essentially the zero-value default — existing border rendering with no theme applied. Costs nothing to keep."

---

### 2. Theme Selection UX

#### First-Run Onboarding (Epic 10 Integration)

**Consensus: Add a theme picker step to the onboarding flow.**

- Show a horizontal preview of all three doors rendered with each theme
- User scrolls through themes with left/right arrows
- Enter to confirm selection
- "Skip" option defaults to Modern/Minimalist
- Preview renders at a fixed width (e.g., 28 chars per door) to show the theme's character

**UX (Jordan):** "The picker should feel like browsing, not configuring. Show the doors rendered with sample task text. Let the visual speak for itself — no lengthy descriptions needed."

**QA (Taylor):** "We need to handle narrow terminals gracefully in the picker. If terminal is too narrow to show three doors side-by-side, show them vertically or one at a time."

#### Settings View

**Consensus: Add a theme selection option to the settings/preferences area.**

- Accessible via `:theme` command or a settings menu item
- Same preview experience as onboarding
- Change takes effect immediately (no restart required)
- Current theme highlighted in the picker

**Dev (Sam):** "The DoorsView already re-renders on every Update cycle. Swapping the theme slice is instant — just reassign `dv.themes` and the next View() call picks it up."

#### Config.yaml Support

**Consensus: Theme preference persisted in config.yaml.**

```yaml
# ~/.threedoors/config.yaml
theme: modern  # Options: classic, modern, scifi, shoji
```

- Theme name is a simple string key, not a path or struct
- Invalid theme name falls back to "modern" with a warning logged
- Config file is the source of truth; settings view writes to config
- Each door in the trio uses the same theme (not mix-and-match for v1)

**Architect (Alex):** "Same theme for all three doors in v1. Mix-and-match is a future enhancement. It complicates the picker UX and the config schema."

**PM (Pat):** "Agreed. Single theme selection keeps the mental model simple. 'Pick your door style' is clearer than 'assign a style to each door.'"

---

### 3. Architecture Decisions

#### DoorTheme Struct

**Consensus: Adopt the research doc's proposed struct with minor refinements.**

```go
// DoorTheme defines the visual frame for a door.
type DoorTheme struct {
    Name        string
    Description string  // Short description for picker UI
    Render      func(content string, width, height int, selected bool) string
    Colors      ThemeColors
    MinWidth    int  // Minimum terminal width for this theme; below this, fall back to classic
}

type ThemeColors struct {
    Frame    lipgloss.Color
    Fill     lipgloss.Color
    Accent   lipgloss.Color
    Selected lipgloss.Color
}
```

**Architect (Alex):** "Adding `MinWidth` to the struct lets each theme declare its own minimum. The DoorsView checks terminal width against the active theme's MinWidth and falls back to classic rendering if too narrow."

**Dev (Sam):** "`Description` is needed for the picker UI — one line like 'Clean lines, minimal ornamentation' displayed below the preview."

#### Theme Registry

```go
// Registry is a simple map, not a plugin system.
var Registry = map[string]DoorTheme{
    "classic": ClassicTheme,
    "modern":  ModernTheme,
    "scifi":   SciFiTheme,
    "shoji":   ShojiTheme,
}

// DefaultTheme is used when no theme is configured or the configured theme is invalid.
const DefaultThemeName = "modern"
```

**Architect (Alex):** "A map keyed by name is simpler than a slice. Lookup by config string is O(1). Registration is explicit — no reflection, no init() functions."

**Dev (Sam):** "The registry lives in `internal/tui/themes/registry.go`. Each theme file registers itself by being included in the map literal. No runtime registration needed for v1."

#### File Organization

```
internal/tui/themes/
    theme.go       // DoorTheme, ThemeColors types
    registry.go    // Registry map, DefaultThemeName, lookup helper
    classic.go     // Classic (current Lipgloss borders, wrapped as a DoorTheme)
    modern.go      // Modern / Minimalist
    scifi.go       // Sci-Fi / Spaceship
    shoji.go       // Japanese Shoji
```

**Architect (Alex):** "The `classic.go` theme wraps the existing Lipgloss border rendering as a DoorTheme. This means the current rendering path is preserved as a theme option, not deleted."

#### Integration with DoorsView

**Dev (Sam):** "The change to `doors_view.go` is minimal:"

```go
// In DoorsView struct, add:
theme themes.DoorTheme

// In View(), replace:
// style := doorStyle.Width(doorWidth)
// renderedDoors = append(renderedDoors, style.Render(content))
// With:
renderedDoors = append(renderedDoors, dv.theme.Render(content, doorWidth, doorHeight, i == dv.selectedDoor))
```

"The theme is loaded from config at DoorsView initialization. The `:theme` command swaps it at runtime."

---

### 4. Testing Strategy

**Consensus: Golden file tests are the primary strategy, with width boundary tests as secondary.**

#### Golden File Tests

- Each theme gets a golden file test: render at 28-char width and 40-char width, compare output
- Golden files stored in `internal/tui/themes/testdata/`
- Use the project's existing golden test infrastructure (`golden_test.go` pattern)
- Test both selected and unselected states

**QA (Taylor):** "Golden files catch visual regressions instantly. If someone changes a box-drawing character, the diff is obvious. Run `go test ./internal/tui/themes/ -update` to regenerate."

#### Width Boundary Tests

- Test each theme at its declared `MinWidth` — should render correctly
- Test at `MinWidth - 1` — should indicate fallback needed
- Test at very wide terminal (120+ chars) — should scale gracefully

#### Content Wrapping Tests

- Test with short task text (1 line), medium (2-3 lines), and long (5+ lines that need wrapping)
- Verify word wrapping works correctly within each theme's content area

#### Cross-Terminal Testing (Manual)

- Pre-release: manual verification across iTerm2, Terminal.app, Alacritty
- Screenshot comparison (not automated for v1)
- Document which themes work in which terminals

**QA (Taylor):** "We should add a `vhs` tape (Charm's terminal recorder) for each theme as part of the docs. Serves as both documentation and visual regression reference."

**Dev (Sam):** "VHS tapes are a nice-to-have for v1.1. For v1, golden files + manual terminal checks are sufficient."

---

### 5. Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| v1 themes | Modern, Sci-Fi, Shoji + Classic (fallback) | Maximum contrast, minimum rendering risk |
| Theme application | Same theme for all 3 doors | Simpler UX and config; mix-and-match deferred |
| Persistence | config.yaml string key | Simple, human-readable, consistent with existing config pattern |
| Default theme | Modern / Minimalist | Cleanest rendering, lowest risk, welcoming aesthetic |
| Per-door colors | Theme colors override per-door colors | ThemeColors struct takes precedence; classic mode preserves existing colors |
| Door number labels | Overlaid separately, not part of theme frame | Functional UI stays consistent across themes |
| Theme preview | Settings view + onboarding picker | Same component, reused in both contexts |
| Seasonal variants | Deferred | Fun but scope creep |
| Mix-and-match doors | Deferred | Per-door theme assignment is a future enhancement |

---

### 6. Risks Identified

1. **Scope creep temptation** — Themes are inherently fun to design. Risk of spending too much time on visual polish vs. shipping. Mitigated by hard cap of 3 themes for v1.
2. **Width calculation bugs** — Manual frame drawing requires precise width math. Off-by-one errors will be visible. Mitigated by golden file tests at multiple widths.
3. **Lipgloss interaction** — Themes bypass `Border()` but still use `lipgloss.Width()` for measurement. ANSI escape codes in styled text can confuse width calculations. Mitigated by using `lipgloss.Width()` consistently.

---

### 7. Recommended Epic Structure

**Epic 17: Door Theme System**

| Story | Title | Effort |
|-------|-------|--------|
| 17.1 | Theme types, registry, and classic theme wrapper | S |
| 17.2 | Modern, Sci-Fi, and Shoji theme implementations | M |
| 17.3 | DoorsView integration — load theme from config, apply in View() | S |
| 17.4 | Theme picker in onboarding flow | M |
| 17.5 | Settings view — `:theme` command with preview | M |
| 17.6 | Golden file tests for all themes | S |

**Total estimated effort:** 2-3 weeks at 2-4 hrs/week

---

*Party mode discussion conducted as part of the BMAD methodology multi-agent workflow.*
