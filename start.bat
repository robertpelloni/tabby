@echo off
setlocal
title Tabby
cd /d "%~dp0"

echo [Tabby] Starting...
where npm >nul 2>nul
if errorlevel 1 (
    echo [Tabby] npm not found. Please install it.
    pause
    exit /b 1
)

npm start

if errorlevel 1 (
    echo [Tabby] Exited with error code %errorlevel%.
    pause
)
endlocal
