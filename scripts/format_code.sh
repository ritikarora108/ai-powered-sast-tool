#!/bin/bash

# Format Go files
echo "Formatting Go files..."
find ./backend -name "*.go" -type f -exec gofmt -w -s {} \;

# Format TypeScript/JavaScript files
echo "Formatting TypeScript/JavaScript files..."
if command -v npx &> /dev/null; then
    cd frontend
    if grep -q "prettier" package.json; then
        npx prettier --write "**/*.{ts,tsx,js,jsx,json}"
    else
        echo "Prettier not found in package.json, attempting to install..."
        npm install --save-dev prettier
        npx prettier --write "**/*.{ts,tsx,js,jsx,json}"
    fi
    cd ..
else
    echo "npx not found, skipping TypeScript/JavaScript formatting"
fi

echo "Formatting complete!" 