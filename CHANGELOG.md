# Changelog

All notable changes to the Tabby (robertpelloni fork) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Version number is maintained in [VERSION.md](VERSION.md) and is the single source of truth.

## [1.0.231-nightly.1] - 2025-04-08

### Added (robertpelloni fork)
- **Go PTY manager**: Process spawning with stdin/stdout/stderr forwarding
- **Go serial port manager**: Stub for go.bug.st/serial integration
- **TypeScript GoBackendService**: Angular service for Go backend communication
  - JSON-RPC 2.0 client over stdin/stdout
  - RxJS observables for data/exit notifications
  - Full TypeScript API for SSH, SFTP, PTY, Serial
- **Config option**: `goBackend.enabled` to switch between russh and Go backends

## [1.0.231-nightly.0] - 2025-04-08

### Upstream Sync
- Synced fork with upstream Eugeny/tabby master (f05a07ae)
- Configured upstream remote (https://github.com/Eugeny/tabby)

### Added
- **russh-based SSH backend**: The SSH client now uses the `russh` Rust library (v0.1.36) via N-API bindings instead of the previous ssh2-based implementation. This provides better performance and reliability for SSH connections.
- **SFTP context menu download**: Files can now be downloaded via SFTP context menu directly from SSH sessions (#10971)
- **tabby:// URL scheme handler**: Custom URL protocol handler allows opening Tabby via `tabby://` URLs (#11005)
- **Quick connection support for third-party plugins**: Plugins can now add quick connection providers (#11004)
- **X11 forwarding fix**: Fixed X11 forwarding for SSH sessions (#10580, #10558, #10237, #8450)
- **Webpack upgrade**: Upgraded webpack from 5.86.0 to 5.104.1 to fix ES6 class extension bug with xterm.js v6 (#11007)
- **Recovered tabs crash fix**: Fixed crashes when recovering tabs after upgrade (#10619, #10635, #10672, #10682, #10688, #10730, #10843)
- **macOS dock hiding fix**: Fixed hiding the app from macOS dock when docking is enabled (#10960)

### Added (robertpelloni fork)
- **Go backend proof of concept** (`tabby-go/`):
  - SSH client with password, public key, and agent authentication
  - Shell sessions with PTY resize support
  - Jump host / proxy jump chains
  - Keepalive with disconnect detection
  - SFTP file management (list, upload, download, delete, rename, mkdir)
  - JSON-RPC 2.0 server over stdin/stdout
  - 25+ RPC methods for SSH, SFTP, PTY, and Serial
  - Unit tests (all passing)
  - Buildable binary (tabby-backend.exe, 7.2MB)
- **Version management system**:
  - `VERSION.md` as single source of truth
  - `scripts/bump-version.mjs` for automated version synchronization
  - Fixed `app/package.json` version mismatch
- **Comprehensive documentation suite**:
  - `VERSION.md` — Version string
  - `VISION.md` — Project vision and architecture
  - `CHANGELOG.md` — This changelog
  - `ROADMAP.md` — Long-term structural plans (5 phases)
  - `TODO.md` — Feature tasks organized by priority
  - `MEMORY.md` — Codebase observations and preferences
  - `DEPLOY.md` — Deployment instructions
  - `IDEAS.md` — 30 creative improvement ideas
  - `HANDOFF.md` — Session handoff documentation
  - `docs/UNIVERSAL_LLM_INSTRUCTIONS.md` — Universal LLM instructions
  - `CLAUDE.md`, `GEMINI.md`, `GPT.md`, `copilot-instructions.md` — Model-specific instructions

### Project Modules
| Module | Version | Description |
|--------|---------|-------------|
| tabby-core | 1.0.231-nightly.0 | Core UI framework, tab management, services |
| tabby-electron | 1.0.231-nightly.0 | Electron-specific bindings (PTY, docking, shell integration) |
| tabby-terminal | 1.0.231-nightly.0 | Terminal emulation (xterm.js frontend, middleware, features) |
| tabby-ssh | 1.0.231-nightly.0 | SSH2 client with connection manager, SFTP, X11 forwarding |
| tabby-local | 1.0.231-nightly.0 | Local shell profiles (WSL, Git-Bash, PowerShell, CMD, etc.) |
| tabby-serial | 1.0.231-nightly.0 | Serial terminal connections |
| tabby-telnet | 1.0.231-nightly.0 | Telnet/socket connections |
| tabby-settings | 1.0.231-nightly.0 | Settings UI (profiles, hotkeys, vault, appearance) |
| tabby-plugin-manager | 1.0.231-nightly.0 | Plugin installation and management |
| tabby-linkifier | 1.0.231-nightly.0 | Clickable URLs, IPs, and file paths in terminal |
| tabby-auto-sudo-password | 1.0.231-nightly.0 | Auto-paste sudo password in SSH sessions |
| tabby-community-color-schemes | 1.0.231-nightly.0 | Community-contributed color schemes |
| tabby-web | 1.0.231-nightly.0 | Web-specific bindings for browser deployment |
| tabby-web-demo | 1.0.231-nightly.0 | Web demo application |
| tabby-uac | N/A | Windows UAC helper (C# solution, not a TS module) |
| app (Electron shell) | 1.0.0-alpha.1 | Main Electron application shell |
| web | N/A | Web app entry point |

### Key Technologies
- **Frontend**: TypeScript, Angular 15, Pug templates, SCSS
- **Backend/Native**: Electron 38, Node.js, node-pty, russh (Rust SSH via N-API)
- **Build**: Webpack 5, TypeScript 4.9, electron-builder
- **Terminal**: xterm.js v6 with image addon
- **UI Framework**: Bootstrap 5 (via ng-bootstrap), FontAwesome 6
- **State**: Angular services, RxJS observables
- **Config**: YAML-based configuration with platform-specific defaults
