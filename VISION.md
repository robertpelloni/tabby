# Tabby - Project Vision

> **Tabby** (formerly Terminus) is a highly configurable terminal emulator, SSH and serial client for Windows, macOS, and Linux.

## Ultimate Vision

To create the **most capable, beautiful, and extensible cross-platform terminal emulator** that serves as a unified interface for all remote and local computing tasks — from local shells to SSH sessions, serial connections, and beyond.

## Core Principles

1. **Cross-Platform Excellence**: First-class experience on Windows, macOS, and Linux
2. **Extensibility**: Plugin architecture allowing community and custom extensions
3. **Security**: Integrated encrypted vault for SSH secrets, secure credential management
4. **Performance**: Efficient terminal rendering that doesn't choke on fast output
5. **Accessibility**: Full Unicode support, configurable hotkeys, theming
6. **Connectivity**: SSH, Telnet, Serial, local shells — all in one application

## Architecture Overview

Tabby is built as an **Electron application** with an **Angular frontend** and a **plugin-based architecture**:

```
┌─────────────────────────────────────────────┐
│                  Electron Shell              │
│  ┌─────────────────────────────────────────┐ │
│  │           Angular 15 Frontend           │ │
│  │  ┌─────┐ ┌──────┐ ┌──────┐ ┌────────┐ │ │
│  │  │Core │ │Term  │ │SSH   │ │Settings│ │ │
│  │  └─────┘ └──────┘ └──────┘ └────────┘ │ │
│  │  ┌─────┐ ┌──────┐ ┌──────┐ ┌────────┐ │ │
│  │  │Local│ │Serial│ │Telnet│ │Plugins │ │ │
│  │  └─────┘ └──────┘ └──────┘ └────────┘ │ │
│  └─────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────┐ │
│  │           Native Layer                  │ │
│  │  node-pty │ russh │ serialport │ keytar │ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

## Plugin Architecture

Each plugin is a self-contained Angular module that:
- Exports a default `NgModule` class
- Provides functionality via Angular DI providers (singular or multi)
- Can extend UI via `ToolbarButtonProvider`, `ProfileProvider`, `TerminalDecorator`, etc.
- Gets loaded from source checkout (dev), user plugins dir, or `TABBY_PLUGINS` env var

### Extension Points
- **ToolbarButtonProvider** — Add buttons to the toolbar
- **ProfileProvider** — Add connection profile types
- **ConfigProvider** — Add configuration sections
- **HotkeyProvider** — Add keyboard shortcuts
- **TerminalDecorator** — Decorate/extend terminal tabs
- **TabContextMenuProvider** — Add context menu items
- **TabRecoveryProvider** — Recover tabs after restart
- **FileProvider** — Provide file system access

## Project Directory Structure

```
tabby/                          # Root project
├── app/                        # Electron app shell
│   ├── main.js                 # Electron main entry point
│   ├── src/                    # Angular app module, plugin loader
│   └── lib/                    # Main process TypeScript (CLI, config)
├── tabby-core/                 # Core framework plugin
│   └── src/
│       ├── api/                # Public API types and interfaces
│       ├── components/         # Base tab, split tab, headers
│       ├── services/           # Config, hotkeys, tabs, themes, vault
│       └── configDefaults*.yaml # Per-platform config defaults
├── tabby-terminal/             # Terminal emulation plugin
│   └── src/
│       ├── api/                # Terminal interfaces, color schemes
│       ├── components/         # Terminal tab UI components
│       ├── frontends/          # xterm.js frontend
│       ├── middleware/         # Input processing, login scripts, OSC, Zmodem
│       ├── features/           # Debug, Zmodem file transfer
│       └── services/           # Terminal management services
├── tabby-ssh/                  # SSH client plugin
│   └── src/
│       ├── api/                # SSH profile interfaces, importers
│       ├── session/            # SSH shell, SFTP, X11, port forwarding
│       ├── services/           # SSH service, password storage, known hosts
│       └── components/         # SSH-specific UI (host key prompt, etc.)
├── tabby-local/                # Local shell profiles plugin
│   └── src/
│       ├── profiles.ts         # WSL, CMD, PowerShell, Git-Bash, etc.
│       ├── session.ts          # Local PTY session management
│       └── services/           # Shell detection and management
├── tabby-serial/               # Serial terminal plugin
├── tabby-telnet/               # Telnet client plugin
├── tabby-settings/             # Settings UI plugin
├── tabby-electron/             # Electron-specific services
│   └── src/
│       ├── shells/             # Platform-specific shell integrations
│       │   ├── wsl.ts          # Windows Subsystem for Linux
│       │   ├── gitBash.ts      # Git Bash
│       │   ├── powershellCore.ts # PowerShell Core
│       │   ├── cmder.ts        # Cmder
│       │   ├── cygwin*.ts      # Cygwin
│       │   ├── msys2.ts        # MSYS2
│       │   └── linuxDefault.ts # Linux default shells
│       └── services/           # Docking, file provider, UAC, updater
├── tabby-plugin-manager/       # Plugin marketplace/manager
├── tabby-linkifier/            # Clickable URLs/IPs/paths
├── tabby-auto-sudo-password/   # Auto sudo password
├── tabby-community-color-schemes/ # Community themes
├── tabby-web/                  # Web-specific bindings
├── tabby-web-demo/             # Web demo app
├── tabby-uac/                  # Windows UAC helper (C# project)
├── web/                        # Web app entry point
├── build/                      # Build assets (icons, installers)
├── scripts/                    # Build and maintenance scripts
├── extras/                     # Clink distribution, UAC.exe, automator
├── locale/                     # i18n translations (25+ languages)
├── patches/                    # Patch-package patches
└── docs/                       # Documentation assets
```

## Key Technology Decisions

### russh (Rust SSH library)
- **Why**: Replaced the Node.js `ssh2` library with `russh` (Rust-based SSH implementation)
- **Benefits**: Better performance, memory safety, more reliable connection handling
- **Integration**: Compiled to native module via N-API, loaded from `app/` dependencies
- **Branch**: Originally developed on `origin/russh`, now merged to master

### Electron 38
- Latest Electron for security patches and performance improvements
- Custom protocol handler (`tabby://`)

### xterm.js v6
- Industry-standard terminal emulator component
- Image addon support (`xterm-addon-image`)
- Full VT220 + extensions emulation

### Angular 15
- Mature component framework for complex UI
- Dependency injection enables the plugin architecture
- ng-bootstrap for Bootstrap 5 UI components

## robertpelloni Fork Goals

1. **Enhanced Documentation**: Comprehensive project documentation, versioning, changelogs
2. **Go Backend Port**: Investigate porting performance-critical native backends to Go
3. **Feature Completeness**: Ensure all features are fully wired up and represented in UI
4. **Robustness**: Comprehensive testing, bug fixes, error handling
5. **Automated Workflows**: Streamlined CI/CD with version management

## Future Directions

### Short-term
- Complete Go backend investigation and proof-of-concept
- Comprehensive test suite
- UI polish and accessibility improvements
- Plugin API documentation improvements

### Medium-term
- Go-based native backend for SSH, PTY, and serial connections
- WebSocket-based terminal multiplexer
- Enhanced SFTP integration (bidirectional file manager)
- Container/Docker integration improvements

### Long-term
- Terminal multiplexing (tmux-like session persistence)
- Collaborative terminal sharing
- AI-assisted command suggestions
- Mobile companion app
