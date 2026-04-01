// Package tunnel implements the core client-side tunneling logic.
//
// It manages a TCP connection to the gotunnel server, multiplexes streams
// using yamux, and proxies each incoming stream as an HTTP request to the
// local service running on the configured port.
package tunnel

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/DpkRn/gotunnel/internal/models/protocol"
	"github.com/hashicorp/yamux"
)

// Tunnel represents an active tunnel connection to the gotunnel server.
// Use NewTunnel to create one, then call Start in a goroutine to begin
// accepting and proxying traffic.
type Tunnel interface {
	// Start blocks and accepts incoming streams from the tunnel server,
	// forwarding each one as an HTTP request to the local service.
	// It returns when the session is closed or an unrecoverable error occurs.
	Start() error

	// Stop closes the yamux session and the underlying TCP connection,
	// terminating all active streams.
	Stop() error

	// GetPublicUrl returns the public URL assigned by the tunnel server
	// when the connection was first established.
	GetPublicUrl() string
}

// tunnel is the concrete implementation of Tunnel.
type tunnel struct {
	// Conn is the raw TCP connection to the tunnel server.
	Conn net.Conn
	// Session is the yamux multiplexer layered on top of Conn.
	Session *yamux.Session
	// PublicUrl is the URL the tunnel server assigned to this client.
	PublicUrl string
	// Port is the local port to which incoming requests are forwarded.
	Port string
}

// NewTunnel dials the gotunnel server at localhost:9000, negotiates a yamux
// session, and reads the public URL assigned by the server.
//
// It returns an initialised Tunnel ready for Start to be called, or an error
// if any step of the handshake fails.
func NewTunnel(port string) (Tunnel, error) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		return nil, fmt.Errorf("could not connect to tunnel server: %w", err)
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("session error: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("did not receive public URL: %w", err)
	}

	publicURL := "http://" + strings.TrimSpace(string(buf[:n]))
	return &tunnel{
		Conn:      conn,
		Session:   session,
		PublicUrl: publicURL,
		Port:      port,
	}, nil
}

// Start blocks and continuously accepts yamux streams from the tunnel server.
// Each stream is handled concurrently in its own goroutine. Start returns when
// the session is closed (e.g. after Stop is called) or an error occurs.
func (t *tunnel) Start() error {
	fmt.Println("tunnel: listening on port", t.Port)

	for {
		stream, err := t.Session.Accept()
		if err != nil {
			return fmt.Errorf("session closed: %w", err)
		}

		go handle(stream, t.Port)
	}
}

// handle processes a single yamux stream: it reads a TunnelRequest, issues
// the equivalent HTTP request to the local service, and writes the response
// back to the stream.
func handle(stream net.Conn, port string) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Println("tunnel: failed to read request:", err)
		return
	}

	var req protocol.TunnelRequest
	if err := json.Unmarshal(data, &req); err != nil {
		fmt.Println("tunnel: failed to decode request:", err)
		return
	}

	httpReq, err := http.NewRequest(
		req.Method,
		"http://localhost:"+port+req.Path,
		bytes.NewReader(req.Body),
	)
	if err != nil {
		fmt.Println("tunnel: failed to build HTTP request:", err)
		return
	}

	for k, v := range req.Headers {
		for _, val := range v {
			httpReq.Header.Add(k, val)
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Println("tunnel: local service error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("tunnel: failed to read response body:", err)
		return
	}

	response := protocol.TunnelResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    body,
	}

	out, err := json.Marshal(response)
	if err != nil {
		fmt.Println("tunnel: failed to encode response:", err)
		return
	}

	stream.Write(append(out, '\n'))
}

// Stop closes the yamux session and the underlying TCP connection.
// Any in-flight streams will be terminated. It is safe to call Stop
// more than once.
func (t *tunnel) Stop() error {
	t.Session.Close()
	t.Conn.Close()
	return nil
}

// GetPublicUrl returns the public URL assigned by the tunnel server.
func (t *tunnel) GetPublicUrl() string {
	return t.PublicUrl
}
