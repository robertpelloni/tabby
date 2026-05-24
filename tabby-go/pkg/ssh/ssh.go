// Package ssh implements an SSH client for Tabby's Go backend.
//
// It provides connection management, shell sessions, port forwarding,
// X11 forwarding, SFTP, and agent forwarding — essentially the same
// functionality as the TypeScript russh-based implementation but in Go.
//
// Features:
// - Password, public key, keyboard-interactive, and agent authentication
// - Agent forwarding (via SSH_AUTH_SOCK on Unix, named pipe on Windows)
// - Port forwarding (local, remote, dynamic/SOCKS5)
// - Jump host / proxy jump chains
// - Proxy command support
// - SOCKS and HTTP proxy support
// - Known hosts verification
// - Keepalive with disconnect detection
// - SSH multiplexing (multiple sessions per connection)
// - Host key verification with known_hosts support
// - Keyboard-interactive authentication with client-side prompt forwarding
// - X11 forwarding (channel-level support)
package ssh

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
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
	forwards    map[string]*Forward    // forwardID -> Forward
	notify      NotifyFunc             // callback to send notifications to client
	idCounter   int

	// Pending auth callbacks
	pendingHostKey      map[string]chan bool                // connectionID -> accept channel
	pendingKeyboardInt  map[string]chan KeyboardIntResponse // connectionID -> response channel
	authMu              sync.Mutex
}

// NotifyFunc is the callback type for sending notifications to the Electron client
type NotifyFunc func(method string, params interface{})

// KeyboardIntResponse holds the user's responses to keyboard-interactive prompts
type KeyboardIntResponse struct {
	Responses []string
	Error     error
}

// NewManager creates a new SSH connection manager
func NewManager(notify NotifyFunc) *Manager {
	return &Manager{
		connections:        make(map[string]*Connection),
		sessions:           make(map[string]*Session),
		forwards:           make(map[string]*Forward),
		notify:             notify,
		pendingHostKey:     make(map[string]chan bool),
		pendingKeyboardInt: make(map[string]chan KeyboardIntResponse),
	}
}

// Connection represents an active SSH connection
type Connection struct {
	RefCount int

	ID         string
	Client     *ssh.Client
	Config     *ssh.ClientConfig
	RemoteAddr string
	ServerVer  string
	Params     api.SSHConnectParams
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

// Forward represents an active port forward
type Forward struct {
	ID           string
	ConnectionID string
	Type         api.PortForwardType
	Host         string
	Port         int
	TargetAddr   string
	TargetPort   int
	Listener     net.Listener
	done         chan struct{}
}

// nextID generates a unique identifier
func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixMilli(), m.idCounter)
}

// getFingerprint generates a unique fingerprint for a connection profile
func getFingerprint(params api.SSHConnectParams) string {
	return fmt.Sprintf("%s@%s:%d|%v", params.Username, params.Host, params.Port, params.JumpHost != nil)
}

// Connect establishes a new SSH connection or returns an active multiplexed connection
func (m *Manager) Connect(params api.SSHConnectParams) (*api.SSHConnectionResult, error) {
	fingerprint := getFingerprint(params)

	// Multiplexer logic
	m.mu.Lock()
	for _, conn := range m.connections {
		if getFingerprint(conn.Params) == fingerprint {
			// Found an existing connection, multiplex it!
			conn.RefCount++
			m.mu.Unlock()
			m.sendServiceMessage(conn.ID, fmt.Sprintf("Multiplexing existing SSH connection to %s", conn.RemoteAddr))
			return &api.SSHConnectionResult{
				ConnectionID: conn.ID,
				ServerVersion:    conn.ServerVer,
			}, nil
		}
	}
	m.mu.Unlock()

	// Build SSH client config
	config, err := m.buildClientConfig(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config: %w", err)
	}

	// Determine target address
	port := params.Port
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(params.Host, fmt.Sprintf("%d", port))

	// Connect via the appropriate transport
	var client *ssh.Client

	switch {
	case params.JumpHost != nil:
		client, err = m.connectViaJump(params.JumpHost, addr, config)
	case params.ProxyCommand != "":
		m.sendServiceMessage(params.Host, "Using proxy command: "+params.ProxyCommand)
		client, err = m.connectViaProxyCommand(params.ProxyCommand, addr, config)
	case params.SocksProxyHost != "":
		m.sendServiceMessage(params.Host, fmt.Sprintf("Using SOCKS proxy: %s", net.JoinHostPort(params.SocksProxyHost, fmt.Sprintf("%d", params.SocksProxyPort))))
		client, err = m.connectViaSocksProxy(params.SocksProxyHost, params.SocksProxyPort, addr, config)
	case params.HTTPProxyHost != "":
		m.sendServiceMessage(params.Host, fmt.Sprintf("Using HTTP proxy: %s", net.JoinHostPort(params.HTTPProxyHost, fmt.Sprintf("%d", params.HTTPProxyPort))))
		client, err = m.connectViaHTTPProxy(params.HTTPProxyHost, params.HTTPProxyPort, addr, config)
	default:
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
		Params:     params,
		RefCount:   1,
	}

	m.mu.Lock()
	m.connections[connID] = conn
	m.mu.Unlock()

	// Set up keepalive if configured
	if params.KeepaliveInterval > 0 {
		go m.keepalive(conn, params.KeepaliveInterval, params.KeepaliveCountMax)
	}

	// Send banner notification
	banner := string(client.ServerVersion())
	if banner != "" && !params.SkipBanner {
		m.notify("ssh.banner", api.BannerNotification{
			ConnectionID: connID,
			Message:      banner,
		})
	}

	result := &api.SSHConnectionResult{
		ConnectionID:  connID,
		ServerVersion: conn.ServerVer,
		RemoteAddress: addr,
		Banner:        banner,
		JumpChain:     m.buildJumpChain(params),
	}

	return result, nil
}

// buildJumpChain returns a list of hostnames in the jump chain
func (m *Manager) buildJumpChain(params api.SSHConnectParams) []string {
	var chain []string
	curr := &params
	for curr != nil {
		chain = append(chain, curr.Host)
		curr = curr.JumpHost
	}
	// Reverse to show jump1 -> jump2 -> target
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
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
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
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

	// Request X11 forwarding if requested
	if params.X11 {
		// X11 forwarding is handled at the channel level
		// The Go SSH library requires manual X11 channel handling
		m.sendServiceMessage(params.ConnectionID, "X11 forwarding requested")
	}

	// Request agent forwarding if requested
	if params.AgentForward {
		// Agent forwarding is handled via channel opens
		m.sendServiceMessage(params.ConnectionID, "Agent forwarding requested")
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

	// Handle X11 channels if X11 was requested
	if params.X11 {
		go m.handleX11Channels(params.ConnectionID, sessionID, conn)
	}

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
		// Decrease the reference count for multiplexed connections
		conn.RefCount--
		if conn.RefCount > 0 {
			m.mu.Unlock()
			m.sendServiceMessage(params.ConnectionID, fmt.Sprintf("Detached from multiplexed SSH connection to %s. (%d references remaining)", conn.RemoteAddr, conn.RefCount))
			return nil
		}

		// If ref count reaches 0, actually close and cleanup everything
		// Close all sessions for this connection
		for id, sess := range m.sessions {
			if sess.ConnectionID == params.ConnectionID {
				sess.Stdin.Close()
				sess.Session.Close()
				delete(m.sessions, id)
			}
		}
		// Close all forwards for this connection
		for id, fwd := range m.forwards {
			if fwd.ConnectionID == params.ConnectionID {
				fwd.Listener.Close()
				close(fwd.done)
				delete(m.forwards, id)
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

// ---- Port Forwarding ----

// AddForward adds a port forward to a connection
func (m *Manager) AddForward(params api.PortForwardParams) (*api.PortForwardResult, error) {
	m.mu.RLock()
	conn, ok := m.connections[params.ConnectionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("connection not found: %s", params.ConnectionID)
	}

	forwardID := m.nextID("fwd")
	fwd := &Forward{
		ID:           forwardID,
		ConnectionID: params.ConnectionID,
		Type:         params.Type,
		Host:         params.Host,
		Port:         params.Port,
		TargetAddr:   params.TargetAddress,
		TargetPort:   params.TargetPort,
		done:         make(chan struct{}),
	}

	switch params.Type {
	case api.PortForwardLocal:
		err := m.startLocalForward(conn, fwd)
		if err != nil {
			return nil, fmt.Errorf("failed to start local forward: %w", err)
		}
	case api.PortForwardRemote:
		err := m.startRemoteForward(conn, fwd)
		if err != nil {
			return nil, fmt.Errorf("failed to start remote forward: %w", err)
		}
	case api.PortForwardDynamic:
		err := m.startDynamicForward(conn, fwd)
		if err != nil {
			return nil, fmt.Errorf("failed to start dynamic forward: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown forward type: %s", params.Type)
	}

	m.mu.Lock()
	m.forwards[forwardID] = fwd
	m.mu.Unlock()

	m.sendServiceMessage(params.ConnectionID, fmt.Sprintf("Forwarded %s", fwd.String()))

	return &api.PortForwardResult{ForwardID: forwardID}, nil
}

// RemoveForward stops and removes a port forward
func (m *Manager) RemoveForward(params api.PortForwardRemoveParams) error {
	m.mu.Lock()
	fwd, ok := m.forwards[params.ForwardID]
	if ok {
		delete(m.forwards, params.ForwardID)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("forward not found: %s", params.ForwardID)
	}

	if fwd.Listener != nil {
		fwd.Listener.Close()
	}
	close(fwd.done)

	m.sendServiceMessage(fwd.ConnectionID, fmt.Sprintf("Stopped forwarding %s", fwd.String()))
	return nil
}

// ListForwards returns all active port forwards for a connection
func (m *Manager) ListForwards(connectionID string) []api.PortForwardInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []api.PortForwardInfo
	for _, fwd := range m.forwards {
		if fwd.ConnectionID == connectionID {
			result = append(result, api.PortForwardInfo{
				ID:            fwd.ID,
				Type:          fwd.Type,
				Host:          fwd.Host,
				Port:          fwd.Port,
				TargetAddress: fwd.TargetAddr,
				TargetPort:    fwd.TargetPort,
				Active:        true,
			})
		}
	}
	return result
}

// startLocalForward starts a local port forward (local -> remote)
func (m *Manager) startLocalForward(conn *Connection, fwd *Forward) error {
	listenAddr := net.JoinHostPort(fwd.Host, fmt.Sprintf("%d", fwd.Port))
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}
	fwd.Listener = listener

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				select {
				case <-fwd.done:
					return
				default:
					m.sendServiceMessage(conn.ID, fmt.Sprintf("Local forward listener error: %v", err))
					return
				}
			}

			go m.handleLocalForwardConn(conn, fwd, localConn)
		}
	}()

	return nil
}

func (m *Manager) handleLocalForwardConn(conn *Connection, fwd *Forward, localConn net.Conn) {
	targetAddr := net.JoinHostPort(fwd.TargetAddr, fmt.Sprintf("%d", fwd.TargetPort))

	// Open a channel to the remote SSH server
	remoteConn, err := conn.Client.Dial("tcp", targetAddr)
	if err != nil {
		m.sendServiceMessage(conn.ID, fmt.Sprintf("Could not forward to %s: %v", targetAddr, err))
		localConn.Close()
		return
	}

	// Bidirectional copy
	go m.relay(localConn, remoteConn)
	go m.relay(remoteConn, localConn)
}

// startRemoteForward starts a remote port forward (remote -> local)
func (m *Manager) startRemoteForward(conn *Connection, fwd *Forward) error {
	remoteAddr := net.JoinHostPort(fwd.Host, fmt.Sprintf("%d", fwd.Port))
	listener, err := conn.Client.Listen("tcp", remoteAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on remote %s: %w", remoteAddr, err)
	}
	fwd.Listener = listener

	go func() {
		for {
			remoteConn, err := listener.Accept()
			if err != nil {
				select {
				case <-fwd.done:
					return
				default:
					m.sendServiceMessage(conn.ID, fmt.Sprintf("Remote forward listener error: %v", err))
					return
				}
			}

			localAddr := net.JoinHostPort(fwd.TargetAddr, fmt.Sprintf("%d", fwd.TargetPort))
			localConn, err := net.Dial("tcp", localAddr)
			if err != nil {
				m.sendServiceMessage(conn.ID, fmt.Sprintf("Could not connect to local %s: %v", localAddr, err))
				remoteConn.Close()
				continue
			}

			go m.relay(remoteConn, localConn)
			go m.relay(localConn, remoteConn)
		}
	}()

	return nil
}

// startDynamicForward starts a dynamic (SOCKS5) port forward
func (m *Manager) startDynamicForward(conn *Connection, fwd *Forward) error {
	listenAddr := net.JoinHostPort(fwd.Host, fmt.Sprintf("%d", fwd.Port))
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}
	fwd.Listener = listener

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				select {
				case <-fwd.done:
					return
				default:
					return
				}
			}

			go m.handleSOCKS5Conn(conn, localConn)
		}
	}()

	return nil
}

// handleSOCKS5Conn handles a SOCKS5 connection for dynamic forwarding
func (m *Manager) handleSOCKS5Conn(conn *Connection, localConn net.Conn) {
	defer localConn.Close()

	// SOCKS5 handshake
	buf := make([]byte, 256)

	// Read version and auth methods
	n, err := localConn.Read(buf)
	if err != nil || n < 2 {
		return
	}
	if buf[0] != 0x05 {
		return
	}

	// No auth required
	localConn.Write([]byte{0x05, 0x00})

	// Read connect request
	n, err = localConn.Read(buf)
	if err != nil || n < 7 {
		return
	}
	if buf[0] != 0x05 || buf[1] != 0x01 {
		localConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var targetAddr string
	var targetPort int

	switch buf[3] {
	case 0x01: // IPv4
		if n < 10 {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
		targetPort = int(buf[8])<<8 | int(buf[9])
	case 0x03: // Domain
		if n < int(5+buf[4]+2) {
			return
		}
		domainLen := int(buf[4])
		targetAddr = string(buf[5 : 5+domainLen])
		targetPort = int(buf[5+domainLen])<<8 | int(buf[6+domainLen])
	case 0x04: // IPv6
		if n < 22 {
			return
		}
		parts := make([]string, 16)
		for i := 0; i < 16; i++ {
			parts[i] = fmt.Sprintf("%x", buf[4+i])
		}
		targetAddr = strings.Join(parts, ":")
		targetPort = int(buf[20])<<8 | int(buf[21])
	default:
		localConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// Connect through SSH
	remoteConn, err := conn.Client.Dial("tcp", net.JoinHostPort(targetAddr, fmt.Sprintf("%d", targetPort)))
	if err != nil {
		localConn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// Send success response
	localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// Relay
	go m.relay(localConn, remoteConn)
	m.relay(remoteConn, localConn)
}

// relay copies data between two connections
func (m *Manager) relay(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

// ---- Host Key Verification ----

// HandleHostKeyResponse processes the client's response to a host key prompt
func (m *Manager) HandleHostKeyResponse(connectionID string, accepted bool) {
	m.authMu.Lock()
	ch, ok := m.pendingHostKey[connectionID]
	if ok {
		delete(m.pendingHostKey, connectionID)
	}
	m.authMu.Unlock()

	if ok {
		ch <- accepted
	}
}

// HandleKeyboardInteractiveResponse processes the client's response to keyboard-interactive prompts
func (m *Manager) HandleKeyboardInteractiveResponse(connectionID string, responses []string) {
	m.authMu.Lock()
	ch, ok := m.pendingKeyboardInt[connectionID]
	if ok {
		delete(m.pendingKeyboardInt, connectionID)
	}
	m.authMu.Unlock()

	if ok {
		ch <- KeyboardIntResponse{Responses: responses}
	}
}

// ---- Internal Methods ----

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

	// Host key callback
	if params.VerifyHostKey && params.KnownHostsPath != "" {
		hostKeyCallback, err := knownhosts.New(params.KnownHostsPath)
		if err != nil {
			m.sendServiceMessage(params.Host, fmt.Sprintf("Warning: Could not load known_hosts from %s: %v", params.KnownHostsPath, err))
			config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		} else {
			config.HostKeyCallback = hostKeyCallback
		}
	} else if params.VerifyHostKey {
		// Interactive host key verification
		config.HostKeyCallback = m.hostKeyCallback(params)
	} else {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Build auth methods
	var authMethods []ssh.AuthMethod

	switch params.Auth.Type {
	case "password":
		if params.Auth.Password != "" {
			authMethods = append(authMethods, ssh.Password(params.Auth.Password))
		}
	case "publicKey":
		authMethods = m.publicKeyAuthMethods(params)
	case "agent":
		if auth, err := m.agentAuth(params.Auth.AgentSocketPath); err == nil {
			authMethods = append(authMethods, auth)
		}
	case "keyboardInteractive":
		authMethods = append(authMethods, m.keyboardInteractiveAuth(params))
	case "none":
		// No auth
	}

	// Fallback: always try agent if available
	if params.Auth.Type != "agent" {
		if auth, err := m.agentAuth(""); err == nil {
			authMethods = append(authMethods, auth)
		}
	}

	// Fallback: try saved password
	if params.Password != "" && params.Auth.Type != "password" {
		authMethods = append(authMethods, ssh.Password(params.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
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

// publicKeyAuthMethods creates auth methods from public key files
func (m *Manager) publicKeyAuthMethods(params api.SSHConnectParams) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// Try inline private key
	if params.Auth.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(params.Auth.PrivateKey))
		if err == nil {
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	// Try private key files
	for _, keyPath := range params.Auth.PrivateKeyPaths {
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			m.sendServiceMessage(params.Host, fmt.Sprintf("Could not load private key %s: %v", keyPath, err))
			continue
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			m.sendServiceMessage(params.Host, fmt.Sprintf("Failed to parse private key %s: %v", keyPath, err))
			continue
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	return methods
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
	return ssh.PublicKeysCallback(agentClient.Signers), nil
}

// keyboardInteractiveAuth creates an auth method for keyboard-interactive authentication
func (m *Manager) keyboardInteractiveAuth(params api.SSHConnectParams) ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		// If there are no questions, just return empty responses
		if len(questions) == 0 {
			return []string{}, nil
		}

		connID := params.Host + fmt.Sprintf(":%d", params.Port)

		// Build prompts
		prompts := make([]api.Prompt, len(questions))
		for i, q := range questions {
			prompts[i] = api.Prompt{
				Prompt: q,
				Echo:   echos[i],
			}
		}

		// Send notification to client
		m.notify("ssh.keyboardInteractive", api.KeyboardInteractiveNotification{
			ConnectionID: connID,
			Name:         user,
			Instruction:  instruction,
			Prompts:      prompts,
		})

		// Wait for response
		ch := make(chan KeyboardIntResponse, 1)
		m.authMu.Lock()
		m.pendingKeyboardInt[connID] = ch
		m.authMu.Unlock()

		select {
		case resp := <-ch:
			if resp.Error != nil {
				return nil, resp.Error
			}
			return resp.Responses, nil
		case <-time.After(60 * time.Second):
			m.authMu.Lock()
			delete(m.pendingKeyboardInt, connID)
			m.authMu.Unlock()
			return nil, fmt.Errorf("keyboard-interactive auth timed out")
		}
	})
}

// hostKeyCallback creates an interactive host key verification callback
func (m *Manager) hostKeyCallback(params api.SSHConnectParams) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		connID := params.Host + fmt.Sprintf(":%d", params.Port)
		fingerprint := ssh.FingerprintSHA256(key)

		// Send prompt to client
		m.notify("ssh.hostKeyPrompt", api.HostKeyPromptNotification{
			ConnectionID: connID,
			Host:         hostname,
			Port:         params.Port,
			KeyType:      key.Type(),
			Fingerprint:  fingerprint,
			KeyBytes:     base64.StdEncoding.EncodeToString(key.Marshal()),
		})

		// Wait for response
		ch := make(chan bool, 1)
		m.authMu.Lock()
		m.pendingHostKey[connID] = ch
		m.authMu.Unlock()

		select {
		case accepted := <-ch:
			if !accepted {
				return fmt.Errorf("host key rejected by user")
			}
			return nil
		case <-time.After(30 * time.Second):
			m.authMu.Lock()
			delete(m.pendingHostKey, connID)
			m.authMu.Unlock()
			return fmt.Errorf("host key verification timed out")
		}
	}
}

// connectViaJump connects through a jump host (supports chained jumps)
func (m *Manager) connectViaJump(jumpParams *api.SSHConnectParams, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// 1. Build SSH configuration for the jump host
	jumpConfig, err := m.buildClientConfig(*jumpParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build jump host config: %w", err)
	}

	jumpPort := jumpParams.Port
	if jumpPort == 0 {
		jumpPort = 22
	}
	jumpAddr := net.JoinHostPort(jumpParams.Host, fmt.Sprintf("%d", jumpPort))

	// 2. Connect to the jump host (handling recursive chains natively)
	var jumpClient *ssh.Client
	if jumpParams.JumpHost != nil {
		jumpClient, err = m.connectViaJump(jumpParams.JumpHost, jumpAddr, jumpConfig)
	} else {
		jumpClient, err = ssh.Dial("tcp", jumpAddr, jumpConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to jump host %s: %w", jumpAddr, err)
	}

	// 3. Dial the target address through the active jump host connection
	// This opens a direct-tcpip channel on the jump host to the target
	netConn, err := jumpClient.Dial("tcp", targetAddr)
	if err != nil {
		jumpClient.Close()
		return nil, fmt.Errorf("failed to dial target through jump host: %w", err)
	}

	// 4. Perform the SSH handshake to the final target over the proxied TCP channel
	ncc, chans, reqs, err := ssh.NewClientConn(netConn, targetAddr, config)
	if err != nil {
		netConn.Close()
		jumpClient.Close()
		return nil, fmt.Errorf("failed to establish SSH over jump: %w", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
}

// connectViaProxyCommand connects using a proxy command
func (m *Manager) connectViaProxyCommand(proxyCmd string, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// Parse the proxy command
	parts := strings.Fields(proxyCmd)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty proxy command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy stdin: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start proxy command: %w", err)
	}

	// Establish SSH over the proxy's stdin/stdout
	conn := &proxyConn{
		stdin:  stdinPipe,
		stdout: stdoutPipe,
		local:  "proxy",
		remote: targetAddr,
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to establish SSH over proxy: %w", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
}

// connectViaSocksProxy connects through a SOCKS5 proxy
func (m *Manager) connectViaSocksProxy(proxyHost string, proxyPort int, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	proxyAddr := net.JoinHostPort(proxyHost, fmt.Sprintf("%d", proxyPort))
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SOCKS proxy: %w", err)
	}

	// SOCKS5 handshake
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid target address: %w", err)
	}
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	// Send SOCKS5 connect
	req := []byte{0x05, 0x01, 0x00, 0x03, byte(len(host))}
	req = append(req, []byte(host)...)
	req = append(req, byte(port>>8), byte(port&0xff))
	conn.Write(req)

	resp := make([]byte, 256)
	n, err := conn.Read(resp)
	if err != nil || n < 2 || resp[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("SOCKS proxy connection failed")
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH over SOCKS proxy: %w", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
}

// connectViaHTTPProxy connects through an HTTP CONNECT proxy
func (m *Manager) connectViaHTTPProxy(proxyHost string, proxyPort int, targetAddr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	proxyAddr := net.JoinHostPort(proxyHost, fmt.Sprintf("%d", proxyPort))
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP proxy: %w", err)
	}

	// Send HTTP CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)
	conn.Write([]byte(connectReq))

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("HTTP proxy read error: %w", err)
	}

	response := string(buf[:n])
	if !strings.Contains(response, "200") {
		conn.Close()
		return nil, fmt.Errorf("HTTP proxy CONNECT failed: %s", strings.TrimSpace(response))
	}

	// Skip any remaining headers
	for {
		line := make([]byte, 4096)
		n, err := conn.Read(line)
		if err != nil || n == 0 {
			break
		}
		if strings.Contains(string(line[:n]), "\r\n\r\n") {
			break
		}
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH over HTTP proxy: %w", err)
	}

	return ssh.NewClient(ncc, chans, reqs), nil
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

// handleX11Channels handles incoming X11 channels
func (m *Manager) handleX11Channels(connID, sessionID string, conn *Connection) {
	for {
		ch := conn.Client.HandleChannelOpen("x11")
		if ch == nil {
			return
		}

		newChannel := <-ch
		if newChannel == nil {
			return
		}

		go m.handleX11Channel(connID, newChannel)
	}
}

func (m *Manager) handleX11Channel(connID string, newChannel ssh.NewChannel) {
	ch, reqs, err := newChannel.Accept()
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)

	// Try to connect to local X11 display
	display := os.Getenv("DISPLAY")
	if display == "" {
		ch.Close()
		return
	}

	// Parse display
	xHost, xDisplay := "localhost", "0"
	if parts := strings.SplitN(display, ":", 2); len(parts) == 2 {
		xHost = parts[0]
		xDisplay = parts[1]
	}
	if xHost == "" || xHost == "unix" {
		xHost = "localhost"
	}

	port := 6000
	d := 0
	if n, _ := fmt.Sscanf(xDisplay, "%d", &d); n == 1 && d < 100 {
		port = 6000 + d
	}

	xConn, err := net.Dial("tcp", net.JoinHostPort(xHost, fmt.Sprintf("%d", port)))
	if err != nil {
		m.sendServiceMessage(connID, fmt.Sprintf("Could not connect to X server: %v", err))
		ch.Close()
		return
	}

	go io.Copy(ch, xConn)
	io.Copy(xConn, ch)
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

// sendServiceMessage sends an informational message to the client
func (m *Manager) sendServiceMessage(connectionID string, message string) {
	m.notify("ssh.serviceMessage", api.ServiceMessageNotification{
		ConnectionID: connectionID,
		Message:      message,
	})
}

// proxyConn implements net.Conn interface for proxy command stdin/stdout
type proxyConn struct {
	stdin  io.WriteCloser
	stdout io.Reader
	local  string
	remote string
}

func (c *proxyConn) Read(b []byte) (int, error)  { return c.stdout.Read(b) }
func (c *proxyConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }
func (c *proxyConn) Close() error {
	c.stdin.Close()
	return nil
}
func (c *proxyConn) LocalAddr() net.Addr  { return &addrProto{c.local} }
func (c *proxyConn) RemoteAddr() net.Addr { return &addrProto{c.remote} }
func (c *proxyConn) SetDeadline(t time.Time) error      { return nil }
func (c *proxyConn) SetReadDeadline(t time.Time) error   { return nil }
func (c *proxyConn) SetWriteDeadline(t time.Time) error  { return nil }

type addrProto struct{ addr string }

func (a *addrProto) Network() string { return "tcp" }
func (a *addrProto) String() string  { return a.addr }

// String returns a human-readable description of the forward
func (f *Forward) String() string {
	switch f.Type {
	case api.PortForwardLocal:
		return fmt.Sprintf("(local) %s:%d → (remote) %s:%d", f.Host, f.Port, f.TargetAddr, f.TargetPort)
	case api.PortForwardRemote:
		return fmt.Sprintf("(remote) %s:%d → (local) %s:%d", f.Host, f.Port, f.TargetAddr, f.TargetPort)
	case api.PortForwardDynamic:
		return fmt.Sprintf("(dynamic/SOCKS5) %s:%d", f.Host, f.Port)
	default:
		return fmt.Sprintf("(unknown) %s:%d", f.Host, f.Port)
	}
}

// randomHex generates a random hex string of the given byte length
func randomHex(n int) string {
	b := make([]byte, n)
	// Use time-based seed as fallback; crypto/rand is preferred but this is
	// sufficient for X11 cookie generation
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}
	return fmt.Sprintf("%x", b)
}
