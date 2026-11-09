# Changelog

## [1.0.271] - 2026-11-04

### Added

- Session Persistence Protocol RPCs: Exposed `session.save`, `session.load`, and `session.clear` over the JSON-RPC daemon server in `tabby-go/internal/server`.
- Server RPC Session Tests: Added `TestSessionPersistenceRPC` in `tabby-go/internal/server/server_test.go` verifying save, load, and clear roundtrips over a duplex JSON-RPC pipe.

### Changed

- Version bump to 1.0.271

## [1.0.270] - 2026-11-04

### Added

- Last Tab Closes Window Behavior: Wired `LastTabClosesWindow` setting in Wails frontend `Tab.close()`, invoking `quitTabbyApplication()` to cleanly shut down active servers and terminate the application when the final tab is closed.

### Changed

- Version bump to 1.0.270

## [1.0.269] - 2026-11-04

### Added

- Agent Workflow Protocol RPCs: Exposed `agent.cancelWorkflow` and `agent.getWorkflow` over the JSON-RPC server in `tabby-go/internal/server`.
- Server RPC Unit Tests: Added `TestAgentWorkflowRPC` in `tabby-go/internal/server/server_test.go` verifying workflow start, status retrieval, and cancellation over the JSON-RPC pipe.

### Changed

- Version bump to 1.0.269

## [1.0.268] - 2026-11-04

### Added

- Command Palette Shortcut Badges: Displayed keyboard shortcut badges (`<span class="palette-shortcut">`) next to command items with dark/light theme styling in `tabby-go/wails-app/frontend`.

### Changed

- Version bump to 1.0.268

## [1.0.267] - 2026-11-04

### Added

- CI Wails Frontend Build Gate: Added `Build Wails frontend` step (`cd tabby-go/wails-app/frontend && npm install && npm run build`) into `.github/workflows/build.yml` to validate Wails UI production bundling during CI.

### Changed

- Version bump to 1.0.267

## [1.0.266] - 2026-11-04

### Added

- Tab Search Keyboard Navigation: Added `Enter` key handling in tab search input to instantly switch to the first matched tab and dismiss the search bar.

### Changed

- Version bump to 1.0.266

## [1.0.265] - 2026-11-04

### Added

- Agentic Workflow Cancellation: Added `StatusCancelled` task status, context cancellation lifecycle to `Workflow`, and `CancelWorkflow` method to `WorkflowManager` in `tabby-go/pkg/agent`.

### Changed

- Version bump to 1.0.265

## [1.0.264] - 2026-11-04

### Added

- Titlebar Close Confirmation: Wired `btn-close` click handler to `quitTabbyApplication()` to safeguard running sessions and ensure background servers disconnect before closing.

### Changed

- Version bump to 1.0.264

## [1.0.263] - 2026-11-04

### Added

- Hotkey Parsing Unit Tests: Added `test/core/hotkeysUtil.test.mjs` verifying key mapping (`getKeyName`) for modifiers, mouse, wheel, and alphanumeric codes, as well as strict canonical ordering (`getKeystrokeName`) from `tabby-core/src/services/hotkeys.util.ts`.

### Changed

- Version bump to 1.0.263

## [1.0.262] - 2026-11-04

### Added

- Core Utility Unit Testing: Added `test/core/resettableTimeout.test.mjs` verifying callback execution, timer clearing, and timer rescheduling for `ResettableTimeout` from `tabby-core/src/utils.ts`.

### Changed

- Version bump to 1.0.262

## [1.0.261] - 2026-11-04

### Added

- Terminal Font Zoom Keybindings: Bound `Ctrl+=` (zoom in), `Ctrl+-` (zoom out), and `Ctrl+0` (zoom reset) into global keyboard handler, with shortcut markers in the Command Palette.

### Changed

- Version bump to 1.0.261

## [1.0.260] - 2026-11-04

### Added

- Center Window Action: Imported `WindowCenter` runtime function and added Center Window on Screen action to Command Palette and System Tray dropdown.

### Changed

- Version bump to 1.0.260

## [1.0.259] - 2026-11-04

### Added

- Titlebar Maximize Refinements: Added dynamic tooltip guidance (`title="Restore"` vs `"Maximize"`) and immediate asynchronous state updating to `btn-maximize`.

### Changed

- Version bump to 1.0.259

## [1.0.258] - 2026-11-04

### Added

- Session State Controls: Added "Clear Saved Session State" action to Command Palette invoking `ClearSessionState()` with toast feedback, and toast notification on "Save Session State".

### Changed

- Version bump to 1.0.258

## [1.0.257] - 2026-11-04

### Added

- Module ESM Optimization: Added `"type": "module"` to `tabby-core/package.json` to eliminate Node.js typeless module warnings during test runs and optimize ESM loader resolution.

### Changed

- Version bump to 1.0.257

## [1.0.256] - 2026-11-04

### Added

- Dynamic System Tray Version: Updated `GetPlatform()` version response in Go backend to `1.0.256` and wired asynchronous dynamic badge updating (`#tray-version-badge`) in `toggleSystemTrayMenu`.

### Changed

- Version bump to 1.0.256

## [1.0.255] - 2026-11-04

### Added

- Server Teardown Implementation: Implemented `CloseAll()` routines in `ssh` and `telnet` packages to enumerate and cleanly disconnect active connections. Wired into `ShutdownServers()` and `QuitApplication()` in `tabby-go/wails-app/app.go`.

### Changed

- Version bump to 1.0.255

## [1.0.254] - 2026-11-04

### Added

- Build Pipeline Validation: Executed end-to-end `tabby-go/build.sh` pipeline on Linux, compiling Vite production assets and generating the cross-compiled desktop binary `tabby-go/wails-app/bin/tabby-wails.exe` (19MB).

### Changed

- Version bump to 1.0.254

## [1.0.253] - 2026-11-04

### Added

- Core Utility Unit Tests: Added `test/core/utils.test.mjs` verifying `TAB_COLORS` palette constants, `ResettableTimeout` lifecycle execution/cancellation, and `serializeFunction` FIFO asynchronous execution.

### Changed

- Version bump to 1.0.253

## [1.0.252] - 2026-11-04

### Added

- Desktop Window Title Sync: Automatically synchronize native operating system window title (`SetWindowTitle`) on tab activation and restore cleanly to "Tabby" when all tabs are closed.

### Changed

- Version bump to 1.0.252

## [1.0.251] - 2026-11-04

### Added

- Window Management Hotkeys: Integrated global keyboard shortcuts in Wails desktop app: `F11` (Toggle Fullscreen), `Ctrl+Shift+M` (Minimize Window), and `Ctrl+Shift+A` (Toggle Always on Top).
- Command Palette Window Actions: Registered Fullscreen, Minimize, and Always on Top commands with shortcut badges in the Command Palette.

### Changed

- Version bump to 1.0.251

## [1.0.250] - 2026-11-04

### Added

- Go Wails Backend Lifecycle: Implemented `ShutdownServers()` and `QuitApplication()` methods on `App` struct in `tabby-go/wails-app/app.go` ensuring background sessions are closed gracefully before application exit.
- Direct UTF8Splitter Testing: Connected `test/core/utfSplitter.test.mjs` directly to `tabby-core/src/utfSplitter.js` verifying module boundary buffering and splitting.

### Changed

- Version bump to 1.0.250
- Hoisted `alwaysOnTop` global state in Wails frontend to prevent any runtime reference ordering errors.

## [1.0.249] - 2026-11-04

### Added

- Always on Top Window Control: Integrated `toggleAlwaysOnTop` into the desktop titlebar system tray dropdown menu with real-time state feedback (Enabled/Disabled) using Wails runtime `WindowSetAlwaysOnTop`.

### Changed

- Version bump to 1.0.249

## [1.0.248] - 2026-11-04

### Added

- CI/CD Pipeline Automation: Integrated `npm test` unit tests and `tabby-go` backend test suite (`go test ./...`) into the `Lint` validation gate in `.github/workflows/build.yml` using `actions/setup-go@v5`.

### Changed

- Version bump to 1.0.248

## [1.0.247] - 2026-11-04

### Added

- Unit Testing Framework: Configured `npm test` utilizing Node.js built-in test runner for rapid, zero-dependency unit testing across core TypeScript and JavaScript utilities.
- Core Utilities Test Suite: Created `test/core/utfSplitter.test.mjs` verifying multi-byte character boundary buffering, fragmented stream handling, and buffer flushing for `UTF8Splitter`.

### Changed

- Version bump to 1.0.247

## [1.0.246] - 2026-11-04

### Added

- Automated Release Workflow: Verified and documented the tagged release workflow (`.github/workflows/release.yml`) triggering on `v*` tags with `action-automatic-releases`.

### Changed

- Version bump to 1.0.246

## [1.0.245] - 2026-11-04

### Added

- Wails Window Runtime Controls: Integrated `WindowFullscreen`, `WindowUnfullscreen`, and `WindowIsFullscreen` into system tray actions with automatic fallback to maximize/unmaximize.

### Changed

- Version bump to 1.0.245

## [1.0.244] - 2026-11-04

### Added

- Electron Version Synchronization: Explicitly defined `electronVersion: 38.0.0` in `electron-builder.yml` to ensure build consistency with `package.json` devDependency `"electron": "38"`.
- Trailing Host Wildcard SSH Fix: Handled edge cases in `ImportSSHConfigAsProfiles` ensuring preceding profiles are saved before encountering `Host *` blocks at the end of files.

### Changed

- Version bump to 1.0.244

## [1.0.243] - 2026-11-04

### Added

- Titlebar Double-Click Toggle: Added double-click event listener to the custom titlebar drag handle (`#titlebar-drag`) to dynamically toggle window maximize/restore via Wails runtime.

### Changed

- Version bump to 1.0.243

## [1.0.242] - 2026-11-04

### Added

- Unified Build Script: Added `tabby-go/build.sh` providing automated test running, frontend Vite asset bundling, and cross-platform compilation of the Wails Go binary.
- Documentation: Updated `DEPLOY.md` to document the unified build workflow for both Unix (`build.sh`) and Windows (`start.bat`).

### Changed

- Version bump to 1.0.242

## [1.0.241] - 2026-11-04

### Added

- Cloud Sync Testing: Added unit test suite `TestSyncPushAndPull` in `tabby-go/pkg/sync/sync_test.go` covering Push, Pull, and initial empty state fallback.
- Block Actions UI Documentation: Formally validated and marked Phase 2 Block Actions UI features (Copy Command, Copy Output, Filter/Search, Share, Keyboard Navigation) as complete in `TODO.md` and `ROADMAP.md`.

### Changed

- Version bump to 1.0.241

## [1.0.240] - 2026-11-04

### Added

- Enhanced OpenSSH Config Parser: Added quote-trimming on configuration values, automatic skipping of wildcard catch-all blocks (`Host *`), path normalization for `IdentityFile`, and support for `ProxyJump`.
- Unit Tests for SSH Config Importer: Added `TestImportSSHConfigAsProfiles` in `tabby-go/pkg/profile/profile_test.go`.
- Merge Strategy for SSH Import: Supported merge strategies in frontend `importSSHConfig(overwriteExisting)`: skip duplicates by default or update existing profiles with Shift-click, displaying granular counts of added, updated, and skipped profiles.

### Changed

- Version bump to 1.0.240

## [1.0.239] - 2026-11-04

### Added

- Profile Group Manager: Added dedicated Profile Groups Management panel under "Saved Profiles" in the settings dashboard.
- Profile Groups Actions: Added UI capabilities to create new groups, rename groups (automatically updating all matching profile references), and delete groups (safely re-assigning members to Ungrouped).
- Default Shell Hardening: Added fallback type verification for `this.shell` in the terminal tab constructor to prevent invocation failures on non-string inputs.

### Changed

- Version bump to 1.0.239

## [1.0.238] - 2026-11-04

### Added

- Serial Terminal Flow Control: Support for Hardware RTS/CTS, Software XON/XOFF, and None options in `tabby-go/pkg/serial` and frontend connection dialog.
- Serial Terminal Hex View Mode: Two-column hex dump with printable ASCII sidebar, supporting dynamic toggling directly via `HEX` toolbar button.
- Serial Session Logging: Ability to capture incoming/outgoing serial data to disk in timestamped `.log` files in `~/.tabby/serial-logs/` and live session log buffer viewer.
- Serial Unit Tests: Added comprehensive unit tests in `tabby-go/pkg/serial/serial_test.go` verifying mode construction, flow control handshake settings, lifecycle management, and session log streaming.

### Changed

- Version bump to 1.0.238
- Updated `tabby-go/pkg/api/types.go` with `FlowControl`, `Logging`, and `LogPath` fields for serial sessions.

## [1.0.237] - 2026-11-04

### Changed

- Version bump to 1.0.237
- Enhanced desktop custom titlebar with an integrated System Tray / Server Control menu widget
- Dynamic visibility for system tray trigger linked to user's `HideTray` preference
- Cleaned up build scripts in `tabby-go/start.bat` to automatically build frontend Vite assets before compiling Wails Go binary

### Added

- Interactive System Tray popover menu featuring:
  * Tabby Go core version and protocol status
  * New Terminal Tab (`Ctrl+Shift+T`)
  * Command Palette (`Ctrl+Shift+P`)
  * Dashboard & Settings panel (`Ctrl+,`)
  * Toggle Fullscreen & Minimize window controls
  * Safe 'Stop All Servers' with active count indicator and 'Quit Tabby'
- Exported `ShutdownServers` and `QuitApplication` methods into Wails JS bindings

## [1.0.236] - 2026-11-03

### Changed

- Version bump to 1.0.236
- Consolidated fragmented settings and dashboard pages into high-value domains: General & System, Appearance & Display, Terminal & Behavior, and Connections & Profiles
- Expanded dashboard settings panel with real-time live keyword search filtering across all configuration options

### Added

- Interactive guidance tooltip markers (ℹ) next to every configuration label in the settings dashboard
- Real-time welcome dashboard status badges (active sessions, profiles count, protocol status)
- Prominent server management actions: 'Stop All Servers' and 'Quit Tabby' with confirmation safeguards
- Go Wails backend lifecycle handlers `ShutdownServers` and `QuitApplication` for graceful connection termination

## [1.0.235] - 2026-07-13

### Changed

- Version bump to 1.0.235
- Merge upstream/master (Eugeny/tabby) with 943 upstream commits
- CI pipeline: windows-2022 runner, libsecret deps
- Tab context menu: Pin tab, command labels
- SSH: enhanced auth methods (publicKey, etc.)
- Release notes: migrated axios→fetch, marked ESM

### Added

- Terminal Broadcast mode (Ctrl+Shift+B, antenna toolbar button)
- Session Logging: capture output, view/download (📜 toolbar button)
- Profile Group dropdown suggestions in profile editor
- Settings Descriptions/Tooltips for all config options

## [1.0.231-nightly.19] - 2026-05-23

### Changed

- Version bump

## [1.0.231-nightly.14] - 2026-05-12

### Added

- Implemented a base `registerCompletionItemProvider` dictionary overlay on top of the Monaco IDE input window, providing intelligent Warp-style shell autocomplete suggestions.
- Bound `Cmd+F` and `Ctrl+F` keystrokes inside the Monaco UI block to redirect focus automatically to Tabby's global terminal output search panel, removing the isolated/redundant Monaco native find widget.

## [1.0.231-nightly.13] - 2026-05-12

### Changed

- Converted `generateCommand()` AI text generation to use an explicit overlay `prompt()` window, preventing the destruction of currently typed user command text in the Monaco IDE box (Warp parity).
- Successfully hooked up the `ReactPluginDecorator` to the Angular lifecycle, validating that simple isolated React components can inject UI frames directly on top of the Terminal canvas without relying on standard Webpack/Angular DI modules.

## [1.0.231-nightly.12] - 2026-05-12

### Changed

- Improved the Workflows Catalog Modal to actively parse template strings (e.g. `{port}`) and prompt the user to fulfill variables via a popup before pasting the finalized command string back into the Warp-style IDE Monaco editor.
- Configured a persistent rolling `tabby-backend.log` file in the user's home directory (`os.UserHomeDir`) alongside `os.Stdout` to ensure trace logs and backend panics are easily retrievable by users reporting issues from the standalone daemon.

## [1.0.231-nightly.11] - 2026-05-12

### Added

- Completed Tabby Agent Sidebar Chat integration, sending real conversation payloads containing `role` and `content` through the `ipcRenderer` directly to the `ai.agentChat` JSON-RPC orchestrator on the backend.
- Extended the `CommandCatalogModalComponent` UI to properly format and highlight `{parameter}` blocks inline matching Warp Drive workflows.

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
- Configured upstream remote (<https://github.com/Eugeny/tabby>)

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

## v1.0.231-nightly.17
- Create `tabby-go/pkg/sync` implementation representing the Warp Drive Cloud Sync backend stub.
- Connect JSON-RPC `sync.push` and `sync.pull` commands to the `server.go` IPC handler to allow the frontend to serialize Workflows and Profiles.
- Increment version tracking documents and prepare HANDOFF logic.

## v1.0.231-nightly.18
- Create `tabby-core/src/services/sync.service.ts` to act as the Cloud Sync frontend proxy.
- Expose `SyncService` through `tabby-core` index API, wired up to call `sync.push` and `sync.pull` JSON-RPC commands in the Go backend.
- Increment version tracking documents and prepare HANDOFF logic.
