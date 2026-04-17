// Standalone inspector: run separately, then point gotunnel (or nodetunnel) at ws://host:port/ingest.
package main

import (
	"log"
	"os"

	"github.com/dpkrn/gotunnel/pkg/inspector"
)

func main() {
	port := "4040"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	log.Fatal(inspector.Run(port, "8080"))
}
