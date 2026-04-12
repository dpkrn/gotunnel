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
	"time"

	"github.com/hashicorp/yamux"
)

// Default control-plane address (tunnel server TCP).
const defaultControlAddr = "clickly.cv:9000"

type clientConn struct {
	conn      net.Conn
	session   *yamux.Session
	publicURL string
	port      string
}

func dialClient(port string) (*clientConn, error) {
	conn, err := net.Dial("tcp", defaultControlAddr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to tunnel server: %w", err)
	}

	//send client hello
	tunnelReq := clientHello{
		TunnelType:   "gotunnel",
		Version:      "1.0.8",
		TunnelID:     "random-tunnel-id", //todo: generate a fixed tunnel for user
		ConnectionID: generateConnectionID(),
	}
	tunnelReqBytes, err := json.Marshal(tunnelReq)
	if err != nil {
		fmt.Println("Error marshalling tunnel request:", err)
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
	return &clientConn{
		conn:      conn,
		session:   session,
		publicURL: publicURL,
		port:      port,
	}, nil
}

func (c *clientConn) Start() error {
	fmt.Fprintf(os.Stderr, "gotunnel: listening on port %s\n", c.port)

	for {
		stream, err := c.session.Accept()
		if err != nil {
			return fmt.Errorf("session closed: %w", err)
		}

		go handleStream(stream, c.port)
	}
}

func handleStream(stream net.Conn, port string) {
	start := time.Now()
	defer stream.Close()

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
		"http://localhost:"+port+req.Path,
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

	AddLog(requestLog{
		ID:          generateID(),
		Method:      req.Method,
		Path:        req.Path,
		Headers:     req.Headers,
		Body:        string(req.Body),
		Status:      resp.StatusCode,
		RespBody:    string(body),
		RespHeaders: resp.Header,
		Timestamp:   time.Now(),
		Duration:    time.Since(start).Milliseconds(),
	})

}

func (c *clientConn) Stop() error {
	c.session.Close()
	c.conn.Close()
	return nil
}

func (c *clientConn) getPublicURL() string {
	return c.publicURL
}

func printSuccess(publicURL string, localURL string) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════╗")
	fmt.Println("  ║   🚇  mytunnel — tunnel is live                  ║")
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Printf("  ║  🌍  Public   →  %-32s║\n", publicURL)
	fmt.Printf("  ║  💻  Local    →  %-32s║\n", localURL)
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Println("  ║  ⚡  Forwarding requests...                      ║")
	fmt.Println("  ║  🛑  Press Ctrl+C to stop                        ║")
	fmt.Println("  ╚══════════════════════════════════════════════════╝")
	fmt.Println()
}
