#!/bin/bash

# Exit on error
set -e

# Default variables
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_NAME="sast"
SCHEMA_FILE="../docs/schema.sql"

# Parse command line arguments
while getopts h:p:u:w:d:f: flag
do
    case "${flag}" in
        h) DB_HOST=${OPTARG};;
        p) DB_PORT=${OPTARG};;
        u) DB_USER=${OPTARG};;
        w) DB_PASSWORD=${OPTARG};;
        d) DB_NAME=${OPTARG};;
        f) SCHEMA_FILE=${OPTARG};;
    esac
done

echo "Setting up database..."
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_USER: $DB_USER"
echo "DB_NAME: $DB_NAME"
echo "SCHEMA_FILE: $SCHEMA_FILE"

# Check if psql is installed
if ! command -v psql &> /dev/null
then
    echo "PostgreSQL client (psql) not found. Please install it and try again."
    exit 1
fi

# Check if schema file exists
if [ ! -f "$SCHEMA_FILE" ]; then
    echo "Schema file not found: $SCHEMA_FILE"
    exit 1
fi

# Create database if it doesn't exist
export PGPASSWORD=$DB_PASSWORD
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo "Database $DB_NAME already exists."
else
    echo "Creating database $DB_NAME..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;"
    echo "Database created."
fi

# Apply schema
echo "Applying schema..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $SCHEMA_FILE

echo "Database setup complete!" 