// Implementation of the inspector HTTP server and WebSockets. The language-agnostic protocol
// (ingest URL, JSON per WebSocket frame) is documented in the package doc in doc.go.
package inspector

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dpkrn/gotunnel/pkg/inspector/logstore"
	"github.com/google/uuid"
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
	// forwardPort is the local HTTP port the tunnel forwards to (digits only), for HTML/JS defaults.
	port string // e.g. "4040"
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func newServer(store *logstore.Logstore, port string) *server {
	return &server{
		viewers: make(map[*websocket.Conn]struct{}),
		store:   store,
		port:    port,
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

func cloneHeaderMap(h map[string][]string) map[string][]string {
	if h == nil {
		return nil
	}
	out := make(map[string][]string, len(h))
	for k, vals := range h {
		out[k] = append([]string(nil), vals...)
	}
	return out
}

// pathForReplayLog returns path + raw query for log display (matches tunnel-style path).
func pathForReplayLog(u *url.URL) string {
	if u == nil {
		return "/"
	}
	p := u.Path
	if p == "" {
		p = "/"
	}
	if u.RawQuery != "" {
		p += "?" + u.RawQuery
	}
	return p
}

func (s *server) recordReplay(ev logstore.RequestEvent) {
	if err := s.ingestEvent(ev); err != nil {
		log.Printf("inspector: replay log: %v", err)
	}
}

// HeaderLogReplay is sent by the inspector UI on POST /replay. When true, the exchange is stored
// and broadcast; otherwise the request is proxied to localhost only (no history entry).
const HeaderLogReplay = "X-Inspector-Log-Replay"

func logReplayRequested(r *http.Request) bool {
	v := strings.TrimSpace(strings.ToLower(r.Header.Get(HeaderLogReplay)))
	if v == "" {
		return false
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// handleReplay proxies a request to localhost only (SSRF-safe). Used by the UI "Replay" action.
func (s *server) handleReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+HeaderLogReplay)
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
	logReplay := logReplayRequested(r)
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
	reqEv := logstore.Request{
		Method:  p.Method,
		Path:    pathForReplayLog(u),
		Body:    []byte(p.Body),
		Headers: cloneHeaderMap(p.Headers),
	}
	if err != nil {
		if logReplay {
			s.recordReplay(logstore.RequestEvent{
				ID:         "req_" + uuid.New().String(),
				Source:     "replay",
				Request:    reqEv,
				Response:   logstore.Response{StatusCode: http.StatusBadGateway, Headers: map[string][]string{}, Body: []byte(err.Error())},
				DurationMs: dur,
			})
		}
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
		if logReplay {
			s.recordReplay(logstore.RequestEvent{
				ID:         "req_" + uuid.New().String(),
				Source:     "replay",
				Request:    reqEv,
				Response:   logstore.Response{StatusCode: http.StatusBadGateway, Headers: map[string][]string{}, Body: []byte(err.Error())},
				DurationMs: dur,
			})
		}
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":      err.Error(),
			"durationMs": dur,
		})
		return
	}
	if logReplay {
		s.recordReplay(logstore.RequestEvent{
			ID:      "req_" + uuid.New().String(),
			Source:  "replay",
			Request: reqEv,
			Response: logstore.Response{
				StatusCode: resp.StatusCode,
				Headers:    cloneHeaderMap(resp.Header),
				Body:       respBody,
			},
			DurationMs: dur,
		})
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

func (s *server) serveInspectorHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	out := bytes.ReplaceAll(inspectorHTML, []byte("__LOCAL_APP_PORT__"), []byte(s.port))
	_, _ = w.Write(out)
}

func buildMux(srv *server) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleViewerWS())
	mux.HandleFunc("GET /ingest", srv.handleIngestWS())
	mux.HandleFunc("GET /logs", srv.serveLogs)
	mux.HandleFunc("GET /log", srv.serveLogByID)
	mux.HandleFunc("POST /replay", srv.handleReplay)
	mux.HandleFunc("OPTIONS /replay", srv.handleReplay)

	mux.HandleFunc("GET /", srv.serveInspectorHTML)
	mux.HandleFunc("GET /inspector.css", serveEmbedded(inspectorCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /theme-postman.css", serveEmbedded(themePostmanCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /theme-terminal.css", serveEmbedded(themeTerminalCSS, "text/css; charset=utf-8"))
	mux.HandleFunc("GET /index.js", serveEmbedded(indexJS, "application/javascript; charset=utf-8"))
	return mux
}

// StartInspector serves the inspector in the background. Call the returned stop function on shutdown
// (e.g. chain with the tunnel stop callback).
// It binds the listen address synchronously so callers (e.g. ingest WebSocket dial) can connect immediately.
func StartInspector(listen string) (stop func(), err error) {
	store := logstore.NewLogstore()
	srv := newServer(store, listen)
	mux := buildMux(srv)

	ln, err := net.Listen("tcp", ":"+listen)
	if err != nil {
		return nil, fmt.Errorf("inspector listen  %w", err)
	}
	hs := &http.Server{Handler: mux}
	go func() {
		if err := hs.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("inspector: server: %v", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = hs.Shutdown(ctx)
	}, nil
}

// Run starts the inspector HTTP server and blocks until the server exits (usually on error).
// listen is a port like "4040" (becomes ":4040") or a full host:port.
//
// Routes: GET / (UI), static /inspector.css, /theme-*.css, /index.js,
// GET /ws, GET /ingest, GET /logs, GET /log, POST /replay.
// forwardPort is the default local app port shown in the UI; empty defaults to "8080".
func Run(listen string) error {
	store := logstore.NewLogstore()
	srv := newServer(store, listen)
	mux := buildMux(srv)
	return http.ListenAndServe(":"+listen, mux)
}
