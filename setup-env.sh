#!/bin/bash

# Setup environment variables for the e-commerce Go application

echo "Setting up environment variables for e-commerce Go application..."

# Create .env file from example
if [ ! -f .env ]; then
    cp env.example .env
    echo "Created .env file from env.example"
else
    echo ".env file already exists"
fi

# Export environment variables for current session
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=test
export DB_PASSWORD=test
export DB_NAME=postgres
export DB_SSLMODE=disable
export PORT=8080
export GIN_MODE=debug
export JWT_SECRET=AANmjnkH0q6f5g9KASpjq6r6Arj0OHHnNhdYPrChFvX8pf1okVESyCPrex9SMBQcLBPNEEvMdGDLyiMPhVkcXg==

echo "Environment variables set for current session:"
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_USER: $DB_USER"
echo "DB_NAME: $DB_NAME"
echo "DB_SSLMODE: $DB_SSLMODE"
echo "PORT: $PORT"
echo "GIN_MODE: $GIN_MODE"
echo "JWT_SECRET: [HIDDEN]"

echo ""
echo "To make these permanent, add them to your shell profile (.bashrc, .zshrc, etc.)"
echo "Or use the .env file that was created."
