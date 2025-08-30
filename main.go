package main

import (
	"fmt"
	"github.com/samir-adh/bytetorrent/torrentClient"
	"runtime/debug"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic recovered: %v\n", r)
			debug.PrintStack() // prints full stack trace
		}
	}()
	filepath := "torrentfile/testdata/debian-12.10.0-amd64-netinst.iso.torrent"
	client, err := torrentClient.New(filepath)
	if err != nil {
		panic(err)
	}
	client.Start()
}
