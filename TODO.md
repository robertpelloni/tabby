# TODO - Feature Tasks, Bug Fixes & Improvements

## Critical / High Priority

### Version Management
- [x] **Fix version mismatch**: `app/package.json` says `1.0.0-alpha.1` while all plugins are `1.0.231-nightly.0`
  - Made `app/package.json` version match plugins (1.0.231-nightly.0)
  - Created `scripts/bump-version.mjs` to synchronize versions
  - File: `scripts/bump-version.mjs` ✅
- [x] **Automated version bump script**: Created `scripts/bump-version.mjs` that:
  - Reads current version from `VERSION.md`
  - Accepts semver bump type (patch/minor/major/nightly)
  - Updates all `package.json` version fields
  - Updates `CHANGELOG.md` with new section

### Go Backend Port
- [x] **Create Go module structure**: `tabby-go/` directory with `go.mod` ✅
- [x] **Go SSH client PoC**: Basic SSH connection using `golang.org/x/crypto/ssh` ✅
  - Authenticate with password and public key
  - Open shell session
  - Handle PTY resize
  - Jump host support
  - Keepalive with disconnect detection
- [x] **Go SFTP client**: Full file management ✅
  - List, download, upload, delete, rename, mkdir
- [x] **Communication layer**: JSON-RPC 2.0 over stdin/stdout ✅
  - 25+ methods registered
  - Async notifications for data/exit events
- [ ] **TypeScript client**: Angular service to communicate with Go backend
- [ ] **PTY management in Go**: Cross-platform PTY spawning (stub exists)
- [ ] **Serial port in Go**: Serial communication (stub exists)

### SSH Feature Completeness
- [ ] **SFTP file manager UI**: Full bidirectional file browser
  - Currently only context-menu download exists (#10971)
  - Need upload, delete, rename, create directory
  - Need drag-and-drop support
  - Need progress indicators
- [ ] **Port forwarding management UI**: 
  - Local, remote, and dynamic port forwarding config exists
  - Need runtime management UI (add/remove forwards while connected)
  - Need status indicators for active forwards
- [ ] **Jump host chain visualization**: Show jump host path in UI
- [ ] **SSH config file import**: Verify and enhance `sshImporters.ts`

### Build & Infrastructure
- [ ] **Sync Electron versions**: electron-builder.yml, package.json, CI all consistent
- [ ] **Automated release workflow**: Tag push → build → draft release
- [ ] **Cross-platform build verification**: Test builds on all platforms

---

## Medium Priority

### UI/UX Improvements
- [ ] **Settings descriptions/tooltips**: Add descriptions to all config options in settings UI
- [ ] **Profile group management**: Improve group editing and organization
- [ ] **Hotkey conflict detection**: Warn when hotkeys conflict
- [ ] **Tab search**: Search/filter open tabs
- [ ] **Command palette**: Quick command execution (like VS Code)
- [ ] **Split pane management**: Better controls for pane creation, navigation, resizing

### Terminal Features
- [ ] **Terminal broadcast**: Send input to all visible terminals
- [ ] **Session logging**: Record terminal output to file
- [ ] **Scrollback buffer configuration**: More options for scrollback size
- [ ] **Terminal profile templates**: Pre-configured profiles for common tasks
- [ ] **Custom prompts integration**: Better shell integration detection

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
- [ ] **AI command suggestions**: Based on history and context
- [ ] **Docker integration**: Built-in Docker container management
- [ ] **Kubernetes integration**: kubectl integration with context awareness
- [ ] **Snippet library**: Reusable command snippets with variables
- [ ] **Credential sync**: Cross-device credential synchronization

---

## Completed ✓
- [x] Upstream sync with Eugeny/tabby
- [x] russh-based SSH backend (upstream)
- [x] SFTP context menu download (upstream)
- [x] tabby:// URL scheme handler (upstream)
- [x] Webpack 5.104.1 upgrade (upstream)
- [x] X11 forwarding fix (upstream)
- [x] Comprehensive project documentation (this session)
- [x] VERSION.md single source of truth (this session)
- [x] CHANGELOG.md (this session)
- [x] VISION.md (this session)
- [x] ROADMAP.md (this session)
- [x] MEMORY.md (this session)
- [x] DEPLOY.md (this session)
- [x] UNIVERSAL_LLM_INSTRUCTIONS.md (this session)
