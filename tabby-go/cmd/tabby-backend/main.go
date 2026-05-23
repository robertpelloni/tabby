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
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/robertpelloni/tabby/tabby-go/internal/server"
)

func main() {
	listenAddr := flag.String("listen", "", "TCP address to listen on (default: stdio mode)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Println("tabby-backend v1.0.231-nightly.19")
		os.Exit(0)
	}

	home, err := os.UserHomeDir()
	if err == nil {
		logFile, err := os.OpenFile(filepath.Join(home, "tabby-backend.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			multiWriter := io.MultiWriter(os.Stderr, logFile)
			log.SetOutput(multiWriter)
		} else {
			log.Printf("Failed to open log file: %v", err)
		}
	} else {
		log.Printf("Failed to get home dir for logging: %v", err)
	}

	log.SetPrefix("[tabby-go] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Release: "tabby-backend@v1.0.231-nightly.19",
		EnableTracing: true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Printf("Sentry initialization failed: %v", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

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
