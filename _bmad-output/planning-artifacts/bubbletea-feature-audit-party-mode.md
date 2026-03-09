# Bubbletea Feature Audit & UX Improvement Roadmap

> Party mode artifact — 2026-03-09
> Agents: UX Designer (Sally), Architect (Winston), TEA (Murat), PM (John), Creative Problem Solver (Dr. Quinn)

## Executive Summary

Deep audit of the charmbracelet ecosystem vs ThreeDoors' current usage. Of 15 unused components/libraries identified, the panel recommends **5 adoptions**, **2 conditional/deferred items**, and **8 rejections**. One new epic proposed (Epic 41: Charm Ecosystem Adoption & TUI Polish).

---

## Phase 1: Ecosystem Audit

### Current Usage

| Library | Version | What We Use |
|---------|---------|-------------|
| bubbletea | v1.3.10 | Core MVU framework — tea.Model, tea.Cmd, tea.Msg, KeyMsg |
| lipgloss | v1.1.0 | 60+ styles, JoinHorizontal, borders (Rounded/Double), 256-color palette |
| bubbles | v1.0.0 | **textinput ONLY** — used in search, add-task, onboarding, values views |
| x/exp/teatest | latest | Headless TUI testing via TestModel |
| x/exp/golden | latest | Golden file snapshot testing |
| x/ansi | v0.11.6 | ANSI utilities (indirect) |

### Unused Components Identified

| # | Component | Category | Description |
|---|-----------|----------|-------------|
| 1 | bubbles/viewport | Scrolling | Scrollable content area with mouse wheel support |
| 2 | bubbles/spinner | Feedback | Loading/activity indicator |
| 3 | bubbles/progress | Feedback | Progress bar |
| 4 | bubbles/list | Selection | Scrollable item selection with filtering |
| 5 | bubbles/help | Display | Built-in keybinding help renderer |
| 6 | bubbles/paginator | Navigation | Multi-page dot/number navigation |
| 7 | bubbles/table | Display | Tabular data rendering |
| 8 | bubbles/textarea | Input | Multi-line text editor |
| 9 | bubbles/timer/stopwatch | Display | Time tracking display |
| 10 | harmonica | Animation | Spring physics, smooth transitions, easing |
| 11 | huh | Input | Form building library with validation |
| 12 | glamour | Display | Terminal markdown rendering |
| 13 | lipgloss.JoinVertical | Layout | Vertical content composition |
| 14 | lipgloss.Place | Layout | Centered/positioned content placement |
| 15 | Adaptive colors | Accessibility | Terminal-aware color degradation |

### Custom Implementations That Overlap With Charm Components

| Our Code | Overlapping Charm Component | Files |
|----------|---------------------------|-------|
| Custom scrolling (3 implementations) | bubbles/viewport | help_view.go, synclog_view.go, keybinding_overlay.go |
| Custom cursor-based selection | bubbles/list | devqueue_view.go, theme_picker.go, search_view.go |
| Custom pagination | bubbles/paginator | help_view.go |
| Custom help display | bubbles/help | help_view.go, keybinding_overlay.go |

---

## Phase 2: Party Mode Critical Review

### Panel Discussion Summary

**Key Insight (Dr. Quinn):** Two fundamentally different categories of improvement:

- **Category A: REPLACEMENT** — Swapping custom code for standard library equivalents (viewport, lipgloss layout). Reduces maintenance burden regardless of future plans. Should be prioritized.
- **Category B: ADDITION** — Adding new capabilities (harmonica, spinner). Genuine feature additions with higher risk. Evaluate individually.

### Per-Component Analysis

#### 1. bubbles/viewport — ADOPT (Replace custom scrolling)

- **UX Impact (Sally):** Users experience inconsistent scroll behavior across 3 custom implementations. Viewport gives them muscle memory. Mouse wheel scrolling is a free accessibility win.
- **Architecture (Winston):** Three hand-rolled scrolling systems = three sets of edge cases. Viewport is battle-tested across hundreds of charm apps. Highest-ROI change on the list — removes maintenance burden.
- **Testing (Murat):** MEDIUM risk. Golden file tests need updating (different ANSI output). But viewport has its own test helpers. Net positive — fewer custom scroll bugs.
- **PM (John):** Scrolling standardization makes every future scrollable view better (Epic 40 stats, any new panels).
- **Effort:** 3 | **Impact:** 4 | **Future Value:** +2 | **Adjusted Score: 2.0**

#### 2. bubbles/spinner — ADOPT (Async feedback)

- **UX Impact (Sally):** When Todoist syncs or providers load, users see nothing. That "nothing" is uncertainty. A spinner says "I'm working on it." The difference between "is it frozen?" and "ah, it's thinking."
- **Architecture (Winston):** Trivial to add. Spinner model embeds into any view. Deterministic frame rendering for tests.
- **Testing (Murat):** LOW risk. Spinner has deterministic frame rendering. Freeze frame for golden tests using spinner.Tick control.
- **PM (John):** Trivial effort, clear value. Ship it.
- **Effort:** 1 | **Impact:** 3 | **Future Value:** +0 | **Score: 3.0**

#### 3. lipgloss.Place + JoinVertical — ADOPT (Layout quality of life)

- **Architecture (Winston):** We only use JoinHorizontal. Adding JoinVertical eliminates manual `\n` concatenation. Place simplifies layout math we're doing manually for centering.
- **Testing (Murat):** NEGLIGIBLE risk. Pure layout functions. Golden tests catch any rendering differences immediately.
- **PM (John):** Infrastructure improvement. Every layout gets cleaner.
- **Effort:** 1 | **Impact:** 2 | **Future Value:** +1 | **Score: 3.0**

#### 4. harmonica (door transitions) — ADOPT with SPIKE

- **UX Impact (Sally):** SOUL.md says "the UI should feel like physical objects — doors that open, selections that click into place." Current transitions are instantaneous state swaps. Spring physics delivers the brand promise. This is the difference between "I clicked a thing" and "I opened a door."
- **Architecture (Winston):** Spring animations require frame-by-frame tea.Cmd ticks. Update() loop gets busier during transitions. Not a problem per se, but golden file tests need deterministic handling. Scope to door transitions ONLY.
- **Testing (Murat):** HIGH risk. Animations are time-dependent. Need to test initial state, final state, and verify animation command chain. Recommend spike story first — implement on ONE transition, verify testing pattern, then expand.
- **PM (John):** High impact but needs spike. Don't commit to full animation system before test pattern is proven.
- **Effort:** 4 (with spike: 2) | **Impact:** 4 | **Future Value:** +1 | **Score: 1.25 (spike: 2.5)**

#### 5. Adaptive color profiles — ADOPT (Accessibility)

- **Architecture (Winston):** Lipgloss supports TrueColor, ANSI256, ANSI, and ASCII profiles. We hardcode 256-color values. Adaptive colors would gracefully degrade in limited terminals. Low effort since Lipgloss handles detection.
- **UX Impact (Sally):** Invisible to most users. But for those in constrained terminals (SSH sessions, old macOS Terminal.app), it prevents garbled output.
- **Effort:** 2 | **Impact:** 2 | **Future Value:** +1 | **Score: 1.5**

#### 6. bubbles/progress — CONDITIONAL

- **PM (John):** Only relevant if Epic 40 stats dashboard needs it. D-096 decided custom Lipgloss sparklines for Phase 1. Don't add speculatively.
- **Verdict:** Adopt ONLY if Epic 40 Phase 2+ needs a progress bar. Don't import preemptively.
- **Effort:** 2 | **Impact:** 2 | **Score: 1.0 (conditional)**

#### 7. glamour (markdown rendering) — DEFER

- **Architecture (Winston):** Markdown rendering is useful but not blocking anything today. Tech-stack.md already deferred this to Phase 3.
- **PM (John):** What markdown are we rendering? No current use case.
- **Effort:** 3 | **Impact:** 2 | **Score: 0.67 (defer)**

### Rejected Components (8 items)

| Component | Why Rejected | SOUL.md Alignment |
|-----------|-------------|-------------------|
| **bubbles/list** | Fights the 3-door constraint. Our custom cursor UIs serve ThreeDoors-specific needs. Generic list would add selection modes we don't want. Prior decision X-068 already rejected this for command completion. | "Three Doors, Not Three Hundred" |
| **bubbles/help** | Our keybinding registry integration is too deep (D-090). Bubbles help is too simple for our multi-category, view-aware system. | Custom is better here |
| **bubbles/table** | No tabular data in ThreeDoors. We show 3 tasks, not a spreadsheet. | "Not a project management tool" |
| **bubbles/textarea** | Multi-line input contradicts friction-reduction philosophy. Single-step captures via textinput are intentional. | "Reduce friction to starting" |
| **bubbles/filepicker** | Path input in onboarding is fine as textinput. File picker adds UI complexity for rare operation. | Keep it simple |
| **bubbles/timer/stopwatch** | Session tracking is backend (JSONL), not displayed. D-077 decided soft progress (step counter), not timer. | "No guilt" — visible timers create pressure |
| **huh** (forms) | Over-engineered for our input needs. We have at most 2-step input flows. Textinput is sufficient. | Don't add complexity |
| **wish** (SSH) | Local-only application per SOUL.md. No remote terminal mode needed. | "Local-First, Privacy-Always" |

### Custom Implementations: Keep vs Replace

| Custom Implementation | Verdict | Rationale |
|----------------------|---------|-----------|
| Custom scrolling (3 views) | **REPLACE with viewport** | Three different behaviors confuse users. Viewport adds mouse wheel for free. |
| Custom cursor selection | **KEEP** | Our cursor UIs are intentionally constrained for ThreeDoors (3 doors, small menus). Generic list fights this. |
| Custom pagination (help_view) | **REPLACE with viewport** | Viewport subsumes pagination. Continuous scrolling > page-based for help content. |
| Custom help display | **KEEP** | Deep integration with keybinding registry (D-090). Bubbles help can't express our category system. |
| Theme system | **KEEP** | DoorTheme interface is home-grown for good reason — no charm equivalent exists for custom door rendering. |
| Keybinding bar | **KEEP** | Responsive width breakpoints and height-based visibility (D-094) are ThreeDoors-specific requirements. |

---

## Phase 3: Priority Matrix

### Ranked Priority List

| Rank | Improvement | Effort (1-5) | Impact (1-5) | Future Value | Adjusted Score | Category | Type |
|------|------------|-------------|-------------|--------------|----------------|----------|------|
| 1 | **Spinner for async feedback** | 1 | 3 | +0 | **3.0** | Addition | New capability |
| 2 | **lipgloss.Place + JoinVertical** | 1 | 2 | +1 | **3.0** | Replacement | Layout QoL |
| 3 | **Viewport scroll standardization** | 3 | 4 | +2 | **2.0** | Replacement | Debt reduction |
| 4 | **Harmonica door transitions (spike)** | 2* | 4 | +1 | **2.5** | Addition | UX signature |
| 5 | **Adaptive color profiles** | 2 | 2 | +1 | **1.5** | Addition | Accessibility |
| 6 | Progress bar (conditional) | 2 | 2 | +0 | 1.0 | Conditional | Epic 40 only |
| 7 | Glamour markdown (defer) | 3 | 2 | +1 | 0.67 | Defer | Future phase |

*Harmonica effort 2 = spike story only; full implementation would be effort 4.

### Story Outlines (Top 5)

#### Story 41.1: Spinner Component for Async Provider Operations

**Title:** Add bubbles/spinner for sync and provider loading feedback
**Description:** Embed spinner model in sync status view and any view that triggers async provider operations. When Todoist sync, Apple Notes fetch, or other provider operations are in flight, show a spinner to indicate activity. Replace the current "no feedback" state.
**Effort:** Small (1-2 days)
**AC:**
- Spinner visible during provider sync operations
- Spinner stops when operation completes or errors
- Golden file tests updated with deterministic spinner frame
- No spinner when operations are instant (<100ms)

#### Story 41.2: Lipgloss Layout Utilities Adoption

**Title:** Adopt lipgloss.JoinVertical and Place for cleaner layout composition
**Description:** Replace manual `\n` concatenation with JoinVertical where appropriate. Use Place for centered content (greeting, empty states). Audit all views for layout improvement opportunities. No functional change — pure refactor.
**Effort:** Small (1-2 days)
**AC:**
- All manual vertical joins replaced with lipgloss.JoinVertical where cleaner
- Centered content uses lipgloss.Place instead of manual padding math
- All golden file tests pass (updated as needed)
- No functional behavior change

#### Story 41.3: Viewport Adoption for Help View

**Title:** Replace custom scrolling in help_view with bubbles/viewport
**Description:** Replace the custom page-based scrolling in help_view.go with bubbles/viewport. This is the first of three viewport migrations. Help view is chosen first because it has the most content and would benefit most from continuous scrolling and mouse wheel support.
**Effort:** Medium (2-3 days)
**AC:**
- help_view uses bubbles/viewport for content display
- Mouse wheel scrolling works in help view
- j/k/up/down keyboard navigation preserved
- Page-up/page-down supported via viewport
- Golden file tests updated
- No regression in help content rendering

#### Story 41.4: Viewport Adoption for Synclog and Keybinding Overlay

**Title:** Migrate synclog_view and keybinding_overlay to bubbles/viewport
**Description:** Complete the viewport migration. Replace custom scrolling in synclog_view.go and keybinding_overlay.go with bubbles/viewport. Create shared viewport factory function for ThreeDoors-specific defaults.
**Effort:** Medium (2-3 days)
**AC:**
- synclog_view uses bubbles/viewport
- keybinding_overlay uses bubbles/viewport
- Shared NewScrollableView() factory function created
- Mouse wheel scrolling works in both views
- All golden file tests updated
- Consistent scroll behavior across all three migrated views

#### Story 41.5: Harmonica Door Transition Spike

**Title:** Spike: Spring-physics door selection animation with harmonica
**Description:** Research spike to determine feasibility of harmonica spring animations for door selection transitions. Implement a proof-of-concept on the door selection → detail view transition. Primary deliverable is the documented testing pattern for frame-based animations in teatest, not the animation itself.
**Effort:** Medium (2-3 days)
**AC:**
- harmonica dependency added to go.mod
- Door selection triggers spring-physics transition (proof of concept)
- Document testing pattern for frame-based animations in teatest
- Golden file test strategy documented (which frame(s) to snapshot)
- Performance verified: no visible lag or CPU spike during animation
- Decision: proceed with full animation system or reject

#### Story 41.6: Adaptive Color Profile Support

**Title:** Terminal-aware color degradation via lipgloss color profiles
**Description:** Leverage lipgloss/termenv color profile detection to gracefully degrade color output in limited terminals. Replace hardcoded 256-color values with adaptive color definitions that work across TrueColor, ANSI256, ANSI, and ASCII terminals.
**Effort:** Small-Medium (2 days)
**AC:**
- Color profile detected at startup
- Styles degrade gracefully in ANSI (16-color) terminals
- Theme colors adapt to terminal capability
- No visual change on modern terminals (TrueColor/ANSI256)
- Tested with ASCII color profile (already used in golden tests)

#### Story 41.7: Viewport-Aware Stats Dashboard (conditional, depends on Epic 40)

**Title:** Stats dashboard uses viewport for scrollable content
**Description:** When Epic 40 stats dashboard is implemented, use viewport for scrollable stats content instead of building another custom scroll system.
**Effort:** Included in Epic 40 stories
**Note:** This is a constraint on Epic 40 implementation, not a separate story.

---

## Decisions Summary

| ID | Decision | Rationale | Alternatives Rejected |
|----|----------|-----------|----------------------|
| D-128 | Adopt bubbles/viewport to replace 3 custom scroll implementations | Standardizes behavior, adds mouse wheel, reduces maintenance (3 impls → 1 dep) | Keep custom (higher maintenance), partial replacement (inconsistent) |
| D-129 | Adopt bubbles/spinner for async operation feedback | Trivial effort, eliminates "is it frozen?" uncertainty, deterministic testing | Custom spinner (reinventing wheel), no feedback (current state, poor UX) |
| D-130 | Adopt lipgloss.JoinVertical + Place for layout | Layout QoL, eliminates manual padding math and `\n` concatenation | Keep manual layout (more code, more bugs) |
| D-131 | Harmonica door transitions via spike-first approach | Spring physics delivers SOUL.md "physical objects" promise, but testing risk requires validation | Full commitment without spike (risky), reject animations entirely (misses SOUL.md promise, but see X-008 prior) |
| D-132 | Reject bubbles/list for ThreeDoors selection UIs | 3-door constraint is intentional; generic list fights core design philosophy | Adopt list everywhere (fights design), hybrid (complex for no benefit) |
| D-133 | Reject bubbles/textarea, table, filepicker, timer, help, huh, wish | Each solves problems ThreeDoors doesn't have or contradicts SOUL.md values | See per-component rejection rationale above |
| D-134 | New Epic 41: Charm Ecosystem Adoption & TUI Polish | 5-7 cohesive stories around ecosystem leverage; too many for infrastructure backlog, too few for multiple epics | Distribute across existing epics (fragments coherent vision), single giant story (too large) |

---

## Epic Recommendation

**Proposed: Epic 41 — Charm Ecosystem Adoption & TUI Polish**

**Description:** Systematically adopt underutilized charmbracelet ecosystem components to reduce custom code maintenance, improve UX consistency, and deliver on SOUL.md's "physical objects" promise. Prioritizes replacements (viewport, layout) over additions (spinner, harmonica).

**Scope:** 6-7 stories (41.1 through 41.6, plus conditional 41.7)
**Priority:** P2 — Nice to have. No stories block other work. Can be interleaved with P1 epics.
**Dependencies:** None hard. Story 41.7 is conditional on Epic 40.

**Relationship to existing epics:**
- Story 41.5 (harmonica spike) extends prior decision X-008 which deferred door animations. This spike validates or invalidates that deferral.
- Story 41.7 constrains Epic 40 implementation (use viewport, don't build custom scroll).
- Stories 41.3-41.4 (viewport) may touch views being modified by Epic 39 (keybinding display). Sequence viewport after Epic 39 to avoid conflicts.

---

## Appendix: SOUL.md Alignment Check

Every recommendation was validated against SOUL.md values:

| Value | How This Audit Aligns |
|-------|----------------------|
| "Keypresses should produce visible, satisfying responses" | Harmonica spring animations, spinner feedback |
| "The UI should feel like physical objects" | Harmonica door transitions |
| "Subtle is not the same as invisible" | Spinner during sync, adaptive colors |
| "Show 3 tasks. Not 5." | Rejecting bubbles/list preserves constraint |
| "Not a project management tool" | Rejecting table, timer, textarea |
| "Progress over perfection" | Spike-first for harmonica, ship spinner immediately |
| "Local-First, Privacy-Always" | Rejecting wish (SSH), keeping everything local |
| "Reduce friction to starting" | Rejecting textarea (multi-line adds friction) |
