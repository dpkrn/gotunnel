# gotunnel architecture

This document describes how `pkg/tunnel` is organized, how traffic flows through the system, and where to extend behavior safely.

## Package layout (`pkg/tunnel`)

| File | Responsibility |
|------|----------------|
| `tunnel.go` | Public API: [StartTunnel](../pkg/tunnel/tunnel.go) orchestration (dial, inspector, forwarder). |
| `options.go` | [TunnelOptions](../pkg/tunnel/models.go), defaults, option merging. |
| `client.go` | Control-plane TCP dial, Yamux session, per-stream HTTP proxy to localhost. |
| `wire.go` | JSON wire types for one stream request/response line. |
| `models.go` | Domain types: [RequestLog](../pkg/tunnel/models.go), [clientHello](../pkg/tunnel/models.go), themes. |
| `logstore.go` | In-memory ring buffer of [RequestLog](../pkg/tunnel/models.go), [AddLog](../pkg/tunnel/logstore.go) / [GetLogs](../pkg/tunnel/logstore.go), **pluggable subscribers** for live updates. |
| `inspector.go` | Loopback HTTP server: UI (embedded HTML), `GET /logs`, `GET /ws`, `POST /replay`. |
| `replay.go` | Replay handler: POST JSON → local app. |
| `utils.go` | IDs (connection, request). |
| `inspector_page.html` | Embedded inspector UI (via `go:embed`). |

**Dependency direction (high level):** `tunnel.go` → `client` + `inspector` + `logstore`; `client` → `logstore` (after each proxied request); `inspector` registers a log subscriber on `logstore`.

## Component diagram

```mermaid
flowchart TB
  subgraph UserApp["User process"]
    LocalHTTP["Local HTTP server\nlocalhost:PORT"]
  end

  subgraph GotunnelPkg["pkg/tunnel"]
    StartTunnel["StartTunnel"]
    Client["clientConn\nYamux + TCP"]
    LogStore["logstore\nring buffer + subscribers"]
    Inspector["inspector HTTP\n:4040"]
    WS["WebSocket hub"]
  end

  subgraph Remote["Tunnel server"]
    PublicURL["Public URL"]
    Edge["Ingress / stream open"]
  end

  Browser["Browser\ninspector UI"]

  StartTunnel --> Client
  StartTunnel --> Inspector
  Inspector --> WS
  Inspector --> LogStore
  Client -->|"AddLog"| LogStore
  LogStore -->|"notify subscribers"| WS
  Browser <-->|"HTTP + WS"| Inspector

  PublicURL --> Edge
  Edge -->|"JSON tunnelRequest"| Client
  Client -->|"HTTP"| LocalHTTP
  LocalHTTP -->|"response"| Client
  Client -->|"JSON tunnelResponse"| Edge
```

## Sequence: public request → local app

```mermaid
sequenceDiagram
  participant Internet
  participant TunnelSrv as Tunnel server
  participant Yamux as Yamux stream
  participant Handle as handleStream
  participant Local as localhost app
  participant Log as logstore

  Internet->>TunnelSrv: HTTP request
  TunnelSrv->>Yamux: open stream + JSON tunnelRequest
  Yamux->>Handle: read line, unmarshal
  Handle->>Local: http.Client Do
  Local-->>Handle: HTTP response
  Handle->>Yamux: JSON tunnelResponse + newline
  Handle->>Log: AddLog(RequestLog)
  Note over Log: ring buffer + notify subscribers async
```

## Sequence: inspector live updates

```mermaid
sequenceDiagram
  participant UI as Browser
  participant Insp as inspector HTTP
  participant Log as logstore
  participant Hub as WebSocket hub

  UI->>Insp: GET /ws
  Insp->>Hub: register connection
  Note over Log,Hub: On AddLog, notifyLogSubscribers runs hub.broadcast in a goroutine
  Log->>Hub: subscriber(entry)
  Hub-->>UI: JSON RequestLog
```

## Extension points (scalability)

1. **More live consumers of traffic** — Call [RegisterLogSubscriber](../pkg/tunnel/logstore.go) from application or library code to receive each [RequestLog](../pkg/tunnel/models.go) without modifying the inspector.
2. **Replace in-memory logs** — Introduce a `LogStore` interface and inject an implementation; keep [GetLogs](../pkg/tunnel/logstore.go) / [AddLog](../pkg/tunnel/logstore.go) as adapters during migration.
3. **Control plane** — [dialClient](../pkg/tunnel/client.go) and [defaultControlAddr](../pkg/tunnel/client.go) are natural places for TLS, auth, or configurable server addresses.
4. **Protocol** — [wire.go](../pkg/tunnel/wire.go) centralizes JSON shapes for the Yamux line protocol.

## Lifecycle: StartTunnel / stop

```mermaid
stateDiagram-v2
  [*] --> Dial: StartTunnel(port)
  Dial --> Running: dial OK + Yamux session
  Running --> Running: Accept streams, proxy, AddLog
  Running --> Stopped: stop()
  Stopped --> [*]: Shutdown inspector HTTP, unsub WebSocket, close Yamux/TCP
```
