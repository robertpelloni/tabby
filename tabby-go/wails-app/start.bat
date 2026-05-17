@echo off
REM Tabby Go - Development Start Script
REM This script builds and runs the Wails application in dev mode.

echo ============================================
echo   Tabby Go - Starting Development Server
echo ============================================
echo.

REM Change to the wails-app directory
cd /d "%~dp0"

REM Check if Wails is installed
where wails >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Wails CLI not found in PATH.
    echo         Install it with: go install github.com/wailsapp/wails/v2/cmd/wails@latest
    echo.
    pause
    exit /b 1
)

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go is not installed or not in PATH.
    echo         Download from: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo [INFO] Wails found: 
where wails

echo [INFO] Go version:
go version

echo.
echo [1/2] Running Go tests...
cd .. && go test ./... 2>&1
if %ERRORLEVEL% neq 0 (
    echo.
    echo [WARN] Some Go tests failed. Continuing anyway...
    echo.
) else (
    echo [OK] All Go tests passed.
    echo.
)

cd wails-app

echo [2/2] Starting Wails dev server...
echo        (This will open the app with hot-reload enabled)
echo        (Press Ctrl+C in this window to stop)
echo.

wails dev

if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] Wails dev exited with an error.
    echo.
    pause
    exit /b 1
)
