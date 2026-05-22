# TODO - Feature Tasks, Bug Fixes & Improvements

## Critical / High Priority

### Version Management
- [x] **Fix version mismatch**: `app/package.json` says `1.0.0-alpha.1` while all plugins are `1.0.231-nightly.0` ✅
- [x] **Automated version bump script**: Created `scripts/bump-version.mjs` ✅
- [x] **Bump version to 1.0.231-nightly.9**: Synchronized across all 15 package.json files. ✅

### Go Backend Port
- [x] **Create Go module structure** ✅
- [x] **Go SSH client PoC** ✅
- [x] **Go SFTP client** ✅
- [x] **Communication layer**: JSON-RPC 2.0 ✅
- [x] **TypeScript client**: Angular service proxy ✅
- [x] **PTY management in Go** ✅
- [x] **Serial port in Go** ✅
- [x] **SSH Port Forwarding** ✅
- [x] **SSH Proxy Support** ✅
- [x] **SSH Host Key Verification** ✅
- [x] **SSH Keyboard-Interactive Auth** ✅
- [x] **BTK Native UI Integration** ✅
- [x] **Go Backend Config** ✅
- [x] **Terminal Middleware** ✅
- [x] **Known Hosts Manager** ✅
- [x] **Session Recovery** ✅
- [x] **Notification System** ✅
- [x] **Cloud Sync**: Integrated `SyncService` to route requests to Go backend. ✅
- [x] **End-to-end testing**: Go backend tests passing. ✅
- [x] **SFTP Progress Tracking**: Real-time progress reporting implemented in Go and wired to frontend. ✅
- [x] **SFTP Path Optimization**: Direct path-based transfers implemented to bypass base64 memory overhead. ✅
- [ ] **Real PTY support**: Cross-platform PTY via creack/pty + Windows ConPTY
- [ ] **Real Serial support**: go.bug.st/serial integration
- [ ] **SFTP File Manager UI**: Basic UI restored (Rename, Create Dir, Upload, Download).
- [ ] **BTK Native Terminal App**: Full native UI with terminal rendering

### SSH Feature Completeness
- [x] **SFTP UI Restoration**: Fixed regressions in Rename and Create Directory functionality. ✅
- [x] **SFTP Progress indicators**: Real-time bars in global transfers menu. ✅
- [ ] **SFTP Drag-and-drop support polish**: Improve folder drop feedback.
- [ ] **Port forwarding management UI**:
  - Local, remote, and dynamic port forwarding config exists.
  - Need runtime management UI (add/remove forwards while connected).
- [x] **Jump host chain visualization**: Restored `jumpHostPath` in SSH tab UI. ✅
- [ ] **SSH config file import**: Verify and enhance `sshImporters.ts`.

### Build & Infrastructure
- [x] **Upstream Sync**: Merged `upstream/master` (Eugeny/tabby) into `master`. ✅
- [x] **Feature Branch Merging**: Merged SyncService and Release branches. ✅
- [ ] **Sync Electron versions**: electron-builder.yml, package.json, CI all consistent.
- [ ] **Automated release workflow**: Tag push → build → draft release.

---
[Rest of file unchanged...]
