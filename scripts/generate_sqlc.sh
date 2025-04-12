#!/bin/bash

# Exit on error
set -e

# Default variables
SQLC_CONFIG="../backend/db/sqlc.yaml"
SQLC_VERSION="v1.25.0"
WORKING_DIR="../backend"

# Parse command line arguments
while getopts c:v:d: flag
do
    case "${flag}" in
        c) SQLC_CONFIG=${OPTARG};;
        v) SQLC_VERSION=${OPTARG};;
        d) WORKING_DIR=${OPTARG};;
    esac
done

echo "Generating SQLC code..."
echo "SQLC_CONFIG: $SQLC_CONFIG"
echo "SQLC_VERSION: $SQLC_VERSION"
echo "WORKING_DIR: $WORKING_DIR"

# Check if docker is installed
if ! command -v docker &> /dev/null
then
    echo "Docker not found. Please install it and try again."
    exit 1
fi

# Check if config file exists
if [ ! -f "$SQLC_CONFIG" ]; then
    echo "SQLC config file not found: $SQLC_CONFIG"
    exit 1
fi

# Run sqlc in docker
echo "Running sqlc..."
cd $WORKING_DIR && docker run --rm -v $(pwd):/src -w /src kjconroy/sqlc:$SQLC_VERSION generate

echo "SQLC code generation complete!" 