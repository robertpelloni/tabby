# Tabby Autonomous Development Protocol

This document defines the rules and workflows for AI-driven autonomous development within the Tabby project. All participating models MUST adhere to these standards to ensure environment sanity, progress retention, and architectural consistency.

## 1. Principal Directive: Autopilot Execution
Operating on autopilot is mandatory for continuous execution.
- **Sequential Execution**: Features and tasks must be executed sequentially.
- **Commit Early, Commit Often**: Perform a clean `git commit` and `git push` (if applicable) after every major feature or verification step.
- **Zero Confirmation**: Proceed to the next task automatically without pausing for user approval unless a critical goal is entirely ambiguous.

## 2. Session Initialization (Context Restoration)
At the absolute start of every session:
1. **Analyze State**: Read local rules (`AGENTS.md`), repository structure, and key documentation (`VISION.md`, `ROADMAP.md`).
2. **Upstream Sync**:
   - Fetch all changes from the upstream parent fork (`Eugeny/tabby`).
   - Merge `upstream/master` into local `master`.
   - Resolve conflicts by prioritizing the native Go backend proxy architecture for SSH, SFTP, and Sync services.
3. **Submodule Cleanup**: Recursively update all submodules to their latest tracking commits.

## 3. Implementation Standards
- **Thin Proxy Architecture**: Angular frontend services (e.g., `SSHSession`, `SFTPSession`, `SyncService`) must be thin IPC proxies that route complex logic to the `tabby-go` daemon.
- **Native-First I/O**: For file operations (SFTP, Logging), always prefer passing absolute file paths to the Go backend instead of streaming binary data over Electron IPC.
- **Strict Versioning**: Use `node scripts/bump-version.mjs [patch|nightly]` to keep all 15+ packages synchronized. Reference the version bump in commit messages.

## 4. Verification & Testing
- **Unified Test Command**: Use `yarn test` from the root to execute both Go unit tests and the integration test suite.
- **Backend**: Every modification to Go code requires running `yarn test:go` (or `cd tabby-go && go test ./...`).
- **Agent Orchestration**: Autonomous workflows must be managed by the `tabby-go/pkg/agent` driver via the `agent.*` RPC methods.
- **Integration**: Verify JSON-RPC protocol updates using `yarn test:integration` (requires a built `build/tabby-backend`).
- **Frontend**: Visually verify UI modifications using Playwright screenshots.

## 5. Session Handoff
When hitting context limits or concluding a session:
1. Summarize learned architecture shifts and notable conflict resolutions.
2. Update `HANDOFF.md` with a structured log of completed merges and pending tasks.
3. Perform a final `git push` including all submodule updates.
