#!/bin/bash

# Set up environment variables from .env file
if [ -f ../.env ]; then
    source ../.env
else
    echo "Error: .env file not found"
    exit 1
fi

# Create the database and initialize it
echo "Creating database..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -f init_db.sql

# Apply the schema
echo "Applying database schema..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ../db/schema/schema.sql

echo "Database setup complete!" 