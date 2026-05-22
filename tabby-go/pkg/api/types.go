// Package api defines the JSON-RPC API types shared between the Go backend
// and the TypeScript/Electron frontend.
//
// The Go backend communicates with the Electron app via JSON-RPC 2.0 over
// stdin/stdout (or a TCP socket). This allows the Electron app to spawn the
// Go backend as a child process and communicate bidirectionally.
package api

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONRPCNotification represents a JSON-RPC 2.0 notification (no ID, no response expected)
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ErrorParseError     = -32700
	ErrorInvalidRequest = -32600
	ErrorMethodNotFound = -32601
	ErrorInvalidParams  = -32602
	ErrorInternal       = -32603

	// Application-specific error codes
	ErrorAuthFailed     = -32001
	ErrorHostKeyReject  = -32002
	ErrorConnectFailed  = -32003
	ErrorSessionExpired = -32004
)

// ---- SSH API Types ----

// SSHConnectParams contains parameters for establishing an SSH connection
type SSHConnectParams struct {
	Host              string            `json:"host"`
	Port              int               `json:"port,omitempty"`
	Username          string            `json:"user"`
	Auth              SSHAuthParams     `json:"auth"`
	KeepaliveInterval int               `json:"keepaliveInterval,omitempty"`
	KeepaliveCountMax int               `json:"keepaliveCountMax,omitempty"`
	ReadyTimeout      int               `json:"readyTimeout,omitempty"`
	AgentForward      bool              `json:"agentForward,omitempty"`
	X11               bool              `json:"x11,omitempty"`
	X11Display        string            `json:"x11Display,omitempty"`
	JumpHost          *SSHConnectParams `json:"jumpHost,omitempty"`
	Algorithms        *SSHAlgorithms    `json:"algorithms,omitempty"`
	ProxyCommand      string            `json:"proxyCommand,omitempty"`
	SocksProxyHost    string            `json:"socksProxyHost,omitempty"`
	SocksProxyPort    int               `json:"socksProxyPort,omitempty"`
	HTTPProxyHost     string            `json:"httpProxyHost,omitempty"`
	HTTPProxyPort     int               `json:"httpProxyPort,omitempty"`
	Environment       map[string]string `json:"environment,omitempty"`
	VerifyHostKey     bool              `json:"verifyHostKey,omitempty"`
	KnownHostsPath    string            `json:"knownHostsPath,omitempty"`
	SkipBanner        bool              `json:"skipBanner,omitempty"`
	Password          string            `json:"password,omitempty"` // Saved password fallback
}

// SSHAuthParams specifies the authentication method and credentials
type SSHAuthParams struct {
	Type                          string   `json:"type"` // "password", "publicKey", "agent", "keyboardInteractive", "none"
	Password                      string   `json:"password,omitempty"`
	PrivateKey                    string   `json:"privateKey,omitempty"`
	PrivateKeyPaths               []string `json:"privateKeyPaths,omitempty"`
	AgentSocketPath               string   `json:"agentSocketPath,omitempty"`
	AgentType                     string   `json:"agentType,omitempty"` // "auto", "unix", "namedPipe", "pageant"
	KeyboardInteractivePassthrough bool     `json:"keyboardInteractivePassthrough,omitempty"`
}

// SSHAlgorithms specifies preferred algorithms for the SSH connection
type SSHAlgorithms struct {
	HMAC          []string `json:"hmac,omitempty"`
	KEX           []string `json:"kex,omitempty"`
	Cipher        []string `json:"cipher,omitempty"`
	ServerHostKey []string `json:"serverHostKey,omitempty"`
	Compression   []string `json:"compression,omitempty"`
}

// SSHSessionParams contains parameters for starting a shell session over SSH
type SSHSessionParams struct {
	ConnectionID string `json:"connectionId"`
	Columns      int    `json:"columns"`
	Rows         int    `json:"rows"`
	Terminal     string `json:"terminal,omitempty"` // e.g., "xterm-256color"
	Command      string `json:"command,omitempty"`  // Optional command instead of shell
	AgentForward bool   `json:"agentForward,omitempty"`
	X11          bool   `json:"x11,omitempty"`
}

// SSHResizeParams contains terminal resize parameters
type SSHResizeParams struct {
	ConnectionID string `json:"connectionId"`
	SessionID    string `json:"sessionId"`
	Columns      int    `json:"columns"`
	Rows         int    `json:"rows"`
}

// SSHWriteParams contains data to write to an SSH session
type SSHWriteParams struct {
	ConnectionID string `json:"connectionId"`
	SessionID    string `json:"sessionId"`
	Data         string `json:"data"` // Base64-encoded binary data
}

// SSHCloseParams closes an SSH connection or session
type SSHCloseParams struct {
	ConnectionID string `json:"connectionId"`
	SessionID    string `json:"sessionId,omitempty"` // If empty, closes the entire connection
}

// SSHConnectionResult is returned after successful connection
type SSHConnectionResult struct {
	ConnectionID  string   `json:"connectionId"`
	ServerVersion string   `json:"serverVersion"`
	RemoteAddress string   `json:"remoteAddress"`
	Banner        string   `json:"banner,omitempty"`
	AuthMethods   []string `json:"authMethods"`
	JumpChain     []string `json:"jumpChain,omitempty"` // Ordered list of jump host identifiers (e.g. ["bastion1", "bastion2"])
}

// SSHSessionResult is returned after starting a shell session
type SSHSessionResult struct {
	SessionID string `json:"sessionId"`
}

// ---- Port Forwarding Types ----

// PortForwardType defines the type of port forwarding
type PortForwardType string

const (
	PortForwardLocal   PortForwardType = "local"
	PortForwardRemote  PortForwardType = "remote"
	PortForwardDynamic PortForwardType = "dynamic"
)

// PortForwardParams contains parameters for adding a port forward
type PortForwardParams struct {
	ConnectionID  string          `json:"connectionId"`
	Type          PortForwardType `json:"type"`
	Host          string          `json:"host"`           // Listen host for local/dynamic, remote host for remote
	Port          int             `json:"port"`            // Listen port for local/dynamic, remote port for remote
	TargetAddress string          `json:"targetAddress"`   // Target address (local forward only)
	TargetPort    int             `json:"targetPort"`      // Target port (local forward only)
}

// PortForwardResult is returned after adding a port forward
type PortForwardResult struct {
	ForwardID string `json:"forwardId"`
}

// PortForwardRemoveParams removes a port forward
type PortForwardRemoveParams struct {
	ConnectionID string `json:"connectionId"`
	ForwardID    string `json:"forwardId"`
}

// PortForwardListResult lists active port forwards
type PortForwardListResult struct {
	Forwards []PortForwardInfo `json:"forwards"`
}

// PortForwardInfo describes an active port forward
type PortForwardInfo struct {
	ID            string          `json:"id"`
	Type          PortForwardType `json:"type"`
	Host          string          `json:"host"`
	Port          int             `json:"port"`
	TargetAddress string          `json:"targetAddress,omitempty"`
	TargetPort    int             `json:"targetPort,omitempty"`
	Active        bool            `json:"active"`
}

// ---- PTY API Types ----

// PTYSpawnParams contains parameters for spawning a local PTY
type PTYSpawnParams struct {
	ID      string            `json:"id"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Columns int               `json:"columns"`
	Rows    int               `json:"rows"`
}

// PTYSpawnResult is returned after spawning a PTY
type PTYSpawnResult struct {
	ID  string `json:"id"`
	PID int    `json:"pid"`
}

// PTYResizeParams resizes a PTY
type PTYResizeParams struct {
	ID      string `json:"id"`
	Columns int    `json:"columns"`
	Rows    int    `json:"rows"`
}

// PTYWriteParams writes data to a PTY
type PTYWriteParams struct {
	ID   string `json:"id"`
	Data string `json:"data"` // Base64-encoded
}

// PTYKillParams kills a PTY process
type PTYKillParams struct {
	ID     string `json:"id"`
	Signal string `json:"signal,omitempty"`
}

// ---- Serial API Types ----

// SerialOpenParams contains parameters for opening a serial port
type SerialOpenParams struct {
	ID          string `json:"id"`
	Port        string `json:"port"`
	BaudRate    int    `json:"baudRate"`
	DataBits    int    `json:"dataBits,omitempty"`    // 5,6,7,8 (default 8)
	StopBits    int    `json:"stopBits,omitempty"`    // 1,2 (default 1)
	Parity      string `json:"parity,omitempty"`      // "none","even","odd" (default "none")
	FlowControl string `json:"flowControl,omitempty"` // "none","hardware","software" (default "none")
}

// SerialOpenResult is returned after opening a serial port
type SerialOpenResult struct {
	ID string `json:"id"`
}

// SerialListPortsResult is returned when listing available serial ports
type SerialListPortsResult struct {
	Ports []SerialPortInfo `json:"ports"`
}

// SerialPortInfo describes an available serial port
type SerialPortInfo struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Product      string `json:"product,omitempty"`
	SerialNumber string `json:"serialNumber,omitempty"`
	VID          string `json:"vid,omitempty"`
	PID          string `json:"pid,omitempty"`
}

// SerialWriteParams writes data to a serial port
type SerialWriteParams struct {
	ID   string `json:"id"`
	Data string `json:"data"` // Base64-encoded
}

// SerialCloseParams closes a serial port
type SerialCloseParams struct {
	ID string `json:"id"`
}

// ---- SFTP API Types ----

// SFTPOpenParams opens an SFTP session over an existing SSH connection
type SFTPOpenParams struct {
	ConnectionID string `json:"connectionId"`
}

// SFTPOpenResult is returned after opening an SFTP session
type SFTPOpenResult struct {
	SessionID string `json:"sessionId"`
}

// SFTPListParams lists files in a directory
type SFTPListParams struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
}

// SFTPFile represents a file or directory in an SFTP listing
type SFTPFile struct {
	Name      string `json:"name"`
	FullPath  string `json:"fullPath"`
	Size      int64  `json:"size"`
	Mode      uint32 `json:"mode"`
	ModTime   string `json:"modTime"`
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
}

// SFTPDownloadParams downloads a file
type SFTPDownloadParams struct {
	SessionID  string `json:"sessionId"`
	RemotePath string `json:"remotePath"`
	LocalPath  string `json:"localPath"`
	TransferID string `json:"transferId,omitempty"`
}

// SFTPUploadParams uploads a file
type SFTPUploadParams struct {
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	TransferID string `json:"transferId,omitempty"`
	Data       string `json:"data,omitempty"` // Base64-encoded fallback
}

// SFTPTransferResult is returned after a file transfer
type SFTPTransferResult struct {
	BytesTransferred int64  `json:"bytesTransferred"`
	Data             string `json:"data,omitempty"` // Base64-encoded for Download fallback
}

// SFTPChmodParams changes file permissions
type SFTPChmodParams struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
	Mode      uint32 `json:"mode"`
}

// SFTPReadlinkParams reads a symbolic link
type SFTPReadlinkParams struct {
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
}

// SFTPSymlinkParams creates a symbolic link
type SFTPSymlinkParams struct {
	SessionID string `json:"sessionId"`
	OldPath   string `json:"oldPath"`
	NewPath   string `json:"newPath"`
}

// ---- SSH Auth Response Types ----

// HostKeyVerifyResponse is sent by the client to accept/reject a host key
type HostKeyVerifyResponse struct {
	ConnectionID string `json:"connectionId"`
	Accepted     bool   `json:"accepted"`
}

// KeyboardInteractiveResponse is sent by the client with auth responses
type KeyboardInteractiveResponse struct {
	ConnectionID string   `json:"connectionId"`
	Responses    []string `json:"responses"`
}

// ---- Notification Types (server → client) ----

// DataNotification is sent when data is received from a terminal/SSH/serial session
type DataNotification struct {
	ConnectionID string `json:"connectionId,omitempty"`
	SessionID    string `json:"sessionId,omitempty"`
	PTYID        string `json:"ptyId,omitempty"`
	SerialID     string `json:"serialId,omitempty"`
	Data         string `json:"data"` // Base64-encoded
}

// ExitNotification is sent when a session exits
type ExitNotification struct {
	ConnectionID string `json:"connectionId,omitempty"`
	SessionID    string `json:"sessionId,omitempty"`
	PTYID        string `json:"ptyId,omitempty"`
	SerialID     string `json:"serialId,omitempty"`
	ExitCode     int    `json:"exitCode,omitempty"`
	Signal       string `json:"signal,omitempty"`
}

// TransferProgressNotification is sent during file transfers
type TransferProgressNotification struct {
	TransferID       string `json:"transferId"`
	BytesTransferred int64  `json:"bytesTransferred"`
	TotalBytes       int64  `json:"totalBytes"`
	Complete         bool   `json:"complete"`
	Error            string `json:"error,omitempty"`
}

// HostKeyPromptNotification is sent when an unknown host key is encountered
type HostKeyPromptNotification struct {
	ConnectionID string `json:"connectionId"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	KeyType      string `json:"keyType"`
	Fingerprint  string `json:"fingerprint"`
	KeyBytes     string `json:"keyBytes,omitempty"` // Base64-encoded public key bytes
}

// KeyboardInteractiveNotification is sent for keyboard-interactive auth prompts
type KeyboardInteractiveNotification struct {
	ConnectionID string   `json:"connectionId"`
	Name         string   `json:"name"`
	Instruction  string   `json:"instruction"`
	Prompts      []Prompt `json:"prompts"`
}

// Prompt represents a single prompt in keyboard-interactive auth
type Prompt struct {
	Prompt string `json:"prompt"`
	Echo   bool   `json:"echo"`
}

// BannerNotification is sent when the SSH server sends a banner message
type BannerNotification struct {
	ConnectionID string `json:"connectionId"`
	Message      string `json:"message"`
}

// PortForwardEventNotification is sent when a port forward connection event occurs
type PortForwardEventNotification struct {
	ConnectionID  string `json:"connectionId"`
	ForwardID     string `json:"forwardId"`
	EventType     string `json:"eventType"` // "connected", "disconnected", "error"
	Message       string `json:"message,omitempty"`
	ClientAddress string `json:"clientAddress,omitempty"`
	ClientPort    int    `json:"clientPort,omitempty"`
}

// ServiceMessageNotification is sent for informational messages during connection
type ServiceMessageNotification struct {
	ConnectionID string `json:"connectionId"`
	Message      string `json:"message"`
}
