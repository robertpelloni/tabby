# Tabby TypeScript vs. Go Backend Parity Analysis

The user wants 1:1 parity between the old TypeScript implementation and the new Go backend.

## PTY (Local Terminal)
### TypeScript (`node-pty`) Features:
*   Spawn process (shell, env vars, cwd).
*   Resize columns/rows.
*   Write stdin.
*   Listen to stdout (`data`).
*   Listen to process exit (`exit`).

### Go Backend (`creack/pty`) Features:
*   Spawn process (shell, env vars, cwd) ✅
*   Resize columns/rows ✅
*   Write stdin ✅
*   Listen to stdout (`PTY.Data` notification) ✅
*   Listen to process exit (`PTY.Exit` notification) ✅
*   **Parity Status: 100% COMPLETE**

## Serial Communication
### TypeScript (`serialport`) Features:
*   List available COM ports.
*   Open connection (baud rate, data bits, stop bits, parity, flow control).
*   Write bytes.
*   Read bytes.
*   Close connection.

### Go Backend (`go.bug.st/serial`) Features:
*   List COM ports (`Serial.ListPorts`) ✅
*   Open connection (`Serial.Connect`) ✅ (Supports all config options like flow control, baud, etc.)
*   Write bytes (`Serial.Write`) ✅
*   Read bytes (`Serial.Data` notification) ✅
*   Close connection (`Serial.Close`) ✅
*   **Parity Status: 100% COMPLETE**

## SSH & SFTP
### TypeScript (`russh`) Features:
*   Authentication (Password, Keyboard-Interactive, Agent, Private Key).
*   Host key verification (interactive prompt, known_hosts).
*   Channel multiplexing (Shell, SFTP, TCP Forwarding).
*   Port Forwarding (Local, Remote, Dynamic).
*   Keepalive intervals.
*   SFTP File Management (Download, Upload, Stat, Readdir, Chmod, Mkdir, etc.).
*   X11 Forwarding.

### Go Backend (`golang.org/x/crypto/ssh`) Features:
*   Authentication ✅
*   Host key verification ✅
*   Channel multiplexing (Shell, SFTP) ✅
*   Port Forwarding (Local, Remote, Dynamic) ✅
*   Keepalive intervals ✅
*   SFTP File Management (via `pkg/sftp`) ✅
*   **Missing Features**:
    *   **X11 Forwarding**: The Go backend does not currently have a dedicated multiplexer for X11 sockets (`ssh:startShell` takes `x11` as a boolean, but the Go implementation doesn't route the `x11` ssh request type back to a local socket server yet).
    *   **Jump Hosts (ProxyJump)**: The TypeScript frontend handles this by opening an `AuthenticatedSSHClient` and asking it for a TCP Forwarding channel to use as the transport for the next SSH client. The Go backend (`pkg/ssh/ssh.go`) currently assumes a direct TCP dial.
*   **Parity Status: ~85% COMPLETE**. (Requires X11 socket forwarding implementation and recursive jump host dialing in Go).

# Tabby vs. Warp Parity Analysis

Tabby currently operates as a traditional terminal emulator (xterm.js) wrapped in an Electron GUI with tabs and split panes.

### The Delta (What Tabby Needs for 1:1 Warp Parity)
1.  **UI Paradigm Shift**: Tabby needs to move away from `xterm.js` rendering a continuous block of text. It needs a dedicated, pinned text area at the bottom built with standard HTML/CSS `contenteditable` or Monaco Editor to provide multi-cursor, syntax-highlighted IDE editing.
2.  **Output Parsing (Blocks)**: Tabby needs to intercept the PTY output, parse shell prompts, and visually separate the output of `Command A` from `Command B` into distinct DOM elements (Blocks).
3.  **Command Interception**: Tabby must intercept the command *before* sending it to the PTY, enabling it to render the command in a "Block Header" rather than letting the shell echo the command.
4.  **AI Integration Engine**: Tabby requires a dedicated backend service (or LLM API keys configuration) to build the "Agentic" chat panel, "Explain Error" button on failed blocks, and natural language command generation.
5.  **Workflows Catalog**: A UI panel (Command Palette style) to search, categorize, and parameterize saved shell scripts.
6.  **Collaboration Cloud**: A backend sync engine to share these workflows and blocks with a team.
