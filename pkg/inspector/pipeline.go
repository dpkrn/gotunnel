package inspector

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dpkrn/gotunnel/pkg/logstore"
	"github.com/gorilla/websocket"
)

type Pipeline struct {
	store *logstore.Logstore
	ws    *websocket.Conn
}

func NewPipeline(store *logstore.Logstore, ws *websocket.Conn) *Pipeline {
	return &Pipeline{store: store, ws: ws}
}

func (p *Pipeline) BrodcastLog(log logstore.RequestEvent) error {
	p.store.AddLog(log)
	payload := struct {
		EventType string                `json:"eventType"`
		Payload   logstore.RequestEvent `json:"payload"`
	}{
		EventType: "request",
		Payload:   log,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return err
	}
	err = p.ws.WriteMessage(websocket.TextMessage, []byte(payloadBytes))
	if err != nil {
		fmt.Println("Error writing message:", err)
		return err
	}
	return nil
}

func (p *Pipeline) GetLog(id string) (logstore.RequestEvent, error) {
	log := p.store.GetLog(id)
	if log.ID == "" {
		return logstore.RequestEvent{}, fmt.Errorf("log not found")
	}
	return log, nil
}

func (p *Pipeline) BindToConn(conn *websocket.Conn) {
	p.ws = conn
}
func (p *Pipeline) UnbindFromConn(conn *websocket.Conn) {
	p.ws = nil
}

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(inspector Inspector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		inspector.BindToConn(conn)
		defer inspector.UnbindFromConn(conn)
		fmt.Println("Bound to connection")
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			fmt.Printf("Received: %s\\n", message)
			if err != nil {
				break
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println("Error writing message:", err)
				break
			}
		}
	}
}
