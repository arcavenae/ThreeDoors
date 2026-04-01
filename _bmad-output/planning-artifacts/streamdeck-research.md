# Elgato Stream Deck Plugin/App Research for ThreeDoors (macOS)

**Date:** 2026-03-31
**Status:** Research Complete
**Provenance:** L3 (AI-autonomous)

---

## Executive Summary

Building a Stream Deck plugin for ThreeDoors is **feasible and well-supported** by the Elgato SDK ecosystem. The recommended approach is a **Node.js thin-client plugin** that communicates with a **Go daemon/server component** inside ThreeDoors via a local Unix socket or HTTP API. This avoids the complexity of writing the plugin entirely in Go (which is possible but uses dormant community SDKs) while keeping the core logic in Go where ThreeDoors's `TaskProvider` interface lives.

**Key decisions needed:**
1. Separate repo (recommended) vs monorepo for the plugin
2. Whether ThreeDoors needs a daemon/server mode (yes — the TUI is not externally addressable today)
3. Node.js plugin + Go daemon (recommended) vs pure Go plugin (riskier, less maintained SDKs)

---

## 1. Stream Deck SDK & Plugin Architecture

### How Plugins Work

Stream Deck plugins are architecturally similar to web apps with a frontend/backend split:

- **Backend (plugin):** Runs on Node.js (v20 or v24, as of SD 7.3). Handles core logic, state management, and communication with external services.
- **Frontend (property inspector):** Runs in Chromium (v130). Provides configuration UI for button settings.
- **Communication:** Bidirectional JSON over WebSocket between the plugin and the Stream Deck application.

The Stream Deck app manages all hardware communication — plugins never talk to hardware directly.

### Plugin Lifecycle

1. Stream Deck launches the plugin process with CLI args: `-port`, `-pluginUUID`, `-registerEvent`, `-info`
2. Plugin connects to the WebSocket on the specified port
3. Plugin sends a registration message with its UUID
4. Bidirectional event flow begins (button presses → plugin, visual updates → Stream Deck)
5. Stream Deck manages crash recovery and plugin restart

### Supported Languages

| Approach | Status | Recommendation |
|----------|--------|----------------|
| **Node.js/TypeScript** (official SDK) | Actively maintained, official CLI scaffolding, v2+ SDK | **Recommended** |
| **Python** (community SDK) | Released March 2025, active | Viable alternative |
| **Go** (community SDKs) | Multiple libraries, all dormant (2019-2022) | Use for daemon only |
| **Native binary** (C++/C#/Go) | Supported but "advanced technique, not recommended" per Elgato | Possible but unsupported path |

### Native Binary Support

The manifest supports `CodePath` pointing to a native executable (not just `.js`). Platform-specific paths via `CodePathMac` and `CodePathWin` are available. However:
- Must be a **single-file executable** (not `.app` bundles on macOS)
- Elgato explicitly recommends Node.js over native binaries
- DRM and Marketplace features are designed around the Node.js SDK
- Native plugins must implement WebSocket communication from scratch

### Key WebSocket Events

**Inbound (from Stream Deck):**
- `keyDown` / `keyUp` — button press/release
- `dialRotate` / `dialDown` / `dialUp` — encoder interaction
- `willAppear` / `willDisappear` — action visibility
- `didReceiveSettings` / `didReceiveGlobalSettings` — configuration changes
- `systemDidWakeUp` — system resume (important for reconnecting to ThreeDoors daemon)

**Outbound (to Stream Deck):**
- `setImage` — update button icon (base64 PNG or file path)
- `setTitle` — update button text
- `setState` — toggle button state (0/1)
- `showOk` / `showAlert` — visual feedback
- `setFeedback` — update encoder LCD content
- `openUrl` — launch browser

### Development Tools

- **Stream Deck CLI:** `npm install -g @elgato/cli` → `streamdeck create` scaffolding wizard
- **Hot reload:** `npm run watch` rebuilds and restarts plugin on file changes
- **Debugging:** Node.js debugger support (VS Code, Chrome DevTools)
- **Packaging:** `streamdeck pack` → `.streamDeckPlugin` installer file

---

## 2. Integration with ThreeDoors

### Current ThreeDoors Architecture

ThreeDoors is a Go TUI (Bubbletea) with:
- `TaskProvider` interface (`internal/core/provider.go`) — `LoadTasks()`, `SaveTask()`, `SaveTasks()`, `DeleteTask()`
- YAML task file storage
- JSONL session logging
- No network-accessible API or daemon mode

### The Core Challenge

**ThreeDoors has no external API.** The TUI runs as a foreground terminal process. A Stream Deck plugin needs to communicate with ThreeDoors to:
1. Read the current three doors (tasks)
2. Trigger door selection (choose a task)
3. Complete/skip/snooze tasks
4. Get status updates for button display

### Recommended Integration Architecture

```
┌─────────────┐     WebSocket      ┌──────────────┐     Unix Socket    ┌─────────────────┐
│ Stream Deck  │◄──────────────────►│  SD Plugin   │◄──────────────────►│  ThreeDoors     │
│ Application  │     (JSON)         │  (Node.js)   │     (JSON-RPC)     │  Daemon (Go)    │
└─────────────┘                     └──────────────┘                    └────────┬────────┘
                                                                                 │
                                                                        ┌────────▼────────┐
                                                                        │  TaskProvider    │
                                                                        │  (YAML files)    │
                                                                        └─────────────────┘
```

**Option A: Go Daemon + Node.js Plugin (Recommended)**

ThreeDoors gains a lightweight daemon/server mode (`threedoors serve`) that exposes the TaskProvider over a local Unix socket or localhost HTTP. The Node.js Stream Deck plugin connects to this daemon.

- **Pros:** Official SDK support, Marketplace compatibility, clean separation, daemon benefits other integrations (MCP, Shortcuts, etc.)
- **Cons:** Two processes to manage, Node.js dependency for plugin

**Option B: Pure Go Plugin (Native Binary)**

Compile a Go binary that implements the Stream Deck WebSocket protocol directly and accesses YAML task files.

- **Pros:** Single language, direct file access, no daemon needed
- **Cons:** Dormant Go SDKs (last updated 2019-2022), no official support, can't use DRM/Marketplace features easily, concurrent file access with TUI is dangerous

**Option C: CLI Bridge**

Plugin shells out to a `threedoors` CLI command for each operation.

- **Pros:** Simplest to implement, no daemon
- **Cons:** High latency per operation, process spawn overhead, no real-time updates, no push notifications

### Recommended: Option A with a Daemon

The daemon approach has the highest value because it unlocks not just Stream Deck but all future integrations (MCP server, Apple Shortcuts, iOS companion app, widgets). It's a one-time investment that pays dividends.

### Feature Mapping to Stream Deck Buttons

| Stream Deck Button | ThreeDoors Action | Implementation |
|---|---|---|
| **Door 1 / 2 / 3** | Select door (start task) | `keyDown` → daemon RPC `selectDoor(1)` |
| **Complete** | Mark current task done | `keyDown` → daemon RPC `completeTask()` |
| **Skip** | Skip current task | `keyDown` → daemon RPC `skipTask()` |
| **Snooze** | Snooze current task | `keyDown` → daemon RPC `snoozeTask()` |
| **Refresh** | Reshuffle doors | `keyDown` → daemon RPC `refresh()` |
| **Stats** | Show session stats | Display on encoder LCD / button title |

Button icons would show task titles (truncated) on each door button, updating via `setTitle` when doors change.

### Stream Deck+ (Encoder) Support

The Stream Deck+ has rotary dials with LCD strips — these could show:
- Task priority/effort as a dial position
- Session timer on the LCD
- Rotate to browse tasks, press to select

---

## 3. Repository Structure

### Option A: Separate Repository (Recommended)

```
github.com/ArcavenAE/threedoors-streamdeck/
├── plugin/                    # Node.js Stream Deck plugin
│   ├── src/
│   │   ├── plugin.ts         # Entry point
│   │   ├── actions/          # Door actions, complete, skip, etc.
│   │   └── threedoors-client.ts  # Daemon communication
│   ├── .sdPlugin/
│   │   ├── manifest.json
│   │   ├── imgs/
│   │   └── ui/               # Property inspector HTML
│   ├── package.json
│   └── tsconfig.json
├── docs/
├── LICENSE
└── README.md
```

**Pros:**
- Independent release cycle (Stream Deck plugin updates don't require ThreeDoors releases)
- Clean dependency separation (Node.js toolchain stays out of Go project)
- Marketplace submission is simpler with a focused repo
- Can be developed/tested independently
- Other contributors can work on the plugin without understanding the full ThreeDoors codebase

**Cons:**
- Two repos to maintain
- Integration testing requires coordination
- Daemon API changes need synchronized releases

### Option B: Monorepo (subdirectory)

```
github.com/ArcavenAE/ThreeDoors/
├── cmd/threedoors/           # Existing TUI entry point
├── cmd/threedoors-daemon/    # New daemon entry point
├── internal/                 # Existing Go packages
├── streamdeck-plugin/        # New: Stream Deck plugin
│   ├── src/
│   ├── .sdPlugin/
│   └── package.json
└── ...
```

**Pros:**
- Single source of truth, atomic changes across daemon + plugin
- Easier to keep daemon API and plugin in sync

**Cons:**
- Mixed Go + Node.js toolchain in one repo
- CI complexity increases
- Marketplace submission from a subdirectory is awkward
- Plugin contributors need full repo context

### Recommendation: Separate Repo

The daemon component (`threedoors serve`) lives in the main ThreeDoors repo as a new `cmd/` entry point. The Stream Deck plugin is a separate repo that depends on the daemon's API contract (documented via OpenAPI or similar).

### Build Pipeline

**Plugin build (CI/CD):**
```bash
npm ci
npm run build           # TypeScript → JS
streamdeck validate     # Validate manifest and structure
streamdeck pack         # Create .streamDeckPlugin
```

**Local dev install:**
```bash
streamdeck link         # Symlink .sdPlugin to Stream Deck plugins dir
npm run watch           # Hot reload on changes
```

### Elgato Marketplace Publishing

1. Register as a Maker at [Maker Console](https://docs.elgato.com/makers/general/become-a-maker/)
2. Sign the Maker Agreement
3. Package plugin via `streamdeck pack`
4. Upload `.streamDeckPlugin` to Maker Console
5. Submit for review (takes ~weeks)
6. Published to Marketplace (searchable in Stream Deck app)

**Requirements:**
- Plugin must not compromise device safety or integrity
- Must include clear description and gallery images
- No external paywalls (if paid, must use Elgato's system)
- Category and action names must be descriptive (≤30 chars)
- UUID in reverse-DNS format (e.g., `com.arcavenae.threedoors`)

**DRM (optional):**
- Available for Marketplace plugins using `@elgato/streamdeck` v2+
- Provides file encryption and integrity checking
- Requires SDKVersion 3 and MinimumVersion 6.9+

---

## 4. macOS Considerations

### Code Signing & Notarization

**For the Node.js plugin:** Not required. Stream Deck plugins using the official Node.js runtime don't need separate code signing — the Stream Deck application hosts the plugin process.

**For the Go daemon (`threedoors serve`):** If distributed via Homebrew (which ThreeDoors already targets):
- Homebrew bottles are not individually signed
- If distributed as a standalone binary: requires Apple Developer ID certificate, `codesign`, and `notarytool` submission
- ThreeDoors already has [code signing research](../_bmad-output/planning-artifacts/code-signing-findings-research.md) and [Apple signing investigation](../_bmad-output/planning-artifacts/apple-signing-investigation-analysis.md) — the daemon would follow the same pipeline

### Sandbox Restrictions

- Stream Deck plugins run outside the App Store sandbox
- Unix socket communication (`/tmp/threedoors.sock` or `~/.threedoors/daemon.sock`) works without restrictions
- localhost HTTP (e.g., `http://127.0.0.1:7333`) also works without special permissions
- No entitlements needed for local IPC

### Homebrew Compatibility

The daemon would be part of the existing `threedoors` Homebrew formula (or a separate `threedoors-daemon` service):
```ruby
# In Homebrew formula
service do
  run [opt_bin/"threedoors", "serve"]
  keep_alive true
  log_path var/"log/threedoors-daemon.log"
end
```

Users would: `brew services start threedoors` to run the daemon.

### macOS-Specific Manifest

```json
{
  "OS": [
    { "Platform": "mac", "MinimumVersion": "13" }
  ],
  "CodePathMac": "bin/plugin.js"
}
```

---

## 5. Prior Art & Open Source

### Go Stream Deck Libraries

| Library | Last Active | Stars | Notes |
|---------|------------|-------|-------|
| [dh1tw/streamdeck](https://github.com/dh1tw/streamdeck) | Dec 2021 | 86 | Direct HID hardware access (bypasses SD app), MIT |
| [tystuyfzand/streamdeck-sdk-go](https://github.com/tystuyfzand/streamdeck-sdk-go) | Mar 2019 | 36 | WebSocket plugin SDK, BSD-3 |
| [SkYNewZ/streamdeck-sdk](https://github.com/SkYNewZ/streamdeck-sdk) | Feb 2022 | 11 | Plugin SDK, Apache-2.0 |
| [samwho/streamdeck-plugin-skeleton](https://github.com/samwho/streamdeck-plugin-skeleton) | ~2019 | — | Skeleton/template project |
| [Luzifer/streamdeck](https://github.com/Luzifer/streamdeck) | — | — | Linux-focused management tool |
| [rafaelmartins.com/p/streamdeck](https://pkg.go.dev/rafaelmartins.com/p/streamdeck) | — | — | Pure Go, multi-model support |

**Assessment:** All Go SDKs are dormant. `tystuyfzand/streamdeck-sdk-go` is the closest to what we'd need (WebSocket plugin protocol) but hasn't been updated since 2019 and targets an older SDK version. Building on these would require significant maintenance burden and porting to SDK v3.

### CLI/TUI Controller Plugins

| Plugin | Description |
|--------|-------------|
| [streamdeck-commandline](https://github.com/mikepowell/streamdeck-commandline) | Execute arbitrary CLI commands from buttons (Windows-only) |
| [StreamDeckWS](https://github.com/ybizeul/StreamDeckWS) | WebSocket proxy for Node-RED integration |
| [WebSocket Proxy](https://marketplace.elgato.com/product/websocket-proxy-5ed6a37a-d6e9-4c95-aa95-42ded37ecff1) | Generic WebSocket bridge (Marketplace) |
| [OpenDeck](https://github.com/nekename/OpenDeck) | Linux SD software with Elgato plugin compatibility |

**Insight:** The WebSocket Proxy plugin on the Marketplace demonstrates that bridging Stream Deck to external services via WebSocket is a proven pattern. ThreeDoors could even work with this existing plugin if the daemon exposes a WebSocket API.

### Official Sample Plugins

[elgatosf/streamdeck-plugin-samples](https://github.com/elgatosf/streamdeck-plugin-samples) — Counter, CPU monitor, and other examples. Good templates for understanding the SDK patterns.

---

## 6. Recommended Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Stream Deck Plugin** | TypeScript + `@elgato/streamdeck` v2 | Official SDK, Marketplace support, hot reload |
| **Plugin ↔ Daemon IPC** | Unix socket with JSON-RPC 2.0 | Low latency, no port conflicts, macOS-native |
| **Daemon** | Go (`threedoors serve` subcommand) | Reuses TaskProvider, same language as ThreeDoors |
| **Daemon ↔ TaskProvider** | Direct Go interface call | Already implemented |
| **Build** | npm + `@elgato/cli` for plugin, `just` for daemon | Standard tooling for each ecosystem |
| **CI** | GitHub Actions | Separate workflows for plugin and daemon |

### Manifest Skeleton

```json
{
  "$schema": "https://schemas.elgato.com/streamdeck/plugins/manifest.json",
  "UUID": "com.arcavenae.threedoors",
  "Name": "ThreeDoors",
  "Version": "0.1.0.0",
  "Author": "ArcavenAE",
  "Description": "Control ThreeDoors task management from your Stream Deck",
  "Icon": "imgs/plugin-icon",
  "SDKVersion": 3,
  "Software": { "MinimumVersion": "6.9" },
  "OS": [{ "Platform": "mac", "MinimumVersion": "13" }],
  "Nodejs": { "Version": "20", "Debug": "enabled" },
  "CodePath": "bin/plugin.js",
  "Actions": [
    {
      "UUID": "com.arcavenae.threedoors.door",
      "Name": "Door",
      "Icon": "imgs/actions/door",
      "Controllers": ["Keypad"],
      "States": [{ "Image": "imgs/actions/door-closed" }]
    },
    {
      "UUID": "com.arcavenae.threedoors.complete",
      "Name": "Complete Task",
      "Icon": "imgs/actions/complete",
      "Controllers": ["Keypad"],
      "States": [{ "Image": "imgs/actions/complete" }]
    }
  ]
}
```

---

## 7. Next Steps

### Phase 1: Foundation (prerequisite — in ThreeDoors repo)
1. **Design daemon API** — Define JSON-RPC methods: `getDoors()`, `selectDoor(index)`, `completeTask()`, `skipTask()`, `snoozeTask()`, `getStats()`
2. **Implement `threedoors serve`** — Lightweight daemon exposing TaskProvider over Unix socket
3. **Add daemon to Homebrew service** — `brew services start threedoors`

### Phase 2: Plugin MVP (separate repo)
4. **Scaffold plugin** — `streamdeck create` with TypeScript template
5. **Implement door actions** — Three door buttons + complete/skip
6. **Dynamic button titles** — Show task names on door buttons via `setTitle`
7. **Connection management** — Auto-reconnect to daemon, show alert on disconnect

### Phase 3: Polish & Publish
8. **Design button icons** — Door-themed icons matching ThreeDoors aesthetic
9. **Property inspector** — Settings UI for daemon socket path, refresh interval
10. **Stream Deck+ support** — Encoder actions for browsing tasks
11. **Marketplace submission** — Gallery images, descriptions, review process

### Phase 4: Future Enhancements
12. **Multi-action profiles** — Pre-built Stream Deck profiles with optimal button layouts
13. **Statistics display** — Session stats on encoder LCD
14. **Theme sync** — Button icons match active ThreeDoors door theme

---

## References

- [Stream Deck SDK — Plugin Environment](https://docs.elgato.com/streamdeck/sdk/introduction/plugin-environment/)
- [Stream Deck SDK — Getting Started](https://docs.elgato.com/streamdeck/sdk/introduction/getting-started/)
- [Stream Deck SDK — Manifest Reference](https://docs.elgato.com/streamdeck/sdk/references/manifest/)
- [Stream Deck SDK — WebSocket Plugin API](https://docs.elgato.com/streamdeck/sdk/references/websocket/plugin/)
- [Stream Deck SDK — Distribution](https://docs.elgato.com/streamdeck/sdk/introduction/distribution/)
- [Elgato Marketplace — Submission Guidelines](https://docs.elgato.com/guidelines/submissions/)
- [Official Plugin Samples](https://github.com/elgatosf/streamdeck-plugin-samples)
- [Stream Deck CLI](https://github.com/elgatosf/cli)
- [Go SDK — tystuyfzand/streamdeck-sdk-go](https://github.com/tystuyfzand/streamdeck-sdk-go)
- [Go SDK — SkYNewZ/streamdeck-sdk](https://github.com/SkYNewZ/streamdeck-sdk)
- [Go Hardware — dh1tw/streamdeck](https://github.com/dh1tw/streamdeck)
- [Python SDK — streamdeck-plugin-sdk](https://pypi.org/project/streamdeck-plugin-sdk/)
- [WebSocket Proxy Plugin](https://marketplace.elgato.com/product/websocket-proxy-5ed6a37a-d6e9-4c95-aa95-42ded37ecff1)
