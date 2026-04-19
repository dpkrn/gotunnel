package client

import (
	"net"

	"github.com/dpkrn/gotunnel/internal/tcp"
)

type Client struct {
	tcpClient *tcp.TCPClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}
