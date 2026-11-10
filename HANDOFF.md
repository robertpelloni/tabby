# Session Handoff — Wails Lifecycle Server Teardown Hook (v1.0.280)

## Completed Operations

### 1. Window Close Lifecycle Teardown (`tabby-go/wails-app/app.go`)
- Wired `a.ShutdownServers()` inside `onBeforeClose(ctx context.Context)` in `App`, ensuring that native OS window closures (e.g. Alt+F4, window manager close events) trigger graceful shutdown across SSH and Telnet connections.
- Re-built Vite production assets in `tabby-go/wails-app/frontend/dist`.

### 2. Versioning & Documentation
- Bumped `VERSION.md` to `1.0.280`.
- Synchronized `app.go` and system tray badge to `1.0.280`.
- Updated `CHANGELOG.md` and `HANDOFF.md`.

## Test Status
- `npm test`: 11/11 tests pass.
- `cd tabby-go && env CGO_ENABLED=1 go test ./...`: All 30 test packages pass.
