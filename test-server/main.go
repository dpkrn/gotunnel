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
