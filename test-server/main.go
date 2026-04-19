package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "hello world"})
	})
	mux.HandleFunc("GET /name", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		firstname := q.Get("firstname")
		lastname := q.Get("lastname")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "hello " + firstname + " " + lastname,
		})
	})
	mux.HandleFunc("POST /sum", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Num1 int `json:"num1"`
			Num2 int `json:"num2"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"sum": body.Num1 + body.Num2})
	})

	tunnelOptions := tunnel.TunnelOptions{
		Inspector:    true,
		InspectorAdd: "4040",
	}

	url, stop, err := tunnel.StartTunnel("8080", tunnelOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer stop()
	fmt.Println("Public URL:", url)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
