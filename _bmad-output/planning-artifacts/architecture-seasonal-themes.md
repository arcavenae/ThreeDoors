---
stepsCompleted: ["step-01-init", "step-02-context", "step-03-decisions", "step-04-components"]
inputDocuments:
  - docs/prd/requirements.md (FR132-FR137, NFR28-NFR30)
  - docs/architecture/components.md (existing theme system architecture)
  - door-themes-analyst-review.md
  - internal/tui/themes/theme.go (DoorTheme struct)
  - internal/tui/themes/registry.go (Registry)
  - internal/tui/themes/modern.go (reference implementation)
workflowType: 'architecture'
project_name: 'ThreeDoors'
user_name: 'arcaven'
date: '2026-03-08'
---

# Architecture Decision Document — Seasonal Door Theme Variants (Epic 33)

This document defines the technical architecture for extending the existing Door Theme System (Epic 17) with seasonal theme variants that auto-switch based on the current date.

---

## 1. Project Context

ThreeDoors has a complete door theme system implemented in `internal/tui/themes/`:

- **`DoorTheme` struct** — Name, Description, Render function, ThemeColors, MinWidth
- **`Registry`** — Map-based theme storage with `Register()`, `Get()`, `Names()`
- **`NewDefaultRegistry()`** — Pre-populates Classic, Modern, Sci-Fi, Shoji themes
- **Render functions** — Pure functions: `(content string, width int, selected bool) string`
- **Config persistence** — `theme: modern` in `~/.threedoors/config.yaml`
- **Golden file tests** — Multi-width, selected/unselected state verification

The existing architecture is explicitly designed for extensibility. Seasonal themes are additive — no existing code requires modification beyond adding new registrations.

---

## 2. Design Decisions

### DD-S1: Seasonal Themes as Standalone DoorTheme Instances (Replacement Model)

**Decision:** Each seasonal theme is a self-contained `DoorTheme` with its own render function, not an overlay or decorator on existing themes.

**Rationale:**
- Overlays add complexity (composing render functions, merging color palettes) with marginal benefit
- Standalone themes are independently testable via golden files
- Matches the existing pattern — Modern, Sci-Fi, Shoji are all standalone
- Each seasonal theme can have its own MinWidth, Colors, and Description

**Alternatives Rejected:**
- Decorator/overlay pattern — too complex for the value, creates coupling between base and seasonal themes
- CSS-like theme inheritance — not a concept in Go/Lipgloss, would require custom framework

### DD-S2: SeasonalResolver as Pure Function

**Decision:** Season-to-theme mapping is a pure function, not an interface or struct.

```go
// ResolveSeason returns the season name for the given date,
// or empty string if seasonal themes are disabled or no match.
func ResolveSeason(now time.Time, ranges []SeasonRange) string
```

**Rationale:**
- Pure function is trivially testable with table-driven tests
- No state, no I/O, no dependencies
- Single responsibility: date → season name
- Caller decides what to do with the result (look up theme, fall back, etc.)

### DD-S3: Season Metadata on DoorTheme Struct

**Decision:** Extend `DoorTheme` with optional seasonal metadata fields.

```go
type DoorTheme struct {
    Name        string
    Description string
    Render      func(content string, width int, selected bool) string
    Colors      ThemeColors
    MinWidth    int
    // Seasonal metadata (zero values for non-seasonal themes)
    Season      string // "spring", "summer", "autumn", "winter", or ""
    SeasonStart MonthDay // {Month: 3, Day: 1} for spring
    SeasonEnd   MonthDay // {Month: 5, Day: 31} for spring
}

type MonthDay struct {
    Month int
    Day   int
}
```

**Rationale:**
- Zero-value `Season` ("") means non-seasonal — backward compatible
- `MonthDay` avoids year dependency and timezone complexity
- Registry can filter by `Season != ""` to list seasonal themes
- No separate seasonal metadata store needed

### DD-S4: Auto-Switch in DoorsView Initialization

**Decision:** Seasonal theme resolution happens at DoorsView construction time, not on every render.

**Flow:**
1. App startup reads `config.yaml` for `theme` and `seasonal_themes` settings
2. If `seasonal_themes: true` (default), call `ResolveSeason(time.Now().UTC(), defaultSeasonRanges)`
3. If a matching seasonal theme exists in registry, use it instead of configured theme
4. Store resolved theme reference in DoorsView — no per-render checks
5. Planning session start also re-checks (handles overnight sessions crossing season boundaries)

**Rationale:**
- Season doesn't change during a session (sessions are minutes, not months)
- Avoids per-render overhead (even though it's negligible)
- Simple to test — mock the time, verify theme selection

### DD-S5: Config Schema Extension

**Decision:** Add `seasonal_themes` boolean to config.yaml.

```yaml
theme: modern              # base theme (existing)
seasonal_themes: true      # enable auto-switching (new, default: true)
```

**Rationale:**
- Minimal config addition — single boolean
- Default `true` provides delight out of the box
- Users who prefer their chosen theme year-round set `false`
- No date-range configuration in config (use code defaults, configurable date ranges are YAGNI for v1)

### DD-S6: Unicode Character Constraint

**Decision:** All seasonal themes use only characters from NFR17-approved ranges:
- Box-drawing: `U+2500–U+257F`
- Block elements: `U+2580–U+259F`
- Geometric shapes: `U+25A0–U+25FF`

**No emoji. No Unicode symbols outside these ranges.**

Seasonal identity comes from *patterns*, not *symbols*:
- **Winter:** Dense dot patterns (`·`), crystalline angular frames (`╬`, `╪`)
- **Spring:** Flowing curved lines (`╭╮╰╯`), light open patterns
- **Summer:** Radiating lines (`╋`), bold geometric shapes (`◆`, `●`)
- **Autumn:** Layered block elements (`▒▓`), angular textures

---

## 3. Component Architecture

### New Files

| File | Purpose |
|------|---------|
| `internal/tui/themes/seasonal.go` | `SeasonRange`, `MonthDay`, `ResolveSeason()`, `DefaultSeasonRanges` |
| `internal/tui/themes/seasonal_test.go` | Table-driven tests for date-range resolution |
| `internal/tui/themes/winter.go` | `NewWinterTheme() *DoorTheme` |
| `internal/tui/themes/spring.go` | `NewSpringTheme() *DoorTheme` |
| `internal/tui/themes/summer.go` | `NewSummerTheme() *DoorTheme` |
| `internal/tui/themes/autumn.go` | `NewAutumnTheme() *DoorTheme` |
| `internal/tui/themes/accessibility_test.go` | WCAG contrast ratio validation helpers |

### Modified Files

| File | Change |
|------|--------|
| `internal/tui/themes/theme.go` | Add `Season`, `SeasonStart`, `SeasonEnd`, `MonthDay` to `DoorTheme` |
| `internal/tui/themes/registry.go` | Add `GetBySeason(season string)` method; register seasonal themes in `NewDefaultRegistry()` |
| `internal/tui/doors_view.go` | Add seasonal theme resolution at construction time |
| `internal/tui/theme_picker.go` | Add seasonal category to theme picker |

### Data Flow

```
App Startup
    │
    ├─ Load config.yaml → theme: "modern", seasonal_themes: true
    │
    ├─ ResolveSeason(time.Now().UTC(), DefaultSeasonRanges) → "winter"
    │
    ├─ Registry.GetBySeason("winter") → *DoorTheme{Name: "winter", ...}
    │
    └─ DoorsView uses winter theme for this session
```

### Registry Extension

```go
// GetBySeason returns the seasonal theme for the given season name, or false.
func (r *Registry) GetBySeason(season string) (*DoorTheme, bool) {
    for _, t := range r.themes {
        if t.Season == season {
            return t, true
        }
    }
    return nil, false
}
```

---

## 4. Testing Strategy

### SeasonalResolver Tests (Table-Driven)

| Test Case | Date | Expected Season |
|-----------|------|----------------|
| Spring start | March 1 | spring |
| Spring end | May 31 | spring |
| Summer start | June 1 | summer |
| Winter boundary | December 31 | winter |
| Winter wrap | January 15 | winter |
| Leap year | February 29 | winter |

### Golden File Tests

24 golden files: 4 seasons × 3 widths (min, 80, 120) × 2 states (selected, unselected).

### Accessibility Tests

Programmatic WCAG AA contrast ratio validation:
- Extract foreground/background Lipgloss color values
- Compute relative luminance per WCAG 2.0 formula
- Assert ratio >= 4.5:1 for all text elements
- Test against both dark (background #000000) and light (background #FFFFFF) terminal schemes

---

## 5. Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Curved box-drawing chars (`╭╮`) render inconsistently | Spring theme is Tier 2 risk — test across iTerm2, Terminal.app, Alacritty before shipping |
| Season date ranges don't match user expectations | Default to widely-accepted meteorological seasons; document in help text |
| Seasonal themes feel gimmicky | Ensure themes are subtle and tasteful, not novelty — patterns, not pictures |

---

## 6. Out of Scope

- Holiday-specific themes (Halloween, Christmas) — future epic
- User-defined seasonal date ranges in config.yaml — YAGNI for v1
- Hemisphere-aware season detection — note for future
- Animated season transitions — not planned
- Community-contributed seasonal themes — future extensibility
