package main

import (
	"log"

	"github.com/dpkrn/gotunnel/pkg/inspector"
)

func main() {
	log.Fatal(inspector.Run("4040"))
}
