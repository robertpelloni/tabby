// Package ssh implements an SSH client for Tabby's Go backend.
//
// It provides connection management, shell sessions, port forwarding,
// X11 forwarding, SFTP, and agent forwarding — essentially the same
// functionality as the TypeScript russh-based implementation but in Go.
//
// The SSH client uses golang.org/x/crypto/ssh which supports:
// - Password, public key, and keyboard-interactive authentication
// - Agent forwarding (via SSH_AUTH_SOCK)
// - Port forwarding (local, remote, dynamic/SOCKS)
// - X11 forwarding (via xorg/xauth integration)
// - Jump host / proxy jump chains
// - SFTP subsystem
package ssh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Manager manages SSH connections and sessions
type Manager struct {
	mu          sync.RWMutex
	connections map[string]*Connection // connectionID -> Connection
	sessions    map[string]*Session    // sessionID -> Session
	notify      NotifyFunc             // callback to send notifications to client
	idCounter   int
}

// NotifyFunc is the callback type for sending notifications to the Electron client
type NotifyFunc func(method string, params interface{})

// NewManager creates a new SSH connection manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		sessions:    make(map[string]*Session),
		notify:      notify,
	}
}

// Connection represents an active SSH connection
type Connection struct {
	ID         string
	Client     *ssh.Client
	Config     *ssh.ClientConfig
	RemoteAddr string
	ServerVer  string
	mu         sync.Mutex
}

// Session represents an active shell session within a connection
type Session struct {
	ID           string
	ConnectionID string
	Session      *ssh.Session
	Stdin        io.WriteCloser
	done         chan struct{}
}

// nextID generates a unique identifier
func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// Connect establishes a new SSH connection
func (m *Manager) Connect(params api.SSHConnectParams) (*api.SSHConnectionResult, error) {
	// Build SSH client config
	config, err := m.buildClientConfig(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}

	// Handle jump hosts (proxy jump)
	var client *ssh.Client
	addr := fmt.Sprintf("%s:%d", params.Host, params.Port)
	if params.Port == 0 {
		addr = fmt.Sprintf("%s:22", params.Host)
	}

	if params.JumpHost != nil {
		client, err = m.connectViaJump(params.JumpHost, addr, config)
	} else if params.ProxyCommand != "" {
		client, err = m.connectViaProxyCommand(params.ProxyCommand, addr, config)
	} else {
		client, err = ssh.Dial("tcp", addr, config)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	connID := m.nextID("ssh-conn")
	conn := &Connection{
		ID:         connID,
		Client:     client,
		Config:     config,
		RemoteAddr: addr,
		ServerVer:  string(client.ServerVersion()),
	}

	m.mu.Lock()
	m.connections[connID] = conn
	m.mu.Unlock()

	// Set up keepalive if configured
	if params.KeepaliveInterval > 0 {
		go m.keepalive(conn, params.KeepaliveInterval, params.KeepaliveCountMax)
	}

	result := &api.SSHConnectionResult{
		ConnectionID:  connID,
		ServerVersion: conn.ServerVer,
		RemoteAddress: addr,
	}

	return result, nil
}

// StartShell opens a shell session on an existing connection
func (m *Manager) StartShell(params api.SSHSessionParams) (*api.SSHSessionResult, error) {
	m.mu.RLock()
	conn, ok := m.connections[params.ConnectionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("connection not found: %s", params.ConnectionID)
	}

	session, err := conn.Client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	term := params.Terminal
	if term == "" {
		term = "xterm-256color"
	}

	// Request PTY
	if err := session.RequestPty(term, params.Rows, params.Columns, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	// Get stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Capture stdout/stderr and forward as notifications
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	sessionID := m.nextID("ssh-session")
	sess := &Session{
		ID:           sessionID,
		ConnectionID: params.ConnectionID,
		Session:      session,
		Stdin:        stdin,
		done:         make(chan struct{}),
	}

	m.mu.Lock()
	m.sessions[sessionID] = sess
	m.mu.Unlock()

	// Start shell (or run command if specified)
	if params.Command != "" {
		err = session.Start(params.Command)
	} else {
		err = session.Shell()
	}
	if err != nil {
		session.Close()
		m.mu.Lock()
		delete(m.sessions, sessionID)
		m.mu.Unlock()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	// Forward output as notifications
	go m.forwardOutput(params.ConnectionID, sessionID, stdout)
	go m.forwardOutput(params.ConnectionID, sessionID, stderr)

	// Monitor session exit
	go m.monitorExit(params.ConnectionID, sessionID, session)

	return &api.SSHSessionResult{SessionID: sessionID}, nil
}

// Resize resizes the terminal for a session
func (m *Manager) Resize(params api.SSHResizeParams) error {
	m.mu.RLock()
	sess, ok := m.sessions[params.SessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", params.SessionID)
	}

	return sess.Session.WindowChange(params.Rows, params.Columns)
}

// Write sends data to a session's stdin
func (m *Manager) Write(params api.SSHWriteParams) error {
	m.mu.RLock()
	sess, ok := m.sessions[params.SessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", params.SessionID)
	}

	data, err := base64.StdEncoding.DecodeString(params.Data)
	if err != nil {
		return fmt.Errorf("invalid base64 data: %w", err)
	}

	_, err = sess.Stdin.Write(data)
	return err
}

// Close closes a session or entire connection
func (m *Manager) Close(params api.SSHCloseParams) error {
	if params.SessionID != "" {
		m.mu.Lock()
		sess, ok := m.sessions[params.SessionID]
		if ok {
			delete(m.sessions, params.SessionID)
		}
		m.mu.Unlock()

		if ok {
			sess.Stdin.Close()
			return sess.Session.Close()
		}
		return fmt.Errorf("session not found: %s", params.SessionID)
	}

	m.mu.Lock()
	conn, ok := m.connections[params.ConnectionID]
	if ok {
		// Close all sessions for this connection
		for id, sess := range m.sessions {
			if sess.ConnectionID == params.ConnectionID {
				sess.Stdin.Close()
				sess.Session.Close()
				delete(m.sessions, id)
			}
		}
		delete(m.connections, params.ConnectionID)
	}
	m.mu.Unlock()

	if ok {
		return conn.Client.Close()
	}
	return fmt.Errorf("connection not found: %s", params.ConnectionID)
}

// ListConnections returns all active connection IDs
func (m *Manager) ListConnections() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.connections))
	for id := range m.connections {
		ids = append(ids, id)
	}
	return ids
}

// GetConnection returns a connection by ID (used internally by SFTP etc)
func (m *Manager) GetConnection(id string) (*ssh.Client, error) {
	m.mu.RLock()
	conn, ok := m.connections[id]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("connection not found: %s", id)
	}
	return conn.Client, nil
}

// buildClientConfig creates an ssh.ClientConfig from the connection parameters
func (m *Manager) buildClientConfig(params api.SSHConnectParams) (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User: params.Username,
	}

	if params.ReadyTimeout > 0 {
		config.Timeout = time.Duration(params.ReadyTimeout) * time.Second
	} else {
		config.Timeout = 30 * time.Second
	}

	// Host key callback — for now, accept all (TODO: implement known_hosts verification)
	config.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	// Build auth methods
	var authMethods []ssh.AuthMethod

	switch params.Auth.Type {
	case "password":
		authMethods = append(authMethods, ssh.Password(params.Auth.Password))
	case "publicKey":
		for _, keyData := range params.Auth.PrivateKeyPaths {
			key, err := os.ReadFile(keyData)
			if err != nil {
				continue
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				continue
			}
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
		// Also try inline private key
		if params.Auth.PrivateKey != "" {
			signer, err := ssh.ParsePrivateKey([]byte(params.Auth.PrivateKey))
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	case "agent":
		authMethod, err := m.agentAuth(params.Auth.AgentSocketPath)
		if err != nil {
			return nil, fmt.Errorf("agent auth failed: %w", err)
		}
		authMethods = append(authMethods, authMethod)
	case "keyboardInteractive":
		// For keyboard-interactive, we'd need to send prompts back to the client
		// and wait for responses. This is a simplified version.
		authMethods = append(authMethods, ssh.Password(params.Auth.Password))
	}

	if len(authMethods) == 0 {
		// Try agent as fallback
		if auth, err := m.agentAuth(""); err == nil {
			authMethods = append(authMethods, auth)
		}
	}

	config.Auth = authMethods

	// Apply custom algorithms if specified
	if params.Algorithms != nil {
		config.Config = ssh.Config{}
		if len(params.Algorithms.KEX) > 0 {
			config.Config.KeyExchanges = params.Algorithms.KEX
		}
		if len(params.Algorithms.Cipher) > 0 {
			config.Config.Ciphers = params.Algorithms.Cipher
		}
		if len(params.Algorithms.HMAC) > 0 {
			config.Config.MACs = params.Algorithms.HMAC
		}
	}

	return config, nil
}

// agentAuth creates an auth method using the SSH agent
func (m *Manager) agentAuth(socketPath string) (ssh.AuthMethod, error) {
	if socketPath == "" {
		socketPath = os.Getenv("SSH_AUTH_SOCK")
	}
	if socketPath == "" {
		return nil, fmt.Errorf("no SSH agent socket found")
	}

	agentConn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent: %w", err)
	}

	agentClient := agent.NewClient(agentConn)
	// Use the forwarded agent
	return ssh.PublicKeysCallback(agentClient.Signers), nil
}

// connectViaJump connects through a jump host
func (m *Manager) connectViaJump(jumpParams *api.SSHConnectParams, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// Connect to jump host first
	jumpConfig, err := m.buildClientConfig(*jumpParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build jump host config: %w", err)
	}

	jumpAddr := fmt.Sprintf("%s:%d", jumpParams.Host, jumpParams.Port)
	if jumpParams.Port == 0 {
		jumpAddr = fmt.Sprintf("%s:22", jumpParams.Host)
	}

	jumpClient, err := ssh.Dial("tcp", jumpAddr, jumpConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to jump host: %w", err)
	}

	// Dial through the jump host to the target
	conn, err := jumpClient.Dial("tcp", targetAddr)
	if err != nil {
		jumpClient.Close()
		return nil, fmt.Errorf("failed to dial target through jump host: %w", err)
	}

	// Establish SSH connection over the proxied connection
	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		conn.Close()
		jumpClient.Close()
		return nil, fmt.Errorf("failed to establish SSH over jump: %w", err)
	}

	client := ssh.NewClient(ncc, chans, reqs)
	return client, nil
}

// connectViaProxyCommand connects using a proxy command
func (m *Manager) connectViaProxyCommand(proxyCmd string, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// TODO: implement proxy command execution
	return nil, fmt.Errorf("proxy command not yet implemented")
}

// forwardOutput reads from a reader and sends data notifications
func (m *Manager) forwardOutput(connID, sessionID string, reader io.Reader) {
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			encoded := base64.StdEncoding.EncodeToString(buf[:n])
			m.notify("ssh.data", api.DataNotification{
				ConnectionID: connID,
				SessionID:    sessionID,
				Data:         encoded,
			})
		}
		if err != nil {
			break
		}
	}
}

// monitorExit waits for a session to exit and sends a notification
func (m *Manager) monitorExit(connID, sessionID string, session *ssh.Session) {
	err := session.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	m.notify("ssh.exit", api.ExitNotification{
		ConnectionID: connID,
		SessionID:    sessionID,
		ExitCode:     exitCode,
	})

	// Clean up session
	m.mu.Lock()
	delete(m.sessions, sessionID)
	m.mu.Unlock()
}

// keepalive sends periodic keepalive requests
func (m *Manager) keepalive(conn *Connection, interval, countMax int) {
	if countMax == 0 {
		countMax = 3
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	failCount := 0
	for range ticker.C {
		conn.mu.Lock()
		client := conn.Client
		conn.mu.Unlock()

		if client == nil {
			return
		}

		_, _, err := client.SendRequest("keepalive@golang.org", true, nil)
		if err != nil {
			failCount++
			if failCount >= countMax {
				m.notify("ssh.exit", api.ExitNotification{
					ConnectionID: conn.ID,
					ExitCode:     -1,
					Signal:       "keepalive-timeout",
				})
				return
			}
		} else {
			failCount = 0
		}
	}
}

// buildKnownHostsCallback creates a HostKeyCallback that verifies against known_hosts
// TODO: implement this with the knownhosts package
func buildKnownHostsCallback(knownHostsPath string) (ssh.HostKeyCallback, error) {
	return knownhosts.New(knownHostsPath)
}

// MarshalJSON is a helper to marshal any value to JSON
func MarshalJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
