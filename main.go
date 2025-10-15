package main

import (
	// "fmt"

	"os"

	"github.com/samir-adh/bytetorrent/torrentclient"
	"github.com/ztrue/tracerr"
	//"runtime/debug"
)

func main() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Printf("panic recovered: %v\n", r)
	// 		debug.PrintStack() // prints full stack trace
	// 	}
	// }()
	filepath := "torrentfile/testdata/debian-13.1.0-amd64-netinst.iso.torrent"
	client, err := torrentclient.New(filepath)
	if err != nil {
		tracerr.Print(err)
		os.Exit(1)
	}
	client.Start()
}
