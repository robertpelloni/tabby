# Handoff Document

## Current Status
The project is currently at version **v1.0.231-nightly.18**.

## Recent Work
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
