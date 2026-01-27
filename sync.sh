#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check if venv exists
if [ ! -d ".venv" ]; then
    echo "Virtual environment not found. Run ./setup.sh first."
    exit 1
fi

# Check if MongoDB is running
if ! docker compose ps --status running 2>/dev/null | grep -q mongodb; then
    echo "MongoDB is not running. Starting..."
    docker compose up -d
    echo "Waiting for MongoDB to be ready..."
    sleep 3
fi

# Run the sync script
.venv/bin/python sync_to_mongodb.py "$@"
