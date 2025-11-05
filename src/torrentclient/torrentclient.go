package torrentclient

import (
	/*
	"crypto/sha1"
	"log"
	"net"
	"sync"
	"time"
	"github.com/ztrue/tracerr"
	pr "github.com/samir-adh/bytetorrent/src/peerconnection"
	*/
	"fmt"
	"net/http"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
	mgr "github.com/samir-adh/bytetorrent/src/downloadmanager"
)

type TorrentClient struct {
	InfoHash [20]byte
	SelfId   [20]byte
	Port     int
	Peers    []tr.Peer
	Pieces   []pc.Piece
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
	pieces := make([]pc.Piece, len(tor.PiecesHash))
	for i := range len(tor.PiecesHash) {
		pieces[i] = pc.Piece{
			Index:  i,
			State:  pc.Missing,
			Hash:   tor.PiecesHash[i],
			Length: tor.GetPieceLength(i),
		}
	}
	downloaded := make([]bool, len(tor.PiecesHash))
	for i := range downloaded {
		downloaded[i] = false
	}
	return &TorrentClient{
		InfoHash: tor.InfoHash,
		SelfId:   self_id,
		Port:     port,
		Peers:    peers,
		Pieces:   pieces,
	}, nil
}

func (client *TorrentClient) Download() {
	wp := mgr.NewWorkerPool(client.SelfId, client.InfoHash, client.Peers, client.Pieces)
	wp.Start()
}

/*

type Downloaded struct {
	tab []bool
	mu  sync.Mutex
}


func (client *TorrentClient) Download() {
	dw := Downloaded{
		make([]bool, len(client.Pieces)),
		sync.Mutex{},
	}
	workQueue := make(chan *pc.Piece, len(client.Pieces))
	results := make(chan *pc.PieceResult, len(client.Pieces))
	for _, peer := range client.Peers {
		client.downloadRoutine(peer, &dw, workQueue, results)
	}
}

type ErrorMissingPiece struct {
	PieceIndex int
}

func (err ErrorMissingPiece) Error() string {
	return fmt.Sprintf("Peer doesn't have piece %d", err.PieceIndex)
}

func (client *TorrentClient) downloadRoutine(peer tr.Peer, downloaded *Downloaded, work chan *pc.Piece, results chan*pc.PieceResult) {
	// Initiate connection with peer
	netConn, err := net.DialTimeout(
		"tcp",
		peer.String(),
		5*time.Second,
	)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer func() {
		if err := netConn.Close(); err != nil {
			log.Print(err.Error())
		}
	}() // Close the connection when the function finishes
	peerConnection, err := pr.New(client.SelfId, peer, client.InfoHash, &netConn)
	if err != nil {
		log.Printf("Could not connect to peer %s", (&peer).String())
		return
	}

	// Start trying to download pieces
	fmt.Printf("There are %d pieces in the file\n", len(client.Pieces))
	for _, piece := range client.Pieces {
		pieceResult, err := client.downloadPiece(piece, peerConnection, &netConn)
		if err != nil {
			log.Println(err)
			continue
		}
		downloaded.mu.Lock()
		downloaded.tab[piece.Index] = true
		downloaded.mu.Unlock()
		results <- pieceResult
	}
}

func (client *TorrentClient) downloadPiece(piece *pc.Piece, peerConnection *pr.PeerConnection, netConn *net.Conn) (*pc.PieceResult, error) {
	// Check that the piece is not already downloaded
	// if client.Downloaded[piece.Index] {
	// 	return nil, tracerr.Errorf("piece %d is already downloaded", piece.Index)
	// }

	// Check that the peer has the piece
	if !peerConnection.CanHandle(piece.Index) {
		return nil, ErrorMissingPiece{piece.Index}
	}

	// Try to download the piece
	fmt.Printf("Downloading piece %d...\n", piece.Index)
	pieceResult, err := peerConnection.Download(piece, netConn)
	if err != nil {
		return nil, err
	}

	// Check integrity
	hash := sha1.Sum(pieceResult.Payload)
	if hash != piece.Hash {
		tracerr.Errorf("hash of downloaded piece %d doesn't match expected hash\n", piece.Index)
	}
	fmt.Printf("Downloaded piece %d...\n", piece.Index)
	return pieceResult, nil

}

*/

/*
Plan to make the download concurrent

What are the steps :

	we get a list of all pieces

	turn this list into a work queue

	for each peer we start a coroutine

	for each peer coroutine we pick a piece and try to download it

	if the peer has it
		we download it
		we write it to the disk (maybe we shoud use a collector coroutine to collect the pieces asyncsly and write them)

	else
		we put it back to the work queue

	pick the next piece

Questions :

	how do long do each peer coroutine need to run ?
	We stop them when the are no more pieces in the work queue
	-> that's a first solution but the program will keep running forever if a piece is impossible to download

	how do we handle the writing of the pieces on the disk ?
	- we can write just after downloading a piece in the coroutine
	- we can send the piece data to a object like a collector that will write in the disk
	- maybe something else ?

What we shoud do now :
	take a look a go concurency patterns to try to find a pattern that solves our problems
	implement it as it even if it looks overengineered
	(optional) optimize it

	i have take a look at a few patterns and the worker pool pattern seems to be the most suitable
*/