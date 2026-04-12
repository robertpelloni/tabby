# Tabby - Project Vision

To create the **most capable, beautiful, and extensible cross-platform terminal emulator** that serves as a unified interface for all remote and local computing tasks.

## Core Architectural Shift
Tabby is transitioning from relying on native Node.js modules (`node-pty`, `serialport`, `russh`) towards a unified, standalone **Go backend (`tabby-go`)**. This backend serves JSON-RPC 2.0 requests over standard input/output, significantly decoupling the performance-intensive system interactions from the Electron/V8 engine.

This sets the foundation for expanding Tabby beyond Electron into native GUI toolkits like BTK.
