package tunnel

import "sync"

const defaultMaxRequestLogs = 100

var (
	maxRequestLogs = defaultMaxRequestLogs
	requestLogsMu  sync.RWMutex
	requestLogs    []requestLog
)

// setMaxRequestLogs configures how many recent entries AddLog retains (minimum 1).
func setMaxRequestLogs(n int) {
	if n < 1 {
		n = defaultMaxRequestLogs
	}
	requestLogsMu.Lock()
	maxRequestLogs = n
	if len(requestLogs) > maxRequestLogs {
		requestLogs = requestLogs[len(requestLogs)-maxRequestLogs:]
	}
	requestLogsMu.Unlock()
}

// AddLog records a proxied request/response and pushes it to live inspector clients.
func AddLog(entry requestLog) {
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
func GetLogs() []requestLog {
	requestLogsMu.RLock()
	defer requestLogsMu.RUnlock()
	out := make([]requestLog, len(requestLogs))
	copy(out, requestLogs)
	return out
}
