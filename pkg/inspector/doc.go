// Package inspector is a standalone HTTP/WebSocket traffic viewer.
//
// # Go (gotunnel / pkg/tunnel)
//
// tunnel.StartTunnel starts the inspector in-process by default (see StartInspector) so nothing
// runs manually; it passes your local app port into the UI so “Replay base” defaults to
// http://localhost:<that port> (traffic forwarding target), not the inspector’s listen port.
// Use tunnel.WithEmbeddedInspector(false) if you prefer a separate process.
//
// # Other languages (e.g. nodetunnel)
//
// Node cannot embed this Go package. Run the same inspector as a subprocess: build cmd/inspector
// and spawn it (see js/spawnInspector.mjs in the repo), or connect to an already-running server.
// Tunnel clients only need the ingest WebSocket URL and JSON message shapes (see below).
//
// Default when listening on port 4040:
//
//	Ingest WebSocket: ws://127.0.0.1:4040/ingest
//	UI (browser):     http://127.0.0.1:4040/
//
// Build the ingest URL from your listen address: see [IngestWebSocketURL] in Go, or use
// ws://<host>:<port>/ingest with the same host/port you pass to Listen.
//
// # Ingest WebSocket (GET /ingest → upgrade)
//
// Send one JSON-encoded request/response record per WebSocket text message. The JSON matches
// [logstore.RequestEvent]: id, request, response, durationMs. Nested [logstore.Request] and
// [logstore.Response] use json tags; body fields are []byte and encode as base64 strings in JSON
// (standard encoding/json behavior). Headers are objects mapping string → array of strings.
//
// Example (fields abbreviated):
//
//	{"id":"…","request":{"method":"GET","path":"/","body":null,"headers":{…}},"response":{"durationMs":0,"statusCode":200,"headers":{…},"body":…},"durationMs":12}
//
// # Viewer WebSocket (GET /ws → upgrade)
//
// The UI subscribes here; the server pushes JSON envelopes:
//
//	{"eventType":"request","payload":<RequestEvent>}
//
// # HTTP API (for the bundled UI)
//
// GET /logs — JSON array of stored events. GET /log?id=… — single event. POST /replay — replay
// helper (localhost only). Static assets: /inspector.css, /theme-*.css, /index.js.
//
package inspector
