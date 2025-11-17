package torrentclient

import (
	"crypto/sha1"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/samir-adh/bytetorrent/src/log"
	pr "github.com/samir-adh/bytetorrent/src/peerconnection"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
	"github.com/ztrue/tracerr"
)

type TorrentClient struct {
	InfoHash         [20]byte
	SelfId           [20]byte
	Port             int
	Peers            []tr.Peer
	Pieces           []pc.Piece
	PieceLength      int
	FileName         string
	Logger           *log.Logger
	DownloadedPieces []bool
}

func New(filepath string, logger *log.Logger) (*TorrentClient, error) {
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
	logger.Printf(log.HighVerbose, "tracker request: %s\n", trackerRequest)
	peers, err := tr.FindPeers(trackerRequest, httpClient)
	if err != nil {
		return nil, err
	}
	pieces := make([]pc.Piece, len(tor.PiecesHash))
	for i := range len(tor.PiecesHash) {
		pieces[i] = pc.Piece{
			Index:  i,
			Hash:   tor.PiecesHash[i],
			Length: tor.GetPieceLength(i),
		}
	}
	downloaded := make([]bool, len(tor.PiecesHash))
	for i := range downloaded {
		downloaded[i] = false
	}
	logger.Printf(log.LowVerbose, "Downloading %s", tor.Name)
	return &TorrentClient{
		InfoHash:         tor.InfoHash,
		SelfId:           self_id,
		Port:             port,
		Peers:            peers,
		Pieces:           pieces,
		FileName:         tor.Name,
		Logger:           logger,
		PieceLength:      tor.PieceLength,
		DownloadedPieces: downloaded,
	}, nil
}

func (client *TorrentClient) Download() error {
	// Create file to store the downloaded data
	file, err := os.Create(fmt.Sprintf("./downloads/%s", client.FileName))
	if err != nil {
		return err
	}
	fileSize := 0
	for _, piece := range client.Pieces {
		fileSize += piece.Length
	}
	client.Logger.Printf(log.HighVerbose, "creating file of size %d bytes", fileSize)
	err = file.Truncate(int64(fileSize))
	if err != nil {
		return err
	}
	defer file.Close()
	client.workerPool(
		file,
	)
	return nil
}

func (client *TorrentClient) workerPool(file *os.File) {
	piecesQueue := make(chan pc.Piece, len(client.Pieces))
	resultsQueue := make(chan pc.PieceResult, len(client.Pieces))
	quit := make(chan bool)
	wg := sync.WaitGroup{}
	for _, piece := range client.Pieces {
		piecesQueue <- piece
	}
	for _, peer := range client.Peers {
		wg.Go(func() {
			client.worker(
				peer,
				piecesQueue,
				resultsQueue,
				quit,
			)
		})
	}

	wg.Go(func() {
		client.collectPieces(
			file,
			resultsQueue,
			quit,
		)
	})
	wg.Wait()
	close(piecesQueue)
	close(resultsQueue)
}

func (client *TorrentClient) worker(
	peer tr.Peer,
	pieceQueue chan pc.Piece,
	resultsQueue chan pc.PieceResult,
	quit chan bool,
) {
	netConn, err := net.DialTimeout(
		"tcp",
		peer.AddressToStr(),
		5*time.Second,
	)
	if err != nil {
		client.Logger.Print(log.HighVerbose, err.Error())
		return
	}
	defer func() {
		if err := netConn.Close(); err != nil {
			client.Logger.Print(log.LowVerbose, err.Error())
		}
	}() // Close the connection when the function finishes
	peerConnection, err := pr.New(client.SelfId, peer, client.InfoHash, &netConn, client.Logger)
	if err != nil {
		client.Logger.Printf(log.LowVerbose, "could not connect to peer %s", (&peer).String())
		return
	}

	for {
		select {
		case piece := <-pieceQueue:
			result := client.downloadPiece(&piece, pieceQueue, peerConnection, &netConn)
			// check if piece is missing from peer
			switch result.State {
			case pc.Downloaded:
				resultsQueue <- *result
			case pc.Missing:
				pieceQueue <- piece
			default:
				client.Logger.Printf(log.HighVerbose, "error downloading piece %d from peer %d with state %d\n", piece.Index, peer.Id, result.State)
				close(quit)

			}
		case <-quit:
			client.Logger.Printf(log.HighVerbose, "stopping connection to peer %d\n", peer.Id)
			return
		}
	}
}

func (client *TorrentClient) downloadPiece(piece *pc.Piece, pieceQueue chan pc.Piece, peerConnection *pr.PeerConnection, netConn *net.Conn) *pc.PieceResult {

	// Check that the peer has the piece
	if !peerConnection.CanHandle(piece.Index) {
		pieceQueue <- *piece
	}

	// Try to download the piece
	client.Logger.Printf(log.HighVerbose, "downloading piece %d from peer %d\n", piece.Index, peerConnection.Peer.Id)
	pieceResult, err := peerConnection.Download(piece)
	client.Logger.Printf(log.HighVerbose, "downloaded piece %d from peer %d\n", piece.Index, peerConnection.Peer.Id)
	if err != nil {
		return &pc.PieceResult{
			Index:   piece.Index,
			Payload: nil,
			State:   pc.Failed,
		}
	}

	// Check integrity
	hash := sha1.Sum(pieceResult.Payload)
	if hash != piece.Hash {
		err = tracerr.Errorf("hash of downloaded piece %d doesn't match expected hash\n", piece.Index)
		return &pc.PieceResult{
			Index:   piece.Index,
			Payload: nil,
			State:   pc.Failed,
		}
	}
	return pieceResult

}

func (client *TorrentClient) collectPieces(file *os.File, resultsQueue chan pc.PieceResult, quit chan bool) {
	for result := range resultsQueue {
		if result.State != pc.Downloaded {
			client.Logger.Printf(log.LowVerbose,
				"failed to download piece %d data in state %d, aborting torrent.\n",
				result.Index, result.State)
			close(quit)
		}
		client.Logger.Printf(log.HighVerbose, "writing data of piece %d \n", result.Index)
		// time.Sleep(time.Duration(rand.Intn(1e3)) * time.Microsecond) // Simulate download time
		// startTime := time.Now()
		pieceDefaultSize := client.PieceLength
		bytesWritten, err := file.WriteAt(result.Payload, int64(result.Index*pieceDefaultSize))
		// ellapsedTime := time.Since(startTime)
		// wp.logger.Printf("writing piece data took %dms\n", ellapsedTime.Milliseconds())
		if err != nil || bytesWritten != len(result.Payload) {
			client.Logger.Printf(log.LowVerbose, "failed to write piece data, aborting torrent : %s\n", err)
			close(quit)
		}
		// client.completedMu.Lock()
		client.DownloadedPieces[result.Index] = true
		downloadIsCompleted := true
		completedCount := 0
		for _, pieceIsCompleted := range client.DownloadedPieces {
			downloadIsCompleted = downloadIsCompleted && pieceIsCompleted
			if pieceIsCompleted {
				completedCount += 1
			}
		}
		// client.completedMu.Unlock()
		percentageComplete := completedCount * 100 / len(client.DownloadedPieces)
		// wp.logger.Printf(log.LowVerbose,"download %d %% complete", percentageComplete)
		client.Logger.ProgressSimple(percentageComplete)
		if downloadIsCompleted {
			close(quit)
			return
		}
	}

}

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
