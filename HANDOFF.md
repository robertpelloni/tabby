# HANDOFF.md

## Current Status (v1.0.231-nightly.8)
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

## Completed in this Session
1.  **Rich Widget Webviews/Images**: Upgraded `renderWidget` to support rendering inline `image` and `iframe` OSC widgets using safe DOM injection.
2.  **Explain Error AI Mock**: Extended `ai.go` to provide comprehensive diagnostic Markdown blocks in response to common shell errors (e.g. Permission Denied, Missing File, Command Not Found).

## Completed in this Session
1.  **Connect AI to Actual LLM API**: Finished the `ai.go` implementation using `net/http` targeting OpenAI's API. Provides a clean fallback mock if the API key isn't detected.
2.  **Hot-Reloading Config**: Finalized the `config.Manager` to poll `os.Stat` on the underlying YAML file. It correctly emits `host:config-change` via the JSON-RPC daemon bridge.

## Next Steps for the Next LLM Session
1.  **Terminal Agent Sidebar AI**: Connect the "Tabby AI Agent" sidebar in `baseTerminalTab.component` to a real conversational stream rather than just the generic "Explain Error" payload router.
2.  **Workflows Catalog UI**: Build out the Workflows execution catalog modal (`CommandCatalogModalComponent`) which is currently stubbed, giving users a way to save parameterized commands directly connected to their IDE input box.
