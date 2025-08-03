#!/bin/bash

# Test script for the static binary
# This verifies that the static binary works without CGO dependencies

set -e

echo "Testing static binary (Pure Go SQLite)..."

# Create a test directory
TEST_DIR=$(mktemp -d)
echo "Test directory: $TEST_DIR"

# Copy the static binary
cp bin/orchestrator-linux-static "$TEST_DIR/orchestrator"

# Change to test directory
cd "$TEST_DIR"

# Set environment variables for testing
export HOST=127.0.0.1
export PORT=8081
export DATABASE_PATH="./test.db"
export DATABASE_DRIVER=sqlite  # Force pure Go driver
export LOG_LEVEL=info
export FIRECRACKER_BINARY=/usr/bin/firecracker
export KERNEL_PATH=./vmlinux.bin
export ROOTFS_PATH=./rootfs.ext4

echo "Starting orchestrator in background..."

# Start the orchestrator in background
./orchestrator &
ORCHESTRATOR_PID=$!

# Wait for startup
sleep 3

echo "Testing health endpoint..."

# Test the health endpoint
if curl -s http://127.0.0.1:8081/api/v1/health | grep -q "healthy"; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
    exit 1
fi

echo "Testing stats endpoint..."

# Test the stats endpoint
if curl -s http://127.0.0.1:8081/api/v1/stats | grep -q "totalVMs"; then
    echo "✅ Stats endpoint passed"
else
    echo "❌ Stats endpoint failed"
    exit 1
fi

# Check if database was created
if [ -f "./test.db" ]; then
    echo "✅ Database file created successfully"
else
    echo "❌ Database file not created"
    exit 1
fi

# Clean up
echo "Stopping orchestrator..."
kill $ORCHESTRATOR_PID
wait $ORCHESTRATOR_PID 2>/dev/null || true

# Return to original directory
cd - > /dev/null

# Clean up test directory
rm -rf "$TEST_DIR"

echo "✅ All tests passed! Static binary is working correctly."
echo ""
echo "🎉 Your orchestrator is ready for deployment!"
echo ""
echo "To deploy on your Linux server:"
echo "1. Copy bin/orchestrator-linux-static to your server"
echo "2. Run: ./orchestrator-linux-static"
echo "3. Access: http://your-server:8080"