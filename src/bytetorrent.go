package main

import (
	"flag"
	"os"

	"github.com/samir-adh/bytetorrent/src/log"
	"github.com/samir-adh/bytetorrent/src/torrentclient"
	"github.com/ztrue/tracerr"
)

func main() {
	defaultFilepath := "./test-env/torrents/test.torrent"
	filepath := flag.String("f", defaultFilepath, "Torrent file to download")
	verbose := flag.Bool("v", false, "Enable verbose output mode")
	flag.Parse()
	verboseLevel := log.LowVerbose
	if *verbose {
		verboseLevel = log.HighVerbose
	}
	logger := log.Logger{Verbose: verboseLevel}
	client, err := torrentclient.New(*filepath, &logger)
	if err != nil {
		tracerr.Print(err)
		os.Exit(1)
	}
	client.Download()
}
