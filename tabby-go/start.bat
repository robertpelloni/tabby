@echo off
setlocal enabledelayedexpansion
title Tabby Go

:: ═══════════════════════════════════════════════════════════════
:: Tabby Go - Terminal Backend & Native App
:: Module:  github.com/robertpelloni/tabby/tabby-go
:: Entries:  cmd/tabby-backend  ->  tabby-backend.exe
::           cmd/tabby-native   ->  tabby-native.exe
::           wails-app/         ->  tabby-wails.exe
:: ═══════════════════════════════════════════════════════════════

cd /d "%~dp0"

set "BACKEND=tabby-backend.exe"
set "NATIVE=tabby-native.exe"
set "WAILS=tabby-wails.exe"

:: ─── Parse command ──────────────────────────────────────────
set "CMD=%1"
if "%CMD%"=="" set "CMD=run"
if /i "%CMD%"=="run"     goto :run
if /i "%CMD%"=="build"   goto :build
if /i "%CMD%"=="backend" goto :backend
if /i "%CMD%"=="native"  goto :native
if /i "%CMD%"=="wails"   goto :wails
if /i "%CMD%"=="test"    goto :test
if /i "%CMD%"=="clean"   goto :clean
if /i "%CMD%"=="help"    goto :help
echo Unknown command: %CMD%
goto :help

:: ─── Build all 3 binaries ───────────────────────────────────
:build
echo.
echo  [Tabby Go] Building all binaries...
echo.
echo  [1/3] %BACKEND% (JSON-RPC over stdio)...
go build -buildvcs=false -o %BACKEND% ./cmd/tabby-backend
if errorlevel 1 ( echo  [FAIL] %BACKEND% & exit /b 1 )
for %%f in (%BACKEND%) do echo  [OK]   %%~zf bytes
echo  [2/3] %NATIVE% (BTK native terminal)...
go build -buildvcs=false -o %NATIVE% ./cmd/tabby-native
if errorlevel 1 ( echo  [FAIL] %NATIVE% & exit /b 1 )
for %%f in (%NATIVE%) do echo  [OK]   %%~zf bytes
echo  [3/3] %WAILS% (Wails GUI)...
cd wails-app
go build -buildvcs=false -o ..\%WAILS% .
if errorlevel 1 ( cd .. & echo  [FAIL] %WAILS% & exit /b 1 )
cd ..
for %%f in (%WAILS%) do echo  [OK]   %%~zf bytes
echo.
echo  [Tabby Go] Build complete.
goto :end

:: ─── Run (launch all) ──────────────────────────────────────
:run
if not exist %BACKEND% call :build
if errorlevel 1 exit /b 1
echo.
echo  [Tabby Go] Launching all services...
start /b %BACKEND%
start /b %NATIVE%
start "" %WAILS%
echo  [Tabby Go] All services started.
goto :end

:: ─── Individual services ────────────────────────────────────
:backend
if not exist %BACKEND% ( go build -buildvcs=false -o %BACKEND% ./cmd/tabby-backend || exit /b 1 )
%BACKEND%
goto :end

:native
if not exist %NATIVE% ( go build -buildvcs=false -o %NATIVE% ./cmd/tabby-native || exit /b 1 )
%NATIVE%
goto :end

:wails
if not exist %WAILS% ( cd wails-app && go build -buildvcs=false -o ..\%WAILS% . && cd .. || exit /b 1 )
%WAILS%
goto :end

:: ─── Test ───────────────────────────────────────────────────
:test
echo  [Tabby Go] Running tests...
go test ./pkg/... ./internal/... -v -count=1 -timeout 120s 2>&1 | findstr /V "no test files"
goto :end

:: ─── Clean ──────────────────────────────────────────────────
:clean
del /q %BACKEND% 2>nul
del /q %NATIVE% 2>nul
del /q %WAILS% 2>nul
go clean
echo  [Tabby Go] Cleaned.
goto :end

:: ─── Help ───────────────────────────────────────────────────
:help
echo.
echo  Tabby Go - Usage: start.bat [command]
echo.
echo  Commands:
echo    run       Build and launch all 3 services (default)
echo    build     Build all binaries
echo    backend   Run JSON-RPC backend only
echo    native    Run BTK native terminal only
echo    wails     Run Wails GUI only
echo    test      Run tests
echo    clean     Remove binaries
echo    help      Show this help
echo.
echo  Binaries:
echo    tabby-backend.exe  JSON-RPC backend (spawned by Electron)
echo    tabby-native.exe   BTK native terminal
echo    tabby-wails.exe    Wails GUI app
echo.
echo  Packages:
echo    pkg/agent        Agent integration
echo    pkg/ai           AI assistance
echo    pkg/api          API layer
echo    pkg/colorscheme  Color theming
echo    pkg/config       Configuration
echo    pkg/hotkey       Hotkey management
echo    pkg/pty          PTY process management
echo    pkg/session      Terminal sessions
echo    pkg/ssh          SSH client
echo    pkg/sftp         SFTP file transfer
echo    pkg/ui           Terminal UI
echo    pkg/vault        Credential vault
echo    pkg/vdom         Virtual DOM
echo    internal/server  Internal server
echo.
goto :end

:end
endlocal
