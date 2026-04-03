// Package tunnel provides [StartTunnel] for embedding a gotunnel client in your app.
//
// Full overview, a minimal main, and patterns for net/http, Gin, Gorilla mux,
// and Fiber are on [github.com/DpkRn/gotunnel/pkg].
package tunnel

import (
	"fmt"

	"github.com/DpkRn/gotunnel/internal/tunnel"
)

// StartTunnel connects to the gotunnel server and begins forwarding public
// traffic to the local HTTP server running on the given port.
//
// It returns:
//   - url: the public URL assigned by the tunnel server (e.g. "http://xyz.example.com")
//   - stop: a function that shuts down the tunnel and releases all resources
//   - err: non-nil if the tunnel could not be established
//
// The tunnel runs in the background. Call stop() when your application exits
// or when you no longer need the tunnel. It is safe to defer stop().
//
//	url, stop, err := tunnel.StartTunnel("3000")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
func StartTunnel(port string) (url string, stop func(), err error) {
	t, err := tunnel.NewTunnel(port)
	if err != nil {
		return "", noop, fmt.Errorf("could not create tunnel: %w", err)
	}

	go func() {
		if err := t.Start(); err != nil {
			fmt.Println("tunnel stopped:", err)
		}
	}()
	fmt.Println("✅Public url:", t.GetPublicUrl())

	return t.GetPublicUrl(), func() { t.Stop() }, nil
}

// noop is a no-op stop function returned on error so callers can safely defer stop().
func noop() {}
