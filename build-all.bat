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
set "SERVER_OUTPUT=%OUTPUT_DIR%\update-server.exe"
set "CLIENT_DIR=%OUTPUT_DIR%\data\clients"

echo [1/5] Cleaning output directory...
REM Kill any running server process first
taskkill /F /IM update-server.exe >nul 2>&1
taskkill /F /IM docufiller-update-server.exe >nul 2>&1

REM Clean output directory to avoid cache issues
if exist "%OUTPUT_DIR%" (
    echo Removing: %OUTPUT_DIR%
    rmdir /s /q "%OUTPUT_DIR%"
    if errorlevel 1 (
        echo WARNING: Failed to remove %OUTPUT_DIR%
        echo Files may be in use. Trying alternative cleanup...
        timeout /t 1 /nobreak >nul
        rmdir /s /q "%OUTPUT_DIR%" 2>nul
    )
)

echo Creating fresh output directory...
if exist "%OUTPUT_DIR%" (
    echo WARNING: Output directory still exists, may contain old files
)
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"
if not exist "%CLIENT_DIR%" mkdir "%CLIENT_DIR%"
echo Created: %OUTPUT_DIR%
echo Created: %CLIENT_DIR%
echo.

echo [2/5] Building Update Server...
cd /d "%SCRIPT_DIR%cmd\update-server"
go build -o "%SERVER_OUTPUT%" .
if errorlevel 1 (
    echo ERROR: Failed to build update-server.exe
    goto :error
)
echo SUCCESS: Built update-server.exe
echo.

echo [3/5] Building Publish Client...
cd /d "%SCRIPT_DIR%cmd\update-publisher"
go build -o "%CLIENT_DIR%\publish-client.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build publish-client.exe
    goto :error
)
echo SUCCESS: Built publish-client.exe
echo.

echo [4/5] Building Update Client...
cd /d "%SCRIPT_DIR%cmd\update-client"
go build -o "%CLIENT_DIR%\update-client.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build update-client.exe
    goto :error
)
echo SUCCESS: Built update-client.exe
echo.

echo [5/5] Copying configuration files...
REM Copy publish-client usage guide
copy /Y "%SCRIPT_DIR%cmd\update-publisher\publish-client.usage.txt" "%CLIENT_DIR%\publish-client.usage.txt" >nul
if errorlevel 1 (
    robocopy "%SCRIPT_DIR%cmd\update-publisher" "%CLIENT_DIR%" "publish-client.usage.txt" /noclobber /njh /njs /ndl /nc /ns >nul
)
echo Copied: publish-client.usage.txt

REM Copy update-client config template
copy /Y "%SCRIPT_DIR%cmd\update-client\update-client.config.yaml" "%CLIENT_DIR%\update-client.config.yaml" >nul
if errorlevel 1 (
    robocopy "%SCRIPT_DIR%cmd\update-client" "%CLIENT_DIR%" "update-client.config.yaml" /noclobber /njh /njs /ndl /nc /ns >nul
)
echo Copied: update-client.config.yaml

REM Copy server config template
copy /Y "%SCRIPT_DIR%config.yaml" "%OUTPUT_DIR%\config.yaml" >nul
if errorlevel 1 (
    robocopy "%SCRIPT_DIR%" "%OUTPUT_DIR%" "config.yaml" /noclobber /njh /njs /ndl /nc /ns >nul
)
echo Copied: config.yaml
echo.

echo ========================================
echo   Build Completed Successfully
echo ========================================
echo.
echo Deployment directory structure:
echo.
echo   bin\
echo     +-- update-server.exe
echo     +-- config.yaml
echo     +-- data\
echo         +-- clients\
echo             +-- publish-client.exe
echo             +-- publish-client.usage.txt
echo             +-- update-client.exe
echo             +-- update-client.config.yaml
echo.
echo You can now copy the entire 'bin' folder to your server.
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
