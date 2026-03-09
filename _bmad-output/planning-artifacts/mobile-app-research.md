# ThreeDoors Mobile App Research Findings

## Party Mode Discussion Summary

This document captures the findings from a multi-agent analysis of how to add an iPhone mobile app version of ThreeDoors. The discussion covered 8 key areas with recommendations accepted across all topics.

---

## 1. Native Swift/SwiftUI vs React Native vs Flutter vs PWA

### Options Evaluated

| Approach | Performance | iOS Integration | Dev Speed | Maintenance |
|----------|------------|-----------------|-----------|-------------|
| **SwiftUI (Native)** | Best (60 FPS, Metal-backed) | Full (iCloud, Apple Notes, Shortcuts) | Moderate | iOS-only codebase |
| **React Native** | Good (JS bridge ~16-32ms latency) | Good (via native modules) | Fast (hot reload) | Cross-platform possible |
| **Flutter** | Very Good (Skia rendering, 60 FPS) | Good (via platform channels) | Fast (hot reload) | Cross-platform possible |
| **PWA** | Adequate | Limited (no background sync, limited offline) | Fastest | Zero App Store friction |

### Recommendation: **Native SwiftUI**

**Rationale:**
- ThreeDoors is an Apple-ecosystem product (macOS TUI + Apple Notes integration)
- SwiftUI provides seamless iCloud/CloudKit integration for data sync
- Direct Apple Notes API access via Swift (vs osascript workarounds)
- No cross-platform need identified (no Android planned)
- SwiftUI adoption is 60-70% of iOS devs for new projects as of 2025
- Best performance for the gesture-heavy "Three Doors" UX concept
- Smallest dependency footprint for a single-platform app
- Apple requires Xcode 16+ / iOS 18 SDK for 2026 submissions — SwiftUI is Apple's preferred path

**Why not others:**
- React Native/Flutter add cross-platform complexity for a single-platform app
- PWA limitations on iOS (no background sync, limited offline, storage constraints) conflict with ThreeDoors' offline-first philosophy
- Cross-platform frameworks would complicate Apple Notes integration

---

## 2. Sharing Go Backend Logic

### Options Evaluated

| Approach | Code Sharing | Complexity | Performance |
|----------|-------------|------------|-------------|
| **gomobile bind** | Go → .xcframework | Medium | Good (compiled) |
| **Shared .dylib/.a** | Go → C shared library | High | Good |
| **HTTP API server** | Go runs as local server | Low-Medium | Network overhead |
| **Rewrite in Swift** | No sharing | Low complexity, high effort | Best |
| **Hybrid: Swift + shared types** | Protocol-level sharing | Low | Best |

### Recommendation: **Protocol-level sharing (rewrite core in Swift, share interfaces)**

**Rationale:**
- The Go codebase is ~2,500 lines of domain logic — small enough to rewrite
- gomobile has limitations: only supports a subset of Go types, no generics support, adds build complexity
- The `TaskProvider` interface pattern is the real reusable asset — port the interface design, not the implementation
- Swift's native CloudKit/iCloud integration is vastly simpler than making Go talk to Apple services through gomobile
- The domain model (Task, TaskStatus, DoorSelection) is simple enough to duplicate without risk
- Sync logic patterns (last-write-wins, write queue) are algorithmic knowledge, not code-dependent

**Sharing strategy:**
- Port `TaskProvider` protocol to Swift protocol
- Port `Task` model to Swift struct (with Codable)
- Port `SyncEngine` conflict resolution algorithm to Swift
- Share task file format (YAML) for interop between TUI and mobile
- Share Apple Notes parsing logic (checkbox format) conceptually

---

## 3. Data Sync Between Desktop TUI and Mobile App

### Options Evaluated

| Approach | Complexity | Offline Support | Real-time |
|----------|-----------|----------------|-----------|
| **iCloud/CloudKit** | Low-Medium | Yes (built-in) | Near real-time |
| **Shared YAML files via iCloud Drive** | Low | Yes | File-level sync delays |
| **Local HTTP server** | Medium | No (requires both running) | Yes |
| **Apple Notes as shared backend** | Already exists | Yes (via iCloud) | Via Notes sync |

### Recommendation: **Apple Notes as primary shared backend + iCloud Drive for config/metrics**

**Rationale:**
- Apple Notes sync is already implemented in ThreeDoors (Epic 2)
- Both TUI (osascript) and mobile app (Swift EventKit/Apple Notes API) can read/write the same Apple Notes
- This gives "free" sync — edit a task on iPhone in Apple Notes, it syncs to Mac, TUI picks it up
- The existing `SyncEngine` with last-write-wins conflict resolution handles concurrent edits
- iCloud Drive can sync config files and metrics between desktop and mobile
- No custom backend infrastructure needed
- CKSyncEngine (Apple's 2023+ recommendation) available if CloudKit sync needed later

**Sync architecture:**
```
iPhone App ←→ Apple Notes (iCloud) ←→ macOS TUI
                                        ↕
iPhone App ←→ iCloud Drive ←→ macOS TUI
             (config, metrics)
```

---

## 4. Three Doors UX Translation to Touch/Mobile

### Recommended Mobile UX

| Desktop (TUI) | Mobile (Touch) |
|---------------|----------------|
| Three doors side-by-side | Three cards stacked vertically or swipeable carousel |
| WASD/Arrow key navigation | Tap to select, swipe to navigate |
| Press key for status change | Tap action buttons or swipe gestures |
| `/` for search | Pull-down search bar |
| `:command` palette | Bottom sheet with actions |
| Mood capture via `M` key | Floating action button or shake gesture |
| Door refresh via `S`/Down | Pull-to-refresh gesture |

### Recommendation: **Swipeable card carousel with tap-to-open**

**Design principles:**
- **Three cards as swipeable carousel**: Each "door" is a card. Swipe left/right to browse. This preserves the "only see a few at a time" philosophy
- **Tap to open**: Tapping a card opens the detail view (replaces Enter/door opening)
- **Swipe down to refresh**: Pull-to-refresh generates new three doors (replaces `S` key)
- **Haptic feedback**: Light haptic on card selection, medium on status change, success haptic on completion
- **Bottom action sheet**: Status actions (Complete, Blocked, In Progress) via bottom sheet — familiar iOS pattern
- **Swipe gestures on cards**: Swipe right = complete, swipe left = defer/skip (Tinder-like, reduces taps)
- **"Progress over perfection" toast**: Appears after completion, same as TUI
- **Minimal chrome**: Full-screen cards, no tab bar, focus on the three doors concept

---

## 5. Apple App Store Requirements and Signing

### Requirements Checklist

| Requirement | Detail |
|------------|--------|
| **Developer Account** | Apple Developer Program ($99/year individual) |
| **Xcode Version** | Xcode 16+ required for 2026 submissions |
| **Target SDK** | iOS 18 SDK minimum |
| **Minimum iOS Version** | Recommend iOS 17+ (covers 90%+ of active devices) |
| **Code Signing** | Development + Distribution certificates via Xcode |
| **App Review** | Must pass Apple's App Review Guidelines |
| **Privacy** | App Privacy labels required; declare data collection |
| **App Size** | Keep under 200MB for cellular download |

### Recommendation: **Target iOS 17+, use Xcode Automatic Signing**

**Notes:**
- ThreeDoors collects no personal data beyond what's in Apple Notes (user's own data)
- Privacy labels would be minimal: "Data Not Collected" or "Data Not Linked to You"
- No in-app purchases planned for MVP — simplifies review
- TestFlight for beta testing before App Store submission
- Consider keeping it as a personal/side-loaded app initially to avoid App Store costs

---

## 6. Minimum Viable Mobile Experience

### MVP Features (Must Have)

1. **Three Doors display** — See three task cards, swipe/tap to browse
2. **Open a door** — Tap to see task detail
3. **Mark complete** — Complete a task with one tap or swipe
4. **Refresh doors** — Pull-to-refresh for new set of three
5. **Apple Notes sync** — Read tasks from Apple Notes (same note the TUI uses)
6. **Status changes** — Mark blocked, in-progress, complete
7. **Session metrics** — Track basic session data locally

### Phase 2 Features (Can Wait)

- Quick add mode (add tasks from mobile)
- Mood tracking
- Search/filter
- Extended task capture with "why"
- Door selection analytics
- Values/goals display
- Write-back to Apple Notes from mobile
- Widget for iOS home screen
- Shortcuts/Siri integration
- Offline queue for Apple Notes writes

### Phase 3 Features (Future)

- Apple Watch complication
- iPad layout
- Notifications/reminders
- Shared task lists
- AI-powered task suggestions

### Recommendation: **Ship MVP with read + complete + refresh**

The core value proposition is "see three doors, pick one, take action." The mobile MVP must deliver exactly this experience and nothing more. Apple Notes read sync ensures tasks are the same across devices.

---

## 7. How This Fits with Existing Apple Notes Integration

### Current State

- ThreeDoors TUI reads/writes Apple Notes via `osascript` (AppleScript automation)
- Tasks stored as checkbox lines: `- [ ] task text` / `- [x] task text`
- `SyncEngine` handles bidirectional sync with last-write-wins conflict resolution
- `WriteQueue` handles retry for failed writes
- `FallbackProvider` provides graceful degradation

### Mobile Integration Points

| Component | TUI (macOS) | Mobile (iOS) |
|-----------|-------------|--------------|
| Apple Notes access | osascript (AppleScript) | Swift Apple Notes API or direct file format |
| Task parsing | Go regex on checkbox lines | Swift regex on same format |
| Sync | SyncEngine (Go) | Port SyncEngine logic to Swift |
| Write queue | WriteQueue (YAML file) | Swift equivalent with iCloud Drive shared queue |
| Conflict resolution | Last-write-wins by timestamp | Same algorithm in Swift |

### Recommendation: **Use Apple Notes as the single source of truth**

- Both apps read/write the same Apple Notes document
- iCloud handles the sync between devices automatically
- No custom sync server needed
- The mobile app can use iOS's native Apple Notes integration (potentially via CloudKit or documented APIs) rather than osascript
- Maintain the same checkbox format (`- [ ]` / `- [x]`) for interoperability
- Share the same note title configuration

---

## 8. Code Sharing Strategy Between Go TUI and Mobile App

### Shared Assets (Non-Code)

| Asset | Sharing Method |
|-------|---------------|
| Task file format (YAML schema) | Documented specification |
| Apple Notes checkbox format | Documented specification |
| Sync algorithm (last-write-wins) | Documented algorithm, implemented independently |
| Session metrics format (JSONL) | Shared schema definition |
| Config file format | Shared YAML schema |
| Door selection algorithm | Documented algorithm |

### Shared via iCloud Drive

| File | Purpose |
|------|---------|
| `~/.threedoors/config.yaml` | Provider configuration |
| `~/.threedoors/sessions.jsonl` | Session metrics (append from both) |
| `~/.threedoors/pending_writes.yaml` | Write queue (shared retry queue) |

### Recommendation: **Share specifications and data formats, not code**

**Strategy:**
1. **Document all interfaces and algorithms** as specifications (not Go code)
2. **Implement natively in Swift** following the same patterns
3. **Share data files via iCloud Drive** for config and metrics interop
4. **Use Apple Notes as shared task store** — both apps read/write same note
5. **Maintain protocol compatibility** — same task ID generation (deterministic UUID from `noteTitle:lineIndex`)
6. **Test interop** — verify both apps can read each other's written data

This avoids gomobile complexity while ensuring the apps work together seamlessly.

---

## Summary of Recommendations

| Area | Recommendation |
|------|---------------|
| Framework | Native SwiftUI |
| Backend sharing | Protocol-level sharing (rewrite in Swift) |
| Data sync | Apple Notes as shared backend + iCloud Drive |
| Mobile UX | Swipeable card carousel with tap-to-open |
| App Store | iOS 17+, automatic signing, minimal privacy footprint |
| MVP scope | Read tasks + Three Doors display + Complete + Refresh |
| Apple Notes fit | Single source of truth, same checkbox format |
| Code sharing | Share specs and data formats, implement natively |

---

## Sources

- [Calling Go code from Swift with Gomobile](https://medium.com/@matryer/tutorial-calling-go-code-from-swift-on-ios-and-vice-versa-with-gomobile-7925620c17a4)
- [Flutter vs React Native vs SwiftUI Comparison 2026](https://www.index.dev/skill-vs-skill/swiftui-vs-flutter-vs-react-native)
- [PWA on iOS Limitations 2025](https://brainhub.eu/library/pwa-on-ios)
- [CloudKit - Apple Developer](https://developer.apple.com/icloud/cloudkit/)
- [CKSyncEngine - WWDC23](https://developer.apple.com/videos/play/wwdc2023/10188/)
- [Apple Developer Program Enrollment 2026](https://www.webtonative.com/blog/apple-developer-program-enrollment)
- [gomobile package](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)
