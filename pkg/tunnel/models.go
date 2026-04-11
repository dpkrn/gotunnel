package tunnel

type ClientHello struct {
	TunnelType   string `json:"tunnel_type"`
	Version      string `json:"version"`
	TunnelID     string `json:"tunnel_id"`
	ConnectionID string `json:"connection_id"`
}
