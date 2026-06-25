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
2.  **IDE-Like Text Editing**: Replace the traditional command prompt with a pinned `contenteditable` (or Monaco Editor) input box at the bottom of the screen.
3.  **Block Actions**: Implement UI for copying, filtering, and navigating blocks.
4.  **Rich Widget Blocks (WaveTerm Parity)**: Allow the output block to detect special metadata tags (via ANSI OSC codes or JSON payloads from the `tabby-go` daemon) to render Markdown, images, webviews, and code editors directly in the terminal block flow.

## Phase 3: Agentic AI Integration (Warp Parity)
**Priority: HIGH** | **Status: PLANNING**

### Goal
Embed a first-class AI assistant directly into the workflow to diagnose errors, generate commands, and automate tasks.

## Phase 4: Workflows & Collaboration (Warp Drive Parity)
**Priority: MEDIUM** | **Status: PLANNING**

### Goal
Provide a platform for teams to save, parameterize, search, and synchronize their terminal environment.

## Phase 5: Infinite Extensibility (Hyper Parity)
**Priority: MEDIUM** | **Status: PLANNING**

## Phase 6: Feature Completeness & Polish
**Priority: MEDIUM** | **Status: PARTIAL**

## Phase 7: Universal Frontends
**Priority: LOW** | **Status: PLANNING**

### Goal
Leverage the decoupled Go backend (`tabby-go`) to power multiple native frontends.
