# Changelog

All notable changes to this project are recorded here. Summaries are derived from [`git log`](https://git-scm.com/docs/git-log) (newest first within each section).

## [Unreleased]

### Changed

- **2026-04-19** — Cleared extra libraries: dropped Gin from `test-server` (stdlib `net/http` + `encoding/json`); `go.mod` / `go.sum` reduced to core deps (`google/uuid`, `gorilla/websocket`, `hashicorp/yamux`). (`6fe49be`)
- **2026-04-19** — Structured / refactored tunnel and related code. (`99e74aa`)

### Added

- **2026-04-18** — Inspector: replay logging, history controls (sort, filter), optional `X-Inspector-Log-Replay` for persisting replays, tunnel forward port wiring, ingest fixes (sync `net.Listen` before ingest dial). (`a28c393`)
- **2026-04-18** — Inspector: configurable listen port; JavaScript helper to spawn the inspector binary (`js/spawnInspector.mjs`). (`cbb1ebb`)
- **2026-04-15** — Inspector UI: traffic viewer, request replay, Postman / Terminal themes. (`fc3c877`)
- **2026-04-15** — Tunnel: embed inspector, WebSocket ingest for live request logs, `/logs` / `/ws` integration. (`dfece9e`, `c15b3ab`)

### Earlier (April 2026)

- **2026-04-15** — Windows binary packaging; clearer tunnel success output. (`66273d2`)
- **2026-04-11** — Client hello and connection ID for tunnel control plane. (`3c37728`)
- **2026-04-05** — Release notes for v0.4.5; success banner moved into client; `mytunnel` CLI and build/docs updates; package docs under `pkg/`. (`2591316`, `12e21a9`, `137bab7`, `ce520c6`, `27dcb38`)

---

## Tags

Published versions include `v0.3.x`–`v0.4.6` and `v1.0.1`–`v1.0.11` (see `git tag -l`). This changelog focuses on recent commits; older releases may be documented in git history and `README.md`.

To regenerate a raw list from git:

```bash
git log --oneline --decorate -50
```
