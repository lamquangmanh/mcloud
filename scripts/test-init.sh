#!/bin/bash

# MCloud Init Test Script
set -e

echo "=== MCloud Init Feature Test ==="
echo

# Kill any existing mcloudd processes
echo "0. Stopping any existing mcloudd processes..."
pkill -f mcloudd || true
sleep 1
echo "✓ Existing processes stopped"
echo

# Clean up any existing database and WAL files
echo "1. Cleaning up existing database files..."
rm -f mcloud.db mcloud.db-shm mcloud.db-wal
echo "✓ Database files cleaned"
echo

# Build the binaries
echo "2. Building binaries..."
go build -o mcloudd ./cmd/mcloudd
go build -o mcloudctl ./cmd/mcloudctl
echo "✓ Binaries built successfully"
echo

# Start the mcloudd server in background
echo "3. Starting mcloudd server..."
./mcloudd &
SERVER_PID=$!
echo "✓ Server started (PID: $SERVER_PID)"

# Wait for server to be ready
sleep 3

# Run the init command
echo
echo "4. Running 'mcloudctl init' command..."
echo
./mcloudctl init --name test-cluster

# Cleanup
echo
echo "5. Cleaning up..."
kill $SERVER_PID 2>/dev/null || true
echo "✓ Server stopped"

echo
echo "=== Test completed successfully! ==="
echo
echo "You can inspect the database:"
echo "  sqlite3 mcloud.db 'SELECT * FROM clusters;'"
echo "  sqlite3 mcloud.db 'SELECT * FROM nodes;'"
echo "  sqlite3 mcloud.db 'SELECT * FROM bootstrap_tokens;'"
echo "  sqlite3 mcloud.db 'SELECT * FROM kv_store;'"
