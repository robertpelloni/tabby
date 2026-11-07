# Session Handoff — Backend Server Teardown Implementation (v1.0.255)

## Completed Operations

### 1. Connection Teardown Logic (`tabby-go/pkg/ssh/close_all.go`, `tabby-go/pkg/telnet/close_all.go`)
- Added `CloseAll()` to `ssh.Manager` and `telnet.Manager` allowing thread-safe enumeration and termination of all active background client sessions.
- Wired `a.sshMgr.CloseAll()` and `a.telnetMgr.CloseAll()` inside `ShutdownServers()` in `tabby-go/wails-app/app.go`.

### 2. Versioning & Documentation
- Bumped `VERSION.md` to `1.0.255`.
- Updated `CHANGELOG.md` and `HANDOFF.md`.
