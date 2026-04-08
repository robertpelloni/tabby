# Memory - Ongoing Observations About the Codebase

## Architecture Observations

### General Architecture
- Tabby is a **monorepo** with 15+ internal packages, all managed via yarn workspaces
- Each internal package is both an Angular module and an npm package
- The `app/` directory is the Electron shell that loads all plugins
- Build uses webpack 5 with a custom multi-config setup (`webpack.config.mjs`, `webpack.plugin.config.mjs`)
- TypeScript 4.9 (not 5.x yet) — Angular 15 compatibility constraint

### Plugin System
- Plugins are discovered by scanning for `package.json` files with `tabby-plugin` or `tabby-builtin-plugin` keywords
- Legacy `terminus-plugin` keyword still supported for backwards compatibility
- Plugin loading order matters: built-in plugins first, then user plugins
- Plugins communicate via Angular DI providers (multi-provider pattern)
- The `TABBY_PLUGINS` env var allows loading plugins from arbitrary paths during development

### SSH Implementation
- Uses `russh` (v0.1.36), a Rust SSH library, loaded via N-API native bindings
- The `origin/russh` branch was the development branch for the russh migration
- SSH features: shell sessions, SFTP, X11 forwarding, port forwarding (local/remote/dynamic), jump hosts, agent forwarding
- Password storage via keytar (system keychain)
- Known hosts management with host key verification prompts

### Terminal Layer
- xterm.js v6 as the terminal frontend
- Middleware pipeline: input processing → login scripts → OSC processing → stream processing → UTF8 splitting
- Features layered on top: Zmodem file transfer, debug mode
- Color scheme management is extensive (community-contributed schemes as a plugin)

### Configuration
- YAML-based configuration with platform-specific defaults (`configDefaults.{linux,macos,windows,web}.yaml`)
- Config merging uses deepmerge via `configMerge()` in ConfigService
- Hotkeys are configurable with multi-chord support
- Profiles are stored as arrays in config, with group support

### Build System
- `electron-builder` for packaging (Windows: NSIS/portable, macOS: DMG, Linux: deb/rpm/pacman/snap)
- Build scripts in `scripts/` directory (per-platform build scripts)
- Native modules compiled via `node-gyp` (keytar, serialport, node-pty, etc.)
- `patch-package` for patching node_modules (see `patches/` directory)

### Localization
- 25+ language translations in `locale/` directory (`.po` files)
- Uses `@ngx-translate/core` for i18n
- Crowdin integration for translation management (`crowdin.yml` might exist)

### Security
- Encrypted vault for SSH secrets (`VaultService`)
- Keytar integration for OS-level credential storage
- Windows UAC helper (`tabby-uac/` - C# project)
- SSH host key verification with user prompts

## Code Style Observations
- TypeScript with strict-ish settings but some `any` usage in IPC boundaries
- Angular components use Pug templates (`.pug`) and SCSS (`.scss`)
- Services follow Angular patterns with `@Injectable()` decorators
- IPC between main and renderer process uses `ipcRenderer.sendSync` in some places (could be improved)
- Some `var` usage in try/catch blocks for conditional requires (Node.js native modules)

## Known Technical Debt
- `app/package.json` version is `1.0.0-alpha.1` while all plugins are `1.0.231-nightly.0` — version mismatch
- No unit tests visible in the repository
- Some `eslint-disable` comments suggest style inconsistencies
- IPC uses synchronous calls in some places (performance bottleneck)
- The `tabby-uac/` module is a separate C# project with no build integration

## Design Preferences (robertpelloni)
- Preference for comprehensive documentation
- Single source of truth for version numbers (VERSION.md)
- Preference for Go as a backend language (porting initiative)
- Automated git workflow with version bumps tied to commits
- Cross-model development (Claude, Gemini, GPT working on same codebase)
