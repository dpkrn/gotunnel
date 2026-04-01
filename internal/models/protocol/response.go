package protocol

import "net/http"

// TunnelResponse is the JSON-encoded message the client sends back to the
// tunnel server after forwarding a request to the local service. The server
// uses this to construct the HTTP response it returns to the original caller.
type TunnelResponse struct {
	// Status is the HTTP status code returned by the local service (e.g. 200, 404).
	Status int

	// Headers contains the HTTP response headers from the local service.
	Headers http.Header

	// Body holds the raw response body returned by the local service.
	Body []byte
}
