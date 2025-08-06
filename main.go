package main

import (
	"github.com/samir-adh/bytetorrent/torrentClient"
)

func main() {
	filepath := "torrentFile/testdata/debian-12.10.0-amd64-netinst.iso.torrent"
	client,err := torrentClient.New(filepath)
	if err != nil {
		panic(err)
	}
	client.Start()
}
