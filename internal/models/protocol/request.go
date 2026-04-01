// Package protocol defines the wire format used between the gotunnel server
// and the client to describe HTTP requests and responses transported over
// yamux streams.
package protocol

import "net/http"

// TunnelRequest is the JSON-encoded message the tunnel server sends to the
// client over a yamux stream. It carries all information needed to reconstruct
// and forward the original HTTP request to the local service.
type TunnelRequest struct {
	// Method is the HTTP method of the original request (e.g. "GET", "POST").
	Method string

	// Path is the request path including any query string (e.g. "/api/users?page=2").
	Path string

	// Headers contains the HTTP headers from the original request.
	Headers http.Header

	// Body holds the raw request body, if any.
	Body []byte
}
