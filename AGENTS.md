# AGENTS.md - Universal LLM Instructions
This file is the single source of truth for all AI dev tools (Google Jules, Claude, Gemini, GPT, Copilot, etc.) contributing to the Tabby Terminal project.

## 1. Project Context
*   **Goal**: Tabby is transitioning from a traditional Node.js/Electron terminal emulator (xterm.js) into a modern, Agentic, Block-Based, Collaborative Workspace powered by a unified Go daemon (`tabby-go`).
*   **The Vision (Warp, WaveTerm, and Hyper Parity)**: The ultimate goal is 1:1 feature parity with modern terminals. This means:
    *   **Blocks**: Parsing shell output into distinct, actionable DOM elements.
    *   **Rich Widgets (WaveTerm)**: Rendering Markdown, images, and Monaco code editors inline instead of just raw text.
    *   **IDE Editing (Warp)**: A pinned, multi-cursor, syntax-highlighted input box.
    *   **AI Integration**: Natural language command generation, error explanation, and conversational agents.
    *   **Workflows**: A searchable catalog of parameterized, saved shell scripts synced to a cloud backend.
    *   **Extensibility (Hyper)**: A simple API for users to drop React/JS scripts into a folder and instantly wrap or extend the UI chrome, alongside instantaneous hot-reloading of configuration files.
*   **The Architecture**: The Angular frontend services (`SSHSession`, `SFTPSession`, `LocalTerminal`, `SerialPort`) are thin IPC proxies. They must route all complex logic (PTY spawning, SSH, SFTP, hardware serial) to the `tabby-go` daemon via asynchronous `ipcRenderer.invoke()` channels in Electron.

## 2. Core Directives
1.  **Always Verify Your Work**: After every action that modifies the codebase, use a read-only tool (e.g., `yarn build`, `go test ./...`, `list_files`, `read_file`) to confirm the outcome. Do not mark a task as complete without verification.
2.  **Edit Source, Not Artifacts**: Never edit files in `dist`, `build`, or `target` directories. Trace the code back to its source, modify it, and run the appropriate build command.
3.  **Proactive Testing**: For any code change, attempt to run relevant unit tests (or write them). Use `yarn test` from the root to run both Go unit tests and the integration test suite. The frontend uses Webpack.
4.  **Diagnose Before Changing Environment**: Read error logs carefully. Do not immediately try to install/uninstall packages without understanding the expected environment setup.
5.  **Autonomous Problem Solving**: Strive to solve problems autonomously. Only ask for help if the request is ambiguous, you are stuck after multiple approaches, or you need a decision that significantly alters the scope.
6.  **Knowledgebase Lookup**: Utilize the `knowledgebase_lookup` tool early and often for guidance on bootstrapping, testing, tool issues, etc.
7.  **Commit Standards**: Every build must increment the version number globally (e.g., via `scripts/bump-version.mjs`). The commit message must reference the version bump and follow standard conventions (50 char subject, blank line, detailed body).
8.  **Document Everything**: You must read and update the project documentation (`VISION.md`, `ROADMAP.md`, `TODO.md`, `CHANGELOG.md`, `MEMORY.md`, `DEPLOY.md`, `HANDOFF.md`, `IDEAS.md`, `parity-expanded-research.md`) in comprehensive detail at the start and end of every session.
9.  **Submodules**: Maintain a global reference list with versions, URLs, and project directory structure locations for all submodules and libraries. Run `go mod vendor` when updating Go dependencies.

## 3. Code Style & Commenting
*   **In-Depth Comments**: Explain the reasoning, optimizations, side effects, and bugs of your code. Why is it there? Why is it the way it is?
*   **Self-Explanatory Code**: Leave simple, self-evident code uncommented. Don't comment just to comment.
*   **TypeScript**: Use strict typing. Avoid `any` where possible.
*   **Go**: Follow idiomatic Go practices.

## 4. Subagents
*   You may use subagents if possible to implement each feature, and commit/push to git in between each major step.
*   Remain autonomous for as long as possible without further confirmations. You may complete a feature, commit/push, and proceed autonomously.

## 5. Never Kill Node Processes
*   **DO NOT taskkill all node processes.** This will kill your own session and any other sessions. Don't do it!
