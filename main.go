package main

import (
	"fmt"
	"sync"

	pr "github.com/samir-adh/bytetorrent/peerconnection"
	tf "github.com/samir-adh/bytetorrent/torrentfile"
	tr "github.com/samir-adh/bytetorrent/tracker"
	"github.com/ztrue/tracerr"
)

func main() {
	filepath := "torrentfile/testdata/debian-12.10.0-amd64-netinst.iso.torrent"
	tor, err := tf.OpenTorrentFile(filepath)
	if err != nil {
		fmt.Printf("Error opening torrent file: %v\n", err)
		tracerr.PrintSource(err)
	}
	self_id, err := tr.RandomPeerId()
	if err != nil {
		tracerr.PrintSource(err)
	}
	port := 6881
	trackerRequest, err := tr.BuildTrackerRequest(tor, self_id, port)
	if err != nil {
		tracerr.PrintSource(err)
	}
	fmt.Printf("Tracker request: %s\n", trackerRequest)
	peers, err := tr.ConnectToTracker(trackerRequest)
	if err != nil {
		tracerr.PrintSource(err)
	}
	try_connect_all(peers, self_id, tor)
}

func try_connect(peer_connection *pr.PeerConnection) error {
	err := pr.ConnectToPeer(peer_connection)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func try_connect_all(peers []tr.Peer, self_id [20]byte, tor *tf.TorrentFile) {
	wg := sync.WaitGroup{}
	for i, peer := range peers {
		wg.Add(1)
		go func() {
			peer_connection := &pr.PeerConnection{
				SelfId:      self_id,
				Peer:        peer,
				TorrentFile: tor,
			}
			err := try_connect(peer_connection)
			if err != nil {
				fmt.Printf("Error connecting to peer %d at %s: %v\n", i, peer.String(), err)
			} else {
				fmt.Printf("Connected to peer %d at %s\n", i, peer.String())
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
