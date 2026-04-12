package tunnel

import "time"

type clientHello struct {
	TunnelType   string `json:"tunnel_type"`
	Version      string `json:"version"`
	TunnelID     string `json:"tunnel_id"`
	ConnectionID string `json:"connection_id"`
}

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
