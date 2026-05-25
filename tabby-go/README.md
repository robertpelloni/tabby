# tabby-go — Tabby Go Backend

Go-based backend for Tabby terminal emulator, providing performance-critical native functionality via a JSON-RPC 2.0 interface.

## Status

**Active Development** — SSH, SFTP, PTY, Serial, Port Forwarding, BTK native UI integration.

## Architecture

```
┌──────────────────────────────────────────────┐
│         Electron / TypeScript Frontend        │
│         or BTK Native UI (future)             │
├──────────────────────────────────────────────┤
│        JSON-RPC 2.0 over stdin/stdout         │
│               (40+ methods)                   │
├──────────────────────────────────────────────┤
│              Go Backend                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ SSH Mgr  │  │ SFTP Mgr │  │ PTY Mgr  │   │
│  │ • shells │  │ • list   │  │ • spawn  │   │
│  │ • auth   │  │ • upload │  │ • resize │   │
│  │ • jumps  │  │ • down   │  │ • kill   │   │
│  │ • fwd    │  │ • chmod  │  └──────────┘   │
│  │ • x11    │  │ • sym    │                  │
│  │ • proxy  │  │ • mkdir  │  ┌──────────┐   │
│  └──────────┘  └──────────┘  │Serial Mgr│   │
│                               │ • open   │   │
│  ┌──────────┐                 │ • write  │   │
│  │ BTK UI   │                 │ • close  │   │
│  │ (CGo)    │                 └──────────┘   │
│  └──────────┘                                 │
└──────────────────────────────────────────────┘
```

## Building

```bash
cd tabby-go

# Build the JSON-RPC backend (pure Go, no CGo)
go build -mod=mod -o ../build/tabby-backend ./cmd/tabby-backend

# Run tests
go test -mod=mod ./...
```

## Running

The backend is designed to be spawned by Electron as a child process:

```bash
# stdio mode (default) — used by Electron
tabby-backend

# Check version
tabby-backend --version
```

## JSON-RPC API (40+ methods)

### Lifecycle
| Method | Description |
|--------|-------------|
| `ping` | Health check with version info |

### SSH
| Method | Description |
|--------|-------------|
| `ssh.connect` | Connect to SSH server |
| `ssh.startShell` | Start shell session |
| `ssh.resize` | Resize terminal |
| `ssh.write` | Write data to session |
| `ssh.close` | Close session/connection |
| `ssh.listConnections` | List active connections |

### SSH Port Forwarding
| Method | Description |
|--------|-------------|
| `ssh.addForward` | Add port forward (local/remote/dynamic) |
| `ssh.removeForward` | Remove port forward |
| `ssh.listForwards` | List active port forwards |

### SSH Authentication
| Method | Description |
|--------|-------------|
| `ssh.verifyHostKey` | Accept/reject host key prompt |
| `ssh.keyboardInteractiveResp` | Respond to keyboard-interactive auth |

### SFTP
| Method | Description |
|--------|-------------|
| `sftp.open` | Open SFTP session |
| `sftp.list` | List directory contents |
| `sftp.readDir` | List directory with symlink info |
| `sftp.download` | Download file |
| `sftp.upload` | Upload file |
| `sftp.delete` | Delete file/directory |
| `sftp.rename` | Rename file/directory |
| `sftp.mkdir` | Create directory |
| `sftp.mkdirAll` | Create directory tree |
| `sftp.stat` | Get file info |
| `sftp.lstat` | Get file info (no follow symlinks) |
| `sftp.chmod` | Change file permissions |
| `sftp.readlink` | Read symbolic link target |
| `sftp.symlink` | Create symbolic link |
| `sftp.rmdir` | Remove directory |
| `sftp.close` | Close SFTP session |

### PTY
| Method | Description |
|--------|-------------|
| `pty.spawn` | Spawn local PTY |
| `pty.resize` | Resize PTY |
| `pty.write` | Write data |
| `pty.kill` | Kill process |

### Serial
| Method | Description |
|--------|-------------|
| `serial.open` | Open serial port |
| `serial.write` | Write data |
| `serial.close` | Close port |
| `serial.listPorts` | List available ports |

### Notifications (server → client)

| Method | Description |
|--------|-------------|
| `ssh.data` | Data received from SSH/PTY/serial session |
| `ssh.exit` | Session exited |
| `pty.data` / `pty.exit` | PTY data/exit |
| `serial.data` / `serial.exit` | Serial data/exit |
| `ssh.banner` | SSH server banner |
| `ssh.hostKeyPrompt` | Host key verification prompt |
| `ssh.keyboardInteractive` | Keyboard-interactive auth prompt |
| `ssh.serviceMessage` | Informational message |
| `ssh.portForwardEvent` | Port forward connection event |

## SSH Features

### Authentication Methods
- **Password**: Direct password authentication
- **Public Key**: Private key file or inline key data
- **Agent**: SSH agent (Unix socket / Windows named pipe / Pageant)
- **Keyboard-Interactive**: Interactive challenge-response with client-side forwarding
- **None**: No authentication (testing)

### Connection Methods
- Direct TCP connection
- Jump host / proxy jump (supports nested chains)
- Proxy command (e.g., `nc %h %p`)
- SOCKS5 proxy
- HTTP CONNECT proxy

### Port Forwarding
- **Local**: `localhost:localPort → remote:targetPort` via SSH tunnel
- **Remote**: `remote:remotePort → localhost:targetPort` (reverse tunnel)
- **Dynamic**: SOCKS5 proxy on `localhost:port` (routes traffic through SSH)

### Other SSH Features
- X11 forwarding (channel-level support)
- Agent forwarding
- Known hosts verification
- Keepalive with disconnect detection
- Custom algorithm selection (KEX, cipher, MAC, compression)

## SFTP Features

- Directory listing with symlink detection and resolution
- File upload/download with progress
- Recursive directory creation
- Symbolic link creation and reading
- File permission changes (chmod)
- File info with and without symlink following

## BTK Native UI

The `pkg/ui/` package provides Go bindings for BTK native widgets via CGo:

- **bridge.h**: C API header (flat functions for CGo compatibility)
- **bridge.cpp**: C++ implementation wrapping BTK's Qt-descended classes
- **ui.go**: Go bindings with type-safe API

### Supported Widgets
App, Window, TabWidget, Splitter, Terminal, MenuBar, Menu, Action,
ToolBar, StatusBar, Label, LineEdit, Button, ComboBox, Layout,
Dialog, FileDialog

### Building with BTK
```bash
# Requires BTK to be compiled first
cmake -B build tabby-go/vendor/btk
cmake --build build
# Then link against BTK libraries
```

## Testing

```bash
go test -mod=mod ./...
```

## Dependencies

- `golang.org/x/crypto/ssh` — SSH protocol implementation
- `github.com/pkg/sftp` — SFTP protocol
- `github.com/robertpelloni/btk` — Native UI toolkit (submodule, optional)

## Project Structure

```
tabby-go/
├── cmd/
│   └── tabby-backend/     # JSON-RPC backend entry point
├── internal/
│   └── server/            # JSON-RPC server with method routing
├── pkg/
│   ├── api/               # Shared types (SSH, SFTP, PTY, Serial, notifications)
│   ├── ssh/               # SSH connection manager
│   ├── sftp/              # SFTP file operations
│   ├── pty/               # PTY process management
│   ├── serial/            # Serial port communication (stub)
│   ├── ui/                # BTK native UI bindings (CGo)
│   │   ├── bridge.h       # C API header
│   │   ├── bridge.cpp     # C++ implementation
│   │   └── ui.go          # Go bindings
│   └── nativeapp/         # Native app orchestration
└── vendor/
    ├── btk/               # BTK submodule
    └── (Go dependencies)  # Vendored Go packages
```
