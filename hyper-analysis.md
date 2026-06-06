# Hyper Analysis & Implementation Status

## Hyper's Architecture Highlights
Hyper is built with Electron using a React + Redux stack. Its key features include:
1.  **Plugin System (PLUGINS.md)**: Plugins can "decorate" virtually any component (Terms, Term, Header, etc.) by wrapping them in Higher-Order Components (HOCs).
2.  **Configuration (hyper.json)**: Supports local and remote plugins, with hot-reloading for theme and layout changes.
3.  **Keymaps & Menus**: Plugins can inject new key bindings and menu items directly.
4.  **RPC Bridge**: Communicates between the main and renderer processes for global plugin features.

## Implementation in Tabby
While Tabby uses Angular, we have implemented several Hyper-inspired extensibility layers to reach parity:

### 1. React/WebComponent Support (`ReactPluginDecorator`)
- **Implemented**: `tabby-terminal/src/api/reactPlugin.ts` provides a global registry `window['tabbyReactPlugins']`.
- **Capability**: Allows loading external JavaScript that can mount React or raw DOM elements over the terminal view without requiring Angular compilation.
- **Parity**: Matches Hyper's ability to overlay widgets (e.g., status line, sidebars) using HOC-like logic.

### 2. Terminal Decorators (`TerminalDecorator`)
- **Implemented**: Tabby's `TerminalDecorator` API (in `tabby-terminal/src/api/decorator.ts`) allows modules to intercept terminal lifecycle events.
- **Parity**: Functionally equivalent to Hyper's component decoration, providing deep access to the `xterm.js` instance and container element.

### 3. Hot-Reloading Config
- **Implemented**: `ConfigService` in `tabby-core` watches the config file and emits events.
- **Parity**: UI elements reactively update when settings change, avoiding the "refresh to apply" friction.

## Roadmap for Full Hyper Parity
1.  **Plugin Management UI**: Need a better way to list, search, and enable local plugins (currently manual).
2.  **Global IPC Plugin Registry**: Extend the Go backend to also support "agent-style" plugins that can run long-running tasks in the background, similar to Hyper's `onApp` hooks.

## Conclusion
The core Hyper extensibility features are already present in Tabby's architecture. The `submodules/hyper` repository served its purpose for structural analysis and is now redundant.
