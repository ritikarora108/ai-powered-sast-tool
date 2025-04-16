# Format Go files
Write-Host "Formatting Go files..." -ForegroundColor Green
Get-ChildItem -Path .\backend -Filter "*.go" -Recurse | ForEach-Object {
    Write-Host "Formatting $($_.FullName)"
    go fmt $_.FullName
}

# Format TypeScript/JavaScript files
Write-Host "Formatting TypeScript/JavaScript files..." -ForegroundColor Green
if (Test-Path -Path .\frontend\package.json) {
    Push-Location -Path .\frontend
    if (Select-String -Path package.json -Pattern "prettier" -Quiet) {
        Write-Host "Running Prettier..."
        npx prettier --write "**/*.{ts,tsx,js,jsx,json}"
    } else {
        Write-Host "Prettier not found in package.json, attempting to install..."
        npm install --save-dev prettier
        npx prettier --write "**/*.{ts,tsx,js,jsx,json}"
    }
    Pop-Location
} else {
    Write-Host "Frontend package.json not found, skipping TypeScript/JavaScript formatting" -ForegroundColor Yellow
}

Write-Host "Formatting complete!" -ForegroundColor Green
