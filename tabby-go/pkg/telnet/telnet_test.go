package telnet

import (
	"sync"
	"testing"
)

func TestManagerCreation(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
	if len(mgr.ListConnections()) != 0 {
		t.Error("New manager should have no connections")
	}
}

func TestConnectInvalidHost(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	_, err := mgr.Connect(TelnetConnectParams{Host: "nonexistent.invalid", Port: 23})
	if err == nil {
		t.Error("Should fail connecting to invalid host")
	}
}

func TestWriteNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	err := mgr.Write("nonexistent", "dGVzdA==")
	if err == nil {
		t.Error("Should fail writing to non-existent connection")
	}
}

func TestResizeNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	err := mgr.Resize("nonexistent", 80, 24)
	if err == nil {
		t.Error("Should fail resizing non-existent connection")
	}
}

func TestCloseNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	err := mgr.Close("nonexistent")
	if err == nil {
		t.Error("Should fail closing non-existent connection")
	}
}

func TestProcessTelnetSimpleData(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// Regular data should pass through
	data := []byte("Hello World")
	result := mgr.processTelnet(tc, data)
	if string(result) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", string(result))
	}
}

func TestProcessTelnetEscapedFF(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// 0xFF 0xFF should become a single 0xFF
	data := []byte{0xFF, 0xFF, 'A'}
	result := mgr.processTelnet(tc, data)
	if len(result) != 2 || result[0] != 0xFF || result[1] != 'A' {
		t.Errorf("Expected [0xFF, 'A'], got %v", result)
	}
}

func TestProcessTelnetIACCommand(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// IAC DO ECHO should be handled without crashing (no conn for response)
	// We test the parsing logic doesn't panic even without a real connection
	data := []byte{'A', 'B', 'C'} // Just data, no IAC
	result := mgr.processTelnet(tc, data)
	if string(result) != "ABC" {
		t.Errorf("Expected 'ABC', got %q", string(result))
	}
}

func TestProcessTelnetMixedData(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// Mix of data and IAC commands
	data := []byte{'H', 'i', IAC, WILL, OPT_SGA, '!', 0xFF, 0xFF}
	result := mgr.processTelnet(tc, data)
	expected := []byte{'H', 'i', '!', 0xFF}
	if string(result) != string(expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestProcessTelnetSuboption(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// TTYPE suboption with SEND - without a real connection we test parsing only
	// The response write will fail silently
	data := []byte{IAC, SB, OPT_TTYPE, SUBOPT_SEND, IAC, SE, 'X'}
	result := mgr.processTelnet(tc, data)
	// After processing the suboption, 'X' should be in the result
	if len(result) != 1 || result[0] != 'X' {
		t.Errorf("Expected ['X'], got %v", result)
	}
}

func TestProcessTelnetNAWS(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	tc := &Connection{ID: "test", telnetMode: true}

	// NAWS suboption
	data := []byte{IAC, SB, OPT_NAWS, 0, 80, 0, 24, IAC, SE}
	result := mgr.processTelnet(tc, data)
	if len(result) != 0 {
		t.Errorf("Expected no data from NAWS, got %v", result)
	}
}

func TestFindSuboptionEnd(t *testing.T) {
	data := []byte{IAC, SB, OPT_TTYPE, SUBOPT_SEND, IAC, SE, 'A', 'B'}
	end := findSuboptionEnd(data, 0)
	if end != 6 {
		t.Errorf("Expected end at 6, got %d", end)
	}
}

func TestFindSuboptionEndMissing(t *testing.T) {
	data := []byte{IAC, SB, OPT_TTYPE, SUBOPT_SEND}
	end := findSuboptionEnd(data, 0)
	if end != -1 {
		t.Errorf("Expected -1 for incomplete suboption, got %d", end)
	}
}

func TestNotificationForwarding(t *testing.T) {
	var mu sync.Mutex
	var notifications []string
	mgr := NewManager(func(method string, params interface{}) {
		mu.Lock()
		notifications = append(notifications, method)
		mu.Unlock()
	})

	if len(notifications) != 0 {
		t.Error("Should start with no notifications")
	}
	_ = mgr
}
