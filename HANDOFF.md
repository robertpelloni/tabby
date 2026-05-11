# HANDOFF.md

## Current Status (v1.0.231-nightly.6)
The user has directed the project to target 100% 1:1 feature parity with the top three modern "Agentic" and "Workspace" terminals: **Warp**, **WaveTerm**, and **Hyper**.

1.  **Deep Research**: I created `parity-expanded-research.md` detailing the specific features we need to adopt from each platform (Warp's IDE input and AI, WaveTerm's Rich Widget blocks and remote file editing, Hyper's aesthetic extensibility and React plugins).
2.  **Documentation Pivot**: I overhauled `VISION.md`, `ROADMAP.md`, `TODO.md`, `IDEAS.md`, and all AI instruction files (`AGENTS.md`, `CLAUDE.md`, etc.) to codify this ambitious roadmap. Every single setting, option, menu item, and function from these three terminals is now our target goal.
3.  **Code Consistency**: I fixed the remaining TypeScript compilation error for the `BlockFrontend` toggle in `tabby-terminal/src/api/baseTerminalTab.component.ts`. The repository builds entirely cleanly without errors (`yarn build` and `go test ./...`).

## Completed in this Session
1.  **Implement WaveTerm-style Rich Widget Blocks**: Integrated `marked` and `dompurify` to parse and render Markdown widgets from `WaveTermWidget` OSC sequences.
2.  **Implement Warp-style IDE Input**: Replaced the `<textarea>` input with a pinned `monaco-editor` instance in `baseTerminalTab.component` to provide multi-cursor support and syntax highlighting (Warp parity).

## Completed in this Session
1.  **Block Actions UI**: Implemented isolated `Copy Command` and `Copy Output` operations that extract text specifically from sub-containers instead of polluting the clipboard with UI text.
2.  **AI Command Search**: Added rudimentary keyword matching to the AI backend mock to demonstrate generating real shell commands (like `tar -xvf`) from natural language queries inside the Monaco editor.

## Next Steps for the Next LLM Session
1.  **Rich Widget Webviews/Images**: Extend the `renderWidget` parser to handle displaying inline image blobs or webviews to fully match WaveTerm's capabilities.
2.  **Explain Error AI Mock**: Fill out the `ai.go` backend mock logic for the `Explain Error` function to provide robust error parsing matching.
