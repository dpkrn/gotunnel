package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
)

type nameBody struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type nameResponse struct {
	FullName string `json:"full_name"`
}

func fullName(first, last string) string {
	f := strings.TrimSpace(first)
	l := strings.TrimSpace(last)
	return strings.TrimSpace(f + " " + l)
}

func writeNameJSON(w http.ResponseWriter, first, last string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(nameResponse{FullName: fullName(first, last)})
}

func handleNameQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	first := q.Get("first_name")
	last := q.Get("last_name")
	writeNameJSON(w, first, last)
}

func handleNamePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body nameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	writeNameJSON(w, body.FirstName, body.LastName)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("→ request:", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello world\n\nTry:\n  GET  /name?first_name=Ada&last_name=Lovelace\n  POST /name  {\"first_name\":\"Ada\",\"last_name\":\"Lovelace\"}\n"))
	})

	http.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleNameQuery(w, r)
		case http.MethodPost:
			handleNamePost(w, r)
		default:
			http.Error(w, "use GET or POST", http.StatusMethodNotAllowed)
		}
	})

	url, stop, err := tunnel.StartTunnel("8080", tunnel.TunnelOptions{
		Inspector:     true,
		Themes:        "terminal",
		Logs:          100,
		InspectorAddr: ":9090",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stop()
	fmt.Println("Public URL:", url)
	fmt.Println("Local URL:", "http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
