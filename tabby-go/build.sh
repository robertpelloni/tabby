#!/usr/bin/env bash
# Tabby Go & Wails Build Script for Linux/macOS
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "============================================"
echo "  Tabby Go - Unified Build Script"
echo "============================================"

# 1. Run backend unit tests
echo "[1/3] Running Go backend tests..."
go test ./...

# 2. Build Wails frontend production bundle
echo "[2/3] Building Wails frontend..."
(
  cd wails-app/frontend
  if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
  fi
  npm run build
)

# 3. Build Go Wails binary (detect OS and target appropriately)
echo "[3/3] Compiling Wails Go application..."
mkdir -p wails-app/bin
if command -v pkg-config >/dev/null 2>&1 && pkg-config --exists gtk+-3.0 webkit2gtk-4.0 2>/dev/null; then
  (
    cd wails-app
    go build -tags "desktop,production" -o bin/tabby-wails .
  )
  echo "  Binary created: tabby-go/wails-app/bin/tabby-wails"
else
  echo "  Note: Native Linux GTK headers not found in environment; cross-compiling Windows binary..."
  (
    cd wails-app
    GOOS=windows go build -tags "desktop,production" -o bin/tabby-wails.exe .
  )
  echo "  Binary created: tabby-go/wails-app/bin/tabby-wails.exe"
fi

echo "============================================"
echo "  Build completed successfully!"
echo "============================================"
