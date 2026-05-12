# Changelog

## [1.0.231-nightly.10] - 2026-05-12
### Added
- Created `ReactPluginDecorator` in the Angular frontend that exposes a `window['tabbyReactPlugins']` interface, giving users the ability to inject native React or DOM extensions over the Tabby window seamlessly (Hyper parity).
- Initialized official Sentry Go error tracking on the Go daemon inside `main.go`. This automatically recovers and captures native panics on the daemon via the `SENTRY_DSN` env var.


## [1.0.231-nightly.8] - 2026-05-11
### Added
- Integrated actual OpenAI API requests to the Go backend (`tabby-go/pkg/ai`) triggered by the `OPENAI_API_KEY` environment variable. If missing, it gracefully falls back to the local mock behavior.
- Implemented Hyper-style hot-reloading configurations. The Go backend (`tabby-go/pkg/config`) now utilizes an `os.Stat` polling loop to watch the active YAML configuration file and broadcasts JSON-RPC `host:config-change` events instantly to the Angular frontend for zero-restart visual updates.


## [1.0.231-nightly.7] - 2026-05-11
### Changed
- Extended `BlockFrontend` OSC rendering to support parsing and safely displaying inline image widgets and iframe/webview widgets, utilizing DOMPurify to prevent arbitrary XSS.
- Refined the `tabby-go/pkg/ai` backend mock to intelligently parse errors for the `ExplainError` action, emitting structured Markdown diagnostic tips for common terminal failures like 'Permission Denied', 'Port in Use', and 'Command Not Found'.


## [1.0.231-nightly.6] - 2026-05-11
### Changed
- Refactored `BlockFrontend` to isolate terminal output from UI actions, allowing clean "Copy Command" and "Copy Output" actions.
- Implemented natural language to shell command generation logic in the AI backend integration mock.


All notable changes to the Tabby (robertpelloni fork) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Version number is maintained in [VERSION.md](VERSION.md) and is the single source of truth.

## [1.0.231-nightly.5] - 2026-04-23

### Changed
- Version bump

## [1.0.231-nightly.4] - 2026-04-15

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **Rich Widget Block Frontend** (Phase 2 Roadmap: WaveTerm Parity)
  - Implemented `BlockFrontend` inside `tabby-terminal` to render terminal output into discrete HTML/DOM elements rather than an xterm.js canvas.
  - Intercepts ANSI OSC 1337 (`\x1b]1337;WaveTermWidget=...`) payloads from the backend to instantly render Markdown, code, or images directly in the output stream.

### Changed
- Version bump

## [1.0.231-nightly.3] - 2025-04-09

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **Terminal Middleware** (`pkg/middleware/`, 400 LOC)
  - UTF8Splitter: Buffers incomplete multibyte UTF-8 sequences
  - InputProcessor: Backspace key mapping (ctrl-h, ctrl-?, delete, backspace)
  - LoginScriptProcessor: Expect/send pattern matching with regex support
  - OSCProcessor: OSC 1337 parsing (CurrentDir with tilde expansion)
  - StreamProcessor: Newline mode conversion (CR/LF/CRLF/implicit)
- **Known Hosts Manager** (`pkg/knownhosts/`, 200 LOC)
  - Thread-safe in-memory storage with SHA-256 fingerprints
  - Host key verification (match/mismatch/unknown)
  - OpenSSH known_hosts file format load/save
  - IPv6 bracket notation handling
- **Session Recovery** (`pkg/recovery/`, 160 LOC)
  - Tab state registration and persistence (JSON)
  - Session tracking with connected/disconnected transitions
  - Platform-specific recovery path (`~/.tabby/recovery.json`)
- **Notification System** (`pkg/notification/`, 120 LOC)
  - Info/Warning/Error severity levels
  - Read/unread tracking with 100-entry cap
  - OnChange callbacks for real-time UI updates
- **21 new RPC methods** (65+ total): knownHosts.*, notifications.*, recovery.*
- **33 new tests** (157 total, all passing)

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **Rich Widget Block Frontend** (Phase 2 Roadmap: WaveTerm Parity)
  - Implemented `BlockFrontend` inside `tabby-terminal` to render terminal output into discrete HTML/DOM elements rather than an xterm.js canvas.
  - Intercepts ANSI OSC 1337 (`\x1b]1337;WaveTermWidget=...`) payloads from the backend to instantly render Markdown, code, or images directly in the output stream.

### Changed
- Binary size: 8.05MB (from 7.95MB)
- Go codebase: 9,686 LOC across 32 source files, 14 packages

## [1.0.231-nightly.2] - 2025-04-08

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **SSH Port Forwarding**: Full local, remote, and dynamic/SOCKS5 forwarding
- **SSH Proxy Support**: Proxy command, SOCKS5 proxy, HTTP CONNECT proxy
- **SSH Host Key Verification**: Interactive prompts and known_hosts file support
- **SSH Keyboard-Interactive Auth**: Client-side prompt forwarding via notifications
- **SSH X11 Forwarding**: Channel-level X11 forwarding support
- **SSH Agent Forwarding**: Agent forwarding over SSH sessions
- **Nested Jump Hosts**: Support for chained jump host connections
- **SFTP Enhancements**: chmod, readlink, symlink, rmdir, lstat, readDir, mkdirAll
- **BTK Native UI Integration**: CGo bridge (bridge.h/bridge.cpp) and Go bindings (pkg/ui/)
  - Full widget set: Window, TabWidget, Splitter, Terminal, MenuBar, Menu,
    Action, ToolBar, StatusBar, Label, LineEdit, Button, ComboBox, Layout, Dialog
  - File dialogs, clipboard, dark mode, screen info
- **BTK Submodule**: Added robertpelloni/btk as git submodule for native UI toolkit
- **TypeScript GoBackendService**: Full API coverage for all 40+ JSON-RPC methods
  - Port forwarding methods (addForward, removeForward, listForwards)
  - Auth callbacks (verifyHostKey, keyboardInteractiveResp)
  - Enhanced SFTP (chmod, readlink, symlink, lstat, readDir, mkdirAll, rmdir)
  - RxJS observables for host key prompts, keyboard-interactive, banners, service messages
- **Go API Types**: Expanded to 500+ lines covering port forwarding, serial ports, SFTP operations
- **Version bump**: 1.0.231-nightly.1 → 1.0.231-nightly.2

## [1.0.231-nightly.1] - 2025-04-08

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **Go PTY manager**: Process spawning with stdin/stdout/stderr forwarding
- **Go serial port manager**: Stub for go.bug.st/serial integration
- **TypeScript GoBackendService**: Angular service for Go backend communication
  - JSON-RPC 2.0 client over stdin/stdout
  - RxJS observables for data/exit notifications
  - Full TypeScript API for SSH, SFTP, PTY, Serial
- **Config option**: `goBackend.enabled` to switch between russh and Go backends

## [1.0.231-nightly.0] - 2025-04-08

### Upstream Sync
- Synced fork with upstream Eugeny/tabby master (f05a07ae)
- Configured upstream remote (https://github.com/Eugeny/tabby)

### Added
- **russh-based SSH backend**: The SSH client now uses the `russh` Rust library (v0.1.36) via N-API bindings instead of the previous ssh2-based implementation. This provides better performance and reliability for SSH connections.
- **SFTP context menu download**: Files can now be downloaded via SFTP context menu directly from SSH sessions (#10971)
- **tabby:// URL scheme handler**: Custom URL protocol handler allows opening Tabby via `tabby://` URLs (#11005)
- **Quick connection support for third-party plugins**: Plugins can now add quick connection providers (#11004)
- **X11 forwarding fix**: Fixed X11 forwarding for SSH sessions (#10580, #10558, #10237, #8450)
- **Webpack upgrade**: Upgraded webpack from 5.86.0 to 5.104.1 to fix ES6 class extension bug with xterm.js v6 (#11007)
- **Recovered tabs crash fix**: Fixed crashes when recovering tabs after upgrade (#10619, #10635, #10672, #10682, #10688, #10730, #10843)
- **macOS dock hiding fix**: Fixed hiding the app from macOS dock when docking is enabled (#10960)

### Added (robertpelloni fork)
- **Terminal Agent Chat UI** (Phase 3 Roadmap: Warp Parity)
  - Built a toggleable chat sidebar into the `baseTerminalTab` providing a dedicated interface for natural-language debugging and agentic chat. The interface routes input through `ai:explainError` placeholder IPC hooks.
- **Port Forwarding Management UI** (Phase 6 Roadmap)
  - Connected the `SSHPortForwardingModalComponent` logic seamlessly over JSON-RPC 2.0 to the Go backend.
- **Jump Host Visualization** (Phase 6 Roadmap)
  - The terminal toolbar explicitly displays nested proxy-jump chains now (e.g., `via bastion1 -> bastion2`).

- **React WebComponent Plugin API Wrapper** (Phase 5 Roadmap: Hyper Parity)
  - Implemented `ReactPluginDecorator` in `tabby-terminal/src/api/reactPlugin.ts`.
  - Allows developers to write standalone scripts exposing DOM/React nodes into `window.tabbyReactPlugins.register` to seamlessly overlay arbitrary UI elements on top of the terminal interface, avoiding the need for heavy Angular plugins for simple extensions.

- **Command Catalog UI** (Phase 4 Roadmap: Warp Parity)
  - Appended a Command Library lookup action on the IDE input toolbar.
  - Built `CommandCatalogModalComponent` exposing searchable, preset workflows and commands parameterized via curly brackets.

- **Block-Based Output Parsing** (Phase 2 Roadmap)
  - Adapted `ansi-to-html` integration inside the new `BlockFrontend` so PTY data execution blocks accurately map foreground/background escape sequences natively into standard DOM elements instead of raw text elements.
- **Warp-Style IDE Input Area** (Phase 2 Roadmap)
  - Implemented `ideInput` `<textarea>` toggle directly into `baseTerminalTab`.
  - Keydown handlers appropriately forward buffered multi-line input down into the PTY stream via `sendInput()`.
- **SSH Algorithms Native Migration** (Phase 1 Roadmap)
  - Replaced the internal `russh` Node native binding dependency that queried OpenSSH capabilities within `tabby-ssh/src/algorithms.ts` to use natively exported constants matching the Go daemon.
- **SSH PTY IPC End-to-End Integration** (Phase 1 Roadmap)
  - Registered all `ipcMain.handle('ssh:*')` routes in `app/lib/ssh.ts` to seamlessly tunnel data buffering from the Angular frontend down into the Go terminal backend PTY, permanently stripping the `russh` Node native bindings for core terminal execution.

- **SFTP IPC Proxy** (Phase 6 Roadmap)
  - Rewrote `tabby-ssh/src/session/sftp.ts` to entirely remove the `russh` dependency.
  - `readdir`, `readlink`, `stat`, `chmod`, `mkdir`, `rmdir`, `upload`, and `download` operations are now seamlessly marshalled via JSON-RPC 2.0 to the unified `tabby-go` daemon over `ipcRenderer`.
- **SSH Multiplexing Pool Manager** (Phase 1 Roadmap)
  - Implemented `Connection` pool manager in `tabby-go/pkg/ssh/ssh.go`.
  - Supports tracking connection `fingerprint` and returning the same active client on subsequent connection attempts to the same host.
  - Supports `RefCount` to keep the active TCP socket running until all multiplexed child sessions are torn down.
- **Go backend proof of concept** (`tabby-go/`):
  - SSH client with password, public key, and agent authentication
  - Shell sessions with PTY resize support
  - Jump host / proxy jump chains
  - Keepalive with disconnect detection
  - SFTP file management (list, upload, download, delete, rename, mkdir)
  - JSON-RPC 2.0 server over stdin/stdout
  - 25+ RPC methods for SSH, SFTP, PTY, and Serial
  - Unit tests (all passing)
  - Buildable binary (tabby-backend.exe, 7.2MB)
- **Version management system**:
  - `VERSION.md` as single source of truth
  - `scripts/bump-version.mjs` for automated version synchronization
  - Fixed `app/package.json` version mismatch
- **Comprehensive documentation suite**:
  - `VERSION.md` — Version string
  - `VISION.md` — Project vision and architecture
  - `CHANGELOG.md` — This changelog
  - `ROADMAP.md` — Long-term structural plans (5 phases)
  - `TODO.md` — Feature tasks organized by priority
  - `MEMORY.md` — Codebase observations and preferences
  - `DEPLOY.md` — Deployment instructions
  - `IDEAS.md` — 30 creative improvement ideas
  - `HANDOFF.md` — Session handoff documentation
  - `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` — Universal LLM instructions
  - `CLAUDE.md`, `GEMINI.md`, `GPT.md`, `copilot-instructions.md` — Model-specific instructions

### Project Modules
| Module | Version | Description |
|--------|---------|-------------|
| tabby-core | 1.0.231-nightly.0 | Core UI framework, tab management, services |
| tabby-electron | 1.0.231-nightly.0 | Electron-specific bindings (PTY, docking, shell integration) |
| tabby-terminal | 1.0.231-nightly.0 | Terminal emulation (xterm.js frontend, middleware, features) |
| tabby-ssh | 1.0.231-nightly.0 | SSH2 client with connection manager, SFTP, X11 forwarding |
| tabby-local | 1.0.231-nightly.0 | Local shell profiles (WSL, Git-Bash, PowerShell, CMD, etc.) |
| tabby-serial | 1.0.231-nightly.0 | Serial terminal connections |
| tabby-telnet | 1.0.231-nightly.0 | Telnet/socket connections |
| tabby-settings | 1.0.231-nightly.0 | Settings UI (profiles, hotkeys, vault, appearance) |
| tabby-plugin-manager | 1.0.231-nightly.0 | Plugin installation and management |
| tabby-linkifier | 1.0.231-nightly.0 | Clickable URLs, IPs, and file paths in terminal |
| tabby-auto-sudo-password | 1.0.231-nightly.0 | Auto-paste sudo password in SSH sessions |
| tabby-community-color-schemes | 1.0.231-nightly.0 | Community-contributed color schemes |
| tabby-web | 1.0.231-nightly.0 | Web-specific bindings for browser deployment |
| tabby-web-demo | 1.0.231-nightly.0 | Web demo application |
| tabby-uac | N/A | Windows UAC helper (C# solution, not a TS module) |
| app (Electron shell) | 1.0.0-alpha.1 | Main Electron application shell |
| web | N/A | Web app entry point |

### Key Technologies
- **Frontend**: TypeScript, Angular 15, Pug templates, SCSS
- **Backend/Native**: Electron 38, Node.js, node-pty, russh (Rust SSH via N-API)
- **Build**: Webpack 5, TypeScript 4.9, electron-builder
- **Terminal**: xterm.js v6 with image addon
- **UI Framework**: Bootstrap 5 (via ng-bootstrap), FontAwesome 6
- **State**: Angular services, RxJS observables
- **Config**: YAML-based configuration with platform-specific defaults
