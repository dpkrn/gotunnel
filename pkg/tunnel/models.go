package tunnel

import "time"

type clientHello struct {
	TunnelType   string `json:"tunnel_type"`
	Version      string `json:"version"`
	TunnelID     string `json:"tunnel_id"`
	ConnectionID string `json:"connection_id"`
}

// RequestLog is one captured HTTP request/response pair (tunnel → local app round trip).
type RequestLog struct {
	ID      string              `json:"id"`
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`

	Status   int    `json:"status"`
	RespBody string `json:"resp_body"`
	// RespHeaders is the upstream HTTP response header map (from your local app).
	RespHeaders map[string][]string `json:"resp_headers,omitempty"`

	Timestamp time.Time `json:"timestamp"`
	Duration  int64     `json:"duration_ms"`
}

// Themes names the built-in traffic inspector color schemes.
type Themes string

const (
	ThemesDark     Themes = "dark" // default
	ThemesLight    Themes = "light"
	ThemesTerminal Themes = "terminal"
)

type TunnelOptions struct {
	// Inspector enables the traffic inspector UI (default true when options are omitted).
	Inspector bool `json:"inspector"`
	// Themes selects the inspector palette: "dark" (default), "terminal", or "light".
	Themes string `json:"themes"`
	// Logs sets the number of logs to store in memory (default: 100).
	Logs int `json:"logs"`
	// InspectorAddr sets the listen address for the traffic inspector (default: ":4040").
	InspectorAddr string `json:"inspector_addr"`
}
