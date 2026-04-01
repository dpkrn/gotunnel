# gotunnel

A lightweight Go library that exposes your local HTTP server to the public internet via a secure tunnel — similar to ngrok, but embeddable directly in your Go application.

## How it works

`gotunnel` connects your local service to a remote tunnel server over TCP. Incoming public requests are multiplexed over a single connection using [yamux](https://github.com/hashicorp/yamux) and forwarded to your local HTTP server. Your app gets a public URL it can share immediately.

```
Internet → Tunnel Server (:9000) → yamux stream → gotunnel client → localhost:<port>
```

## Requirements

- Go 1.21+
- A running `gotunnel` server accessible at `localhost:9000` (or configure your own)

## Installation

```bash
go get github.com/DpkRn/gotunnel
```

## Quick Start

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

    // Start the tunnel on the same port your server will listen on
    url, stop, err := tunnel.StartTunnel("8080")
    if err != nil {
        log.Fatal("tunnel error:", err)
    }
    defer stop()

    fmt.Println("Public URL:", url) // e.g. http://abc123.yourtunnelserver.com

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

That's it. The tunnel runs in the background alongside your HTTP server. When your program exits, `stop()` cleans up the connection.

## API

### `tunnel.StartTunnel(port string) (url string, stop func(), err error)`

Connects to the tunnel server, registers your local port, and returns:

| Return | Type | Description |
|--------|------|-------------|
| `url` | `string` | The public URL assigned to your tunnel (e.g. `http://xyz.example.com`) |
| `stop` | `func()` | Call this to close the tunnel and release the connection |
| `err` | `error` | Non-nil if the tunnel could not be established |

## Usage Patterns

### With `net/http`

```go
url, stop, err := tunnel.StartTunnel("3000")
if err != nil {
    log.Fatal(err)
}
defer stop()

fmt.Println("Share this URL:", url)

mux := http.NewServeMux()
mux.HandleFunc("/webhook", handleWebhook)
http.ListenAndServe(":3000", mux)
```

### Graceful shutdown with OS signals

```go
url, stop, err := tunnel.StartTunnel("8080")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Public URL:", url)

quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

go http.ListenAndServe(":8080", nil)

<-quit
fmt.Println("Shutting down...")
stop()
```

### Testing webhooks locally

```go
func main() {
    http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        fmt.Printf("Received webhook: %s\n", body)
        w.WriteHeader(http.StatusOK)
    })

    url, stop, err := tunnel.StartTunnel("4000")
    if err != nil {
        log.Fatal(err)
    }
    defer stop()

    // Register this URL with your webhook provider (Stripe, GitHub, etc.)
    fmt.Println("Webhook URL:", url+"/webhook")

    log.Fatal(http.ListenAndServe(":4000", nil))
}
```

## Project Structure

```
gotunnel/
├── pkg/
│   └── tunnel/
│       └── tunnel.go       # Public API — StartTunnel()
├── internal/
│   ├── tunnel/
│   │   └── tunnel.go       # Core TCP/yamux tunnel logic
│   └── models/
│       └── protocol/
│           ├── request.go  # TunnelRequest wire format
│           └── response.go # TunnelResponse wire format
└── cmd/
    └── test.go             # Example usage
```

## License

MIT
