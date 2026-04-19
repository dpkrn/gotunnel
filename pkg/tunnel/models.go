package tunnel

type clientHello struct {
	TunnelType   string `json:"tunnel_type"`
	Version      string `json:"version"`
	TunnelID     string `json:"tunnel_id"`
	ConnectionID string `json:"connection_id"`
}

type Method string

const (
	MethodGet     Method = "GET"
	MethodPost    Method = "POST"
	MethodPut     Method = "PUT"
	MethodDelete  Method = "DELETE"
	MethodPatch   Method = "PATCH"
	MethodOptions Method = "OPTIONS"
	MethodHead    Method = "HEAD"
	MethodAny     Method = "ANY"
)
