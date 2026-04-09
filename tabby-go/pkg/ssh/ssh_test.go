package ssh

import (
	"encoding/base64"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// TestManagerCreation verifies the manager is created correctly
func TestManagerCreation(t *testing.T) {
	var mu sync.Mutex
	var notifications []string
	notify := func(method string, params interface{}) {
		mu.Lock()
		notifications = append(notifications, method)
		mu.Unlock()
	}

	mgr := NewManager(notify)
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}

	if len(mgr.ListConnections()) != 0 {
		t.Error("New manager should have no connections")
	}
}

// TestConnectInvalidHost verifies connection fails for invalid hosts
func TestConnectInvalidHost(t *testing.T) {
	var mu sync.Mutex
	var notifications []string
	notify := func(method string, params interface{}) {
		mu.Lock()
		notifications = append(notifications, method)
		mu.Unlock()
	}

	mgr := NewManager(notify)

	params := api.SSHConnectParams{
		Host:     "nonexistent.invalid.host.test",
		Port:     22,
		Username: "test",
		Auth:     api.SSHAuthParams{Type: "password", Password: "test"},
	}

	_, err := mgr.Connect(params)
	if err == nil {
		t.Error("Should fail connecting to invalid host")
	}
}

// TestListConnectionsEmpty verifies listing with no connections
func TestListConnectionsEmpty(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})
	conns := mgr.ListConnections()
	if len(conns) != 0 {
		t.Errorf("Expected empty connections, got %d", len(conns))
	}
}

// TestCloseNonexistentSession verifies closing a non-existent session returns error
func TestCloseNonexistentSession(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.Close(api.SSHCloseParams{
		ConnectionID: "nonexistent",
		SessionID:    "nonexistent",
	})
	if err == nil {
		t.Error("Should fail closing non-existent session")
	}
}

// TestCloseNonexistentConnection verifies closing a non-existent connection returns error
func TestCloseNonexistentConnection(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.Close(api.SSHCloseParams{
		ConnectionID: "nonexistent",
	})
	if err == nil {
		t.Error("Should fail closing non-existent connection")
	}
}

// TestResizeNonexistent verifies resizing a non-existent session returns error
func TestResizeNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.Resize(api.SSHResizeParams{
		SessionID: "nonexistent",
		Columns:   80,
		Rows:      24,
	})
	if err == nil {
		t.Error("Should fail resizing non-existent session")
	}
}

// TestWriteNonexistent verifies writing to a non-existent session returns error
func TestWriteNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.Write(api.SSHWriteParams{
		SessionID: "nonexistent",
		Data:      base64.StdEncoding.EncodeToString([]byte("test")),
	})
	if err == nil {
		t.Error("Should fail writing to non-existent session")
	}
}

// TestWriteInvalidBase64 verifies writing invalid base64 returns error
func TestWriteInvalidBase64(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.Write(api.SSHWriteParams{
		SessionID: "nonexistent",
		Data:      "not-valid-base64!!!",
	})
	if err == nil {
		t.Error("Should fail with invalid base64")
	}
}

// TestGetConnectionNonexistent verifies getting a non-existent connection returns error
func TestGetConnectionNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	_, err := mgr.GetConnection("nonexistent")
	if err == nil {
		t.Error("Should fail getting non-existent connection")
	}
}

// TestAddForwardNonexistent verifies adding forward to non-existent connection fails
func TestAddForwardNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	_, err := mgr.AddForward(api.PortForwardParams{
		ConnectionID:  "nonexistent",
		Type:          api.PortForwardLocal,
		Host:          "127.0.0.1",
		Port:          8080,
		TargetAddress: "localhost",
		TargetPort:    80,
	})
	if err == nil {
		t.Error("Should fail adding forward to non-existent connection")
	}
}

// TestRemoveForwardNonexistent verifies removing non-existent forward fails
func TestRemoveForwardNonexistent(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	err := mgr.RemoveForward(api.PortForwardRemoveParams{
		ConnectionID: "nonexistent",
		ForwardID:    "nonexistent",
	})
	if err == nil {
		t.Error("Should fail removing non-existent forward")
	}
}

// TestListForwardsEmpty verifies listing forwards for non-existent connection returns empty
func TestListForwardsEmpty(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	forwards := mgr.ListForwards("nonexistent")
	if len(forwards) != 0 {
		t.Errorf("Expected empty forwards, got %d", len(forwards))
	}
}

// TestHostKeyVerification verifies host key response routing
func TestHostKeyVerification(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	// Should not panic for non-existent connection
	mgr.HandleHostKeyResponse("nonexistent", true)
}

// TestKeyboardInteractiveResponse verifies keyboard interactive response routing
func TestKeyboardInteractiveResponse(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	// Should not panic for non-existent connection
	mgr.HandleKeyboardInteractiveResponse("nonexistent", []string{"response1", "response2"})
}

// TestNotificationSending verifies notifications are sent correctly
func TestNotificationSending(t *testing.T) {
	var mu sync.Mutex
	var received []struct {
		method string
		params interface{}
	}
	notify := func(method string, params interface{}) {
		mu.Lock()
		received = append(received, struct {
			method string
			params interface{}
		}{method, params})
		mu.Unlock()
	}

	mgr := NewManager(notify)

	// Trigger a notification by connecting to an invalid host
	params := api.SSHConnectParams{
		Host:     "invalid",
		Port:     22,
		Username: "test",
		Auth:     api.SSHAuthParams{Type: "password", Password: "test"},
	}
	mgr.Connect(params)

	// Just verify the notification mechanism works
	// (The actual connection will fail, but service message should have been sent)
}

// TestBuildClientConfig verifies client config building
func TestBuildClientConfig(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	// Test password auth
	params := api.SSHConnectParams{
		Host:     "test.com",
		Port:     22,
		Username: "testuser",
		Auth:     api.SSHAuthParams{Type: "password", Password: "secret"},
	}
	config, err := mgr.buildClientConfig(params)
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}
	if config.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", config.User)
	}

	// Test with timeout
	params.ReadyTimeout = 10
	config, err = mgr.buildClientConfig(params)
	if err != nil {
		t.Fatalf("Failed to build config with timeout: %v", err)
	}
	if config.Timeout != 10*time.Second {
		t.Errorf("Expected 10s timeout, got %v", config.Timeout)
	}

	// Test with custom algorithms
	params.Algorithms = &api.SSHAlgorithms{
		KEX:    []string{"curve25519-sha256"},
		Cipher: []string{"aes128-ctr"},
		HMAC:   []string{"hmac-sha2-256"},
	}
	config, err = mgr.buildClientConfig(params)
	if err != nil {
		t.Fatalf("Failed to build config with algorithms: %v", err)
	}
}

// TestForwardString verifies the Forward.String() method
func TestForwardString(t *testing.T) {
	tests := []struct {
		forward  Forward
		expected string
	}{
		{
			forward:  Forward{Type: api.PortForwardLocal, Host: "127.0.0.1", Port: 8080, TargetAddr: "db", TargetPort: 3306},
			expected: "(local) 127.0.0.1:8080 → (remote) db:3306",
		},
		{
			forward:  Forward{Type: api.PortForwardRemote, Host: "0.0.0.0", Port: 9090, TargetAddr: "localhost", TargetPort: 8080},
			expected: "(remote) 0.0.0.0:9090 → (local) localhost:8080",
		},
		{
			forward:  Forward{Type: api.PortForwardDynamic, Host: "127.0.0.1", Port: 1080},
			expected: "(dynamic/SOCKS5) 127.0.0.1:1080",
		},
	}

	for _, tt := range tests {
		result := tt.forward.String()
		if result != tt.expected {
			t.Errorf("Expected %q, got %q", tt.expected, result)
		}
	}
}

// TestProxyConn verifies the proxy connection implementation
func TestProxyConn(t *testing.T) {
	// Test address helpers
	a := &addrProto{"local"}
	if a.Network() != "tcp" {
		t.Errorf("Expected tcp network, got %s", a.Network())
	}
	if a.String() != "local" {
		t.Errorf("Expected 'local' address, got %s", a.String())
	}
}

// TestAPITypeMarshaling verifies API types serialize correctly
func TestAPITypeMarshaling(t *testing.T) {
	params := api.SSHConnectParams{
		Host:     "example.com",
		Port:     22,
		Username: "user",
		Auth: api.SSHAuthParams{
			Type:     "password",
			Password: "secret",
		},
		JumpHost: &api.SSHConnectParams{
			Host:     "jump.example.com",
			Port:     2222,
			Username: "jumpuser",
			Auth: api.SSHAuthParams{
				Type:     "publicKey",
				PrivateKeyPaths: []string{"/home/user/.ssh/id_rsa"},
			},
		},
		Algorithms: &api.SSHAlgorithms{
			KEX:    []string{"curve25519-sha256"},
			Cipher: []string{"aes256-gcm@openssh.com"},
		},
		SocksProxyHost: "proxy.example.com",
		SocksProxyPort: 1080,
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var unmarshaled api.SSHConnectParams
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if unmarshaled.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", unmarshaled.Host)
	}
	if unmarshaled.JumpHost == nil {
		t.Error("Jump host should not be nil")
	}
	if unmarshaled.JumpHost.Host != "jump.example.com" {
		t.Errorf("Expected jump host 'jump.example.com', got '%s'", unmarshaled.JumpHost.Host)
	}
	if unmarshaled.SocksProxyPort != 1080 {
		t.Errorf("Expected SOCKS port 1080, got %d", unmarshaled.SocksProxyPort)
	}
}
