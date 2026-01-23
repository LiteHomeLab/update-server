@echo off
REM Test runner script for update-server
REM This script runs all tests with appropriate configuration

setlocal enabledelayedexpansion

echo ========================================
echo Update Server Test Runner
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if errorlevel 1 (
    echo ERROR: Go is not installed or not in PATH
    exit /b 1
)

REM Parse command line arguments
set TEST_TYPE=all
set VERBOSE=0

:parse_args
if "%~1"=="" goto done_parsing
if /i "%~1"=="unit" set TEST_TYPE=unit
if /i "%~1"=="integration" set TEST_TYPE=integration
if /i "%~1"=="e2e" set TEST_TYPE=e2e
if /i "%~1"=="all" set TEST_TYPE=all
if /i "%~1"=="-v" set VERBOSE=1
if /i "%~1"=="--verbose" set VERBOSE=1
shift
goto parse_args
:done_parsing

echo Running tests: %TEST_TYPE%
echo.

REM Set common test flags
set TEST_FLAGS=-timeout=5m
if %VERBOSE%==1 set TEST_FLAGS=%TEST_FLAGS% -v

REM Run tests based on type
if /i "%TEST_TYPE%"=="unit" goto run_unit
if /i "%TEST_TYPE%"=="integration" goto run_integration
if /i "%TEST_TYPE%"=="e2e" goto run_e2e
goto run_all

:run_unit
echo [Running Unit Tests]
echo ========================================
go test ./... -run "^Test[^E]" %TEST_FLAGS%
goto end

:run_integration
echo [Running Integration Tests]
echo ========================================
go test ./tests/integration/... %TEST_FLAGS%
goto end

:run_e2e
echo [Running E2E Tests]
echo ========================================
echo NOTE: E2E tests require Playwright drivers to be installed.
echo Install them with: go install github.com/playwright-community/playwright-go/cmd/playwright@latest
echo.
go test ./tests/e2e/... %TEST_FLAGS%
goto end

:run_all
echo [Running All Tests]
echo ========================================
echo.
echo [1/3] Unit Tests
echo ----------------------------------------
go test ./... -run "^Test[^E]" %TEST_FLAGS%
if errorlevel 1 (
    echo Unit tests failed!
    exit /b 1
)

echo.
echo [2/3] Integration Tests
echo ----------------------------------------
go test ./tests/integration/... %TEST_FLAGS%
if errorlevel 1 (
    echo Integration tests failed!
    exit /b 1
)

echo.
echo [3/3] E2E Tests
echo ----------------------------------------
echo NOTE: E2E tests require Playwright drivers.
go test ./tests/e2e/... %TEST_FLAGS%

echo.
echo ========================================
echo All tests completed!
echo ========================================
goto end

:end
endlocal
