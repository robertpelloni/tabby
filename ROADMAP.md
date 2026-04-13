# Roadmap - Long-term Structural Plans

## Current Phase: Documentation & Stabilization

### Status: IN PROGRESS
- [x] Create comprehensive project documentation suite
- [x] Establish version management (VERSION.md as single source of truth)
- [x] Sync fork with upstream Eugeny/tabby
- [x] Document all modules and their purposes
- [ ] Establish automated version bump workflow
- [ ] Fix version mismatch (app: 1.0.0-alpha.1 vs plugins: 1.0.231-nightly.0)

---

## Phase 1: Go Backend Investigation & Proof of Concept
**Priority: HIGH** | **Status: PoC COMPLETE, Integration PENDING**

### Completed ✓
- [x] Created `tabby-go/` directory with Go module
- [x] Go SSH client with password, pubkey, agent auth
- [x] Shell sessions with PTY resize
- [x] Jump host support
- [x] Keepalive with disconnect detection
- [x] SFTP file management (list, upload, download, delete, rename, mkdir)
- [x] JSON-RPC 2.0 server over stdin/stdout
- [x] 25+ RPC methods registered
- [x] API types for all operations
- [x] Unit tests (all passing)
- [x] Buildable binary

### Remaining
- [ ] TypeScript client service for Electron-Go communication
- [ ] Wire Go backend into existing Angular SSH service
- [ ] Configuration option to switch between russh and Go backends
- [ ] PTY management in Go (using creack/pty for Unix, ConPTY for Windows)
- [ ] Serial port management in Go (using go.bug.st/serial)
- [ ] End-to-end integration testing

---

## Phase 2: Feature Completeness
**Priority: HIGH** | **Status: PARTIAL**

### Goal
Ensure every implemented feature is fully wired up, tested, and comprehensively represented in the UI.

### Areas Needing Attention
1. **SFTP Integration**
   - [ ] Full bidirectional file manager UI
   - [ ] Drag and drop file transfer
   - [ ] Progress indicators with cancel support
   - [ ] Directory tree browsing

2. **SSH Features**
   - [ ] SSH config file import (partial — importer exists but may be incomplete)
   - [ ] SSH multiplexer service (exists but needs testing)
   - [ ] Jump host chain management UI
   - [ ] Port forwarding management UI (config exists, UI may be limited)

3. **Terminal Features**
   - [ ] Split pane management improvements
   - [ ] Custom shell profile configuration polish
   - [ ] Terminal broadcast input (send to all tabs)

4. **Settings UI**
   - [ ] All config options represented with proper descriptions/tooltips
   - [ ] Import/export configuration
   - [ ] Profile group management

5. **Serial Terminal**
   - [ ] Hex view mode
   - [ ] Advanced flow control settings
   - [ ] Connection logging

---

## Phase 3: Testing & Quality
**Priority: MEDIUM** | **Status: NOT STARTED**

### Goal
Comprehensive test coverage for all modules.

### Plan
1. Unit tests for core services (ConfigService, HotkeysService, etc.)
2. Integration tests for SSH session lifecycle
3. E2E tests for terminal rendering
4. Plugin loading tests
5. Config migration tests
6. Cross-platform build verification

---

## Phase 4: Documentation & Developer Experience
**Priority: MEDIUM** | **Status: IN PROGRESS**

### Goal
World-class documentation for users and plugin developers.

### Plan
1. Complete API documentation (TypeDoc)
2. Plugin development guide with examples
3. User manual with screenshots
4. Architecture decision records
5. Contribution guidelines update
6. In-app help system

---

## Phase 5: Advanced Features
**Priority: LOW** | **Status: PLANNING**

### Terminal Multiplexing
- Persistent sessions (like tmux/screen)
- Session detach/attach
- Collaborative terminal sharing

### AI Integration
- Command suggestions based on history
- Natural language to command translation
- Error diagnosis and fix suggestions

### Mobile Companion
- Remote terminal access from mobile
- Session management
- Quick command execution

---

## Version Milestones

| Version | Target | Key Features |
|---------|--------|-------------|
| 1.0.231 | Current | Upstream sync, documentation, Go PoC |
| 1.0.232 | Next | Go backend SSH PoC, version management fix |
| 1.1.0 | Future | Go backend for all native modules |
| 2.0.0 | Long-term | Full Go backend, terminal multiplexing |
