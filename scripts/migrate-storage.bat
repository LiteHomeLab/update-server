@echo off
setlocal enabledelayedexpansion
REM Storage Migration Script for DocuFiller Update Server (Windows)
REM This script migrates existing package files to the new multi-program structure

echo Starting storage migration...
echo ================================
echo.

REM Define paths
set OLD_BASE=.\data\packages
set NEW_BASE=.\data\packages

REM Step 1: Create new directory structure
echo [Step 1] Creating new directory structure...
if not exist "%NEW_BASE%\docufiller\stable" mkdir "%NEW_BASE%\docufiller\stable"
if not exist "%NEW_BASE%\docufiller\beta" mkdir "%NEW_BASE%\docufiller\beta"
echo Created directories:
echo   - %NEW_BASE%\docufiller\stable
echo   - %NEW_BASE%\docufiller\beta
echo.

REM Step 2: Migrate stable versions
echo [Step 2] Migrating stable versions...
if exist "%OLD_BASE%\stable" (
    dir /b /a-d "%OLD_BASE%\stable" >nul 2>&1
    if %errorlevel% equ 0 (
        echo Found files in stable channel
        for %%F in ("%OLD_BASE%\stable\*") do (
            set filename=%%~nxF
            set version=%%~nF
            set target_dir=%NEW_BASE%\docufiller\stable\!version!
            if not exist "!target_dir!" mkdir "!target_dir!"
            move "%%F" "!target_dir!\" >nul
            echo   Migrated: %%~nxF -^> docufiller\stable\!version!\
        )
        rd "%OLD_BASE%\stable" 2>nul && echo   Removed old stable directory
        echo Done migrating stable versions
    ) else (
        echo   No files found in stable channel
    )
) else (
    echo   Old stable directory not found, skipping
)
echo.

REM Step 3: Migrate beta versions
echo [Step 3] Migrating beta versions...
if exist "%OLD_BASE%\beta" (
    dir /b /a-d "%OLD_BASE%\beta" >nul 2>&1
    if %errorlevel% equ 0 (
        echo Found files in beta channel
        for %%F in ("%OLD_BASE%\beta\*") do (
            set filename=%%~nxF
            set version=%%~nF
            set target_dir=%NEW_BASE%\docufiller\beta\!version!
            if not exist "!target_dir!" mkdir "!target_dir!"
            move "%%F" "!target_dir!\" >nul
            echo   Migrated: %%~nxF -^> docufiller\beta\!version!\
        )
        rd "%OLD_BASE%\beta" 2>nul && echo   Removed old beta directory
        echo Done migrating beta versions
    ) else (
        echo   No files found in beta channel
    )
) else (
    echo   Old beta directory not found, skipping
)
echo.

REM Summary
echo ================================
echo Storage migration completed!
echo.
echo New directory structure:
echo   data\packages\docufiller\stable\
echo   data\packages\docufiller\beta\
echo.
echo Next steps:
echo 1. Run: go run scripts\migrate.go (to migrate database)
echo 2. Verify the package files are in the correct locations
echo 3. Test the download endpoint
echo.

pause
