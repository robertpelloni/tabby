# Tabby Go

A modern, native terminal emulator built with Go (Wails) + xterm.js, replacing the original Electron-based Tabby architecture for improved performance and reduced resource usage.

## Architecture

| Component | Technology |
|-----------|-----------|
| Backend | Go (Wails v2) |
| Terminal | xterm.js 5.x with addons |
| Frontend | Vanilla JS + CSS |
| Bindings | 68 Wails RPC methods |
| Config | TOML-based settings |

## Go Packages (17 tested)

| Package | Description |
|---------|-------------|
| `pkg/ssh` | SSH client with password/key/agent/keyboard-interactive auth |
| `pkg/sftp` | SFTP client with 17 operations |
| `pkg/telnet` | Telnet client with NAWS resize |
| `pkg/serial` | Serial port communication |
| `pkg/pty` | Cross-platform PTY (ConPTY/Unix) |
| `pkg/vault` | AES-256-CBC encrypted credential storage |
| `pkg/keychain` | OS keychain integration with vault fallback |
| `pkg/profile` | Connection profile management |
| `pkg/session` | Session persistence and recovery |
| `pkg/settings` | TOML configuration management |
| `pkg/config` | Application configuration |
| `pkg/audit` | Connection audit logging with rotation |
| `pkg/updater` | GitHub releases update checker |
| `pkg/notification` | System notification backend |
| `pkg/hotkey` | Global hotkey registration |
| `pkg/knownhosts` | SSH known hosts management |
| `pkg/colorscheme` | Color scheme definitions |
| `pkg/recovery` | Crash recovery mechanisms |
| `pkg/api` | JSON-RPC API layer |
| `pkg/middleware` | RPC middleware |

## Connection Types

- **Local Shell**: PTY (ConPTY on Windows, `creack/pty` on Unix), multi-shell, reconnect
- **SSH**: Password/key/agent/keyboard-interactive auth, jump hosts, ProxyJump, keepalive, timeout, agent forwarding, passphrase
- **Serial Port**: Configurable baud rate, data bits, stop bits, parity
- **Telnet**: Full protocol support, NAWS resize, reconnect

## SSH Features

- SFTP browser with 17 operations + drag & drop upload + directory picker downloads
- Port forwarding (local/remote/dynamic SOCKS5)
- Jump Host chain with ProxyJump support
- Login scripts (auto-run on connect)
- Host key verification dialog
- SSH config import from `~/.ssh/config`
- Key passphrase support
- Agent forwarding toggle
- Keepalive & timeout configuration
- Keyboard-interactive auth with inline password dialog

## Terminal & UI

- **Tab Management**: Status indicators, color labels, badge notifications, drag-and-drop reordering
- **Split Pane**: Vertical/horizontal with auto-resize during drag
- **Command Palette**: 38+ commands (Ctrl+Shift+P)
- **Terminal Context Menu**: Copy/Paste/Search Web/Clear
- **Terminal Title Tracking**: OSC 0/2 sequences update tab + window title
- **Scroll-to-Bottom Button**: Floating button when scrolled up
- **Settings Panel**: Font, size, scrollback, cursor, idle timeout, color scheme, shell
- **Quick Connect**: Welcome screen with action buttons
- **Notification Center**: Bell button with event history
- **Custom CSS**: User-defined styles applied on load
- **Clickable URLs**: Open in default browser
- **Multi-line Paste Warning**: Confirms before pasting multi-line content
- **Loading Spinner**: Animated spinner during SSH connection
- **13+ Color Schemes**: Including light theme
- **Thin Overlay Scrollbar**: Custom scrollbar styling
- **Application Menu Bar**: File/Edit/View/Help menus

## Data & Persistence

- **Profile CRUD**: Groups, duplicate, type switching, import/export (JSON)
- **Profile Group Collapse/Expand**: Click to toggle, persists in localStorage
- **Session Persistence & Auto-Reconnect**: Survives app restart
- **Reconnect Overlay**: Connection lost dialog with Reconnect/Close buttons
- **Snippets System**: Save and run command snippets
- **Connection Log Viewer**: Filterable modal with color-coded entries

## Security & Operations

- **AES-256-CBC Encrypted Vault**: Secure local credential storage
- **OS Keychain Integration**: With vault fallback for unsupported platforms
- **Connection Audit Logging**: JSON-line format with 10MB rotation
- **Auto-Updater**: Checks GitHub releases on startup (silent, 5s delay)
- **Idle Connection Monitor**: Auto-disconnects after configurable timeout
- **Inline Password Dialog**: Modal for keyboard-interactive SSH auth

## File Transfer

- **Zmodem Support**: rz/sz via addon integration
- **SFTP Upload/Download**: Drag & drop, directory picker

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Shift+P | Command Palette |
| Ctrl+Shift+F | Tab Search |
| Ctrl+Shift+S | Serial Dialog |
| Ctrl+Shift+N | Telnet Dialog |
| Ctrl+Shift+T | New Tab |
| Ctrl+Shift+L | Connection Log |
| Ctrl+Shift+O | Settings Panel |
| Ctrl+Shift+E | Export Profiles |
| Ctrl+W | Close Tab |
| Ctrl+Tab | Next Tab |
| Ctrl+Shift+Tab | Previous Tab |
| Alt+1-9 | Switch to Tab N |

## Build

```bash
# Prerequisites
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Development
cd wails-app && wails dev

# Production Build
cd wails-app && wails build

# Run Tests
cd tabby-go && go test ./...
```

## Code Stats

| File | Lines |
|------|-------|
| main.js | ~3,900 |
| app.go | 519 |
| app.css | ~1,400 |
| index.html | 221 |
| **Total** | **~6,000+** |

## Project Structure

```
tabby-go/
├── pkg/
│   ├── ssh/          # SSH client
│   ├── sftp/         # SFTP client
│   ├── telnet/       # Telnet client
│   ├── serial/       # Serial port
│   ├── pty/          # Cross-platform PTY
│   ├── vault/        # Encrypted vault
│   ├── keychain/     # OS keychain
│   ├── profile/      # Connection profiles
│   ├── session/      # Session persistence
│   ├── settings/     # TOML settings
│   ├── audit/        # Audit logging
│   ├── updater/      # Update checker
│   └── ...           # Other packages
└── wails-app/
    ├── app.go        # Wails bindings (68 methods)
    ├── main.go       # App entrypoint
    └── frontend/
        ├── index.html
        └── src/
            ├── main.js   # App logic (~3,900 lines)
            └── app.css   # Styling (~1,400 lines)
```

## License

MIT
