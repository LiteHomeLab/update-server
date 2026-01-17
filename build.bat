@echo off
setlocal

REM Get the absolute path of the directory where this batch file is located
set SCRIPT_DIR=%~dp0
REM Remove trailing backslash
set SCRIPT_DIR=%SCRIPT_DIR:~0,-1%

echo ========================================
echo DocuFiller Build Script
echo ========================================
echo.

REM Create bin directory if not exists
if not exist "%SCRIPT_DIR%\bin" (
    echo Creating bin directory...
    mkdir "%SCRIPT_DIR%\bin"
)

echo Building docufiller-update-server...
cd /d "%SCRIPT_DIR%"
go build -o bin\docufiller-update-server.exe .

if errorlevel 1 (
    echo.
    echo ========================================
    echo Build failed for docufiller-update-server
    echo ========================================
    pause
    exit /b 1
)

echo Building upload-admin...
cd /d "%SCRIPT_DIR%\clients\go\admin"
go mod tidy
go build -o "%SCRIPT_DIR%\bin\upload-admin.exe" .

if errorlevel 1 (
    echo.
    echo ========================================
    echo Build failed for upload-admin
    echo ========================================
    pause
    exit /b 1
)

echo.
echo ========================================
echo Build succeeded!
echo Output: %SCRIPT_DIR%\bin\docufiller-update-server.exe
echo Output: %SCRIPT_DIR%\bin\upload-admin.exe
echo ========================================
echo.

pause
