package main

import (
	"flag"
	"os"

	"github.com/samir-adh/bytetorrent/src/log"
	"github.com/samir-adh/bytetorrent/src/torrentclient"
	"github.com/ztrue/tracerr"
)

func main() {
	filepath := "./test-env/torrents/test.torrent"
	verbose := flag.Bool("v", false, "Enable verbose output mode")
	verboseLevel := log.LowVerbose
	if *verbose {
		verboseLevel = log.HighVerbose
	}
	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		filepath = args[0]
	}
	logger := log.Logger{Verbose: verboseLevel}
	client, err := torrentclient.New(filepath, &logger)
	if err != nil {
		tracerr.Print(err)
		os.Exit(1)
	}
	client.Download()
}
