// Package server implements the JSON-RPC 2.0 server that communicates with
// the Electron frontend over stdin/stdout or a TCP socket.
//
// Protocol:
// - The Electron app spawns this binary as a child process
// - Communication happens via JSON-RPC 2.0 over stdin/stdout
// - Each line is a complete JSON-RPC message
// - The server sends notifications for async events (data, exit, etc.)
//
// Registered methods:
//   - ping                        - Health check
//   - ssh.connect                 - Connect to SSH server
//   - ssh.startShell              - Start shell session
//   - ssh.resize                  - Resize terminal
//   - ssh.write                   - Write data to session
//   - ssh.close                   - Close session/connection
//   - ssh.listConnections         - List active connections
//   - ssh.addForward              - Add port forward
//   - ssh.removeForward           - Remove port forward
//   - ssh.listForwards            - List port forwards
//   - ssh.verifyHostKey           - Respond to host key prompt
//   - ssh.keyboardInteractiveResp - Respond to keyboard-interactive prompt
//   - sftp.open                   - Open SFTP session
//   - sftp.list                   - List directory
//   - sftp.download               - Download file
//   - sftp.upload                 - Upload file
//   - sftp.delete                 - Delete file/directory
//   - sftp.rename                 - Rename file/directory
//   - sftp.mkdir                  - Create directory
//   - sftp.mkdirAll               - Create directory tree
//   - sftp.stat                   - Get file info
//   - sftp.lstat                  - Get file info (no follow symlinks)
//   - sftp.readDir                - Read directory with symlink info
//   - sftp.chmod                  - Change file permissions
//   - sftp.readlink               - Read symbolic link target
//   - sftp.symlink                - Create symbolic link
//   - sftp.rmdir                  - Remove directory
//   - sftp.close                  - Close SFTP session
//   - pty.spawn                   - Spawn local PTY
//   - pty.resize                  - Resize PTY
//   - pty.write                   - Write data
//   - pty.kill                    - Kill process
//   - serial.open                 - Open serial port (stub)
//   - serial.write                - Write data (stub)
//   - serial.close                - Close port (stub)
//   - serial.listPorts            - List available ports (stub)
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/robertpelloni/tabby/tabby-go/pkg/ai"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/config"
	"github.com/robertpelloni/tabby/tabby-go/pkg/knownhosts"
	"github.com/robertpelloni/tabby/tabby-go/pkg/notification"
	"github.com/robertpelloni/tabby/tabby-go/pkg/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/recovery"
	"github.com/robertpelloni/tabby/tabby-go/pkg/serial"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	"github.com/robertpelloni/tabby/tabby-go/pkg/telnet"
)

// Server is the JSON-RPC server for Tabby's Go backend
type Server struct {
	sshMgr       *ssh.Manager
	sftpMgr      *sftp.Manager
	ptyMgr       *pty.Manager
	serialMgr    *serial.Manager
	telnetMgr    *telnet.Manager
	aiMgr        *ai.Manager
	knownHosts   *knownhosts.Manager
	notifMgr     *notification.Manager
	recoveryMgr  *recovery.Manager
	configMgr    *config.Manager
	reader       *bufio.Reader
	writer       io.Writer
	mu           sync.Mutex
	running      bool
}

// New creates a new Server using stdin/stdout
func New() *Server {
	s := &Server{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
	s.sshMgr = ssh.NewManager(s.sendNotification)
	s.sftpMgr = sftp.NewManager(s.sshMgr)
	s.ptyMgr = pty.NewManager(s.sendNotification)
	s.serialMgr = serial.NewManager(s.sendNotification)
	s.telnetMgr = telnet.NewManager(s.sendNotification)
	s.aiMgr = ai.NewManager()
	s.knownHosts = knownhosts.NewManager()
	s.notifMgr = notification.NewManager()
	s.recoveryMgr = recovery.NewManager()
	s.notifMgr.OnChange(func(notifs []notification.Notification) {
		s.sendNotification("notifications.changed", notifs)
	})
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

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
	s.telnetMgr = telnet.NewManager(s.sendNotification)
	s.knownHosts = knownhosts.NewManager()
	s.notifMgr = notification.NewManager()
	s.recoveryMgr = recovery.NewManager()
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s
}

// Run starts the server's main loop
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

		line = line[:len(line)-1]

		var req api.JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(0, api.ErrorParseError, "Parse error", nil)
			continue
		}

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
		result = map[string]string{"status": "ok", "version": "1.0.231-nightly.1"}

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

	// ---- SSH Port Forwarding ----
	case "ssh.addForward":
		result, err = s.handleSSHAddForward(req.Params)
	case "ssh.removeForward":
		err = s.handleSSHRemoveForward(req.Params)
	case "ssh.listForwards":
		result, err = s.handleSSHListForwards(req.Params)

	// ---- SSH Auth Callbacks ----
	case "ssh.verifyHostKey":
		err = s.handleSSHVerifyHostKey(req.Params)
	case "ssh.keyboardInteractiveResp":
		err = s.handleSSHKeyboardInteractiveResp(req.Params)

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
	case "sftp.mkdirAll":
		err = s.handleSFTPMkdirAll(req.Params)
	case "sftp.stat":
		result, err = s.handleSFTPStat(req.Params)
	case "sftp.lstat":
		result, err = s.handleSFTPLstat(req.Params)
	case "sftp.readDir":
		result, err = s.handleSFTPReadDir(req.Params)
	case "sftp.chmod":
		err = s.handleSFTPChmod(req.Params)
	case "sftp.readlink":
		result, err = s.handleSFTPReadlink(req.Params)
	case "sftp.symlink":
		err = s.handleSFTPSymlink(req.Params)
	case "sftp.rmdir":
		err = s.handleSFTPRmdir(req.Params)
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
	case "serial.listPorts":
		result, err = s.handleSerialListPorts(req.Params)

	// ---- Telnet ----
	case "telnet.connect":
		result, err = s.handleTelnetConnect(req.Params)
	case "telnet.write":
		err = s.handleTelnetWrite(req.Params)
	case "telnet.resize":
		err = s.handleTelnetResize(req.Params)
	case "telnet.close":
		err = s.handleTelnetClose(req.Params)
	case "telnet.listConnections":
		result = s.telnetMgr.ListConnections()

	// ---- Known Hosts ----
	case "knownHosts.get":
		result, err = s.handleKnownHostsGet(req.Params)
	case "knownHosts.store":
		err = s.handleKnownHostsStore(req.Params)
	case "knownHosts.remove":
		err = s.handleKnownHostsRemove(req.Params)
	case "knownHosts.list":
		result = s.knownHosts.List()
	case "knownHosts.verify":
		result, err = s.handleKnownHostsVerify(req.Params)
	case "knownHosts.loadFile":
		err = s.handleKnownHostsLoadFile(req.Params)
	case "knownHosts.saveFile":
		err = s.handleKnownHostsSaveFile(req.Params)

	// ---- Notifications ----
	case "notifications.info":
		err = s.handleNotificationInfo(req.Params)
	case "notifications.warning":
		err = s.handleNotificationWarning(req.Params)
	case "notifications.error":
		err = s.handleNotificationError(req.Params)
	case "notifications.getUnread":
		result = s.notifMgr.GetUnread()
	case "notifications.getAll":
		result = s.notifMgr.GetAll()
	case "notifications.markRead":
		err = s.handleNotificationMarkRead(req.Params)
	case "notifications.clear":
		s.notifMgr.Clear()

	// ---- Recovery ----
	case "recovery.registerTab":
		err = s.handleRecoveryRegisterTab(req.Params)
	case "recovery.unregisterTab":
		err = s.handleRecoveryUnregisterTab(req.Params)
	case "recovery.updateTab":
		err = s.handleRecoveryUpdateTab(req.Params)
	case "recovery.getTabs":
		result = s.recoveryMgr.GetRecoverableTabs()
	case "recovery.save":
		err = s.handleRecoverySave(req.Params)
	case "recovery.load":
		result, err = s.handleRecoveryLoad(req.Params)
	case "recovery.clear":
		s.recoveryMgr.Clear()

	// ---- AI ----
	case "ai.generateCommand":
		var p ai.GenerateCommandParams
		if err = reMarshal(req.Params, &p); err == nil {
			result, err = s.aiMgr.GenerateCommand(p)
		}
	case "ai.chat":
			var cp ai.ChatParams
			if err = reMarshal(req.Params, &cp); err == nil {
				result, err = s.aiMgr.Chat(cp)
			}
		case "ai.explainError":
		var p ai.ExplainErrorParams
		if err = reMarshal(req.Params, &p); err == nil {
			result, err = s.aiMgr.ExplainError(p)
		}

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

// ---- Messaging ----

func (s *Server) sendResult(id int, result interface{}) {
	resp := api.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendMessage(resp)
}

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

func (s *Server) sendNotification(method string, params interface{}) {
	notif := api.JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	s.sendMessage(notif)
}

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
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.Connect(p)
}

func (s *Server) handleSSHStartShell(params interface{}) (*api.SSHSessionResult, error) {
	var p api.SSHSessionParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.StartShell(p)
}

func (s *Server) handleSSHResize(params interface{}) error {
	var p api.SSHResizeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.Resize(p)
}

func (s *Server) handleSSHWrite(params interface{}) error {
	var p api.SSHWriteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.Write(p)
}

func (s *Server) handleSSHClose(params interface{}) error {
	var p api.SSHCloseParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.Close(p)
}

func (s *Server) handleSSHAddForward(params interface{}) (*api.PortForwardResult, error) {
	var p api.PortForwardParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.AddForward(p)
}

func (s *Server) handleSSHRemoveForward(params interface{}) error {
	var p api.PortForwardRemoveParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.RemoveForward(p)
}

func (s *Server) handleSSHListForwards(params interface{}) ([]api.PortForwardInfo, error) {
	type connParams struct {
		ConnectionID string `json:"connectionId"`
	}
	var p connParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sshMgr.ListForwards(p.ConnectionID), nil
}

func (s *Server) handleSSHVerifyHostKey(params interface{}) error {
	var p api.HostKeyVerifyResponse
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.sshMgr.HandleHostKeyResponse(p.ConnectionID, p.Accepted)
	return nil
}

func (s *Server) handleSSHKeyboardInteractiveResp(params interface{}) error {
	var p api.KeyboardInteractiveResponse
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.sshMgr.HandleKeyboardInteractiveResponse(p.ConnectionID, p.Responses)
	return nil
}

// ---- SFTP Handlers ----

func (s *Server) handleSFTPOpen(params interface{}) (*api.SFTPOpenResult, error) {
	var p api.SFTPOpenParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Open(p)
}

func (s *Server) handleSFTPList(params interface{}) ([]api.SFTPFile, error) {
	var p api.SFTPListParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.List(p)
}

func (s *Server) handleSFTPDownload(params interface{}) (*api.SFTPTransferResult, error) {
	var p api.SFTPDownloadParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Download(p)
}

func (s *Server) handleSFTPUpload(params interface{}) (*api.SFTPTransferResult, error) {
	var p api.SFTPUploadParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
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
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
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
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
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
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Mkdir(p.SessionID, p.Path)
}

func (s *Server) handleSFTPMkdirAll(params interface{}) error {
	type mkdirParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p mkdirParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.MkdirAll(p.SessionID, p.Path)
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
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Stat(p.SessionID, p.Path)
}

func (s *Server) handleSFTPLstat(params interface{}) (*api.SFTPFile, error) {
	type statParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p statParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Lstat(p.SessionID, p.Path)
}

func (s *Server) handleSFTPReadDir(params interface{}) ([]api.SFTPFile, error) {
	type readDirParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p readDirParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.ReadDir(p.SessionID, p.Path)
}

func (s *Server) handleSFTPChmod(params interface{}) error {
	var p api.SFTPChmodParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Chmod(p.SessionID, p.Path, p.Mode)
}

func (s *Server) handleSFTPReadlink(params interface{}) (map[string]string, error) {
	var p api.SFTPReadlinkParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	target, err := s.sftpMgr.Readlink(p.SessionID, p.Path)
	if err != nil {
		return nil, err
	}
	return map[string]string{"target": target}, nil
}

func (s *Server) handleSFTPSymlink(params interface{}) error {
	var p api.SFTPSymlinkParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Symlink(p.SessionID, p.OldPath, p.NewPath)
}

func (s *Server) handleSFTPRmdir(params interface{}) error {
	type rmdirParams struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
	}
	var p rmdirParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Rmdir(p.SessionID, p.Path)
}

func (s *Server) handleSFTPClose(params interface{}) error {
	type closeParams struct {
		SessionID string `json:"sessionId"`
	}
	var p closeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.sftpMgr.Close(p.SessionID)
}

// ---- PTY Handlers ----

func (s *Server) handlePTYSpawn(params interface{}) (*api.PTYSpawnResult, error) {
	var p api.PTYSpawnParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.ptyMgr.Spawn(p)
}

func (s *Server) handlePTYResize(params interface{}) error {
	var p api.PTYResizeParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.ptyMgr.Resize(p.ID, p.Columns, p.Rows)
}

func (s *Server) handlePTYWrite(params interface{}) error {
	var p api.PTYWriteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.ptyMgr.Write(p.ID, p.Data)
}

func (s *Server) handlePTYKill(params interface{}) error {
	var p api.PTYKillParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.ptyMgr.Kill(p.ID, p.Signal)
}

// ---- Serial Handlers ----

func (s *Server) handleSerialOpen(params interface{}) (*api.SerialOpenResult, error) {
	var p api.SerialOpenParams
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.serialMgr.Open(p)
}

func (s *Server) handleSerialWrite(params interface{}) error {
	var p api.SerialWriteParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.serialMgr.Write(p.ID, p.Data)
}

func (s *Server) handleSerialClose(params interface{}) error {
	var p api.SerialCloseParams
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.serialMgr.Close(p.ID)
}

func (s *Server) handleSerialListPorts(params interface{}) (*api.SerialListPortsResult, error) {
	ports, err := s.serialMgr.ListPorts()
	if err != nil {
		return nil, err
	}
	return &api.SerialListPortsResult{Ports: convertPorts(ports)}, nil
}

func convertPorts(ports []string) []api.SerialPortInfo {
	result := make([]api.SerialPortInfo, len(ports))
	for i, p := range ports {
		result[i] = api.SerialPortInfo{Name: p}
	}
	return result
}

// ---- Telnet Handlers ----

func (s *Server) handleTelnetConnect(params interface{}) (interface{}, error) {
	var p struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.telnetMgr.Connect(telnet.TelnetConnectParams{Host: p.Host, Port: p.Port})
}

func (s *Server) handleTelnetWrite(params interface{}) error {
	var p struct {
		ConnectionID string `json:"connectionId"`
		Data         string `json:"data"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.telnetMgr.Write(p.ConnectionID, p.Data)
}

func (s *Server) handleTelnetResize(params interface{}) error {
	var p struct {
		ConnectionID string `json:"connectionId"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.telnetMgr.Resize(p.ConnectionID, p.Width, p.Height)
}

func (s *Server) handleTelnetClose(params interface{}) error {
	var p struct {
		ConnectionID string `json:"connectionId"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.telnetMgr.Close(p.ConnectionID)
}

// reMarshal is a helper to re-marshal interface{} params into a typed struct
func reMarshal(from, to interface{}) error {
	data, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, to)
}

// ---- Known Hosts Handlers ----

func (s *Server) handleKnownHostsGet(params interface{}) (*knownhosts.KnownHost, error) {
	var p knownhosts.Selector
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.knownHosts.GetFor(p), nil
}

func (s *Server) handleKnownHostsStore(params interface{}) error {
	var p struct {
		Selector knownhosts.Selector `json:"selector"`
		Digest   string              `json:"digest"`
		KeyBytes []byte              `json:"keyBytes"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.knownHosts.Store(p.Selector, p.Digest, p.KeyBytes)
	return nil
}

func (s *Server) handleKnownHostsRemove(params interface{}) error {
	var p knownhosts.Selector
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.knownHosts.Remove(p)
	return nil
}

func (s *Server) handleKnownHostsVerify(params interface{}) (map[string]interface{}, error) {
	var p struct {
		Selector knownhosts.Selector `json:"selector"`
		KeyBytes []byte              `json:"keyBytes"`
	}
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	ok, err := s.knownHosts.Verify(p.Selector, p.KeyBytes)
	return map[string]interface{}{"ok": ok}, err
}

func (s *Server) handleKnownHostsLoadFile(params interface{}) error {
	var p struct {
		Path string `json:"path"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.knownHosts.LoadFromFile(p.Path)
}

func (s *Server) handleKnownHostsSaveFile(params interface{}) error {
	var p struct {
		Path string `json:"path"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.knownHosts.SaveToFile(p.Path)
}

// ---- Notification Handlers ----

func (s *Server) handleNotificationInfo(params interface{}) error {
	var p struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.notifMgr.Info(p.Title, p.Message)
	return nil
}

func (s *Server) handleNotificationWarning(params interface{}) error {
	var p struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.notifMgr.Warning(p.Title, p.Message)
	return nil
}

func (s *Server) handleNotificationError(params interface{}) error {
	var p struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.notifMgr.Error(p.Title, p.Message)
	return nil
}

func (s *Server) handleNotificationMarkRead(params interface{}) error {
	var p struct {
		ID string `json:"id"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.notifMgr.MarkRead(p.ID)
	return nil
}

// ---- Recovery Handlers ----

func (s *Server) handleRecoveryRegisterTab(params interface{}) error {
	var p recovery.TabState
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.recoveryMgr.RegisterTab(p)
	return nil
}

func (s *Server) handleRecoveryUnregisterTab(params interface{}) error {
	var p struct {
		TabID string `json:"tabId"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.recoveryMgr.UnregisterTab(p.TabID)
	return nil
}

func (s *Server) handleRecoveryUpdateTab(params interface{}) error {
	var p struct {
		TabID string `json:"tabId"`
		Title string `json:"title"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	s.recoveryMgr.UpdateTab(p.TabID, p.Title)
	return nil
}

func (s *Server) handleRecoverySave(params interface{}) error {
	var p struct {
		Path string `json:"path"`
	}
	if err := reMarshal(params, &p); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	if p.Path == "" {
		p.Path = recovery.GetRecoveryPath()
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.recoveryMgr.Save(p.Path)
}

func (s *Server) handleRecoveryLoad(params interface{}) (*recovery.RecoveryFile, error) {
	var p struct {
		Path string `json:"path"`
	}
	if err := reMarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.Path == "" {
		p.Path = recovery.GetRecoveryPath()
	}
	s.configMgr = config.NewManager()

	// Start polling the config file from the standard path and emit host:config-change IPC
	configPath := config.GetConfigPath()
	if configPath != "" {
		s.configMgr.Load(configPath)
		s.configMgr.OnChange(func(store *config.Store) {
			s.sendNotification("host:config-change", nil)
		})
		s.configMgr.StartWatching()
	}

	return s.recoveryMgr.Load(p.Path)
}
