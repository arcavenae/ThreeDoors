# Testing Frame-Based Animations in Bubbletea

## Overview

This document describes the testing pattern for spring-physics animations using
`charmbracelet/harmonica` in a Bubbletea TUI. It was produced by Story 41.5 (Harmonica
Door Transition Spike).

## The Challenge

Frame-based animations are time-dependent: a spring moves toward its target over
multiple frames driven by `tea.Tick`. This creates two testing problems:

1. **Non-determinism**: real timers produce variable frame counts across runs
2. **Golden file instability**: mid-animation frames vary with system load/timing

## Testing Strategy

### Unit-Test the Animation Model Directly

The `DoorAnimation` struct is decoupled from Bubbletea. Its `Update()` method
advances one frame deterministically — no timers involved. This is the primary
testing surface.

```go
da := NewDoorAnimation()
da.SetSelection(1)

// Drive frames manually
for da.Update() {
    // each call advances exactly one spring step
}

// Assert final state
if da.Emphasis(1) != 1.0 { ... }
```

**What to assert:**
- Convergence: animation settles within a bounded frame count
- Target accuracy: emphasis reaches target values after settling
- Initial motion: first frame moves toward target (not away)
- Boundary: out-of-range indices return zero
- Re-selection: switching targets mid-animation still converges

### Golden File Strategy

**Snapshot only settled states.** Do NOT snapshot mid-animation frames — they are
inherently non-deterministic in integration tests because `tea.Tick` timing varies.

- **Initial state** (before any selection): all doors at emphasis 0.0 — stable
- **Final state** (after animation settles): selected door at emphasis 1.0 — stable
- **Mid-animation**: test via unit tests on the model, not golden files

### Deterministic Timing for CI

The animation model is driven by `harmonica.FPS(60)` — each `Update()` call
simulates exactly 1/60th of a second regardless of wall-clock time. This means:

- **Unit tests are fully deterministic** — no real timers
- **Integration tests** (teatest) should wait for animation to settle before
  asserting, or test with animation disabled

### Integration with teatest

For teatest-based smoke tests:

1. Send the key that triggers selection
2. Wait long enough for the animation to settle (>2s at default spring params)
3. Assert the final rendered output

```go
tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
// Wait for spring to settle — generous timeout
time.Sleep(3 * time.Second)
out := tm.FinalOutput(t)
// Assert final state, not mid-animation
```

Alternatively, disable animations in test mode by not initializing `DoorAnimation`
or by providing a "skip animation" flag.

## Performance Considerations

- Spring computation is trivial: three multiply-add operations per frame per door
- At 60fps with 3 doors, this is ~180 floating-point ops/second — negligible
- The `tea.Tick` message scheduling is the only overhead, and Bubbletea handles
  this efficiently
- No allocations in the hot loop (`Update` mutates in place)
- Animation settles in ~60–120 frames (1–2 seconds) — no long-running loops

## Spring Parameter Tuning

Current defaults (chosen for "snappy but bouncy" feel):
- **Frequency**: 6.0 Hz — fast response
- **Damping**: 0.4 — moderate bounce (visible overshoot, settles quickly)

Adjust in `internal/tui/animation.go` constants. Higher damping (→1.0) reduces
bounce. Lower frequency slows the animation.
