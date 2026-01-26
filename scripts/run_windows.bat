@echo off
REM Primortal - Windows Launch Helper
REM This script unblocks the game files to bypass SmartScreen warnings

setlocal
cd /d "%~dp0"

echo Primortal - Windows Launch Helper
echo ==================================
echo.

REM Check if binary exists
if not exist "primortal.exe" (
    echo Error: primortal.exe not found in current directory
    pause
    exit /b 1
)

echo Unblocking game files...
powershell -Command "Get-ChildItem -Path '%~dp0' -Recurse | Unblock-File"

if %ERRORLEVEL% EQU 0 (
    echo Successfully unblocked files
    echo.
    echo Launching Primortal...
    echo.
    start "" "primortal.exe"
) else (
    echo Warning: Failed to unblock files automatically
    echo You may need to:
    echo   1. Right-click primortal.exe
    echo   2. Select Properties
    echo   3. Check "Unblock" and click OK
    echo.
    pause
    echo Launching anyway...
    start "" "primortal.exe"
)
