# DEPLOY.md - Tabby Terminal Deployment Instructions

## Overview
Tabby is an Electron application with an Angular frontend and a native Go backend daemon (`tabby-go`). Deployment involves building the Go binary, compiling the TypeScript frontend, and packaging them together via `electron-builder`.

## Prerequisites
*   **Node.js**: v18 or v20 (Strictly required for native module compatibility).
*   **Yarn**: v1.22.x (Classic). Do not use npm or yarn v3+.
*   **Go**: v1.21 or higher (Required to compile the backend daemon).
*   **Python 3**: For building native dependencies (if any remain).
*   **C++ Build Tools**:
    *   Windows: Visual Studio Build Tools (C++ workload).
    *   macOS: Xcode Command Line Tools.
    *   Linux: `build-essential`, `libx11-dev`, `libxext-dev` (depending on the distro).

## 1. Building the Go Backend
The Go backend must be compiled for the target architecture *before* packaging the Electron app.

```bash
cd tabby-go
# Install dependencies into vendor/
go mod vendor

# Run tests to ensure stability
env CGO_ENABLED=1 go test ./...

# Build the binary (Outputs to the bin/ directory)
# For Windows
GOOS=windows GOARCH=amd64 go build -o bin/tabby-backend.exe ./cmd/tabby-backend

# For macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/tabby-backend-darwin-arm64 ./cmd/tabby-backend

# For Linux
GOOS=linux GOARCH=amd64 go build -o bin/tabby-backend-linux-amd64 ./cmd/tabby-backend
```

*Note: `electron-builder` is configured to copy the contents of `tabby-go/bin/` into the final application bundle under `resources/`.*

## 2. Building the Frontend Workspaces
Tabby uses Lerna/Yarn Workspaces. You must install dependencies from the root and build all packages.

```bash
cd .. # Back to project root

# Install dependencies (This takes a while)
yarn install

# Build all Angular plugins and the Electron app
yarn build
```

## 3. Running in Development Mode
To test the application locally without packaging it:

```bash
yarn start
```
This will spawn the Electron process, which in turn spawns the local `tabby-go` binary from your `bin/` folder via the JSON-RPC bridge.

## 4. Packaging for Production
To create a distributable installer (e.g., `.exe`, `.dmg`, `.AppImage`):

```bash
# Package for the current OS/Architecture
yarn app:build

# To specify a platform (Requires appropriate cross-compilation tools):
yarn app:build:win
yarn app:build:mac
yarn app:build:linux
```

The resulting installers will be placed in the `app/dist/` directory.

## 5. Versioning and Releases
Before cutting a release, ensure the global version number is synchronized across all `package.json` files and the `CHANGELOG.md`.

```bash
# Increment the nightly version
node scripts/bump-version.mjs nightly

# Increment patch/minor/major
node scripts/bump-version.mjs patch

# Commit the version bump
git commit -am "chore: release vX.Y.Z"
git tag vX.Y.Z
# Git pushing the tags manually omitted here
```
