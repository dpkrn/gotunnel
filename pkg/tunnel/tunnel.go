// Package tunnel is the public entry point for embedding gotunnel in your Go application.
//
// to install the library run the following command:
// go get github.com/DpkRn/gotunnel
//
// to use the library in your project, import the tunnel subpackage
//
//	import "github.com/DpkRn/gotunnel/pkg/tunnel"
//
// That function exposes a local HTTP server on a public URL by connecting to a
// tunnel server you run separately. Your app keeps serving on localhost; the
// tunnel forwards internet traffic to that port.
//
// Typical flow: start your HTTP server on a fixed port, call StartTunnel with
// the same port string, print the returned URL, and call the returned stop
// function on shutdown.
//
// # Requirements
//
//   - A reachable tunnel server (defaults match the gotunnel/mytunnel stack).
//   - The port you pass to StartTunnel must be the port your HTTP server listens on.
//
// Simple Main function in Go:
// package main
//
// import (
// 	"fmt"
// 	"log"
// 	"net/http"
//
// 	"github.com/DpkRn/gotunnel/pkg/tunnel"
// )
//
// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("→ request:", r.Method, r.URL.Path)
// 		w.WriteHeader(200)
// 		w.Write([]byte("hello world"))
// 	})
// 	// your server runs normally
// 	log.Fatal(http.ListenAndServe(":8000", nil))
// }
// ------------------------------------------------------------
// # Minimal program
//
// A complete small program using only the standard library:
//
//	package main
//
// import (
// 	"fmt"
// 	"log"
// 	"net/http"
//
// 	"github.com/DpkRn/gotunnel/pkg/tunnel"
// )
//
// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 	  fmt.Println("→ request:", r.Method, r.URL.Path)
//       w.WriteHeader(200)
// 	  w.Write([]byte("hello world"))
//     })
// 	url, stop, err := tunnel.StartTunnel("8080")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer stop()

// 	fmt.Println("Public URL:", url)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// # net/http
//
// Run [http.ListenAndServe] (or [http.Server.ListenAndServe]) in a goroutine,
// then StartTunnel with the same port:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/", apiHandler)
//	go func() {
//		log.Fatal(http.ListenAndServe(":3000", mux))
//	}()
//
//	url, stop, err := tunnel.StartTunnel("3000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Gin
//
// Run gin’s Run in a goroutine so StartTunnel can run on the main goroutine
// (or the reverse; the important part is both are running):
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
//	go func() { r.Run(":8080") }()
//
//	url, stop, err := tunnel.StartTunnel("8080")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Gorilla mux
//
// Pass your gorilla/mux Router to [http.ListenAndServe]:
//
//	r := mux.NewRouter()
//	r.HandleFunc("/", homeHandler)
//	go func() {
//		log.Fatal(http.ListenAndServe(":9000", r))
//	}()
//
//	url, stop, err := tunnel.StartTunnel("9000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Fiber
//
// Call fiber’s Listen in a goroutine with the address that matches your port:
//
//	app := fiber.New()
//	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })
//	go func() { log.Fatal(app.Listen(":4000")) }()
//
//	url, stop, err := tunnel.StartTunnel("4000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stop()
//
// # Shutdown
//
// Always call the stop function returned by StartTunnel when exiting (for example
// after [os.Signal] on SIGINT) so the tunnel connection closes cleanly.
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
//
// # How to use it

//main function

// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/DpkRn/gotunnel/pkg/tunnel"
// )

// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("→ request:", r.Method, r.URL.Path)
// 		w.WriteHeader(200)
// 		w.Write([]byte("hello world"))
// 	})
// 	log.Fatal(http.ListenAndServe(":8000", nil))
// }

// ##install the library
// go get github.com/DpkRn/gotunnel

// ##use the library in your project

// A complete small program using only the tunnel library  library:
//
// package main
//
// import (
// 	"fmt"
// 	"log"
// 	"net/http"
//
// 	"github.com/DpkRn/gotunnel/pkg/tunnel"
// )
//
// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("→ request:", r.Method, r.URL.Path)
// 		w.WriteHeader(200)
// 		w.Write([]byte("hello world"))
// 	})

// 	url, stop, err := tunnel.StartTunnel("8000")
// 	if err != nil {
// 		log.Fatal("tunnel error:", err)
// 	}
// 	defer stop()

// 	fmt.Println("🌍 Public URL:", url)
// 	// your server runs normally — tunnel stays alive alongside it
// 	log.Fatal(http.ListenAndServe(":8000", nil))
// }

// It will provide you public url and a stop function to stop the tunnel.
// - public url is the url you can use to access your server publicly

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
