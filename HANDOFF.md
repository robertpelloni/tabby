# Session Handoff — Window Controls Hotkeys & Wails Runtime Integration (v1.0.251)

## Completed Operations

### 1. Window Control Keybindings (`tabby-go/wails-app/frontend/src/main.js`)
- Added global keyboard shortcuts to `bindGlobalKeys()`:
  * `F11`: Toggle Fullscreen via Wails runtime (`WindowIsFullscreen`, `WindowFullscreen`, `WindowUnfullscreen`).
  * `Ctrl+Shift+M`: Minimize window (`WindowMinimise`).
  * `Ctrl+Shift+A`: Toggle Always on Top (`toggleAlwaysOnTop`).
- Registered window control actions with shortcut badges inside the Command Palette.
- Re-built production Vite frontend assets into `tabby-go/wails-app/frontend/dist`.

### 2. Versioning & Documentation
- Bumped `VERSION.md` to `1.0.251`.
- Updated `CHANGELOG.md` and `HANDOFF.md`.
