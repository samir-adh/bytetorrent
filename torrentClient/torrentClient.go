package torrentClient

import (
	"fmt"
	"sync"

	"github.com/samir-adh/bytetorrent/peerConnection"
	"github.com/samir-adh/bytetorrent/torrentFile"
	"github.com/samir-adh/bytetorrent/tracker"
)

type PieceState int

const (
	Missing PieceState = iota
	Downloading
	Downloaded
)

type TorrentClient struct {
	File        torrentFile.TorrentFile
	SelfId      [20]byte
	Port        int
	Peers       []tracker.Peer
	PiecesState []PieceState
}

func New(filepath string) (*TorrentClient, error) {
	tor, err := torrentFile.OpenTorrentFile(filepath)
	if err != nil {
		return nil, err
	}
	self_id, err := tracker.RandomPeerId()
	if err != nil {
		return nil, err
	}
	port := 6881
	trackerRequest, err := tracker.BuildTrackerRequest(tor, self_id, port)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Tracker request: %s\n", trackerRequest)
	peers, err := tracker.FindPeers(trackerRequest)
	if err != nil {
		return nil, err
	}
	piecesState := make([]PieceState, len(tor.PiecesHash))
	for i := range piecesState {
		piecesState[i] = Missing
	}
	return &TorrentClient{
		File:        *tor,
		SelfId:      self_id,
		Port:        port,
		Peers:       peers,
		PiecesState: piecesState,
	}, nil
}

func (client *TorrentClient) Start() {
	wg := sync.WaitGroup{}
	for i, peer := range client.Peers {
		wg.Add(1)
		go func() {
			peerConnection := peerConnection.PeerConnection{
				SelfId: client.SelfId,
				Peer: peer,
				TorrentFile: &client.File,
			}
			
		}
	}
}

func try_connect(peer_connection *peerConnection.PeerConnection) error {
	err := peerConnection.ConnectToPeer(peer_connection)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func try_connect_all(peers []tracker.Peer, self_id [20]byte, tor *torrentFile.TorrentFile) {
	wg := sync.WaitGroup{}
	for i, peer := range peers {
		wg.Add(1)
		go func() {
			peer_connection := &peerConnection.PeerConnection{
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
