# Tabby Roadmap

## Current Phase: Go Backend Porting & Integration

### Go Backend Integration (In Progress)
- [x] Implement native Go PTY and Serial backends
- [x] Wire up `tabby-local` to route PTY requests via Go IPC
- [x] Wire up `tabby-serial` to route Serial requests via Go IPC
- [ ] **Crucial Next Step**: Wire up `tabby-ssh` to route SSH/SFTP requests via Go IPC.
      *Note*: This requires a comprehensive rewrite of `tabby-ssh/src/session/ssh.ts` and `sftp.ts` to remove the `russh` dependency and replace it with asynchronous `ipcRenderer.invoke` calls mapped to the `GoBackendService`.

## Future Phases
- [ ] Native BTK UI Front-end implementations.
- [ ] Mobile companion applications.
