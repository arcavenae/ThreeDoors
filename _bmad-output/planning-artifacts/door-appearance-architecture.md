# Architecture Assessment: Door Visual Appearance Redesign

**Date:** 2026-03-08
**Source:** Party mode discussion recommendations (door-appearance-party-mode.md)
**Status:** Assessed — Changes Required (Low Risk)

---

## 1. Executive Summary

The door appearance redesign requires **moderate architectural changes** to the theme system. The core change is extending the render function signature to accept height, adding a door anatomy helper, and updating DoorsView to manage vertical space distribution. No new dependencies are needed — all changes use existing box-drawing characters and Lipgloss capabilities.

**Risk Level:** Low. Changes are backward-compatible with compact mode fallback.

---

## 2. Changes Required

### 2.1 Theme Interface Extension

**Current signature:**
```go
Render func(content string, width int, selected bool) string
```

**New signature:**
```go
Render func(content string, width int, height int, selected bool) string
```

**Impact:** All 4 theme render functions must be updated. Classic, Modern, Sci-Fi, and Shoji themes all need the height parameter.

**Migration strategy:** Update all theme render functions simultaneously. The height parameter is additive — existing width logic is preserved.

### 2.2 DoorTheme Struct Extension

Add `MinHeight int` field to `DoorTheme` struct in `internal/tui/themes/theme.go`:

```go
type DoorTheme struct {
    Name        string
    Description string
    Render      func(content string, width int, height int, selected bool) string
    Colors      ThemeColors
    MinWidth    int
    MinHeight   int  // NEW: minimum terminal height for door-like rendering
}
```

### 2.3 DoorAnatomy Helper (New Type)

New file: `internal/tui/themes/anatomy.go`

```go
type DoorAnatomy struct {
    LintelRow      int // row 0 — top frame with door number
    ContentStart   int // row 2 — after lintel + padding
    PanelDivider   int // ~45% of height — upper/lower panel divide
    HandleRow      int // ~60% of height — knob/handle position
    ThresholdRow   int // height-1 — bottom frame/floor
    TotalHeight    int
}

func NewDoorAnatomy(height int) DoorAnatomy { ... }
```

This is a pure-data helper — no interfaces, no dependencies. Each theme uses it to calculate where to place structural elements.

### 2.4 DoorsView Height Management

**Current:** `DoorsView` only manages horizontal space (width distribution across 3 doors).

**New:** `DoorsView` must also calculate available vertical space:
- Terminal height minus header/footer/status chrome
- Pass calculated height to `theme.Render()`
- Compare against `theme.MinHeight` for compact mode fallback

**File:** `internal/tui/doors_view.go` — update `View()` method to calculate and pass height.

### 2.5 Compact Mode Fallback

When terminal height < theme's MinHeight:
- Use current landscape card layout (existing render path)
- No visual degradation — just switches to the card-style rendering

This mirrors the existing MinWidth fallback pattern.

---

## 3. Files Affected

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/tui/themes/theme.go` | Modify | Add MinHeight field, update Render signature |
| `internal/tui/themes/anatomy.go` | **New** | DoorAnatomy helper type |
| `internal/tui/themes/classic.go` | Modify | Update render function for height-aware rendering |
| `internal/tui/themes/modern.go` | Modify | Update render function with door anatomy |
| `internal/tui/themes/scifi.go` | Modify | Update render function with door anatomy |
| `internal/tui/themes/shoji.go` | Modify | Update render function with door anatomy |
| `internal/tui/doors_view.go` | Modify | Height calculation, pass height to Render |
| `internal/tui/themes/anatomy_test.go` | **New** | Tests for DoorAnatomy calculations |
| `internal/tui/themes/*_test.go` | Modify | Update golden files for new proportions |
| `internal/tui/themes/testdata/` | Modify | New golden file baselines |

---

## 4. No Architectural Changes Needed

The following areas require **no changes**:
- **Theme registry** — No changes to registry pattern
- **Config system** — No new config fields (MinHeight is code-defined, not user-configured)
- **Onboarding/theme picker** — Theme picker works with any proportion
- **Data models** — No task model changes
- **Provider pattern** — No storage changes
- **Dependencies** — No new external packages

---

## 5. Design Decisions

### DD-DA1: Height parameter vs. auto-detect in Render

**Decision:** Pass height explicitly to Render function.

**Rationale:** Render functions should be pure (content+dimensions in, string out). Querying terminal size inside Render would add I/O dependency and complicate testing. DoorsView already has terminal dimensions from Bubbletea's WindowSizeMsg.

### DD-DA2: DoorAnatomy as helper struct vs. embedded in DoorTheme

**Decision:** Standalone helper struct, not embedded.

**Rationale:** Each theme may interpret anatomy positions differently (e.g., Shoji puts lattice bars at panel divider; Sci-Fi puts bulkhead). The anatomy provides suggested positions; themes decide how to use them.

### DD-DA3: Compact mode threshold

**Decision:** Per-theme MinHeight (12-16 rows typical) with fallback to current card-style rendering.

**Rationale:** Different themes have different vertical requirements (Shoji needs more rows for lattice pattern). Per-theme thresholds match the existing per-theme MinWidth pattern.

---

## 6. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Golden file churn | High | Low | Expected — regenerate all golden files as part of implementation |
| Vertical overflow on small terminals | Medium | Low | Compact mode fallback prevents broken layout |
| Theme picker preview needs height | Low | Low | Theme picker can use fixed preview height |
| Content truncation in portrait mode | Medium | Low | Word-wrap already handles this; just fewer chars per line |

---

*Assessment conducted following BMAD architecture review methodology.*
