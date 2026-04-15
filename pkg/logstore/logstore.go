package logstore

import (
	"sync"
)

type RequestEvent struct {
	ID              string `json:"id"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	RequestBody     []byte `json:"requestBody"`
	RequestHeaders  map[string][]string `json:"requestHeaders"`
	ResponseTime    int64 `json:"responseTime"`
	StatusCode      int `json:"statusCode"`
	ResponseHeaders map[string][]string `json:"responseHeaders"`
	ResponseBody    []byte `json:"responseBody"`
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

func (l *Logstore) GetLog(id string) RequestEvent {
	l.Mu.Lock()
	defer l.Mu.Unlock()
	for _, log := range l.Logs {
		if log.ID == id {
			return log
		}
	}
	return RequestEvent{}
}
