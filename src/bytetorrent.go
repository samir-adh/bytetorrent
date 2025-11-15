package main

import (
	"os"

	"github.com/samir-adh/bytetorrent/src/torrentclient"
	"github.com/ztrue/tracerr"
)

func main() {
	filepath := "./test-env/torrents/test.torrent"
	if len(os.Args) > 1 {
		filepath = os.Args[1]
	}
	client, err := torrentclient.New(filepath)
	if err != nil {
		tracerr.Print(err)
		os.Exit(1)
	}
	client.Download()
}
