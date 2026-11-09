# Session Handoff — Last Tab Closes Window Behavior (v1.0.270)

## Completed Operations

### 1. Wails Window Lifecycle & Settings Integration (`tabby-go/wails-app/frontend/src/main.js`)
- Wired `settings.LastTabClosesWindow` in `Tab.close()`: when the last remaining active tab is closed and `LastTabClosesWindow` is enabled, `quitTabbyApplication()` is called to cleanly shut down background connection managers (SSH/Telnet `CloseAll`) and exit the application.
- Re-built Vite production assets in `tabby-go/wails-app/frontend/dist`.

### 2. Versioning & Documentation
- Bumped `VERSION.md` to `1.0.270`.
- Updated `CHANGELOG.md` and `HANDOFF.md`.

## Test Status
- `npm test`: 11/11 tests pass.
- `cd tabby-go && env CGO_ENABLED=1 go test ./...`: All 28 test packages pass.
