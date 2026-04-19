package tunnel

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dpkrn/gotunnel/pkg/inspector/logstore"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

// Default control-plane address (tunnel server TCP).
const defaultControlAddr = "clickly.cv:9000"

type clientConn struct {
	conn         net.Conn
	session      *yamux.Session
	publicURL    string
	port         string
	ingestConn *websocket.Conn
	ingestMu   sync.Mutex
}

func dialClient(port string) (*clientConn, error) {
	conn, err := net.Dial("tcp", defaultControlAddr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to tunnel server: %w", err)
	}

	tunnelReq := ClientHello{
		TunnelType:   "gotunnel",
		Version:      "1.0.8",
		TunnelID:     "random-tunnel-id", //todo: generate a fixed tunnel for user
		ConnectionID: GenerateConnectionID(),
	}
	tunnelReqBytes, err := json.Marshal(tunnelReq)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error marshalling tunnel request: %w", err)
	}
	conn.Write(append(tunnelReqBytes, '\n'))

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

	line := strings.TrimSpace(string(buf[:n]))
	publicURL := line
	if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
		publicURL = "http://" + line
	}

	// Connect inspector ingest after the tunnel control plane is ready so StartInspector's HTTP
	// server has time to accept, and the tunnel path is never blocked waiting on the inspector.

	return &clientConn{
		conn:      conn,
		session:   session,
		publicURL: publicURL,
		port:      port,
	}, nil
}

func (c *clientConn) pushLog(ev logstore.RequestEvent) {
	c.ingestMu.Lock()
	conn := c.ingestConn
	c.ingestMu.Unlock()
	if conn == nil {
		return
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	c.ingestMu.Lock()
	defer c.ingestMu.Unlock()
	if c.ingestConn == nil {
		return
	}
	if err := c.ingestConn.WriteMessage(websocket.TextMessage, b); err != nil {
		_ = c.ingestConn.Close()
		c.ingestConn = nil
		fmt.Fprintf(os.Stderr, "gotunnel: inspector ingest write failed: %v\n", err)
	}
}

func (c *clientConn) Start() error {
	fmt.Fprintf(os.Stderr, "gotunnel: listening on port %s\n", c.port)

	for {
		stream, err := c.session.Accept()
		if err != nil {
			return fmt.Errorf("session closed: %w", err)
		}

		go handleStream(stream, c)
	}
}

func handleStream(stream net.Conn, c *clientConn) {
	defer stream.Close()
	startTime := time.Now()

	reader := bufio.NewReader(stream)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "gotunnel: failed to read request: %v\n", err)
		writeWireErr(stream, http.StatusBadRequest, []byte("read request: "+err.Error()))
		return
	}

	var req tunnelRequest
	if err := json.Unmarshal(data, &req); err != nil {
		fmt.Fprintf(os.Stderr, "gotunnel: failed to unmarshal request: %v\n", err)
		writeWireErr(stream, http.StatusBadRequest, []byte("unmarshal request: "+err.Error()))
		return
	}

	httpReq, err := http.NewRequest(
		req.Method,
		"http://localhost:"+c.port+req.Path,
		bytes.NewReader(req.Body),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gotunnel: failed to build HTTP request: %v\n", err)
		writeWireErr(stream, http.StatusBadRequest, []byte("build HTTP request: "+err.Error()))
		return
	}

	for k, v := range req.Headers {
		for _, val := range v {
			httpReq.Header.Add(k, val)
		}
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gotunnel: failed to do local request: %v\n", err)
		writeWireErr(stream, http.StatusBadGateway, []byte("local request failed: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		writeWireErr(stream, http.StatusBadGateway, []byte("read local response: "+err.Error()))
		return
	}

	response := tunnelResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    body,
	}

	out, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gotunnel: failed to marshal response: %v\n", err)
		writeWireErr(stream, http.StatusInternalServerError, []byte("marshal response: "+err.Error()))
		return
	}

	stream.Write(append(out, '\n'))
	go c.pushLog(
		logstore.RequestEvent{
			ID:     GenerateRequestID(),
			Source: "ingest",
			Request: logstore.Request{
				Method:  req.Method,
				Path:    req.Path,
				Body:    req.Body,
				Headers: req.Headers,
			},
			Response: logstore.Response{
				StatusCode: response.Status,
				Headers:    response.Headers,
				Body:       body,
			},
			DurationMs: time.Since(startTime).Milliseconds(),
		})
}

func (c *clientConn) Stop() error {
	c.ingestMu.Lock()
	if c.ingestConn != nil {
		_ = c.ingestConn.Close()
		c.ingestConn = nil
	}
	c.ingestMu.Unlock()
	c.session.Close()
	c.conn.Close()
	return nil
}

func (c *clientConn) getPublicURL() string {
	return c.publicURL
}

func printSuccess(publicURL string, localURL string, inspectorIngestURL string) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════╗")
	fmt.Println("  ║   🚇  mytunnel — tunnel is live                  ║")
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Printf("  ║  🌍  Public   →  %-32s║\n", publicURL)
	fmt.Printf("  ║  💻  Local    →  %-32s║\n", localURL)
	fmt.Printf("  ║  🔍  Inspector →  %-32s║\n", inspectorIngestURL)
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Println("  ║  ⚡  Forwarding requests...                      ║")
	fmt.Println("  ║  🛑  Press Ctrl+C to stop                        ║")
	fmt.Println("  ╚══════════════════════════════════════════════════╝")
	fmt.Println()
}
