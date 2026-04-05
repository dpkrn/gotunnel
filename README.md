# gotunnel

Expose your local HTTP server to the public internet — use it as a **CLI tool** or embed it directly in your **Go application** as a library.

```
Internet ──► Tunnel Server ──► yamux stream ──► gotunnel ──► localhost:<port>
```

---

## CLI — `mytunnel`

The `mytunnel` binary lets you expose any local port with a single command.

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/DpkRn/devtunnel/master/install.sh | bash
```

Auto-detects your OS and CPU architecture (macOS Apple Silicon, macOS Intel, Linux x86\_64) and installs to `/usr/local/bin`.

**Or install the CLI with Go** (separate module path so it does not clutter the library on [pkg.go.dev](https://pkg.go.dev/github.com/DpkRn/gotunnel)):

```bash
go install github.com/DpkRn/gotunnel/mytunnel@latest
```

From a clone, use the repo [go.work](go.work) and build:

```bash
go build -C mytunnel -o mytunnel .
```

### Usage

```bash
mytunnel http <port>
```

**Example — expose a React dev server running on port 3000:**

```
$ mytunnel http 3000

  ╔══════════════════════════════════════════════════╗
  ║   🚇  mytunnel — tunnel is live                  ║
  ╠══════════════════════════════════════════════════╣
  ║  🌍  Public   →  http://abc123.example.com       ║
  ║  💻  Local    →  http://localhost:3000            ║
  ╠══════════════════════════════════════════════════╣
  ║  ⚡  Forwarding requests...                      ║
  ║  🛑  Press Ctrl+C to stop                        ║
  ╚══════════════════════════════════════════════════╝
```

Press `Ctrl+C` to stop the tunnel.

### Commands

| Command | Description |
|---------|-------------|
| `mytunnel http <port>` | Forward public HTTP traffic to `localhost:<port>` |
| `mytunnel help` | Show help |

---

## Go Library

Embed the tunnel directly in your Go application — no separate process needed.

### Install

```bash
go get github.com/DpkRn/gotunnel
```

### Quick Start

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/DpkRn/gotunnel/pkg/tunnel"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from my local server!"))
    })

    url, stop, err := tunnel.StartTunnel("8080")
    if err != nil {
        log.Fatal("tunnel error:", err)
    }
    defer stop()

    fmt.Println("Public URL:", url)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

The tunnel runs in the background alongside your server. `stop()` closes the connection cleanly — it is safe to `defer` it.

### API

#### `tunnel.StartTunnel(port string) (url string, stop func(), err error)`

| Return | Type | Description |
|--------|------|-------------|
| `url` | `string` | Public URL assigned by the tunnel server, e.g. `http://abc123.example.com` |
| `stop` | `func()` | Closes the tunnel and releases all resources |
| `err` | `error` | Non-nil if the tunnel could not be established |

### Examples

**Graceful shutdown with OS signals**

```go
url, stop, err := tunnel.StartTunnel("8080")
if err != nil {
    log.Fatal(err)
}
defer stop()

fmt.Println("Public URL:", url)

quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

go log.Fatal(http.ListenAndServe(":8080", nil))

<-quit
fmt.Println("Shutting down...")
```

**Testing webhooks locally**

Register the printed URL with Stripe, GitHub, or any webhook provider:

```go
http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    fmt.Printf("Received: %s\n", body)
    w.WriteHeader(http.StatusOK)
})

url, stop, err := tunnel.StartTunnel("4000")
if err != nil {
    log.Fatal(err)
}
defer stop()

fmt.Println("Webhook URL:", url+"/webhook")
log.Fatal(http.ListenAndServe(":4000", nil))
```

---

## How It Works

1. `StartTunnel` dials the gotunnel server over TCP and negotiates a [yamux](https://github.com/hashicorp/yamux) multiplexed session.
2. The server assigns a public URL and sends it back during the handshake.
3. Each public HTTP request arrives as a new yamux stream.
4. The client decodes the request, forwards it to `localhost:<port>`, and streams the response back.

---



### pkg.go.dev layout

The module **[github.com/DpkRn/gotunnel](https://pkg.go.dev/github.com/DpkRn/gotunnel)** ships only **`pkg/`** (the library). The **`mytunnel`** binary is module **[github.com/DpkRn/gotunnel/mytunnel](https://pkg.go.dev/github.com/DpkRn/gotunnel/mytunnel)** so the library index stays free of old `cmd/` / `internal/` trees.

## Requirements

- Go 1.25+
- A running gotunnel server (the client dials **`clickly.cv:9000`** by default; change `defaultControlAddr` in `pkg/tunnel/client.go` for your own server)

---

## License

MIT
