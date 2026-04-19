package tcp

import "net"

type TCPClient struct {
	conn net.Conn
}

func NewTCPClient(addr string) (*TCPClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &TCPClient{conn: conn}, nil
}
