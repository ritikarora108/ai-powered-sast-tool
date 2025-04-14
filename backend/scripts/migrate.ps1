# PowerShell script for managing database migrations with goose

param (
    [string]$command = "up",
    [string]$name = ""
)

$directory = "./db/migrations"

# Display help if requested
if ($command -eq "help") {
    Write-Host "Usage: ./scripts/migrate.ps1 [command] [name]"
    Write-Host "Commands:"
    Write-Host "  up     - Apply all migrations"
    Write-Host "  down   - Roll back the latest migration"
    Write-Host "  status - Show migration status"
    Write-Host "  reset  - Roll back all migrations"
    Write-Host "  create - Create a new migration file (requires -name parameter)"
    exit 0
}

# Check for valid command
$validCommands = @("up", "down", "status", "reset", "create")
if ($validCommands -notcontains $command) {
    Write-Host "Invalid command: $command"
    Write-Host "Available commands: up, down, reset, status, create"
    exit 1
}

# Check for required name parameter for create command
if ($command -eq "create" -and $name -eq "") {
    Write-Host "Migration name is required for create command"
    Write-Host "Usage: ./scripts/migrate.ps1 create -name your_migration_name"
    exit 1
}

# Set database URL
$databaseUrl = $env:DATABASE_URL
if ([string]::IsNullOrEmpty($databaseUrl)) {
    $databaseUrl = "postgres://postgres:postgres@localhost:5432/sast?sslmode=disable"
}

# Run migration command
try {
    switch ($command) {
        "up" {
            Write-Host "Applying migrations..."
            go run scripts/migrate.go -command up
        }
        "down" {
            Write-Host "Rolling back latest migration..."
            go run scripts/migrate.go -command down
        }
        "status" {
            Write-Host "Getting migration status..."
            go run scripts/migrate.go -command status
        }
        "reset" {
            $confirm = Read-Host "Are you sure you want to roll back all migrations? (y/n)"
            if ($confirm -eq "y") {
                Write-Host "Rolling back all migrations..."
                go run scripts/migrate.go -command reset
            } else {
                Write-Host "Operation canceled"
            }
        }
        "create" {
            Write-Host "Creating new migration: $name"
            go run scripts/migrate.go -command create -name $name
        }
    }
    Write-Host "Migration command '$command' executed successfully"
} catch {
    Write-Host "Error executing migration command: $_"
    exit 1
} 