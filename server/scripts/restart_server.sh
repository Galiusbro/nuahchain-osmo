#!/bin/bash

# Script to properly restart the server
# Kills existing server process and starts a new one

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT/server"

PID_FILE="/tmp/server.pid"
LOG_FILE="/tmp/server.log"

echo "🔄 Restarting server..."

# Kill process on port 8080 first (most reliable)
echo "   Checking port 8080..."
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "   Killing process on port 8080..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    sleep 1
fi

# Kill existing server if running
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p "$OLD_PID" > /dev/null 2>&1; then
        echo "   Killing existing server (PID: $OLD_PID)..."
        kill -9 "$OLD_PID" 2>/dev/null || true
        sleep 1
    fi
    rm -f "$PID_FILE"
fi

# Kill any other go run processes
pkill -9 -f "go run.*main.go" 2>/dev/null || true
pkill -9 -f "server/main" 2>/dev/null || true
sleep 2

# Double check port is free
if lsof -ti:8080 > /dev/null 2>&1; then
    echo "   ⚠️  Port 8080 still in use, force killing..."
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    sleep 1
fi

# Start new server
echo "   Starting new server..."
go run main.go > "$LOG_FILE" 2>&1 &
NEW_PID=$!
echo "$NEW_PID" > "$PID_FILE"

echo "   Server started with PID: $NEW_PID"
echo "   Log file: $LOG_FILE"
echo "   PID file: $PID_FILE"

# Wait a bit and check if server is running
sleep 3
if ps -p "$NEW_PID" > /dev/null 2>&1; then
    echo "✅ Server is running"

    # Check if server responds
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Server is responding"
    else
        echo "⚠️  Server started but not responding yet"
    fi
else
    echo "❌ Server failed to start"
    echo "   Check logs: tail -f $LOG_FILE"
    exit 1
fi

echo ""
echo "To stop server: kill \$(cat $PID_FILE)"
echo "To view logs: tail -f $LOG_FILE"

