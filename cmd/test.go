package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DpkRn/gotunnel/pkg/tunnel"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("→ request:", r.Method, r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte("hello world"))
	})

	url, stop, err := tunnel.StartTunnel("8000")
	if err != nil {
		log.Fatal("tunnel error:", err)
	}
	defer stop()

	fmt.Println("🌍 Public URL:", url)
	// your server runs normally — tunnel stays alive alongside it
	log.Fatal(http.ListenAndServe(":8000", nil))
}
