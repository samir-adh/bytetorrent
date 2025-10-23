#!/usr/bin/env bash

# This file is used to clean our environment when we are done using it
# It will stop and remove the containers as well as removing the files 
# so we have a fresh start the next time we run 'start.sh'

set -e

echo "=== Cleaning BitTorrent Test Environment ==="

# Stop and remove containers
echo "Stopping containers..."
docker compose down -v

# Remove directories
echo "Removing directories..."
rm -rf downloads1 downloads2 watch1 watch2 torrents test-files/test-file.dat

# Optional: Keep test-files directory
# Uncomment to also remove test files:
# rm -rf test-files

echo "✓ Cleanup complete!"