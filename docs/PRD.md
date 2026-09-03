# SideLive — Product Requirements Document

**Status:** Working product specification  
**Project:** SideLive  
**Last updated:** 2026-09-03

---

## 1. Product Overview

SideLive is an open-source, local-first desktop companion for streamers who want to monitor live audience interactions without constantly switching attention to the device or application running the livestream.

The first implementation targets TikTok LIVE comments on desktop. The long-term product is platform-agnostic and supports multiple streaming platforms, multiple live streams, and multiple floating overlays.

> **SideLive — Keep your live audience in sight.**

## 2. Problem Statement

A streamer may broadcast using a phone while focusing on a laptop, game, browser, IDE, presentation, or other primary screen. When the phone is being used as the livestream camera, reading comments on that same phone becomes inconvenient or disruptive.

SideLive solves a narrower problem than OBS or TikTok Live Studio:

> The livestream happens elsewhere. SideLive only helps the streamer stay aware of the live audience.

## 3. Product Vision

SideLive should become a lightweight multi-platform LIVE interaction companion.

```text
TikTok ────┐
YouTube ───┼──► SideLive ───► one or more overlays
Twitch ────┘
```

SideLive monitors live interactions. It does not broadcast video.

## 4. Product Principles

1. Platform-agnostic by design.
2. TikTok-first in implementation.
3. UI-independent core.
4. Local-first.
5. Lightweight.
6. Multi-stream by design.
7. Multi-overlay by design.
8. Failure isolation.
9. Open-source friendly.
10. SDK is a first-class product surface.
11. No mandatory SideLive account.
12. No mandatory SideLive backend.
13. Product scope remains focused on audience interaction, not broadcasting.

## 5. Target Users

- Streamers using a phone as the livestream camera.
- Creators who need to monitor chat while working in another application.
- Gamers who want chat visible without changing windows.
- Developers or creators streaming coding sessions.
- Users running livestreams on more than one platform.
- Developers building custom live-chat tools on top of SideLive.

## 6. Core Product Statement

> SideLive is an open-source, local-first desktop companion that lets streamers monitor live interactions from one or multiple streaming platforms through customizable floating overlays.

## 7. Core Concepts

```text
Provider
   ↓
Stream
   ↓
Normalized Live Events
   ↓
Overlay Source
   ↓
Overlay
```

Primary entities:

- Provider
- Stream
- LiveEvent
- Overlay
- OverlaySource
- ProviderCapability
- Workspace

## 8. Provider

A Provider represents an integration with a livestream platform.

```text
TikTokProvider
YouTubeProvider
TwitchProvider
```

A Provider is not a livestream itself. One Provider can support many Streams.

## 9. Provider Capabilities

```ts
type ProviderCapabilities = {
  comments: boolean
  donations: boolean
  subscriptions: boolean
  follows: boolean
  likes: boolean
  viewers: boolean
}
```

The UI adapts to capabilities rather than assuming every Provider behaves identically.

## 10. Stream

A Stream represents one monitored live source.

```ts
type Stream = {
  id: string
  providerId: string
  identity: {
    id?: string
    username?: string
    displayName?: string
  }
  status: StreamStatus
}
```

Do not model streams as `primaryStream` and `secondaryStream`. Use `streams[]`.

## 11. Stream Lifecycle

Possible states:

```text
idle
connecting
connected
reconnecting
offline
disconnected
failed
```

Each Stream owns an independent lifecycle.

## 12. Live Event

Provider-native events are converted into normalized SideLive events.

```ts
type LiveEvent = {
  id: string
  streamId: string
  provider: string
  type: "comment" | "donation" | "subscription" | "follow"
  timestamp: number
  user: {
    id?: string
    username?: string
    displayName: string
    avatarUrl?: string
  }
  payload: unknown
}
```

The official UI must consume SideLive events, not provider-native event objects.

## 13. Do Not Over-Normalize

TikTok Gift, YouTube Super Chat, and Twitch Bits are similar but not identical. SideLive provides a useful common representation without deleting provider-specific information.

```ts
type LiveEvent<T> = {
  id: string
  streamId: string
  provider: string
  type: string
  timestamp: number
  payload: T
  providerMetadata?: unknown
}
```

## 14. Overlay

Overlay is a logical SideLive entity, not merely a native OS window.

```ts
type Overlay = {
  id: string
  name: string
  sources: OverlaySource[]
  bounds: {
    x: number
    y: number
    width: number
    height: number
  }
  appearance: {
    opacity: number
    fontSize: number
  }
  behavior: {
    alwaysOnTop: boolean
    clickThrough: boolean
  }
}
```

## 15. Overlay Source

```ts
type OverlaySource = {
  overlayId: string
  streamId: string
}
```

OverlaySource explicitly connects Streams to Overlays.

## 16. Stream ↔ Overlay Relationship

The relationship is many-to-many.

Do not implement:

```ts
stream.overlayId
```

Use OverlaySource relationships.

### Separate

```text
TikTok Stream  ───► Overlay A
YouTube Stream ───► Overlay B
```

### Unified

```text
TikTok ────┐
YouTube ───┼──► Overlay A
Twitch ────┘
```

### Hybrid

```text
TikTok ─────┬──► TikTok Overlay
            └──► Unified Overlay

YouTube ────┬──► YouTube Overlay
            └──► Unified Overlay
```

No hardcoded SeparateMode, UnifiedMode, or HybridMode is required.

## 17. Workspace

Workspace is a future-facing aggregate.

```text
Workspace
│
├── Stream[]
└── Overlay[]
```

It may become the long-term persistence/configuration boundary, but it does not need to be prominent in the v0.1 UI.

## 18. Product Modes

### Companion Window
Normal desktop window showing stream connections and live activity.

### Overlay Mode
Transparent, borderless, always-on-top live interaction presentation over another application.

### Focus Mode
Future mode emphasizing high-value interactions such as questions or mentions instead of every event.

## 19. MVP Definition

```text
Provider
└── TikTok

Streams
└── 1 TikTok stream

Events
└── Comment

Overlay
└── 1 overlay
```

The architecture supports more, but v0.1 does not expose everything.

## 20. MVP Features — P0

- Provider system.
- TikTok provider.
- Add/remove stream.
- Connect/disconnect stream.
- Realtime comments.
- Normalized comment model.
- Connection status.
- Automatic reconnect.
- Floating overlay.
- Always-on-top behavior.
- Local configuration.
- No mandatory account.
- No mandatory backend.

## 21. MVP Supporting Features

- Connection health indicator.
- Temporary offline state.
- Overlay position persistence.
- Overlay size persistence.
- Overlay opacity.
- Overlay font size.
- Show/hide overlay.
- Basic click-through.
- Overlay edit/lock mode.

## 22. Post-MVP Core Features — P1

- Multiple TikTok streams.
- Multiple overlays.
- Overlay → Stream assignment.
- Independent overlay positions.
- Independent overlay settings.
- Unified overlay.
- Second platform Provider.

## 23. Provider Roadmap

Tentative sequence:

```text
v0.1   TikTok
later  YouTube
later  Twitch
later  community providers
```

YouTube is especially useful as an architecture validation. If adding YouTube requires rewriting large portions of the core, the Provider abstraction is insufficient.

## 24. Event Roadmap

Potential event types:

- Comment
- Gift / donation / Super Chat / Bits
- Follow
- Share
- Subscribe
- Join
- Likes
- Provider-specific live interaction events

## 25. Filtering

Potential filters:

- comments only;
- hide event types;
- keyword highlight;
- questions only;
- hide spam;
- followers/subscribers only;
- user blacklist;
- mentions only;
- gift/donation thresholds.

Filtering occurs after normalization.

## 26. Highlighting

Potential rules:

- mentions;
- questions;
- moderator/VIP users;
- followers/subscribers;
- donation thresholds;
- repeated topics;
- selected users.

## 27. Question Detection

Question detection can begin with deterministic heuristics. AI is not required initially.

Potential signals:

- question mark;
- question words;
- mention + interrogative pattern;
- repeated similar questions.

## 28. Pin / Answer Queue

Future features:

- pin comment;
- add to answer queue;
- mark answered;
- remove from queue;
- group similar questions.

## 29. Chat Expiration

Potential behavior:

- fixed message duration;
- maximum visible comments;
- pause/freeze;
- dynamic duration based on chat volume;
- pinned messages ignore expiration.

## 30. TTS

Future modes may include:

```text
all comments
questions only
mentions only
selected users
subscribers only
gifts above threshold
```

TTS is deferred because queueing, spam, language, emoji, and interruption policies must be handled well.

## 31. Smart Features

Future possibilities:

- Chat Pulse for rapidly repeated topics;
- similar-comment grouping;
- smart filtering;
- local AI where useful.

AI should not require a SideLive-owned cloud backend by default.

## 32. Customization

Initial:

- opacity;
- font size;
- maximum visible comments.

Later:

- font family;
- background;
- text color;
- spacing;
- border radius;
- animation;
- avatar size;
- message duration;
- themes;
- custom CSS;
- community themes;
- import/export configuration.

## 33. Global Hotkeys

Potential actions:

- show/hide overlay;
- toggle click-through;
- increase/decrease opacity;
- pause chat;
- enter overlay edit mode.

Exact key combinations are not locked by the PRD.

## 34. History

Future session functionality:

- temporary history;
- search;
- export session.

Long-term history is added only when there is a clear need.

## 35. Privacy

SideLive is local-first.

- No mandatory SideLive account.
- No mandatory SideLive backend.
- No database unless needed.
- No mandatory telemetry.
- Diagnostics remain local by default.
- Crash reporting, if introduced, is opt-in.

Preferred data path:

```text
Streaming Platform
        ↓
User Desktop
```

## 36. Reliability Requirements

- Clean disconnect.
- Automatic reconnect.
- Network interruption handling.
- Provider failure isolation.
- Bounded event queues.
- Stable long-running sessions.
- Useful error states.
- Normal reconnect must not require application restart.

## 37. Multi-Stream Isolation

A failure in Stream A must not terminate Stream B or Stream C.

Each Stream is an independent runtime unit.

## 38. Overlay Isolation

One Overlay failure should not require all overlays to restart.

Overlay does not own provider network connections.

## 39. Event Flow

```text
Provider
   ↓
Platform-native event
   ↓
Normalizer
   ↓
LiveEvent
   ↓
Event Router
   ↓
Consumer
   ├── Overlay
   ├── Main Window
   ├── History
   ├── TTS
   └── future consumers
```

## 40. UX Philosophy

SideLive should feel like a companion, not a control room.

The interface should be lightweight, minimal, unobtrusive, quick to configure, usable during an active livestream, and easy to understand at a glance.

Avoid turning SideLive into a large analytics dashboard.

## 41. Main Application Concept

Potential primary areas:

```text
Streams
Overlays
Providers
Settings
```

The main application manages connections and overlay configuration.

## 42. Overlay Concept

Overlay is intended for passive viewing while another application remains primary.

Key behavior:

- always-on-top;
- minimal chrome;
- transparent where supported;
- movable in edit mode;
- click-through in locked mode;
- efficient rendering;
- configurable source assignment.

## 43. Open-Source Goals

- Transparent implementation.
- Contributor-friendly architecture.
- Reusable SDK.
- Provider extension contract.
- Examples using public APIs.
- Community provider support.
- No requirement to use the official UI.

## 44. SDK as a Product Surface

Third-party developers should eventually be able to use SideLive without SideLive Desktop.

Examples:

```text
custom desktop UI
CLI
custom overlay
local automation
accessibility tool
experimental event consumer
```

## 45. Provider Extensibility

Long-term, a developer should be able to add a Provider without modifying SideLive core.

Conceptual community package:

```text
sidelive-provider-kick
```

Exact distribution depends on the implementation stack.

## 46. Out of Scope

SideLive does not aim to provide:

- video streaming;
- screen capture;
- virtual camera;
- OBS replacement;
- TikTok Live Studio replacement;
- video recording;
- stream editing;
- multi-stream broadcasting;
- broadcasting itself;
- stream scheduling;
- analytics dashboard;
- revenue analytics platform.

## 47. Development Priorities

```text
P0 Critical
P1 High
P2 Medium
P3 Low
```

### P0
Provider abstraction, TikTok provider, stream lifecycle, normalized comments, realtime comments, reconnect, single floating overlay, always-on-top.

### P1
Multiple streams, multiple overlays, overlay source assignment, independent overlay configuration.

### P2
Additional event types, filters, highlighting, questions, second Provider.

### P3
TTS, Chat Pulse, smart grouping, sophisticated customization.

## 48. Tentative Milestones

```text
v0.1  TikTok + single stream + single overlay
v0.2  overlay quality and reliability
v0.3  multi-stream + multi-overlay
v0.4  second provider / multi-platform validation
v0.5  filtering / questions / additional events
v0.6+ TTS, Chat Pulse, smarter features
```

## 49. MVP Success Criteria

The MVP succeeds when a user can:

```text
Open SideLive
    ↓
Select TikTok
    ↓
Enter username
    ↓
Connect
    ↓
Receive realtime comments
    ↓
Open floating overlay
    ↓
Position it over another application
    ↓
Continue using the laptop
```

Temporary disconnects should recover without restart, and no project-owned account/backend should be required.

## 50. Architecture Success Criteria

- TikTok is isolated behind Provider boundary.
- UI never consumes TikTok-native event types.
- Adding a second Stream does not require redesign.
- Adding a second Overlay does not require redesign.
- Unified overlays require configuration, not a new architecture.
- Third-party consumers can use the SDK without SideLive Desktop.
- One Stream failure does not kill the application.
- Providers can evolve independently.

## 51. Non-Goals for v0.1

Do not require:

- YouTube;
- Twitch;
- gifts;
- follows;
- subscriptions;
- TTS;
- AI;
- history search;
- analytics;
- cloud sync;
- automatic updates;
- frozen public SDK;
- multiple workspaces.

## 52. Product Vocabulary

Preferred generic names:

```text
provider
stream
event
overlay
overlay source
workspace
```

Avoid names such as `tiktokComments`, `tiktokOverlay`, or `tiktokUsername` outside the TikTok Provider boundary.

## 53. Guiding Product Rule

> **Design for N platforms, N streams, and N overlays. Build TikTok, one stream, one overlay first.**
