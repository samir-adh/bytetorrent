# BitTorrent Test Environment

A local BitTorrent testing environment with tracker and peers for development.

## Quick Start

```bash
# Start environment
./start.sh

# Check status
./status.sh

# View logs
docker-compose logs -f

# Clean up
./clean.sh
```

## Components

- **Tracker**: `http://localhost:6969/announce`
- **Seeder1**: Web UI at `http://localhost:9091`, peer port `51413`
- **Seeder2**: Web UI at `http://localhost:9092`, peer port `51414`

## Using with Your Client

1. Start the environment: `./start.sh`
2. Use the torrent file: `./torrents/test.torrent`
3. Your client should connect to tracker at `http://localhost:6969/announce`
4. Peers will be discovered and file transfer will begin

## Directory Structure

```
.
├── docker-compose.yml    # Container definitions
├── start.sh              # Setup script
├── clean.sh              # Cleanup script
├── status.sh             # Status checker
├── downloads1/           # Seeder1 downloads
├── downloads2/           # Seeder2 downloads
├── watch1/               # Seeder1 watch dir
├── watch2/               # Seeder2 watch dir
├── test-files/           # Source files
└── torrents/             # Generated torrents
```

## Testing Scenarios

### Add more peers

```bash
docker run -d --name bt-seeder3 \
  --network bittorrent-test-env_bt-network \
  -p 9093:9091 -p 51415:51413 \
  -v $(pwd)/watch3:/watch \
  -v $(pwd)/downloads3:/downloads \
  linuxserver/transmission
```

### Create custom torrent

```bash
transmission-create -o torrents/custom.torrent \
  -t http://bt-tracker:6969/announce \
  your-file.dat
```

## Troubleshooting

- Check logs: `docker-compose logs`
- Verify connectivity: `./status.sh`
- Restart: `docker-compose restart`
- Clean start: `./clean.sh && ./start.sh`