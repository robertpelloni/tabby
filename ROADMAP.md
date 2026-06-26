# Tabby - Long-Term Structural Roadmap

## Phase 1: The Go Backend Migration (In Progress)
**Priority: CRITICAL** | **Status: ALMOST COMPLETE**

### Goal
Eliminate brittle Node.js native dependencies (`node-pty`, `serialport`, `russh`) by porting all OS-level and protocol logic to a unified Go daemon (`tabby-go`). Achieve 1:1 parity with the legacy TypeScript implementation.

### Current Status
*   **PTY**: Implemented in Go (`creack/pty`). Electron bridge handles spawn, resize, write, and async data streams. ✅
*   **Serial**: Implemented in Go (`go.bug.st/serial`). Electron bridge handles port listing, connect, write, and async data streams. ✅
*   **Communication**: JSON-RPC 2.0 bridge over stdin/stdout. Tests are written and passing. ✅
*   **SSH & SFTP Backend**: `SSHSession` and `SFTPSession` Angular services rewritten to proxy all calls to the Go daemon via `ipcRenderer`. The Electron main process correctly multiplexes these requests. ✅
*   **X11 Forwarding**: Implemented missing `x11-req` socket forwarding in the Go backend (`pkg/ssh/ssh.go`). ✅
*   **Jump Hosts (ProxyJump)**: Implemented recursive TCP Dialing via SSH natively in the Go backend (`pkg/ssh/ssh.go`). Verified via unit tests. ✅
*   **End-to-End Integration Testing**: Validated `tabby-go` multiplexing, raw `Uint8Array` data propagation over RPC to the frontend. Node `russh` module completely removed. ✅

### Immediate Next Steps (Phase 1)
1. **Stabilize Production CI**: Resolve any underlying CI script or packaging dependencies missed during the Node native modules swap out before cutting a stable release.

## Phase 2: The UI/UX Paradigm Shift (Warp & WaveTerm Parity)
**Priority: HIGH** | **Status: EXPERIMENTAL**

### Goal
Completely overhaul the traditional continuous text stream (`xterm.js`) into an **Agentic, Block-based, Collaborative Workspace** with IDE-like text editing and Rich Widget output. Achieve and exceed 1:1 feature parity with Warp Terminal and WaveTerm.

### Current Status
*   **BlockFrontend Scaffold**: Created the foundational `BlockFrontend` implementation in `tabby-terminal/src/frontends/blockFrontend.ts` to begin experimenting with DOM-based rendering over canvas-based rendering. ✅

### Key Milestones
1.  **Block-Based Output Parsing**: Intercept the PTY data stream, parse ANSI escape sequences into styled HTML, parse shell prompts, and render each command execution as a distinct, isolated DOM element ("Block").
2.  **IDE-Like Text Editing**: Replace the traditional command prompt with a pinned `contenteditable` (or Monaco Editor) input box at the bottom of the screen. Implement:
    *   Mouse click to place cursor.
    *   Multi-cursor editing (`Alt+Click`, `Cmd+D`).
    *   Intelligent fuzzy-search Tab completion.
    *   Real-time syntax highlighting and error validation (red squiggles).
3.  **Block Actions**: Implement UI for:
    *   Copying command/output of a specific block.
    *   Filtering/searching within a block.
    *   Generating shareable web links for a block.
    *   Keyboard navigation between blocks.
4.  **Rich Widget Blocks (WaveTerm Parity)**: Allow the output block to detect special metadata tags (via ANSI OSC codes or JSON payloads from the `tabby-go` daemon) to render Markdown, images, webviews, and code editors directly in the terminal block flow.

## Phase 3: Agentic AI Integration (Warp Parity)
**Priority: HIGH** | **Status: PLANNING**

### Goal
Embed a first-class AI assistant directly into the workflow to diagnose errors, generate commands, and automate tasks.

### Key Milestones
1.  **AI Command Search**: Natural language to shell command generation within the IDE input box.
2.  **Explain Error Action**: A one-click button on failed blocks that reads the command, stderr, and environment to suggest a fix.
3.  **Terminal Agent Chat**: A dedicated sidebar/panel for conversational interaction with an AI that can autonomously run diagnostic commands (with user permission).
4.  **Workflow Generation**: AI-assisted creation of parameterized, saved shell scripts.

## Phase 4: Workflows & Collaboration (Warp Drive Parity)
**Priority: MEDIUM** | **Status: PLANNING**

### Goal
Provide a platform for teams to save, parameterize, search, and synchronize their terminal environment.

### Key Milestones
1.  **Command Catalog**: A searchable UI (Command Palette style) for saved, parameterized commands (e.g., `docker build -t {{image_name}} .`).
2.  **Cloud Sync Backend**: A secure backend service to synchronize Workflows, Environment Variables, and SSH Profiles across a user's devices.
3.  **Team Sharing**: Permissions and organizational structure to share workflows with colleagues.

## Phase 5: Infinite Extensibility (Hyper Parity)
**Priority: MEDIUM** | **Status: PLANNING**

### Goal
Ensure the Tabby plugin ecosystem is as rich, accessible, and fast-reloading as Vercel's Hyper terminal.

### Key Milestones
1.  **Hot Reloading Configuration**: Watch `~/.tabby.yaml` or `config.json` for file changes and instantly update the Angular UI (themes, shortcuts, layouts) without a full app refresh.
2.  **Simplified Plugin API**: Abstract away Angular Dependency Injection where possible so users can wrap existing UI components with raw React or Web Components simply by pushing a JS file to a local `~/.tabby-plugins/` directory.

## Phase 6: Feature Completeness & Polish
**Priority: MEDIUM** | **Status: PARTIAL**

### Goal
Ensure every legacy feature is fully wired up, tested, and comprehensively represented in the UI before replacing `xterm.js`.

### Key Milestones
1.  **SFTP File Manager UI**: Full bidirectional file browser, drag-and-drop file transfer, progress indicators, directory tree browsing.
2.  **SSH Features**: Port forwarding management UI, Jump host chain management UI, SSH config file import polish.
3.  **Terminal Features**: Split pane management improvements, custom shell profile configuration polish, terminal broadcast input (send to all tabs).
4.  **Settings UI**: All config options represented with proper descriptions/tooltips, import/export configuration, profile group management.
5.  **Serial Terminal**: Hex view mode, advanced flow control settings, connection logging.

## Phase 7: Universal Frontends
**Priority: LOW** | **Status: PLANNING**

### Goal
Leverage the decoupled Go backend (`tabby-go`) to power multiple native frontends.

### Key Milestones
1.  **Pure Native UI**: A blazing fast, GPU-accelerated desktop client built on a modern framework (e.g., Tauri, Rust/Metal, or a Go native UI).
2.  **Web UI**: A browser-based remote terminal client connecting to a headless `tabby-go` daemon running on a server.
3.  **Mobile Companion**: iOS/Android apps for remote session management and quick command execution.
