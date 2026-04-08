# Universal LLM Instructions

> **This file is the single source of instructions for ALL AI models working on this project.**
> 
> Each model may have its own file (CLAUDE.md, GEMINI.md, GPT.md, copilot-instructions.md) that references this file and appends model-specific instructions.

## Project: Tabby (robertpelloni fork)

**Repository**: https://github.com/robertpelloni/tabby  
**Upstream**: https://github.com/Eugeny/tabby  
**Language**: TypeScript (frontend), Go (backend - in progress)  
**Framework**: Angular 15, Electron 38  
**Build**: Webpack 5, electron-builder  
**Version**: See [VERSION.md](../VERSION.md) — single source of truth

## Mandatory Rules

### 1. Session Start Protocol
1. Read this file completely
2. Read [VERSION.md](../VERSION.md) for current version
3. Read [MEMORY.md](../MEMORY.md) for ongoing observations
4. Read [TODO.md](../TODO.md) for current task list
5. Read [ROADMAP.md](../ROADMAP.md) for long-term plans
6. Check git status: `git status && git log --oneline -5`

### 2. Session End Protocol
1. Update [MEMORY.md](../MEMORY.md) with new observations
2. Update [TODO.md](../TODO.md) with completed/new items
3. Update [ROADMAP.md](../ROADMAP.md) if structural plans changed
4. Update [CHANGELOG.md](../CHANGELOG.md) with all changes
5. Document everything in [HANDOFF.md](../HANDOFF.md) for next model
6. Commit and push: `git add -A && git commit -m "descriptive message" && git push`

### 3. Version Management
- **VERSION.md** is the single source of truth for the version number
- Every build should have a new version number
- When version is updated:
  1. Update `VERSION.md`
  2. Run version sync script (or manually update all `package.json` files)
  3. Update `CHANGELOG.md` with new section
  4. Git commit with version bump in message
  5. Git push

### 4. Git Workflow
- **Branch**: Work on `master` branch (robertpelloni fork)
- **Upstream sync**: `git fetch upstream && git merge upstream/master`
- **Commit frequency**: Commit after each feature/fix
- **Push frequency**: Push after each commit
- **Merge conflicts**: Always resolve intelligently, never lose features
- **Feature branches**: Merge any robertpelloni feature branches into main

### 5. Code Style
- TypeScript with Angular patterns
- Components use Pug templates and SCSS
- Services use `@Injectable()` decorators
- Comment code that needs commenting, don't over-comment
- Follow existing code patterns in the codebase
- No `taskkill` commands — will kill the session

### 6. Documentation Requirements
- [VISION.md](../VISION.md) — Ultimate project goal and design
- [MEMORY.md](../MEMORY.md) — Ongoing observations and preferences
- [DEPLOY.md](../DEPLOY.md) — Deployment instructions (always up to date)
- [CHANGELOG.md](../CHANGELOG.md) — Detailed changelog with version
- [ROADMAP.md](../ROADMAP.md) — Long-term structural plans
- [TODO.md](../TODO.md) — Individual features, bug fixes, details
- [IDEAS.md](../IDEAS.md) — Creative improvement ideas
- [HANDOFF.md](../HANDOFF.md) — Session handoff documentation
- [VERSION.md](../VERSION.md) — Version string only (single source of truth)

### 7. Project Structure Awareness
- This is a **fork** of Eugeny/tabby with custom modifications
- 15+ internal packages (all `tabby-*` directories)
- Plugin-based architecture — each package is an Angular module
- `app/` is the Electron shell, `web/` is the web app entry
- `tabby-uac/` is a C# project (Windows UAC helper)
- All internal packages share the same version (in their `package.json`)

### 8. Go Porting Initiative
- Goal: Port performance-critical native backends to Go
- Priority areas: SSH client, PTY management, Serial port, SFTP
- Go code should go in `tabby-go/` directory
- Communication via JSON-RPC or gRPC to Electron
- Progressive migration — don't break existing functionality

### 9. Handoff Protocol
After completing work, document in HANDOFF.md:
- What was accomplished
- What was attempted but failed
- What the next model should do
- Any blockers or issues found
- Files modified
- Decisions made and rationale

### 10. Quality Standards
- Test all functions after implementation
- Double-check all features work end-to-end
- No bugs in committed code (fix before committing)
- Every feature comprehensively represented in UI
- All functionality documented in UI (labels, tooltips, descriptions)
- Robust error handling

## Key Files to Know

| File | Purpose |
|------|---------|
| `VERSION.md` | Version number (single source of truth) |
| `CHANGELOG.md` | Changelog with version history |
| `app/main.js` | Electron main entry point |
| `app/src/app.module.ts` | Angular root module |
| `tabby-core/src/api/` | Public API interfaces |
| `tabby-core/src/services/` | Core Angular services |
| `tabby-core/src/configDefaults.yaml` | Default configuration |
| `tabby-ssh/src/session/ssh.ts` | SSH session implementation |
| `tabby-ssh/src/session/sftp.ts` | SFTP implementation |
| `tabby-terminal/src/frontends/` | xterm.js frontend |
| `tabby-terminal/src/middleware/` | Terminal middleware pipeline |
| `tabby-electron/src/pty.ts` | PTY interface |
| `tabby-electron/src/shells/` | Platform-specific shells |
| `electron-builder.yml` | Build configuration |
| `webpack.config.mjs` | Webpack configuration |
| `scripts/` | Build and utility scripts |
