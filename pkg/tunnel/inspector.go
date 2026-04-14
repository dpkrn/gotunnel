package tunnel

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultInspectorAddr = ":4040"

// inspectorHTTPBaseURL returns the URL users open in a browser (matches opts.InspectorAddr / default).
func inspectorHTTPBaseURL(opts TunnelOptions) string {
	addr := strings.TrimSpace(opts.InspectorAddr)
	if addr == "" {
		addr = defaultInspectorAddr
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	return "http://" + addr
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

// inspectorHub owns WebSocket clients for live log streaming. It is safe for
// concurrent use.
type inspectorHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newInspectorHub() *inspectorHub {
	return &inspectorHub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *inspectorHub) register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *inspectorHub) unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	_ = c.Close()
}

func (h *inspectorHub) closeAll() {
	h.mu.Lock()
	for c := range h.clients {
		delete(h.clients, c)
		_ = c.Close()
	}
	h.mu.Unlock()
}

func (h *inspectorHub) broadcast(entry RequestLog) {
	h.mu.Lock()
	list := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		list = append(list, c)
	}
	h.mu.Unlock()

	for _, c := range list {
		if err := c.WriteJSON(entry); err != nil {
			h.unregister(c)
		}
	}
}

// startInspector runs the traffic inspector UI. Address comes from
// opts.InspectorAddr or [defaultInspectorAddr]. Theme comes from opts.Themes
// ("dark", "terminal", "light"). localPort is the tunnel target for POST /replay.
func startInspector(opts TunnelOptions, localPort string) func() {
	addr := strings.TrimSpace(opts.InspectorAddr)
	if addr == "" {
		addr = defaultInspectorAddr
	}
	themeClass := normalizeInspectorTheme(opts.Themes)

	hub := newInspectorHub()
	unsub := RegisterLogSubscriber(hub.broadcast)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		serveInspectorUI(w, r, themeClass)
	})
	mux.HandleFunc("GET /logs", serveInspectorLogs)
	mux.HandleFunc("POST /replay", handleReplay(localPort))
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		handleInspectorWS(hub, w, r)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Fprintf(os.Stderr, "gotunnel: traffic inspector → http://127.0.0.1%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "gotunnel: inspector stopped: %v\n", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)

		unsub()
		hub.closeAll()
	}
}

func normalizeInspectorTheme(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "terminal":
		return "theme-terminal"
	case "light":
		return "theme-light"
	case "dark", "":
		return "theme-dark"
	default:
		return "theme-dark"
	}
}

func serveInspectorUI(w http.ResponseWriter, r *http.Request, bodyClass string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := strings.Replace(inspectorPageHTML, "__THEME_CLASS__", bodyClass, 1)
	w.Write([]byte(page))
}

//go:embed inspector_page.html
var inspectorPageHTML string

func serveInspectorLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(GetLogs())
}

func handleInspectorWS(hub *inspectorHub, w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	hub.register(conn)

	// Drain pings / detect disconnect; broadcast is server → client only.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.unregister(conn)
			return
		}
	}
}
