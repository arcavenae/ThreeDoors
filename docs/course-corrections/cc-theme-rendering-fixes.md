# Course Correction: Theme Rendering Fixes

**Date:** 2026-03-07
**Triggered by:** Visual QA review of all four door themes
**Epic affected:** 17 — Door Theme System
**New stories:** 17.7, 17.8, 17.9

## Problem Statement

Visual inspection of the four door themes revealed issues ranging from critical rendering bugs to design problems that undermine the core ThreeDoors value proposition (quick task scanning with minimal friction).

## Findings by Theme

### Classic (No Issues)
Working as intended. Clean rounded borders, readable text, clear selected state. This is the baseline.

### Modern (Low Severity)
- Selected vs unselected state difference is too subtle — only line weight changes from `─│` to `━┃`, same color range
- Doorknob `●` sits very low, creating awkward bottom-heavy visual weight
- Excessive empty space without enough visual payoff

### Sci-Fi (Medium Severity)
- Visual noise overwhelms task content — shade rails (`░░`), double-line frame, mid-bar separator, and inner borders compete for attention
- Border overhead consumes 8 characters per door, leaving only ~12 chars of usable text width on 80-column terminals
- The `[ACCESS]` label in a separate lower panel fragments the already-small door space
- Task text gets lost in decoration, defeating the purpose of the app

### Shoji (Critical Severity)
- **Bug:** Raw ANSI escape codes (`[38;5;180m`) visible as literal text in the terminal
- **Root cause:** `countRunes()` in `modern.go:112` (shared across themes) counts raw runes instead of visual width. When `style.Render()` wraps box-drawing characters with ANSI sequences, the width math breaks. Content overflows, and the terminal truncates mid-escape-sequence, leaving raw codes visible.
- **Design issue:** Dense lattice grid (many small cells) creates visual noise. Real shoji screens have large paper panes with thin frames — this renders as graph paper.

## Shared Technical Issue

All custom themes use `countRunes()` for width calculation, which does not account for ANSI escape sequence byte length. The function should be replaced with `ansi.StringWidth()` from `charmbracelet/x/ansi`, which properly handles escape sequences and wide characters.

## Target Experience

The core principle: **every theme must put task text first**. Decoration exists to give the door personality, not to compete with content. A user should be able to glance at any themed door and read the task in under 2 seconds.

**Classic** — Already achieves this. Baseline.

**Modern** — Should feel like a high-end apartment door: clean, quiet, confident. The thin lines and single doorknob say "I don't need to shout." But the selected state needs to actually *glow* — right now it barely whispers. Fix: widen the brightness gap between selected and unselected.

**Sci-Fi** — Should feel like tapping an airlock panel on a spaceship. The double-line frame and shade rails say "industrial" and "mechanical." But the current implementation has so many visual layers (rails + inner borders + mid-bar + separate panels) that it reads like a circuit diagram, not a door. Fix: flatten to one content panel, slim the rails, keep the `[ACCESS]` label as flavor text inside the content area.

**Shoji** — Should feel like looking through a Japanese sliding screen: thin wooden lattice with large paper panes. The grid pattern IS the identity, but the current implementation subdivides into tiny cells that create visual noise at real terminal widths. Fix: use a few horizontal lattice bars with one cross junction instead of a dense grid. The open space between bars IS the paper pane.

## Corrective Actions

### Story 17.7: Fix width calculation and Shoji ANSI leak (Critical)
Replace `countRunes()` with `ansi.StringWidth()`. Fix Shoji `buildContentRow` width math. Add visual width regression tests.

### Story 17.8: Redesign Shoji theme visual layout (Medium)
Reduce grid density — use fewer, larger panes. Increase content-to-decoration ratio.

### Story 17.9: Simplify Sci-Fi theme and improve Modern theme contrast (Medium)
Reduce Sci-Fi decoration overhead. Improve Modern selected state visibility.

## Impact Assessment

- **User impact:** Shoji theme is unusable, Sci-Fi theme is hard to read
- **Code impact:** Localized to `internal/tui/themes/` — no architectural changes needed
- **Risk:** Low — changes are isolated to theme rendering functions with existing golden file test coverage
- **Testing:** Golden files will need regeneration after visual fixes

## Party Mode Participants

- Sally (UX Designer) — visual design analysis
- Winston (Architect) — rendering pipeline diagnosis
- Amelia (Dev) — implementation-level root cause
- Quinn (QA) — testing gap identification
