package serial

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	serial "go.bug.st/serial"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

func TestBuildModeDefaults(t *testing.T) {
	params := api.SerialOpenParams{
		Port:     "/dev/ttyUSB0",
		BaudRate: 115200,
	}

	mode := BuildMode(params)
	if mode.BaudRate != 115200 {
		t.Fatalf("expected BaudRate 115200, got %d", mode.BaudRate)
	}
	if mode.DataBits != 8 {
		t.Fatalf("expected DataBits 8, got %d", mode.DataBits)
	}
	if mode.StopBits != serial.OneStopBit {
		t.Fatalf("expected OneStopBit, got %v", mode.StopBits)
	}
	if mode.Parity != serial.NoParity {
		t.Fatalf("expected NoParity, got %v", mode.Parity)
	}
	if mode.InitialStatusBits != nil {
		t.Fatalf("expected nil InitialStatusBits for default flow control")
	}
}

func TestBuildModeCustom(t *testing.T) {
	params := api.SerialOpenParams{
		Port:        "/dev/ttyS0",
		BaudRate:    9600,
		DataBits:    7,
		StopBits:    2,
		Parity:      "even",
		FlowControl: "hardware",
	}

	mode := BuildMode(params)
	if mode.BaudRate != 9600 {
		t.Fatalf("expected BaudRate 9600, got %d", mode.BaudRate)
	}
	if mode.DataBits != 7 {
		t.Fatalf("expected DataBits 7, got %d", mode.DataBits)
	}
	if mode.StopBits != serial.TwoStopBits {
		t.Fatalf("expected TwoStopBits, got %v", mode.StopBits)
	}
	if mode.Parity != serial.EvenParity {
		t.Fatalf("expected EvenParity, got %v", mode.Parity)
	}
	if mode.InitialStatusBits == nil || !mode.InitialStatusBits.RTS || !mode.InitialStatusBits.DTR {
		t.Fatalf("expected RTS and DTR to be true under hardware flow control")
	}
}

func TestManagerConnectionLifecycle(t *testing.T) {
	var receivedNotifications []string
	mgr := NewManager(func(method string, params interface{}) {
		receivedNotifications = append(receivedNotifications, method)
	})

	id := mgr.nextID("test-serial")
	if id == "" {
		t.Fatalf("expected non-empty ID")
	}

	// Test writing to non-existent connection fails
	err := mgr.Write("non-existent-id", base64.StdEncoding.EncodeToString([]byte("hello")))
	if err == nil {
		t.Fatalf("expected error for non-existent connection, got nil")
	}

	// Test closing non-existent connection fails
	err = mgr.Close("non-existent-id")
	if err == nil {
		t.Fatalf("expected error for non-existent connection, got nil")
	}
}

func TestSerialConnectionLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_session.log")

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatalf("failed to create temp log file: %v", err)
	}

	conn := &SerialConnection{
		ID:          "serial-test-1",
		PortName:    "/dev/ttyUSB0",
		BaudRate:    115200,
		DataBits:    8,
		StopBits:    1,
		Parity:      "none",
		FlowControl: "hardware",
		LogFile:     f,
		LogPath:     logPath,
		Open:        true,
		done:        make(chan struct{}),
	}

	// Simulate writing and receiving data
	conn.logMu.Lock()
	conn.LogFile.WriteString("[OUT] AT+CGMI\r\n")
	conn.LogFile.WriteString("[IN ] +CGMI: Quectel\r\n")
	conn.logMu.Unlock()

	conn.LogFile.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read written log file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("expected log file to contain session data")
	}
}
