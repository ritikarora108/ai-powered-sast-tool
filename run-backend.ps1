# Script to stop, build and run backend
try {
    taskkill /F /IM backend.exe 2>$null
} catch {
    # Process might not be running, continue
}

# Change to backend directory
cd ./backend

# Build the backend
Write-Host "Building backend..."
go build -o backend.exe

# Run if build was successful
if ($LASTEXITCODE -eq 0) {
    Write-Host "Starting backend..."
    ./backend.exe
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
}
