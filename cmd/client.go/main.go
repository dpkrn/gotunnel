package main

import (
	"fmt"
	"os"

	internal "github.com/DpkRn/gotunnel/internal/tunnel"
	"github.com/DpkRn/gotunnel/pkg/tunnel"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: mytunnel http <port>")
		return
	}

	port := os.Args[2]

	publicUrl := internal.GetPublicUrl()
	err := tunnel.Start(port, publicUrl)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Tunnel started")
	fmt.Println("Public URL:", publicUrl)
}
