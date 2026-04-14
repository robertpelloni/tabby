### Tabby Architecture Overview

**1. General Structure & Paradigm**
Tabby (formerly Terminus) is a highly extensible, cross-platform terminal emulator. It is built fundamentally on two layers:
- **Frontend/Shell:** An Electron application featuring an Angular 15 frontend. The frontend is heavily modular and relies entirely on a plugin-based architecture.
- **Backend/Native Layer:** Historically relying on Node.js native modules (like `node-pty`, `serialport`, and `russh` via N-API), the project is currently in the process of porting these performance-critical, low-level system operations to a **Go backend**. 

**2. The Go Backend (`tabby-go`)**
The `tabby-go` subproject acts as a standalone binary spawned by the Electron main process. 
- **Communication:** It communicates with the frontend via **JSON-RPC 2.0** over standard input and standard output (`stdin`/`stdout`). This avoids IPC overhead native to Electron and guarantees cross-platform native execution.
- **Components:**
  - **SSH/SFTP (`pkg/ssh`, `pkg/sftp`):** Full SSH2 client connection manager, supporting port forwarding, agent forwarding, X11, interactive auth, and an SFTP file manager.
  - **PTY (`pkg/pty`):** Pseudo-terminal management using `creack/pty` on Unix-like systems. This replaces Node's `node-pty`. It manages the lifecycle, resizing, and output forwarding of local shell processes.
  - **Serial (`pkg/serial`):** Hardware serial port communication using `go.bug.st/serial`. Supports baud rate configurations, parities, and real-time streaming of serial data.
  - **Telnet (`pkg/telnet`):** Basic telnet connectivity and input/output routing.
  - **Native UI (`pkg/ui`):** Experimental bindings via CGo to `BTK` (a Qt-descended C++ UI framework), hinting at a potential future native, non-Electron GUI version.
- **Event Loop & Notifications:** The backend implements an asynchronous event notification loop. When data is received from PTY or Serial, it base64-encodes the bytes and emits a JSON-RPC notification (e.g., `"pty.data"` or `"serial.data"`) back to the Electron parent process.

**3. Frontend Plugin Architecture**
Everything in Tabby's UI is a plugin. Core functionalities are bundled as default plugins (e.g., `tabby-core`, `tabby-terminal`, `tabby-ssh`, `tabby-serial`, `tabby-settings`).
- **Angular Dependency Injection:** Plugins register themselves by providing classes to Angular DI tokens (Extension Points). For example, a new profile type is added by providing a `ProfileProvider`, and a new toolbar button is added via `ToolbarButtonProvider`.
- **xterm.js Integration:** The actual terminal emulation is handled by `xterm.js`, wrapped inside an Angular component in `tabby-terminal`.

**4. Design Decisions & Patterns**
- **Security & Vault:** Tabby features a built-in encrypted vault (`tabby-core/src/services/vault.ts`) for storing sensitive credentials like SSH passwords and keys, avoiding plain-text config storage.
- **Single Binary / Multi-role:** The Go backend serves as a versatile daemon handling multiple connection types asynchronously using Go's strong concurrency features (goroutines and channels).
- **Graceful Degradation:** The PTY and Serial packages are designed to fall back or return errors cleanly if the host OS does not support certain features (e.g., using `exec.Cmd` pipes if full PTY isn't available, though we just upgraded it to use `creack/pty`).

**5. Immediate Roadmap & Refactoring Strategy**
- The primary goal is to fully port all local PTY, Serial, and SSH session management from the Node/Electron native layer to the `tabby-go` JSON-RPC backend. 
- Ensure that the TypeScript frontend cleanly routes these connections to the Go backend instead of `node-pty` and `serialport`.
- Expand documentation (which I have started by creating/updating `ROADMAP.md`, `VISION.md`, `CHANGELOG.md`, `DEPLOY.md`, etc.) to maintain a clear path forward for future AI/human maintainers.