#!/bin/bash

# Simple usage demo
echo "=== MCloud Init Demo ==="
echo

# Build
echo "Building..."
go build -o mcloudd ./cmd/mcloudd
go build -o mcloudctl ./cmd/mcloudctl

# Start server in background
echo "Starting mcloudd server..."
./mcloudd > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2

# Initialize cluster - now just need --name!
echo
echo "Initializing cluster..."
./mcloudctl init --name my-cluster

# Cleanup
echo
kill $SERVER_PID 2>/dev/null || true
echo "Done!"
