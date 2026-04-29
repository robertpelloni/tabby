# Warp Feature Analysis — Tabby Go Parity Research

> Source: `warp/` submodule (github.com/robertpelloni/warp)
> 3,122 Rust files · 1,239,297 lines of code · 65+ crates · Rust/WGPU/WASM

---

## Architecture Overview

Warp is a **Rust-native terminal emulator** with a custom GPU-rendered UI framework (WarpUI). It does NOT use Electron or xterm.js. The entire terminal grid, text rendering, and UI is drawn via **WGPU** (WebGPU) shaders with a Flutter-inspired Element system.

### Rendering Stack

| Layer         | Technology                                |
| ------------- | ----------------------------------------- |
| Windowing     | `winit` (custom fork)                     |
| GPU Rendering | `wgpu` (DX12/Vulkan/Metal/GLES)           |
| Text Shaping  | `font-kit` + custom glyph cache           |
| UI Framework  | Custom `warpui` (Entity-Component-Handle) |
| Terminal Grid | Custom `vte` fork (ANSI parser)           |
| Shaders       | WGSL (WebGPU Shading Language)            |

### Build Targets

- **Native** (macOS, Windows, Linux) — full feature set
- **WASM** — browser-based terminal (via `serve-wasm` crate)

---

## Feature Catalog (65 Crates)

### 1. CORE TERMINAL (`app/src/terminal/`, `crates/warp_terminal/`)

#### PTY Management

- **Local TTY** (`local_tty/`) — Full PTY spawn with platform-specific backends
  - **Unix**: `posix_openpt` / `fork` based PTY
  - **Windows**: **ConPTY** (`windows/conpty_api.rs`, `windows/pipes.rs`, `windows/child.rs`)
  - **Docker Sandbox** (`docker_sandbox.rs`) — Run commands in Docker containers
  - **WSL** (`wsl/`) — Windows Subsystem for Linux integration
- **Remote TTY** (`remote_tty/`) — Execute on remote servers via SSH
- **Event Loop** (`local_tty/event_loop.rs`) — Async PTY I/O with `mio`
- **Recorder** (`recorder.rs`) — Record/replay terminal sessions
- **Shell Detection** (`available_shells.rs`) — Auto-detect installed shells (bash, zsh, fish, powershell, cmd)
- **Terminal Manager** (`terminal_manager.rs`) — Lifecycle management for PTY instances

#### Terminal Model

- **Block-Based Output** (`model/block.rs`, `model/blocks.rs`) — Warp's signature feature: parses shell output into distinct "blocks" (command + output pairs)
- **Grid Rendering** (`model/grid/`) — Full ANSI/VT100 escape sequence handling via custom `vte` fork
  - `ansi_handler.rs` — CSI/OSC/DCS sequence processing
  - `grid_storage.rs` — Efficient in-memory grid with resize
  - `selection.rs` / `selection_cursor.rs` — Text selection (character, word, line, block)
  - `image.rs` — Inline image rendering (iTerm2, Kitty protocols)
  - `secrets.rs` — Secret masking (API keys, passwords detected via regex)
- **Alternate Screen** (`model/alt_screen.rs`) — Full support for vim, less, top, etc.
- **Rich Content** (`model/rich_content.rs`) — Markdown, code blocks, links in terminal output
- **Session Model** (`model/session/`) — Multiple session types:
  - `local_command_executor.rs` — Local shell commands
  - `remote_command_executor.rs` — SSH remote commands
  - `remote_server_executor.rs` — Warp server execution
  - `tmux_executor.rs` — Tmux integration
  - `wsl_command_executor.rs` — WSL commands
  - `msys2_command_executor.rs` — MSYS2 on Windows

#### Input System (`terminal/input/`)

- **IDE-Style Editor** — Multi-line, multi-cursor input box (Warp's signature)
  - `buffer_model.rs` — Input buffer with undo/redo
  - `decorations.rs` — Syntax highlighting, error underlines
  - `agent.rs` — AI agent input mode
  - `classic.rs` — Traditional terminal input mode
  - `universal.rs` — Universal search input
- **Inline Menu** (`inline_menu/`) — Autocomplete dropdown with fuzzy matching
- **Slash Commands** (`slash_commands/`) — `/` commands (like Discord)
- **Suggestions** (`suggestions_mode_menu.rs`) — AI-powered suggestions
- **Conversations** (`conversations/`) — AI conversation history
- **History Menu** — Up-arrow history with inline search (`inline_history/`)
- **Rewind** (`rewind/`) — Command history rewind/undo

#### Block System (`terminal/model/block/`)

- **Interaction Modes** (`interaction_mode.rs`) — Selectable, editable blocks
- **Serialized Blocks** (`serialized_block.rs`) — Persistence of block state
- **Block Filtering** (`block_filter.rs`) — Filter output by type
- **Block List Viewport** (`block_list_viewport.rs`) — Virtualized scrollable block list

### 2. SSH (`app/src/terminal/ssh/`)

- **SSH Connection** — Full SSH protocol support
- **Warpify** (`warpify.rs`) — Install Warp server on remote SSH host
- **Tmux Integration** (`install_tmux.rs`) — Auto-install tmux for session persistence
- **Root Access** (`root_access.rs`) — Sudo/root detection
- **SSH Detection** (`ssh_detection.rs`) — Auto-detect SSH environment
- **File Upload** (`view/ssh_file_upload.rs`) — SCP file transfer UI

### 3. AI SYSTEM (`crates/ai/`, `app/src/ai/`)

#### AI Agent (`ai/src/agent/`)

- **Actions** (`action/mod.rs`) — Tool use: file read, write, search, shell execution
- **Action Results** (`action_result/mod.rs`) — Structured output from AI actions
- **Citations** (`citation.rs`) — Source code citations in AI responses
- **File Locations** (`file_locations.rs`) — AI-aware file navigation

#### Codebase Indexing (`ai/src/index/`)

- **Full Source Code Embedding** (`full_source_code_embedding/`) — Semantic code search
  - `chunker/semantic.rs` — AST-aware code chunking
  - `merkle_tree/` — Incremental index updates via Merkle trees
  - `store_client.rs` — Vector store integration
- **File Outline** (`file_outline/`) — Symbol extraction via LSP

#### AI Skills (`ai/src/skills/`)

- **Skill Parser** (`parse_skill.rs`) — Parse `.md` skill definitions
- **Skill Provider** (`skill_provider.rs`) — Plugin system for AI capabilities
- **Read Skills** (`read_skills.rs`) — Load skills from filesystem

#### Other AI Features

- **API Keys** (`api_keys.rs`) — Multi-provider key management (OpenAI, Anthropic, etc.)
- **AWS Credentials** (`aws_credentials.rs`) — AWS Bedrock integration
- **Project Context** (`project_context/`) — AI context about current project
- **LLM ID** (`llm_id.rs`) — Multi-model selection (GPT-4, Claude, etc.)
- **Diff Validation** (`diff_validation/`) — Validate AI-generated code diffs

#### Agent Mode (`app/src/ai/`)

- **Agent Conversations** — Full conversational AI in terminal
- **Agent Management** — Notification system for AI agents
- **Multi-Agent** — Multiple AI agents running concurrently
- **CLI Agent Sessions** (`terminal/cli_agent_sessions/`) — Integration with Claude, Codex, Gemini, OpenCode
- **Computer Use** (`crates/computer_use/`) — AI can interact with GUI (screenshots, keyboard, mouse)

### 4. EDITOR (`crates/editor/`)

- **Multi-line Editor** — IDE-quality text editing in terminal input
- **Rich Content** — Markdown rendering (headers, lists, tables, images, Mermaid diagrams)
- **Text Operations** — Undo/redo, find/replace, selection, multi-cursor
- **Code Highlighting** — Syntax-aware with AST integration
- **Runnable Commands** — Clickable command blocks in output
- **Diff View** — Inline code diff with AI suggestions
- **Validation** — Input validation with error markers

### 5. COMMAND COMPLETION (`crates/warp_completer/`)

- **Autocomplete Engine** (`completer/engine/`) — Context-aware command completion
  - `command.rs` — Command name completion
  - `argument/` — Argument completion with descriptions
  - `flag/` — Flag/option completion
  - `path.rs` — File path completion
  - `variable.rs` — Environment variable completion
- **Command Signatures** (`signatures/`)
  - `clap.rs` — Parse Rust clap definitions
  - `v2/` — Enhanced signature format with JS extensions
  - `legacy/` — Backward-compatible signatures
  - `registry.rs` — 1000+ command definitions
- **Fuzzy Matching** (`matchers.rs`) — Fuzzy filter completions
- **Alias Expansion** (`suggest/alias.rs`) — Shell alias awareness
- **Priority System** (`suggest/priority/`) — ML-ranked suggestions

### 6. WORKFLOWS (`app/src/workflows/`)

- **Parameterized Scripts** — Saved, searchable shell scripts with parameters
- **Workflow Categories** (`categories.rs`) — Browsable catalog
- **Workflow Views** (`workflow_view/`) — UI for editing/running workflows
  - `argument_editor.rs` — Parameter input UI
  - `alias_argument_selector.rs` — Alias selection
  - `env_var_selector.rs` — Environment variable picker
  - `syntax_highlightable.rs` — Syntax-highlighted script editing
- **Cloud Sync** — Workflows synced to Warp Drive (cloud backend)
- **Sources**: Global, Local, Project, Team, Personal Cloud, Warp AI
- **Export** (`export_workflow.rs`) — Share workflows

### 7. THEMES (`app/src/themes/`)

- **23 Built-in Themes**: Dark, Light, Dracula, Solarized Dark/Light, Gruvbox Dark/Light, Jellyfish, Koi, Leafy, Marble, Pink City, Snowy, Dark City, Red Rock, Cyber Wave, Willow Dream, Fancy Dracula, Phenomenon, Solar Flare, Adeberry
- **Custom Themes** — User-created themes via UI
- **Base16 Themes** — Standard Base16 color scheme support
- **Theme Creator** (`theme_creator.rs`) — Visual theme builder with:
  - Color picking from wallpaper
  - Auto contrast calculation
  - Custom accent colors
- **Theme Chooser** (`theme_chooser.rs`) — Gallery view
- **Theme Deletion** (`theme_deletion_modal.rs`) — Manage themes

### 8. VERTICAL TABS (`workspace/view/vertical_tabs.rs`)

- **Left Sidebar Tabs** — Vertical tab list (your requested feature!)
- **Tab Data Model** — Tab type, title, icon, color, status
- **Context Chips** — Git status, language, project info per tab
- **Tab Settings** — Display granularity, compact mode, primary info selection
- **Drag & Drop** — Reorder tabs via drag
- **Close Buttons** — Individual tab close with confirmation
- **Undo Close** (`undo_close/`) — Reopen recently closed tabs

### 9. VIM MODE (`crates/vim/`)

- **Full Vim Emulation** — Modal editing (normal, insert, visual, command)
- **Text Objects** — `iw`, `i"`, `ip`, `ib` (word, quote, paragraph, block)
- **Find Char** — `f`, `F`, `t`, `T` motions
- **Matching Brackets** — `%` motion
- **Paragraph Iterator** — `{`, `}` motions
- **Registers** — Named yank/delete registers

### 10. WORKSPACE / PANE MANAGEMENT (`app/src/workspace/`, `app/src/pane_group/`)

- **Split Panes** — Horizontal and vertical splits
- **Pane Groups** — Nested split layouts
- **Tab Configurations** (`tab_configs/`) — Per-tab settings
  - `session_config.rs` — SSH/Local/Remote session config
  - `branch_picker.rs` — Git branch selection per tab
  - `repo_picker.rs` — Repository selection
  - `params_modal.rs` — Workflow parameter input
- **Left Panel** — Tabs, project explorer, Warp Drive, AI conversations
- **Right Panel** — AI agent panel
- **Global Search** (`workspace/view/global_search/`) — Search across files, commands, history
- **Command Palette** — VS Code-style command palette

### 11. SHARED SESSIONS (`terminal/shared_session/`)

- **Real-time Collaboration** — Multiple users in same terminal
- **Roles** — Sharer, Viewer with permission management
- **Presence** — See other users' cursors and selections
- **Heartbeat** (`network/heartbeat.rs`) — Keepalive mechanism
- **AI Agent Integration** (`ai_agent.rs`) — Shared AI sessions
- **Session Replay** — Replay shared sessions
- **Share Modal** — Invite users, manage permissions

### 12. NOTEBOOKS (`app/src/notebooks/`)

- **Interactive Notebooks** — Jupyter-like code notebooks in terminal
- **Code Cells** — Executable code blocks with output
- **Notebook Locations** — Local, cloud, team

### 13. LSP INTEGRATION (`crates/lsp/`)

- **Language Server Protocol** — Full LSP client
- **Supported Servers**:
  - TypeScript (`typescript_language_server.rs`)
  - Rust (`rust.rs`) — rust-analyzer
  - Python (`pyright.rs`)
  - C/C++ (`clangd.rs`)
  - Go (`go.rs`) — gopls
- **Auto-Install** (`install.rs`) — Install language servers automatically
- **Server Watcher** (`server_repo_watcher.rs`) — Watch for server updates

### 14. SEARCH & NAVIGATION

- **Find in Terminal** (`terminal/find/`) — Search within terminal output
  - Block list search, alt screen search, rich content search
- **Global Search** (`workspace/view/global_search/`) — Search everything
- **Ripgrep Integration** (`crates/warp_ripgrep/`) — Fast file content search
- **File Navigation** (`crates/warp_files/`) — File system operations
- **Link Detection** (`util/link_detection.rs`) — Clickable URLs in terminal

### 15. REMOTE SERVER (`crates/remote_server/`)

- **Remote Environments** — SSH-based remote development
- **Authentication** (`auth.rs`) — OAuth, token-based auth
- **Transport** (`transport.rs`) — SSH transport layer
- **Setup** (`setup.rs`) — Auto-configure remote servers
- **Protocol** (`protocol.rs`) — Binary communication protocol

### 16. CLOUD / WARP DRIVE (`app/src/drive/`)

- **Cloud Sync** — Settings, workflows, themes, conversations synced
- **Object Model** — Cloud objects (workflows, notebooks, conversations)
- **Team Management** (`workspaces/team.rs`) — Team-based sharing
- **User Profiles** (`workspaces/user_profiles.rs`)
- **GraphQL Client** (`crates/graphql/`) — Server communication

### 17. AUTHENTICATION (`app/src/auth/`)

- **Firebase Auth** (`crates/firebase/`) — Google authentication
- **OAuth** — GitHub, Google sign-in
- **Session Management** — Persistent login

### 18. INPUT CLASSIFICATION (`crates/input_classifier/`)

- **Natural Language Detection** (`natural_language_detection/`) — Detect if input is English or shell
- **ONNX Model** (`onnx/`) — ML-based classification
- **Heuristic Classifier** (`heuristic_classifier/`) — Rule-based fallback
- **FastText** (`fasttext/`) — Text classification model

### 19. MANAGED SECRETS (`crates/managed_secrets/`)

- **Encrypted Storage** — HPKE envelope encryption
- **GCP Integration** (`gcp.rs`) — Google Cloud secret management
- **Secret Values** (`secret_value.rs`) — Secure string handling

### 20. MISCELLANEOUS

- **Voice Input** (`crates/voice_input/`) — Dictate commands via microphone
- **Auto-Update** (`app/src/autoupdate/`) — In-app updates
- **Crash Reporting** (`app/src/crash_reporting/`) — Sentry integration
- **Package Installers** (`terminal/package_installers.rs`) — Auto-install tools
- **Tips System** (`app/src/tips/`) — Contextual tips and onboarding
- **Notifications** (`app/src/antivirus/`) — System notifications
- **Key Bindings** (`terminal/keys.rs`) — Fully customizable hotkeys
- **Image Support** — iTerm2 and Kitty image protocols
- **Kitty Keyboard Protocol** — Extended keyboard input
- **Settings** (`app/src/settings/`) — TOML-based configuration with schema validation

---

## Warp Feature Priority Matrix for Tabby Go

Given your priorities (local shells, many tabs, vertical tabs, stability, speed):

### Phase 1 — Core Terminal (CRITICAL)

| Feature             | Warp Implementation               | Tabby Go Status          |
| ------------------- | --------------------------------- | ------------------------ |
| ConPTY (Windows)    | `local_tty/windows/conpty_api.rs` | ✅ Done (via conpty lib) |
| PTY (Unix)          | `local_tty/unix.rs`               | ✅ Done (via creack/pty) |
| Terminal Grid/ANSI  | Custom `vte` fork                 | ⚠️ Using xterm.js        |
| Block-Based Output  | `model/block.rs`                  | ❌ Not started           |
| Multi-tab           | `workspace/vertical_tabs.rs`      | ✅ Basic sidebar done    |
| Terminal Resize     | `pty.Resize()`                    | ✅ Done                  |
| Shell Detection     | `available_shells.rs`             | ❌ Hardcoded             |
| Session Persistence | `undo_close/`                     | ❌ Not started           |

### Phase 2 — Editor & Input (HIGH)

| Feature          | Warp Implementation     | Tabby Go Status         |
| ---------------- | ----------------------- | ----------------------- |
| IDE Input Editor | `editor/` crate         | ❌ Using xterm.js stdin |
| Autocomplete     | `warp_completer/`       | ❌ Not started          |
| Command History  | `terminal/history/`     | ❌ Not started          |
| Find/Search      | `terminal/find/`        | ❌ Not started          |
| Vim Mode         | `crates/vim/`           | ❌ Not started          |
| Rich Content     | `model/rich_content.rs` | ❌ Not started          |

### Phase 3 — AI (MEDIUM)

| Feature         | Warp Implementation    | Tabby Go Status |
| --------------- | ---------------------- | --------------- |
| AI Agent        | `crates/ai/`           | ❌ Not started  |
| Codebase Index  | `ai/index/`            | ❌ Not started  |
| Skills          | `ai/skills/`           | ❌ Not started  |
| LSP Integration | `crates/lsp/`          | ❌ Not started  |
| Computer Use    | `crates/computer_use/` | ❌ Not started  |

### Phase 4 — Cloud & Collab (LOWER for now)

| Feature              | Warp Implementation | Tabby Go Status      |
| -------------------- | ------------------- | -------------------- |
| Themes (23 built-in) | `themes/`           | ❌ 1 hardcoded theme |
| Workflows            | `workflows/`        | ❌ Not started       |
| Shared Sessions      | `shared_session/`   | ❌ Not started       |
| SSH Remote           | `terminal/ssh/`     | ✅ Go backend done   |
| Cloud Sync           | `drive/`            | ❌ Not started       |
| Notebooks            | `notebooks/`        | ❌ Not started       |

---

## Key Technical Insights for Tabby Go

1. **Warp uses xterm.js differently** — Actually, Warp does NOT use xterm.js at all. It has a custom GPU-rendered terminal grid. We're using xterm.js which is fine for our purposes but has limitations for block-based output.

2. **Block-Based Output is Warp's killer feature** — Every command+output is a discrete "block" that can be selected, copied, edited, and shared. This is 50K+ lines of Rust.

3. **The IDE Input Editor is the other killer feature** — A multi-line, syntax-highlighted input box pinned to the bottom. This replaces the traditional terminal prompt.

4. **Vertical Tabs** — Warp has full vertical tab implementation with drag-drop, context chips, and compact mode.

5. **Warp is cross-platform** — macOS, Windows (ConPTY), Linux (posix_openpt), and WASM.

6. **The autocompleter has 1000+ command signatures** — A massive database of CLI tool argument/flag completions.
