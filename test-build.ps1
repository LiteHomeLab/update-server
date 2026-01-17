# Test script to verify build.bat works correctly

Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "Testing build.bat" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Clean bin directory
Remove-Item -Path "bin\*.exe" -Force -ErrorAction SilentlyContinue
Write-Host "Cleaned bin directory" -ForegroundColor Yellow

# Run build.bat (with auto-confirmation for pause)
Write-Host ""
Write-Host "Running build.bat..." -ForegroundColor Yellow
$process = Start-Process -FilePath "cmd" -ArgumentList "/c build.bat" -Wait -PassThru -NoNewWindow

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build complete. Exit code: $($process.ExitCode)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Check results
Write-Host ""
Write-Host "Bin directory contents:" -ForegroundColor Yellow
Get-ChildItem -Path "bin\" -Filter "*.exe" | Format-Table Name, Length -AutoSize

if (Test-Path "bin\docufiller-update-server.exe") {
    Write-Host "[SUCCESS] docufiller-update-server.exe built" -ForegroundColor Green
} else {
    Write-Host "[FAIL] docufiller-update-server.exe NOT found" -ForegroundColor Red
}

if (Test-Path "bin\upload-admin.exe") {
    Write-Host "[SUCCESS] upload-admin.exe built" -ForegroundColor Green
} else {
    Write-Host "[FAIL] upload-admin.exe NOT found" -ForegroundColor Red
}

Write-Host ""
Read-Host "Press Enter to exit"
