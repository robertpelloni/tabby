# tabby-go — Tabby Go Backend

Go-based backend for Tabby terminal emulator, providing performance-critical native functionality via a JSON-RPC 2.0 interface.

## Status

**Proof of Concept** — SSH client and SFTP are functional, PTY and Serial are stubs.

## Architecture

```
┌──────────────────────────────────────┐
│         Electron / TypeScript        │
│     (tabby-electron, tabby-core)     │
├──────────────────────────────────────┤
│   JSON-RPC 2.0 over stdin/stdout     │
├──────────────────────────────────────┤
│          Go Backend                  │
│  ┌──────────┐ ┌──────────┐          │
│  │ SSH Mgr  │ │ SFTP Mgr │          │
│  └──────────┘ └──────────┘          │
│  ┌──────────┐ ┌──────────┐          │
│  │ PTY Mgr  │ │Serial Mgr│          │
│  └──────────┘ └──────────┘          │
└──────────────────────────────────────┘
```

## Building

```bash
cd tabby-go
go build ./...
go build -o ../build/tabby-backend.exe ./cmd/tabby-backend
```

## Running

The backend is designed to be spawned by Electron as a child process:

```bash
# stdio mode (default)
tabby-backend

# Check version
tabby-backend --version
```

## JSON-RPC API

### SSH

| Method | Description |
|--------|-------------|
| `ssh.connect` | Connect to SSH server |
| `ssh.startShell` | Start shell session |
| `ssh.resize` | Resize terminal |
| `ssh.write` | Write data to session |
| `ssh.close` | Close session/connection |
| `ssh.listConnections` | List active connections |

### SFTP

| Method | Description |
|--------|-------------|
| `sftp.open` | Open SFTP session |
| `sftp.list` | List directory contents |
| `sftp.download` | Download file |
| `sftp.upload` | Upload file |
| `sftp.delete` | Delete file/directory |
| `sftp.rename` | Rename file/directory |
| `sftp.mkdir` | Create directory |
| `sftp.stat` | Get file info |
| `sftp.close` | Close SFTP session |

### PTY (stub)

| Method | Description |
|--------|-------------|
| `pty.spawn` | Spawn local PTY |
| `pty.resize` | Resize PTY |
| `pty.write` | Write data |
| `pty.kill` | Kill process |

### Serial (stub)

| Method | Description |
|--------|-------------|
| `serial.open` | Open serial port |
| `serial.write` | Write data |
| `serial.close` | Close port |

### Notifications (server → client)

| Method | Description |
|--------|-------------|
| `ssh.data` | Data received from SSH session |
| `ssh.exit` | SSH session exited |

## Testing

```bash
go test ./...
```

## Dependencies

- `golang.org/x/crypto/ssh` — SSH protocol implementation
- `github.com/pkg/sftp` — SFTP protocol
- `golang.org/x/term` — Terminal handling
