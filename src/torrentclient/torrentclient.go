package torrentclient

import (
	"fmt"
	"net/http"
	"os"

	mgr "github.com/samir-adh/bytetorrent/src/downloadmanager"
	"github.com/samir-adh/bytetorrent/src/log"
	pc "github.com/samir-adh/bytetorrent/src/piece"
	"github.com/samir-adh/bytetorrent/src/torrentfile"
	tr "github.com/samir-adh/bytetorrent/src/tracker"
)

type TorrentClient struct {
	InfoHash [20]byte
	SelfId   [20]byte
	Port     int
	Peers    []tr.Peer
	Pieces   []pc.Piece
	PieceLength int
	FileName string
	Logger   *log.Logger
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
			State:  pc.Missing,
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
		InfoHash: tor.InfoHash,
		SelfId:   self_id,
		Port:     port,
		Peers:    peers,
		Pieces:   pieces,
		FileName: tor.Name,
		Logger:   logger,
		PieceLength: tor.PieceLength,
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
	wp := mgr.NewWorkerPool(client.SelfId, client.InfoHash, client.Peers, client.Pieces, client.PieceLength, file, client.Logger)
	wp.Start()
	return nil
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
