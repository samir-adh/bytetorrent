package torrentclient

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/samir-adh/bytetorrent/src/peerconnection"
	"github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	"github.com/samir-adh/bytetorrent/src/tracker"
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

	httpClient := http.DefaultClient
	fmt.Printf("Tracker request: %s\n", trackerRequest)
	peers, err := tracker.FindPeers(trackerRequest, httpClient)
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
	client.startDownloading()
}

func (client *TorrentClient) initiatePeerConnections() {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	fmt.Println("Connecting to peers...")
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
				fmt.Printf("failed to connect to peer %d : %s\n", i, err)

			} else {
				mu.Lock()
				client.PeerConnections = append(client.PeerConnections, *peerConnection)
				mu.Unlock()
			}
		}(peer)
	}
	wg.Wait()
}

func (client *TorrentClient) startDownloading() {
	fmt.Println("Starting download...")
	wg := sync.WaitGroup{}
	//n := len(client.Queue)
	for _, piece := range client.Queue {
		// fmt.Printf("Downloading piece %d of %d\n", i, n)
		wg.Add(1)
		for _, peerConnection := range client.PeerConnections {
			if peerConnection.CanHandle(piece.Index) {
				go peerConnection.Download(piece)
				break
			}

		}
		wg.Done()
	}
	wg.Wait()
}
