// Package inspector is a standalone HTTP/WebSocket traffic inspector.
//
// Run [Run] to serve the UI, log APIs, and WebSocket endpoints. Tunnels (gotunnel,
// nodetunnel, etc.) connect to GET /ingest and send one JSON-encoded
// logstore.RequestEvent per WebSocket text message; the inspector stores each event
// and broadcasts the same envelope to browser clients on GET /ws
// ({ "eventType": "request", "payload": <RequestEvent> }).
package inspector

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dpkrn/gotunnel/pkg/inspector/logstore"
	"github.com/gorilla/websocket"
)

//go:embed inspector.html
var inspectorHTML []byte

//go:embed inspector.css
var inspectorCSS []byte

//go:embed theme-postman.css
var themePostmanCSS []byte

//go:embed theme-terminal.css
var themeTerminalCSS []byte

//go:embed index.js
var indexJS []byte

type server struct {
	mu      sync.RWMutex
	viewers map[*websocket.Conn]struct{}
	store   *logstore.Logstore
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func newServer(store *logstore.Logstore) *server {
	return &server{
		viewers: make(map[*websocket.Conn]struct{}),
		store:   store,
	}
}

func (s *server) registerViewer(c *websocket.Conn) {
	s.mu.Lock()
	s.viewers[c] = struct{}{}
	s.mu.Unlock()
}

func (s *server) unregisterViewer(c *websocket.Conn) {
	s.mu.Lock()
	delete(s.viewers, c)
	s.mu.Unlock()
}

// ingestEvent persists ev and fans out to UI WebSocket clients on /ws.
func (s *server) ingestEvent(ev logstore.RequestEvent) error {
	s.store.AddLog(ev)

	env := struct {
		EventType string                `json:"eventType"`
		Payload   logstore.RequestEvent `json:"payload"`
	}{
		EventType: "request",
		Payload:   ev,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal broadcast: %w", err)
	}

	s.mu.RLock()
	list := make([]*websocket.Conn, 0, len(s.viewers))
	for c := range s.viewers {
		list = append(list, c)
	}
	s.mu.RUnlock()

	for _, c := range list {
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			s.unregisterViewer(c)
			_ = c.Close()
		}
	}
	return nil
}

func (s *server) handleViewerWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.registerViewer(conn)
		defer func() {
			s.unregisterViewer(conn)
			_ = conn.Close()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}

// handleIngestWS accepts tunnel producers: each text frame is JSON for one [logstore.RequestEvent].
func (s *server) handleIngestWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = conn.Close() }()
		log.Printf("inspector: ingest connection from %s", r.RemoteAddr)

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var ev logstore.RequestEvent
			if err := json.Unmarshal(data, &ev); err != nil {
				log.Printf("inspector: ingest decode: %v", err)
				continue
			}
			if err := s.ingestEvent(ev); err != nil {
				log.Printf("inspector: ingest: %v", err)
			}
		}
	}
}

func (s *server) serveLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	logs := s.store.GetLogs()
	_ = json.NewEncoder(w).Encode(logs)
}

func (s *server) serveLogByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	log, err := s.store.GetLog(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(log)
}

func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.Trim(strings.TrimSpace(host), "[]"))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func allowReplayURL(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return isLoopbackHost(u.Hostname())
}

// handleReplay proxies a request to localhost only (SSRF-safe). Used by the UI "Replay" action.
func (s *server) handleReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p struct {
		Method  string              `json:"method"`
		URL     string              `json:"url"`
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(p.Method) == "" {
		p.Method = http.MethodGet
	}
	u, err := url.Parse(p.URL)
	if err != nil || u.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if !allowReplayURL(u) {
		http.Error(w, "only http(s) URLs on localhost are allowed", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, p.Method, p.URL, bytes.NewReader([]byte(p.Body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for k, vals := range p.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	cli := &http.Client{Timeout: 120 * time.Second}
	start := time.Now()
	resp, err := cli.Do(req)
	dur := time.Since(start).Milliseconds()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      err.Error(),
			"durationMs": dur,
		})
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      err.Error(),
			"durationMs": dur,
		})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"statusCode": resp.StatusCode,
		"headers":    resp.Header,
		"body":       respBody,
		"durationMs": dur,
	})
}

func serveEmbedded(content []byte, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(content)
	}
}

func serveInspectorHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(inspectorHTML)
}

// listenAddr turns "4040", ":4040", or "127.0.0.1:9090" into a value suitable for [http.ListenAndServe].
func listenAddr(port string) string {
	p := strings.TrimSpace(port)
	if p == "" {
		return ":4040"
	}
	if strings.Contains(p, ":") {
		return p
	}
	return ":" + p
}

// Run starts the inspector HTTP server and blocks until the server exits (usually on error).
// listen is a port like "4040" (becomes ":4040") or a full host:port.
//
// Routes: GET / (UI), static /inspector.css, /theme-*.css, /index.js,
// GET /ws, GET /ingest, GET /logs, GET /log, POST /replay.
func Run(listen string) error {
	store := logstore.NewLogstore()
	srv := newServer(store)
	addr := listenAddr(listen)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleViewerWS())
	mux.HandleFunc("GET /ingest", srv.handleIngestWS())
	mux.HandleFunc("GET /logs", srv.serveLogs)
	mux.HandleFunc("GET /log", srv.serveLogByID)
	mux.HandleFunc("POST /replay", srv.handleReplay)
	mux.HandleFunc("OPTIONS /replay", srv.handleReplay)

	mux.HandleFunc("GET /", serveInspectorHTML)
	mux.HandleFunc("GET /inspector.css", serveEmbedded(inspectorCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /theme-postman.css", serveEmbedded(themePostmanCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /theme-terminal.css", serveEmbedded(themeTerminalCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /index.js", serveEmbedded(indexJS, "application/javascript; charset=utf-8"))

	listenURL := "http://" + addr
	if strings.HasPrefix(addr, ":") {
		listenURL = "http://127.0.0.1" + addr
	}
	log.Printf("inspector: listening on %s (UI /, viewers /ws, ingest /ingest)\n", listenURL)
	return http.ListenAndServe(addr, mux)
}
