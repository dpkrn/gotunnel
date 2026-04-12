# Shared tunnel inspector

This folder defines the **cross-language** story for the traffic inspector: one **browser** UI, one **protocol** ([`PROTOCOL.md`](./PROTOCOL.md)), multiple tunnel implementations.

## Goals

| Goal | How |
|------|-----|
| UI runs in the **browser** (client) | Ship the SPA as static files and/or an npm package; no UI logic required inside Go/Node except optional static hosting. |
| **localhost:4040** (default) | Inspector **runtime** listens here for API + WebSocket; same port in every implementation by default. |
| **Independent of tunnel** | Runtime only implements: ring buffer of logs, WS broadcast, replay to local app. Tunnel transport is irrelevant. |
| **gotunnel + nodetunnel** | Both implement the same [`PROTOCOL.md`](./PROTOCOL.md). |
| **Plug-and-play import** | **Node:** e.g. `npx @scope/tunnel-inspector` or `import` from an npm package that opens the UI with `INSPECTOR_BASE_URL` defaulting to `http://127.0.0.1:4040`. **Go:** optional `embed` of the same `web/` assets, or document “install from npm; open URL”. |

## Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│  Browser (inspector UI — static SPA)                         │
│  fetch /logs · WebSocket /ws · POST /replay                   │
└───────────────────────────────┬─────────────────────────────┘
                                │ HTTP / WS
                                ▼
┌─────────────────────────────────────────────────────────────┐
│  Inspector runtime (localhost:4040)                          │
│  — append log · GET /logs · WS push · POST /replay → local app│
└───────────────────────────────┬─────────────────────────────┘
                                │ tunnel-specific
                                ▼
┌─────────────────────────────────────────────────────────────┐
│  Tunnel client (gotunnel / nodetunnel) … public URL, etc.   │
└─────────────────────────────────────────────────────────────┘
```

The **runtime** is the only part that must match the protocol. The **tunnel** only feeds it captured `RequestLog` objects.

## Repository layout (target)

```text
inspector/
  PROTOCOL.md       # versioned API (source of truth)
  README.md         # this file
  web/              # (future) SPA: index.html, JS, CSS — built once, shared
```

Today, gotunnel still embeds HTML/CSS/JS in `pkg/tunnel/inspector.go` for convenience. The migration path is:

1. **Freeze** behavior against [`PROTOCOL.md`](./PROTOCOL.md).
2. **Extract** the embedded document into `inspector/web/` (same behavior, relative URLs).
3. **Go:** `//go:embed web/*` and serve `GET /` from that FS, or serve API-only and point users to the npm-hosted UI.
4. **Node:** implement the same routes and publish the same `web/` bundle alongside `nodetunnel`.

## Configuration (conceptual)

| Variable / option | Meaning |
|-------------------|---------|
| `INSPECTOR_ADDR` / `InspectorAddr` | Listen address for runtime (e.g. `:4040`). |
| `INSPECTOR_BASE_URL` (UI) | Where the browser should call APIs (default `http://127.0.0.1:4040`). |
| Theme | UI-only; can be query param, `localStorage`, or build flag — not part of protocol v1. |

## Security note

The inspector runtime is intended for **local development**. It exposes captured traffic and can replay arbitrary requests to your local app. Bind to **loopback** only unless you understand the risk.
