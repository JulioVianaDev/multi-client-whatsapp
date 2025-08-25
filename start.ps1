Write-Host "Starting WhatsApp Bridge Project..." -ForegroundColor Green
Write-Host ""

Write-Host "[1/3] Installing Go dependencies..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to install Go dependencies" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host "[2/3] Installing Node.js dependencies..." -ForegroundColor Yellow
Set-Location nodejs-project
npm install
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to install Node.js dependencies" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
Set-Location ..

Write-Host "[3/3] Starting both services..." -ForegroundColor Yellow
Write-Host ""

Write-Host "Starting Go WhatsApp Bridge on port 4444..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run main.go" -WindowStyle Normal

Write-Host "Starting Node.js Webhook Receiver on port 5555..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd nodejs-project; npm start" -WindowStyle Normal

Write-Host ""
Write-Host "Both services are starting..." -ForegroundColor Green
Write-Host ""
Write-Host "Go WhatsApp Bridge: http://localhost:4444" -ForegroundColor White
Write-Host "Node.js Webhook Receiver: http://localhost:5555" -ForegroundColor White
Write-Host ""
Write-Host "Instructions:" -ForegroundColor Yellow
Write-Host "1. Wait for both services to start" -ForegroundColor White
Write-Host "2. Visit http://localhost:4444/qr to get QR code" -ForegroundColor White
Write-Host "3. Scan QR code with WhatsApp mobile app" -ForegroundColor White
Write-Host "4. Visit http://localhost:5555/scan for instructions" -ForegroundColor White
Write-Host "5. Monitor events in the Node.js console" -ForegroundColor White
Write-Host ""
Read-Host "Press Enter to exit"
