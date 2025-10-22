package torrentclient

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/samir-adh/bytetorrent/src/peerconnection"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
	"github.com/ztrue/tracerr"
)

type TorrentClient struct {
	File            torrentfile.TorrentFile
	SelfId          [20]byte
	Port            int
	Peers           []tr.Peer
	PeerConnections []peerconnection.PeerConnection
	Queue           []pc.Piece
	QueueMutex      sync.Mutex
}

func New(filepath string) (*TorrentClient, error) {
	tor, err := torrentfile.OpenTorrentFile(filepath)
	if err != nil {
		return nil, err
	}
	self_id, err := tr.RandomPeerId()
	if err != nil {
		return nil, err
	}
	port := 6881
	trackerRequest, err := tr.BuildTrackerRequest(tor, self_id, port)
	if err != nil {
		return nil, err
	}

	httpClient := http.DefaultClient
	fmt.Printf("Tracker request: %s\n", trackerRequest)
	peers, err := tr.FindPeers(trackerRequest, httpClient)
	if err != nil {
		return nil, err
	}
	piecesQueue := make([]pc.Piece, len(tor.PiecesHash))
	for i := range len(tor.PiecesHash) {
		piecesQueue = append(piecesQueue, pc.Piece{
			Index:  i,
			State:  pc.Missing,
			Hash:   tor.PiecesHash[i],
			Length: tor.GetPieceLength(i),
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
		go func(peer tr.Peer) {
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

type ErrorMissingPiece struct {
	PieceIndex int
}

func (err *ErrorMissingPiece) Error() string {
	return fmt.Sprintf("Peer doesn't have piece %d")
}

func (client *TorrentClient) downloadRoutine(peer tr.Peer, piece pc.Piece) (*pc.PieceResult, error) {
	netConn, err := net.DialTimeout(
		"tcp",
		peer.String(),
		5*time.Second,
	)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer netConn.Close()
	peerConnection, err := peerconnection.New(client.SelfId, peer, client.File, &netConn)
	if err != nil {
		log.Printf("Could not connect to peer %s", (&peer).String())
		return  nil, err
	}
	if !peerConnection.CanHandle(piece.Index) {
		return nil, tracerr.Wrap(&ErrorMissingPiece{piece.Index})
	}
	pieceResult, err := peerConnection.Download(piece, &netConn)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return pieceResult, nil
}

func (client *TorrentClient) startDownloading() {
	fmt.Println("Starting download...")
	
}
