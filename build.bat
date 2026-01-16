@echo off
REM DocuFiller Update Server Build Script for Windows

echo ========================================
echo DocuFiller Update Server Build Script
echo ========================================
echo.

REM Check if bin directory exists, create if not
if not exist "bin" (
    echo Creating bin directory...
    mkdir bin
)

echo Building project...
go build -o bin/docufiller-update-server.exe .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Build succeeded!
    echo Output: bin/docufiller-update-server.exe
    echo ========================================
) else (
    echo.
    echo ========================================
    echo Build failed with error code: %ERRORLEVEL%
    echo ========================================
    exit /b %ERRORLEVEL%
)

echo.
