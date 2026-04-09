// tabby-native — Native BTK-based Tabby terminal emulator
//
// This is an experimental native UI build of Tabby using the BTK
// widget toolkit instead of Electron. It provides a lighter-weight
// alternative with faster startup and lower memory usage.
//
// Build (requires BTK to be compiled):
//
//	go build -tags btk -o tabby-native ./cmd/tabby-native
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/robertpelloni/tabby/tabby-go/pkg/nativeapp"
)

func main() {
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Println("tabby-native v1.0.231-nightly.2 (BTK)")
		os.Exit(0)
	}

	app := nativeapp.NewNativeApp()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
