package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/DpkRn/gotunnel/pkg/tunnel"
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
		url, stop, err := tunnel.StartTunnel(port)
		if err != nil {
			stop()
			fmt.Println("could not start tunnel", err)
			return
		}
		defer stop()

		publicURL := strings.TrimSpace(url)
		localURL := "http://localhost:" + port

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

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Run 'mytunnel help' to see available commands.")
	}
}
