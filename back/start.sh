#!/bin/bash

# Load .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Start server
go run cmd/server/main.go
