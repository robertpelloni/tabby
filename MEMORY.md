# Memory - Ongoing Observations About the Codebase

## Architecture Observations

### The Go Migration (Rebuild Phase)
- Tabby is transitioning from a Node.js/Rust backend to a native **Go backend daemon** (`tabby-go`).
- The Go backend communicates with the Electron frontend via **JSON-RPC 2.0** over stdin/stdout.
- Core protocols (PTY, SSH, SFTP, Serial, Telnet) are now handled by Go packages in `tabby-go/pkg`.
- Frontend services (e.g., `SSHSession`, `SFTPSession`, `AgentService`) act as thin IPC proxies routing logic to Go.

### Agentic Workflow Engine (Phase 8)
- Implemented a **7-phase stateful workflow engine** in `tabby-go/pkg/agent` (Discovery, Exploration, Clarification, Design, Implementation, Review, Summary).
- Supports **Blocking Clarification**: The workflow pauses and waits for user input via `agent.submitWorkflowResponse`.
- **VDOM Protocol**: A Virtual DOM system in Go (`pkg/vdom`) allows agents to push interactive UI trees to the frontend.
- **Context Management**: `ContextManager` tracks codebase symbols and file access history to provide agents with persistent memory.

### Frontend Integration
- **Widget System**: `AgentWidgetsComponent` and `WidgetVDOMComponent` in `tabby-core` render agent-driven rich media (Markdown, Dashboards).
- **Workflow Visualization**: `WorkflowProgressComponent` provides real-time feedback on the agent's reasoning process.
- **AOT Compatibility**: The project strictly adheres to Angular 15 AOT compilation. `AppModule` must maintain its class-based decorator structure for dynamic plugin loading.

### Build & Environment
- **Heap Limits**: Webpack builds for large packages (e.g., `tabby-terminal`) require `NODE_OPTIONS="--max-old-space-size=8192"`.
- **TSLib Versioning**: Due to multi-package hoisting issues, `tslib` (v2.5+) must be manually symlinked from the root `node_modules` into each sub-package's `node_modules` to ensure `__spreadArray` availability.
- **BrowserWindow instantiation**: In Electron environments, `BrowserWindow` must be cast to `any` or required directly to avoid TypeScript naming conflicts.

## Testing & Verification
- **Go Test Suite**: Core backend logic is verified via `go test ./...`.
- **Integration Benchmark**: `integration_test.py` validates the full stack, ensuring <1.1ms RPC latency and stable agent/widget lifecycles.

## Design Preferences
- Single source of truth for versions in `VERSION.md`.
- Comprehensive documentation in `VISION.md`, `ROADMAP.md`, and individual tool analysis files.
- Automated submodule analysis and porting workflow.
