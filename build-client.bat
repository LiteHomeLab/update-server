@echo off
echo Building update-client...
cd cmd\update-client
go build -o ../../bin/update-client.exe .
if errorlevel 1 (
    echo Build failed
    cd ../..
    exit /b 1
)
cd ../..
echo Build successful: bin/update-client.exe
