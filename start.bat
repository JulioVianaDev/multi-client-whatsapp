@echo off
echo Starting WhatsApp Bridge Project...
echo.

echo [1/3] Installing Go dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo Error: Failed to install Go dependencies
    pause
    exit /b 1
)

echo [2/3] Installing Node.js dependencies...
cd nodejs-project
npm install
if %errorlevel% neq 0 (
    echo Error: Failed to install Node.js dependencies
    pause
    exit /b 1
)
cd ..

echo [3/3] Starting both services...
echo.
echo Starting Go WhatsApp Bridge on port 4444...
start "Go WhatsApp Bridge" cmd /k "go run main.go"

echo Starting Node.js Webhook Receiver on port 5555...
start "Node.js Webhook Receiver" cmd /k "cd nodejs-project && npm start"

echo.
echo Both services are starting...
echo.
echo Go WhatsApp Bridge: http://localhost:4444
echo Node.js Webhook Receiver: http://localhost:5555
echo.
echo Instructions:
echo 1. Wait for both services to start
echo 2. Visit http://localhost:4444/qr to get QR code
echo 3. Scan QR code with WhatsApp mobile app
echo 4. Visit http://localhost:5555/scan for instructions
echo 5. Monitor events in the Node.js console
echo.
pause
