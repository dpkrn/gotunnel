package tunnel

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func Start(port string, serverAddr string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}

		req := string(buffer[:n])
		fmt.Println("Incoming request:", req)

		parts := strings.Split(req, " ")
		if len(parts) < 2 {
			continue
		}

		path := parts[1]

		resp, err := http.Get("http://localhost:" + port + path)
		if err != nil {
			conn.Write([]byte("Error calling local server"))
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		conn.Write(body)
	}
}
