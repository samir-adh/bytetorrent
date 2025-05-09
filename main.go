package main

import (
	"fmt"

	tf "github.com/samir-adh/bytetorrent/torrentfile"
	tr "github.com/samir-adh/bytetorrent/tracker"
)

func main() {
	filepath := "torrentfile/testdata/debian-12.10.0-amd64-netinst.iso.torrent"
	tor, err := tf.OpenTorrentFile(filepath)
	if err != nil {
		panic(err)
	}
	peer_id, err := tr.GeneratePeerId()
	if err != nil {
		panic(err)
	}
	port := 6881
	trackerRequest, err := tr.BuildTrackerRequest(tor, peer_id, port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Tracker request : %s", trackerRequest)
	body, err := tr.ConnectToTracker(trackerRequest)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response body : %s", string(body))
}
