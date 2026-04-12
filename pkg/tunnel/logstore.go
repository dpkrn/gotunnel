package tunnel

import "sync"

const maxRequestLogs = 100

var (
	requestLogsMu sync.RWMutex
	requestLogs   []RequestLog
)

// AddLog records a proxied request/response and pushes it to live inspector clients.
func AddLog(entry RequestLog) {
	requestLogsMu.Lock()
	requestLogs = append(requestLogs, entry)
	if len(requestLogs) > maxRequestLogs {
		requestLogs = requestLogs[len(requestLogs)-maxRequestLogs:]
	}
	requestLogsMu.Unlock()

	// Notify outside the lock so inspector / WebSocket work cannot block other AddLog/GetLogs callers.
	notifyInspectorSubscribers(entry)
}

// GetLogs returns a snapshot of recent logs (newest appended last).
func GetLogs() []RequestLog {
	requestLogsMu.RLock()
	defer requestLogsMu.RUnlock()
	out := make([]RequestLog, len(requestLogs))
	copy(out, requestLogs)
	return out
}
