# HANDOFF.md

## Current Status (v1.0.231-nightly.4)
The user has directed the project to target 100% 1:1 feature parity with the top three modern "Agentic" and "Workspace" terminals: **Warp**, **WaveTerm**, and **Hyper**.

1.  **Deep Research**: I created `parity-expanded-research.md` detailing the specific features we need to adopt from each platform (Warp's IDE input and AI, WaveTerm's Rich Widget blocks and remote file editing, Hyper's aesthetic extensibility and React plugins).
2.  **Documentation Pivot**: I overhauled `VISION.md`, `ROADMAP.md`, `TODO.md`, `IDEAS.md`, and all AI instruction files (`AGENTS.md`, `CLAUDE.md`, etc.) to codify this ambitious roadmap. Every single setting, option, menu item, and function from these three terminals is now our target goal.
3.  **Code Consistency**: I fixed the remaining TypeScript compilation error for the `BlockFrontend` toggle in `tabby-terminal/src/api/baseTerminalTab.component.ts`. The repository builds entirely cleanly without errors (`yarn build` and `go test ./...`).

## Next Steps for the Next LLM Session
1.  **Implement WaveTerm-style Rich Widget Blocks**: The `BlockFrontend` currently just outputs raw text to a DOM span. We need to write an interceptor that detects specific OSC sequences from the `tabby-go` JSON-RPC bridge and renders a Markdown viewer or Monaco code editor block instead of text.
2.  **Implement Warp-style IDE Input**: Create a separate, pinned input `<textarea>` or Monaco editor instance at the bottom of `baseTerminalTab.component.pug` to decouple user input from the `xterm.js` output canvas. Handle multi-cursor and syntax highlighting logic.
