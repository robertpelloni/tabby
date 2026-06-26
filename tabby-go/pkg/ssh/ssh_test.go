package ssh

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// TestMockDataForwarding verifies that the manager correctly processes binary keystrokes/responses.
func TestMockDataForwarding(t *testing.T) {
	var receivedBase64 string

	m := NewManager(func(method string, payload interface{}) {
		if method == "ssh.data" {
			if notif, ok := payload.(api.DataNotification); ok {
				receivedBase64 = notif.Data
			}
		}
	})

	buf := bytes.NewBuffer([]byte("keystroke_data"))

	// Invoke forwardOutput manually
	go m.forwardOutput("mock-conn", "mock-session", buf)

	time.Sleep(100 * time.Millisecond) // wait for async buffer forward

	expected := base64.StdEncoding.EncodeToString([]byte("keystroke_data"))
	if receivedBase64 != expected {
		t.Fatalf("Expected base64 %s, got %s", expected, receivedBase64)
	}
}
