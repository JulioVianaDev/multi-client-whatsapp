#!/bin/bash

echo "Setting up environment for local development..."

# Set the environment variable
export EASY_ZAP_WEBHOOK_URL=http://localhost:4444

echo "Environment variable set: EASY_ZAP_WEBHOOK_URL=$EASY_ZAP_WEBHOOK_URL"

# Download dependencies
echo "Downloading Go dependencies..."
go mod tidy

# Build the application
echo "Building the application..."
go build -o main

# Run the application
echo "Starting the application..."
./main
