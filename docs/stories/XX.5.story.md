# Story XX.5: README Visual Demo Section

## Status: Not Started

## Epic

Epic XX: README Overhaul

## References

- Plan: `_bmad-output/planning-artifacts/readme-overhaul-plan.md` (Section 8)

## Story

As a potential user visiting the repository,
I want to see what ThreeDoors looks like before installing it,
So that I can quickly understand the product and decide if it's worth trying.

## Background

The README has no visual element showing what ThreeDoors looks like. Best-in-class TUI project READMEs (bubbletea, lazygit, glow) all feature hero GIFs or screenshots prominently. Since capturing actual terminal recordings requires running the application, this story creates placeholder infrastructure and an ASCII art mockup that works immediately.

## Acceptance Criteria

**Given** the README
**When** a user views it on GitHub
**Then** they see a visual demo section after "What is ThreeDoors?" showing what the TUI looks like

**Given** the visual demo
**When** rendered on GitHub
**Then** it includes either an ASCII art mockup of the three doors view OR a placeholder with clear instructions for adding a GIF later

**Given** the `docs/assets/` directory
**When** screenshots or GIFs are added in the future
**Then** the directory exists and the README references it

**Given** the Screenshots section
**When** a user expands it (foldable)
**Then** they see a gallery layout with placeholders for key views: doors, task detail, dashboard, themes, search, onboarding

## Technical Notes

- ASCII art mockup is preferred over placeholder text — it provides immediate value and is maintainable
- The mockup should reflect the actual door rendering style (use existing golden file tests in `internal/tui/themes/testdata/` as reference)
- If creating a GIF, use `charmbracelet/vhs` — it's the Charm ecosystem tool for terminal recordings
- Store images in `docs/assets/` directory (create if needed)
- Max image width: 800px for GitHub rendering
- This story creates the section structure and ASCII mockup; actual GIF capture can be a follow-up

## Tasks

- [ ] Create `docs/assets/` directory if it doesn't exist
- [ ] Create ASCII art mockup of the three doors view (reference golden files in themes/testdata/)
- [ ] Add `## 📸 Screenshots` section to README after "What is ThreeDoors?"
- [ ] Add hero visual (ASCII mockup or centered placeholder)
- [ ] Add foldable gallery section with placeholders for future screenshots
- [ ] Add comment noting `charmbracelet/vhs` as recommended tool for future GIF capture
- [ ] Update story status to Done
