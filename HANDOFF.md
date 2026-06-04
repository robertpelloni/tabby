# Handoff Document

## Current Status
The project is currently at version **v1.0.231-nightly.19**.

## Recent Work
*   **Agent Driver (Go Backend):** Implemented the core `agent` package in `tabby-go` for autonomous task orchestration. This provides task status tracking, progress indicators, and async execution within the Go daemon.
*   **Agent RPC Integration:** Registered `agent.runTask`, `agent.listTasks`, and `agent.getTaskStatus` JSON-RPC 2.0 methods in the Go server, enabling the Electron frontend to drive backend-level automation.
*   **Block Output Parsing:** Implemented heuristic shell prompt detection in `BlockFrontend` to automatically partition terminal output into discrete, actionable blocks.
*   **Agent Management UI:** Integrated `AgentService` and `AgentStatusComponent` into the frontend, providing a real-time status tracker and progress indicator for autonomous backend tasks in the main toolbar.
*   **Block Actions:** Enhanced `BlockFrontend` with a comprehensive set of block-level actions: copy command, copy output, integrated search/filtering, and shareable link generation. Implemented `Ctrl+Arrow` navigation for cycling through blocks.
*   **Command Catalog:** Built the `CommandCatalogModalComponent` UI to allow users to search and apply parameterized workflows synced from the cloud backend.
*   **Warp Drive Parity (Cloud Sync):** Built out the Angular frontend service `SyncService` in `tabby-core`. This service exposes `push()` and `pull()` commands which forward payloads across the `ipcRenderer` border to the `tabby-go` daemon's `sync.push` and `sync.pull` JSON-RPC targets.
*   **API Export:** Exposed the `SyncService` cleanly via `tabby-core/src/index.ts` so plugins (like the upcoming Workflow catalog) can depend on it natively.
*   **Build Verification:** Successfully ran the memory-intensive `yarn build` using the 8GB node limit constraint to prevent OOMs. The Go tests and builds completed without regressions.

## Next Steps
*   **Command Catalog UI:** Now that the `SyncService` exists, build out the `CommandCatalogModalComponent` (or similar overlay). It needs to pull Workflows from the sync service, render them in a searchable list, and inject the selected command back into the `baseTerminalTab`'s Monaco IDE input editor.
*   **Hyper-Style Extensibility (Phase 5):** The next major target is implementing the React Web Component API wrapper. Developers should be able to push raw `.tsx` functional components into the configuration directory, and the Angular app should render them seamlessly as terminal overlays.

## Relevant Notes
*   If `yarn build` fails with `Allocation failed - JavaScript heap out of memory`, ensure you run it with `export NODE_OPTIONS=--max-old-space-size=8192`.
*   Remember to update `SUBMODULE_INVENTORY.md` and `VERSION.md` systematically after completing tasks.
*   Do not kill `node` processes blindly, as it terminates the underlying LLM agent environment.
