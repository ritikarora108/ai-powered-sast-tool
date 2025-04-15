# Script to stop any running instances, install dependencies if needed, and run the frontend
try {
    # Try to find any running Next.js processes
    $nextProcesses = Get-Process -Name "node" -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -match "next" }
    
    if ($nextProcesses) {
        Write-Host "Stopping existing Next.js processes..."
        $nextProcesses | ForEach-Object { Stop-Process -Id $_.Id -Force }
    }
} catch {
    # Process might not be running, continue
    Write-Host "No Next.js processes found to stop or error accessing processes"
}

# Change to frontend directory
cd ./frontend

# Check if node_modules exists
if (-not (Test-Path -Path "node_modules")) {
    Write-Host "Installing dependencies..."
    npm install
}

# Run the frontend in development mode
Write-Host "Starting frontend in development mode..."
npm run dev 