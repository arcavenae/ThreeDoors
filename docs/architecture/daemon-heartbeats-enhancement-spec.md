# multiclaude Enhancement Spec: Daemon-Native Heartbeats

**Spec for:** multiclaude daemon changes
**Story:** 73.6 — Daemon-Native Heartbeats
**Date:** 2026-03-29
**Design doc:** `docs/architecture/daemon-native-heartbeats.md`

---

## Summary

Replace the fixed 2-minute `wakeAgents()` loop with a configurable per-agent heartbeat scheduler that supports:
1. Per-agent intervals (5m–20m range, configurable)
2. Activity detection (skip nudges for active agents)
3. Workflow triggers (periodic messages like SYNC_OPERATIONAL_DATA)

This eliminates the last CronCreate dependency and makes heartbeats a first-class daemon feature.

---

## Affected Files

| File | Change Type | Description |
|------|------------|-------------|
| `internal/daemon/daemon.go` | Modify | Replace `wakeAgents()` with heartbeat scheduler |
| `internal/daemon/heartbeat.go` | New | Heartbeat scheduler, activity detection, trigger system |
| `internal/daemon/config.go` | Modify | Add heartbeat config parsing |
| `pkg/tmux/client.go` | No change | Existing `SendKeysLiteralWithEnter` is reused |
| `internal/config/repo.go` | Modify | Add heartbeat section to repo config schema |
| `cmd/multiclaude/config.go` | Modify | Add `multiclaude config heartbeats` subcommand |

---

## Change 1: Heartbeat Scheduler (`heartbeat.go`)

### New Type: HeartbeatScheduler

```go
package daemon

import (
    "sync"
    "time"
)

// HeartbeatScheduler manages per-agent heartbeat timers and workflow triggers.
type HeartbeatScheduler struct {
    mu          sync.Mutex
    agents      map[string]*agentTimer
    triggers    []*workflowTrigger
    config      HeartbeatConfig
    tmux        TmuxClient
    messages    MessageSender
    logger      Logger
    stopCh      chan struct{}
}

type agentTimer struct {
    Name         string
    Interval     time.Duration
    LastDelivery time.Time
    Offset       time.Duration // Boot-time stagger offset
}

type workflowTrigger struct {
    Name         string
    Target       string
    Interval     time.Duration
    Message      string
    Delivery     string // "tmux" or "message"
    SkipIfActive bool
    LastFired    time.Time
}
```

### Core Loop

Replace `wakeAgents()` tick-based approach with a single goroutine that checks all agents each tick:

```go
func (s *HeartbeatScheduler) Run() {
    // Tick every 30 seconds — check which agents/triggers are due
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case now := <-ticker.C:
            s.mu.Lock()
            s.processHeartbeats(now)
            s.processTriggers(now)
            s.mu.Unlock()
        case <-s.stopCh:
            return
        }
    }
}

func (s *HeartbeatScheduler) processHeartbeats(now time.Time) {
    for _, agent := range s.agents {
        elapsed := now.Sub(agent.LastDelivery)

        // Not due yet (account for boot-time offset on first run)
        if agent.LastDelivery.IsZero() {
            if now.Sub(s.bootTime) < agent.Offset {
                continue
            }
        } else if elapsed < agent.Interval {
            continue
        }

        // Activity detection
        if s.isAgentActive(agent.Name) && !s.hasStaleMessages(agent.Name, agent.Interval) {
            s.logger.Debug("[heartbeat] %s: skipped (active)", agent.Name)
            continue
        }

        // Deliver heartbeat via tmux paste
        prompt := s.getWakePrompt(agent.Name)
        s.tmux.SendKeysLiteralWithEnter(agent.Name, prompt)
        agent.LastDelivery = now
        s.logger.Info("[heartbeat] %s: delivered (idle %s, interval %s)",
            agent.Name, elapsed, agent.Interval)
    }
}

func (s *HeartbeatScheduler) processTriggers(now time.Time) {
    for _, trigger := range s.triggers {
        elapsed := now.Sub(trigger.LastFired)
        if elapsed < trigger.Interval {
            continue
        }

        // Workflow triggers always fire (skip_if_active defaults to false)
        if trigger.SkipIfActive && s.isAgentActive(trigger.Target) {
            continue
        }

        switch trigger.Delivery {
        case "message":
            s.messages.Send(trigger.Target, trigger.Message)
        default:
            s.tmux.SendKeysLiteralWithEnter(trigger.Target, trigger.Message)
        }

        trigger.LastFired = now
        s.logger.Info("[trigger] %s → %s: delivered (interval %s)",
            trigger.Name, trigger.Target, trigger.Interval)
    }
}
```

### Activity Detection

```go
func (s *HeartbeatScheduler) isAgentActive(name string) bool {
    // Check tmux pane activity timestamp
    // tmux display -p -t <pane> '#{pane_last_activity}' returns Unix timestamp
    lastActivity, err := s.tmux.GetPaneActivity(name)
    if err != nil {
        return false // Can't determine — assume idle, deliver heartbeat
    }

    idleDuration := time.Now().UTC().Sub(lastActivity)
    return idleDuration < s.config.ActivityThreshold
}

func (s *HeartbeatScheduler) hasStaleMessages(name string, interval time.Duration) bool {
    // Check if agent has unacknowledged messages older than 2x interval
    pending := s.messages.PendingFor(name)
    staleThreshold := 2 * interval
    for _, msg := range pending {
        if time.Since(msg.Timestamp) > staleThreshold {
            return true // Stale messages — deliver heartbeat even if active
        }
    }
    return false
}
```

---

## Change 2: Configuration Schema

### Config File Addition

In the repo config file (`~/.multiclaude/repos/<repo>/config.yaml`), add:

```yaml
heartbeats:
  enabled: true
  default_interval: 5m
  activity_threshold: 60s
  agents:
    merge-queue: 5m
    pr-shepherd: 5m
    envoy: 10m
    project-watchdog: 15m
    arch-watchdog: 20m
    retrospector: 15m
  triggers:
    - name: SYNC_OPERATIONAL_DATA
      target: project-watchdog
      interval: 3h
      message: "SYNC_OPERATIONAL_DATA"
      delivery: message
      skip_if_active: false
```

### Go Types

```go
type HeartbeatConfig struct {
    Enabled           bool                    `yaml:"enabled"`
    DefaultInterval   time.Duration           `yaml:"default_interval"`
    ActivityThreshold time.Duration           `yaml:"activity_threshold"`
    Agents            map[string]time.Duration `yaml:"agents"`
    Triggers          []TriggerConfig         `yaml:"triggers"`
}

type TriggerConfig struct {
    Name         string        `yaml:"name"`
    Target       string        `yaml:"target"`
    Interval     time.Duration `yaml:"interval"`
    Message      string        `yaml:"message"`
    Delivery     string        `yaml:"delivery"` // "tmux" or "message"
    SkipIfActive bool          `yaml:"skip_if_active"`
}

// Defaults applied when config is missing or partial
func DefaultHeartbeatConfig() HeartbeatConfig {
    return HeartbeatConfig{
        Enabled:           true,
        DefaultInterval:   5 * time.Minute,
        ActivityThreshold: 60 * time.Second,
        Agents: map[string]time.Duration{
            "merge-queue":      5 * time.Minute,
            "pr-shepherd":      5 * time.Minute,
            "envoy":            10 * time.Minute,
            "project-watchdog": 15 * time.Minute,
            "arch-watchdog":    20 * time.Minute,
            "retrospector":     15 * time.Minute,
        },
        Triggers: []TriggerConfig{
            {
                Name:         "SYNC_OPERATIONAL_DATA",
                Target:       "project-watchdog",
                Interval:     3 * time.Hour,
                Message:      "SYNC_OPERATIONAL_DATA",
                Delivery:     "message",
                SkipIfActive: false,
            },
        },
    }
}
```

---

## Change 3: Daemon Integration

### Replace `wakeAgents()` in `daemon.go`

```go
// Before (current):
func (d *Daemon) wakeAgents() {
    // Fixed 2-min loop nudging all agents
    for _, agent := range d.agents {
        if agent.Type == state.AgentTypeWorkspace { continue }
        if time.Since(agent.LastNudge) < 2*time.Minute { continue }
        d.tmux.SendKeysLiteralWithEnter(agent.Pane, agent.WakePrompt)
        agent.LastNudge = time.Now()
    }
}

// After:
func (d *Daemon) startHeartbeats() {
    config := d.loadHeartbeatConfig()
    d.heartbeats = NewHeartbeatScheduler(config, d.tmux, d.messages, d.logger)

    // Register all active agents
    for _, agent := range d.agents {
        if agent.Type == state.AgentTypeWorkspace { continue }
        interval := config.IntervalFor(agent.Name)
        d.heartbeats.RegisterAgent(agent.Name, interval)
    }

    go d.heartbeats.Run()
}
```

### Agent Spawn/Despawn Hooks

When agents are spawned or removed, update the scheduler:

```go
func (d *Daemon) onAgentSpawned(agent *AgentState) {
    interval := d.heartbeats.config.IntervalFor(agent.Name)
    d.heartbeats.RegisterAgent(agent.Name, interval)
}

func (d *Daemon) onAgentRemoved(name string) {
    d.heartbeats.UnregisterAgent(name)
}
```

---

## Change 4: CLI Subcommand

### `multiclaude config heartbeats`

```
Usage:
  multiclaude config heartbeats [flags]

Flags:
  --default-interval duration    Default heartbeat interval (e.g., 5m)
  --activity-threshold duration  Idle threshold for activity detection (e.g., 60s)
  --agent string                 Agent name to configure
  --interval duration            Interval for --agent
  --add-trigger string           Add a workflow trigger by name
  --target string                Target agent for --add-trigger
  --trigger-interval duration    Interval for --add-trigger
  --remove-trigger string        Remove a trigger by name
  --enabled bool                 Enable/disable heartbeats

Examples:
  multiclaude config heartbeats                          # Show current config
  multiclaude config heartbeats --agent merge-queue --interval 3m
  multiclaude config heartbeats --add-trigger QUOTA_CHECK --target supervisor --trigger-interval 5m
  multiclaude config heartbeats --enabled false
```

---

## Change 5: New tmux Helper

### `GetPaneActivity()` in `pkg/tmux/client.go`

```go
// GetPaneActivity returns the last time the pane received output.
func (c *Client) GetPaneActivity(paneTarget string) (time.Time, error) {
    output, err := c.run("display-message", "-p", "-t", paneTarget,
        "#{pane_last_activity}")
    if err != nil {
        return time.Time{}, fmt.Errorf("get pane activity for %s: %w", paneTarget, err)
    }

    epoch, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
    if err != nil {
        return time.Time{}, fmt.Errorf("parse pane activity timestamp: %w", err)
    }

    return time.Unix(epoch, 0).UTC(), nil
}
```

---

## Migration Checklist (for implementer)

When implementing this spec:

1. [ ] Create `internal/daemon/heartbeat.go` with HeartbeatScheduler
2. [ ] Add HeartbeatConfig to repo config schema
3. [ ] Add `GetPaneActivity()` to tmux client
4. [ ] Replace `wakeAgents()` call in daemon main loop with `startHeartbeats()`
5. [ ] Add agent spawn/despawn hooks
6. [ ] Add `multiclaude config heartbeats` CLI subcommand
7. [ ] Add default SYNC_OPERATIONAL_DATA trigger to default config
8. [ ] Write tests for:
   - HeartbeatScheduler interval logic
   - Activity detection thresholds
   - Trigger firing logic
   - Config parsing and defaults
   - Staggering offset calculation
9. [ ] Update daemon logs to include heartbeat decisions
10. [ ] Remove SYNC_OPERATIONAL_DATA CronCreate from supervisor MEMORY.md startup checklist

## Backward Compatibility

- The existing 2-minute wake loop behavior is preserved as the default if no heartbeat config exists
- Agents that aren't listed in the config get `default_interval` (5m)
- Existing agent definitions don't need changes — they respond to any wake prompt
- The workspace window continues to be excluded from all heartbeats/triggers
