package logstore

import (
	"fmt"
	"sync"
)

// Request is the inbound HTTP request captured for the inspector.
type Request struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Body    []byte              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

// Response is the upstream HTTP response returned to the tunnel.
type Response struct {
	DurationMs int64               `json:"durationMs"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}

// RequestEvent is one logged exchange: id + request + response.
type RequestEvent struct {
	ID         string   `json:"id"`
	Request    Request  `json:"request"`
	Response   Response `json:"response"`
	DurationMs int64    `json:"durationMs"`
}

// Logstore holds captured request/response events (thread-safe).
type Logstore struct {
	Logs []RequestEvent
	Mu   sync.Mutex
}

func NewLogstore() *Logstore {
	return &Logstore{Logs: make([]RequestEvent, 0)}
}

func (l *Logstore) AddLog(log RequestEvent) {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	l.Logs = append(l.Logs, log)
}

func (l *Logstore) GetLog(id string) (RequestEvent, error) {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	for _, log := range l.Logs {
		if log.ID == id {
			return log, nil
		}
	}
	return RequestEvent{}, fmt.Errorf("log not found")
}

// GetLogs returns a snapshot of stored events (newest appended last).
func (l *Logstore) GetLogs() []RequestEvent {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	out := make([]RequestEvent, len(l.Logs))
	copy(out, l.Logs)
	return out
}
