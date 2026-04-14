package tunnel

import (
	"sync"
	"sync/atomic"
)

const defaultMaxrequestLogs = 100

var (
	maxrequestLogs = defaultMaxrequestLogs
	requestLogsMu  sync.RWMutex
	requestLogs    []requestLog

	// logSubscribers receives each new log after it is stored (for live UIs, metrics, etc.).
	logSubMu       sync.RWMutex
	logSubscribers = make(map[int64]LogSubscriber)
	nextSubscriber atomic.Int64
)

// LogSubscriber is called after each [AddLog], outside the log mutex. Implementations should return quickly;
// heavy work should run in a new goroutine if needed.
type LogSubscriber func(requestLog)

// RegisterLogSubscriber adds a callback invoked after each [AddLog]. Call the returned function to unsubscribe.
// Safe for concurrent use. Passing nil returns a no-op unregister.
func RegisterLogSubscriber(fn LogSubscriber) (unregister func()) {
	if fn == nil {
		return func() {}
	}
	id := nextSubscriber.Add(1)
	logSubMu.Lock()
	logSubscribers[id] = fn
	logSubMu.Unlock()
	return func() {
		logSubMu.Lock()
		delete(logSubscribers, id)
		logSubMu.Unlock()
	}
}

func notifyLogSubscribers(entry requestLog) {
	logSubMu.RLock()
	subs := make([]LogSubscriber, 0, len(logSubscribers))
	for _, fn := range logSubscribers {
		subs = append(subs, fn)
	}
	logSubMu.RUnlock()

	for _, fn := range subs {
		go fn(entry)
	}
}

// setMaxrequestLogs configures how many recent entries AddLog retains (minimum 1).
func setMaxrequestLogs(n int) {
	if n < 1 {
		n = defaultMaxrequestLogs
	}
	requestLogsMu.Lock()
	maxrequestLogs = n
	if len(requestLogs) > maxrequestLogs {
		requestLogs = requestLogs[len(requestLogs)-maxrequestLogs:]
	}
	requestLogsMu.Unlock()
}

// AddLog records a proxied request/response and pushes it to registered subscribers (e.g. inspector WebSocket).
func addLog(entry requestLog) {
	requestLogsMu.Lock()
	requestLogs = append(requestLogs, entry)
	if len(requestLogs) > maxrequestLogs {
		requestLogs = requestLogs[len(requestLogs)-maxrequestLogs:]
	}
	requestLogsMu.Unlock()

	notifyLogSubscribers(entry)
}

// GetLogs returns a snapshot of recent logs (newest appended last).
func getLogs() []requestLog {
	requestLogsMu.RLock()
	defer requestLogsMu.RUnlock()
	out := make([]requestLog, len(requestLogs))
	copy(out, requestLogs)
	return out
}
