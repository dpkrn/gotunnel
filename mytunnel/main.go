package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/dpkrn/gotunnel/pkg/tunnel"
)

func printHelp() {
	fmt.Println("mytunnel — expose your local server to the internet")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mytunnel <command> <port>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  http <port>   Forward HTTP traffic to localhost:<port>")
	fmt.Println("  help          Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  mytunnel http 3000")
	fmt.Println("  mytunnel http 8080")
	fmt.Println()
	fmt.Println("Optional: send traffic to a standalone inspector (run: go run ./cmd/inspector)")
	fmt.Println("  GOTUNNEL_INSPECTOR_WS=ws://127.0.0.1:4040/ingest mytunnel http 8080")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]

	if command == "help" || command == "--help" || command == "-h" {
		printHelp()
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: mytunnel http <port>")
		return
	}

	port := os.Args[2]

	switch command {
	case "http":
		opts := tunnel.Options{}
		if u := os.Getenv("GOTUNNEL_INSPECTOR_WS"); u != "" {
			opts.InspectorIngestURL = u
		}
		_, stop, err := tunnel.StartTunnelWithOptions(port, opts)
		if err != nil {
			stop()
			fmt.Println("could not start tunnel", err)
			return
		}
		defer stop()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Run 'mytunnel help' to see available commands.")
	}
}
