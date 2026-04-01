package protocol

import "net/http"

type TunnelRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}
