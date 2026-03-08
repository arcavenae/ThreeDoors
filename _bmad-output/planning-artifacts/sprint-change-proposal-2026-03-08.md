# Sprint Change Proposal — Seasonal Door Theme Variants

**Date:** 2026-03-08
**Proposed by:** BMAD Correct Course Workflow
**Change Scope:** Minor — New epic, no existing work affected

---

## Section 1: Issue Summary

**Problem Statement:** ThreeDoors' door theme system (Epic 17, COMPLETE) currently provides four static themes (Classic, Modern, Sci-Fi, Shoji). The analyst review (2026-03-03) identified seasonal/holiday theme variants as a deferred opportunity. With Epic 17's infrastructure now fully implemented and proven — DoorTheme struct, Registry, render functions, config persistence, golden file testing — the foundation exists to add time-based visual variety.

**Context:** The analyst review explicitly noted seasonal variants as "Fun idea, but pure scope creep for v1. Note as future opportunity." With v1 theme system complete and stable, this proposal creates a new epic to realize that opportunity.

**Evidence:**
- Epic 17 complete: 6/6 stories merged (PRs #119-#124)
- Theme infrastructure proven: `DoorTheme` struct, `Registry`, render functions, golden file tests
- Analyst recommendation: "Proceed post-v1" (Section 6, Q4)
- Existing themes use only safe Unicode ranges (box-drawing, block elements, geometric shapes)

---

## Section 2: Impact Analysis

### Epic Impact
- **No existing epics affected.** This is a purely additive new epic (proposed as Epic 33).
- **Prerequisite:** Epic 17 (Door Theme System) — already COMPLETE.
- **No blocking dependencies** on any in-progress or planned work.

### Story Impact
- No current or future stories require changes.
- New stories will extend the existing `internal/tui/themes/` package.

### Artifact Conflicts
- **PRD:** Needs new functional requirements (FR132+) for seasonal themes.
- **Architecture:** Minor extension to components doc — seasonal theme metadata, time-based selection logic.
- **UI/UX:** Theme picker may need seasonal category; settings view needs auto-switch toggle.

### Technical Impact
- **Code:** Extend `DoorTheme` struct with optional season/date metadata. Add seasonal theme files. Add auto-switch logic in DoorsView.
- **Infrastructure:** No changes needed.
- **Deployment:** No changes needed.

---

## Section 3: Recommended Approach

**Selected Path:** Direct Adjustment — Add new Epic 33 with focused stories.

**Rationale:**
- Epic 17's architecture was explicitly designed for theme extensibility (registry pattern, render functions)
- Adding seasonal themes is a natural, low-risk extension
- No rollback or MVP scope changes needed
- Effort: Medium (4-5 stories)
- Risk: Low (building on proven infrastructure)
- Timeline impact: None on existing P1 work

**Alternatives Considered:**
- Bundling into Epic 17 — rejected (Epic 17 is complete and merged)
- Deferring indefinitely — rejected (infrastructure is ready, value is clear)

---

## Section 4: Detailed Change Proposals

### 4.1 PRD Changes

**File:** `docs/prd/requirements.md`
**Section:** New requirements block after FR62

**NEW Requirements:**
```
**Seasonal Door Theme Variants:**

FR132: The system shall provide seasonal door theme variants that apply
time-appropriate visual styling (e.g., autumn leaves for fall, snowflakes
for winter, cherry blossoms for spring, sun motifs for summer), using only
Unicode characters from the safe rendering ranges (NFR17).

FR133: The system shall support automatic seasonal theme switching based
on the current date, with configurable season date ranges stored in
config.yaml — defaulting to meteorological seasons (spring: Mar 1,
summer: Jun 1, autumn: Sep 1, winter: Dec 1).

FR134: The system shall allow users to opt out of automatic seasonal
switching via a `seasonal_themes: false` setting in config.yaml,
reverting to their manually selected theme.

FR135: The system shall provide a `:seasonal` command in the TUI that
previews all seasonal variants and allows manual season override for
testing or preference.

FR136: All seasonal theme variants shall maintain WCAG AA contrast ratios
(4.5:1 minimum for text) in both light and dark terminal color schemes,
ensuring readability is never sacrificed for visual novelty.

FR137: Seasonal themes shall gracefully fall back to the user's base
theme when the terminal width is below the seasonal variant's declared
minimum width, consistent with FR61 behavior.
```

### 4.2 Architecture Changes

**File:** `docs/architecture/components.md`
**Impact:** Add seasonal theme metadata to DoorTheme struct description. Add SeasonalThemeLoader component.

**NEW:**
```
ThemeMetadata extension:
- Season field (spring/summer/autumn/winter/holiday)
- DateRange (start month-day, end month-day)
- BasedOn field linking seasonal variant to its parent theme
- Auto-switch logic checks current date against registered seasonal themes

SeasonalThemeLoader:
- On app startup, check date against seasonal theme date ranges
- If seasonal_themes enabled and a matching seasonal theme exists, override active theme
- Emit theme_switch event to session metrics
```

### 4.3 Epics & Stories Changes

**File:** `docs/prd/epics-and-stories.md`
**Action:** Add Epic 33: Seasonal Door Theme Variants

**NEW Epic 33** with stories covering:
1. Seasonal theme metadata model and date-range matching
2. Seasonal theme implementations (4 seasons)
3. Auto-switch integration in DoorsView
4. Seasonal theme picker and `:seasonal` command
5. Accessibility validation and golden file tests

---

## Section 5: Implementation Handoff

**Change Scope Classification:** Minor — Direct implementation by development team.

**Handoff Recipients:**
- **PM Agent:** Update PRD with new FRs (FR132-FR137)
- **Architect Agent:** Update architecture docs with seasonal theme components
- **SM Agent:** Add Epic 33 to ROADMAP.md and sprint planning
- **Dev Team:** Implement stories once planning artifacts are finalized

**Success Criteria:**
- New FRs added to PRD
- Epic 33 created with 4-5 stories
- Architecture updated with seasonal theme components
- All seasonal themes pass golden file tests and accessibility checks
- ROADMAP.md updated with Epic 33

---

## Approval

**Status:** Approved (automated BMAD pipeline)
**Next Steps:** Proceed to Party Mode discussion, PRD update, architecture design, and epic/story creation.
