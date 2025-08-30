package torrentClient

import (
	"fmt"
	"sync"

	"github.com/samir-adh/bytetorrent/peerconnection"
	"github.com/samir-adh/bytetorrent/piece"
	"github.com/samir-adh/bytetorrent/torrentfile"
	"github.com/samir-adh/bytetorrent/tracker"
)

type TorrentClient struct {
	File            torrentfile.TorrentFile
	SelfId          [20]byte
	Port            int
	Peers           []tracker.Peer
	PeerConnections []peerconnection.PeerConnection
	Queue           []piece.Piece
	QueueMutex      sync.Mutex
}

func New(filepath string) (*TorrentClient, error) {
	tor, err := torrentfile.OpenTorrentFile(filepath)
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
	piecesQueue := make([]piece.Piece, len(tor.PiecesHash))
	for i := range len(tor.PiecesHash) {
		piecesQueue = append(piecesQueue, piece.Piece{
			Index: i,
			State: piece.Missing,
			Hash:  tor.PiecesHash[i],
		})
	}
	return &TorrentClient{
		File:   *tor,
		SelfId: self_id,
		Port:   port,
		Peers:  peers,
		Queue:  piecesQueue,
	}, nil
}
func (client *TorrentClient) Start() {
	client.initiatePeerConnections()
}

func (client *TorrentClient) initiatePeerConnections() {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for i, peer := range client.Peers {
		wg.Add(1)
		go func(peer tracker.Peer) {
			defer wg.Done()
			// Spawn peer connection
			peerConnection, err := peerconnection.New(
				client.SelfId,
				peer,
				client.File,
			)
			if err != nil {
				fmt.Errorf("failed to connect to peer %d : %s\n", i, err)

			}
			mu.Lock()
			client.PeerConnections = append(client.PeerConnections, *peerConnection)
			mu.Unlock()
		}(peer)
	}
	wg.Wait()
}

func (client *TorrentClient) startDownloading() {
	wg := sync.WaitGroup{}
	for _,piece := range client.Queue {
		wg.Add(1)
		for _, peerConnection := range client.PeerConnections {
			if peerConnection.CanHandle(piece.Index) {
				go peerConnection.Download(piece, &wg)
			}
		}

	}
	wg.Wait()
}
