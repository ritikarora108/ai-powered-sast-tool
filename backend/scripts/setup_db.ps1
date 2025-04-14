# PowerShell script to set up the database

# Read environment variables from .env file
if (Test-Path "../.env") {
    Get-Content "../.env" | ForEach-Object {
        if (!$_.StartsWith("#") -and $_.Contains("=")) {
            $key, $value = $_ -split "=", 2
            [Environment]::SetEnvironmentVariable($key, $value)
        }
    }
}
else {
    Write-Error "Error: .env file not found"
    exit 1
}

$DB_HOST = $env:DB_HOST
$DB_PORT = $env:DB_PORT
$DB_USER = $env:DB_USER
$DB_PASSWORD = $env:DB_PASSWORD
$DB_NAME = $env:DB_NAME

# Create the database and initialize it
Write-Output "Creating database..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -f init_db.sql

# Apply the schema
Write-Output "Applying database schema..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ../db/schema/schema.sql

Write-Output "Database setup complete!" 