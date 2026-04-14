// # gotunnel
//
// Package tunnel exposes a local HTTP server on a public URL by establishing
// a persistent outbound TCP connection to a gotunnel server.
//
// It creates a secure outbound connection to a tunnel server and forwards
// incoming requests to your local application (e.g., localhost:8080).
//
// ## Introduction
//
// ### Benefits
//
// - Sharing your local server with others
//
// - Testing webhooks (Stripe, GitHub, etc.)
//
// - Remote debugging without deployment
//
// - No port forwarding or firewall configuration needed
//
// - Works behind NAT or private networks
//
// - Simple integration with existing Go HTTP servers
//
// - Traffic inspector, replay, modify request unlimited times
//
// Incoming traffic reaches the public URL, is forwarded through the tunnel,
// and is proxied to your local HTTP server (e.g., localhost:8080).
//
// This enables exposing local development servers without port forwarding,
// firewall changes, or public hosting.
//
// ### Requirements
//
// - A gotunnel server must be running and reachable.
//
// - The port passed to [StartTunnel] must match your local HTTP server port.
//
// - Your local server must be running BEFORE or concurrently with StartTunnel.
//
// ## Overview
//
// Package **tunnel** exposes a local HTTP server on a public URL by connecting to a gotunnel server you run separately. Traffic hits the tunnel first, then your app on `localhost`.
//
// it can be used in two way
//
// - Expose your local HTTP server to the public internet — embed it directly in your **Go application** as a library
//
// - use it as a **CLI tool**.
//
// The following mirrors the [`pkg/tunnel`](pkg/tunnel/tunnel.go) package comment (same order).
//
// ### API
//
// The only public entry point is [StartTunnel].
//
// ### Quick Example
//
// Step 1 — local server only (no gotunnel yet)
//
// Run this first using only the standard library. Visit `http://localhost:8080` to confirm the handler works.
//
// ```go
// package main
//
// import (
//
//	"fmt"
//	"log"
//	"net/http"
//
// )
//
//	func main() {
//		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//			fmt.Println("→ request:", r.Method, r.URL.Path)
//			w.WriteHeader(200)
//			w.Write([]byte("hello world"))
//		})
//		log.Fatal(http.ListenAndServe(":8080", nil))
//	}
//
// ```
//
// ## Step 2 — same server, add the tunnel
//
// ### Install
//
//	$ go get github.com/dpkrn/gotunnel
//
// ### Import
//
//	import "github.com/dpkrn/gotunnel/pkg/tunnel"
//
// Add the import, call **`StartTunnel`** with the same port as **`http.ListenAndServe`**, `defer stop()`, and print the public URL before you block in **`ListenAndServe`**.
//
// ```go
// package main
//
// import (
//
//	"fmt"
//	"log"
//	"net/http"
//
//	"github.com/dpkrn/gotunnel/pkg/tunnel"
//
// )
//
//	func main() {
//		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//			fmt.Println("→ request:", r.Method, r.URL.Path)
//			w.WriteHeader(200)
//			w.Write([]byte("hello world"))
//		})
//		url, stop, err := tunnel.StartTunnel("8080")
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer stop()
//		fmt.Println("Public URL:", url)
//		log.Fatal(http.ListenAndServe(":8080", nil))
//	}
//
// ```
//
// ### Traffic inspector
//
// By default, StartTunnel starts a small HTTP server on loopback (see
// [TunnelOptions.InspectorAddr], default ":4040") that serves the traffic
// inspector UI and APIs. Open:
//
//	http://127.0.0.1:4040
//
// You can browse captured requests and responses, and replay requests against your
// local app. Customize appearance with [TunnelOptions.Themes] ("dark", "terminal",
// or "light"), retention with [TunnelOptions.Logs], or the listen address with
// [tunnel.TunnelOptions{}].
//
// ```go
//
//	url, stop, err := tunnel.StartTunnel("8080", tunnel.TunnelOptions{
//	    Inspector: true, //default true
//	    Themes:    "terminal", //default dark
//	    Logs:      100,
//	    InspectorAddr: ":9090", //default 4040
//	})
//
// ```
//
// ### `net/http`
//
// Run **`http.ListenAndServe`** in a goroutine, then **`StartTunnel`** with the same port:
//
// ```go
// mux := http.NewServeMux()
// mux.HandleFunc("/api/", apiHandler)
//
//	go func() {
//		log.Fatal(http.ListenAndServe(":3000", mux))
//	}()
//
// url, stop, err := tunnel.StartTunnel("3000")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer stop()
// ```
//
// ### Gin
//
// Run Gin’s **`Run`** in a goroutine so the tunnel and server both run (add Gin to your `go.mod`):
//
// ```go
// r := gin.Default()
// r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
// go func() { r.Run(":8080") }()
// url, stop, err := tunnel.StartTunnel("8080")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer stop()
// ```
//
// ### Gorilla mux
//
// Pass a **gorilla/mux** `Router` to **`http.ListenAndServe`**:
//
// ```go
// r := mux.NewRouter()
// r.HandleFunc("/", homeHandler)
//
//	go func() {
//		log.Fatal(http.ListenAndServe(":9000", r))
//	}()
//
// url, stop, err := tunnel.StartTunnel("9000")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer stop()
// ```
//
// ### Fiber
//
// Call Fiber’s **`Listen`** in a goroutine with the same port as **`StartTunnel`** (add Fiber to your `go.mod`):
//
// ```go
// app := fiber.New()
// app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
// go func() { log.Fatal(app.Listen(":4000")) }()
// url, stop, err := tunnel.StartTunnel("4000")
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
// defer stop()
// ```
//
// ### Shutdown
//
// Always call the **`stop`** function returned from **`StartTunnel`** on exit (for example after **`os.Signal`** on `SIGINT`) so the tunnel connection closes cleanly.
//
// ---
//
// ## Go library
//
// Embed the tunnel directly in your Go application — no separate process needed.
//
// ### Install
//
//	$ go get github.com/dpkrn/gotunnel
//
// ### Quick Start
//
// ```go
// package main
//
// import (
//
//	"fmt"
//	"log"
//	"net/http"
//
//	"github.com/dpkrn/gotunnel/pkg/tunnel"
//
// )
//
//	func main() {
//	    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	        w.Write([]byte("Hello from my local server!"))
//	    })
//
//	    url, stop, err := tunnel.StartTunnel("8080")
//	    if err != nil {
//	        log.Fatal("tunnel error:", err)
//	    }
//	    defer stop()
//
//	    fmt.Println("Public URL:", url)
//	    log.Fatal(http.ListenAndServe(":8080", nil))
//	}
//
// ```
//
// The tunnel runs in the background alongside your server. `stop()` closes the connection cleanly — it is safe to `defer` it.
//
// ### API
//
// #### `tunnel.StartTunnel(port string) (url string, stop func(), err error)`
//
// | Return | Type | Description |
// |--------|------|-------------|
// | `url` | `string` | Public URL assigned by the tunnel server, e.g. `http://abc123.example.com` |
// | `stop` | `func()` | Closes the tunnel and releases all resources |
// | `err` | `error` | Non-nil if the tunnel could not be established |
//
// ### Examples
//
// **Graceful shutdown with OS signals**
//
// ```go
// url, stop, err := tunnel.StartTunnel("8080")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// defer stop()
//
// fmt.Println("Public URL:", url)
//
// quit := make(chan os.Signal, 1)
// signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
//
// go log.Fatal(http.ListenAndServe(":8080", nil))
//
// <-quit
// fmt.Println("Shutting down...")
// ```
//
// **Testing webhooks locally**
//
// Register the printed URL with Stripe, GitHub, or any webhook provider:
//
// ```go
//
//	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//	    fmt.Printf("Received: %s\n", body)
//	    w.WriteHeader(http.StatusOK)
//	})
//
// url, stop, err := tunnel.StartTunnel("4000")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// defer stop()
//
// fmt.Println("Webhook URL:", url+"/webhook")
// log.Fatal(http.ListenAndServe(":4000", nil))
// ```
//
// Always call the stop function returned by [StartTunnel].
//
// ### Notes
//
// This ensures:
//
//   - TCP connection is closed cleanly
//
//   - tunnel is deregistered on the server
//
//   - resources are released
//
// Example:
//
//	stop()
//
// Or handle signals:
//
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt)
//	<-c
//	stop()
//
// ## Notes
//
//   - Only HTTP traffic is supported currently.
//
//   - Each tunnel maps to a single local port.
//
//   - One client maintains a persistent connection (multiplexed internally).
//
//   - Future versions may support TLS, authentication, and traffic inspection.
//
//   - Advanced but useful features will come soon
//
// ## Troubleshooting
//
//   - Ensure your local server is running before starting the tunnel.
//
//   - Verify the correct port is passed to StartTunnel.
//
//   - Check tunnel server connectivity if no public URL is returned.
//
//   - If requests fail, confirm your local handler responds correctly.
package tunnel

import (
	"fmt"
	"os"
	"sync"
)

// StartTunnel dials the tunnel server, starts forwarding in a background goroutine, and returns
// the public URL, a stop function (safe to defer), and an error if setup failed.
//
// Options are optional: call StartTunnel("8080") to use [DefaultTunnelOptions], or pass a
// [TunnelOptions] value to override only the fields you set (e.g. Themes: "terminal").
func StartTunnel(port string, opts ...TunnelOptions) (url string, stop func(), err error) {
	options := applyTunnelOptions(opts...)
	setMaxrequestLogs(options.Logs)

	c, err := dialClient(port)
	if err != nil {
		return "", noop, fmt.Errorf("could not create tunnel: %w", err)
	}
	// dialClient succeeded: TCP session is up and public URL is known.
	inspURL := inspectorHTTPBaseURL(options)
	printSuccess(c.getPublicURL(), "http://localhost:"+port, inspURL)

	stopInspector := startInspector(options, port)

	go func() {
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "gotunnel: tunnel stopped: %v\n", err)
		}
	}()

	var stopOnce sync.Once
	return c.getPublicURL(), func() {
		stopOnce.Do(func() {
			stopInspector()
			_ = c.Stop()
		})
	}, nil
}

func noop() {}
