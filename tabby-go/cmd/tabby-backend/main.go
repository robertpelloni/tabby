// cmd/tabby-backend is the main entry point for the Tabby Go backend.
//
// The Go backend is spawned by the Electron app as a child process and
// communicates via JSON-RPC 2.0 over stdin/stdout.
//
// Usage:
//
//	tabby-backend          # Start in stdio mode (default)
//	tabby-backend --listen :9090  # Start in TCP mode
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/robertpelloni/tabby/tabby-go/internal/server"
)

func main() {
	listenAddr := flag.String("listen", "", "TCP address to listen on (default: stdio mode)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Println("tabby-backend v1.0.231-nightly.0")
		os.Exit(0)
	}

	log.SetPrefix("[tabby-go] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	srv := server.New()

	if *listenAddr != "" {
		// TCP mode — not yet implemented, would need net.Listener wrapper
		log.Fatalf("TCP mode not yet implemented, use stdio mode")
	}

	// Default: stdio mode
	log.Println("Starting Tabby Go backend (stdio mode)")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
