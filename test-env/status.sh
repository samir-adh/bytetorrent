#!/usr/bin/env bash

# This file allows us to checl the status of the different containers
# It checks that each container has access to the torrent
# and that the can communicate between each other

echo "=== BitTorrent Test Environment Status ==="
echo ""

# Check container status
echo "Container Status:"
docker-compose ps
echo ""

# Check seeder1
echo "=== Seeder1 (localhost:9091) ==="
docker exec bt-seeder1 transmission-remote -l 2>/dev/null || echo "Not responding"
echo ""

# Check seeder2
echo "=== Seeder2 (localhost:9092) ==="
docker exec bt-seeder2 transmission-remote -l 2>/dev/null || echo "Not responding"
echo ""

# Check tracker
echo "Tracker Status:"
curl -s http://localhost:6969/scrape | head -c 100
echo ""
echo ""

# Network connectivity
echo "Network Connectivity:"
docker exec bt-seeder2 ping -c 1 bt-seeder1 > /dev/null 2>&1 && \
    echo "✓ seeder2 can reach seeder1" || \
    echo "✗ seeder2 CANNOT reach seeder1"

docker exec bt-seeder1 curl -s http://bt-tracker:6969/announce > /dev/null 2>&1 && \
    echo "✓ seeder1 can reach tracker" || \
    echo "✗ seeder1 CANNOT reach tracker"