# Changelog

## [1.0.231-nightly.1] - 2026-04-12
### Added
- Go Backend: Replaced `node-pty` with `creack/pty` mapped through Go daemon JSON-RPC interface.
- Go Backend: Replaced `@serialport` with `go.bug.st/serial` mapped through Go daemon JSON-RPC interface.
- Documentation: Added explicit ROADMAP, TODO, and CHANGELOG to track the porting process.

### Changed
- `tabby-local`: Routes shell instantiation requests through Electron IPC to `tabby-backend`.
- `tabby-serial`: Routes hardware communication through Electron IPC to `tabby-backend`.
- Preserved Web Serial functionality for `Platform.Web` while migrating desktop versions to Go.
