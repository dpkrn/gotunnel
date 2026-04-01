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

type Tunnel interface {
	Start() error
	Stop() error
	GetPublicUrl() string
}

type tunnel struct {
	Conn      net.Conn
	Session   *yamux.Session
	PublicUrl string
	Port      string
}

func NewTunnel(port string) (Tunnel, error) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		return nil, err
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

func (t *tunnel) Start() error {

	fmt.Println("Starting tunnel on port", t.Port)

	for {
		stream, err := t.Session.Accept()
		if err != nil {
			return err
		}

		go handle(stream, t.Port)
	}
}

func handle(stream net.Conn, port string) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	data, _ := reader.ReadBytes('\n')

	var req protocol.TunnelRequest
	json.Unmarshal(data, &req)

	httpReq, _ := http.NewRequest(
		req.Method,
		"http://localhost:"+port+req.Path,
		bytes.NewReader(req.Body),
	)

	for k, v := range req.Headers {
		for _, val := range v {
			httpReq.Header.Add(k, val)
		}
	}

	resp, _ := http.DefaultClient.Do(httpReq)
	body, _ := io.ReadAll(resp.Body)

	response := protocol.TunnelResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    body,
	}

	out, _ := json.Marshal(response)
	stream.Write(append(out, '\n'))
}

func (t *tunnel) Stop() error {
	t.Session.Close()
	t.Conn.Close()
	return nil
}

func (t *tunnel) GetPublicUrl() string {
	return t.PublicUrl
}
