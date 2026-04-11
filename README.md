# gotunnel

Package tunnel exposes a local HTTP server on a public URL by establishing
a persistent outbound TCP connection to a gotunnel server.

It creates a secure outbound connection to a tunnel server and forwards
incoming requests to your local application (e.g., localhost:8080).

## Introduction

### Benefits

- Sharing your local server with others
- Testing webhooks (Stripe, GitHub, etc.)
- Remote debugging without deployment
- No port forwarding or firewall configuration needed
- Works behind NAT or private networks
- Simple integration with existing Go HTTP servers

Incoming traffic reaches the public URL, is forwarded through the tunnel,
and is proxied to your local HTTP server (e.g., localhost:8080).

This enables exposing local development servers without port forwarding,
firewall changes, or public hosting.

### Requirements

- A gotunnel server must be running and reachable.
- The port passed to [StartTunnel] must match your local HTTP server port.
- Your local server must be running BEFORE or concurrently with StartTunnel.

## Overview

Package **tunnel** exposes a local HTTP server on a public URL by connecting to a gotunnel server you run separately. Traffic hits the tunnel first, then your app on `localhost`.
it can be used in two way

- Expose your local HTTP server to the public internet — embed it directly in your **Go application** as a library
- use it as a **CLI tool**.

The following mirrors the [`pkg/tunnel`](pkg/tunnel/tunnel.go) package comment (same order).

### API

The only public entry point is [StartTunnel].

### Quick Example

Step 1 — local server only (no gotunnel yet)

Run this first using only the standard library. Visit `http://localhost:8080` to confirm the handler works.

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("→ request:", r.Method, r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte("hello world"))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Step 2 — same server, add the tunnel

### Install

```bash
go get github.com/dpkrn/gotunnel
```

### Import

```go
import "github.com/dpkrn/gotunnel/pkg/tunnel"
```

Add the import, call **`StartTunnel`** with the same port as **`http.ListenAndServe`**, `defer stop()`, and print the public URL before you block in **`ListenAndServe`**.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("→ request:", r.Method, r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte("hello world"))
	})
	url, stop, err := tunnel.StartTunnel("8080")
	if err != nil {
		log.Fatal(err)
	}
	defer stop()
	fmt.Println("Public URL:", url)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### `net/http`

Run **`http.ListenAndServe`** in a goroutine, then **`StartTunnel`** with the same port:

```go
mux := http.NewServeMux()
mux.HandleFunc("/api/", apiHandler)
go func() {
	log.Fatal(http.ListenAndServe(":3000", mux))
}()
url, stop, err := tunnel.StartTunnel("3000")
if err != nil {
	log.Fatal(err)
}
defer stop()
```

### Gin

Run Gin’s **`Run`** in a goroutine so the tunnel and server both run (add Gin to your `go.mod`):

```go
r := gin.Default()
r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
go func() { r.Run(":8080") }()
url, stop, err := tunnel.StartTunnel("8080")
if err != nil {
	log.Fatal(err)
}
defer stop()
```

### Gorilla mux

Pass a **gorilla/mux** `Router` to **`http.ListenAndServe`**:

```go
r := mux.NewRouter()
r.HandleFunc("/", homeHandler)
go func() {
	log.Fatal(http.ListenAndServe(":9000", r))
}()
url, stop, err := tunnel.StartTunnel("9000")
if err != nil {
	log.Fatal(err)
}
defer stop()
```

### Fiber

Call Fiber’s **`Listen`** in a goroutine with the same port as **`StartTunnel`** (add Fiber to your `go.mod`):

```go
app := fiber.New()
app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
go func() { log.Fatal(app.Listen(":4000")) }()
url, stop, err := tunnel.StartTunnel("4000")
if err != nil {
	log.Fatal(err)
}
defer stop()
```

### Shutdown

Always call the **`stop`** function returned from **`StartTunnel`** on exit (for example after **`os.Signal`** on `SIGINT`) so the tunnel connection closes cleanly.

---

## Go library

Embed the tunnel directly in your Go application — no separate process needed.

### Install

```bash
go get github.com/dpkrn/gotunnel
```

### Quick Start

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/dpkrn/gotunnel/pkg/tunnel"
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

## Using CLI (no import needed)

Expose your local HTTP server to the public internet — use it as a **CLI tool**, or embed it directly in your **Go application** as a library.

```
Internet ──► Tunnel Server ──► yamux stream ──► gotunnel ──► localhost:<port>
```

---

The `mytunnel` binary lets you expose any local port with a single command.

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/dpkrn/gotunnel/main/install.sh | bash
```

Auto-detects your OS and CPU architecture (macOS Apple Silicon, macOS Intel, Linux x86\_64) and installs to `/usr/local/bin`.

### Usage

Run your local server.

```bash
mytunnel http <port>
```

*(Note: use the same port your local server is listening on.)*

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

## How it works (high level)

1. Your app starts a local HTTP server.
2. StartTunnel establishes a persistent TCP connection to the tunnel server.
3. The server assigns a public URL.
4. Incoming requests are forwarded over the tunnel to your local server.
5. Responses are sent back through the same tunnel.

## Requirements

- Go 1.25+
- A running gotunnel server (the client dials **`clickly.cv`** by default; change `defaultControlAddr` in `pkg/tunnel/client.go` for your own server)

---

## License

MIT
