# Project Memory & Knowledge Base

- Tabby is transitioning all core backend communication to a Go JSON-RPC 2.0 daemon (`tabby-go/cmd/tabby-backend`).
- The Electron frontend spawns the Go daemon via `GoBackendService` (`app/lib/goBackend.ts`).
- **PTY**: Replaced `node-pty`. The frontend uses synchronous UUID generation mapped to the async Go Backend `pty.spawn` request to preserve the frontend IPC architecture.
- **Serial**: Replaced `@serialport`. The frontend utilizes `ipcRenderer.invoke` via `HostAppService.ipcInvoke` to communicate with the Go backend, retaining Web Serial API fallback for browser targets.
- **SSH/SFTP**: Next major target for Go integration. Requires extensive rewriting of `SSHSession` state management.
- Always use `yarn install` and `yarn build` for the frontend. For the backend, use `go build` and `go mod vendor` to maintain dependencies.
