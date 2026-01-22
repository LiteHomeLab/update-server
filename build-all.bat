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
set "SERVER_CLIENT_DIR=%SCRIPT_DIR%data\clients"

echo [1/6] Cleaning output directories...
REM Clean output directories to avoid cache issues
if exist "%OUTPUT_DIR%" (
    echo Removing: %OUTPUT_DIR%
    rmdir /s /q "%OUTPUT_DIR%" 2>nul
)
if exist "%SERVER_CLIENT_DIR%" (
    echo Removing: %SERVER_CLIENT_DIR%
    rmdir /s /q "%SERVER_CLIENT_DIR%" 2>nul
)

echo Creating fresh output directories...
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"
if not exist "%CLIENT_OUTPUT_DIR%" mkdir "%CLIENT_OUTPUT_DIR%"
echo Created: %OUTPUT_DIR%
echo Created: %CLIENT_OUTPUT_DIR%
echo.

echo [2/6] Building Update Server...
cd /d "%SCRIPT_DIR%cmd\update-server"
go build -o "%OUTPUT_DIR%\update-server.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-server.exe
    goto :error
)
echo SUCCESS: Built update-server.exe
echo.

echo [3/6] Building Update Publisher...
cd /d "%SCRIPT_DIR%cmd\update-publisher"
go build -o "%CLIENT_OUTPUT_DIR%\update-publisher.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-publisher.exe
    goto :error
)
echo SUCCESS: Built update-publisher.exe
echo.

echo [4/6] Building Update Client...
cd /d "%SCRIPT_DIR%cmd\update-client"
go build -o "%CLIENT_OUTPUT_DIR%\update-client.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-client.exe
    goto :error
)
echo SUCCESS: Built update-client.exe
echo.

echo [5/6] Copying client executables to server data directory...
if not exist "%SERVER_CLIENT_DIR%" mkdir "%SERVER_CLIENT_DIR%"

REM Copy update publisher
copy /Y "%CLIENT_OUTPUT_DIR%\update-publisher.exe" "%SERVER_CLIENT_DIR%\update-publisher.exe"
if errorlevel 1 (
    echo WARNING: Standard copy failed, using robocopy fallback...
    robocopy "%CLIENT_OUTPUT_DIR%" "%SERVER_CLIENT_DIR%" "update-publisher.exe" /noclobber /njh /njs /ndl /nc /ns
)
echo Copied: update-publisher.exe -^> %SERVER_CLIENT_DIR%

REM Copy update client
copy /Y "%CLIENT_OUTPUT_DIR%\update-client.exe" "%SERVER_CLIENT_DIR%\update-client.exe"
if errorlevel 1 (
    echo WARNING: Standard copy failed, using robocopy fallback...
    robocopy "%CLIENT_OUTPUT_DIR%" "%SERVER_CLIENT_DIR%" "update-client.exe" /noclobber /njh /njs /ndl /nc /ns
)
echo Copied: update-client.exe -^> %SERVER_CLIENT_DIR%

echo.
echo [6/6] Copying client configuration files...
REM Copy update-publisher usage guide
copy /Y "%SCRIPT_DIR%cmd\update-publisher\update-publisher.usage.txt" "%CLIENT_OUTPUT_DIR%\update-publisher.usage.txt"
if errorlevel 1 (
    robocopy "%SCRIPT_DIR%cmd\update-publisher" "%CLIENT_OUTPUT_DIR%" "update-publisher.usage.txt" /noclobber /njh /njs /ndl /nc /ns
)
echo Copied: update-publisher.usage.txt -^> %CLIENT_OUTPUT_DIR%

REM Copy update-client config template
copy /Y "%SCRIPT_DIR%cmd\update-client\update-client.config.yaml" "%CLIENT_OUTPUT_DIR%\update-client.config.yaml"
if errorlevel 1 (
    robocopy "%SCRIPT_DIR%cmd\update-client" "%CLIENT_OUTPUT_DIR%" "update-client.config.yaml" /noclobber /njh /njs /ndl /nc /ns
)
echo Copied: update-client.config.yaml -^> %CLIENT_OUTPUT_DIR%

REM Copy to server deployment directory
copy /Y "%CLIENT_OUTPUT_DIR%\update-publisher.usage.txt" "%SERVER_CLIENT_DIR%\update-publisher.usage.txt"
if errorlevel 1 (
    robocopy "%CLIENT_OUTPUT_DIR%" "%SERVER_CLIENT_DIR%" "update-publisher.usage.txt" /noclobber /njh /njs /ndl /nc /ns
)

copy /Y "%CLIENT_OUTPUT_DIR%\update-client.config.yaml" "%SERVER_CLIENT_DIR%\update-client.config.yaml"
if errorlevel 1 (
    robocopy "%CLIENT_OUTPUT_DIR%" "%SERVER_CLIENT_DIR%" "update-client.config.yaml" /noclobber /njh /njs /ndl /nc /ns
)
echo.
echo Configuration files copied to server deployment directory.

echo.
echo ========================================
echo   All Build Steps Completed Successfully
echo ========================================
goto :end

:error
echo.
echo ========================================
echo   Build Failed!
echo ========================================
echo.
pause

:end
endlocal
pause
