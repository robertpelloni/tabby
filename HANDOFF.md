# HANDOFF.md

## Current Status (1.0.231-nightly.8)
The user has directed the project to target 100% 1:1 feature parity with the top three modern "Agentic" and "Workspace" terminals: **Warp**, **WaveTerm**, and **Hyper**.

1.  **Deep Research**: `parity-expanded-research.md` detailing the specific features we need to adopt from each platform (Warp's IDE input and AI, WaveTerm's Rich Widget blocks and remote file editing, Hyper's aesthetic extensibility and React plugins).
2.  **Documentation Pivot**: `VISION.md`, `ROADMAP.md`, `TODO.md`, `IDEAS.md`, and all AI instruction files (`AGENTS.md`, `CLAUDE.md`, etc.) codify this ambitious roadmap. Every single setting, option, menu item, and function from these three terminals is now our target goal.
3.  **AI Chat functionality completed**: Added `ai:chat` IPC channel handling to the frontend, electron app bridge, and the Go backend logic to give actual conversational mocked responses.

## Next Steps for the Next LLM Session
1.  **SFTP File Manager UI**: Implement full bidirectional file browser. Currently only context-menu download exists. Needs upload, delete, rename, create directory functionality, progress indicators, etc.
