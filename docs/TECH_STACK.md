# SideLive — Technical Stack

**Status:** Accepted for initial implementation  
**Last updated:** 2026-09-03  
**Project:** SideLive  
**Architecture reference:** `ARCHITECTURE.md`  
**Product reference:** `PRD.md`

---

## 1. Purpose

This document defines the initial technical stack for SideLive.

The goal is not to choose the largest or most fashionable stack. The stack must support the architectural contract already defined for SideLive:

- platform-agnostic domain model;
- TikTok-first implementation;
- multi-stream by design;
- multi-overlay by design;
- Stream ↔ Overlay many-to-many;
- local-first;
- lightweight desktop runtime;
- low idle resource usage;
- failure isolation;
- reusable core and SDK;
- provider implementations isolated from the core;
- official desktop UI as only one consumer of the core;
- good contributor experience;
- strong long-term maintainability.

Electron is intentionally excluded.

---

# 2. Final Stack Decision

SideLive will start with:

```text
Desktop runtime       Wails v3
Backend / core        Go
Frontend              React
Frontend language     TypeScript
Bundler               Vite
Styling               Tailwind CSS
Frontend state        Zustand
Go ↔ Frontend RPC     Wails generated bindings
Go → Frontend events  Wails event system
Persistence           JSON configuration via Go
Logging               Go slog
Frontend formatting   Biome
Frontend testing      Vitest + Testing Library
Go testing            standard go test
E2E / desktop tests   Playwright later + focused native integration tests
Package manager       pnpm
CI                    GitHub Actions
Release               Wails packaging + GitHub Releases
Auto-update           Wails updater, post-MVP
First provider        TikTok
TikTok implementation PirateTok/live-go behind our provider adapter
```

Recommended baseline versions at project initialization:

```text
Go          1.27.x
Wails       v3.0.0-beta.x, pinned to an exact beta
React       19.2.x
TypeScript  7.x
Vite        8.1.x
Biome       2.5.x
Vitest      4.1.x
pnpm        current stable major
```

Exact patch versions must be committed through lockfiles/tooling and upgraded intentionally.

We should not use floating `latest` versions in CI.

---

# 3. Why Wails v3

Wails v3 is selected instead of Wails v2.

Although Wails v3 is still beta at the time of this decision, its desktop API is documented as stable during the beta period and its window model matches SideLive unusually well.

SideLive requires:

```text
multiple windows
frameless windows
always-on-top
transparent overlays
click-through overlays
window positioning
system tray
global shortcuts
native packaging
type-safe Go ↔ TypeScript communication
```

Wails v3 provides native support for almost all of these concerns.

The most important difference for SideLive is that multi-window is a first-class concept in Wails v3 rather than something we need to design around.

Example conceptual mapping:

```text
SideLive Main Window
        │
        ├── Overlay A
        ├── Overlay B
        └── Overlay C
```

Each SideLive Overlay maps naturally to a Wails `WebviewWindow`.

---

# 4. Why Wails Instead of Electron

Electron is explicitly excluded from SideLive.

SideLive is a companion application expected to remain open for long livestream sessions while the user is also running:

- games;
- browsers;
- IDEs;
- streaming applications;
- video calls;
- other resource-intensive software.

Bundling a dedicated Chromium and Node runtime does not align with the project's lightweight objective.

Wails instead uses the operating system's native WebView:

```text
Windows → WebView2
macOS   → WebKit
Linux   → WebKitGTK
```

The application backend compiles to a native Go binary.

This does not make WebViews free, especially when several overlay windows exist, but it avoids shipping a complete independent browser runtime with SideLive.

Performance will still be benchmarked rather than assumed.

---

# 5. Known Wails v3 Risk

Wails v3 is beta.

We accept this risk because:

1. SideLive is itself a new project with no production compatibility burden yet.
2. The Wails v3 desktop API is already declared stable during beta.
3. The v3 window architecture is substantially better aligned with SideLive.
4. We can pin an exact beta version instead of following `latest`.
5. Wails is isolated primarily at the application/desktop adapter boundary.

The version must be pinned.

Do not use this pattern in reproducible builds:

```bash
go get github.com/wailsapp/wails/v3@latest
```

Instead, SideLive commits an exact known-good version.

Upgrading Wails is an intentional maintenance task with overlay regression testing.

---

# 6. Platform Strategy

Initial development priority:

```text
P0  Windows
P1  macOS
P1  Linux
```

The architecture remains cross-platform from day one.

However, identical native window behavior must not be assumed across operating systems.

---

# 7. Critical Linux Overlay Limitation

This is the most important platform caveat found during technical research.

Wails v3 currently documents:

```text
macOS   transparent window backgrounds supported
Windows transparent window backgrounds supported
Linux   solid background only
```

for the relevant `BackgroundType` API.

Linux window managers also have different behavior regarding:

- always-on-top;
- positioning;
- resizing;
- fullscreen overlays;
- compositor transparency.

Therefore SideLive must not promise identical overlay behavior on every Linux desktop environment in v0.1.

The application architecture must support platform capabilities.

Conceptually:

```go
type DesktopCapabilities struct {
    TransparentWindows bool
    ClickThrough       bool
    AlwaysOnTop        bool
    GlobalShortcuts    bool
}
```

UI behavior adapts to capabilities instead of pretending unsupported features exist.

This mirrors the provider capability system already defined in the architecture.

---

# 8. Backend Language — Go

Go owns the SideLive application domain and runtime.

The frontend must not become the source of truth for stream connections.

Go owns:

```text
Provider Registry
Stream Manager
Stream lifecycle
Connection state
Reconnect policy
Event normalization
Event Router
Overlay configuration domain
Workspace configuration
Persistence
Logging
Diagnostics
```

The frontend owns presentation.

This division is intentional.

---

# 9. Why Go

Go is a strong fit for SideLive's workload.

SideLive is dominated by:

- long-running network connections;
- WebSockets;
- event streams;
- independent stream lifecycles;
- cancellation;
- reconnection;
- fan-out;
- concurrent providers.

These map naturally to:

```text
goroutines
channels
context.Context
interfaces
sync primitives
```

Example:

```text
TikTok stream ───────┐
                     │
YouTube stream ──────┼──→ Event Router
                     │
Twitch stream ───────┘
```

Each connection can run independently while emitting normalized events.

Failure in one stream must not crash another.

---

# 10. Concurrency Rules

Goroutines are not an architecture.

They are an implementation primitive.

SideLive will follow these rules:

1. Every long-running operation must have an owning lifecycle.
2. Long-running provider operations receive a `context.Context`.
3. A stream owns cancellation for its connection.
4. Goroutines must not be spawned without a shutdown path.
5. Provider goroutines must not directly update UI state.
6. Event channels must have documented ownership and closure semantics.
7. Unbounded queues are forbidden.
8. Slow consumers must not permanently block provider network loops.
9. One provider panic must never crash the whole application.

Conceptual stream runtime:

```go
type Runtime struct {
    cancel context.CancelFunc
    done   chan struct{}
}
```

---

# 11. Go Package Strategy

SideLive starts as a single Go module inside the monorepo.

This is intentional.

We will not begin with several Go modules merely to imitate npm package boundaries.

Initial structure:

```text
sidelive/
├── go.mod
├── go.sum
│
├── cmd/
│   └── sidelive/
│
├── internal/
│   ├── app/
│   ├── core/
│   ├── streams/
│   ├── events/
│   ├── overlays/
│   ├── routing/
│   ├── config/
│   └── desktop/
│
├── sdk/
│
└── providers/
    └── tiktok/
```

`internal/` contains implementation details.

`sdk/` is the deliberate public Go API.

`providers/` contains provider adapters.

This provides clean boundaries without early multi-module versioning overhead.

If provider release cadence later becomes independent, a provider may be moved into its own Go module or repository without changing the core domain contract.

---

# 12. Public SDK

The public SDK is a first-class product surface.

It is not the Wails binding layer.

These are separate concepts:

```text
SideLive Go SDK
    ↓
public API for developers

Wails bindings
    ↓
private desktop UI bridge
```

Third-party developers must not depend on:

```text
internal/*
Wails runtime types
React types
desktop window types
```

Desired future experience:

```go
client := sidelive.NewClient()

client.RegisterProvider(tiktok.New())

stream, err := client.CreateStream(...)
```

The exact API is not locked yet.

The SDK should be designed only after the first vertical slice proves the internal model.

Do not prematurely freeze interfaces.

---

# 13. Provider Contract

Providers implement a SideLive-owned interface.

The core never imports provider-specific event models.

Conceptually:

```go
type Provider interface {
    ID() ProviderID
    Capabilities() ProviderCapabilities

    Connect(
        ctx context.Context,
        config StreamConfig,
        sink EventSink,
    ) error
}
```

The exact signature may change during the initial implementation.

Important direction:

```text
SideLive Core
      ↑
Provider Contract
      ↑
TikTok Adapter
```

Never:

```text
SideLive Core
      ↓
TikTok SDK types
```

---

# 14. TikTok Provider Decision

The first experimental TikTok implementation will use:

```text
github.com/PirateTok/live-go
```

behind our own `providers/tiktok` adapter.

Reasons for choosing it for the initial spike:

- native Go implementation;
- context-oriented connection API;
- event channel API;
- WebSocket-based realtime events;
- comments and additional LIVE event types;
- documented automatic reconnection;
- no Node/Python sidecar required;
- no signing server or API key required according to its current documentation;
- recent 2026 activity.

However, this dependency is considered **replaceable infrastructure**.

It is not part of SideLive's architecture contract.

TikTok's Webcast interface is unofficial and can change without notice.

Therefore:

```text
PirateTok/live-go
        ↓
TikTok Adapter
        ↓
Normalized SideLive Events
        ↓
SideLive Core
```

If the library breaks, we replace the adapter implementation rather than rewrite SideLive.

Alternatives retained for evaluation/fallback include:

```text
Davincible/gotiktoklive
steampoweredtaco/gotiktoklive
direct protocol implementation
```

No direct protocol rewrite should be attempted before existing libraries are proven inadequate.

---

# 15. TikTok Spike Gate

Before implementing most UI features, the TikTok provider must pass a technical spike.

Success criteria:

```text
username
   ↓
resolve LIVE room
   ↓
connect
   ↓
receive comment
   ↓
normalize
   ↓
print LiveEvent
```

Minimum validation:

- connect to a currently live user;
- receive realtime comments;
- disconnect cleanly;
- cancellation works;
- offline user produces useful error;
- temporary network loss does not crash the app;
- library errors can be converted to SideLive errors;
- connection can run for an extended session;
- two simultaneous connections can run independently.

If these fail consistently, the provider dependency is reconsidered before building deeper product features.

---

# 16. Frontend — React

React 19 is selected for the desktop frontend.

React is responsible for:

```text
main application UI
overlay rendering
settings
forms
stream cards
event presentation
animations
interaction
```

React does not own stream connections.

No networking to TikTok should originate from React.

---

# 17. TypeScript

TypeScript is mandatory for all frontend code.

We will use strict typing.

Recommended configuration principles:

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true
  }
}
```

Exact settings may be adjusted for compatibility.

SideLive should use current stable TypeScript.

At initialization time, TypeScript 7 is preferred if the Vite/React/Wails toolchain validates cleanly.

If generated Wails bindings or frontend tooling have compatibility problems, TypeScript 6 is the approved fallback.

This is a tooling compatibility decision, not an architectural one.

---

# 18. Vite

Vite is selected as the frontend build tool.

Reasons:

- fast development cycle;
- native React ecosystem fit;
- Wails works naturally with frontend bundlers such as Vite;
- Vite 8 uses Rolldown as its unified bundler;
- simple production output suitable for embedding into Wails.

No Next.js.

SideLive is a local desktop application, not a server-rendered website.

SSR, server actions and route-level server architecture provide no meaningful benefit here.

---

# 19. pnpm

pnpm is selected as the JavaScript package manager.

Responsibilities:

```text
dependency installation
frontend scripts
workspace packages
lockfile
```

Root:

```text
pnpm-workspace.yaml
```

Initially the frontend may be the only JavaScript workspace package.

We still establish pnpm workspace semantics because examples or reusable frontend packages may be added later.

---

# 20. Styling — Tailwind CSS

Tailwind CSS is selected.

Reasons:

- fast UI iteration;
- good fit for custom overlay design;
- easy dynamic variants;
- familiar project ecosystem;
- no runtime CSS-in-JS dependency;
- good control over minimal, transparent and responsive overlay presentation.

Tailwind should not leak into core logic.

---

# 21. Component Strategy

Do not install a large component framework.

Use:

```text
React
Tailwind
small accessible primitives where necessary
local SideLive components
```

shadcn-style component ownership is acceptable.

The preference is copied/owned components rather than runtime dependence on a large visual framework.

SideLive needs a distinctive compact overlay UI rather than a generic dashboard visual system.

---

# 22. Frontend State — Zustand

Zustand is selected for frontend-local state.

Important boundary:

Zustand is **not** the canonical SideLive domain store.

Go remains the source of truth for:

```text
streams
connections
provider state
overlay configuration
workspace configuration
```

Zustand owns ephemeral presentation state such as:

```text
selected stream
active settings panel
temporary filters
overlay editor selection
draft form state
local UI preferences
frontend event buffers
```

For domain updates:

```text
Go runtime
   ↓
Wails event
   ↓
frontend adapter
   ↓
Zustand projection
   ↓
React
```

This avoids duplicate business logic.

---

# 23. Why Not Redux

Redux is not selected.

SideLive does not currently need:

- large reducer architecture;
- action middleware ecosystem;
- normalized frontend domain database;
- global state as the source of truth.

The authoritative domain runtime already exists in Go.

Adding Redux would duplicate architecture unnecessarily.

---

# 24. Why Not TanStack Query Initially

TanStack Query is not part of the initial stack.

SideLive is not primarily an HTTP request/cache application.

Most data arrives through:

```text
Wails method calls
+
realtime events
```

If a future provider or cloud service introduces meaningful request caching requirements, TanStack Query can be introduced locally.

Do not use a caching framework for a problem we do not yet have.

---

# 25. Go ↔ React Communication

Communication has two different semantics.

## Commands

Frontend asks Go to perform an operation:

```text
ConnectStream
DisconnectStream
CreateOverlay
UpdateOverlay
GetWorkspace
SaveSettings
```

Use Wails generated method bindings.

Example:

```text
React
  │
  │ typed method call
  ▼
Wails generated binding
  │
  ▼
Go application service
```

## Events

Go informs the frontend that something happened:

```text
comment received
connection status changed
stream ended
provider error
overlay state changed
```

Use the Wails event system.

Do not implement our own localhost REST API or WebSocket server for desktop IPC.

Wails already provides the bridge and generated typed bindings.

---

# 26. Binding Boundary

The Wails binding layer belongs to the desktop application adapter.

It must not become the public SideLive SDK.

Recommended layering:

```text
React
   ↓
desktop frontend adapter
   ↓
generated Wails bindings
   ↓
application services
   ↓
SideLive core
```

React components should preferably call small frontend service wrappers rather than generated binding files everywhere.

This gives us a replacement boundary if Wails changes.

---

# 27. Normalized Events

Provider-native events are converted once, immediately after the provider boundary.

Conceptual Go model:

```go
type LiveEvent struct {
    ID        EventID
    StreamID  StreamID
    Provider  ProviderID
    Type      EventType
    Timestamp time.Time
    User      User
    Payload   EventPayload
}
```

Provider-specific diagnostic metadata may be retained separately.

The frontend receives SideLive events, never PirateTok events.

---

# 28. Event Delivery

For high-frequency events, do not expose a separate Wails call per component.

Go receives provider events and publishes normalized events.

Conceptual pipeline:

```text
Provider
   ↓
Normalizer
   ↓
Event Router
   ↓
Desktop Event Bridge
   ↓
React
```

A bounded frontend event buffer should be used.

Rendering every event forever is forbidden.

Overlay rendering should keep only the relevant visible/recent messages.

This matters more for performance than micro-optimizing Go structs.

---

# 29. Overlay Runtime

Logical Overlay configuration belongs to the domain.

Native windows belong to the desktop adapter.

```text
Overlay domain
     ↓
Overlay Manager
     ↓
Wails Window Adapter
     ↓
native WebView window
```

Wails-specific types must not appear in the core overlay model.

---

# 30. Overlay Window Configuration

The primary overlay should use Wails capabilities equivalent to:

```text
Frameless
AlwaysOnTop
IgnoreMouseEvents when click-through is enabled
Transparent background where supported
Explicit position
Explicit size
```

The overlay begins with click-through disabled while editing.

Typical lifecycle:

```text
Edit Mode
    ↓
draggable / resizable
    ↓
Lock Overlay
    ↓
click-through enabled
```

A global shortcut must allow restoring interactive mode.

This prevents users from making an overlay impossible to control.

---

# 31. Multi-Window Performance Rule

Each overlay may create another system WebView.

Therefore unlimited overlays are not free.

SideLive should not prematurely optimize this, but it must measure:

```text
idle memory
memory per overlay
CPU while idle
CPU during high-volume chat
startup latency
event-to-render latency
```

Benchmark scenarios:

```text
1 stream + 1 overlay
1 stream + 3 overlays
3 streams + 1 unified overlay
3 streams + 3 overlays
high-volume synthetic provider
```

Performance budgets should be defined after the first working vertical slice.

---

# 32. Fake Provider

A deterministic fake provider is required early.

Example:

```text
provider-demo
```

It emits:

```text
comments
connection state
disconnects
bursts
errors
```

without TikTok.

Benefits:

- UI development without a real LIVE;
- reproducible tests;
- load testing;
- overlay testing;
- contributor onboarding;
- provider-independent architecture validation.

This is development infrastructure, not a user-facing MVP feature.

---

# 33. Persistence

No database is required for v0.1.

Configuration will be stored through Go in the operating system's user config directory.

Example conceptual files:

```text
config.json
workspace.json
```

Persist:

```text
stream definitions
overlay definitions
overlay bounds
appearance
behavior
application preferences
```

Do not persist:

```text
active WebSocket connection
goroutines
reconnect timers
runtime errors
current connection handles
```

Runtime state is rebuilt when SideLive starts.

---

# 34. Why JSON First

JSON is selected because:

- trivial to inspect;
- trivial to debug;
- easy migration during early development;
- no schema engine;
- no database lifecycle;
- ideal for small local configuration.

The storage layer must still be abstracted behind a repository/config service.

This allows later migration to SQLite without changing domain services.

---

# 35. SQLite Decision

SQLite is deferred.

Introduce SQLite when SideLive adds features such as:

```text
large chat history
session search
analytics-like local queries
long-lived event archive
```

Do not add SQLite merely because desktop applications often use it.

---

# 36. Configuration Migrations

Even JSON needs a version.

Example:

```json
{
  "version": 1,
  "workspace": {}
}
```

Every persisted schema must be migratable.

Never assume configuration written by v0.1 will remain structurally identical forever.

---

# 37. Logging

Backend logging uses the Go standard library:

```text
log/slog
```

Use structured logs.

Example fields:

```text
provider
stream_id
overlay_id
event_type
operation
duration
error
```

Do not log full chat content by default.

Do not log authentication cookies, tokens or sensitive provider payloads.

---

# 38. Local Diagnostics

SideLive is local-first.

There is no mandatory telemetry.

Initial diagnostics:

```text
local structured logs
version information
provider status
runtime information
optional exported diagnostics bundle later
```

Crash/error telemetry may only be introduced as explicit opt-in functionality.

---

# 39. Frontend Formatting and Linting

Use Biome.

Biome handles:

```text
formatting
linting
import organization
```

Do not use Prettier + ESLint unless a concrete unsupported rule requires ESLint.

The goal is one fast frontend toolchain.

Type correctness remains the responsibility of TypeScript.

---

# 40. Go Formatting and Static Analysis

Mandatory:

```text
gofmt
go vet
go test
```

Recommended CI lint:

```text
golangci-lint
```

Do not enable hundreds of strict linters on day one.

Start with a useful, low-noise configuration and tighten it based on real bugs.

---

# 41. Frontend Testing

Use:

```text
Vitest
React Testing Library
```

Test:

```text
event-to-view transformations
overlay filtering
frontend store behavior
important components
settings flows
```

Avoid snapshot-heavy tests.

Prefer behavior tests.

---

# 42. Go Testing

Use Go's standard test tooling:

```bash
go test ./...
```

Important areas:

```text
event normalization
stream lifecycle
reconnection
router fan-out
filtering
configuration migration
provider adapters
failure isolation
```

Concurrency-sensitive tests should also run with:

```bash
go test -race ./...
```

in CI where practical.

---

# 43. Provider Replay Tests

Provider tests should not depend exclusively on a live TikTok stream.

Where legally and technically appropriate, retain sanitized/replayable protocol fixtures or normalized event fixtures.

The ideal testing chain:

```text
captured fixture
      ↓
provider parser
      ↓
normalized event
      ↓
assert
```

The fake provider covers higher-level SideLive behavior.

---

# 44. Desktop Integration Tests

Full desktop E2E testing is deferred until the vertical slice works.

Use focused integration tests for native requirements:

```text
window creation
overlay creation
always-on-top configuration
click-through toggle
position restoration
multi-window isolation
```

Playwright may test the web UI surface where practical, but it is not assumed to fully validate OS window semantics.

Native behavior needs platform-specific smoke tests.

---

# 45. Performance Testing

Performance is a product requirement, not an assumption based on language choice.

Create a benchmark mode using the fake provider.

Measure:

```text
process memory
CPU
event throughput
event routing latency
render latency
overlay count scaling
stream count scaling
```

Test at event rates above normal usage.

Example synthetic loads:

```text
10 events/sec
100 events/sec
500 events/sec
```

The goal is not to render 500 messages/sec.

The goal is to ensure SideLive remains responsive and applies backpressure/drop policies safely.

---

# 46. Event Backpressure

A viewer should not lose control of the desktop because a provider emits an extreme burst.

The system requires bounded queues.

Potential policy:

```text
Provider Network Reader
       ↓
bounded internal queue
       ↓
Normalizer
       ↓
Event Router
       ↓
consumer-specific bounded queue
```

Presentation consumers may discard obsolete events.

Connection/control events have higher importance than transient chat render events.

Exact policies will be decided after profiling.

---

# 47. Build Tooling

Wails v3 uses its own build/task workflow.

Primary developer commands should eventually be wrapped at repository level:

```bash
pnpm dev
pnpm test
pnpm lint
pnpm build
```

These may delegate to:

```text
wails3 dev
go test
Biome
Vitest
wails3 build
```

Contributors should not need to remember many implementation-specific commands.

---

# 48. Repository Layout

Initial target:

```text
sidelive/
│
├── cmd/
│   └── sidelive/
│
├── internal/
│   ├── app/
│   ├── core/
│   ├── events/
│   ├── streams/
│   ├── overlays/
│   ├── routing/
│   ├── config/
│   └── desktop/
│
├── providers/
│   ├── demo/
│   └── tiktok/
│
├── sdk/
│
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   ├── components/
│   │   ├── features/
│   │   ├── overlays/
│   │   ├── stores/
│   │   └── lib/
│   └── bindings/
│
├── docs/
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── TECH_STACK.md
│   └── adr/
│
├── examples/
│
├── scripts/
│
├── go.mod
├── go.sum
├── package.json
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
└── Taskfile.yml
```

The exact Wails-generated structure may require adjustments.

Architecture boundaries take priority over visual folder symmetry.

---

# 49. Application Composition Root

`cmd/sidelive` is the composition root.

It creates and wires:

```text
config repository
logger
provider registry
stream manager
event router
overlay manager
desktop services
Wails application
```

Core packages should not instantiate global dependencies internally.

Explicit construction is preferred.

No dependency injection framework.

---

# 50. Error Model

Provider-native errors are converted into SideLive errors.

Conceptual categories:

```text
ProviderUnavailable
StreamOffline
ConnectionFailed
ConnectionLost
AuthenticationRequired
RateLimited
UnsupportedFeature
InvalidIdentity
InternalProviderError
```

Provider diagnostics may be attached internally.

UI consumes normalized errors suitable for presentation.

---

# 51. Reconnection

Reconnection belongs to the stream/provider runtime layer.

Never the overlay.

Conceptually:

```text
stream
  ↓
connect
  ↓
connection lost
  ↓
backoff
  ↓
reconnect
```

Use:

```text
context cancellation
bounded exponential backoff
jitter
maximum retry policy where appropriate
```

The user can always manually disconnect, which must cancel retries immediately.

---

# 52. Security Model

SideLive is not a browser wrapper for arbitrary web content.

Rules:

- do not expose unnecessary Go services to Wails;
- validate every frontend argument in Go;
- do not execute arbitrary provider payloads;
- sanitize URLs before exposing them to UI;
- no arbitrary remote HTML in overlay windows;
- never persist provider credentials unencrypted if credentials are later required;
- secrets must never be placed in frontend state unless unavoidable;
- do not enable DevTools in production builds by default.

Provider integrations are a security boundary.

---

# 53. Packaging

Use Wails native packaging.

Targets:

```text
Windows
    installer / executable

macOS
    .app / DMG

Linux
    native package formats where supported
```

Release artifacts are produced by CI.

Do not build release artifacts manually on a developer laptop as the canonical release process.

---

# 54. CI

Use GitHub Actions.

Minimum PR pipeline:

```text
Go format check
Go vet
Go test
Go race tests where practical
golangci-lint
Biome check
TypeScript typecheck
Vitest
frontend build
Wails build smoke test
```

Release pipeline adds platform builds and packaging.

Matrix:

```text
Windows
macOS
Linux
```

Native window behavior cannot be completely validated through compilation alone.

---

# 55. Release Strategy

Early versions use semantic-ish pre-1.0 releases:

```text
v0.1.0
v0.2.0
v0.3.0
```

Breaking changes to the public SDK are allowed before v1.0 but must be documented.

Internal packages have no compatibility promise.

Provider changes may be released independently later if/when providers become separate modules.

---

# 56. Auto Update

Auto-update is not required for the first vertical slice.

When introduced, use Wails v3's updater rather than implementing a custom update protocol.

GitHub Releases is the preferred initial release provider.

Update signing must be configured before enabling automatic installation.

---

# 57. No Mandatory Backend Service

SideLive will not have a project-owned cloud backend for the MVP.

Architecture:

```text
Streaming Platform
        ↓
SideLive Desktop
```

Not:

```text
Streaming Platform
        ↓
SideLive Server
        ↓
SideLive Desktop
```

This keeps:

```text
cost low
privacy high
latency low
deployment simple
open-source self-use easy
```

A cloud service may only be introduced later for a feature that clearly requires one.

---

# 58. Dependency Policy

Prefer the standard library where it is sufficient.

A dependency is justified when it:

- solves a real difficult problem;
- is maintained;
- reduces platform complexity;
- has acceptable licensing;
- can be isolated behind our interface when unstable.

Avoid dependency chains for trivial helpers.

Provider libraries especially must remain replaceable.

---

# 59. What We Explicitly Do Not Use

Initial stack does **not** include:

```text
Electron
Node backend
Python sidecar
Rust core
Tauri
Next.js
Redux
TanStack Query
GraphQL
local HTTP API
local WebSocket IPC server
PostgreSQL
MySQL
SQLite
Redis
Docker
Kubernetes
cloud backend
microservices
```

This list can change if future requirements justify it.

It is not ideological.

It exists to prevent accidental architecture inflation.

---

# 60. Initial Vertical Slice

The first real application milestone is:

```text
Open SideLive
      ↓
Enter TikTok username
      ↓
Go TikTok provider connects
      ↓
Provider event received
      ↓
Normalize to SideLive LiveEvent
      ↓
Event Router
      ↓
Wails desktop bridge
      ↓
React main window
      ↓
Create overlay
      ↓
Comment appears in overlay
```

Then:

```text
network interruption
      ↓
reconnect
      ↓
UI status updates
      ↓
comments continue
```

No additional feature should take priority over proving this path.

---

# 61. Technical Spike Order

Before polishing UI:

## Spike A — Wails Overlay

Prove:

```text
main window
second window
frameless
always-on-top
transparent background
click-through
drag/edit mode
global shortcut
position restore
```

Test Windows first.

Also document macOS and Linux differences.

## Spike B — TikTok Go Provider

Prove:

```text
username
connect
comment
disconnect
cancel
reconnect
two simultaneous streams
```

## Spike C — Event Bridge

Prove:

```text
Go synthetic event
     ↓
Wails event
     ↓
React
     ↓
overlay
```

Only after these three spikes pass do we treat the stack as fully validated.

---

# 62. Architecture Mapping

```text
                   SideLive Desktop
                         │
            ┌────────────┴────────────┐
            │                         │
       React Frontend            Wails Runtime
            │                         │
            └──── typed bindings ─────┘
                         │
                  Application Layer
                         │
        ┌────────────────┼────────────────┐
        │                │                │
 Stream Manager      Event Router    Overlay Manager
        │                │                │
        └────────────────┼────────────────┘
                         │
                     Core Domain
                         │
                 Provider Contract
                         │
            ┌────────────┴────────────┐
            │                         │
      Demo Provider              TikTok Provider
                                      │
                              PirateTok/live-go
```

Long-term:

```text
                         Provider Contract
                    ┌─────────┼─────────┐
                    │         │         │
                 TikTok    YouTube    Twitch
```

No provider-specific type crosses the Provider Contract boundary.

---

# 63. State Ownership Summary

| State | Owner |
|---|---|
| Provider connection | Go |
| Stream lifecycle | Go |
| Reconnect state | Go |
| Normalized events | Go |
| Workspace configuration | Go |
| Overlay domain configuration | Go |
| Native overlay windows | Wails desktop adapter |
| React rendering | React |
| Current frontend selection | Zustand |
| Form drafts | Zustand / local React state |
| Temporary rendered event list | frontend |
| Long-term history | future persistence layer |

The answer to "where should this state live?" should normally be determined by this table.

---

# 64. Performance Philosophy

SideLive chooses Go + Wails partly for efficiency, but no performance claim is accepted without measurement.

Priority order:

```text
1. Correctness
2. Resource stability during long sessions
3. UI responsiveness
4. Predictable memory usage
5. Low idle CPU
6. Startup speed
7. Binary size
```

Do not sacrifice maintainability for tiny synthetic benchmark improvements.

The biggest likely performance risks are not Go itself.

They are more likely to be:

```text
too many WebView windows
unbounded event history
excessive React re-renders
large avatars/images
unbounded event channels
bad reconnect loops
verbose logging
```

Optimize those first.

---

# 65. Developer Experience

Target onboarding:

```bash
git clone ...
cd sidelive
pnpm install
wails3 doctor
pnpm dev
```

Long-term we should provide a single setup document and optionally a bootstrap task.

Development should provide:

```text
Go hot reload
React HMR
generated bindings
clear logs
fake provider
replayable events
```

A frontend contributor should not need TikTok LIVE access to work on the UI.

---

# 66. Learning Curve Strategy

Because Go is also a learning goal for the initial maintainer, SideLive should use idiomatic Go rather than introducing complex framework abstractions.

Prefer:

```text
interfaces
structs
constructors
context.Context
channels where appropriate
standard errors
explicit dependencies
small packages
```

Avoid prematurely introducing:

```text
large DI framework
actor framework
custom async abstraction
enterprise repository framework
generic event framework
reflection-heavy magic
```

The project should teach Go by using Go naturally.

---

# 67. Open Questions

These are intentionally not locked by this document:

1. Exact public SDK API.
2. Final event payload generic design.
3. Exact event backpressure algorithm.
4. Whether SQLite is eventually required.
5. Whether providers become separate Go modules.
6. Exact update UX.
7. Linux fallback UX for transparent overlays.
8. Exact overlay animation system.
9. Whether YouTube is the second provider.
10. Whether a separate CLI ships publicly.

These questions do not block the first vertical slice.

---

# 68. Accepted Technical Decisions

The following are now considered accepted unless new evidence invalidates them:

```text
ADR: Wails v3 desktop runtime
ADR: Go owns SideLive core/runtime
ADR: React + TypeScript frontend
ADR: Vite frontend build
ADR: Tailwind styling
ADR: Zustand only for frontend presentation state
ADR: Wails binding/event bridge for desktop IPC
ADR: no local REST/WebSocket IPC layer
ADR: JSON configuration before database
ADR: slog logging
ADR: Biome frontend lint/format
ADR: Go standard testing + Vitest frontend testing
ADR: fake provider required
ADR: TikTok library hidden behind provider adapter
ADR: PirateTok/live-go is the first spike implementation
ADR: Electron excluded
```

These should later be split into individual ADR files when implementation begins.

---

# 69. Research Basis

This decision was based on current public documentation and repositories checked on 2026-09-03.

Primary sources:

- Wails v3 documentation  
  https://v3.wails.io/

- Wails v3 installation and platform requirements  
  https://v3.wails.io/getting-started/installation/

- Wails v3 window options  
  https://v3.wails.io/features/windows/options/

- Wails v3 frameless/overlay windows  
  https://v3.wails.io/features/windows/frameless/

- Wails v3 Go ↔ JavaScript method bindings  
  https://v3.wails.io/features/bindings/methods/

- Wails v3 FAQ / packaging / updater  
  https://v3.wails.io/faq/

- Wails v3 changelog  
  https://v3.wails.io/changelog/

- Go release history  
  https://go.dev/doc/devel/release

- React versions  
  https://react.dev/versions

- Vite 8 / 8.1 announcements  
  https://vite.dev/blog/announcing-vite8  
  https://vite.dev/blog/announcing-vite8-1

- TypeScript official blog  
  https://devblogs.microsoft.com/typescript/

- Biome releases  
  https://biomejs.dev/blog/

- Vitest releases  
  https://vitest.dev/blog

- PirateTok Go LIVE connector  
  https://github.com/PirateTok/live-go

- Davincible GoTikTokLive  
  https://github.com/Davincible/gotiktoklive

- steampoweredtaco GoTikTokLive fork  
  https://github.com/steampoweredtaco/gotiktoklive

---

# 70. Final Technical Statement

SideLive will be built as a lightweight local-first desktop application where:

```text
Go
owns the domain and realtime runtime.

Wails v3
owns the native desktop integration.

React + TypeScript
own presentation.

Providers
are replaceable adapters.

Wails bindings
are private desktop IPC.

The SideLive SDK
is a separate public contract.
```

The first implementation must prove the following before broad feature development:

> A Go TikTok provider can receive and normalize LIVE comments, route them through the SideLive core, and render them in a lightweight Wails floating overlay without compromising process stability or platform isolation.

That vertical slice is the technical foundation of SideLive.
