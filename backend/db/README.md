# Database Package

This package provides database access and SQL query functionality for the AI-powered SAST tool.

## Structure

- `db.go` - Main database connection and query interface
- `migrations/` - Database migration files
- `query/` - SQL query files for SQLC
- `schema/` - Database schema definitions
- `sqlc/` - Generated code from SQLC (not committed to git)

## Database Setup

1. Install PostgreSQL if you haven't already
2. Create a database:
   ```
   psql -U postgres -c "CREATE DATABASE sast_tool;"
   ```
3. Run the migration scripts:
   ```
   psql -U postgres -d sast_tool -f migrations/001_init.sql
   ```

## Generating Code with SQLC

1. Install SQLC:
   ```
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```
2. Generate code:
   ```
   cd backend
   sqlc generate
   ```

## Environment Variables

The following environment variables are used for database connection:

- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - PostgreSQL user (default: postgres)
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - PostgreSQL database name (default: sast_tool)

These can be configured in the `.env` file.
