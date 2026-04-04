package utils

import (
	"encoding/json"
	"net"

	"github.com/DpkRn/gotunnel/internal/models/protocol"
)

func WriteErr(stream net.Conn, code int, msg []byte) {
	r := protocol.TunnelResponse{Status: code, Body: msg}
	out, _ := json.Marshal(r)
	_, _ = stream.Write(append(out, '\n'))
}
