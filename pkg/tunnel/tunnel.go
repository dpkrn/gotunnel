// Package tunnel exposes a local HTTP server on a public URL by establishing
// a persistent outbound TCP connection to a gotunnel server.
//
// It creates a secure outbound connection to a tunnel server and forwards
// incoming requests to your local application (e.g., localhost:<port>).
//
// This is useful for:
//   - Sharing your local server with others
//   - Testing webhooks (Stripe, GitHub, etc.)
//   - Remote debugging without deployment
//
// Incoming traffic reaches the public URL, is forwarded through the tunnel,
// and is proxied to your local HTTP server (e.g., localhost:8080).
//
// This enables exposing local development servers without port forwarding,
// firewall changes, or public hosting.
//
// # API
//
// The only public entry point is [StartTunnel].
//
// # Requirements
//
//   - A gotunnel server must be running and reachable.
//   - The port passed to [StartTunnel] must match your local HTTP server port.
//   - Your local server must be running BEFORE or concurrently with StartTunnel.
//
// # Benefits
//
//   - No port forwarding or firewall configuration needed
//   - Works behind NAT or private networks
//   - Simple integration with existing Go HTTP servers
//
// # How it works (high level)
//
//  1. Your app starts a local HTTP server.
//  2. StartTunnel establishes a persistent TCP connection to the tunnel server.
//  3. The server assigns a public URL.
//  4. Incoming requests are forwarded over the tunnel to your local server.
//  5. Responses are sent back through the same tunnel.
//
// # Step 1 — local server only (no tunnel)
//
// Run this first to confirm your server works locally:
//
//	package main
//
//	import (
//	    "fmt"
//	    "log"
//	    "net/http"
//	)
//
//	func main() {
//	    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	        fmt.Println("→ request:", r.Method, r.URL.Path)
//	        w.WriteHeader(200)
//	        w.Write([]byte("hello world"))
//	    })
//	    log.Fatal(http.ListenAndServe(":8080", nil))
//	}
//
// Visit: http://localhost:8080
//
// # Install
//
//	go get github.com/dpkrn/gotunnel
//
// # Import
//
//	import "github.com/dpkrn/gotunnel/pkg/tunnel"
//
// # Step 2 — expose using tunnel
//
// Add StartTunnel with the SAME port:
//
//	package main
//
//	import (
//	    "fmt"
//	    "log"
//	    "net/http"
//
//	    "github.com/dpkrn/gotunnel/pkg/tunnel"
//	)
//
//	func main() {
//	    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	        fmt.Println("→ request:", r.Method, r.URL.Path)
//	        w.WriteHeader(200)
//	        w.Write([]byte("hello world"))
//	    })
//
//	    url, stop, err := tunnel.StartTunnel("8080")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer stop()
//
//	    fmt.Println("Public URL:", url)
//
//	    log.Fatal(http.ListenAndServe(":8080", nil))
//	}
//
// # Framework examples
//
// ## net/http (custom mux)
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/", apiHandler)
//
//	go func() {
//	    log.Fatal(http.ListenAndServe(":3000", mux))
//	}()
//
//	url, stop, err := tunnel.StartTunnel("3000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//
// ## Gin
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
//
//	go func() { r.Run(":8080") }()
//
//	url, stop, err := tunnel.StartTunnel("8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//
// ## Gorilla mux
//
//	r := mux.NewRouter()
//	r.HandleFunc("/", homeHandler)
//
//	go func() {
//	    log.Fatal(http.ListenAndServe(":9000", r))
//	}()
//
//	url, stop, err := tunnel.StartTunnel("9000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//
// ## Fiber
//
//	app := fiber.New()
//	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
//
//	go func() { log.Fatal(app.Listen(":4000")) }()
//
//	url, stop, err := tunnel.StartTunnel("4000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//
// # Shutdown
//
// Always call the stop function returned by [StartTunnel].
//
// This ensures:
//   - TCP connection is closed cleanly
//   - tunnel is deregistered on the server
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
// # Notes
//
//   - Only HTTP traffic is supported currently.
//   - Each tunnel maps to a single local port.
//   - One client maintains a persistent connection (multiplexed internally).
//   - Future versions may support TLS, authentication, and traffic inspection.
//   - Advanced but useful features will come soon
//
// # Troubleshooting
//
//   - Ensure your local server is running before starting the tunnel.
//   - Verify the correct port is passed to StartTunnel.
//   - Check tunnel server connectivity if no public URL is returned.
//   - If requests fail, confirm your local handler responds correctly.
package tunnel

import (
	"fmt"
	"os"
)

// StartTunnel dials the tunnel server, starts forwarding in a background goroutine, and returns
// the public URL, a stop function (safe to defer), and an error if setup failed.
func StartTunnel(port string) (url string, stop func(), err error) {
	c, err := dialClient(port)
	if err != nil {
		return "", noop, fmt.Errorf("could not create tunnel: %w", err)
	}

	go func() {
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "gotunnel: tunnel stopped: %v\n", err)
		}
		printSuccess(c.getPublicURL(), "http://localhost:"+port)
	}()

	return c.getPublicURL(), func() { c.Stop() }, nil
}

func noop() {}
