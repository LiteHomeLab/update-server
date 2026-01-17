# Automated test script for build.bat

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "Automated build.bat Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get the script directory
$SCRIPT_DIR = $PSScriptRoot

# Clean bin directory
Remove-Item -Path "bin\*.exe" -Force -ErrorAction SilentlyContinue
Write-Host "Cleaned bin directory" -ForegroundColor Yellow
Write-Host ""

# Build docufiller-update-server
Write-Host "Building docufiller-update-server..." -ForegroundColor Yellow
cd $SCRIPT_DIR
go build -o "bin\docufiller-update-server.exe" .

if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] docufiller-update-server build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] docufiller-update-server.exe built" -ForegroundColor Green

# Build upload-admin
Write-Host ""
Write-Host "Building upload-admin..." -ForegroundColor Yellow
cd "$SCRIPT_DIR\clients\go\admin"
go mod tidy
go build -o "$SCRIPT_DIR\bin\upload-admin.exe" .

if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] upload-admin build failed with exit code $LASTEXITCODE" -ForegroundColor Red
    cd $SCRIPT_DIR
    exit 1
}
Write-Host "[OK] upload-admin.exe built" -ForegroundColor Green

cd $SCRIPT_DIR

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build SUCCEEDED!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Bin directory contents:" -ForegroundColor Yellow
Get-ChildItem -Path "bin\" -Filter "*.exe" | Format-Table Name, Length -AutoSize
Write-Host ""
