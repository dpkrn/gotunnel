# Tunnel inspector protocol

Single contract for **any** tunnel runtime (gotunnel, nodetunnel, …) so one **browser** UI can attach without knowing the implementation language.

**Protocol version:** `1` (bump when breaking JSON shapes or routes; keep backward-compatible additions when possible.)

---

## Roles

| Piece | Responsibility |
|--------|----------------|
| **Inspector UI** | Static web app (HTML/JS/CSS). Runs in the browser. Connects to the runtime only via HTTP + WebSocket below. |
| **Inspector runtime** | Small HTTP server bound to localhost (default **`:4040`**). Stores recent captures, streams new ones over WebSocket, executes replay against the **local app** (e.g. `127.0.0.1:<port>`). No requirement to serve the UI HTML (optional). |

The tunnel client (Go/Node) **implements the runtime**. The UI may be shipped as an npm static package, embedded assets, or opened from a dev server.

---

## Base URL

Configurable. Default:

```text
http://127.0.0.1:4040
```

All paths below are relative to this origin. The UI should read the base URL from config (env, query string, or build-time constant).

---

## HTTP

### `GET /logs`

Returns a JSON **array** of captured request/response records (order: implementation-defined; UI may sort by `timestamp` / `id`).

**Response:** `200`  
**Content-Type:** `application/json`

### `POST /replay`

Replays a synthetic request against the **local application** the tunnel forwards to (same host the tunnel targets, e.g. `http://127.0.0.1:<localPort>`).

**Request Content-Type:** `application/json`

**Body:**

```json
{
  "method": "GET",
  "path": "/api/foo",
  "headers": { "Accept": ["application/json"] },
  "body": ""
}
```

`headers` is optional; values are **arrays of strings** (Go `http.Header` style).

**Response:** `200`  
**Content-Type:** `application/json`

**Success body:**

```json
{
  "status": 200,
  "headers": { "Content-Type": ["application/json"] },
  "body": "…"
}
```

**Error body** (still JSON, often with non-2xx status):

```json
{
  "error": "human-readable message"
}
```

Other methods on `/replay` should return `405` with JSON `{ "error": "…" }` where practical.

---

## WebSocket

### `GET /ws`

- **Direction:** server → client for log events (client may send pings/pongs; server may ignore message frames).
- **Each message** is one JSON object, **same shape as one element of `/logs`** (see [Request log](#request-log)).

Upgrade failures should use normal HTTP error responses.

---

## Request log (JSON object)

Shared by `/logs` entries and each `/ws` message. Field names are stable for protocol v1.

| Field | Type | Notes |
|------|------|--------|
| `id` | string | Unique per capture (client-generated acceptable). |
| `method` | string | HTTP method. |
| `path` | string | Path + query as observed. |
| `headers` | object | Map of string → array of strings. |
| `body` | string | Request body (often empty). |
| `status` | number | Response status from upstream. |
| `resp_body` | string | Response body. |
| `resp_headers` | object | Optional. Map of string → array of strings (upstream response headers). |
| `timestamp` | string | RFC3339 time. |
| `duration_ms` | number | Integer milliseconds. |

Implementations may add extra fields with new names; UIs should ignore unknown fields.

---

## CORS

If the UI is served from a **different origin** than the runtime (e.g. UI on `http://localhost:5173`, API on `http://127.0.0.1:4040`), the runtime must:

- Send appropriate `Access-Control-Allow-Origin` (and related headers) for `GET /logs`, `POST /replay`, and WebSocket upgrade.
- For WebSocket, use a library or framework that allows the upgrade with CORS policy aligned to your threat model (dev: often `*` or `http://localhost:5173`).

For **same-origin** UI (runtime serves `GET /` with static files), CORS is unnecessary.

---

## Optional `GET /`

Serving the inspector SPA from the runtime is **optional**. If the runtime only exposes `/logs`, `/ws`, `/replay`, the UI is loaded from elsewhere (npm package, CDN, separate static server) and must point at the correct **base URL** for API calls.

---

## Version negotiation (optional future)

Clients may send `X-Inspector-Protocol: 1` or `?v=1` on HTTP; WebSocket may use a query param. Servers not supporting the header should ignore it. Bump protocol version when breaking changes occur.

---

## Reference implementation

- **Go:** `gotunnel` — `pkg/tunnel/inspector.go`, `replay.go`, log store + broadcast.
