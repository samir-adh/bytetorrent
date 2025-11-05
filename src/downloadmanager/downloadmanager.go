package downloadmanager

import (
	"fmt"
	"github.com/ztrue/tracerr"
	"log"
	"crypto/sha1"
	"math/rand"
	"net"
	"sync"
	"time"

	pr "github.com/samir-adh/bytetorrent/src/peerconnection"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
)

type WorkerPool struct {
	selfId      [20]byte
	infoHash    [20]byte
	peers       []tr.Peer
	pieceQueue  chan pc.Piece
	resultQueue chan pc.PieceResult
	quit        chan bool
	wg          sync.WaitGroup
}

func NewWorkerPool(selfId [20]byte, infoHash [20]byte, peers []tr.Peer, pieces []pc.Piece) *WorkerPool {
	jobQueue := make(chan pc.Piece, len(pieces))
	for _, piece := range pieces {
		jobQueue <- piece
	}
	return &WorkerPool{
		selfId:      selfId,
		infoHash:    infoHash,
		peers:       peers,
		pieceQueue:  jobQueue,
		resultQueue: make(chan pc.PieceResult, len(jobQueue)),
		quit:        make(chan bool),
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
		fmt.Printf("writing data of piece %d \n", result.Index)
		time.Sleep(time.Duration(rand.Intn(1e3)) * time.Microsecond) // Simulate download time
		if len(wp.resultQueue) == 0 && len(wp.pieceQueue) == 0 {
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
		log.Print(err.Error())
		return
	}
	defer func() {
		if err := netConn.Close(); err != nil {
			log.Print(err.Error())
		}
	}() // Close the connection when the function finishes
	peerConnection, err := pr.New(wp.selfId, peer, wp.infoHash, &netConn)
	if err != nil {
		log.Printf("could not connect to peer %s", (&peer).String())
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
					log.Printf("error downloading piece %d from peer %d: %s\n", piece.Index, peer.Id, result.Error.Error())
					close(wp.quit)
				}
			}
			wp.resultQueue <- *result

		case <-wp.quit:
			log.Printf("stopping connection to peer %d\n", peer.Id)
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

func (wp *WorkerPool) downloadPiece(piece *pc.Piece, peerConnection *pr.PeerConnection, netConn *net.Conn) *pc.PieceResult {


	// Check that the peer has the piece
	if !peerConnection.CanHandle(piece.Index) {
		return &pc.PieceResult{
			Index:   piece.Index,
			Payload: nil,
			Error:   ErrorMissingPiece{piece.Index},
		}
	}

	// Try to download the piece
	// log.Printf("Downloading piece %d from peer %d\n", piece.Index, peerConnection.Peer.Id)
	pieceResult, err := peerConnection.Download(piece, netConn)
	if err != nil {
		return &pc.PieceResult{
			Index:   piece.Index,
			Payload: nil,
			Error:   err,
		}
	}

	// Check integrity
	hash := sha1.Sum(pieceResult.Payload)
	if hash != piece.Hash {
		tracerr.Errorf("hash of downloaded piece %d doesn't match expected hash\n", piece.Index)
	}
	fmt.Printf("downloaded piece %d...\n", piece.Index)
	return pieceResult

}


func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
	// close(wp.quit)
	close(wp.pieceQueue)
	close(wp.resultQueue)
}
