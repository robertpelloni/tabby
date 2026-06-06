# WaveTerm Analysis & Porting Plan

## Overview
WaveTerm introduces the concept of persistent, rich-media "Blocks" that go beyond standard terminal text. Each block is governed by a `BlockController` in the backend and a `View` in the frontend.

## Core Components

### 1. Block Lifecycle & Controllers (`pkg/blockcontroller`)
- **ShellController**: Manages a standard shell session within a block.
- **TsunamiController**: Part of WaveTerm's custom UI engine for high-performance rendering.
- **VDOMController**: Manages dynamic UI components described via a JSON-based Virtual DOM protocol.

### 2. View Types (`frontend/app/view`)
- `term`: Xterm.js based terminal.
- `preview`: Rich preview for Markdown, JSON, CSV, and images.
- `web`: Iframe-based or Electron-webview based browser.
- `sysinfo`: Real-time system monitoring.
- `launcher`: Quick access to commands and connections.
- `vdom`: Renders the Virtual DOM trees sent by the backend.

### 3. Widget Configuration (`schema/widgets.json`)
- Defines metadata like icons, colors, labels, and the associated `BlockDef`.
- `BlockDef` contains `view` type, `controller` type, and specific settings (`term:fontsize`, `cmd:cwd`, etc.).

## Implementation Strategy for Tabby

### Backend (Go)
1. **Extend `tabby-go/pkg/blocks`**:
    - Add a `WidgetType` to the `Block` struct.
    - Implement a `WidgetManager` to handle specialized block types.
2. **VDOM Protocol**:
    - Port the VDOM JSON structures from `pkg/vdom`.
    - Implement an `AgentWidget` that allows LLM agents to push VDOM updates via JSON-RPC.
3. **JSON-RPC Extensions**:
    - `agent.createWidget`: Creates a specialized block (e.g., a "sysinfo" or "vdom" block).
    - `agent.updateWidget`: Sends data/updates to an existing widget.

### Frontend (Angular)
1. **Block Component Factory**:
    - Update the terminal UI to render specialized components based on the block's `WidgetType`.
    - Create `WidgetPreviewComponent`, `WidgetWebComponent`, and `WidgetVDOMComponent`.

## Porting VDOM (The "LLM Harness" Killer Feature)
WaveTerm's VDOM allows the backend to describe a UI tree:
```json
{
  "tag": "div",
  "props": { "className": "p-4" },
  "children": [
    { "tag": "h1", "children": ["System Diagnostic"] },
    { "tag": "button", "props": { "onClick": "run_fix" }, "children": ["Fix Issues"] }
  ]
}
```
This is perfect for an AI agent to generate interactive dashboards or custom tools.

## Status
- **Analysis**: Complete.
- **Implementation**: Backend VDOM and Agent Widget creation implemented in Go. Frontend rendering implemented in `WidgetVDOMComponent`.
- **Redundancy**: Submodule `waveterm` is now redundant and has been removed.
