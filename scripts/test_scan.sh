#!/bin/bash

# Exit on error
set -e

# Default variables
API_URL="http://localhost:8080"
OWNER="OWASP"
REPO="NodeGoat"
WAIT_TIME=10

# Parse command line arguments
while getopts a:o:r:w: flag
do
    case "${flag}" in
        a) API_URL=${OPTARG};;
        o) OWNER=${OPTARG};;
        r) REPO=${OPTARG};;
        w) WAIT_TIME=${OPTARG};;
    esac
done

echo "Testing repository scan..."
echo "API_URL: $API_URL"
echo "OWNER: $OWNER"
echo "REPO: $REPO"

# Check if curl is installed
if ! command -v curl &> /dev/null
then
    echo "curl not found. Please install it and try again."
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq not found. Please install it and try again."
    exit 1
fi

# Create repository
echo "Creating repository..."
REPO_RESPONSE=$(curl -s -X POST "$API_URL/api/repositories" \
    -H "Content-Type: application/json" \
    -d "{\"owner\":\"$OWNER\",\"name\":\"$REPO\",\"url\":\"https://github.com/$OWNER/$REPO\"}")

# Get repository ID
REPO_ID=$(echo $REPO_RESPONSE | jq -r '.id')

if [ "$REPO_ID" == "null" ] || [ -z "$REPO_ID" ]; then
    echo "Failed to create repository. Response: $REPO_RESPONSE"
    exit 1
fi

echo "Repository created with ID: $REPO_ID"

# Start scan
echo "Starting scan..."
SCAN_RESPONSE=$(curl -s -X POST "$API_URL/api/repositories/$REPO_ID/scan" \
    -H "Content-Type: application/json")

# Get scan status
SCAN_STATUS=$(echo $SCAN_RESPONSE | jq -r '.status')

if [ "$SCAN_STATUS" != "scan_initiated" ]; then
    echo "Failed to start scan. Response: $SCAN_RESPONSE"
    exit 1
fi

echo "Scan initiated successfully."
echo "Waiting for $WAIT_TIME seconds for the scan to progress..."
sleep $WAIT_TIME

# Get vulnerabilities
echo "Fetching vulnerabilities..."
VULNS_RESPONSE=$(curl -s -X GET "$API_URL/api/repositories/$REPO_ID/vulnerabilities" \
    -H "Content-Type: application/json")

echo "Scan results:"
echo $VULNS_RESPONSE | jq '.'

echo "Test scan complete!" 