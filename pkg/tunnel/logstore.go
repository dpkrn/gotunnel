package tunnel

import (
	"sync"
	"sync/atomic"
)

const defaultMaxRequestLogs = 100

var (
	maxRequestLogs = defaultMaxRequestLogs
	requestLogsMu  sync.RWMutex
	requestLogs    []RequestLog

	// logSubscribers receives each new log after it is stored (for live UIs, metrics, etc.).
	logSubMu       sync.RWMutex
	logSubscribers = make(map[int64]LogSubscriber)
	nextSubscriber atomic.Int64
)

// LogSubscriber is called after each [AddLog], outside the log mutex. Implementations should return quickly;
// heavy work should run in a new goroutine if needed.
type LogSubscriber func(RequestLog)

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

func notifyLogSubscribers(entry RequestLog) {
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

// AddLog records a proxied request/response and pushes it to registered subscribers (e.g. inspector WebSocket).
func AddLog(entry RequestLog) {
	requestLogsMu.Lock()
	requestLogs = append(requestLogs, entry)
	if len(requestLogs) > maxRequestLogs {
		requestLogs = requestLogs[len(requestLogs)-maxRequestLogs:]
	}
	requestLogsMu.Unlock()

	notifyLogSubscribers(entry)
}

// GetLogs returns a snapshot of recent logs (newest appended last).
func GetLogs() []RequestLog {
	requestLogsMu.RLock()
	defer requestLogsMu.RUnlock()
	out := make([]RequestLog, len(requestLogs))
	copy(out, requestLogs)
	return out
}
