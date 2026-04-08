# Deployment Instructions

## Prerequisites

### All Platforms
- [Node.js](https://nodejs.org/) **v15+** (v22 recommended for builds)
- [Yarn](https://yarnpkg.com/) package manager
- Git

### Linux (Debian/Ubuntu)
```bash
sudo apt install libfontconfig-dev libsecret-1-dev libarchive-tools \
  libnss3 libatk1.0-0 libatk-bridge2.0-0 libgdk-pixbuf2.0-0 \
  libgtk-3-0 libgbm1 cmake
```

### Windows
- Visual Studio Build Tools (for native module compilation)
- Python 3 (for node-gyp)

### macOS
- Xcode Command Line Tools: `xcode-select --install`

## Development Setup

```bash
# Clone the repository
git clone https://github.com/robertpelloni/tabby.git
cd tabby

# Set up upstream remote (one-time)
git remote add upstream https://github.com/Eugeny/tabby

# Install dependencies
yarn

# Build all modules
yarn run build

# Start in development mode
yarn start
```

## Building for Production

### 1. Build all TypeScript modules
```bash
yarn run build
```

### 2. Pre-package plugins
```bash
node scripts/prepackage-plugins.mjs
```

### 3. Build platform installer

#### Windows
```bash
node scripts/build-windows.mjs
```
Outputs: `dist/tabby-*-setup-*.exe` (NSIS installer), `dist/tabby-*-portable-*.exe` (portable)

#### Linux
```bash
node scripts/build-linux.mjs
```
Outputs: `dist/tabby-*.deb`, `dist/tabby-*.rpm`, `dist/tabby-*.pacman`

#### macOS
```bash
node scripts/build-macos.mjs
```
Outputs: `dist/tabby-*-macos-*.dmg` (or .zip for notarized builds)

## Portable Mode (Windows)

Create a `data` folder in the same directory as `Tabby.exe` to run in portable mode.

## CI/CD

### GitHub Actions Workflows
- **build.yml** — Lint + build for macOS (x86_64/arm64), Windows, Linux
- **release.yml** — Automatic draft release on tag push (`v*`)
- **docs.yml** — Build and publish API documentation
- **codeql-analysis.yml** — Security analysis

### Creating a Release
```bash
# Update VERSION.md with new version number
# Update CHANGELOG.md with release notes
# Commit and tag
git add -A
git commit -m "Release v1.0.231"
git tag v1.0.231
git push origin master --tags
```

## Web App Deployment

The web version can be deployed separately:
```bash
cd web
yarn
yarn build
# Deploy the dist/ folder to your hosting
```

### Firebase Hosting
The project includes `firebase.json` for Firebase deployment.

## Syncing with Upstream

```bash
git fetch upstream
git merge upstream/master
# Resolve any conflicts, then:
git push origin master
```

## Troubleshooting

### Native Module Build Failures
```bash
# Clean and rebuild
rm -rf node_modules app/node_modules
yarn
```

### Electron Cache Issues
```bash
export ELECTRON_GET_USE_PROXY=true
# or set the electron mirror:
export ELECTRON_MIRROR="https://npmmirror.com/mirrors/electron/"
```

### WSL/Windows Build Issues
Make sure you have:
- Windows SDK
- Visual Studio Build Tools with C++ workload
- Python in PATH
