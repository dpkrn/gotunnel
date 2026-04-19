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
All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

Each version section lists commits between consecutive git tags (newest releases first). The [Complete commit log](#complete-commit-log) table lists every commit in the repository (newest first).

## [Unreleased]

- refactor: standardize naming conventions by renaming RequestLog to requestLog and updating related logging functions for consistency across the tunnel package (`3a88486`)

## [v1.0.10] - 2026-04-15

- minor fix (`ab1cd5a`)

## [v1.0.9] - 2026-04-15

- fix: update install.sh to use temporary paths for installation files based on OS, improving compatibility when running from a repo clone (`5b5c99f`)
- Merge pull request #1 from dpkrn/f/inspector-themes (`e6f9482`)
- docs: expand tunnel package documentation with detailed usage examples, benefits, and requirements for better user guidance (`a6758ac`)

## [v1.0.8] - 2026-04-15

- feat: add features.md for comprehensive documentation on YourTool capabilities, implement request logging and replay functionality, and introduce a traffic inspector for real-time monitoring (`16e389b`)
- refactor: rename RequestLog to requestLog for consistency, implement theme support in traffic inspector, and enhance TunnelOptions for better configuration management (`f1a5059`)
- docs: enhance README and documentation with traffic inspector details, including usage instructions and customization options (`f52d515`)
- fix: ensure public URL is printed upon successful TCP session establishment in StartTunnel function (`f795b78`)
- feat: enhance tunnel output by adding inspector URL to success message and improve shutdown handling with SIGTERM (`1d40164`)
- refactor: update RequestLog type and logging mechanism, implement log subscriber functionality, and clean up inspector handling (`129396f`)

## [v1.0.7] - 2026-04-12

- chore: update .gitignore to include cross-built mytunnel binaries, enhance documentation in doc.md for build instructions, and improve install.sh for better OS detection and installation process (`0466104`)
- fix: update install.sh to use temporary paths for installation files based on OS, improving compatibility when running from a repo clone (`5b5c99f`)

## [v1.0.6] - 2026-04-11

- added make file (`a67d5ed`)

## [v1.0.5] - 2026-04-11

- make file changes (`f9d1b0f`)
- fix: improve output formatting in Makefile for version tagging and pushing (`9c4a941`)

## [v1.0.4] - 2026-04-11

- minor fix in indexes and types (`bff21f1`)

## [v1.0.3] - 2026-04-11

- chore: update .gitignore to include .commands.md, add commands.md for documentation, and enhance README.md with CLI usage instructions (`cee0400`)

## [v1.0.2] - 2026-04-11

- feat: add Makefile for packaging and tagging releases, enhance documentation in pkg/doc.go and pkg/tunnel/tunnel.go (`20a8001`)

## [v1.0.1] - 2026-04-11

- chore: add google/uuid dependency to go.mod and update go.sum (`fd57156`)
- chore: update .gitignore to include .doc.md and fix README.md server address (`2764379`)

## [v0.4.6] - 2026-04-11

- chore: add release instructions for v0.4.5 and update binary files (`2591316`)
- feat: implement client hello message and connection ID generation (`3c37728`)
- chore: update binary files for Linux, macOS, and ARM64 platforms (`235ae77`)

## [v0.4.5] - 2026-04-05

- refactor: streamline tunnel output by moving success message to client package (`12e21a9`)

## [v0.4.4] - 2026-04-05

- added new release (`4f84eb5`)

## [v0.4.3] - 2026-04-05

- fixed pkg import path (`4164f26`)

## [v0.4.2] - 2026-04-05

- minor ifx (`7d10107`)

## [v0.4.1] - 2026-04-05

- added doc (`4baf966`)

## [v0.4.0] - 2026-04-05

- Update build instructions and Go version; remove unused files and add mytunnel CLI (`137bab7`)

## [v0.3.6] - 2026-04-05

- cleareed doc (`a024c7e`)

## [v0.3.5] - 2026-04-05

- added doc in tunnel also (`ddc0ecb`)

## [v0.3.4] - 2026-04-05

- added doc (`fe17aac`)

## [v0.3.3] - 2026-04-05

- minor fix (`2b9c836`)

## [v0.3.2] - 2026-04-03

- docs: add usage examples and installation instructions for tunnel library (`ce520c6`)

## [v0.3.1] - 2026-04-03

- docs: move package docs to pkg (`27dcb38`)

## [v0.3.0] - 2026-04-03

- remove tunnel.go file and its associated functionality (`23ddb5a`)
- remove build instructions for source installation from README.md (`cf61c3b`)

## [v0.0.2] - 2026-04-03

- Add .gitignore and update documentation for mytunnel CLI (`9f62e26`)
- minor fix to publish (`504ed70`)
- fix module + add package (`6491e2e`)
- add MIT license (`ac9b38c`)
- license added (`64cf4c6`)

## [v0.0.1] - 2026-04-02

- added tunnel lib initial commit (`b1719b3`)
- added tunnel lib required code (`8448b14`)
- added desc (`84d25b4`)
- added install.sh (`caac897`)

---

## Complete commit log

| Date | Commit | Subject |
|------|--------|---------|
| 2026-04-15 | `3a88486` | refactor: standardize naming conventions by renaming RequestLog to requestLog and updating related logging functions for consistency across the tunnel package |
| 2026-04-15 | `ab1cd5a` | minor fix |
| 2026-04-15 | `a6758ac` | docs: expand tunnel package documentation with detailed usage examples, benefits, and requirements for better user guidance |
| 2026-04-15 | `e6f9482` | Merge pull request #1 from dpkrn/f/inspector-themes |
| 2026-04-15 | `129396f` | refactor: update RequestLog type and logging mechanism, implement log subscriber functionality, and clean up inspector handling |
| 2026-04-12 | `1d40164` | feat: enhance tunnel output by adding inspector URL to success message and improve shutdown handling with SIGTERM |
| 2026-04-12 | `f795b78` | fix: ensure public URL is printed upon successful TCP session establishment in StartTunnel function |
| 2026-04-12 | `f52d515` | docs: enhance README and documentation with traffic inspector details, including usage instructions and customization options |
| 2026-04-12 | `5b5c99f` | fix: update install.sh to use temporary paths for installation files based on OS, improving compatibility when running from a repo clone |
| 2026-04-12 | `f1a5059` | refactor: rename RequestLog to requestLog for consistency, implement theme support in traffic inspector, and enhance TunnelOptions for better configuration management |
| 2026-04-12 | `16e389b` | feat: add features.md for comprehensive documentation on YourTool capabilities, implement request logging and replay functionality, and introduce a traffic inspector for real-time monitoring |
| 2026-04-11 | `0466104` | chore: update .gitignore to include cross-built mytunnel binaries, enhance documentation in doc.md for build instructions, and improve install.sh for better OS detection and installation process |
| 2026-04-11 | `a67d5ed` | added make file |
| 2026-04-11 | `9c4a941` | fix: improve output formatting in Makefile for version tagging and pushing |
| 2026-04-11 | `f9d1b0f` | make file changes |
| 2026-04-11 | `bff21f1` | minor fix in indexes and types |
| 2026-04-11 | `cee0400` | chore: update .gitignore to include .commands.md, add commands.md for documentation, and enhance README.md with CLI usage instructions |
| 2026-04-11 | `20a8001` | feat: add Makefile for packaging and tagging releases, enhance documentation in pkg/doc.go and pkg/tunnel/tunnel.go |
| 2026-04-11 | `2764379` | chore: update .gitignore to include .doc.md and fix README.md server address |
| 2026-04-11 | `fd57156` | chore: add google/uuid dependency to go.mod and update go.sum |
| 2026-04-11 | `235ae77` | chore: update binary files for Linux, macOS, and ARM64 platforms |
| 2026-04-11 | `3c37728` | feat: implement client hello message and connection ID generation |
| 2026-04-05 | `2591316` | chore: add release instructions for v0.4.5 and update binary files |
| 2026-04-05 | `12e21a9` | refactor: streamline tunnel output by moving success message to client package |
| 2026-04-05 | `4f84eb5` | added new release |
| 2026-04-05 | `4164f26` | fixed pkg import path |
| 2026-04-05 | `7d10107` | minor ifx |
| 2026-04-05 | `4baf966` | added doc |
| 2026-04-05 | `137bab7` | Update build instructions and Go version; remove unused files and add mytunnel CLI |
| 2026-04-05 | `a024c7e` | cleareed doc |
| 2026-04-05 | `ddc0ecb` | added doc in tunnel also |
| 2026-04-05 | `fe17aac` | added doc |
| 2026-04-05 | `2b9c836` | minor fix |
| 2026-04-03 | `ce520c6` | docs: add usage examples and installation instructions for tunnel library |
| 2026-04-03 | `27dcb38` | docs: move package docs to pkg |
| 2026-04-03 | `cf61c3b` | remove build instructions for source installation from README.md |
| 2026-04-03 | `23ddb5a` | remove tunnel.go file and its associated functionality |
| 2026-04-03 | `64cf4c6` | license added |
| 2026-04-02 | `ac9b38c` | add MIT license |
| 2026-04-02 | `6491e2e` | fix module + add package |
| 2026-04-02 | `504ed70` | minor fix to publish |
| 2026-04-02 | `9f62e26` | Add .gitignore and update documentation for mytunnel CLI |
| 2026-04-02 | `caac897` | added install.sh |
| 2026-04-02 | `84d25b4` | added desc |
| 2026-04-02 | `8448b14` | added tunnel lib required code |
| 2026-03-28 | `b1719b3` | added tunnel lib initial commit |

## Maintenance

After tagging a release, move items from **[Unreleased]** into a new `## [vX.Y.Z]` section and append new rows to the **Complete commit log** table (or sync with `git log`).

Release tags follow the `v*` pattern (for example `v1.0.10`). There is also a legacy tag `gotunnel` on an older commit.
