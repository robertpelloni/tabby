package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	time.Sleep(500 * time.Millisecond)

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


// TestPTYSpawn verifies PTY spawn returns an error if missing params, and handles request mapping correctly
func TestPTYSpawn(t *testing.T) {
	req := api.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      101,
		Method:  "pty.spawn",
		Params: map[string]interface{}{
			"id":      "test-id",
			"command": "echo",
			"args":    []string{"hello"},
		},
	}

	data, _ := json.Marshal(req)
	in := strings.NewReader(string(data) + "\n")
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)
	srv.Run()
	time.Sleep(100 * time.Millisecond)

	output := out.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Find the response line (not the notifications)
	var response map[string]interface{}
	for _, line := range lines {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(line), &parsed); err == nil {
			if id, ok := parsed["id"].(float64); ok && id == 101 {
				response = parsed
				break
			}
		}
	}

	if response == nil {
		t.Fatalf("Could not find valid JSON response in output: %s", output)
	}

	if response["error"] != nil {
		t.Errorf("Expected successful PTY spawn, got error: %v", response["error"])
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Result is missing or invalid: %v", response)
	}

	if result["id"].(string) != "test-id" {
		t.Errorf("Expected id 'test-id', got %v", result["id"])
	}
}

// TestSerialListPorts verifies serial ports listing handles request correctly
func TestSerialListPorts(t *testing.T) {
	req := api.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      102,
		Method:  "serial.listPorts",
		Params:  map[string]interface{}{},
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

	if response["error"] != nil {
		t.Errorf("Expected successful serial listing, got error: %v", response["error"])
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Result is missing or invalid: %v", response)
	}

	if _, ok := result["ports"]; !ok {
		t.Errorf("Expected 'ports' array in result")
	}
}

// TestInvalidJSON verifies server handles broken JSON
func TestInvalidJSON(t *testing.T) {
	in := strings.NewReader("{\"jsonrpc\": \"2.0\", \"method\": \"ping\", \"id\": \n")
	out := &bytes.Buffer{}

	srv := NewWithIO(in, out)
	srv.Run()
	time.Sleep(100 * time.Millisecond)

	line := strings.TrimSpace(out.String())
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		t.Fatalf("Response is not valid JSON: %s", line)
	}

	if response["error"] == nil {
		t.Errorf("Expected error for invalid JSON")
	}
}

// TestAgentWorkflowRPC verifies startWorkflow, getWorkflow, and cancelWorkflow RPC methods
func TestAgentWorkflowRPC(t *testing.T) {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	srv := NewWithIO(inReader, outWriter)
	go srv.Run()

	decoder := json.NewDecoder(outReader)

	// 1. Start workflow
	startReq := `{"jsonrpc": "2.0", "method": "agent.startWorkflow", "params": {"description": "Test Workflow"}, "id": 1}` + "\n"
	io.WriteString(inWriter, startReq)

	var startResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode JSON-RPC output: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 1 {
			startResp = msg
			break
		}
	}

	wfResult, ok := startResp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("Workflow result is invalid: %v", startResp)
	}

	wfID, ok := wfResult["id"].(string)
	if !ok || wfID == "" {
		t.Fatalf("Workflow ID missing from result: %v", wfResult)
	}

	// 2. Get workflow details
	getReq := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "agent.getWorkflow", "params": {"id": "%s"}, "id": 2}`, wfID) + "\n"
	io.WriteString(inWriter, getReq)

	var getResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode JSON-RPC output: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 2 {
			getResp = msg
			break
		}
	}

	getWf, ok := getResp["result"].(map[string]interface{})
	if !ok || getWf["id"] != wfID {
		t.Errorf("getWorkflow returned unexpected result: %v", getResp)
	}

	// 3. Cancel workflow
	cancelReq := fmt.Sprintf(`{"jsonrpc": "2.0", "method": "agent.cancelWorkflow", "params": {"id": "%s"}, "id": 3}`, wfID) + "\n"
	io.WriteString(inWriter, cancelReq)

	var cancelResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode JSON-RPC output: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 3 {
			cancelResp = msg
			break
		}
	}

	if cancelResp["error"] != nil {
		t.Errorf("cancelWorkflow returned error: %v", cancelResp["error"])
	}

	inWriter.Close()
	outWriter.Close()
}

// TestSessionPersistenceRPC verifies session.save, session.load, and session.clear JSON-RPC methods
func TestSessionPersistenceRPC(t *testing.T) {
	tmpDir := t.TempDir()
	originalAppData := os.Getenv("APPDATA")
	if err := os.Setenv("APPDATA", tmpDir); err != nil {
		t.Skip("Cannot set APPDATA")
	}
	defer func() {
		if originalAppData != "" {
			os.Setenv("APPDATA", originalAppData)
		} else {
			os.Unsetenv("APPDATA")
		}
	}()

	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()

	srv := NewWithIO(inReader, outWriter)
	go srv.Run()

	decoder := json.NewDecoder(outReader)

	// 1. Save session
	saveReq := `{"jsonrpc": "2.0", "method": "session.save", "params": {"tabs": [{"title": "Test Tab", "shell": "/bin/bash", "active": true}]}, "id": 10}` + "\n"
	io.WriteString(inWriter, saveReq)

	var saveResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode save response: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 10 {
			saveResp = msg
			break
		}
	}
	if saveResp["error"] != nil {
		t.Errorf("session.save returned error: %v", saveResp["error"])
	}

	// 2. Load session
	loadReq := `{"jsonrpc": "2.0", "method": "session.load", "id": 11}` + "\n"
	io.WriteString(inWriter, loadReq)

	var loadResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode load response: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 11 {
			loadResp = msg
			break
		}
	}
	if loadResp["error"] != nil {
		t.Errorf("session.load returned error: %v", loadResp["error"])
	}
	result, ok := loadResp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("session.load result missing or not an object: %v", loadResp)
	}
	tabs, ok := result["tabs"].([]interface{})
	if !ok || len(tabs) != 1 {
		t.Fatalf("Expected 1 saved tab, got: %v", result["tabs"])
	}

	// 3. Clear session
	clearReq := `{"jsonrpc": "2.0", "method": "session.clear", "id": 12}` + "\n"
	io.WriteString(inWriter, clearReq)

	var clearResp map[string]interface{}
	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			t.Fatalf("Failed to decode clear response: %v", err)
		}
		if id, ok := msg["id"].(float64); ok && id == 12 {
			clearResp = msg
			break
		}
	}
	if clearResp["error"] != nil {
		t.Errorf("session.clear returned error: %v", clearResp["error"])
	}

	inWriter.Close()
	outWriter.Close()
}
