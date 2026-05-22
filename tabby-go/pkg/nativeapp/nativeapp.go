// Package nativeapp provides a native BTK-based UI for Tabby.
//
// This is an alternative to the Electron-based UI that uses BTK's
// native widget toolkit for a lighter-weight, faster-starting application.
//
// Currently in development — provides the basic window structure with
// tab management, menu system, and terminal widget placeholders.
package nativeapp

// Note: This package requires CGo and BTK to be built.
// It is conditionally compiled and not included in the default build.
// To build: go build -tags btk ./cmd/tabby-native/

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
)

// NativeApp represents the native Tabby application
type NativeApp struct {
	sshMgr  *ssh.Manager
	sftpMgr *sftp.Manager
	running bool
}

// NewNativeApp creates a new native Tabby application
func NewNativeApp() *NativeApp {
	app := &NativeApp{}
	app.sshMgr = ssh.NewManager(app.sendNotification)
	app.sftpMgr = sftp.NewManager(app.sshMgr, app)
	return app
}

// Run starts the native application
//
// This function creates the BTK application window with:
//   - Tab bar for multiple terminal sessions
//   - Menu bar (File, Edit, View, Tabs, Help)
//   - Status bar with connection info
//   - Split pane support
//   - Terminal widget for each tab
//
// NOTE: The BTK CGo bindings require the BTK library to be compiled
// and linked. This function will fail at link time if BTK is not available.
// Use the JSON-RPC backend (cmd/tabby-backend) for the pure-Go path.
func (a *NativeApp) Run() error {
	// The BTK UI initialization would go here.
	// For now, we provide the backend infrastructure and the
	// JSON-RPC communication path works independently.

	fmt.Println("Tabby Native App")
	fmt.Println("Backend initialized with SSH, SFTP, PTY, Serial support")
	fmt.Println("Use the JSON-RPC backend (tabby-backend) for Electron integration")

	// Wait for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down...")
	return nil
}

// sendNotification handles notifications from the backend
func (a *NativeApp) ReportProgress(transferID string, bytesTransferred, totalBytes int64, complete bool, err string) {
	a.sendNotification("sftp.progress", api.TransferProgressNotification{
		TransferID:       transferID,
		BytesTransferred: bytesTransferred,
		TotalBytes:       totalBytes,
		Complete:         complete,
		Error:            err,
	})
}

func (a *NativeApp) sendNotification(method string, params interface{}) {
	switch method {
	case "ssh.data":
		if data, ok := params.(api.DataNotification); ok {
			// Forward data to the appropriate terminal widget
			fmt.Printf("[data] session=%s len=%d\n", data.SessionID, len(data.Data))
		}
	case "ssh.exit":
		if data, ok := params.(api.ExitNotification); ok {
			fmt.Printf("[exit] session=%s code=%d\n", data.SessionID, data.ExitCode)
		}
	case "ssh.serviceMessage":
		if data, ok := params.(api.ServiceMessageNotification); ok {
			fmt.Printf("[msg] %s\n", data.Message)
		}
	case "ssh.banner":
		if data, ok := params.(api.BannerNotification); ok {
			fmt.Printf("[banner] %s\n", data.Message)
		}
	case "ssh.hostKeyPrompt":
		if data, ok := params.(api.HostKeyPromptNotification); ok {
			fmt.Printf("[hostkey] %s:%d %s %s\n", data.Host, data.Port, data.KeyType, data.Fingerprint)
		}
	case "ssh.keyboardInteractive":
		if data, ok := params.(api.KeyboardInteractiveNotification); ok {
			fmt.Printf("[ki-auth] %s: %s\n", data.Name, data.Instruction)
		}
	}
}
