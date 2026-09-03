# SideLive — Architecture

**Status:** Accepted conceptual architecture  
**Project:** SideLive  
**Last updated:** 2026-09-03

---

## 1. Document Purpose

This document defines the conceptual software architecture of SideLive. It intentionally separates architecture from technology selection. Technology choices are documented in `TECH_STACK.md`.

## 2. Architectural Goals

SideLive must be platform-agnostic, UI-independent, multi-stream by design, multi-overlay by design, local-first, failure-isolated, provider-extensible, testable without real livestreams, reusable through a public SDK, and suitable for modular open-source development.

## 3. Architectural Principles

1. TikTok is the first Provider, not the center of the domain.
2. Core does not depend on desktop UI.
3. Core does not depend on provider-specific event models.
4. UI does not consume native provider events.
5. Providers normalize before events enter the core.
6. Streams have independent lifecycles.
7. Stream ↔ Overlay is many-to-many.
8. Overlay does not own network connections.
9. Desktop-native behavior belongs to a desktop adapter.
10. Public API is smaller and more stable than internal API.
11. Dependency direction points inward.
12. Extensibility should not require core modification.
13. Package boundaries matter more than Git repository boundaries.
14. Failures should be isolated.
15. Avoid premature abstraction.

## 4. High-Level Architecture

```text
Streaming Platforms
        │
        ▼
    Providers
        │
        ▼
     Streams
        │
        ▼
Normalized Live Events
        │
        ▼
    Event Router
        │
 ┌──────┼───────────────┐
 ▼      ▼               ▼
Overlay Main UI     Future Consumers
                    ├── History
                    ├── TTS
                    └── Automation
```

## 5. Architectural Layers

```text
DOMAIN
────────
core
events
streams
providers
routing
overlay model

       ↓

PUBLIC API
──────────
sdk

       ↓

ADAPTERS
────────
provider-tiktok
provider-youtube
desktop adapter
storage adapter

       ↓

APPLICATIONS
────────────
desktop
cli
examples
third-party applications
```

Dependencies point toward stable inner abstractions.

## 6. Core Domain Model

```text
Workspace
│
├── Stream[]
│     │
│     └── Provider
│
└── Overlay[]
      │
      └── OverlaySource[]
              │
              └── Stream
```

Conceptually:

```text
Provider 1 ─── N Stream
Stream   N ─── N Overlay
```

## 7. Provider

Provider represents a platform integration and is responsible for provider-specific identity, connection, native event reception, capabilities, failure conversion, and normalization.

Provider does not render UI.

## 8. Provider Registry

Concrete Provider implementations are registered at application composition time.

```text
Application Composition
        ↓
Provider Registry
        ├── TikTok
        ├── YouTube
        └── Twitch
```

Core does not hardcode Provider implementations.

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

Use capabilities rather than platform-name branches whenever behavior is truly capability-based.

## 10. Provider Dependency Rule

Correct:

```text
Core Contract
     ↑
TikTok Provider
```

Incorrect:

```text
Core
 ↓
TikTok library
```

Provider implementations depend on SideLive contracts; SideLive core does not depend on concrete provider packages.

## 11. Stream

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

## 12. Stream Identity

SideLive internal IDs are separate from provider identities.

```text
SideLive Stream ID  → stream_...
Provider identity   → TikTok @alice
```

Provider IDs must not become universal SideLive primary keys.

## 13. Stream Lifecycle

```text
idle
  ↓
connecting
  ↓
connected
  ↓
reconnecting
  ↓
connected
```

Other transitions may lead to offline, failed, or disconnected states.

## 14. Stream Manager

Stream Manager coordinates runtime instances: create, connect, disconnect, expose status, cancel work, coordinate reconnect, and contain failures.

It does not render overlays.

## 15. Stream Isolation

```text
Stream A ── connection A
Stream B ── connection B
Stream C ── connection C
```

Failure in one Stream must not automatically stop the others.

## 16. Live Event

```ts
type LiveEvent<T> = {
  id: string
  streamId: string
  provider: string
  type: string
  timestamp: number
  user?: LiveUser
  payload: T
  providerMetadata?: unknown
}
```

Events are normalized and treated as immutable after normalization.

## 17. Live Event Types

Initial:

```text
comment
```

Potential future types:

```text
donation
gift
subscription
follow
share
join
like
viewer
provider-specific extension
```

## 18. Event Normalization

```text
TikTok CommentEvent
       ↓
TikTok Adapter
       ↓
SideLive CommentEvent
```

UI and core consumers never need TikTok SDK objects.

## 19. Provider Metadata

Normalization must not destroy useful provider-specific information. Common data belongs to SideLive models; uncommon data may be retained as provider metadata.

## 20. Event Immutability

Consumers may derive projections from events but should not mutate the shared normalized event.

## 21. Event Router

```text
Provider
   ↓
Stream
   ↓
LiveEvent
   ↓
Event Router
   ├── Overlay A
   ├── Overlay B
   ├── Main Window
   ├── History
   └── TTS
```

The Event Router contains no provider-specific logic.

## 22. Event Consumer Model

Any feature reacting to LiveEvents is conceptually an Event Consumer:

```text
Overlay
History
TTS
Question Queue
Chat Pulse
Export
```

## 23. Event Flow

```text
Platform
   ↓
Provider connection
   ↓
Native event
   ↓
Normalizer
   ↓
LiveEvent
   ↓
Event Router
   ↓
Consumer
```

## 24. Event Ordering

Preserve event order within one Stream where feasible. Do not promise perfect total ordering across providers because latency, timestamps, reconnect behavior, and delivery mechanisms differ.

## 25. Overlay

Overlay is a domain entity describing presentation configuration. It is not equivalent to a native operating-system window.

## 26. Overlay Source

```ts
type OverlaySource = {
  overlayId: string
  streamId: string
}
```

OverlaySource may later carry source-specific filtering or appearance overrides.

## 27. Stream ↔ Overlay Relationship

```text
Stream N ─── N Overlay
```

This is a core architectural contract. Do not put `overlayId` directly on Stream.

## 28. Unified Overlay

```text
TikTok ────┐
YouTube ───┼──► Overlay
Twitch ────┘
```

Unified behavior is configuration, not a separate architecture.

## 29. Hybrid Overlay

```text
Stream A ───┬──► Overlay A
            └──► Overlay C

Stream B ───┬──► Overlay B
            └──► Overlay C
```

## 30. Overlay Manager

Overlay Manager creates, updates, removes, assigns sources, exposes configuration, and coordinates the runtime adapter. It does not connect to livestream providers.

## 31. Overlay Runtime vs Overlay Domain

```text
Overlay domain model
        ↓
Overlay runtime
        ↓
Desktop adapter
        ↓
Native window
```

Core understands configuration. Desktop adapter executes OS behavior.

## 32. Desktop Overlay Adapter

Desktop-only responsibilities:

- create/destroy native window;
- always-on-top;
- click-through;
- move/resize;
- transparent background;
- tray;
- global shortcuts;
- restore bounds.

## 33. Desktop Capability Model

```text
DesktopCapabilities
├── TransparentWindows
├── AlwaysOnTop
├── ClickThrough
└── GlobalShortcuts
```

UI adapts to platform capabilities rather than assuming identical native behavior everywhere.

## 34. Workspace

```text
Workspace
│
├── Streams
└── Overlays
```

Workspace may become the persistence boundary.

## 35. Workspace as Configuration Boundary

Persist stable configuration such as Stream definitions, Overlay definitions, source assignments, appearance, bounds, and preferences.

Do not persist WebSocket handles, goroutines, timers, temporary errors, or native window handles.

## 36. Core vs Runtime State

Core/domain state:

```text
Stream definition
Overlay configuration
Provider capabilities
Workspace
```

Runtime state:

```text
connection
reconnect attempt
heartbeat
event channel
native window
```

## 37. SDK

SideLive SDK is a developer-facing public API. SideLive Desktop is one consumer, not the SDK itself.

Potential consumers include CLI, custom desktop apps, custom overlays, local automation, and testing tools.

## 38. Core vs SDK

```text
core
    internal implementation
    evolves faster

sdk
    intentional public contract
    smaller
    versioned
    documented
```

The SDK must not simply re-export all internals.

## 39. SDK Responsibilities

Potential public surface:

- client construction;
- Provider registration;
- Stream management;
- event subscription;
- normalized models;
- capability inspection.

Exact API is intentionally not locked yet.

## 40. SDK Stability

Internal APIs have no compatibility guarantee. Public SDK changes require deliberate versioning. Before v1.0, breaking changes are allowed but documented.

## 41. Provider SDK / Extension Contract

Community Providers should eventually integrate without forking SideLive core.

```text
sidelive-provider-tiktok
sidelive-provider-youtube
sidelive-provider-twitch
community-provider-kick
```

## 42. Application Composition Root

Concrete dependencies are assembled at the application edge.

```text
Desktop App
   ├── Core
   ├── SDK
   ├── TikTok Provider
   ├── Desktop Adapter
   └── Storage Adapter
```

## 43. Dependency Injection Principle

Prefer explicit dependency passing. No DI framework is architecturally required. Avoid hidden service locators and implicit globals.

## 44. Package Architecture

Conceptually:

```text
apps/
└── desktop/

packages/
├── core/
├── sdk/
├── overlay/
├── provider-tiktok/
└── shared/

examples/
docs/
tooling/
```

Implementation-specific layout is refined in `TECH_STACK.md`.

## 45. Why Monorepo First

A modular monorepo provides atomic changes, easier local development, unified CI, consistent tooling, easier refactors, and simpler contributor onboarding.

## 46. Monorepo Does Not Mean Monolith

> Package boundary is more important than repository boundary.

A single Git repository does not imply tightly coupled modules.

## 47. Package Dependency Direction

```text
applications
      ↓
adapters/providers
      ↓
sdk/application contracts
      ↓
core domain
```

Core never imports desktop UI.

## 48. Repository Extraction Strategy

A package may move to its own repository later for real reasons such as independent maintainers, release cadence, governance, security boundaries, or ecosystem adoption.

Do not split repositories solely because concepts are different.

## 49. Versioning Strategy

Potential independent surfaces:

```text
SideLive Desktop
SideLive SDK
TikTok Provider
```

A TikTok protocol patch should ideally update only its Provider implementation if public contracts remain unchanged.

## 50. Error Model

Potential normalized categories:

```text
ProviderUnavailable
StreamOffline
ConnectionFailed
ConnectionLost
RateLimited
AuthenticationRequired
UnsupportedFeature
InvalidIdentity
InternalProviderError
```

Provider diagnostics may remain attached internally.

## 51. Reconnection Architecture

Reconnect belongs to Stream/provider runtime, never Overlay.

```text
Stream
  ↓
connection lost
  ↓
retry policy
  ↓
reconnect
```

Manual disconnect stops retries immediately.

## 52. Filtering Architecture

```text
Provider
  ↓
normalize
  ↓
LiveEvent
  ↓
Filter
  ↓
Consumer
```

Filtering after normalization lets one filter implementation work across platforms.

## 53. Overlay Filtering

Overlay-specific filters may be attached to Overlay or OverlaySource.

```text
Overlay A → questions only
Overlay B → all comments
Overlay C → gifts above threshold
```

## 54. Internal IDs vs External IDs

Use SideLive-generated internal IDs for entities and preserve provider IDs as external identity.

## 55. Logging Boundary

Logging is infrastructure. Domain logic should not assume cloud telemetry.

## 56. Observability Without Mandatory Telemetry

Support local logs, provider status, runtime diagnostics, and optional debug export. Telemetry is not mandatory.

## 57. Testing Boundaries

Test independently:

```text
core
provider
stream runtime
event router
overlay domain
desktop adapter
SDK
```

Most tests must not require a real TikTok LIVE.

## 58. Demo / Fake Provider

A fake Provider should emit deterministic comments, status changes, disconnects, bursts, and errors for UI development, load testing, architecture tests, contributor onboarding, and CI.

## 59. Replay Provider

A replay Provider may later consume sanitized fixtures to reproduce burst traffic, reconnects, malformed events, event ordering, and edge cases.

## 60. SDK Examples as Architecture Tests

`examples/` consumes only the public SDK. If an ordinary example requires private imports, the SDK is insufficient.

## 61. Example — Single Stream

```text
TikTok Provider
      ↓
TikTok Stream
      ↓
LiveEvent
      ↓
Overlay A
```

## 62. Example — Multi Stream

```text
TikTok Provider
   ├── Stream A
   └── Stream B

YouTube Provider
   └── Stream C
```

## 63. Example — Unified Overlay

```text
Stream A ───┐
Stream B ───┼──► Overlay Unified
Stream C ───┘
```

## 64. Example — Hybrid Setup

```text
Stream A ───┬──► Overlay TikTok
            └──► Overlay Unified

Stream B ───┬──► Overlay YouTube
            └──► Overlay Unified
```

## 65. Architectural Non-Goals

Architecture does not attempt to solve broadcasting, video encoding, screen capture, cloud scale, analytics warehousing, microservices, arbitrary plugin sandboxing, or every future Provider concept on day one.

## 66. Avoid Premature Abstraction

Abstract known stable concepts:

```text
Provider
Stream
LiveEvent
Overlay
OverlaySource
```

Preserve extension points for unknown behavior.

## 67. Avoid Generic Everything

A universal interface that removes important semantics is worse than a small amount of Provider-specific extension data.

## 68. Architecture Evolution Strategy

```text
1. TikTok single stream
2. TikTok + single overlay
3. multiple TikTok streams
4. multiple overlays
5. second Provider
6. unified overlay
7. additional consumers
```

Each stage validates the architecture.

## 69. Architecture Validation Criteria

Architecture remains healthy if:

- second Stream does not require redesign;
- second Overlay does not require redesign;
- second Provider does not force UI rewrite;
- unified Overlay is configuration;
- third-party SDK consumer does not require desktop dependencies;
- TikTok library can be replaced inside its adapter;
- fake Providers work without network access.

## 70. Dependency Smell Checklist

Investigate if:

- core imports a Provider library;
- core imports React/UI;
- Provider imports desktop window APIs;
- Overlay reconnects Streams;
- React stores Provider SDK objects;
- second Stream introduces primary/secondary branches;
- unified chat requires a separate event model;
- examples import internal packages;
- one Stream crash kills all Streams.

## 71. Architectural Decision Records

Use:

```text
docs/adr/
├── 0001-modular-monorepo.md
├── 0002-provider-abstraction.md
├── 0003-normalized-events.md
├── 0004-stream-overlay-many-to-many.md
└── 0005-ui-independent-core.md
```

Each ADR contains Context, Decision, Alternatives, and Consequences.

## 72. Initial Repository Shape

```text
sidelive/
│
├── apps/
│   └── desktop/
│
├── packages/
│   ├── core/
│   ├── sdk/
│   ├── overlay/
│   ├── provider-tiktok/
│   └── shared/
│
├── examples/
│   └── basic-chat/
│
├── docs/
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── TECH_STACK.md
│   └── adr/
│
├── tests/
└── tooling/
```

The implementation-specific shape is refined in `TECH_STACK.md`.

## 73. Initial Runtime Model

```text
SideLive Desktop
      │
      ├── Provider Registry
      │      └── TikTok Provider
      │
      ├── Stream Manager
      │      └── Stream A
      │
      ├── Event Router
      │
      └── Overlay Manager
             └── Overlay A
```

## 74. Target Runtime Model

```text
SideLive Runtime
│
├── Provider Registry
│   ├── TikTok
│   ├── YouTube
│   └── Twitch
│
├── Stream Manager
│   ├── Stream A
│   ├── Stream B
│   └── Stream C
│
├── Event Router
│
├── Overlay Manager
│   ├── Overlay A
│   ├── Overlay B
│   └── Overlay C
│
└── Other Consumers
    ├── History
    ├── TTS
    └── Focus
```

## 75. Architecture Summary

```text
Provider
   ↓
Stream
   ↓
Normalized Event
   ↓
Event Router
   ↓
Consumer
```

Overlay is one Consumer. Desktop is one application. TikTok is one Provider.

## 76. Core Architectural Contract

1. SideLive is platform-agnostic.
2. SideLive core is UI-independent.
3. TikTok is the first Provider, not the domain center.
4. Providers produce normalized events.
5. Streams have independent lifecycle.
6. SideLive can have many Streams.
7. SideLive can have many Overlays.
8. Stream ↔ Overlay is many-to-many.
9. Overlay does not own provider connection.
10. Core does not depend on desktop UI.
11. SDK is separate from internal core.
12. Providers can be developed independently.
13. Repository begins as a modular monorepo.
14. Package boundary matters more than repository boundary.
15. Desktop-specific capabilities live in a desktop adapter.
16. Failures are isolated per Stream/provider/runtime component.

## 77. Guiding Engineering Rule

> **Design for N platforms. Build TikTok first.**

> **Design for N Streams and N Overlays. Expose one of each first.**

## 78. Related Documents

```text
PRD.md
    What are we building and why?

ARCHITECTURE.md
    How is the system conceptually structured?

TECH_STACK.md
    Which technologies implement the architecture?
```

Potential later documents:

```text
EVENT_MODEL.md
PROVIDER_SPEC.md
REPOSITORY.md
```

Only create them when implementation complexity justifies them.
