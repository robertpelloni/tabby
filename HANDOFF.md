# Handoff Document

## Current Status
The project is currently at version **v1.0.231-nightly.17**.

## Recent Work
*   **Warp Drive Parity (Cloud Sync):** Built out the stub implementation for synchronizing terminal workflows, SSH profiles, and environment variables. The `tabby-go/pkg/sync` Go package now defines the `SyncData` format and exposes `Push()` and `Pull()` methods. These serialize data locally representing the "cloud" state.
*   **IPC Handlers:** Successfully mapped the `sync.push` and `sync.pull` JSON-RPC methods in the `server.go` event router.
*   **Build Verification:** Successfully ran the memory-intensive `yarn build` using the 8GB node limit constraint to prevent OOMs. The Go tests and builds completed without regressions.

## Next Steps
*   **Cloud Sync Frontend:** Wire up the Angular services in `tabby-core` or `tabby-terminal` to actually invoke `sync.push` and `sync.pull` through `ipcRenderer`. Construct the "Command Catalog" searchable UI to save and execute parameterized workflows.
*   **Hyper-Style Extensibility (Phase 5):** The next major target is implementing the React Web Component API wrapper. Developers should be able to push raw `.tsx` functional components into the configuration directory, and the Angular app should render them seamlessly as terminal overlays.

## Relevant Notes
*   If `yarn build` fails with `Allocation failed - JavaScript heap out of memory`, ensure you run it with `export NODE_OPTIONS=--max-old-space-size=8192`.
*   Remember to update `SUBMODULE_INVENTORY.md` and `VERSION.md` systematically after completing tasks.
*   Do not kill `node` processes blindly, as it terminates the underlying LLM agent environment.
