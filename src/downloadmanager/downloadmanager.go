package downloadmanager

import (
	"crypto/sha1"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/samir-adh/bytetorrent/src/log"
	pr "github.com/samir-adh/bytetorrent/src/peerconnection"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
	"github.com/ztrue/tracerr"
)

type WorkerPool struct {
	selfId      [20]byte
	infoHash    [20]byte
	peers       []tr.Peer
	pieces		[]pc.Piece
	pieceQueue  chan pc.Piece
	resultQueue chan pc.PieceResult
	quit        chan bool
	completed   []bool
	completedMu sync.Mutex
	wg          sync.WaitGroup
	file        *os.File
	logger      *log.Logger
	pieceLength int
}

func NewWorkerPool(selfId [20]byte, infoHash [20]byte, peers []tr.Peer, pieces []pc.Piece,pieceLength int, file *os.File, logger *log.Logger) *WorkerPool {
	jobQueue := make(chan pc.Piece, len(pieces))
	for _, piece := range pieces {
		jobQueue <- piece
	}
	return &WorkerPool{
		selfId:      selfId,
		infoHash:    infoHash,
		peers:       peers,
		pieces: pieces,
		pieceQueue:  jobQueue,
		resultQueue: make(chan pc.PieceResult, len(jobQueue)),
		quit:        make(chan bool),
		file:        file,
		completed:   make([]bool, len(pieces)),
		logger:      logger,
		pieceLength : pieceLength,
	}

}

func (wp *WorkerPool) Start() {
	for _, peer := range wp.peers {
		wp.wg.Go(func() {
			wp.worker(peer)
		})
	}

	wp.wg.Go(wp.collectPieces)
	wp.Stop()

}

func (wp *WorkerPool) collectPieces() {
	for result := range wp.resultQueue {
		if result.Error != nil {
			wp.logger.Printf(log.LowVerbose, "failed to download piece %d data, aborting torrent : %s\n", result.Index, result.Error)
			close(wp.quit)
		}
		wp.logger.Printf(log.HighVerbose, "writing data of piece %d \n", result.Index)
		// time.Sleep(time.Duration(rand.Intn(1e3)) * time.Microsecond) // Simulate download time
		// startTime := time.Now()
		pieceDefaultSize := wp.pieceLength
		bytesWritten, err := wp.file.WriteAt(result.Payload, int64(result.Index*pieceDefaultSize))
		// ellapsedTime := time.Since(startTime)
		// wp.logger.Printf("writing piece data took %dms\n", ellapsedTime.Milliseconds())
		if err != nil || bytesWritten != len(result.Payload) {
			wp.logger.Printf(log.LowVerbose, "failed to write piece data, aborting torrent : %s\n", err)
			close(wp.quit)
		}
		wp.completedMu.Lock()
		wp.completed[result.Index] = true
		downloadIsCompleted := true
		completedCount := 0
		for _, pieceIsCompleted := range wp.completed {
			downloadIsCompleted = downloadIsCompleted && pieceIsCompleted
			if pieceIsCompleted {
				completedCount += 1
			}
		}
		wp.completedMu.Unlock()
		percentageComplete := completedCount * 100 / len(wp.completed)
		// wp.logger.Printf(log.LowVerbose,"download %d %% complete", percentageComplete)
		wp.logger.ProgressSimple(percentageComplete)
		if downloadIsCompleted {
			close(wp.quit)
			return
		}
	}

}

func (wp *WorkerPool) worker(peer tr.Peer) {
	// Initiate connection with peer
	netConn, err := net.DialTimeout(
		"tcp",
		peer.AddressToStr(),
		5*time.Second,
	)
	if err != nil {
		wp.logger.Print(log.HighVerbose, err.Error())
		return
	}
	defer func() {
		if err := netConn.Close(); err != nil {
			wp.logger.Print(log.LowVerbose, err.Error())
		}
	}() // Close the connection when the function finishes
	peerConnection, err := pr.New(wp.selfId, peer, wp.infoHash, &netConn, wp.logger)
	if err != nil {
		wp.logger.Printf(log.LowVerbose, "could not connect to peer %s", (&peer).String())
		return
	}

	for {
		select {
		case piece := <-wp.pieceQueue:
			// fmt.Printf("Connection to peer %d processing job %d\n", peer.Id, job.Index)
			result := wp.downloadPiece(&piece, peerConnection, &netConn)
			// check if piece is missing from peer

			if result.Error != nil {
				if _, ok := result.Error.(ErrorMissingPiece); ok {
					wp.pieceQueue <- piece
				} else {
					wp.logger.Printf(log.HighVerbose, "error downloading piece %d from peer %d: %s\n", piece.Index, peer.Id, result.Error.Error())
					close(wp.quit)
				}
			}
			wp.resultQueue <- *result

		case <-wp.quit:
			wp.logger.Printf(log.HighVerbose, "stopping connection to peer %d\n", peer.Id)
			return
		}
	}
}

type ErrorMissingPiece struct {
	PieceIndex int
}


func (err ErrorMissingPiece) Error() string {
	return fmt.Sprintf("piece %d is missing from peer", err.PieceIndex)
}
type ErrorConnectionFailed struct {
	PeerId int
}
func (err ErrorConnectionFailed) Error() string {
	return  fmt.Sprintf("failed to connect to peer %d", err.PeerId)
}

func (wp *WorkerPool) downloadPiece(piece *pc.Piece, peerConnection *pr.PeerConnection, netConn *net.Conn) *pc.PieceResult {

	// Check that the peer has the piece
	if !peerConnection.CanHandle(piece.Index) {
		wp.pieceQueue <- *piece
	}

	// Try to download the piece
	wp.logger.Printf(log.HighVerbose, "downloading piece %d from peer %d\n", piece.Index, peerConnection.Peer.Id)
	pieceResult, err := peerConnection.Download(piece, netConn)
	wp.logger.Printf(log.HighVerbose, "downloaded piece %d from peer %d\n", piece.Index, peerConnection.Peer.Id)
	if err != nil {
		// wp.logger.Print(log.HighVerbose, err)
		return &pc.PieceResult{
			Index:   piece.Index,
			Payload: nil,
			Error:   err,
		}
	}

	// Check integrity
	hash := sha1.Sum(pieceResult.Payload)
	if hash != piece.Hash {
		err = tracerr.Errorf("hash of downloaded piece %d doesn't match expected hash\n", piece.Index)
		return &pc.PieceResult{
			Index: piece.Index,
			Payload: nil,
			Error: err,
		}
	}
	// wp.logger.Printf("downloaded piece %d...\n", piece.Index)
	return pieceResult

}

func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
	// close(wp.quit)
	close(wp.pieceQueue)
	close(wp.resultQueue)
}
