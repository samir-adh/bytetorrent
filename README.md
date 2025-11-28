# Bytetorrent

A BitTorent client written in go from scratch.

## Installation

```bash
# Clone this repository
git clone https://github.com/samir-adh/bytetorrent

# Install bytetorrent
go install bytetorrent/src/bytetorrent.go
```

## Usage

```bash
bytetorrent <your_torrent_file>
```

Try to download the Debian 13 disk image !

```bash
bytetorrent -f test-env/torrents/debian-13.1.0-amd64-netinst.iso.torrent
```
