// Package tunnel exposes a local HTTP server on a public URL by connecting to a
// gotunnel server you run separately. Traffic hits the tunnel, then your app on localhost.
//
// # API
//
// Use [StartTunnel] for defaults, or [StartTunnelWithOptions] to attach a standalone
// inspector WebSocket ingest URL (see [Options]).
//
// Install:
//
//	go get github.com/dpkrn/gotunnel
//
// Import:
//
//	import "github.com/dpkrn/gotunnel/pkg/tunnel"
//
// # Requirements
//
//   - A reachable tunnel server (defaults match the gotunnel/mytunnel stack).
//   - The port passed to [StartTunnel] must be the port your HTTP server listens on.
//
// # Step 1 — local server only (no gotunnel yet)
//
// Run this first: standard library only. Visit http://localhost:8080 to confirm the handler works.
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"net/http"
//	)
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
// # Step 2 — same server, add the tunnel
//
// Add the import, call [StartTunnel] with the same port as [http.ListenAndServe], defer stop(),
// and print the public URL before you block in ListenAndServe.
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"net/http"
//
//		"github.com/dpkrn/gotunnel/pkg/tunnel"
//	)
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
// # net/http
//
// Run [http.ListenAndServe] in a goroutine, then [StartTunnel] with the same port:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/", apiHandler)
//	go func() {
//		log.Fatal(http.ListenAndServe(":3000", mux))
//	}()
//	url, stop, err := tunnel.StartTunnel("3000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Gin
//
// Run gin’s Run in a goroutine so the tunnel and server both run (add gin to your go.mod):
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
//	go func() { r.Run(":8080") }()
//	url, stop, err := tunnel.StartTunnel("8080")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Gorilla mux
//
// Pass a gorilla/mux Router to [http.ListenAndServe]:
//
//	r := mux.NewRouter()
//	r.HandleFunc("/", homeHandler)
//	go func() {
//		log.Fatal(http.ListenAndServe(":9000", r))
//	}()
//	url, stop, err := tunnel.StartTunnel("9000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Fiber
//
// Call fiber’s Listen in a goroutine with the same port as [StartTunnel] (add fiber to your go.mod):
//
//	app := fiber.New()
//	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
//	go func() { log.Fatal(app.Listen(":4000")) }()
//	url, stop, err := tunnel.StartTunnel("4000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Shutdown
//
// Always call the stop function from [StartTunnel] on exit (e.g. after [os.Signal] on SIGINT)
// so the tunnel connection closes cleanly.
package pkg
