# HANDOFF.md

## Current Status (v1.0.231-nightly.4)
The mission to port the Tabby backend native dependencies to Go has reached a major milestone.

*   **Accomplished**:
    *   Integrated local PTYs (`creack/pty`) and Hardware Serial (`go.bug.st/serial`) into the Go JSON-RPC bridge.
    *   **SSH & SFTP Rewrite**: We successfully removed the massive (~1,000 line) `russh` dependency from the `tabby-ssh` frontend. The `SSHSession` and `SFTPSession` Angular services have been entirely rewritten to act as thin proxy clients. They now route connection parameters, SFTP file handle calls, and terminal stream data through asynchronous `ipcRenderer.invoke()` channels over to the Electron main process.
    *   The Electron main process (`app/lib/goBackend.ts`) now correctly registers all `ipcMain.handle` endpoints (e.g., `ssh:connect`, `ssh:resize`, `sftp:readDir`) and multiplexes them to the `tabby-go` daemon.
    *   Deleted obsolete/failing CGo code (`tabby-go/pkg/ui`) that broke testing.

*   **Pending Verification & E2E Testing**:
    *   We need to actually boot the desktop application (`yarn start`) and verify that typing in an SSH terminal forwards keystrokes to the Go daemon properly.
    *   SFTP UI (Drag and drop) testing needs to be mapped to the new asynchronous stream handling.

## Next Steps for the Next LLM Session
1.  **Run an E2E Test**: Attempt to spawn an SSH connection using the new architecture. Pay close attention to how `MockChannel` handles raw `Uint8Array` data vs. Base64 encoded JSON-RPC frames.
2.  **Phase 2: Feature Completeness**: As defined in `ROADMAP.md`, focus on the *SFTP File Manager UI*. The core bindings are now built into Go and the Electron bridge, but the frontend context menu only supports "Download". We need to build the full Drag-and-drop React/Angular interface for SFTP directory tree browsing.
3.  Continue documenting everything meticulously per the user's vision.
