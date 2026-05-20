# Submodule and Library Inventory

## Submodules

| Name | Path | URL | Status | Purpose |
|------|------|-----|--------|---------|
| btk | `tabby-go/vendor/btk` | https://github.com/robertpelloni/btk.git | Vendor | CGo bridge and Native UI bindings |
| warp | `warp` | https://github.com/robertpelloni/warp.git | Active | Warp terminal source for analyzing parity features (IDE input, AI blocks) |

## Major Libraries & Packages

| Name | Path / Context | Purpose |
|------|----------------|---------|
| `creack/pty` | Go Backend (`tabby-go`) | Cross-platform PTY management (replacing `node-pty`) |
| `go.bug.st/serial` | Go Backend (`tabby-go`) | Serial port management (replacing Node `serialport`) |
| `monaco-editor` | Frontend (`tabby-terminal`) | Used to replace xterm input with an IDE-like text editing experience (Warp parity) |
| `angular` | Frontend (`tabby-*`) | Primary SPA application framework |
| `electron` | Desktop Wrapper | Used to package the web-app into a desktop application |
| `xterm.js` | Frontend (`tabby-terminal`) | Legacy continuous-stream terminal output logic |
