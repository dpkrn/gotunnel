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

## Project Structure

```
gotunnel/
├── cmd/
│   └── client/
│       └── main.go         # mytunnel CLI entry point
├── pkg/
│   └── tunnel/
│       └── tunnel.go       # Public library API — StartTunnel()
├── internal/
│   ├── tunnel/
│   │   └── tunnel.go       # Core TCP + yamux tunnel logic
│   └── models/
│       └── protocol/
│           ├── request.go  # Wire format: TunnelRequest
│           └── response.go # Wire format: TunnelResponse
└── install.sh              # One-liner installer script
```

---

## Requirements

- Go 1.24+
- A running gotunnel server reachable at `localhost:9000`

---

## License

MIT
