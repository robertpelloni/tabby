# Tabby Terminal Vision

Tabby is an infinitely customizable, cross-platform terminal emulator designed for modern developers. It aims to completely replace native terminals (Terminal.app, conhost, gnome-terminal) and legacy emulators (PuTTY, iTerm2, MobaXterm) by offering a unified, highly extensible platform built on web technologies (Electron + Angular).

## The Ultimate Long-Term Vision (1:1 Parity with Warp, WaveTerm, and Hyper)

Tabby's ultimate goal is to evolve from a traditional, static text-stream emulator (like xterm.js) into an **Agentic, Block-based, Collaborative Workspace**. To achieve this, we are executing a massive roadmap to reach and exceed 100% feature parity with the three leading modern terminals: **Warp, WaveTerm, and Hyper**.

### 1. The Block-Based Paradigm (Warp & WaveTerm Parity)
The traditional continuous stream of text is obsolete. Tabby is transitioning to a UI where every command execution creates a distinct, addressable **Block** (inspired by Warp) that can render rich content (inspired by WaveTerm).
*   **Isolation**: Command output is bounded. You can scroll, collapse, and copy exactly the output of one command without dragging your mouse over hundreds of lines.
*   **Rich Widgets**: A block shouldn't just be text. Tabby will support rendering Markdown, images, web browsers, and Monaco code editors directly inside an output block (WaveTerm parity).
*   **Post-Processing**: Filter (grep), search, and export the contents of a single block without re-running the command.
*   **Shareability**: Generate secure web links to share the exact input/output of a block with a colleague for debugging.

### 2. IDE-Like Text Editing (Warp Parity)
The terminal input should feel like VS Code, not `ed`.
*   **Pinned Input Box**: A dedicated composing area that is separate from the scrolling output.
*   **Multi-Cursor & Selection**: Full mouse support, `Alt+Click` for multiple cursors, and IDE-standard keybindings (`Cmd+D`).
*   **Intelligent Autocompletion**: Context-aware, fuzzy-searchable dropdowns for commands, flags, and file paths with inline documentation.
*   **Syntax Highlighting & Validation**: Real-time syntax highlighting of shell scripts and visual error indicators (red squiggles) before hitting Enter.

### 3. Agentic AI Integration (Warp Parity)
Tabby will embed a first-class AI assistant directly into the workflow.
*   **Natural Language to Command**: Type "how to extract a tar.gz file" and the AI places the exact command (`tar -xvzf`) into your input box.
*   **Explain Error**: A dedicated action on failed blocks. The AI reads the command, the stderr output, and your environment to diagnose the issue and suggest a fix.
*   **Terminal Agent**: A chat interface that can autonomously execute diagnostic commands (with permission) to resolve complex server issues.
*   **Workflow Generation**: AI-assisted creation of parameterized shell scripts for common tasks.

### 4. Workflows & Collaboration (Warp Parity)
*   **Command Catalog**: Save, parameterize, and search your most-used commands (e.g., `docker build -t {{image_name}} .`).
*   **Team Sync**: A cloud backend (like Warp Drive) to securely share Workflows, Environment Variables, and SSH Profiles with your entire team.

### 5. Infinite Extensibility (Hyper Parity)
*   **Hot Reloading Configuration**: A unified, dead-simple JSON/JS configuration file that reloads the UI instantly upon save.
*   **Plugin Ecosystem**: A frictionless API to build, share, and install plugins via `npm`, dramatically expanding the UI capabilities (e.g., aesthetics, themes, custom widgets).

## The Architecture Vision

To support this massive leap in functionality without compromising performance or stability, Tabby is undergoing a fundamental architectural shift:

*   **The Go Backend (`tabby-go`)**: We are aggressively porting all OS-level, native, and complex protocol logic (PTY spawning, Hardware Serial, SSH, SFTP, Multiplexing) out of brittle Node.js native modules (`node-gyp`) and into a unified, statically compiled Go daemon.
    *   **JSON-RPC 2.0 Bridge**: The Electron main process communicates with the Go daemon via a robust JSON-RPC bridge over `stdin`/`stdout`.
    *   **Universal Frontends**: The Go backend is designed to be entirely decoupled from Electron. It will eventually serve as the engine for a pure Native UI frontend, a Web UI (browser-based remote terminal), and mobile companion apps.
    *   **Remote Orchestration**: `tabby-go` will eventually act like WaveTerm's `wsh`, allowing the backend to coordinate remote file editing, metrics, and complex SSH routing over a single multiplexed channel.
*   **Frontend Refactoring**: The Angular frontend services are being stripped of heavy logic and transformed into thin IPC proxy clients. The UI concerns itself *only* with rendering blocks, IDE text boxes, and handling user input.

This dual approach—a hyper-modern, block-based, AI-driven UI powered by a bulletproof, natively compiled Go daemon—is the definitive vision for Tabby.
