# Tabby Project TODO

## High Priority
- [ ] **Refactor `tabby-ssh`**: Completely remove `russh` dependency. Rewrite `SSHSession` class to use asynchronous `ipcRenderer.invoke('ssh:connect', params)` mapped to the GoBackendService and handle interactive auth prompts via IPC event listeners.
- [ ] **Refactor `tabby-sftp`**: Rewrite `SFTPSession` to proxy file operations via `ipcRenderer.invoke('sftp:...')`.

## Medium Priority
- [ ] **Automated Tests**: Implement complete test suite for the Go JSON-RPC layer.
- [ ] **CI/CD**: Fix git tags and `yarn build` errors linked to missing versions.

## Completed ✓
- [x] Implement native Go PTY backend
- [x] Wire up `tabby-local` to route PTY requests via Go IPC
- [x] Implement native Go Serial backend
- [x] Wire up `tabby-serial` to route Serial requests via Go IPC
- [x] Implement GoBackendService in Electron main process
