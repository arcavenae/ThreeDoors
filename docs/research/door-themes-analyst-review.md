# Door Theme System — Analyst Review

**Date:** 2026-03-03
**Reviewer:** Business Analyst (Mary)
**Source:** `docs/research/door-themes-research.md` (PR #116)
**Status:** Reviewed — Recommended for Implementation

---

## 1. Executive Assessment

The door theme system is a **high-feasibility, low-risk feature** that directly supports ThreeDoors' core value proposition. The research document is thorough, technically grounded, and proposes a pragmatic architecture. The recommendation to start with three high-reliability themes (Modern, Sci-Fi, Shoji) is sound.

**Overall Recommendation:** Proceed with implementation as a new epic. The theme system adds visual personality to the Three Doors interface without compromising the core UX principles of reduced friction and "progress over perfection."

---

## 2. Feasibility Assessment

### Architecture: DoorTheme Struct + Render Functions

The proposed `DoorTheme` struct with render functions is the right approach:

- **Simplicity:** No interface hierarchies, no abstract factories — just a slice of render functions. This aligns with the project's Go idiom of "a little copying is better than a little dependency."
- **Integration point is clean:** Replacing `style.Render(content)` with `theme.Render(content, width, selected)` in `DoorsView.View()` is a minimal, low-risk change to the existing TUI code.
- **Testability:** Each theme is a pure function (content in → styled string out), making golden file testing straightforward. The project already has `golden_test.go` infrastructure.

**Risk:** Custom frame drawing bypasses Lipgloss `Border()`, meaning themes must handle padding, alignment, and width calculation manually. This is manageable but increases per-theme code complexity.

### Terminal Compatibility

The feasibility matrix in the research is accurate:

| Risk Level | Themes | Character Classes Used |
|-----------|--------|----------------------|
| **Low** | Modern, Sci-Fi, Shoji, Classic Wooden, Vault | Box-drawing (`─│┌┐═║╔╗`), block elements (`░▒▓`), basic shapes (`●◉`) |
| **Medium** | Castle, Saloon, Garden Gate | Curved box-drawing (`╭╮╰╯`), decorative Unicode (`◠◡⚒╳`) |

The low-risk themes use characters supported by virtually all modern terminal emulators (iTerm2, Alacritty, kitty, Windows Terminal, GNOME Terminal). Medium-risk themes rely on characters with inconsistent width rendering across fonts.

### Performance

Theme rendering is pure string manipulation — no I/O, no computation beyond string building. Performance is a non-concern. Even the most complex theme adds microseconds to a render cycle triggered by keypress events.

---

## 3. Theme Practicality Ranking (v1 Recommendations)

### Tier 1 — Ship in v1 (High feasibility, strong visual identity)

1. **Modern / Minimalist** — Simplest theme, maximum reliability. Clean lines, negative space, asymmetric doorknob. Functions as the "safe default."
2. **Sci-Fi / Spaceship** — Strong visual identity using only double-line box-drawing (universal support). The panel-and-rivet aesthetic is distinctive. Minor risk: `◈` rivet character.
3. **Japanese Shoji** — Pure grid of `┼─│` characters — zero rendering risk. The lattice pattern is visually unique and scales naturally by adding/removing grid cells.

### Tier 2 — Ship in v1.1 (Moderate risk, needs cross-terminal testing)

4. **Classic Wooden** — `▒` fill for wood grain is well-supported. Doorknob `◉` is reliable. Good candidate for quick addition.
5. **Vault / Safe** — Standard double-line frame. Only risk is `╳` crosshatch inside the handle — easily simplified.

### Tier 3 — Defer (Higher risk, complex alignment)

6. **Castle / Medieval** — Arch requires `╭╮╰╯` curves plus precise alignment. `◆` studs and `⚒` symbol have variable widths.
7. **Saloon / Western** — Complex hinge bracket construction. Horizontal slat alignment is fragile.
8. **Garden Gate** — `◠◡` wave characters are the highest rendering risk in the entire set.

---

## 4. Product Fit Analysis

### Alignment with Core Philosophy

The theme system aligns strongly with ThreeDoors' product vision:

- **"Personal achievement partner disguised as a todo app"** — Themed doors add personality and delight. The doors become characters, not just containers.
- **"Works with human psychology"** — Visual variety combats interface fatigue. Distinct door appearances create stronger mental associations with tasks, potentially improving recall and engagement.
- **"Progress over perfection"** — Theme selection during onboarding is a low-stakes, fun first interaction. No wrong answer. Immediate visual payoff.

### Alignment with Existing Features

| Feature | Theme System Interaction |
|---------|------------------------|
| **First-run onboarding (Epic 10, FR38)** | Theme picker is a natural addition to the welcome flow — browse previews, select a set |
| **Config.yaml (FR29, FR32)** | Theme preference stored alongside other user settings |
| **Per-door colors** | Themes can incorporate or replace the existing color system — the research addresses this as Open Question #5 |
| **Session metrics** | Theme selection events can be logged for pattern analysis (which themes correlate with higher engagement?) |
| **Golden file tests** | Theme rendering output is ideal for golden file comparison |

### Strategic Value

- **Low cost, high delight:** Theme implementation is estimated at ~3 theme files + integration in `DoorsView.View()`. The effort-to-impact ratio is excellent.
- **Extensibility:** The theme registry pattern makes it trivial to add community-contributed themes later.
- **Differentiation:** No other TUI task manager offers themed visual frames. This is a unique feature that reinforces ThreeDoors' identity.

---

## 5. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Unicode characters render at wrong width in some terminals | Medium | Medium | Stick to Tier 1 themes for v1; test across iTerm2, Terminal.app, Alacritty |
| Theme rendering breaks at narrow terminal widths | Medium | Low | Define minimum width per theme; fall back to simple bordered box below threshold |
| Theme selection adds friction to onboarding | Low | Medium | Make theme selection optional with a sensible default; allow skipping |
| Scope creep (seasonal variants, animated themes, etc.) | Medium | Low | Defer all enhancements to post-v1; keep initial scope to 3 static themes |
| Per-door color system conflicts with theme colors | Low | Low | Theme `Colors` struct overrides per-door colors when a theme is active |

---

## 6. Open Questions — Analyst Recommendations

From the research document's open questions:

1. **Should themes persist across sessions?** → **Yes.** Store in config.yaml. Users form attachment to their chosen aesthetic. Re-randomizing would feel jarring.

2. **Should door number labels be part of the theme frame?** → **Overlaid separately.** Door numbers are functional UI, not decorative. Keep them consistent across themes.

3. **Do we want a theme preview command?** → **Yes, in settings view.** Essential for theme selection UX. Show all available themes rendered at standard width.

4. **Should themes have seasonal variants?** → **Defer.** Fun idea, but pure scope creep for v1. Note as future opportunity.

5. **How do themes interact with per-door colors?** → **Theme colors take precedence.** Each `DoorTheme` defines its own `ThemeColors` struct. The existing per-door color system becomes the fallback for the "Classic" (no-theme) mode.

---

## 7. Summary

The door theme system is well-researched, architecturally sound, and strategically aligned with ThreeDoors' product vision. The recommended starting trio (Modern, Sci-Fi, Shoji) provides strong visual contrast using only reliable Unicode characters.

**Recommendation:** Create Epic 17: Door Theme System with requirements covering theme registry, theme picker in onboarding, settings view for theme changes, and config.yaml persistence.

---

*Review conducted as part of the BMAD methodology analyst workflow.*
