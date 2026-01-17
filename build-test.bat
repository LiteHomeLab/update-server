@echo off
setlocal

set SCRIPT_DIR=%~dp0

echo ========================================
echo DocuFiller Build Script (TEST)
echo ========================================
echo.

if not exist "%SCRIPT_DIR%bin" mkdir "%SCRIPT_DIR%bin"

echo Building docufiller-update-server...
cd /d "%SCRIPT_DIR%"
go build -o bin\docufiller-update-server.exe .

if errorlevel 1 (
    echo Build failed for docufiller-update-server
    exit /b 1
)

echo Building upload-admin...
cd /d "%SCRIPT_DIR%clients\go\admin"
go mod tidy
go build -o "%SCRIPT_DIR%bin\upload-admin.exe" .

if errorlevel 1 (
    echo Build failed for upload-admin
    exit /b 1
)

echo.
echo ========================================
echo Build succeeded!
echo Output: bin\docufiller-update-server.exe
echo Output: bin\upload-admin.exe
echo ========================================
