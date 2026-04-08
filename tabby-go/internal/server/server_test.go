package server

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

// TestMethodNotFound verifies unknown methods return proper error
func TestMethodNotFound(t *testing.T) {
	req := api.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      99,
		Method:  "nonexistent.method",
	}

	data, _ := json.Marshal(req)
	in := strings.NewReader(string(data) + "\n")
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)

	// Run the server (it reads one line, processes it, then exits on EOF)
	err := srv.Run()
	if err != nil {
		t.Fatalf("Server error: %v", err)
	}

	// Give the goroutine time to write the response
	time.Sleep(100 * time.Millisecond)

	if out.Len() == 0 {
		t.Fatal("Expected output from server")
	}

	line := strings.TrimSpace(out.String())
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		t.Fatalf("Response is not valid JSON: %s", line)
	}

	if response["id"].(float64) != 99 {
		t.Errorf("Expected id 99, got %v", response["id"])
	}

	if response["error"] == nil {
		t.Error("Expected error in response for unknown method")
	}

	errObj := response["error"].(map[string]interface{})
	if errObj["code"].(float64) != float64(api.ErrorMethodNotFound) {
		t.Errorf("Expected error code %d, got %v", api.ErrorMethodNotFound, errObj["code"])
	}
}

// TestPing verifies the ping method works
func TestPing(t *testing.T) {
	req := api.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	}

	data, _ := json.Marshal(req)
	in := strings.NewReader(string(data) + "\n")
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)
	srv.Run()
	time.Sleep(100 * time.Millisecond)

	line := strings.TrimSpace(out.String())
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		t.Fatalf("Response is not valid JSON: %s", line)
	}

	result := response["result"].(map[string]interface{})
	if result["status"] != "ok" {
		t.Errorf("Expected status ok, got %v", result["status"])
	}
}

// TestSSHConnectInvalidParams verifies SSH connect with invalid params returns error
func TestSSHConnectInvalidParams(t *testing.T) {
	req := api.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      100,
		Method:  "ssh.connect",
		Params: map[string]interface{}{
			"host": "nonexistent.invalid",
			"port": 22,
			"user": "test",
			"auth": map[string]interface{}{
				"type":     "password",
				"password": "test",
			},
		},
	}

	data, _ := json.Marshal(req)
	in := strings.NewReader(string(data) + "\n")
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)
	srv.Run()
	time.Sleep(100 * time.Millisecond)

	line := strings.TrimSpace(out.String())
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		t.Fatalf("Response is not valid JSON: %s, raw: %q", line, out.String())
	}

	// Should get an error because the host doesn't exist
	if response["error"] == nil {
		t.Error("Expected error connecting to nonexistent host")
	}
}

// TestConcurrentRequests verifies the server handles multiple requests
func TestConcurrentRequests(t *testing.T) {
	var input strings.Builder
	for i := 1; i <= 5; i++ {
		req := api.JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      i,
			Method:  "ping",
		}
		data, _ := json.Marshal(req)
		input.Write(data)
		input.WriteByte('\n')
	}

	in := strings.NewReader(input.String())
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)
	srv.Run()
	time.Sleep(200 * time.Millisecond)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// We might get the responses in any order due to goroutine scheduling
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 response lines, got %d", len(lines))
	}
}
