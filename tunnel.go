package gotunnel

// Package tunnel provides a simple API for exposing a local HTTP server
// to the public internet through a gotunnel server.
//
// It establishes a persistent TCP connection to the tunnel server, receives
// a public URL, and proxies incoming requests to the specified local port.
//
// Basic usage:
//
//	url, stop, err := tunnel.StartTunnel("8080")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stop()
//
//	fmt.Println("Public URL:", url)
//	log.Fatal(http.ListenAndServe(":8080", nil))

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

	return t.GetPublicUrl(), func() { t.Stop() }, nil
}

// noop is a no-op stop function returned on error so callers can safely defer stop().
func noop() {}
