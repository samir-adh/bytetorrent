#!/usr/bin/env bash

# This file is used to create and setup containers that will emulate a BitTorrent environment with:
# - A tracker that will anounce the torrent
# - Seeder 1 that already have the file and can seed it
# - Seeder 2 that will download the file and then seed it
# The goal is to have a controlled and reproducible environment for testing our BitTorrent client

set -e

echo "=== BitTorrent Test Environment Setup ==="

# Create directory structure
echo "Creating directories..."
mkdir -p watch1 watch2 downloads1 downloads2 test-files torrents
mkdir -p downloads1/complete downloads1/incomplete downloads2/complete downloads2/incomplete

# Create test file if it doesn't exist
if [ ! -f test-files/test-file.dat ]; then
    echo "Creating test file..."
    # echo "Hello from BitTorrent test environment!" > test-files/test-file.dat
    dd if=/dev/urandom of=test-files/test-file.dat bs=8M count=1 iflag=fullblock
fi

# Start Docker containers
echo "Starting Docker containers..."
docker compose up -d

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 10

# Create torrent file for containers
echo "Creating torrent file..."
transmission-create -o torrents/test.torrent \
    -t http://bt-tracker:6969/announce \
    test-files/test-file.dat

# Copy complete file to seeder1
echo "Preparing seeder1 with complete file..."
cp test-files/test-file.dat downloads1/complete

# Copy torrent to watch directories
echo "Adding torrents to watch directories..."
cp torrents/test.torrent watch1/
cp torrents/test.torrent watch2/

# Wait for torrents to be processed
echo "Waiting for torrents to be added..."
sleep 5

echo ""
echo "=== Setup Complete! ==="
echo "Tracker:         http://localhost:6969/announce"
echo "Seeder1 Web UI:  http://localhost:9091"
echo "Seeder2 Web UI:  http://localhost:9092"
echo "Torrent file:    ./torrents/test.torrent"
echo ""
echo "Check status with: docker exec bt-seeder1 transmission-remote -l"
