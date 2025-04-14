//go:build ignore
// +build ignore

package main

/*
This file is not meant to be compiled.
It serves as documentation for how to run sqlc to generate the database code.

To generate the database code, run:
  cd backend
  sqlc generate

Make sure you have sqlc installed:
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

After running sqlc generate, the generated code will be in the backend/db/sqlc directory.
This includes strongly typed functions matching the SQL queries defined in the backend/db/query directory.

The types of queries are:
- :one - Returns a single row
- :many - Returns multiple rows
- :exec - Executes a query without returning rows
- :batchexec - Like :exec but for batch operations
- :copyfrom - Used for PostgreSQL's COPY command

For more information, visit https://docs.sqlc.dev/
*/

func main() {
	// This function is intentionally empty
}
