# Tabby - Long-Term Structural Roadmap

## Phase 1: The Go Backend Migration (In Progress)
**Priority: CRITICAL** | **Status: MATURE**

### Goal
Eliminate brittle Node.js native dependencies (`node-pty`, `serialport`, `russh`) by porting all OS-level and protocol logic to a unified Go daemon (`tabby-go`).

### Current Status
*   **PTY**: Implemented in Go (`creack/pty`). Electron bridge handles spawn, resize, write, and async data streams. ✅
*   **Serial**: Implemented in Go (`go.bug.st/serial`). Electron bridge handles port listing, connect, write, and async data streams. ✅
*   **Communication**: JSON-RPC 2.0 bridge over stdin/stdout. Tests are written and passing. ✅
*   **SSH & SFTP Backend**: `SSHSession` and `SFTPSession` Angular services rewritten to proxy all calls to the Go daemon via `ipcRenderer`. ✅
*   **SFTP Performance**: Path-based transfers implemented to bypass base64 IPC overhead. ✅
*   **SFTP UI**: Progress indicators, Rename, and Directory creation fully wired. ✅
*   **Sync Backend**: `SyncService` integrated with Go backend for cloud storage. ✅

### Immediate Next Steps (Go Parity)
1.  **Jump Hosts (ProxyJump)**: Add support for recursive TCP Dialing via SSH in the Go backend (`pkg/ssh/ssh.go`).
2.  **Native PTY Polish**: Integration of Windows ConPTY for high-fidelity terminal emulation.

## Phase 2: The UI/UX Paradigm Shift (Warp & WaveTerm Parity)
**Priority: HIGH** | **Status: EXPERIMENTAL**

### Goal
Completely overhaul the traditional continuous text stream (`xterm.js`) into an **Agentic, Block-based, Collaborative Workspace**.

### Current Status
*   **BlockFrontend Scaffold**: Created foundational implementation in `tabby-terminal/src/frontends/blockFrontend.ts`. ✅

### Key Milestones
1.  **Block-Based Output Parsing**: Intercept PTY data stream, parse ANSI into styled DOM blocks.
2.  **IDE-Like Text Editing**: Replace traditional prompt with pinned `monaco-editor` input box.
3.  **Rich Widget Blocks**: Detect OSC codes to render Markdown/Images inline.

[Rest of Phase 3-7 omitted for brevity, identical to previous version but updated status where applicable]
