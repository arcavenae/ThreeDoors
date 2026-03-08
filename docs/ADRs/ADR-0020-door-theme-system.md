# ADR-0020: Door Theme System Architecture

- **Status:** Accepted
- **Date:** 2026-02-20
- **Decision Makers:** Analyst review, party mode session
- **Related PRs:** #116-#124, #178, #183, #186
- **Related ADRs:** ADR-0002 (Bubbletea), ADR-0019 (Docker E2E Testing)

## Context

The original "three doors" interface used hardcoded ASCII art. Users requested visual customization. The theme system needed to support multiple visual styles while maintaining terminal compatibility.

## Decision

Implement a **registry-based theme system** with:

1. `DoorTheme` interface defining rendering methods
2. Theme registry for discovery and selection
3. Config persistence for user's theme preference
4. Theme picker in both onboarding flow and `:theme` command
5. Golden file tests for every theme

## Themes Implemented

| Theme | Style | PR |
|-------|-------|-----|
| Classic | ASCII art (original) | #119 (wrapper) |
| Modern | Clean Unicode borders | #120 |
| Sci-Fi | Futuristic with glyphs | #120, #183 |
| Shoji | Japanese sliding doors | #120, #186 |

## Design Decisions

- **No per-door theming** (Decision L9) — single theme applies to all doors for visual cohesion
- **Terminal width fallback** (Decision M1) — falls back to Classic theme below 60 columns
- **Theme preview** (Decision M2) — shows 3 themes side-by-side, arrow keys to scroll, `a`/`w`/`d` to highlight, `s` to shuffle
- **Theme-agnostic door content** — theme controls frame/decoration only, not task content layout

## Consequences

### Positive
- Users can personalize their experience
- Theme interface enables community contributions
- Golden file tests catch visual regressions per theme
- Fallback ensures usability on narrow terminals

### Negative
- Each new theme needs golden file tests for all views
- Theme rendering is one of the most complex view components
- Seasonal theme variants (Epic 33) will add time-based complexity
