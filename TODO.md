# TODO - Feature Tasks, Bug Fixes & Improvements

## Critical / High Priority

### Warp & WaveTerm UI/UX Parity (Phase 2)
- [x] **Block-Based Output Parsing (BlockFrontend)**:
  - Extend the experimental `tabby-terminal/src/frontends/blockFrontend.ts`.
  - Implement a basic ANSI-to-HTML parser (or adapt `xterm.js` logic) so `span.textContent` doesn't just swallow control codes.
  - Intercept the PTY data stream and parse shell prompts (detect when a command ends and a new one begins).
  - Render each command execution as a distinct, isolated DOM element ("Block").
- [x] **IDE-Like Text Editing**:
  - Build a dedicated, pinned `contenteditable` (or Monaco Editor) input box at the bottom of the screen.
  - Implement Mouse click to place cursor.
  - Implement Multi-cursor editing (`Alt+Click`, `Cmd+D`).
  - Implement Intelligent fuzzy-search Tab completion.
  - Implement Real-time syntax highlighting and error validation (red squiggles).
- [ ] **Block Actions UI**:
  - Implement Copy command / Copy output for a specific block.
  - Implement Filtering/searching within a block.
  - Implement Generating shareable web links for a block.
  - Implement Keyboard navigation between blocks.
- [x] **Rich Widget Blocks (WaveTerm)**:
  - Add logic to intercept specific OSC codes from the `tabby-go` backend that tell the frontend to render the next block as Markdown, an Image, or a Code Editor buffer.

### Agentic AI Integration (Phase 3)
- [x] **AI Command Search**: Natural language to shell command generation within the IDE input box.
- [x] **Explain Error Action**: A one-click button on failed blocks that reads the command, stderr, and environment to suggest a fix.
- [x] **Terminal Agent Chat**: A dedicated sidebar/panel for conversational interaction with an AI.
- [x] **Workflow Generation**: AI-assisted creation of parameterized, saved shell scripts.

### Workflows & Collaboration (Phase 4)
- [x] **Command Catalog**: A searchable UI (Command Palette style) for saved, parameterized commands.
- [ ] **Cloud Sync Backend**: A secure backend service to synchronize Workflows, Environment Variables, and SSH Profiles.

---

## Medium Priority

### Hyper Parity (Phase 5)
- [x] **Hot Reloading Configuration**: Watch the Tabby config file and instantly re-render UI elements (like themes) without a restart.
- [x] **React/Web Component Plugin API**: Develop a wrapper around the Angular Dependency Injection system so users can write simple scripts exporting React components to customize the terminal chrome.

### Feature Completeness & Polish (Phase 6)
- [x] **SFTP File Manager UI**: Full bidirectional file browser, drag-and-drop file transfer, progress indicators, directory tree browsing.
- [x] **Port Forwarding Management UI**: Add/remove forwards while connected, status indicators for active forwards.
- [x] **Jump Host Chain Visualization**: Show jump host path in UI.
- [ ] **SSH Config File Import**: Verify and enhance `sshImporters.ts`.
- [ ] **Settings Descriptions/Tooltips**: Add descriptions to all config options in settings UI.
- [ ] **Profile Group Management**: Improve group editing and organization.
- [ ] **Terminal Broadcast**: Send input to all visible terminals.
- [ ] **Session Logging**: Record terminal output to file.
- [ ] **Serial Terminal**: Hex view mode, advanced flow control settings, connection logging.

### Serial Terminal
- [ ] **Hex view mode**: Show hex dump alongside ASCII
- [ ] **Advanced flow control settings**: Hardware flow control, parity, stop bits UI
- [ ] **Connection logging**: Log serial communication to file
- [ ] **Macro support**: Send pre-defined command sequences

### Testing
- [ ] **Unit test framework setup**: Jest or Karma for TypeScript tests
- [ ] **Core service tests**: ConfigService, HotkeysService, TabsService
- [ ] **SSH session tests**: Connection, authentication, data transfer
- [ ] **Plugin loading tests**: Discovery, loading, error handling
- [ ] **Config migration tests**: Upgrade paths between versions

### Code Quality
- [ ] **Migrate to TypeScript 5.x**: Currently on 4.9 (Angular 15 may need update)
- [ ] **Replace synchronous IPC**: Convert `sendSync` calls to async patterns
- [ ] **Reduce `any` types**: Add proper types in IPC boundaries
- [ ] **ESLint strict mode**: Enable stricter linting rules

---

## Low Priority

### Documentation
- [ ] **API documentation**: Complete TypeDoc for all public APIs
- [ ] **Plugin development guide**: Step-by-step tutorial
- [ ] **Architecture decision records**: Document why decisions were made
- [ ] **In-app help system**: Contextual help in settings

### Localization
- [ ] **Complete translation coverage**: Ensure all strings are translatable
- [ ] **RTL support**: Right-to-left language support
- [ ] **Language-specific fonts**: Ensure CJK, Arabic, etc. render correctly

### Performance
- [ ] **Lazy loading**: Load plugins on demand
- [ ] **Memory optimization**: Profile and reduce memory usage
- [ ] **Startup time**: Reduce cold start time

### Advanced Features
- [ ] **Terminal multiplexing**: tmux-like session persistence
- [ ] **Collaborative terminals**: Share terminal sessions
- [x] **AI command suggestions**: Based on history and context
- [ ] **Docker integration**: Built-in Docker container management
- [ ] **Kubernetes integration**: kubectl integration with context awareness
- [ ] **Snippet library**: Reusable command snippets with variables
- [ ] **Credential sync**: Cross-device credential synchronization

---

## Completed ✓
- [x] **Go Backend Parity (Phase 1)**: Integrated PTY, Serial, SSH, SFTP protocols within the Go daemon proxying through JSON-RPC 2.0. Finished proxyJump jump hosts integration and End-to-End integration testing.
- [x] **X11 Forwarding**: Implemented X11 socket forwarding in the Go backend (`pkg/ssh/x11.go`). Sent `x11-req` payload packet correctly.
- [x] **BlockFrontend Stub**: Built the experimental UI toggle for DOM-based block rendering over traditional `xterm.js` continuous streams.
- [x] **Version Management**: Fixed version mismatch and created `scripts/bump-version.mjs`.
- [x] **Go Backend Port**: Created `tabby-go/` directory, implemented PTY, Serial, SSH, and SFTP Go clients.
- [x] **Communication Layer**: Implemented JSON-RPC 2.0 over stdin/stdout with async notifications.
- [x] **Frontend Refactoring**: Rewrote `SSHSession` and `SFTPSession` to act as thin IPC proxy clients. Removed `russh` Node.js native dependency.
- [x] **Electron Main Process Router**: Registered `ipcMain.handle` endpoints to multiplex `ssh:*` and `sftp:*` requests to the `tabby-go` daemon.
- [x] **Build Fixes**: Cleaned up failing CGo code (`tabby-go/pkg/ui`) and fixed TypeScript strict compilation errors.
- [x] **Documentation**: Updated `VISION.md`, `ROADMAP.md`, `TODO.md`, `IDEAS.md`, `CHANGELOG.md`, `MEMORY.md`, `DEPLOY.md`, and `HANDOFF.md` to reflect the massive Warp, WaveTerm, and Hyper parity vision.
