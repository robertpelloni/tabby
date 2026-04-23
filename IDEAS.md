# IDEAS.md - Tabby Improvement & Feature Ideas

This document contains a brainstorming list of features, refactors, and pivots for the Tabby Terminal project.

## 1. The Warp Parity Pivot (Agentic & Block-Based)
The traditional terminal emulator is dying. To stay relevant, Tabby must adopt the features of Warp and Fig:
*   **Blocks**: Parse the output stream so every command executed becomes a distinct DOM element (a "Block"). This allows users to hover over a command, copy *only* its output, or share a permalink to that specific block.
*   **IDE Input Area**: Instead of typing at the bottom of a scrolling text stream, pin a Monaco-editor style input box to the bottom. This allows multi-line editing, multi-cursor (`Alt+Click`), and real-time syntax highlighting before pressing Enter.
*   **Warp AI (Agent)**: Add a sidebar chat interface that has context of the current terminal buffer. It can read error messages and suggest fixes, or generate complex `ffmpeg` or `awk` commands based on natural language prompts.
*   **Workflows (Command Palette)**: A searchable catalog (`Cmd+P`) of parameterized, saved shell scripts that a team can sync to the cloud.

## 2. The WaveTerm Parity Pivot (Rich Content Workspaces)
Taking blocks a step further by embracing the terminal as a dynamic IDE:
*   **Rich Widgets**: A block shouldn't just be plain text. If a user runs `cat README.md`, Tabby should render a beautifully styled Markdown viewer inline within the block flow. If they run `cat image.png`, the image should render inline.
*   **Code Editor Blocks**: Running `edit config.json` shouldn't open `nano` in a text stream. It should pop open a fully functional Monaco text editor block with file-saving capabilities.
*   **Remote File Intelligence (`wsh`)**: We can leverage `tabby-go` by seamlessly installing it on remote SSH servers (like WaveTerm does with its `wsh` binary). Then, Tabby can negotiate complex remote file edits, system metrics (CPU/RAM), and process management natively without relying on parsing SSH text streams.

## 3. The Hyper Parity Pivot (Aesthetic Extensibility)
*   **Zero-Friction Plugins**: Tabby currently relies heavily on Angular Dependency Injection for plugins. This is powerful but steep. We should write a wrapper API that lets a user write a basic React component, drop it into `~/.tabby-plugins/`, and instantly render it over the terminal chrome (e.g., adding a CPU usage graph to the tab bar).
*   **Hot-Reloading Configurations**: Modifying `config.yaml` or a `.js` config file should never require clicking "Restart App". The UI should reactively reload colors, shortcuts, and plugins.

## 4. Refactoring & Architecture
*   **Rust/Tauri Rewrite (The "Nuclear" Option)**: Electron is memory-heavy. We are already moving native logic to Go (`tabby-go`). The next logical step is to replace Electron entirely with Tauri (Rust) for a tiny memory footprint, while keeping the Angular frontend.
*   **Headless Daemon Mode**: The `tabby-go` backend is fully decoupled from the UI via JSON-RPC. We should add a CLI flag (`tabby-go --daemon`) to run it headless on remote servers. Then, the Tabby UI could connect over WebSockets to this remote daemon, providing a seamless local-feeling terminal for remote machines.
*   **WebAssembly (Wasm) Frontend**: If we move away from Electron, the Angular frontend could be compiled to WebAssembly (or we could switch to a framework like Yew/Leptos) for near-native UI performance.

## 5. Quality of Life & Polish
*   **Command History Sync**: Sync bash/zsh history across all machines using a lightweight encrypted cloud backend.
*   **Built-in Snippet Manager**: A UI to manage and inject reusable text snippets into the terminal with variable placeholders (e.g., `ssh {{user}}@{{host}}`).
*   **Hex Viewer (Serial)**: For embedded developers using the Serial port feature, add a split-pane hex dump view alongside the ASCII output.
*   **Visual Jump Host Builder**: A node-graph UI to drag-and-drop SSH servers to construct complex `ProxyJump` chains visually.
*   **Live Collaboration**: "Multiplayer" terminal sessions where two users can type in the same terminal simultaneously (like VS Code Live Share).

## 6. UI/Theming Innovations
*   **Native MacOS/Windows Styling**: Currently, Tabby uses generic HTML/CSS styling. It should detect the host OS and seamlessly blend into MacOS (Vibrancy/Acrylic blurs) or Windows 11 (Mica materials).
*   **Command Status Indicators**: Instead of relying on shell prompt themes (like Starship), Tabby should natively render a green checkmark or red X next to the input box based on the exit code of the last block.
