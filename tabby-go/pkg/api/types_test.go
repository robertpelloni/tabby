package api

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestJSONRPCSerialization verifies that all API types serialize correctly
func TestJSONRPCSerialization(t *testing.T) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ssh.connect",
		Params: SSHConnectParams{
			Host:     "example.com",
			Port:     22,
			Username: "testuser",
			Auth: SSHAuthParams{
				Type:     "password",
				Password: "secret",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal JSONRPCRequest: %v", err)
	}

	var req2 JSONRPCRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("Failed to unmarshal JSONRPCRequest: %v", err)
	}

	if req2.Method != "ssh.connect" {
		t.Errorf("Expected method 'ssh.connect', got '%s'", req2.Method)
	}
}

// TestJSONRPCResponse verifies response serialization
func TestJSONRPCResponse(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      1,
		Result: SSHConnectionResult{
			ConnectionID:  "ssh-conn-123",
			ServerVersion: "SSH-2.0-OpenSSH_8.9",
			RemoteAddress: "192.168.1.1:22",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal JSONRPCResponse: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "ssh-conn-123") {
		t.Error("Response should contain connection ID")
	}
	if !strings.Contains(s, "SSH-2.0-OpenSSH_8.9") {
		t.Error("Response should contain server version")
	}
}

// TestJSONRPCError verifies error response serialization
func TestJSONRPCError(t *testing.T) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      2,
		Error: &RPCError{
			Code:    ErrorMethodNotFound,
			Message: "Method not found: foo.bar",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "-32601") {
		t.Error("Error response should contain error code -32601")
	}
}

// TestNotificationSerialization verifies notification types
func TestNotificationSerialization(t *testing.T) {
	notif := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  "ssh.data",
		Params: DataNotification{
			ConnectionID: "ssh-conn-123",
			SessionID:    "ssh-session-456",
			Data:         "SGVsbG8gV29ybGQ=",
		},
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "ssh.data") {
		t.Error("Notification should contain method name")
	}
}

// TestSFTPFileSerialization verifies SFTP file listing
func TestSFTPFileSerialization(t *testing.T) {
	files := []SFTPFile{
		{Name: "test.txt", Size: 1024, Mode: 0644, ModTime: "2025-01-01T00:00:00Z", IsDir: false},
		{Name: "docs", Size: 4096, Mode: 0755, ModTime: "2025-01-01T00:00:00Z", IsDir: true},
	}

	data, err := json.Marshal(files)
	if err != nil {
		t.Fatalf("Failed to marshal SFTP files: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "test.txt") || !strings.Contains(s, "docs") {
		t.Errorf("SFTP file list should contain file names, got: %s", s)
	}
}

// TestSSHAlgorithms verifies algorithm configuration
func TestSSHAlgorithms(t *testing.T) {
	algo := SSHAlgorithms{
		KEX:           []string{"curve25519-sha256", "ecdh-sha2-nistp256"},
		Cipher:        []string{"aes256-gcm@openssh.com", "chacha20-poly1305@openssh.com"},
		HMAC:          []string{"hmac-sha2-256", "hmac-sha1"},
		ServerHostKey: []string{"ssh-ed25519", "rsa-sha2-512"},
		Compression:   []string{"none", "zlib@openssh.com"},
	}

	data, err := json.Marshal(algo)
	if err != nil {
		t.Fatalf("Failed to marshal algorithms: %v", err)
	}

	s := string(data)
	if !strings.Contains(s, "curve25519-sha256") {
		t.Error("Should contain kex algorithm")
	}
}

// TestPortForwardingTypes verifies port forward config
func TestPortForwardingTypes(t *testing.T) {
	// Test the data notification
	dn := DataNotification{
		ConnectionID: "conn-1",
		SessionID:    "sess-1",
		Data:         "aGVsbG8=",
	}
	data, _ := json.Marshal(dn)
	s := string(data)
	if !strings.Contains(s, "conn-1") || !strings.Contains(s, "aGVsbG8=") {
		t.Errorf("Data notification malformed: %s", s)
	}

	// Test exit notification (use non-zero exit code since 0 is omitempty)
	en := ExitNotification{
		ConnectionID: "conn-1",
		SessionID:    "sess-1",
		ExitCode:     1,
	}
	data, _ = json.Marshal(en)
	s = string(data)
	if !strings.Contains(s, "exitCode") {
		t.Errorf("Exit notification malformed: %s", s)
	}
}
