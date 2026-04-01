package protocol

import "net/http"

type TunnelResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}
