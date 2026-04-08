// Package server implements the JSON-RPC 2.0 server that communicates with
// the Electron frontend over stdin/stdout or a TCP socket.
//
// Protocol:
// - The Electron app spawns this binary as a child process
// - Communication happens via JSON-RPC 2.0 over stdin/stdout
// - Each line is a complete JSON-RPC message
// - The server sends notifications for async events (data, exit, etc.)
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/serial"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
)

// Server is the JSON-RPC server for Tabby's Go backend
type Server struct {
	sshMgr    *ssh.Manager
	sftpMgr   *sftp.Manager
	ptyMgr    *pty.Manager
	serialMgr *serial.Manager
	reader    *bufio.Reader
	writer    io.Writer
	mu        sync.Mutex
	running   bool
}

// New creates a new Server
func New() *Server {
	s := &Server{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
	s.sshMgr = ssh.NewManager(s.sendNotification)
	s.sftpMgr = sftp.NewManager(s.sshMgr)
	s.ptyMgr = pty.NewManager(s.sendNotification)
	s.serialMgr = serial.NewManager(s.sendNotification)
	return s
}

// NewWithIO creates a new Server with custom I/O (for testing)
func NewWithIO(in io.Reader, out io.Writer) *Server {
	s := &Server{
		reader: bufio.NewReader(in),
		writer: out,
	}
	s.sshMgr = ssh.NewManager(s.sendNotification)
	s.sftpMgr = sftp.NewManager(s.sshMgr)
	s.ptyMgr = pty.NewManager(s.sendNotification)
	s.serialMgr = serial.NewManager(s.sendNotification)
	return s
}

// Run starts the server's main loop, reading and processing JSON-RPC messages
func (s *Server) Run() error {
	s.running = true
	log.Println("Tabby Go backend started")

	for s.running {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
				break
			}
			log.Printf("Read error: %v", err)
			continue
		}

		// Trim the newline
		line = line[:len(line)-1]

		// Parse the request
		var req api.JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(0, api.ErrorParseError, "Parse error", nil)
			continue
		}

		// Handle the request
		go s.handleRequest(req)
	}

	return nil
}

// handleRequest dispatches a JSON-RPC request to the appropriate handler
func (s *Server) handleRequest(req api.JSONRPCRequest) {
	var result interface{}
	var err error

	switch req.Method {
	// ---- Lifecycle ----
	case "ping":
		result = map[string]string{"status": "ok"}

	// ---- SSH ----
	case "ssh.connect":
		result, err = s.handleSSHConnect(req.Params)
	case "ssh.startShell":
		result, err = s.handleSSHStartShell(req.Params)
	case "ssh.resize":
		err = s.handleSSHResize(req.Params)
	case "ssh.write":
		err = s.handleSSHWrite(req.Params)
	case "ssh.close":
		err = s.handleSSHClose(req.Params)
	case "ssh.listConnections":
		result = s.sshMgr.ListConnections()

	// ---- SFTP ----
	case "sftp.open":
		result, err = s.handleSFTPOpen(req.Params)
	case "sftp.list":
		result, err = s.handleSFTPList(req.Params)
	case "sftp.download":
		result, err = s.handleSFTPDownload(req.Params)
	case "sftp.upload":
		result, err = s.handleSFTPUpload(req.Params)
	case "sftp.delete":
		err = s.handleSFTPDelete(req.Params)
	case "sftp.rename":
		err = s.handleSFTPRename(req.Params)
	case "sftp.mkdir":
		err = s.handleSFTPMkdir(req.Params)
	case "sftp.stat":
		result, err = s.handleSFTPStat(req.Params)
	case "sftp.close":
		err = s.handleSFTPClose(req.Params)

	// ---- PTY ----
	case "pty.spawn":
		result, err = s.handlePTYSpawn(req.Params)
	case "pty.resize":
		err = s.handlePTYResize(req.Params)
	case "pty.write":
		err = s.handlePTYWrite(req.Params)
	case "pty.kill":
		err = s.handlePTYKill(req.Params)

	// ---- Serial ----
	case "serial.open":
		result, err = s.handleSerialOpen(req.Params)
	case "serial.write":
		err = s.handleSerialWrite(req.Params)
	case "serial.close":
		err = s.handleSerialClose(req.Params)

	default:
		s.sendError(req.ID, api.ErrorMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method), nil)
		return
	}

	if err != nil {
		s.sendError(req.ID, api.ErrorInternal, err.Error(), nil)
		return
	}

	s.sendResult(req.ID, result)
}

// sendResult sends a successful JSON-RPC response
func (s *Server) sendResult(id int, result interface{}) {
	resp := api.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendMessage(resp)
}

// sendError sends an error JSON-RPC response
func (s *Server) sendError(id int, code int, message string, data interface{}) {
	resp := api.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &api.RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.sendMessage(resp)
}

// sendNotification sends a JSON-RPC notification to the client
func (s *Server) sendNotification(method string, params interface{}) {
	notif := api.JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	s.sendMessage(notif)
}

// sendMessage writes a JSON-RPC message as a single line
func (s *Server) sendMessage(msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	s.writer.Write(data)
	s.writer.Write([]byte("\n"))
}

// ---- SSH Handlers ----

func (s *Server) handleSSHConnect(params interface{}) (*api.SSHConnectionResult, error) {
	var p api.SSHConnectParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sshMgr.Connect(p)
}

func (s *Server) handleSSHStartShell(params interface{}) (*api.SSHSessionResult, error) {
	var p api.SSHSessionParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sshMgr.StartShell(p)
}

func (s *Server) handleSSHResize(params interface{}) error {
	var p api.SSHResizeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sshMgr.Resize(p)
}

func (s *Server) handleSSHWrite(params interface{}) error {
	var p api.SSHWriteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sshMgr.Write(p)
}

func (s *Server) handleSSHClose(params interface{}) error {
	var p api.SSHCloseParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sshMgr.Close(p)
}

// ---- SFTP Handlers ----

func (s *Server) handleSFTPOpen(params interface{}) (*api.SFTPOpenResult, error) {
	var p api.SFTPOpenParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Open(p)
}

func (s *Server) handleSFTPList(params interface{}) ([]api.SFTPFile, error) {
	var p api.SFTPListParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.List(p)
}

func (s *Server) handleSFTPDownload(params interface{}) (*api.SFTPTransferResult, error) {
	var p api.SFTPDownloadParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Download(p)
}

func (s *Server) handleSFTPUpload(params interface{}) (*api.SFTPTransferResult, error) {
	var p api.SFTPUploadParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Upload(p)
}

func (s *Server) handleSFTPDelete(params interface{}) error {
	type deleteParams struct {
		SessionID  string `json:"sessionId"`
		RemotePath string `json:"remotePath"`
	}
	var p deleteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Delete(p.SessionID, p.RemotePath)
}

func (s *Server) handleSFTPRename(params interface{}) error {
	type renameParams struct {
		SessionID string `json:"sessionId"`
		OldPath   string `json:"oldPath"`
		NewPath   string `json:"newPath"`
	}
	var p renameParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Rename(p.SessionID, p.OldPath, p.NewPath)
}

func (s *Server) handleSFTPMkdir(params interface{}) error {
	type mkdirParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p mkdirParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Mkdir(p.SessionID, p.Path)
}

func (s *Server) handleSFTPStat(params interface{}) (*api.SFTPFile, error) {
	type statParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p statParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Stat(p.SessionID, p.Path)
}

func (s *Server) handleSFTPClose(params interface{}) error {
	type closeParams struct {
		SessionID string `json:"sessionId"`
	}
	var p closeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.sftpMgr.Close(p.SessionID)
}

// ---- PTY Handlers ----

func (s *Server) handlePTYSpawn(params interface{}) (*api.PTYSpawnResult, error) {
	var p api.PTYSpawnParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return s.ptyMgr.Spawn(p)
}

func (s *Server) handlePTYResize(params interface{}) error {
	var p api.PTYResizeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.ptyMgr.Resize(p.ID, p.Columns, p.Rows)
}

func (s *Server) handlePTYWrite(params interface{}) error {
	var p api.PTYWriteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.ptyMgr.Write(p.ID, p.Data)
}

func (s *Server) handlePTYKill(params interface{}) error {
	var p api.PTYKillParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return s.ptyMgr.Kill(p.ID, p.Signal)
}

// ---- Serial Handlers (stubs) ----
// Serial port management requires go.bug.st/serial or similar

func (s *Server) handleSerialOpen(params interface{}) (*api.SerialOpenResult, error) {
	// TODO: Implement using go.bug.st/serial
	return nil, fmt.Errorf("Serial support not yet implemented")
}

func (s *Server) handleSerialWrite(params interface{}) error {
	// TODO: Implement
	return fmt.Errorf("Serial support not yet implemented")
}

func (s *Server) handleSerialClose(params interface{}) error {
	// TODO: Implement
	return fmt.Errorf("Serial support not yet implemented")
}

// reMarshal is a helper to re-marshal interface{} params into a typed struct
func reMarshal(from, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}
