package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const directory = "./db/migrations"

func main() {
	flags := flag.NewFlagSet("migrate", flag.ExitOnError)

	// Define the command to run
	cmdFlag := flags.String("command", "up", "Migration command (up, down, status, create)")

	// Define the migration name (for create command)
	nameFlag := flags.String("name", "", "Migration name (for create command)")

	// Parse the command line arguments
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/sast?sslmode=disable"
	}

	db, err := goose.OpenDBWithDriver("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	command := *cmdFlag

	switch command {
	case "up":
		if err := goose.Up(db, directory); err != nil {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
	case "down":
		if err := goose.Down(db, directory); err != nil {
			log.Fatalf("Failed to roll back migration: %v", err)
		}
	case "reset":
		if err := goose.Reset(db, directory); err != nil {
			log.Fatalf("Failed to reset migrations: %v", err)
		}
	case "status":
		if err := goose.Status(db, directory); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
	case "create":
		name := *nameFlag
		if name == "" {
			log.Fatalf("Migration name is required for create command")
		}
		if err := goose.Create(db, directory, name, "sql"); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
	default:
		fmt.Printf("Invalid command: %s\n", command)
		fmt.Println("Available commands: up, down, reset, status, create")
		os.Exit(1)
	}

	fmt.Printf("Migration command '%s' executed successfully\n", command)
}
