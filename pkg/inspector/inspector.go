package inspector

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dpkrn/gotunnel/pkg/logstore"
	"github.com/gorilla/websocket"
)

//go:embed inspector.html
var inspectorHTML []byte

type inspector struct {
	Port     string
	Pipeline *Pipeline
}
type Inspector interface {
	BrodcastLog(log logstore.RequestEvent) error
	GetLogs() ([]logstore.RequestEvent, error)
	GetLog(id string) (logstore.RequestEvent, error)
	BindToConn(conn *websocket.Conn)
	UnbindFromConn(conn *websocket.Conn)
}

func NewInspector(port string) Inspector {
	return &inspector{
		Port: port,
		Pipeline: NewPipeline(
			logstore.NewLogstore(),
			nil,
		),
	}
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

// StartInspector serves the inspector UI and WebSocket on addr derived from inspectorListenPort (e.g. "4040" → ":4040").
func StartInspector(inspectorListenPort string) Inspector {
	mux := http.NewServeMux()
	insp := NewInspector(inspectorListenPort)
	addr := listenAddr(inspectorListenPort)

	mux.HandleFunc("GET /ws", wsHandler(insp))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(inspectorHTML)
	})

	mux.HandleFunc("GET /logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logs, err := insp.GetLogs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(logs)
	})
	listenURL := "http://" + addr
	if strings.HasPrefix(addr, ":") {
		listenURL = "http://127.0.0.1" + addr
	}
	log.Printf("inspector: listening on %s (GET /, WS /ws)\n", listenURL)
	go func() {
		log.Fatal(http.ListenAndServe(addr, mux))
	}()
	return insp
}

func (i *inspector) BrodcastLog(log logstore.RequestEvent) error {
	err := i.Pipeline.BrodcastLog(log)
	if err != nil {
		return fmt.Errorf("failed to brodcast log: %w", err)
	}
	return nil
}

func (i *inspector) GetLog(id string) (logstore.RequestEvent, error) {
	log, err := i.Pipeline.GetLog(id)
	if err != nil {
		return logstore.RequestEvent{}, err
	}
	return log, nil
}

func (i *inspector) BindToConn(conn *websocket.Conn) {
	i.Pipeline.BindToConn(conn)
}

func (i *inspector) UnbindFromConn(conn *websocket.Conn) {
	i.Pipeline.UnbindFromConn(conn)
}

func (i *inspector) GetLogs() ([]logstore.RequestEvent, error) {
	return i.Pipeline.store.Logs, nil
}
