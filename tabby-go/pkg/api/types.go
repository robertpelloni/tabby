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
	JumpHost          *SSHConnectParams `json:"jumpHost,omitempty"`
	Algorithms        *SSHAlgorithms    `json:"algorithms,omitempty"`
	ProxyCommand      string            `json:"proxyCommand,omitempty"`
	Environment       map[string]string `json:"environment,omitempty"`
}

// SSHAuthParams specifies the authentication method and credentials
type SSHAuthParams struct {
	Type              string   `json:"type"` // "password", "publicKey", "agent", "keyboardInteractive"
	Password          string   `json:"password,omitempty"`
	PrivateKey        string   `json:"privateKey,omitempty"`
	PrivateKeyPaths   []string `json:"privateKeyPaths,omitempty"`
	AgentSocketPath   string   `json:"agentSocketPath,omitempty"`
	KeyboardInteractivePassthrough bool `json:"keyboardInteractivePassthrough,omitempty"`
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
	ConnectionID   string   `json:"connectionId"`
	ServerVersion  string   `json:"serverVersion"`
	RemoteAddress  string   `json:"remoteAddress"`
	Banner         string   `json:"banner,omitempty"`
	AuthMethods    []string `json:"authMethods"`
}

// SSHSessionResult is returned after starting a shell session
type SSHSessionResult struct {
	SessionID string `json:"sessionId"`
}

// ---- PTY API Types ----

// PTYSpawnParams contains parameters for spawning a local PTY
type PTYSpawnParams struct {
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
	Port         string `json:"port"`
	BaudRate     int    `json:"baudRate"`
	DataBits     int    `json:"dataBits,omitempty"`     // 5,6,7,8 (default 8)
	StopBits     int    `json:"stopBits,omitempty"`     // 1,2 (default 1)
	Parity       string `json:"parity,omitempty"`       // "none","even","odd" (default "none")
	FlowControl  string `json:"flowControl,omitempty"`  // "none","hardware","software" (default "none")
}

// SerialOpenResult is returned after opening a serial port
type SerialOpenResult struct {
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
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    uint32 `json:"mode"`
	ModTime string `json:"modTime"`
	IsDir   bool   `json:"isDir"`
}

// SFTPDownloadParams downloads a file
type SFTPDownloadParams struct {
	SessionID string `json:"sessionId"`
	RemotePath string `json:"remotePath"`
	LocalPath  string `json:"localPath"`
}

// SFTPUploadParams uploads a file
type SFTPUploadParams struct {
	SessionID  string `json:"sessionId"`
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
}

// SFTPTransferResult is returned after a file transfer
type SFTPTransferResult struct {
	BytesTransferred int64 `json:"bytesTransferred"`
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
	TransferID      string `json:"transferId"`
	BytesTransferred int64  `json:"bytesTransferred"`
	TotalBytes      int64  `json:"totalBytes"`
	Complete        bool   `json:"complete"`
	Error           string `json:"error,omitempty"`
}

// HostKeyPromptNotification is sent when an unknown host key is encountered
type HostKeyPromptNotification struct {
	ConnectionID string `json:"connectionId"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	KeyType      string `json:"keyType"`
	Fingerprint  string `json:"fingerprint"`
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
