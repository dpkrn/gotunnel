package tunnel

import (
	"strings"
)

// listenAddrForInspector normalizes a port or host:port like the inspector server ("4040", ":4040", "127.0.0.1:9090").
func listenAddrForInspector(port string) string {
	p := strings.TrimSpace(port)
	if p == "" {
		return ":4040"
	}
	if strings.Contains(p, ":") {
		return p
	}
	return ":" + p
}

// InspectorHTTPURL is the URL users open in a browser for the inspector UI (GET /), for the same
// listen string as the inspector server ("4040", ":9090", "127.0.0.1:8080", …). This is not the
// WebSocket ingest URL.
func InspectorHTTPURL(listen string) string {
	addr := listenAddrForInspector(listen)
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}

// IngestWebSocketURL returns the ws:// URL tunnel clients should dial for ingest (GET /ingest),
// for the same listen string as the inspector server ("4040", ":9090", "127.0.0.1:8080", …).
// It does not import the inspector package — same URL nodetunnel or any runtime should use.
