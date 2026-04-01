# Developer Reference

Build, release, and maintenance commands for the `gotunnel` / `mytunnel` project.

---

## Table of Contents

- [Build](#build)
- [Install Locally](#install-locally)
- [Release](#release)
- [Git](#git)
- [Install Script](#install-script)
- [Debugging](#debugging)

---

## Build

### Standard build

```bash
go build -o mytunnel ./cmd/client
```

### Cross-platform builds

```bash
# macOS — Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o mytunnel-mac-arm64 ./cmd/client

# macOS — Intel
GOOS=darwin GOARCH=amd64 go build -o mytunnel-mac ./cmd/client

# Linux — x86_64
GOOS=linux GOARCH=amd64 go build -o mytunnel-linux ./cmd/client
```

| Variable | Values | Description |
|----------|--------|-------------|
| `GOOS` | `darwin`, `linux`, `windows` | Target operating system |
| `GOARCH` | `arm64`, `amd64` | Target CPU architecture |
| `-o <name>` | any filename | Output binary name |

### Force rebuild (skip cache)

Use `-a` when you want to guarantee a clean compile — useful before cutting a release:

```bash
GOOS=darwin GOARCH=arm64 go build -a -o mytunnel-mac-arm64 ./cmd/client
GOOS=darwin GOARCH=amd64 go build -a -o mytunnel-mac ./cmd/client
GOOS=linux  GOARCH=amd64 go build -a -o mytunnel-linux  ./cmd/client
```

### Full clean rebuild

Wipes the build cache and all downloaded modules — use only when something is deeply broken:

```bash
go clean -cache -modcache -i -r
go build -a -o mytunnel ./cmd/client
```

| Flag | Meaning |
|------|---------|
| `-cache` | Delete the build cache |
| `-modcache` | Delete downloaded module cache (`~/go/pkg/mod`) |
| `-i` | Remove installed packages |
| `-r` | Apply recursively to all dependencies |

---

## Install Locally

Copy a built binary into `$PATH` so you can run `mytunnel` from anywhere:

```bash
sudo cp mytunnel-mac-arm64 /usr/local/bin/mytunnel
sudo chmod +x /usr/local/bin/mytunnel
```

Verify:

```bash
which mytunnel
mytunnel help
```

---

## Release

### Prerequisites

```bash
brew install gh   # GitHub CLI
gh auth login     # authenticate — choose GitHub.com → HTTPS → web browser
```

### Create a new release with binaries

```bash
gh release create v0.2.0 \
  mytunnel-mac \
  mytunnel-mac-arm64 \
  mytunnel-linux \
  --title "v0.2.0" \
  --notes "What changed in this release."
```

### Upload binaries to an existing release

```bash
gh release upload <tag> mytunnel-mac mytunnel-mac-arm64 mytunnel-linux --clobber
```

| Flag | Meaning |
|------|---------|
| `<tag>` | The release tag to upload to (e.g. `v0.2.0`) |
| `--clobber` | Overwrite existing assets with the same name |

### Typical release workflow

```bash
# 1. Build all targets
GOOS=darwin GOARCH=arm64 go build -a -o mytunnel-mac-arm64 ./cmd/client
GOOS=darwin GOARCH=amd64 go build -a -o mytunnel-mac       ./cmd/client
GOOS=linux  GOARCH=amd64 go build -a -o mytunnel-linux     ./cmd/client

# 2. Tag and release
gh release create v0.3.0 \
  mytunnel-mac mytunnel-mac-arm64 mytunnel-linux \
  --title "v0.3.0" \
  --notes "Release notes here."
```

---

## Git

```bash
git status                  # show modified, staged, and untracked files
git add .                   # stage all changes
git commit -m "your message"
git push origin main
git log --oneline           # compact commit history
git log --oneline --graph   # branch graph
```

---

## Install Script

The `install.sh` one-liner for end users:

```bash
curl -fsSL https://raw.githubusercontent.com/DpkRn/devtunnel/master/install.sh | bash
```

| Flag | Meaning |
|------|---------|
| `-f` | Fail silently on HTTP errors (non-zero exit code) |
| `-s` | Silent — no progress output |
| `-S` | Show errors even in silent mode |
| `-L` | Follow redirects |

**What the script does:**

1. Runs `uname` to detect the OS (`Darwin` or `Linux`)
2. Runs `uname -m` to detect CPU architecture (`arm64` or `x86_64`)
3. Downloads the matching binary from the GitHub release
4. Verifies it is a binary (not an HTML error page)
5. Marks it executable and moves it to `/usr/local/bin/mytunnel`

---

## Debugging

### Inspect an installed binary

```bash
file /usr/local/bin/mytunnel   # check architecture and format
uname -m                        # check your Mac's CPU architecture
```

### Manually download and inspect before installing

```bash
curl -fsSL <URL> -o /tmp/test-binary
file /tmp/test-binary
xxd /tmp/test-binary | head -3   # first bytes should be ELF or Mach-O magic
```

### Check Go module state

```bash
go mod tidy        # sync go.sum with actual imports
go mod verify      # verify cached modules match checksums
go build ./...     # confirm everything compiles
```
