@echo off
REM Update Server All-in-One Build Script
REM This script builds the server and all clients in one go.

setlocal enabledelayedexpansion

echo ========================================
echo   Update Server All-in-One Builder
echo ========================================
echo.

REM Get the script directory
set "SCRIPT_DIR=%~dp0"
set "OUTPUT_DIR=%SCRIPT_DIR%bin"
set "CLIENT_OUTPUT_DIR=%OUTPUT_DIR%\clients"

echo [1/5] Creating output directories...
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"
if not exist "%CLIENT_OUTPUT_DIR%" mkdir "%CLIENT_OUTPUT_DIR%"
echo Created: %OUTPUT_DIR%
echo Created: %CLIENT_OUTPUT_DIR%
echo.

echo [2/5] Building Update Server...
cd /d "%SCRIPT_DIR%"
go build -o "%OUTPUT_DIR%\update-server.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-server.exe
    goto :error
)
echo SUCCESS: Built update-server.exe
echo.

echo [3/5] Building Publish Client (update-admin)...
cd /d "%SCRIPT_DIR%clients\go\admin"
go build -o "%CLIENT_OUTPUT_DIR%\update-admin.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-admin.exe
    goto :error
)
echo SUCCESS: Built update-admin.exe
echo.

echo [4/5] Building Update Client (update-client)...
cd /d "%SCRIPT_DIR%cmd\update-client"
go build -o "%CLIENT_OUTPUT_DIR%\update-client.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-client.exe
    goto :error
)
echo SUCCESS: Built update-client.exe
echo.

echo [5/5] Copying client executables to server data directory...
set "SERVER_CLIENT_DIR=%SCRIPT_DIR%data\clients"
if not exist "%SERVER_CLIENT_DIR%" mkdir "%SERVER_CLIENT_DIR%"

REM Copy publish client
copy /Y "%CLIENT_OUTPUT_DIR%\update-admin.exe" "%SERVER_CLIENT_DIR%\publish-client.exe" >nul
echo Copied: publish-client.exe -^> %SERVER_CLIENT_DIR%

REM Copy update client
copy /Y "%CLIENT_OUTPUT_DIR%\update-client.exe" "%SERVER_CLIENT_DIR%\update-client.exe" >nul
echo Copied: update-client.exe -^> %SERVER_CLIENT_DIR%

echo.
echo ========================================
echo   Build Completed Successfully!
echo ========================================
echo.
echo Output files:
echo   - Server: %OUTPUT_DIR%\update-server.exe
echo   - Publish Client: %CLIENT_OUTPUT_DIR%\update-admin.exe
echo   - Update Client: %CLIENT_OUTPUT_DIR%\update-client.exe
echo.
echo Server deployment directory:
echo   - Clients: %SERVER_CLIENT_DIR%\
echo.
echo To deploy:
echo   1. Copy update-server.exe to your server
echo   2. Create 'data\clients' directory on server
echo   3. Copy publish-client.exe to 'data\clients' directory
echo   4. Copy update-client.exe to 'data\clients' directory
echo   5. Run update-server.exe
echo.
goto :end

:error
echo.
echo ========================================
echo   Build Failed!
echo ========================================
exit /b 1

:end
endlocal
