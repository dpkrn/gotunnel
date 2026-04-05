package tunnel

import (
	"encoding/json"
	"net"
	"net/http"
)

// Wire types for JSON over yamux (same shape as the tunnel server protocol).
type tunnelRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

type tunnelResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

func writeWireErr(stream net.Conn, code int, msg []byte) {
	r := tunnelResponse{Status: code, Body: msg}
	out, _ := json.Marshal(r)
	_, _ = stream.Write(append(out, '\n'))
}
